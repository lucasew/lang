package ru

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
)

// RegisterCoreRussianRules ports Russian.getRelevantRules (+ chunker / ignored chars).
// Java list only — no invent SharedLayout extras (no DoublePunctuation / EmptyLine /
// generic UNPAIRED_BRACKETS / WhitespaceBeforePunct / PunctuationMarkAtParagraphEnd).
func RegisterCoreRussianRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	WireRussianChunker(lt)
	lt.PriorityForId = language.RussianPriorityForId
	lt.IgnoredCharacters = languagetool.RussianIgnoredCharactersRegex

	cw := rules.NewCommaWhitespaceRule(nil)
	lt.AddRuleChecker(cw.GetID(), rules.AsSentenceCheckerSimple(cw.Match))

	// Java DoublePunctuationRule is commented out (XML rule instead).

	up := rules.NewUppercaseSentenceStartRule(nil, "ru")
	lt.AddRuleChecker(up.GetID(), rules.AsSentenceCheckerSimple(func(s *languagetool.AnalyzedSentence) []*rules.RuleMatch {
		return up.MatchList([]*languagetool.AnalyzedSentence{s})
	}))

	sp := NewMorfologikRussianSpellerRule()
	if p := morfologik.DiscoverLanguageDict(RussianSpellerDict); p != "" {
		if WireRussianFilterSpeller(p) {
			sp.IsMisspelled = FilterDictIsMisspelled
		}
	}
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))

	// Java WordRepeatRule commented out — moved to RussianSimpleWordRepeatRule later.

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

	// EmptyLineRule commented out in Java.

	ls := rules.NewLongSentenceRule(map[string]string{
		"long_sentence_rule_msg2": "This sentence is too long (%d words)",
	}, 50)
	lt.AddTextLevelRuleChecker(ls.GetID(), rules.AsTextLevelChecker(ls.MatchList))

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

	fw := NewRussianFillerWordsRule(nil)
	lt.AddRuleChecker(fw.GetID(), rules.AsSentenceCheckerSimple(fw.Match))

	// Java PunctuationMarkAtParagraphEnd2 — always setDefaultOff.
	ppe2 := rules.NewPunctuationMarkAtParagraphEnd2(map[string]string{
		"punctuation_mark_paragraph_end2": "Add a punctuation mark at paragraph end",
	})
	lt.AddTextLevelRuleChecker(ppe2.GetID(), rules.AsTextLevelChecker(ppe2.MatchList))
	lt.MarkDefaultOff(ppe2.GetID())

	// specific to Russian
	yo := NewMorfologikRussianYOSpellerRule()
	if p := morfologik.DiscoverLanguageDict(RussianYOSpellerDict); p != "" {
		_ = p
	}
	lt.AddRuleChecker(yo.GetID(), rules.AsSentenceChecker(yo.Match))
	lt.MarkDefaultOff(MorfologikRussianYOSpellerRuleID)

	ub := NewRussianUnpairedBracketsRule(nil)
	lt.AddTextLevelRuleChecker(ub.GetID(), rules.AsTextLevelChecker(ub.MatchList))

	cr := NewRussianCompoundRule(nil)
	lt.AddRuleChecker(cr.GetID(), rules.AsSentenceCheckerSimple(cr.Match))

	sr := NewRussianSimpleReplaceRule(nil)
	lt.AddRuleChecker(sr.GetID(), rules.AsSentenceCheckerSimple(sr.Match))

	// Java RussianSimpleWordRepeatRule → WORD_REPEAT_RULE id.
	wr := NewRussianSimpleWordRepeatRule(map[string]string{"repetition": "Повтор слова"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))

	wc := NewRussianWordCoherencyRule(nil)
	lt.AddTextLevelRuleChecker(wc.GetID(), rules.AsTextLevelChecker(wc.MatchList))

	rwr := NewRussianWordRepeatRule(map[string]string{"repetition": "Повтор слов в предложении"})
	lt.AddRuleChecker(rwr.GetID(), rules.AsSentenceCheckerSimple(rwr.Match))

	// Java RussianWordRootRepeatRule / RussianVerbConjugationRule — missing from prior invent path.
	wrr := NewRussianWordRootRepeatRule(nil)
	lt.AddTextLevelRuleChecker(wrr.GetID(), rules.AsTextLevelChecker(wrr.MatchList))

	vc := NewRussianVerbConjugationRule(nil)
	lt.AddRuleChecker(vc.GetID(), rules.AsSentenceCheckerSimple(vc.Match))

	dr := NewRussianDashRule(nil)
	lt.AddRuleChecker(dr.GetID(), rules.AsSentenceCheckerSimple(dr.Match))

	sc := NewRussianSpecificCaseRule(nil)
	lt.AddRuleChecker(sc.GetID(), rules.AsSentenceCheckerSimple(sc.Match))
}
