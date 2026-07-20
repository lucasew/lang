package fr

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
)

// RegisterCoreFrenchRules ports French.getRelevantRules / createDefaultSpellingRule.
// Java list only — no invent SharedLayout extras (no empty-line / unpaired-quotes /
// whitespace-before-punct / invent FR_SENTENCE_WHITESPACE / invent TOO_LONG_SENTENCE_FR).
// Java has no WordRepeatRule / WordRepeatBeginning.
func RegisterCoreFrenchRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	if languagetool.FrenchPriorityForIdHook != nil {
		lt.PriorityForId = languagetool.FrenchPriorityForIdHook
	}
	if languagetool.FilterFrenchRuleMatchesHook != nil {
		lt.FilterRuleMatches = languagetool.FilterFrenchRuleMatchesHook
	}

	// Java CommaWhitespaceRule(messages, false) → quotesWhitespaceCheck = false.
	cw := rules.NewCommaWhitespaceRule(nil)
	cw.QuotesWhitespaceCheck = false
	lt.AddRuleChecker(cw.GetID(), rules.AsSentenceCheckerSimple(cw.Match))

	dp := rules.NewDoublePunctuationRule(nil)
	lt.AddRuleChecker(dp.GetID(), rules.AsSentenceCheckerSimple(dp.Match))

	ub := NewFrenchUnpairedBracketsRule(nil)
	lt.AddTextLevelRuleChecker(ub.GetID(), rules.AsTextLevelChecker(ub.MatchList))

	// Java createDefaultSpellingRule → MorfologikFrenchSpellerRule (FR_SPELLING_RULE).
	sp := NewMorfologikFrenchSpellerRule()
	if p := morfologik.DiscoverLanguageDict(FrenchSpellerDict); p != "" {
		if WireFrenchFilterSpeller(p) {
			sp.IsMisspelled = FilterDictIsMisspelled
		}
	}
	ltRef := lt
	WireFrenchSpellerTagPOS(sp, func(token string) []languagetool.TokenTag {
		if ltRef.TagWord == nil {
			return nil
		}
		return ltRef.TagWord(token)
	})
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))

	up := rules.NewUppercaseSentenceStartRule(nil, "fr")
	lt.AddRuleChecker(up.GetID(), rules.AsSentenceCheckerSimple(func(s *languagetool.AnalyzedSentence) []*rules.RuleMatch {
		return up.MatchList([]*languagetool.AnalyzedSentence{s})
	}))

	ws := rules.NewMultipleWhitespaceRule(map[string]string{
		"whitespace_repetition": "Possible typo: you repeated a whitespace",
	})
	lt.AddRuleChecker(ws.GetID(), rules.AsSentenceCheckerSimple(func(s *languagetool.AnalyzedSentence) []*rules.RuleMatch {
		return ws.Match([]*languagetool.AnalyzedSentence{s})
	}))

	// Java SentenceWhitespaceRule → SENTENCE_WHITESPACE (not invent FR_).
	sw := rules.NewSentenceWhitespaceRule(nil)
	lt.AddTextLevelRuleChecker(sw.GetID(), rules.AsTextLevelChecker(sw.MatchList))

	// Java LongSentenceRule → TOO_LONG_SENTENCE (not invent FR_).
	ls := rules.NewLongSentenceRule(map[string]string{
		"long_sentence_rule_msg2": "This sentence is too long (%d words)",
	}, 40)
	lt.AddTextLevelRuleChecker(ls.GetID(), rules.AsTextLevelChecker(ls.MatchList))

	// Java LongParagraphRule(messages, this, userConfig) → maxWords 220, defaultOff.
	lp := rules.NewLongParagraphRule(map[string]string{
		"long_paragraph_rule_msg": "This paragraph is too long (%d words)",
	}, 220)
	lt.AddTextLevelRuleChecker(lp.GetID(), rules.AsTextLevelChecker(lp.MatchList))
	if lp.IsDefaultOff() {
		lt.MarkDefaultOff(lp.GetID())
	}

	// specific to French
	cr := NewCompoundRule(nil)
	lt.AddRuleChecker(cr.GetID(), rules.AsSentenceCheckerSimple(cr.Match))

	qws := NewQuestionWhitespaceStrictRule(nil)
	lt.AddRuleChecker(qws.GetID(), rules.AsSentenceCheckerSimple(qws.Match))

	qw := NewQuestionWhitespaceRule(nil)
	lt.AddRuleChecker(qw.GetID(), rules.AsSentenceCheckerSimple(qw.Match))

	sr := NewSimpleReplaceRule(nil)
	lt.AddRuleChecker(sr.GetID(), rules.AsSentenceCheckerSimple(sr.Match))

	rw := NewFrenchRepeatedWordsRule(nil)
	lt.AddTextLevelRuleChecker(rw.GetID(), rules.AsTextLevelChecker(rw.MatchList))
}
