package be

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCoreBelarusianRules ports Belarusian.getRelevantRules (+ priority / ignored chars).
// Java list only — no invent SharedLayout extras (no unpaired / empty-line /
// whitespace-before-punct / generic PunctuationMarkAtParagraphEnd).
func RegisterCoreBelarusianRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	lt.PriorityForId = language.BelarusianPriorityForId
	// Java Belarusian.getIgnoredCharactersRegex: soft hyphen + combining acute/grave.
	lt.IgnoredCharacters = languagetool.BelarusianIgnoredCharactersRegex

	cw := rules.NewCommaWhitespaceRule(nil)
	lt.AddRuleChecker(cw.GetID(), rules.AsSentenceCheckerSimple(cw.Match))

	dp := rules.NewDoublePunctuationRule(nil)
	lt.AddRuleChecker(dp.GetID(), rules.AsSentenceCheckerSimple(dp.Match))

	sp := NewMorfologikBelarusianSpellerRule()
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))

	up := rules.NewUppercaseSentenceStartRule(nil, "be")
	lt.AddRuleChecker(up.GetID(), rules.AsSentenceCheckerSimple(func(s *languagetool.AnalyzedSentence) []*rules.RuleMatch {
		return up.MatchList([]*languagetool.AnalyzedSentence{s})
	}))

	ws := rules.NewMultipleWhitespaceRule(map[string]string{
		"whitespace_repetition": "Possible typo: you repeated a whitespace",
	})
	lt.AddRuleChecker(ws.GetID(), rules.AsSentenceCheckerSimple(func(s *languagetool.AnalyzedSentence) []*rules.RuleMatch {
		return ws.Match([]*languagetool.AnalyzedSentence{s})
	}))

	sw := rules.NewSentenceWhitespaceRule(nil)
	lt.AddTextLevelRuleChecker(sw.GetID(), rules.AsTextLevelChecker(sw.MatchList))

	wpe := rules.NewWhiteSpaceBeforeParagraphEnd(map[string]string{
		"whitespace_before_parapgraph_end_msg": "Don't end a paragraph with whitespace",
	})
	lt.AddTextLevelRuleChecker(wpe.GetID(), rules.AsTextLevelChecker(wpe.MatchList))
	if wpe.IsDefaultOff() {
		lt.MarkDefaultOff(wpe.GetID())
	}

	wpb := rules.NewWhiteSpaceAtBeginOfParagraph(map[string]string{
		"whitespace_at_begin_parapgraph_msg": "Don't start a paragraph with whitespace",
	})
	lt.AddRuleChecker(wpb.GetID(), rules.AsSentenceCheckerSimple(wpb.Match))
	if wpb.IsDefaultOff() {
		lt.MarkDefaultOff(wpb.GetID())
	}

	// Java LongSentenceRule(messages, userConfig, 50) → TOO_LONG_SENTENCE.
	ls := rules.NewLongSentenceRule(map[string]string{
		"long_sentence_rule_msg2": "This sentence is too long (%d words)",
	}, 50)
	lt.AddTextLevelRuleChecker(ls.GetID(), rules.AsTextLevelChecker(ls.MatchList))

	// Java LongParagraphRule(messages, this, userConfig) → maxWords 220, defaultOff.
	lp := rules.NewLongParagraphRule(map[string]string{
		"long_paragraph_rule_msg": "This paragraph is too long (%d words)",
	}, 220)
	lt.AddTextLevelRuleChecker(lp.GetID(), rules.AsTextLevelChecker(lp.MatchList))
	if lp.IsDefaultOff() {
		lt.MarkDefaultOff(lp.GetID())
	}

	prb := rules.NewParagraphRepeatBeginningRule(map[string]string{
		"repetition_paragraph_beginning_last_msg": "Paragraphs should not begin with the same words",
	})
	lt.AddTextLevelRuleChecker(prb.GetID(), rules.AsTextLevelChecker(prb.MatchList))
	if prb.IsDefaultOff() {
		lt.MarkDefaultOff(prb.GetID())
	}

	// Java PunctuationMarkAtParagraphEnd2 — always setDefaultOff.
	ppe2 := rules.NewPunctuationMarkAtParagraphEnd2(map[string]string{
		"punctuation_mark_paragraph_end2": "Add a punctuation mark at paragraph end",
	})
	lt.AddTextLevelRuleChecker(ppe2.GetID(), rules.AsTextLevelChecker(ppe2.MatchList))
	lt.MarkDefaultOff(ppe2.GetID())

	sr := NewSimpleReplaceRule(nil)
	lt.AddRuleChecker(sr.GetID(), rules.AsSentenceCheckerSimple(sr.Match))
	sc := NewBelarusianSpecificCaseRule(nil)
	lt.AddRuleChecker(sc.GetID(), rules.AsSentenceCheckerSimple(sc.Match))
}
