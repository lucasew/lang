package pt

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
)

// RegisterCorePortugueseRules ports Portuguese.getRelevantRules / createDefaultSpellingRule
// plus locale extras (PortugalPortuguese when pt-PT / bare pt; BrazilianPortuguese when pt-BR).
// Java list only — no invent SharedLayout extras.
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

	// Specific to Portuguese (Java Portuguese.getRelevantRules order)
	postC := NewPostReformPortugueseCompoundRule(nil)
	lt.AddRuleChecker(postC.GetID(), rules.AsSentenceCheckerSimple(postC.Match))

	col := NewPortugueseColourHyphenationRule(nil)
	lt.AddRuleChecker(col.GetID(), rules.AsSentenceCheckerSimple(col.Match))

	or := NewPortugueseOrthographyReplaceRule(nil)
	lt.AddRuleChecker(or.GetID(), rules.AsSentenceCheckerSimple(or.Match))

	sr := NewPortugueseReplaceRule(nil)
	lt.AddRuleChecker(sr.GetID(), rules.AsSentenceCheckerSimple(sr.Match))

	// Base barbarisms: Java Portuguese → "/pt/barbarisms.txt" (do not use pt-BR pack here).
	bar := NewPortugueseBarbarismsRule(nil, PortugueseBarbarismsFilePath)
	lt.AddRuleChecker(bar.GetID(), rules.AsSentenceCheckerSimple(bar.Match))

	cl := NewPortugueseClicheRule(nil, PortugueseClicheFilePath)
	lt.AddRuleChecker(cl.GetID(), rules.AsSentenceCheckerSimple(cl.Match))

	fw := NewPortugueseFillerWordsRule(nil)
	lt.AddRuleChecker(fw.GetID(), rules.AsSentenceCheckerSimple(fw.Match))
	// Java AbstractStatisticStyleRule / AbstractFillerWordsRule → setDefaultOff().
	if fw.IsDefaultOff() {
		lt.MarkDefaultOff(fw.GetID())
	}

	rd := NewPortugueseRedundancyRule(nil, PortugueseRedundancyFilePath)
	lt.AddRuleChecker(rd.GetID(), rules.AsSentenceCheckerSimple(rd.Match))

	pe := NewPortugueseWordinessRule(nil, PortugueseWordinessFilePath)
	lt.AddRuleChecker(pe.GetID(), rules.AsSentenceCheckerSimple(pe.Match))

	wiki := NewPortugueseWikipediaRule(nil, PortugueseWikipediaFilePath)
	lt.AddRuleChecker(wiki.GetID(), rules.AsSentenceCheckerSimple(wiki.Match))

	wr := NewPortugueseWordRepeatRule(map[string]string{"repetition": "Repetição de palavra"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))

	wrb := NewPortugueseWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "Três frases sucessivas começam com a mesma palavra.",
		"desc_repetition_beginning_adv":  "Três frases sucessivas começam com o mesmo advérbio.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))

	acc := NewPortugueseAccentuationCheckRule(nil)
	lt.AddRuleChecker(acc.GetID(), rules.AsSentenceCheckerSimple(acc.Match))
	if acc.DefaultOff {
		lt.MarkDefaultOff(acc.GetID())
	}

	di := NewPortugueseDiacriticsRule(nil)
	lt.AddRuleChecker(di.GetID(), rules.AsSentenceCheckerSimple(di.Match))
	if di.IsDefaultOff() {
		lt.MarkDefaultOff(di.GetID())
	}

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

	// Locale extras after base (Java *Portuguese.getRelevantRules after super).
	RegisterPortugalPortugueseExtraRules(lt)
	RegisterBrazilianPortugueseExtraRules(lt)
}

// RegisterPortugalPortugueseExtraRules ports PortugalPortuguese.getRelevantRules
// rules added after super.getRelevantRules. Only for PortugalPortuguese codes
// (pt-PT / bare pt product surface). Same getId as base is registered again when
// Java constructs a second instance with a different resource path (or a second
// PostReformPortugueseCompoundRule). Irish precedent for identical second
// compound instance: still one compound registration; path packs re-register so
// both lists run under the shared getId.
func RegisterPortugalPortugueseExtraRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	code := lt.GetLanguageCode()
	if !languageIsPortugalPortuguese(code) {
		return
	}

	// Java: rules.add(new PostReformPortugueseCompoundRule(...)) — second instance,
	// same getId. Irish twin: same id → one registration (already in base).

	// Java: new PostReformPortugueseDashRule(messages)
	dash := NewPostReformPortugueseDashRule(nil)
	lt.AddRuleChecker(dash.GetID(), rules.AsSentenceCheckerSimple(dash.Match))

	// Java: new PortugueseAgreementReplaceRule(messages, this)
	agr := NewPortugueseAgreementReplaceRule(nil)
	lt.AddRuleChecker(agr.GetID(), rules.AsSentenceCheckerSimple(agr.Match))

	// Java: new PortugalPortugueseReplaceRule(messages, "/pt/pt-PT/replace.txt", this)
	ptRepl := NewPortugalPortugueseReplaceRule(nil, PortugalPortugueseReplaceFilePath)
	lt.AddRuleChecker(ptRepl.GetID(), rules.AsSentenceCheckerSimple(ptRepl.Match))

	// Java: new PortugueseBarbarismsRule(messages, "/pt/pt-PT/barbarisms.txt", this)
	// Same getId as base /pt/barbarisms.txt — second instance with pt-PT pack.
	barPT := NewPortugueseBarbarismsRule(nil, PortugueseBarbarismsPTFilePath)
	lt.AddRuleChecker(barPT.GetID(), rules.AsSentenceCheckerSimple(barPT.Match))

	// Java: new PortugueseArchaismsRule(messages, "/pt/pt-PT/archaisms.txt", this)
	arch := NewPortugueseArchaismsRule(nil, PortugueseArchaismsPTFilePath)
	lt.AddRuleChecker(arch.GetID(), rules.AsSentenceCheckerSimple(arch.Match))

	// Java: new PortugueseClicheRule(messages, "/pt/pt-PT/cliches.txt", this)
	clPT := NewPortugueseClicheRule(nil, PortugueseClichePTFilePath)
	lt.AddRuleChecker(clPT.GetID(), rules.AsSentenceCheckerSimple(clPT.Match))

	// Java: new PortugueseRedundancyRule(messages, "/pt/pt-PT/redundancies.txt", this)
	rdPT := NewPortugueseRedundancyRule(nil, PortugueseRedundancyPTFilePath)
	lt.AddRuleChecker(rdPT.GetID(), rules.AsSentenceCheckerSimple(rdPT.Match))

	// Java: new PortugueseWordinessRule(messages, "/pt/pt-PT/wordiness.txt", this)
	wiPT := NewPortugueseWordinessRule(nil, PortugueseWordinessPTFilePath)
	lt.AddRuleChecker(wiPT.GetID(), rules.AsSentenceCheckerSimple(wiPT.Match))

	// Java: new PortugueseWikipediaRule(messages, "/pt/pt-PT/wikipedia.txt", this)
	wikiPT := NewPortugueseWikipediaRule(nil, PortugueseWikipediaPTFilePath)
	lt.AddRuleChecker(wikiPT.GetID(), rules.AsSentenceCheckerSimple(wikiPT.Match))
}

func languageIsPortugalPortuguese(code string) bool {
	// Mirror language.isPortugalPortugueseCode without importing cycle helpers on short forms.
	// language package already exposes PortuguesePriorityForIdForCode; use same cases.
	c := strings.ToLower(strings.ReplaceAll(code, "_", "-"))
	switch c {
	case "pt", "pt-pt":
		return true
	}
	return strings.HasPrefix(c, "pt-pt")
}

// RegisterBrazilianPortugueseExtraRules ports BrazilianPortuguese.getRelevantRules
// rules added after super.getRelevantRules. Only for BrazilianPortuguese codes (pt-BR).
// Differs from PortugalPortuguese: BrazilianPortugueseReplaceRule + pt-BR packs;
// no PortugueseAgreementReplaceRule / PortugalPortugueseReplaceRule.
// Same getId as base is registered again when Java constructs a second instance with
// a different resource path (or a second PostReformPortugueseCompoundRule).
func RegisterBrazilianPortugueseExtraRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	code := lt.GetLanguageCode()
	if !languageIsBrazilianPortuguese(code) {
		return
	}

	// Java: rules.add(new PostReformPortugueseCompoundRule(...)) — second instance,
	// same getId. Irish twin: same id → one registration (already in base).

	// Java: new PostReformPortugueseDashRule(messages)
	dash := NewPostReformPortugueseDashRule(nil)
	lt.AddRuleChecker(dash.GetID(), rules.AsSentenceCheckerSimple(dash.Match))

	// Java: new BrazilianPortugueseReplaceRule(messages, "/pt/pt-BR/replace.txt", this)
	brRepl := NewBrazilianPortugueseReplaceRule(nil, BrazilianPortugueseReplaceFilePath)
	lt.AddRuleChecker(brRepl.GetID(), rules.AsSentenceCheckerSimple(brRepl.Match))

	// Java: new PortugueseBarbarismsRule(messages, "/pt/pt-BR/barbarisms.txt", this)
	// Same getId as base /pt/barbarisms.txt — second instance with pt-BR pack.
	barBR := NewPortugueseBarbarismsRule(nil, PortugueseBarbarismsBRFilePath)
	lt.AddRuleChecker(barBR.GetID(), rules.AsSentenceCheckerSimple(barBR.Match))

	// Java: new PortugueseArchaismsRule(messages, "/pt/pt-BR/archaisms.txt", this)
	arch := NewPortugueseArchaismsRule(nil, PortugueseArchaismsBRFilePath)
	lt.AddRuleChecker(arch.GetID(), rules.AsSentenceCheckerSimple(arch.Match))

	// Java: new PortugueseClicheRule(messages, "/pt/pt-BR/cliches.txt", this)
	clBR := NewPortugueseClicheRule(nil, PortugueseClicheBRFilePath)
	lt.AddRuleChecker(clBR.GetID(), rules.AsSentenceCheckerSimple(clBR.Match))

	// Java: new PortugueseRedundancyRule(messages, "/pt/pt-BR/redundancies.txt", this)
	rdBR := NewPortugueseRedundancyRule(nil, PortugueseRedundancyBRFilePath)
	lt.AddRuleChecker(rdBR.GetID(), rules.AsSentenceCheckerSimple(rdBR.Match))

	// Java: new PortugueseWordinessRule(messages, "/pt/pt-BR/wordiness.txt", this)
	wiBR := NewPortugueseWordinessRule(nil, PortugueseWordinessBRFilePath)
	lt.AddRuleChecker(wiBR.GetID(), rules.AsSentenceCheckerSimple(wiBR.Match))

	// Java: new PortugueseWikipediaRule(messages, "/pt/pt-BR/wikipedia.txt", this)
	wikiBR := NewPortugueseWikipediaRule(nil, PortugueseWikipediaBRFilePath)
	lt.AddRuleChecker(wikiBR.GetID(), rules.AsSentenceCheckerSimple(wikiBR.Match))
}

func languageIsBrazilianPortuguese(code string) bool {
	c := strings.ToLower(strings.ReplaceAll(code, "_", "-"))
	switch c {
	case "pt-br":
		return true
	}
	return strings.HasPrefix(c, "pt-br")
}
