package spelling

import (
	"fmt"
	"os"
	"path/filepath"
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
		// Dict resource for isMisspelled / morfologik twin: /de/hunspell/de_XX.dict
		// (Go GermanSpellerRule uses the same CFSA2 dict as FilterDictIsMisspelled).
		dictClasspath := germanHunspellDictClasspath(fullLower)
		return newHunspellVagueFromMorfologikDict(dictClasspath)
	case "en":
		return &morfologikSpellerRuleForVague{fileName: englishMorfologikDictClasspath(fullLower)}
	case "fr":
		// MorfologikFrenchSpellerRule.getFileName → /fr/french.dict
		return &morfologikSpellerRuleForVague{fileName: "/fr/french.dict"}
	case "pt":
		return &morfologikSpellerRuleForVague{fileName: portugueseMorfologikDictClasspath(fullLower)}
	case "es":
		return &morfologikSpellerRuleForVague{fileName: "/es/es.dict"}
	case "nl":
		return &morfologikSpellerRuleForVague{fileName: "/nl/spelling/nl_NL.dict"}
	case "pl":
		return &morfologikSpellerRuleForVague{fileName: "/pl/hunspell/pl_PL.dict"}
	case "it":
		return &morfologikSpellerRuleForVague{fileName: "/it/hunspell/it_IT.dict"}
	case "ca":
		return &morfologikSpellerRuleForVague{fileName: "/ca/ca.dict"}
	case "ru":
		return &morfologikSpellerRuleForVague{fileName: "/ru/hunspell/ru_RU.dict"}
	case "uk":
		return &morfologikSpellerRuleForVague{fileName: "/uk/hunspell/uk_UA.dict"}
	case "da", "sv", "eo", "gl", "km", "is":
		// HunspellRule / HunspellNoSuggestionRule languages (is.dic etc.)
		return newHunspellVagueFromDicClasspath(hunspellDicClasspath(short, fullLower))
	default:
		// Unknown: no spelling rule surface (Java RuntimeException via getSpellingCheckRule)
		return nil
	}
}

func germanHunspellDictClasspath(fullLower string) string {
	// Java getDictFilenameInResources: /de/hunspell/de_DE.dic (+ morfo de_DE.dict for suggestions)
	// Go GermanSpellerRule FilterDict uses de_DE.dict (CFSA2). Prefer that for isMisspelled twin.
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

func hunspellDicClasspath(short, fullLower string) string {
	// Java HunspellRule.getDictFilenameInResources: /{short}/hunspell/{langCountry}.dic
	switch short {
	case "da":
		return "/da/hunspell/da_DK.dic"
	case "sv":
		return "/sv/hunspell/sv_SE.dic"
	case "eo":
		return "/eo/hunspell/eo.dic"
	case "gl":
		return "/gl/hunspell/gl_ES.dic"
	case "km":
		return "/km/hunspell/km_KH.dic"
	case "is":
		return "/is/hunspell/is.dic"
	default:
		_ = fullLower
		return "/" + short + "/hunspell/" + short + ".dic"
	}
}

// newHunspellVagueFromMorfologikDict builds a HunspellRule-like probe using a CFSA2
// .dict (same resource Go GermanSpellerRule / FilterDictIsMisspelled uses).
func newHunspellVagueFromMorfologikDict(classpath string) *hunspellRuleForVague {
	disk := discoverVagueResourceDict(classpath)
	if disk == "" {
		// Java hunspell null → isMisspelled false (not misspelled); keep fail-closed soft-false
		return &hunspellRuleForVague{isMisspelledFn: func(string) bool { return false }}
	}
	d, err := atticmorfo.OpenDictionary(disk)
	if err != nil || d == nil {
		return &hunspellRuleForVague{isMisspelledFn: func(string) bool { return false }}
	}
	// Capture dict; IsMisspelled matches Dictionary.IsMisspelled (Speller metadata gates)
	return &hunspellRuleForVague{
		isMisspelledFn: func(word string) bool {
			if word == "" {
				return false
			}
			return d.IsMisspelled(word)
		},
	}
}

// newHunspellVagueFromDicClasspath loads a .dic word list (pure-Go FileHunspell twin, no affix).
// Used for languages where Java registers HunspellRule with .dic resources.
func newHunspellVagueFromDicClasspath(classpath string) *hunspellRuleForVague {
	disk := discoverVagueResourceDic(classpath)
	if disk == "" {
		return &hunspellRuleForVague{isMisspelledFn: func(string) bool { return false }}
	}
	// Load words from .dic (word before '/', skip header count)
	words, err := loadHunspellDicWords(disk)
	if err != nil {
		return &hunspellRuleForVague{isMisspelledFn: func(string) bool { return false }}
	}
	return &hunspellRuleForVague{
		isMisspelledFn: func(word string) bool {
			if word == "" || word == "--" {
				return false
			}
			_, ok := words[word]
			return !ok
		},
	}
}

func loadHunspellDicWords(path string) (map[string]struct{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	out := map[string]struct{}{}
	lines := strings.Split(string(data), "\n")
	start := 0
	if len(lines) > 0 {
		// optional count header
		first := strings.TrimSpace(lines[0])
		if first != "" {
			allDigit := true
			for _, r := range first {
				if r < '0' || r > '9' {
					allDigit = false
					break
				}
			}
			if allDigit {
				start = 1
			}
		}
	}
	for _, line := range lines[start:] {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		w := line
		if i := strings.IndexByte(line, '/'); i >= 0 {
			w = line[:i]
		}
		if w != "" {
			out[w] = struct{}{}
		}
	}
	return out, nil
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

// discoverVagueResourceDic finds a Java resource-path Hunspell .dic on disk.
func discoverVagueResourceDic(classpath string) string {
	rel := strings.TrimPrefix(classpath, "/")
	if rel == "" || !strings.HasSuffix(rel, ".dic") {
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
			filepath.Join(dir, "third_party", "languagetool-dicts", "org", "languagetool", "resource", rel),
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
