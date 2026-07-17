package ar

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// RegisterCoreArabicRules installs shared layout + Arabic word-repeat + beginning.
func RegisterCoreArabicRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "ar")
	wr := NewArabicWordRepeatRule(map[string]string{"repetition": "تكرار كلمة"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	wrb := NewWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "ثلاث جمل متتالية تبدأ بنفس الكلمة.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))
	patterns.RegisterTokenSequences(lt, "ar", []patterns.TokenSequenceSpec{
		{ID: "AR_FI_FI", Tokens: []string{"في", "في"}, Message: "تكرار محتمل لحرف الجر «في».", Suggestion: "في"},
	})
}
