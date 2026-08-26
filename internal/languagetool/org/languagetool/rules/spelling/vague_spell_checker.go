package spelling

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	atticmorfo "github.com/lucasew/lang/internal/attic/morfologik"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// VagueSpellChecker ports org.languagetool.rules.spelling.VagueSpellChecker.
//
// A spell checker that's fast but not guaranteed to always agree with the
// "real" spell checker used in LT. Doesn't offer corrections.
// Can be used for guessing the language of short texts where ngram
// or AI-based language identification isn't exact enough.
//
// Java static caches langToRule / langToDict are process-level maps keyed by
// Language object identity; Go keys by GetShortCodeWithCountryAndVariant (same
// Languages registry entry for a given code).
type VagueSpellChecker struct{}

// NewVagueSpellChecker ports new VagueSpellChecker().
func NewVagueSpellChecker() *VagueSpellChecker {
	return &VagueSpellChecker{}
}

// Java: private final static Map<Language, Rule> langToRule
// Java: private final static Map<Language, Dictionary> langToDict
var (
	vagueLangToRuleMu sync.Mutex
	vagueLangToRule   = map[string]any{}
	vagueLangToDict   = map[string]*atticmorfo.Dictionary{}
)

// clearVagueSpellCheckerCaches clears process caches (tests).
func clearVagueSpellCheckerCaches() {
	vagueLangToRuleMu.Lock()
	vagueLangToRule = map[string]any{}
	vagueLangToDict = map[string]*atticmorfo.Dictionary{}
	vagueLangToRuleMu.Unlock()
}

// IsValidWord ports VagueSpellChecker.isValidWord(String word, Language lang).
func (v *VagueSpellChecker) IsValidWord(word string, lang languagetool.Language) bool {
	if v == nil || lang == nil {
		panic("VagueSpellChecker.isValidWord: nil receiver or language")
	}
	key := languageCacheKey(lang)

	vagueLangToRuleMu.Lock()
	rule := vagueLangToRule[key]
	if rule == nil {
		tempRule := getSpellingCheckRule(lang)
		switch r := tempRule.(type) {
		case *hunspellRuleForVague:
			rule = r
		case *morfologikSpellerRuleForVague:
			// Java: rule = new NonThreadSafeSpellRule(messages, lang, null)
			rule = newNonThreadSafeSpellRule(lang)
		default:
			vagueLangToRuleMu.Unlock()
			panic(fmt.Sprintf("Unknown rule type for language %s: %T",
				lang.GetShortCodeWithCountryAndVariant(), tempRule))
		}
		vagueLangToRule[key] = rule
	}
	vagueLangToRuleMu.Unlock()

	switch r := rule.(type) {
	case *nonThreadSafeSpellRule:
		// Java: indicates a Morfologik-based speller — not thread-safe, re-create Speller in isMisspelled()
		// Note: NonThreadSafeSpellRule.isMisspelled returns !speller.isMisspelled (Java polarity)
		return r.isMisspelled(word)
	case *hunspellRuleForVague:
		// Java: hunspell-based isMisspelled is thread-safe; cached rule OK
		return !r.IsMisspelled(word)
	default:
		panic(fmt.Sprintf("Unknown rule type for language %s: %T",
			lang.GetShortCodeWithCountryAndVariant(), rule))
	}
}

// getSpellingCheckRule ports private SpellingCheckRule getSpellingCheckRule(Language lang).
//
// Java:
//
//	JLanguageTool lt = new JLanguageTool(lang);
//	for (Rule r : lt.getAllActiveRules()) {
//	  if (r instanceof HunspellRule || r instanceof MorfologikSpellerRule) {
//	    spellRule = (SpellingCheckRule) r; break;
//	  }
//	}
//
// Go cannot import language packs from this package (import cycle), so the first
// active HunspellRule / MorfologikSpellerRule is resolved the same way Language
// getRelevantRules would register it: known official rule type + dict path for
// the language short code / variant (same resources Java loads).
func getSpellingCheckRule(lang languagetool.Language) any {
	if lang == nil {
		panic("getSpellingCheckRule: nil language")
	}
	code := lang.GetShortCodeWithCountryAndVariant()
	// Java: new JLanguageTool(lang); first HunspellRule or MorfologikSpellerRule
	// among getAllActiveRules (from Language.getRelevantRules).
	// Go: same first rule type + official dict path without importing language packs
	// (import cycle: language packs → spelling).
	r := firstActiveSpellingRuleFromLanguage(lang)
	if r == nil {
		panic(fmt.Sprintf("No spelling rule found for language %s"+
			" - make sure to set 'preferredVariants' so a variant with a spell checker can be selected", code))
	}
	return r
}

// morfologikSpellerRuleForVague is the MorfologikSpellerRule surface VagueSpellChecker needs
// (instanceof MorfologikSpellerRule + getFileName).
type morfologikSpellerRuleForVague struct {
	fileName string
}

// GetFileName ports MorfologikSpellerRule.getFileName().
func (m *morfologikSpellerRuleForVague) GetFileName() string {
	if m == nil {
		return ""
	}
	return m.fileName
}

// hunspellRuleForVague is the HunspellRule surface VagueSpellChecker needs
// (instanceof HunspellRule + isMisspelled). Used for GermanSpellerRule and pure HunspellRule langs.
type hunspellRuleForVague struct {
	isMisspelledFn func(string) bool
}

// IsMisspelled ports HunspellRule.isMisspelled.
func (h *hunspellRuleForVague) IsMisspelled(word string) bool {
	if h == nil || h.isMisspelledFn == nil {
		return false
	}
	return h.isMisspelledFn(word)
}

// nonThreadSafeSpellRule ports VagueSpellChecker.NonThreadSafeSpellRule
// (private nested class extending SpellingCheckRule).
type nonThreadSafeSpellRule struct {
	language languagetool.Language
}

func newNonThreadSafeSpellRule(lang languagetool.Language) *nonThreadSafeSpellRule {
	return &nonThreadSafeSpellRule{language: lang}
}

// getId ports NonThreadSafeSpellRule.getId → "FAKE_FOR_VAGUE_SPELL_CHECKER".
func (r *nonThreadSafeSpellRule) getId() string { return "FAKE_FOR_VAGUE_SPELL_CHECKER" }

// getDescription ports NonThreadSafeSpellRule.getDescription → "internal".
func (r *nonThreadSafeSpellRule) getDescription() string { return "internal" }

// isMisspelled ports NonThreadSafeSpellRule.isMisspelled.
//
// Java polarity (intentional):
//
//	return !speller.isMisspelled(word);
//
// so a true result means the word is valid. isValidWord returns this value directly
// for the Morfologik branch.
func (r *nonThreadSafeSpellRule) isMisspelled(word string) bool {
	if r == nil || r.language == nil {
		panic("NonThreadSafeSpellRule.isMisspelled: nil")
	}
	key := languageCacheKey(r.language)
	vagueLangToRuleMu.Lock()
	dict := vagueLangToDict[key]
	if dict == nil {
		spellingRule := getSpellingCheckRule(r.language)
		ms, ok := spellingRule.(*morfologikSpellerRuleForVague)
		if !ok {
			// Java casts to MorfologikSpellerRule — panic on wrong type
			vagueLangToRuleMu.Unlock()
			panic(fmt.Sprintf("expected MorfologikSpellerRule for NonThreadSafeSpellRule, got %T", spellingRule))
		}
		fileName := ms.GetFileName()
		// Java: Dictionary.read(JLanguageTool.getDataBroker().getFromResourceDirAsUrl(getFileName()))
		disk := discoverVagueResourceDict(fileName)
		if disk == "" {
			vagueLangToRuleMu.Unlock()
			panic(fmt.Sprintf("dictionary not found for %s (Java resource %s)",
				r.language.GetShortCodeWithCountryAndVariant(), fileName))
		}
		d, err := atticmorfo.OpenDictionary(disk)
		if err != nil || d == nil {
			vagueLangToRuleMu.Unlock()
			panic(fmt.Sprintf("Dictionary.read failed for %s: %v", fileName, err))
		}
		dict = d
		vagueLangToDict[key] = dict
	}
	vagueLangToRuleMu.Unlock()

	// Java: Speller speller = new Speller(dict, 1);
	speller := atticmorfo.NewSpeller(dict, 1)
	// Java: return !speller.isMisspelled(word);
	return !speller.IsMisspelled(word)
}

// firstActiveSpellingRuleFromLanguage mirrors the first HunspellRule or
// MorfologikSpellerRule among Language.getRelevantRules / createDefaultSpellingRule.
func firstActiveSpellingRuleFromLanguage(lang languagetool.Language) any {
	if lang == nil {
		return nil
	}
	short := strings.ToLower(lang.GetShortCode())
	full := lang.GetShortCodeWithCountryAndVariant()
	fullLower := strings.ToLower(full)

	switch short {
	case "de":
		// GermanSpellerRule extends CompoundAwareHunspellRule extends HunspellRule.
		// Probe matches GermanSpellerRule.isMisspelled (hard cases + prohibit/ignore + dict).
		return newGermanSpellerVague(fullLower)
	case "en":
		return &morfologikSpellerRuleForVague{fileName: englishMorfologikDictClasspath(fullLower)}
	case "fr":
		// MorfologikFrenchSpellerRule.getFileName → /fr/french.dict
		return &morfologikSpellerRuleForVague{fileName: "/fr/french.dict"}
	case "pt":
		return &morfologikSpellerRuleForVague{fileName: portugueseMorfologikDictClasspath(fullLower)}
	case "es":
		// MorfologikSpanishSpellerRule.getFileName → /es/es-ES.dict
		return &morfologikSpellerRuleForVague{fileName: "/es/es-ES.dict"}
	case "nl":
		return &morfologikSpellerRuleForVague{fileName: "/nl/spelling/nl_NL.dict"}
	case "pl":
		return &morfologikSpellerRuleForVague{fileName: "/pl/hunspell/pl_PL.dict"}
	case "it":
		return &morfologikSpellerRuleForVague{fileName: "/it/hunspell/it_IT.dict"}
	case "ca":
		// MorfologikCatalanSpellerRule.getFileName → /ca/ca-ES_spelling.dict
		return &morfologikSpellerRuleForVague{fileName: "/ca/ca-ES_spelling.dict"}
	case "ru":
		return &morfologikSpellerRuleForVague{fileName: "/ru/hunspell/ru_RU.dict"}
	case "uk":
		return &morfologikSpellerRuleForVague{fileName: "/uk/hunspell/uk_UA.dict"}
	case "da", "sv", "eo", "gl", "km", "is":
		// Java: da/sv/eo/gl/km → HunspellRule; is → HunspellNoSuggestionRule.
		// Both call native hunspell.spell with full .dic+.aff expansion
		// (HunspellRule.isMisspelled ~429–445).
		// Pure-Go FileHunspellDictionary loads stems only (affix rules not expanded).
		// Dic-stem maps as isValidWord would false-negative affix forms (e.g. Danish
		// billetter/hunde/biler). No invent: hard-fail on isMisspelled until an
		// affix-capable pure-Go HunspellDictionary leaf exists (no cgo).
		return pureHunspellBlockedRule(full)
	default:
		// Unknown: no spelling rule surface (Java RuntimeException via getSpellingCheckRule)
		return nil
	}
}

func germanHunspellDictClasspath(fullLower string) string {
	// Java getDictFilenameInResources: /de/hunspell/de_DE.dic
	// Go GermanSpellerRule FilterDict / Morfo uses de_DE.dict (CFSA2).
	switch {
	case strings.Contains(fullLower, "at"):
		return "/de/hunspell/de_AT.dict"
	case strings.Contains(fullLower, "ch"):
		return "/de/hunspell/de_CH.dict"
	default:
		return "/de/hunspell/de_DE.dict"
	}
}

func englishMorfologikDictClasspath(fullLower string) string {
	// Java Morfologik*SpellerRule RESOURCE_FILENAME
	switch {
	case strings.Contains(fullLower, "gb"):
		return "/en/hunspell/en_GB.dict"
	case strings.Contains(fullLower, "en-ca") || strings.Contains(fullLower, "en_ca"):
		return "/en/hunspell/en_CA.dict"
	case strings.Contains(fullLower, "au"):
		return "/en/hunspell/en_AU.dict"
	case strings.Contains(fullLower, "nz"):
		return "/en/hunspell/en_NZ.dict"
	case strings.Contains(fullLower, "za"):
		return "/en/hunspell/en_ZA.dict"
	default:
		return "/en/hunspell/en_US.dict"
	}
}

func portugueseMorfologikDictClasspath(fullLower string) string {
	// Java MorfologikPortugueseSpellerRule: /pt/spelling/ + getDictFilename
	// pt-PT → pt-PT-90; pt-BR → pt-BR
	if strings.Contains(fullLower, "br") {
		return "/pt/spelling/pt-BR.dict"
	}
	return "/pt/spelling/pt-PT-90.dict"
}

// HunspellRuleID ports org.languagetool.rules.spelling.hunspell.HunspellRule.RULE_ID
// (duplicated string — spelling cannot import hunspell subpackage: cycle).
const vagueHunspellRuleID = "HUNSPELL_RULE"

// pureHunspellBlockedMsg documents why VagueSpellChecker cannot serve pure-Hunspell
// languages yet. Prefer panic over dic-stem invent that disagrees with Java polarity.
func pureHunspellBlockedMsg(langFull string) string {
	return fmt.Sprintf(
		"VagueSpellChecker pure-Hunspell language %s blocked: pure-Go HunspellDictionary has no affix expansion; Java HunspellRule.isMisspelled requires hunspell.spell(.dic+.aff); dependency leaf missing (pure Go / no cgo)",
		langFull,
	)
}

// pureHunspellBlockedRule returns a HunspellRule surface (Java rule type) that hard-fails
// on isMisspelled instead of inventing dic-stem polarity. Panic runs outside the
// langToRule mutex (after cache store) so IsValidWord does not deadlock.
func pureHunspellBlockedRule(langFull string) *hunspellRuleForVague {
	msg := pureHunspellBlockedMsg(langFull)
	return &hunspellRuleForVague{
		isMisspelledFn: func(string) bool {
			panic(msg)
		},
	}
}

// GermanSpellerRule.isMisspelled hard-case patterns (Java Pattern.matches → full string).
var (
	vagueReStartWithSpiel = regexp.MustCompile(`^(?:Spielzugs?|Spielzugangs?|Spielzuganges|Spielzugbuchs?|Spielzugbüchern?|Spielzuges|Spielzugverluste?|Spielzugverluste[ns])$`)
	vagueReEndWithSchafte = regexp.MustCompile(`^[A-ZÖÄÜ][a-zöäß-]+schafte$`)
	vagueReSchafPattern   = regexp.MustCompile(`^.{3,}schaf(s|en)?$`)
	vagueReSchafePattern  = regexp.MustCompile(`^(?:Alpenschaf|Berberschaf|Bergschaf|Blauschaf|Brillenschaf|Dachsteinschaf|Deichschaf|Dickhornschaf|Feinwollschaf|Fettschwanzschaf|Fleischschaf|Fuchsschaf|Glücksschaf|Hausschaf|Jungschaf|Karakulschaf|Klonschaf|Merinoschaf|Milchschaf|Mondschaf|Nutzschaf|Rhönschaf|Riesenwildschaf|Schaukelschaf|Schneeschaf|Steinschaf|Steppenschaf|Superschaf|Waldschaf|Weideschaf|Wildschaf|Wollschaf|Zackelschaf|Zuchtschaf|Zwergblauschaf)(s|en)?$`)
	vagueReReplaceSchaf   = regexp.MustCompile(`schaf$`)
	vagueReReplaceSchafs  = regexp.MustCompile(`schafs$`)
	vagueReReplaceSchafen = regexp.MustCompile(`schafen$`)
)

// newGermanSpellerVague builds a HunspellRule-surface probe matching
// GermanSpellerRule.isMisspelled as ported in rules/de (hard cases SCHAF/Spielzug/
// Standart/schafte; isProhibited; ignoreWord; then CFSA2 de_*.dict membership —
// same resource Go GermanSpellerRule / FilterDictIsMisspelled uses).
//
// Import cycle prevents calling rules/de.GermanSpellerRule from this package;
// control flow and resources are the same as that twin.
func newGermanSpellerVague(fullLower string) *hunspellRuleForVague {
	classpath := germanHunspellDictClasspath(fullLower)
	disk := discoverVagueResourceDict(classpath)

	// SpellingCheckRule word lists (ignore/spelling/prohibit) for de — same init lists
	// HunspellRule/SpellingCheckRule load; GermanSpellerRule also loads these.
	base := NewSpellingCheckRule(vagueHunspellRuleID, DescSpelling, "de")
	// Variant-specific spelling file (de-AT / de-CH) — GermanSpellerRule languageSpecificPlainTextDict
	if strings.Contains(fullLower, "at") {
		base.GetLanguageVariantSpellingFileNameFn = func() string { return "de/hunspell/spelling-de-AT.txt" }
	} else if strings.Contains(fullLower, "ch") {
		base.GetLanguageVariantSpellingFileNameFn = func() string { return "de/hunspell/spelling-de-CH.txt" }
	}
	// GermanSpellerRule.init: ignoreWordsWithLength = 1
	base.IgnoreWordsWithLength = 1
	ApplyDefaultSpellingWordLists(base)

	var dict *atticmorfo.Dictionary
	if disk != "" {
		if d, err := atticmorfo.OpenDictionary(disk); err == nil {
			dict = d
		}
	}

	var isMisspelled func(string) bool
	isMisspelled = func(word string) bool {
		// GermanSpellerRule.isMisspelled hard cases (before super.isMisspelled)
		if vagueReSchafPattern.MatchString(word) && !vagueReSchafePattern.MatchString(word) {
			variant := vagueReReplaceSchaf.ReplaceAllString(word, "schaft")
			variant = vagueReReplaceSchafs.ReplaceAllString(variant, "schaft")
			variant = vagueReReplaceSchafen.ReplaceAllString(variant, "schaften")
			if !isMisspelled(variant) {
				return true
			}
		}
		if strings.HasPrefix(word, "Spielzug") && !vagueReStartWithSpiel.MatchString(word) {
			return true
		}
		if strings.HasPrefix(word, "Standart") &&
			word != "Standarte" &&
			word != "Standarten" &&
			!strings.HasPrefix(word, "Standartenträger") &&
			!strings.HasPrefix(word, "Standartenführer") {
			return true
		}
		if strings.HasSuffix(word, "schafte") && vagueReEndWithSchafte.MatchString(word) {
			return true
		}
		// HunspellRule.isMisspelled: isProhibited(cutOffDot) forces true
		if base != nil && base.IsProhibited(cutOffDotVague(word)) {
			return true
		}
		// ignoreWord → not misspelled
		if base != nil && base.IgnoreWord(word) {
			return false
		}
		// Java hunspell null → false (not misspelled); Go FilterDict nil → false
		if dict == nil || word == "" {
			return false
		}
		return dict.IsMisspelled(word)
	}

	return &hunspellRuleForVague{isMisspelledFn: isMisspelled}
}

func cutOffDotVague(s string) string {
	if strings.HasSuffix(s, ".") {
		return s[:len(s)-1]
	}
	return s
}

// discoverVagueResourceDict finds a Java resource-path Morfologik .dict on disk
// (same layout as morfologik.DiscoverLanguageDict + french/portuguese Maven pos-dicts).
func discoverVagueResourceDict(classpath string) string {
	rel := strings.TrimPrefix(classpath, "/")
	if rel == "" || !strings.HasSuffix(rel, ".dict") {
		return ""
	}
	lang, rest, ok := strings.Cut(rel, "/")
	if !ok || lang == "" || rest == "" {
		return ""
	}
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for {
		candidates := []string{
			filepath.Join(dir, "inspiration", "languagetool", "languagetool-language-modules", lang,
				"src", "main", "resources", "org", "languagetool", "resource", lang, rest),
			filepath.Join(dir, "inspiration", "languagetool", "languagetool-language-modules", lang,
				"src", "main", "resources", "org", "languagetool", "resource", rel),
			filepath.Join(dir, "inspiration", "languagetool", "languagetool-core",
				"src", "test", "resources", "org", "languagetool", "resource", rel),
			filepath.Join(dir, "inspiration", "languagetool", "languagetool-core",
				"src", "main", "resources", "org", "languagetool", "resource", rel),
			filepath.Join(dir, "third_party", "languagetool-dicts", "org", "languagetool", "resource", rel),
			filepath.Join(dir, "third_party", "english-pos-dict", "org", "languagetool", "resource", rel),
			filepath.Join(dir, "third_party", "french-pos-dict", "org", "languagetool", "resource", rel),
			filepath.Join(dir, "third_party", "portuguese-pos-dict", "org", "languagetool", "resource", rel),
			filepath.Join(dir, "third_party", "fr", "french.dict"),
		}
		for _, p := range candidates {
			if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
				return p
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}
