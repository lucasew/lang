package ga

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
)

// RegisterCoreIrishRules installs shared layout + word-repeat + beginning.
func RegisterCoreIrishRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	lt.PriorityForId = language.IrishPriorityForId
	rules.RegisterSharedLayoutRules(lt, "ga")
	wr := rules.NewWordRepeatRule(map[string]string{"repetition": "Athrá"})
	wr.IDOverride = "GA_WORD_REPEAT_RULE"
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	wrb := NewWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "Tosaíonn trí abairt as a chéile leis an bhfocal céanna.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))
	// Soft invent token sequences removed (faithful-port): incomplete without grammar.xml, not invented.

	// Official replace tables (embedded from upstream).
	ir := NewIrishReplaceRule(nil)
	lt.AddRuleChecker(ir.GetID(), rules.AsSentenceCheckerSimple(ir.Match))
	fgb := NewIrishFGBEqReplaceRule(nil)
	lt.AddRuleChecker(fgb.GetID(), rules.AsSentenceCheckerSimple(fgb.Match))
	dp := NewDativePluralStandardReplaceRule(nil)
	lt.AddRuleChecker(dp.GetID(), rules.AsSentenceCheckerSimple(dp.Match))

	// Official compounds, specific-case, prestandard replace.
	cr := NewCompoundRule(nil)
	lt.AddRuleChecker(cr.GetID(), rules.AsSentenceCheckerSimple(cr.Match))
	sc := NewIrishSpecificCaseRule(nil)
	lt.AddRuleChecker(sc.GetID(), rules.AsSentenceCheckerSimple(sc.Match))
	pr := NewPrestandardReplaceRule(nil)
	lt.AddRuleChecker(pr.GetID(), rules.AsSentenceCheckerSimple(pr.Match))

	// Official placenames, people, spaces, English homophones.
	lg := NewLogainmRule(nil)
	lt.AddRuleChecker(lg.GetID(), rules.AsSentenceCheckerSimple(lg.Match))
	pp := NewPeopleRule(nil)
	lt.AddRuleChecker(pp.GetID(), rules.AsSentenceCheckerSimple(pp.Match))
	sp := NewSpacesRule(nil)
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceCheckerSimple(sp.Match))
	eh := NewEnglishHomophoneRule(nil)
	lt.AddRuleChecker(eh.GetID(), rules.AsSentenceCheckerSimple(eh.Match))

	// Java createDefaultSpellingRule → MorfologikIrishSpellerRule.
	// Always full Match (hyphen tokenizing + maths/halfwidth isMisspelled).
	ms := NewMorfologikIrishSpellerRule()
	if p := morfologik.DiscoverLanguageDict(IrishSpellerDict); p != "" {
		// Optional binary path; map Words stay fail-closed without invent.
		_ = p
	}
	lt.AddRuleChecker(ms.GetID(), rules.AsSentenceChecker(ms.Match))
}
