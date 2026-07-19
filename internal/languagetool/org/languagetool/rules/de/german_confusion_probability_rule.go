package de

import (
	"embed"
	"regexp"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"
)

//go:embed data/confusion_sets.txt
var deConfusionSetsFS embed.FS

// German sentence-level exception patterns (Java GermanConfusionProbabilityRule.SENTENCE_EXCEPTION_PATTERNS).
var deConfusionSentenceExceptions = []*regexp.Regexp{
	regexp.MustCompile(`wir \(`),
	regexp.MustCompile(`Wie .*?en Sie`),
	regexp.MustCompile(`fiel(e|en)? .* (aus|auf|anheim)`),
	regexp.MustCompile(`(regnet|schneit)e? es viel`),
	regexp.MustCompile(`(regnet|schneit)e? es (im|jeden) [A-ZÄÖÜ][a-zäöü\-ß]+ viel`),
	regexp.MustCompile(`viel in [A-ZÄÖÜ][a-zäöü\-ß]+ unterwegs`),
	regexp.MustCompile(`viel am [A-ZÄÖÜ][a-zäöü\-ß]+`),
	regexp.MustCompile(`[Ii]hr .* seht`),
	regexp.MustCompile(`fiel .*in die Kategorie`),
	regexp.MustCompile(`fiel .*nicht leicht`),
	regexp.MustCompile(`wie fiel das ins Gewicht`),
}

// Java GermanConfusionProbabilityRule.EXCEPTIONS (all-lowercase; match is case-insensitive).
var deConfusionExceptions = []string{
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
	", dir bei",
	"fiel hinaus",
	"setz dir",
	"du hast dir",
	"vielen als held",
	"seht gut",
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
	"war sich für nichts",
	"dover corporation",
	"bringt dich ein",
	"bringt dich eine",
	"womit arbeitet",
	"womit arbeiten",
	"ich drei bin",
	"was wird unser",
	"die wird wieder",
	"damit wir für",
	"wie finden sie",
	"ach die armen",
	"wie stehen da die",
	"wir würden sie",
	"damit wir ihre daten",
	"kannst du doch gerne",
	"wie ist hier der Stand",
	"wie ist der Stand",
	"dass da Potenzial zu",
	"das auch hergibt",
	"hat mich angeschrieben",
	"sehe gerade",
	"hole dich auch ab",
	"würdest du dich vorstellen",
	"daten wir über",
	"anders seht",
	"weit fallendem",
	"weit fallenden",
	"weit fallendes",
	"weit fallende",
	"weit fallender",
	"wir ja.",
	"weißt, wie",
	"weißt ja, wie",
	"weißt, dass",
	"weißt ja, dass",
	"viel Spass",
	"viel Bock",
	"so viel",
	"viel laenger",
	"voll viel",
	"fasst nichts an",
	"fasst mit an",
	"kann das wer bestätigen",
	"fasst keiner an",
	"fasst keine mehr an",
	"fasst keiner mehr an",
	"Vorgestern und Gestern",
}

var deConfusionCommonWord = regexp.MustCompile(`^[\wöäüßÖÄÜ]+$`)

var (
	deConfusionPairsOnce sync.Once
	deConfusionPairs     map[string][]*rules.ConfusionPair
)

func loadDEConfusionPairs() map[string][]*rules.ConfusionPair {
	deConfusionPairsOnce.Do(func() {
		f, err := deConfusionSetsFS.Open("data/confusion_sets.txt")
		if err != nil {
			deConfusionPairs = map[string][]*rules.ConfusionPair{}
			return
		}
		defer f.Close()
		m, err := rules.NewConfusionSetLoader(nil).LoadConfusionPairs(f)
		if err != nil || m == nil {
			deConfusionPairs = map[string][]*rules.ConfusionPair{}
			return
		}
		deConfusionPairs = m
	})
	return deConfusionPairs
}

// GermanConfusionProbabilityRule ports org.languagetool.rules.de.GermanConfusionProbabilityRule.
// Full Match needs a LanguageModel; without LM Match is empty (fail-closed, not invented).
type GermanConfusionProbabilityRule struct {
	*ngrams.ConfusionProbabilityRule
}

// deConfusionMessages ports MessagesBundle_de.properties statistics_* keys used by CPR.
var deConfusionMessages = map[string]string{
	"statistics_rule_description":   "Mögliche Verwechselungen zwischen ''{0}'' und ''{1}'' erkennen",
	"statistics_suggest_short_desc": "Mögliche Wortverwechselung",
	"statistics_suggest1_new":       "Bitte prüfen Sie, ob ''{0}'' ({1}) hier das richtige Wort ist anstelle von ''{2}'' ({3}).",
	"statistics_suggest2_new":       "Bitte prüfen Sie, ob ''{0}'' ({1}) hier das richtige Wort ist anstelle von ''{2}''.",
	"statistics_suggest3_new":       "Bitte prüfen Sie, ob ''{0}'' hier das richtige Wort ist anstelle von ''{1}''.",
	"statistics_suggest4_new":       "Bitte prüfen Sie, ob ''{0}'' hier das richtige Wort ist anstelle von ''{1}'' ({2}).",
}

// NewGermanConfusionProbabilityRule builds the rule with nil LM (safe default).
func NewGermanConfusionProbabilityRule(messages map[string]string) *GermanConfusionProbabilityRule {
	r := NewGermanConfusionProbabilityRuleWithLM(nil)
	if r != nil && r.ConfusionProbabilityRule != nil {
		// Prefer caller bundle, fall back to DE defaults for missing keys.
		merged := map[string]string{}
		for k, v := range deConfusionMessages {
			merged[k] = v
		}
		for k, v := range messages {
			if v != "" {
				merged[k] = v
			}
		}
		r.Messages = merged
		// Java ConfusionProbabilityRule: TYPOS + NonConformance.
		ngrams.InitConfusionProbabilityMeta(r.ConfusionProbabilityRule, merged)
	}
	return r
}

// NewGermanConfusionProbabilityRuleWithLM ports the Java constructor (grams=3).
func NewGermanConfusionProbabilityRuleWithLM(lm ngrams.LanguageModel) *GermanConfusionProbabilityRule {
	base := ngrams.NewConfusionProbabilityRule(lm, 3)
	base.RuleIDOverride = "DE_CONFUSION_RULE"
	base.Messages = deConfusionMessages
	ngrams.InitConfusionProbabilityMeta(base, deConfusionMessages)
	base.Exceptions = append([]string{}, deConfusionExceptions...)
	base.IsCommonWord = func(token string) bool {
		return deConfusionCommonWord.MatchString(token)
	}
	base.IsException = func(sentenceText string, startPos, endPos int) bool {
		for _, p := range deConfusionSentenceExceptions {
			if p.FindStringIndex(sentenceText) != nil {
				return true
			}
		}
		return false
	}
	base.SetWordToPairs(loadDEConfusionPairs())
	// Java: ANTI_PATTERNS → DisambiguationPatternRule immunization (getSentenceWithImmunization).
	base.IsCoveredByAntiPattern = deConfusionIsCoveredByAntiPattern
	return &GermanConfusionProbabilityRule{ConfusionProbabilityRule: base}
}

func (r *GermanConfusionProbabilityRule) GetID() string {
	if r == nil || r.ConfusionProbabilityRule == nil {
		return "DE_CONFUSION_RULE"
	}
	return r.ConfusionProbabilityRule.GetID()
}

func (r *GermanConfusionProbabilityRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || r.ConfusionProbabilityRule == nil {
		return nil
	}
	return r.ConfusionProbabilityRule.Match(sentence)
}
