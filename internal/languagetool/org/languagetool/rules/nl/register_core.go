package nl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCoreDutchRules ports Dutch.getRelevantRules surface (class getId parity).
// Java list only — no invent SharedLayout extras (no empty-line / whitespace-before-punct /
// paragraph-begin/end / punctuation-paragraph-end / word-repeat).
func RegisterCoreDutchRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	lt.PriorityForId = language.DutchPriorityForId

	cw := rules.NewCommaWhitespaceRule(nil)
	lt.AddRuleChecker(cw.GetID(), rules.AsSentenceCheckerSimple(cw.Match))

	dp := rules.NewDoublePunctuationRule(nil)
	lt.AddRuleChecker(dp.GetID(), rules.AsSentenceCheckerSimple(dp.Match))

	ubr := NewDutchUnpairedBracketsRule(nil)
	lt.AddTextLevelRuleChecker(ubr.GetID(), rules.AsTextLevelChecker(ubr.MatchList))

	up := rules.NewUppercaseSentenceStartRule(nil, "nl")
	lt.AddRuleChecker(up.GetID(), rules.AsSentenceCheckerSimple(func(s *languagetool.AnalyzedSentence) []*rules.RuleMatch {
		return up.MatchList([]*languagetool.AnalyzedSentence{s})
	}))

	sp := NewMorfologikDutchSpellerRule()
	if TryWireDutchFilterSpeller() {
		sp.IsMisspelled = FilterDictIsMisspelled
	}
	_ = TryWireDutchFilterTagger()
	BindDefaultCompoundAcceptorFilters()
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))

	ws := rules.NewMultipleWhitespaceRule(map[string]string{
		"whitespace_repetition": "Possible typo: you repeated a whitespace",
	})
	lt.AddRuleChecker(ws.GetID(), rules.AsSentenceCheckerSimple(func(s *languagetool.AnalyzedSentence) []*rules.RuleMatch {
		return ws.Match([]*languagetool.AnalyzedSentence{s})
	}))

	cr := NewCompoundRule(nil)
	lt.AddRuleChecker(cr.GetID(), rules.AsSentenceCheckerSimple(cr.Match))

	ww := NewDutchWrongWordInContextRule(nil)
	lt.AddRuleChecker(ww.GetID(), rules.AsSentenceCheckerSimple(ww.Match))

	wc := NewWordCoherencyRule(nil)
	lt.AddTextLevelRuleChecker(wc.GetID(), rules.AsTextLevelChecker(wc.MatchList))

	sr := NewSimpleReplaceRule(nil)
	lt.AddRuleChecker(sr.GetID(), rules.AsSentenceCheckerSimple(sr.Match))

	// Java LongSentenceRule → TOO_LONG_SENTENCE.
	ls := NewLongSentenceRule(map[string]string{
		"long_sentence_rule_msg2": "Deze zin is te lang (%d woorden)",
	}, 40)
	lt.AddTextLevelRuleChecker(ls.GetID(), rules.AsTextLevelChecker(ls.MatchList))

	// Java LongParagraphRule 5-arg always setDefaultOff (defaultActive param unused).
	lp := rules.NewLongParagraphRule(map[string]string{
		"long_paragraph_rule_msg": "Deze alinea is te lang (%d woorden)",
	}, 220)
	lt.AddTextLevelRuleChecker(lp.GetID(), rules.AsTextLevelChecker(lp.MatchList))
	if lp.IsDefaultOff() {
		lt.MarkDefaultOff(lp.GetID())
	}

	pw := NewPreferredWordRule(nil)
	lt.AddRuleChecker(pw.GetID(), rules.AsSentenceCheckerSimple(pw.Match))

	sc := NewSpaceInCompoundRule(nil)
	lt.AddRuleChecker(sc.GetID(), rules.AsSentenceCheckerSimple(sc.Match))

	sw := NewSentenceWhitespaceRule(nil)
	lt.AddTextLevelRuleChecker(sw.GetID(), rules.AsTextLevelChecker(sw.MatchList))

	cc := NewCheckCaseRule(nil)
	lt.AddRuleChecker(cc.GetID(), rules.AsSentenceCheckerSimple(cc.Match))
}
