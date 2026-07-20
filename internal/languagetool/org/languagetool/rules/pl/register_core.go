package pl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
)

// RegisterCorePolishRules ports Polish.getRelevantRules / createDefaultSpellingRule.
// Java list only — no invent SharedLayout extras (no double-punct / empty-line /
// long-paragraph / whitespace-before-punct / invent PL_SENTENCE_WHITESPACE).
func RegisterCorePolishRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	lt.PriorityForId = language.PolishPriorityForId

	cw := rules.NewCommaWhitespaceRule(nil)
	lt.AddRuleChecker(cw.GetID(), rules.AsSentenceCheckerSimple(cw.Match))

	up := rules.NewUppercaseSentenceStartRule(nil, "pl")
	lt.AddRuleChecker(up.GetID(), rules.AsSentenceCheckerSimple(func(s *languagetool.AnalyzedSentence) []*rules.RuleMatch {
		return up.MatchList([]*languagetool.AnalyzedSentence{s})
	}))

	// Java WordRepeatRule then later PolishWordRepeatRule (Advanced).
	gwr := rules.NewWordRepeatRule(map[string]string{"repetition": "Powtórzenie"})
	lt.AddRuleChecker(gwr.GetID(), rules.AsSentenceCheckerSimple(gwr.Match))

	ws := rules.NewMultipleWhitespaceRule(map[string]string{
		"whitespace_repetition": "Possible typo: you repeated a whitespace",
	})
	lt.AddRuleChecker(ws.GetID(), rules.AsSentenceCheckerSimple(func(s *languagetool.AnalyzedSentence) []*rules.RuleMatch {
		return ws.Match([]*languagetool.AnalyzedSentence{s})
	}))

	// Java SentenceWhitespaceRule → SENTENCE_WHITESPACE (not invent PL_).
	sw := rules.NewSentenceWhitespaceRule(nil)
	lt.AddTextLevelRuleChecker(sw.GetID(), rules.AsTextLevelChecker(sw.MatchList))

	ub := NewPolishUnpairedBracketsRule(nil)
	lt.AddTextLevelRuleChecker(ub.GetID(), rules.AsTextLevelChecker(ub.MatchList))

	sp := NewMorfologikPolishSpellerRule()
	if p := morfologik.DiscoverLanguageDict(PolishSpellerDict); p != "" {
		if WirePolishFilterSpeller(p) {
			sp.IsMisspelled = FilterDictIsMisspelled
		}
	}
	ltRef := lt
	WirePolishSpellerTagPOS(sp, func(token string) []languagetool.TokenTag {
		if ltRef.TagWord == nil {
			return nil
		}
		return ltRef.TagWord(token)
	})
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))

	pwr := NewPolishWordRepeatRule(map[string]string{"repetition": "Powtórzenie"})
	lt.AddRuleChecker(pwr.GetID(), rules.AsSentenceCheckerSimple(pwr.Match))

	cr := NewCompoundRule(nil)
	lt.AddRuleChecker(cr.GetID(), rules.AsSentenceCheckerSimple(cr.Match))

	sr := NewSimpleReplaceRule(nil)
	lt.AddRuleChecker(sr.GetID(), rules.AsSentenceCheckerSimple(sr.Match))

	wc := NewWordCoherencyRule(nil)
	lt.AddTextLevelRuleChecker(wc.GetID(), rules.AsTextLevelChecker(wc.MatchList))

	dr := NewDashRule(nil)
	lt.AddRuleChecker(dr.GetID(), rules.AsSentenceCheckerSimple(dr.Match))
}
