package ar

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"
)

//go:embed data/confusion_sets.txt
var arConfusionSetsFS embed.FS

// ArabicConfusionProbabilityRule ports
// org.languagetool.rules.ar.ArabicConfusionProbabilityRule.
// Java does not override getId() → inherits ConfusionProbabilityRule.RULE_ID ("CONFUSION_RULE").
// Java does not override getFilenames() → default confusion_sets.txt.
// Java does not add exceptions / anti-patterns.
type ArabicConfusionProbabilityRule struct {
	*ngrams.ConfusionProbabilityRule
}

// ArabicConfusionFilenames ports ConfusionProbabilityRule.getFilenames() default
// used by ArabicConfusionProbabilityRule (no override → confusion_sets.txt).
func ArabicConfusionFilenames() []string {
	return []string{"confusion_sets.txt"}
}

var (
	arConfusionPairsOnce sync.Once
	arConfusionPairs     map[string][]*rules.ConfusionPair
)

// loadARConfusionPairs ports ConfusionProbabilityRule constructor loop:
// for filename in getFilenames(): load "/" + language + "/" + filename via ConfusionSetLoader.
// Official AR resource is header-only (no data rows) → empty map after parse is correct.
// Fail-closed: missing file or parse error → empty map (no invent pairs).
func loadARConfusionPairs() map[string][]*rules.ConfusionPair {
	arConfusionPairsOnce.Do(func() {
		merged := map[string][]*rules.ConfusionPair{}
		for _, fn := range ArabicConfusionFilenames() {
			part := loadCachedARConfusionPairs(fn)
			for k, v := range part {
				merged[k] = append(merged[k], v...)
			}
		}
		arConfusionPairs = merged
	})
	return arConfusionPairs
}

func loadCachedARConfusionPairs(filename string) map[string][]*rules.ConfusionPair {
	if filename == "" {
		return map[string][]*rules.ConfusionPair{}
	}
	// Embedded official copy (same bytes as inspiration resource/ar/confusion_sets.txt).
	f, err := arConfusionSetsFS.Open("data/" + filename)
	if err != nil {
		return map[string][]*rules.ConfusionPair{}
	}
	defer f.Close()
	// Java: new ConfusionSetLoader(language) — ShortDescriptionProvider optional; nil = no invent defs.
	m, err := rules.NewConfusionSetLoader(nil).LoadConfusionPairs(f)
	if err != nil || m == nil {
		return map[string][]*rules.ConfusionPair{}
	}
	return m
}

// NewArabicConfusionProbabilityRule ports
// ArabicConfusionProbabilityRule(messages, languageModel, language):
// super(messages, languageModel, language) → default grams=3, then example pair only.
func NewArabicConfusionProbabilityRule(lm ngrams.LanguageModel) *ArabicConfusionProbabilityRule {
	return NewArabicConfusionProbabilityRuleGrams(lm, 3)
}

// NewArabicConfusionProbabilityRuleGrams ports the grams constructor path used by
// the 3-arg super (default grams) plus example pair.
func NewArabicConfusionProbabilityRuleGrams(lm ngrams.LanguageModel, grams int) *ArabicConfusionProbabilityRule {
	base := ngrams.NewConfusionProbabilityRule(lm, grams)
	// Java ArabicConfusionProbabilityRule does not override getId() → CONFUSION_RULE.
	// Java CPR ctor: load getFilenames() → /ar/confusion_sets.txt (header-only upstream).
	base.SetWordToPairs(loadARConfusionPairs())
	// Java demo (trailing unclosed <marker> kept as upstream)
	base.AddExamplePair(
		rules.Wrong("إن بعض <marker>الضن</marker> إثم.<marker>"),
		rules.Fixed("إن بعض <marker>الظن</marker> إثم.<marker>"),
	)
	return &ArabicConfusionProbabilityRule{ConfusionProbabilityRule: base}
}

// Match ports ConfusionProbabilityRule.match through the Arabic wrapper.
func (r *ArabicConfusionProbabilityRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || r.ConfusionProbabilityRule == nil {
		return nil
	}
	return r.ConfusionProbabilityRule.Match(sentence)
}

// GetID ports inherited ConfusionProbabilityRule.getId() → CONFUSION_RULE.
func (r *ArabicConfusionProbabilityRule) GetID() string {
	if r == nil || r.ConfusionProbabilityRule == nil {
		return ngrams.ConfusionRuleID
	}
	return r.ConfusionProbabilityRule.GetID()
}

// GetFilenames ports ConfusionProbabilityRule.getFilenames for Arabic (default list).
func (r *ArabicConfusionProbabilityRule) GetFilenames() []string {
	return ArabicConfusionFilenames()
}

// GetIncorrectExamples ports Rule.getIncorrectExamples via base CPR.
func (r *ArabicConfusionProbabilityRule) GetIncorrectExamples() []rules.IncorrectExample {
	if r == nil || r.ConfusionProbabilityRule == nil {
		return nil
	}
	return r.ConfusionProbabilityRule.GetIncorrectExamples()
}

// GetCorrectExamples ports Rule.getCorrectExamples via base CPR.
func (r *ArabicConfusionProbabilityRule) GetCorrectExamples() []rules.CorrectExample {
	if r == nil || r.ConfusionProbabilityRule == nil {
		return nil
	}
	return r.ConfusionProbabilityRule.GetCorrectExamples()
}
