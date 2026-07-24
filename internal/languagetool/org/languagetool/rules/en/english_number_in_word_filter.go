package en

import (
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// EnglishNumberInWordFilter ports org.languagetool.rules.en.EnglishNumberInWordFilter
// (extends AbstractNumberInWordFilter; static MorfologikAmericanSpellerRule for
// isMisspelled / getSpellingSuggestions — same as Java).
type EnglishNumberInWordFilter struct {
	inner *rules.NumberInWordFilter
}

// Java: private static MorfologikAmericanSpellerRule englishSpellerRule;
var (
	englishNumberInWordSpellerOnce sync.Once
	englishNumberInWordSpeller     *MorfologikAmericanSpellerRule
)

// englishSpellerRule ports the static englishSpellerRule lazy init in the Java constructor.
func englishSpellerRule() *MorfologikAmericanSpellerRule {
	englishNumberInWordSpellerOnce.Do(func() {
		// Java: new MorfologikAmericanSpellerRule(messages, AmericanEnglish.getInstance())
		englishNumberInWordSpeller = NewMorfologikAmericanSpellerRule()
	})
	return englishNumberInWordSpeller
}

// NewEnglishNumberInWordFilter ports EnglishNumberInWordFilter()
// super(AmericanEnglish.getInstance()) + static MorfologikAmericanSpellerRule.
func NewEnglishNumberInWordFilter() *EnglishNumberInWordFilter {
	_ = englishSpellerRule()
	return &EnglishNumberInWordFilter{
		inner: &rules.NumberInWordFilter{
			// Java: englishSpellerRule.isMisspelled / getSpellingSuggestions
			IsMisspelled:   englishNumberInWordIsMisspelled,
			GetSuggestions: englishNumberInWordGetSuggestions,
		},
	}
}

// englishNumberInWordIsMisspelled ports EnglishNumberInWordFilter.isMisspelled.
func englishNumberInWordIsMisspelled(word string) bool {
	r := englishSpellerRule()
	if r == nil || r.IsMisspelled == nil {
		// Java always constructs the speller; empty-dict isMisspelledWord is fail-closed false.
		return false
	}
	return r.IsMisspelled(word)
}

// englishNumberInWordGetSuggestions ports EnglishNumberInWordFilter.getSuggestions
// → MorfologikAmericanSpellerRule.getSpellingSuggestions.
func englishNumberInWordGetSuggestions(word string) []string {
	r := englishSpellerRule()
	if r == nil {
		return nil
	}
	return r.GetSpellingSuggestions(word)
}

// Suggestions ports AbstractNumberInWordFilter.acceptRuleMatch candidate list.
func (f *EnglishNumberInWordFilter) Suggestions(word string) []string {
	if f == nil || f.inner == nil {
		return nil
	}
	return f.inner.Suggestions(word)
}

// AcceptRuleMatch ports AbstractNumberInWordFilter.acceptRuleMatch.
func (f *EnglishNumberInWordFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, patternTokenPos int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	if f == nil || match == nil || f.inner == nil {
		return nil
	}
	return f.inner.AcceptRuleMatch(match, arguments, patternTokenPos, patternTokens, tokenPositions)
}
