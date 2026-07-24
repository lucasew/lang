package pt

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
)

// RegisterCorePortugueseRules ports Portuguese.getRelevantRules / createDefaultSpellingRule.
// Java list only — no invent SharedLayout extras; no invent agreement/pre-reform/
// regional replace always-on (those are variant or separate packages).
func RegisterCorePortugueseRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	lt.PriorityForId = language.PortuguesePriorityForIdForCode(lt.GetLanguageCode())

	cw := rules.NewCommaWhitespaceRule(nil)
	lt.AddRuleChecker(cw.GetID(), rules.AsSentenceCheckerSimple(cw.Match))

	ub := NewPortugueseUnpairedBracketsRule(nil)
	lt.AddTextLevelRuleChecker(ub.GetID(), rules.AsTextLevelChecker(ub.MatchList))

	// Speller early (Java order: after unpaired, before long-sentence).
	var sp *MorfologikPortugueseSpellerRule
	if strings.Contains(strings.ToLower(lt.GetLanguageCode()), "br") {
		sp = NewMorfologikBrazilianPortugueseSpellerRule()
	} else {
		sp = NewMorfologikPortugalPortugueseSpellerRule()
	}
	if p := morfologik.DiscoverLanguageDict(sp.GetFileName()); p != "" {
		if WirePortugueseFilterSpeller(p) {
			sp.IsMisspelled = FilterDictIsMisspelled
		}
	}
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))

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

	up := rules.NewUppercaseSentenceStartRule(nil, "pt")
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

	el := rules.NewEmptyLineRule(map[string]string{
		"empty_line_rule_msg": "Empty line",
	})
	lt.AddTextLevelRuleChecker(el.GetID(), rules.AsTextLevelChecker(el.MatchList))
	if el.IsDefaultOff() {
		lt.MarkDefaultOff(el.GetID())
	}

	prb := rules.NewParagraphRepeatBeginningRule(map[string]string{
		"repetition_paragraph_beginning_last_msg": "Paragraphs should not begin with the same words",
	})
	lt.AddTextLevelRuleChecker(prb.GetID(), rules.AsTextLevelChecker(prb.MatchList))
	if prb.IsDefaultOff() {
		lt.MarkDefaultOff(prb.GetID())
	}

	// Java PunctuationMarkAtParagraphEnd(messages, this, true) → defaultActive true.
	ppe := rules.NewPunctuationMarkAtParagraphEnd(map[string]string{
		"punctuation_mark_paragraph_end_msg": "Add a punctuation mark at paragraph end",
	})
	ppe.DefaultOff = false
	lt.AddTextLevelRuleChecker(ppe.GetID(), rules.AsTextLevelChecker(ppe.MatchList))

	// Specific to Portuguese (Java order)
	postC := NewPostReformPortugueseCompoundRule(nil)
	lt.AddRuleChecker(postC.GetID(), rules.AsSentenceCheckerSimple(postC.Match))

	col := NewPortugueseColourHyphenationRule(nil)
	lt.AddRuleChecker(col.GetID(), rules.AsSentenceCheckerSimple(col.Match))

	or := NewPortugueseOrthographyReplaceRule(nil)
	lt.AddRuleChecker(or.GetID(), rules.AsSentenceCheckerSimple(or.Match))

	sr := NewPortugueseReplaceRule(nil)
	lt.AddRuleChecker(sr.GetID(), rules.AsSentenceCheckerSimple(sr.Match))

	bar := NewPortugueseBarbarismsRule(nil)
	lt.AddRuleChecker(bar.GetID(), rules.AsSentenceCheckerSimple(bar.Match))

	cl := NewPortugueseClicheRule(nil)
	lt.AddRuleChecker(cl.GetID(), rules.AsSentenceCheckerSimple(cl.Match))

	fw := NewPortugueseFillerWordsRule(nil)
	lt.AddRuleChecker(fw.GetID(), rules.AsSentenceCheckerSimple(fw.Match))

	rd := NewPortugueseRedundancyRule(nil)
	lt.AddRuleChecker(rd.GetID(), rules.AsSentenceCheckerSimple(rd.Match))

	pe := NewPortugueseWordinessRule(nil)
	lt.AddRuleChecker(pe.GetID(), rules.AsSentenceCheckerSimple(pe.Match))

	wiki := NewPortugueseWikipediaRule(nil)
	lt.AddRuleChecker(wiki.GetID(), rules.AsSentenceCheckerSimple(wiki.Match))

	wr := NewPortugueseWordRepeatRule(map[string]string{"repetition": "Repetição de palavra"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))

	wrb := NewPortugueseWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "Três frases sucessivas começam com a mesma palavra.",
		"desc_repetition_beginning_adv":  "Três frases sucessivas começam com o mesmo advérbio.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))

	acc := NewPortugueseAccentuationCheckRule()
	lt.AddRuleChecker(acc.GetID(), rules.AsSentenceCheckerSimple(acc.Match))
	if acc.DefaultOff {
		lt.MarkDefaultOff(acc.GetID())
	}

	di := NewPortugueseDiacriticsRule(nil)
	lt.AddRuleChecker(di.GetID(), rules.AsSentenceCheckerSimple(di.Match))

	ww := NewPortugueseWrongWordInContextRule(nil)
	lt.AddRuleChecker(ww.GetID(), rules.AsSentenceCheckerSimple(ww.Match))

	wc := NewPortugueseWordCoherencyRule(nil)
	lt.AddTextLevelRuleChecker(wc.GetID(), rules.AsTextLevelChecker(wc.MatchList))

	uc := NewPortugueseUnitConversionRule(nil)
	lt.AddRuleChecker(uc.GetID(), rules.AsSentenceCheckerSimple(uc.Match))

	// Java: tooEasy true then false; default off.
	readEasy := NewPortugueseReadabilityRule(true, -1)
	lt.AddTextLevelRuleChecker(readEasy.GetID(), rules.AsTextLevelChecker(readEasy.MatchList))
	if readEasy.IsDefaultOff() {
		lt.MarkDefaultOff(readEasy.GetID())
	}
	readDiff := NewPortugueseReadabilityRule(false, -1)
	lt.AddTextLevelRuleChecker(readDiff.GetID(), rules.AsTextLevelChecker(readDiff.MatchList))
	if readDiff.IsDefaultOff() {
		lt.MarkDefaultOff(readDiff.GetID())
	}

	dp := rules.NewDoublePunctuationRule(nil)
	lt.AddRuleChecker(dp.GetID(), rules.AsSentenceCheckerSimple(dp.Match))

	ec := NewEnglishContractionSpellingRule(nil)
	lt.AddRuleChecker(ec.GetID(), rules.AsSentenceCheckerSimple(ec.Match))
}
