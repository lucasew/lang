package it

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
)

// RegisterCoreItalianRules ports Italian.getRelevantRules / createDefaultSpellingRule.
// Java list only — no invent SharedLayout extras (no empty-line / long-paragraph /
// sentence-whitespace / invent IT_UNPAIRED_BRACKETS).
func RegisterCoreItalianRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	wbp := rules.NewWhitespaceBeforePunctuationRule(map[string]string{
		"no_space_before_colon":     "Don't put a space before the colon",
		"no_space_before_semicolon": "Don't put a space before the semicolon",
	})
	lt.AddRuleChecker(wbp.GetID(), rules.AsSentenceCheckerSimple(wbp.Match))

	cw := rules.NewCommaWhitespaceRule(nil)
	lt.AddRuleChecker(cw.GetID(), rules.AsSentenceCheckerSimple(cw.Match))

	dp := rules.NewDoublePunctuationRule(nil)
	lt.AddRuleChecker(dp.GetID(), rules.AsSentenceCheckerSimple(dp.Match))

	ub := NewUnpairedBracketsRule(nil)
	lt.AddTextLevelRuleChecker(ub.GetID(), rules.AsTextLevelChecker(ub.MatchList))

	sp := NewMorfologikItalianSpellerRule()
	if p := morfologik.DiscoverLanguageDict(ItalianSpellerDict); p != "" {
		if WireItalianFilterSpeller(p) {
			sp.IsMisspelled = FilterDictIsMisspelled
		}
	}
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))

	up := rules.NewUppercaseSentenceStartRule(nil, "it")
	lt.AddRuleChecker(up.GetID(), rules.AsSentenceCheckerSimple(func(s *languagetool.AnalyzedSentence) []*rules.RuleMatch {
		return up.MatchList([]*languagetool.AnalyzedSentence{s})
	}))

	// Java ItalianWordRepeatRule only — no WordRepeatBeginning.
	wr := NewItalianWordRepeatRule(map[string]string{"repetition": "Ripetizione di parola"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))

	ws := rules.NewMultipleWhitespaceRule(map[string]string{
		"whitespace_repetition": "Possible typo: you repeated a whitespace",
	})
	lt.AddRuleChecker(ws.GetID(), rules.AsSentenceCheckerSimple(func(s *languagetool.AnalyzedSentence) []*rules.RuleMatch {
		return ws.Match([]*languagetool.AnalyzedSentence{s})
	}))
}
