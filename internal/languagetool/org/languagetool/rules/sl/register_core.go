package sl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// RegisterCoreSlovenianRules installs shared layout + Slovenian word-repeat + beginning.
func RegisterCoreSlovenianRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "sl")
	wr := rules.NewWordRepeatRule(map[string]string{"repetition": "Ponovitev besede"})
	wr.IDOverride = "SL_WORD_REPEAT_RULE"
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	wrb := NewWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "Tri zaporedne povedi se začnejo z isto besedo.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))
	patterns.RegisterTokenSequences(lt, "sl", []patterns.TokenSequenceSpec{
		{ID: "SL_V_V", Tokens: []string{"v", "v"}, Message: "Možna ponovitev predloga 'v'.", Suggestion: "v"},
	})
}
