package ga

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
)

// RegisterCoreIrishRules ports Irish.getRelevantRules / createDefaultSpellingRule.
// Java list only — no invent SharedLayout extras (no EMPTY_LINE / WHITESPACE_PUNCTUATION /
// invent GA_SENTENCE_WHITESPACE / invent GA_UNPAIRED_BRACKETS).
func RegisterCoreIrishRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	lt.PriorityForId = language.IrishPriorityForId

	cw := rules.NewCommaWhitespaceRule(nil)
	lt.AddRuleChecker(cw.GetID(), rules.AsSentenceCheckerSimple(cw.Match))

	ub := NewUnpairedBracketsRule(nil)
	lt.AddTextLevelRuleChecker(ub.GetID(), rules.AsTextLevelChecker(ub.MatchList))

	dp := rules.NewDoublePunctuationRule(nil)
	lt.AddRuleChecker(dp.GetID(), rules.AsSentenceCheckerSimple(dp.Match))

	// Java registers UppercaseSentenceStartRule twice; same id → one registration.
	up := rules.NewUppercaseSentenceStartRule(nil, "ga")
	lt.AddRuleChecker(up.GetID(), rules.AsSentenceCheckerSimple(func(s *languagetool.AnalyzedSentence) []*rules.RuleMatch {
		return up.MatchList([]*languagetool.AnalyzedSentence{s})
	}))

	// Java LongSentenceRule → TOO_LONG_SENTENCE (not invent GA_).
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

	ws := rules.NewMultipleWhitespaceRule(map[string]string{
		"whitespace_repetition": "Possible typo: you repeated a whitespace",
	})
	lt.AddRuleChecker(ws.GetID(), rules.AsSentenceCheckerSimple(func(s *languagetool.AnalyzedSentence) []*rules.RuleMatch {
		return ws.Match([]*languagetool.AnalyzedSentence{s})
	}))

	// Java SentenceWhitespaceRule → SENTENCE_WHITESPACE (not invent GA_SENTENCE_WHITESPACE).
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

	prb := rules.NewParagraphRepeatBeginningRule(map[string]string{
		"repetition_paragraph_beginning_last_msg": "Paragraphs should not begin with the same words",
	})
	lt.AddTextLevelRuleChecker(prb.GetID(), rules.AsTextLevelChecker(prb.MatchList))
	if prb.IsDefaultOff() {
		lt.MarkDefaultOff(prb.GetID())
	}

	// Java WordRepeatRule default id WORD_REPEAT_RULE — no WordRepeatBeginning.
	wr := rules.NewWordRepeatRule(map[string]string{"repetition": "Athrá"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))

	ms := NewMorfologikIrishSpellerRule()
	if p := morfologik.DiscoverLanguageDict(IrishSpellerDict); p != "" {
		_ = p
	}
	lt.AddRuleChecker(ms.GetID(), rules.AsSentenceChecker(ms.Match))

	lg := NewLogainmRule(nil)
	lt.AddRuleChecker(lg.GetID(), rules.AsSentenceCheckerSimple(lg.Match))
	pp := NewPeopleRule(nil)
	lt.AddRuleChecker(pp.GetID(), rules.AsSentenceCheckerSimple(pp.Match))
	spa := NewSpacesRule(nil)
	lt.AddRuleChecker(spa.GetID(), rules.AsSentenceCheckerSimple(spa.Match))
	cr := NewCompoundRule(nil)
	lt.AddRuleChecker(cr.GetID(), rules.AsSentenceCheckerSimple(cr.Match))
	pr := NewPrestandardReplaceRule(nil)
	lt.AddRuleChecker(pr.GetID(), rules.AsSentenceCheckerSimple(pr.Match))
	ir := NewIrishReplaceRule(nil)
	lt.AddRuleChecker(ir.GetID(), rules.AsSentenceCheckerSimple(ir.Match))
	fgb := NewIrishFGBEqReplaceRule(nil)
	lt.AddRuleChecker(fgb.GetID(), rules.AsSentenceCheckerSimple(fgb.Match))
	eh := NewEnglishHomophoneRule(nil)
	lt.AddRuleChecker(eh.GetID(), rules.AsSentenceCheckerSimple(eh.Match))
	// Java DhaNoBeirtRule — was missing from prior invent registration.
	dn := NewDhaNoBeirtRule(nil)
	lt.AddRuleChecker(dn.GetID(), rules.AsSentenceCheckerSimple(dn.Match))
	dpl := NewDativePluralStandardReplaceRule(nil)
	lt.AddRuleChecker(dpl.GetID(), rules.AsSentenceCheckerSimple(dpl.Match))
	sc := NewIrishSpecificCaseRule(nil)
	lt.AddRuleChecker(sc.GetID(), rules.AsSentenceCheckerSimple(sc.Match))
}
