package lt

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCoreLithuanianRules ports Lithuanian.getRelevantRules / createDefaultSpellingRule.
// Java: CommaWhitespace, DoublePunctuation, UnpairedBrackets, Morfologik, Uppercase,
// MultipleWhitespace only — no invent SharedLayout extras.
func RegisterCoreLithuanianRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	cw := rules.NewCommaWhitespaceRule(nil)
	lt.AddRuleChecker(cw.GetID(), rules.AsSentenceCheckerSimple(cw.Match))

	dp := rules.NewDoublePunctuationRule(nil)
	lt.AddRuleChecker(dp.GetID(), rules.AsSentenceCheckerSimple(dp.Match))

	lt.AddRuleChecker("UNPAIRED_BRACKETS", languagetool.SimpleUnpairedBracketsChecker())

	sp := NewMorfologikLithuanianSpellerRule()
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))

	up := rules.NewUppercaseSentenceStartRule(nil, "lt")
	lt.AddRuleChecker(up.GetID(), rules.AsSentenceCheckerSimple(func(s *languagetool.AnalyzedSentence) []*rules.RuleMatch {
		return up.MatchList([]*languagetool.AnalyzedSentence{s})
	}))

	ws := rules.NewMultipleWhitespaceRule(map[string]string{
		"whitespace_repetition": "Possible typo: you repeated a whitespace",
	})
	lt.AddRuleChecker(ws.GetID(), rules.AsSentenceCheckerSimple(func(s *languagetool.AnalyzedSentence) []*rules.RuleMatch {
		return ws.Match([]*languagetool.AnalyzedSentence{s})
	}))
}
