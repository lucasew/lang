package pt

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
)

// RegisterCorePortugueseRules installs shared layout + Portuguese word-repeat + beginning.
func RegisterCorePortugueseRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	// Java Portuguese / PortugalPortuguese.getPriorityForId (pt-PT variant map then super).
	lt.PriorityForId = language.PortuguesePriorityForIdForCode(lt.GetLanguageCode())
	rules.RegisterSharedLayoutRules(lt, "pt")
	wr := NewPortugueseWordRepeatRule(map[string]string{"repetition": "Repetição de palavra"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	wrb := NewPortugueseWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "Três frases sucessivas começam com a mesma palavra.",
		"desc_repetition_beginning_adv":  "Três frases sucessivas começam com o mesmo advérbio.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))

	// Soft invent token sequences removed (faithful-port): incomplete without grammar.xml, not invented.

	// Official replace.txt + coherency (embedded from upstream).
	sr := NewPortugueseReplaceRule(nil)
	lt.AddRuleChecker(sr.GetID(), rules.AsSentenceCheckerSimple(sr.Match))
	wc := NewPortugueseWordCoherencyRule(nil)
	lt.AddTextLevelRuleChecker(wc.GetID(), rules.AsTextLevelChecker(wc.MatchList))

	// Official diacritics, style tables, wrong-word, orthography, agreement.
	di := NewPortugueseDiacriticsRule(nil)
	lt.AddRuleChecker(di.GetID(), rules.AsSentenceCheckerSimple(di.Match))
	rd := NewPortugueseRedundancyRule(nil)
	lt.AddRuleChecker(rd.GetID(), rules.AsSentenceCheckerSimple(rd.Match))
	pe := NewPortugueseWordinessRule(nil)
	lt.AddRuleChecker(pe.GetID(), rules.AsSentenceCheckerSimple(pe.Match))
	ww := NewPortugueseWrongWordInContextRule(nil)
	lt.AddRuleChecker(ww.GetID(), rules.AsSentenceCheckerSimple(ww.Match))
	or := NewPortugueseOrthographyReplaceRule(nil)
	lt.AddRuleChecker(or.GetID(), rules.AsSentenceCheckerSimple(or.Match))
	ag := NewPortugueseAgreementReplaceRule(nil)
	lt.AddRuleChecker(ag.GetID(), rules.AsSentenceCheckerSimple(ag.Match))

	// Regional replace + reform compounds/dash + EN contractions in PT + unit conversion.
	br := NewBrazilianPortugueseReplaceRule(nil)
	lt.AddRuleChecker(br.GetID(), rules.AsSentenceCheckerSimple(br.Match))
	ptR := NewPortugalPortugueseReplaceRule(nil)
	lt.AddRuleChecker(ptR.GetID(), rules.AsSentenceCheckerSimple(ptR.Match))
	postC := NewPostReformPortugueseCompoundRule(nil)
	lt.AddRuleChecker(postC.GetID(), rules.AsSentenceCheckerSimple(postC.Match))
	preC := NewPreReformPortugueseCompoundRule(nil)
	lt.AddRuleChecker(preC.GetID(), rules.AsSentenceCheckerSimple(preC.Match))
	postD := NewPostReformPortugueseDashRule(nil)
	lt.AddRuleChecker(postD.GetID(), rules.AsSentenceCheckerSimple(postD.Match))
	preD := NewPreReformPortugueseDashRule(nil)
	lt.AddRuleChecker(preD.GetID(), rules.AsSentenceCheckerSimple(preD.Match))
	ec := NewEnglishContractionSpellingRule(nil)
	lt.AddRuleChecker(ec.GetID(), rules.AsSentenceCheckerSimple(ec.Match))
	uc := NewPortugueseUnitConversionRule(nil)
	lt.AddRuleChecker(uc.GetID(), rules.AsSentenceCheckerSimple(uc.Match))

	// Java Portuguese.createDefaultSpellingRule → MorfologikPortugueseSpellerRule
	// (getId uses shortCodeWithCountryAndVariant: PT_PT / PT_BR).
	// Always register PT Match so getRuleMatches post-filters run (dialect SpecificRuleId,
	// clitic, diaeresis, titlecase-hyphen). SimplePredicateSpellerChecker alone would drop them.
	var sp *MorfologikPortugueseSpellerRule
	if strings.Contains(strings.ToLower(lt.GetLanguageCode()), "br") {
		sp = NewMorfologikBrazilianPortugueseSpellerRule()
	} else {
		sp = NewMorfologikPortugalPortugueseSpellerRule()
	}
	// When CFSA2 dict is on disk, wire filter dict for isMisspelled (map Speller stays empty).
	if p := morfologik.DiscoverLanguageDict(sp.GetFileName()); p != "" {
		if WirePortugueseFilterSpeller(p) {
			sp.IsMisspelled = FilterDictIsMisspelled
		}
	}
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))
}
