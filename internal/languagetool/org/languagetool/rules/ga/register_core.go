package ga

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// RegisterCoreIrishRules installs shared layout + word-repeat + beginning.
func RegisterCoreIrishRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "ga")
	wr := rules.NewWordRepeatRule(map[string]string{"repetition": "Athrá"})
	wr.IDOverride = "GA_WORD_REPEAT_RULE"
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	wrb := NewWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "Tosaíonn trí abairt as a chéile leis an bhfocal céanna.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))
	patterns.RegisterTokenSequences(lt, "ga", []patterns.TokenSequenceSpec{
		{ID: "GA_AGUS_AGUS", Tokens: []string{"agus", "agus"}, Message: "Athrá indéanta ar 'agus'.", Suggestion: "agus"},
	})

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
}
