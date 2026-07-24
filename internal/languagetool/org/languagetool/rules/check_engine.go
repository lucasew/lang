package rules

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// SentenceRule matches one analyzed sentence.
type SentenceRule interface {
	Match(sentence *languagetool.AnalyzedSentence) ([]*RuleMatch, error)
}

// SentenceRuleFunc adapts a function to SentenceRule.
type SentenceRuleFunc func(sentence *languagetool.AnalyzedSentence) ([]*RuleMatch, error)

func (f SentenceRuleFunc) Match(sentence *languagetool.AnalyzedSentence) ([]*RuleMatch, error) {
	return f(sentence)
}

// CheckEngine runs sentence-level rules and optional match filters.
// Lives in rules to avoid languagetool ⇄ rules import cycle.
type CheckEngine struct {
	LanguageCode string
	Rules        []SentenceRule
	Filters      []RuleMatchFilter
	Listener     RuleMatchListener
}

func NewCheckEngine(languageCode string) *CheckEngine {
	return &CheckEngine{LanguageCode: languageCode}
}

// CheckText analyzes text and runs all rules; returns filtered matches.
func (e *CheckEngine) CheckText(text string) ([]*RuleMatch, error) {
	lt := languagetool.NewJLanguageTool(e.LanguageCode)
	sentences := lt.Analyze(text)
	return e.CheckSentences(sentences, text)
}

// CheckSentences runs rules over pre-analyzed sentences.
func (e *CheckEngine) CheckSentences(sentences []*languagetool.AnalyzedSentence, text string) ([]*RuleMatch, error) {
	var all []*RuleMatch
	for _, s := range sentences {
		if s == nil {
			continue
		}
		for _, r := range e.Rules {
			if r == nil {
				continue
			}
			ms, err := r.Match(s)
			if err != nil {
				return all, err
			}
			for _, m := range ms {
				NotifyListeners(m, e.Listener)
				all = append(all, m)
			}
		}
	}
	for _, f := range e.Filters {
		if f == nil {
			continue
		}
		all = f.Filter(all, text)
	}
	return all, nil
}
