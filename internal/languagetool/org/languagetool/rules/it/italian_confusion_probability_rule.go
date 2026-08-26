package it

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"
)

//go:embed data/confusion_sets.txt
var itConfusionSetsFS embed.FS

// ItalianConfusionProbabilityRule ports
// org.languagetool.rules.it.ItalianConfusionProbabilityRule.
// Java does not override getId() → inherits ConfusionProbabilityRule.RULE_ID ("CONFUSION_RULE").
// Java does not override getFilenames() → default confusion_sets.txt.
// Java example pair is commented TODO only — no AddExamplePair.
// Java does not add exceptions / anti-patterns.
type ItalianConfusionProbabilityRule struct {
	*ngrams.ConfusionProbabilityRule
}

// ItalianConfusionFilenames ports ConfusionProbabilityRule.getFilenames() default
// used by ItalianConfusionProbabilityRule (no override → confusion_sets.txt).
func ItalianConfusionFilenames() []string {
	return []string{"confusion_sets.txt"}
}

var (
	itConfusionPairsOnce sync.Once
	itConfusionPairs     map[string][]*rules.ConfusionPair
)

// loadITConfusionPairs ports ConfusionProbabilityRule constructor loop:
// for filename in getFilenames(): load "/" + language + "/" + filename via ConfusionSetLoader.
// Official IT resource is header-only (no data rows) → empty map after parse is correct.
// Fail-closed: missing file or parse error → empty map (no invent pairs).
func loadITConfusionPairs() map[string][]*rules.ConfusionPair {
	itConfusionPairsOnce.Do(func() {
		merged := map[string][]*rules.ConfusionPair{}
		for _, fn := range ItalianConfusionFilenames() {
			part := loadCachedITConfusionPairs(fn)
			for k, v := range part {
				merged[k] = append(merged[k], v...)
			}
		}
		itConfusionPairs = merged
	})
	return itConfusionPairs
}

func loadCachedITConfusionPairs(filename string) map[string][]*rules.ConfusionPair {
	if filename == "" {
		return map[string][]*rules.ConfusionPair{}
	}
	// Embedded official copy (same bytes as inspiration resource/it/confusion_sets.txt).
	f, err := itConfusionSetsFS.Open("data/" + filename)
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

// NewItalianConfusionProbabilityRule ports
// ItalianConfusionProbabilityRule(messages, languageModel, language):
// super(messages, languageModel, language) → default grams=3 only (no example pair).
func NewItalianConfusionProbabilityRule(lm ngrams.LanguageModel) *ItalianConfusionProbabilityRule {
	return NewItalianConfusionProbabilityRuleGrams(lm, 3)
}

// NewItalianConfusionProbabilityRuleGrams ports the grams constructor path used by
// the 3-arg super (default grams). Java Italian subclass only calls super — no examples.
func NewItalianConfusionProbabilityRuleGrams(lm ngrams.LanguageModel, grams int) *ItalianConfusionProbabilityRule {
	base := ngrams.NewConfusionProbabilityRule(lm, grams)
	// Java ItalianConfusionProbabilityRule does not override getId() → CONFUSION_RULE.
	// Java CPR ctor: load getFilenames() → /it/confusion_sets.txt (header-only upstream).
	base.SetWordToPairs(loadITConfusionPairs())
	// Java example pair is commented TODO — do not invent example pairs.
	return &ItalianConfusionProbabilityRule{ConfusionProbabilityRule: base}
}

// Match ports ConfusionProbabilityRule.match through the Italian wrapper.
func (r *ItalianConfusionProbabilityRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || r.ConfusionProbabilityRule == nil {
		return nil
	}
	return r.ConfusionProbabilityRule.Match(sentence)
}

// GetID ports inherited ConfusionProbabilityRule.getId() → CONFUSION_RULE.
func (r *ItalianConfusionProbabilityRule) GetID() string {
	if r == nil || r.ConfusionProbabilityRule == nil {
		return ngrams.ConfusionRuleID
	}
	return r.ConfusionProbabilityRule.GetID()
}

// GetFilenames ports ConfusionProbabilityRule.getFilenames for Italian (default list).
func (r *ItalianConfusionProbabilityRule) GetFilenames() []string {
	return ItalianConfusionFilenames()
}

// GetIncorrectExamples ports Rule.getIncorrectExamples via base CPR.
func (r *ItalianConfusionProbabilityRule) GetIncorrectExamples() []rules.IncorrectExample {
	if r == nil || r.ConfusionProbabilityRule == nil {
		return nil
	}
	return r.ConfusionProbabilityRule.GetIncorrectExamples()
}

// GetCorrectExamples ports Rule.getCorrectExamples via base CPR.
func (r *ItalianConfusionProbabilityRule) GetCorrectExamples() []rules.CorrectExample {
	if r == nil || r.ConfusionProbabilityRule == nil {
		return nil
	}
	return r.ConfusionProbabilityRule.GetCorrectExamples()
}
