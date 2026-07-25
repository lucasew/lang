package de

import (
	"embed"
	"os"
	"path/filepath"
	"regexp"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"
	detok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/de"
)

//go:embed data/confusion_sets.txt
var deConfusionSetsFS embed.FS

// GermanConfusionProbabilityRule ports
// org.languagetool.rules.de.GermanConfusionProbabilityRule.
// Java does not override getId() → inherits ConfusionProbabilityRule.RULE_ID ("CONFUSION_RULE").
type GermanConfusionProbabilityRule struct {
	*ngrams.ConfusionProbabilityRule
}

// GermanConfusionFilenames ports ConfusionProbabilityRule.getFilenames() default
// used by GermanConfusionProbabilityRule (no override → confusion_sets.txt).
func GermanConfusionFilenames() []string {
	return []string{"confusion_sets.txt"}
}

// German sentence-level exception patterns (Java GermanConfusionProbabilityRule.SENTENCE_EXCEPTION_PATTERNS).
var deConfusionSentenceExceptions = []*regexp.Regexp{
	regexp.MustCompile(`wir \(`),        // "Hallo, wir (die Dingsbums Gmbh)"
	regexp.MustCompile(`Wie .*?en Sie`), // "Wie heizen Sie das Haus?"
	regexp.MustCompile(`fiel(e|en)? .* (aus|auf|anheim)`),
	regexp.MustCompile(`(regnet|schneit)e? es viel`),                                 // Schneit es viel im Herbst?
	regexp.MustCompile(`(regnet|schneit)e? es (im|jeden) [A-ZÄÖÜ][a-zäöü\-ß]+ viel`), // In Hamburg regnet es im August viel.
	regexp.MustCompile(`viel in [A-ZÄÖÜ][a-zäöü\-ß]+ unterwegs`),                     // "sodass Sie viel in Tirol unterwegs ist"
	regexp.MustCompile(`viel am [A-ZÄÖÜ][a-zäöü\-ß]+`),                               // "sodass Sie viel am Lernen ist"
	regexp.MustCompile(`[Ii]hr .* seht`),                                             // vs "sieht"
	regexp.MustCompile(`fiel .*in die Kategorie`),                                    // vs "viel"
	regexp.MustCompile(`fiel .*nicht leicht`),                                        // vs "viel"
	regexp.MustCompile(`wie fiel das ins Gewicht`),                                   // vs "viel"
}

// GermanConfusionExceptions ports Java GermanConfusionProbabilityRule.EXCEPTIONS
// (all-lowercase comments; match is case-insensitive in CPR).
// See https://github.com/languagetool-org/languagetool/issues/1516
var GermanConfusionExceptions = []string{
	"weist bei interesse auf",
	"weist bei bedarf auf",
	"wir bei der",
	"seht ihr",
	"seht zu, dass",
	"seht zu dass",
	"seht es euch",
	"seht selbst",
	"seht an",
	"viel hin und her",
	"möglichkeit weißt",
	"du doch trotzdem",
	"wir stark ausgelastet sind",
	"wir entwickeln für",
	"nutzen wir Google",
	"vertreiben wir",
	"wir auch nicht",
	", dir bei", // "froh, dir bei deiner Arbeit zu helfen"
	"fiel hinaus",
	"setz dir", // "Setz dir doch bitte einen Termin am Donnerstag"
	"du hast dir",
	"vielen als held",
	"seht gut", // "Ihr seht gut aus"
	"so viel das",
	"wie erinnern sie sich",
	"dürfen wir nicht",
	"kann dich auch",
	"wie schicken wir",
	"wie benutzen sie",
	"wir ja nicht",
	"wie wir oder",
	"eine uhrzeit hatten",
	"damit wir das",
	"damit wir die",
	"damit wir dir",
	"was wird in",
	"warum wird da",
	"da mir der",
	"das wir uns",
	"so wir können",
	"bestellt Botschafter ein",
	"bestellt Botschafterin ein",
	"wie zahlen sie",
	"unser business",
	"journalisten gefiltert worden",
	"für uns filtern",
	"leinwand gezeigte",
	"war sich für nichts", // war sich für nichts zu schade
	"dover corporation",
	"bringt dich ein",
	"bringt dich eine",
	"womit arbeitet",
	"womit arbeiten",
	"ich drei bin", // seit ich drei bin.
	"was wird unser",
	"die wird wieder",
	"damit wir für",
	"wie finden sie",
	"ach die armen",
	"wie stehen da die",      // vs wir
	"wir würden sie",         // vs wird
	"damit wir ihre daten",   // vs wird
	"kannst du doch gerne",   // vs dich
	"wie ist hier der Stand", // vs Sand
	"wie ist der Stand",      // vs Sand
	"dass da Potenzial zu",
	"das auch hergibt",           // vs ergibt
	"hat mich angeschrieben",     // vs abgeschrieben
	"sehe gerade",                // vs siehe gerade
	"hole dich auch ab",          // vs dir
	"würdest du dich vorstellen", // vs dir
	"daten wir über",             // "welche Daten wir über unsere Nutzer erfassen"
	"anders seht",                // falls ihr das anders seht (weht)
	"weit fallendem",             // vs weiht
	"weit fallenden",             // vs weiht
	"weit fallendes",             // vs weiht
	"weit fallende",              // vs weiht
	"weit fallender",             // vs weiht
	"wir ja.",                    // vs wie
	"weißt, wie",                 // vs weist
	"weißt ja, wie",              // vs weist
	"weißt, dass",                // vs weist
	"weißt ja, dass",             // vs weist
	"viel Spass",                 // vs fiel
	"viel Bock",                  // vs fiel
	"so viel",                    // vs fiel
	"viel laenger",               // vs fiel
	"voll viel",                  // vs fiel
	"fasst nichts an",            // vs fast
	"fasst mit an",               // vs fast
	"kann das wer bestätigen",    // vs wär
	"fasst keiner an",            // vs fast
	"fasst keine mehr an",        // vs fast
	"fasst keiner mehr an",       // vs fast
	"Vorgestern und Gestern",     // vs Gesten
}

// Java GermanConfusionProbabilityRule.COMMON_WORD; matcher.matches() → full match.
var deConfusionCommonWord = regexp.MustCompile(`^[\wöäüßÖÄÜ]+$`)

// deConfusionMessages ports MessagesBundle_de.properties statistics_* keys used by CPR.
var deConfusionMessages = map[string]string{
	"statistics_rule_description":   "Mögliche Verwechselungen zwischen ''{0}'' und ''{1}'' erkennen",
	"statistics_suggest_short_desc": "Mögliche Wortverwechselung",
	"statistics_suggest1_new":       "Bitte prüfen Sie, ob ''{0}'' ({1}) hier das richtige Wort ist anstelle von ''{2}'' ({3}).",
	"statistics_suggest2_new":       "Bitte prüfen Sie, ob ''{0}'' ({1}) hier das richtige Wort ist anstelle von ''{2}''.",
	"statistics_suggest3_new":       "Bitte prüfen Sie, ob ''{0}'' hier das richtige Wort ist anstelle von ''{1}''.",
	"statistics_suggest4_new":       "Bitte prüfen Sie, ob ''{0}'' hier das richtige Wort ist anstelle von ''{1}'' ({2}).",
}

var (
	deConfusionPairsOnce sync.Once
	deConfusionPairs     map[string][]*rules.ConfusionPair
)

// loadDEConfusionPairs ports ConfusionProbabilityRule constructor loop:
// for filename in getFilenames(): load "/" + language + "/" + filename via ConfusionSetLoader.
// Prefer official inspiration resource/de; fall back to embedded official copy (identical).
// Fail-closed: missing file or parse error → empty map (no invent pairs).
func loadDEConfusionPairs() map[string][]*rules.ConfusionPair {
	deConfusionPairsOnce.Do(func() {
		merged := map[string][]*rules.ConfusionPair{}
		for _, fn := range GermanConfusionFilenames() {
			part := loadCachedDEConfusionPairs(fn)
			for k, v := range part {
				merged[k] = append(merged[k], v...)
			}
		}
		deConfusionPairs = merged
	})
	return deConfusionPairs
}

func loadCachedDEConfusionPairs(filename string) map[string][]*rules.ConfusionPair {
	if filename == "" {
		return map[string][]*rules.ConfusionPair{}
	}
	// Official /de/<filename> under resource dir (Java DataBroker path).
	if root := DiscoverGermanResourceDir(); root != "" {
		path := filepath.Join(root, filename)
		if m := loadDEConfusionPairsFromFile(path); m != nil {
			return m
		}
	}
	// Embedded official copy (same bytes as inspiration resource/de/confusion_sets.txt).
	f, err := deConfusionSetsFS.Open(filepath.ToSlash(filepath.Join("data", filename)))
	if err != nil {
		return map[string][]*rules.ConfusionPair{}
	}
	defer f.Close()
	m, err := rules.NewConfusionSetLoader(nil).LoadConfusionPairs(f)
	if err != nil || m == nil {
		return map[string][]*rules.ConfusionPair{}
	}
	return m
}

func loadDEConfusionPairsFromFile(path string) map[string][]*rules.ConfusionPair {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	m, err := rules.NewConfusionSetLoader(nil).LoadConfusionPairs(f)
	if err != nil || m == nil {
		return nil
	}
	return m
}

// discoverDEConfusionSet finds official /de/<filename> (for tests / GetFilenames path checks).
func discoverDEConfusionSet(filename string) string {
	if filename == "" {
		return ""
	}
	if root := DiscoverGermanResourceDir(); root != "" {
		cand := filepath.Join(root, filename)
		if st, err := os.Stat(cand); err == nil && !st.IsDir() {
			return cand
		}
	}
	return ""
}

// NewGermanConfusionProbabilityRule ports
// GermanConfusionProbabilityRule(messages, languageModel, language) with grams=3.
// Messages default to MessagesBundle_de statistics_* (Java passes ResourceBundle).
func NewGermanConfusionProbabilityRule(lm ngrams.LanguageModel) *GermanConfusionProbabilityRule {
	return NewGermanConfusionProbabilityRuleGrams(lm, 3)
}

// NewGermanConfusionProbabilityRuleGrams ports the grams constructor:
// super(messages, languageModel, language, grams, EXCEPTIONS, ANTI_PATTERNS)
// + example pair mir/mit.
func NewGermanConfusionProbabilityRuleGrams(lm ngrams.LanguageModel, grams int) *GermanConfusionProbabilityRule {
	return newGermanConfusionProbabilityRule(deConfusionMessages, lm, grams)
}

// NewGermanConfusionProbabilityRuleWithMessages ports the Java constructor when the
// caller supplies an explicit ResourceBundle (e.g. TestTools.getEnglishMessages()).
func NewGermanConfusionProbabilityRuleWithMessages(messages map[string]string, lm ngrams.LanguageModel) *GermanConfusionProbabilityRule {
	merged := map[string]string{}
	for k, v := range deConfusionMessages {
		merged[k] = v
	}
	for k, v := range messages {
		if v != "" {
			merged[k] = v
		}
	}
	return newGermanConfusionProbabilityRule(merged, lm, 3)
}

// NewGermanConfusionProbabilityRuleWithLM is an alias for NewGermanConfusionProbabilityRule
// (historical name kept for in-package callers).
func NewGermanConfusionProbabilityRuleWithLM(lm ngrams.LanguageModel) *GermanConfusionProbabilityRule {
	return NewGermanConfusionProbabilityRule(lm)
}

func newGermanConfusionProbabilityRule(messages map[string]string, lm ngrams.LanguageModel, grams int) *GermanConfusionProbabilityRule {
	base := ngrams.NewConfusionProbabilityRule(lm, grams)
	// Java GermanConfusionProbabilityRule does not override getId() → CONFUSION_RULE.
	// Java: LanguageModelUtils.getGoogleStyleWordTokenizer(language) → GermanWordTokenizer
	wt := detok.NewGermanWordTokenizer()
	base.Tokenize = wt.Tokenize
	if messages == nil {
		messages = deConfusionMessages
	}
	base.Messages = messages
	ngrams.InitConfusionProbabilityMeta(base, messages)
	// Java: super(..., EXCEPTIONS, ANTI_PATTERNS)
	base.Exceptions = append([]string(nil), GermanConfusionExceptions...)
	// Java isCommonWord override: COMMON_WORD.matcher(token).matches()
	base.IsCommonWord = func(token string) bool {
		return deConfusionCommonWord.MatchString(token)
	}
	// Java isException override: SENTENCE_EXCEPTION_PATTERNS
	base.IsException = deConfusionIsException
	// Java: ANTI_PATTERNS → DisambiguationPatternRule immunization
	base.IsCoveredByAntiPattern = deConfusionIsCoveredByAntiPattern
	// Java CPR ctor: load getFilenames() → /de/confusion_sets.txt
	base.SetWordToPairs(loadDEConfusionPairs())
	// Java: mir → mit
	base.AddExamplePair(
		rules.Wrong("Während Sie das Ganze <marker>mir</marker> einem Holzlöffel rühren…"),
		rules.Fixed("Während Sie das Ganze <marker>mit</marker> einem Holzlöffel rühren…"),
	)
	return &GermanConfusionProbabilityRule{ConfusionProbabilityRule: base}
}

// deConfusionIsException ports GermanConfusionProbabilityRule.isException.
func deConfusionIsException(sentenceText string, startPos, endPos int) bool {
	for _, p := range deConfusionSentenceExceptions {
		if p.FindStringIndex(sentenceText) != nil {
			return true
		}
	}
	return false
}

// Match ports ConfusionProbabilityRule.match through the German wrapper.
func (r *GermanConfusionProbabilityRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || r.ConfusionProbabilityRule == nil {
		return nil
	}
	return r.ConfusionProbabilityRule.Match(sentence)
}

// GetID ports inherited ConfusionProbabilityRule.getId() → CONFUSION_RULE.
func (r *GermanConfusionProbabilityRule) GetID() string {
	if r == nil || r.ConfusionProbabilityRule == nil {
		return ngrams.ConfusionRuleID
	}
	return r.ConfusionProbabilityRule.GetID()
}

// GetFilenames ports ConfusionProbabilityRule.getFilenames for German (default list).
func (r *GermanConfusionProbabilityRule) GetFilenames() []string {
	return GermanConfusionFilenames()
}
