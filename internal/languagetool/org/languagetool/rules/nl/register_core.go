package nl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// RegisterCoreDutchRules installs shared layout + NL word-repeat + beginning.
func RegisterCoreDutchRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "nl")
	wr := NewWordRepeatRule(map[string]string{"repetition": "Woordherhaling"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	wrb := NewWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "Drie opeenvolgende zinnen beginnen met hetzelfde woord.",
		"desc_repetition_beginning_adv":  "Drie opeenvolgende zinnen beginnen met hetzelfde bijwoord.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))

	patterns.RegisterTokenSequences(lt, "nl", []patterns.TokenSequenceSpec{
		{ID: "NL_ALS_OF", Tokens: []string{"als", "of"}, Message: "Bedoelde u 'alsof'?", Suggestion: "alsof"},
	})

	// Official replace + coherency tables (embedded from upstream).
	sr := NewSimpleReplaceRule(nil)
	lt.AddRuleChecker(sr.GetID(), rules.AsSentenceCheckerSimple(sr.Match))
	wc := NewWordCoherencyRule(nil)
	lt.AddTextLevelRuleChecker(wc.GetID(), rules.AsTextLevelChecker(wc.MatchList))
	// Official compounds + spaced compound detection.
	cr := NewCompoundRule(nil)
	lt.AddRuleChecker(cr.GetID(), rules.AsSentenceCheckerSimple(cr.Match))
	sc := NewSpaceInCompoundRule(nil)
	lt.AddRuleChecker(sc.GetID(), rules.AsSentenceCheckerSimple(sc.Match))

	// Official wrong-word-in-context + check-case.
	ww := NewDutchWrongWordInContextRule(nil)
	lt.AddRuleChecker(ww.GetID(), rules.AsSentenceCheckerSimple(ww.Match))
	cc := NewCheckCaseRule(nil)
	lt.AddRuleChecker(cc.GetID(), rules.AsSentenceCheckerSimple(cc.Match))
}
