package tl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCoreTagalogRules ports Tagalog.getRelevantRules / createDefaultSpellingRule.
// Java: CommaWhitespace, DoublePunctuation, UnpairedBrackets, Uppercase, MultipleWhitespace,
// Morfologik only — no invent SharedLayout extras.
func RegisterCoreTagalogRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	cw := rules.NewCommaWhitespaceRule(nil)
	lt.AddRuleChecker(cw.GetID(), rules.AsSentenceCheckerSimple(cw.Match))

	dp := rules.NewDoublePunctuationRule(nil)
	lt.AddRuleChecker(dp.GetID(), rules.AsSentenceCheckerSimple(dp.Match))

	lt.AddRuleChecker("UNPAIRED_BRACKETS", languagetool.SimpleUnpairedBracketsChecker())

	up := rules.NewUppercaseSentenceStartRule(nil, "tl")
	lt.AddRuleChecker(up.GetID(), rules.AsSentenceCheckerSimple(func(s *languagetool.AnalyzedSentence) []*rules.RuleMatch {
		return up.MatchList([]*languagetool.AnalyzedSentence{s})
	}))

	ws := rules.NewMultipleWhitespaceRule(map[string]string{
		"whitespace_repetition": "Possible typo: you repeated a whitespace",
	})
	lt.AddRuleChecker(ws.GetID(), rules.AsSentenceCheckerSimple(func(s *languagetool.AnalyzedSentence) []*rules.RuleMatch {
		return ws.Match([]*languagetool.AnalyzedSentence{s})
	}))

	sp := NewMorfologikTagalogSpellerRule()
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))
}
