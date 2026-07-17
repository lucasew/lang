package sv

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// RegisterCoreSwedishRules installs shared layout + language word-repeat + beginning.
func RegisterCoreSwedishRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "sv")
	wr := NewWordRepeatRule(map[string]string{"repetition": "Ordrepetition"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	wrb := NewWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "Tre meningar i rad börjar med samma ord.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))

	patterns.RegisterTokenSequences(lt, "sv", []patterns.TokenSequenceSpec{
		{ID: "SV_I_I", Tokens: []string{"i", "i"}, Message: "Möjlig upprepning av 'i'.", Suggestion: "i"},
	})

	// Official coherency.txt (embedded from upstream).
	wc := NewWordCoherencyRule(nil)
	lt.AddTextLevelRuleChecker(wc.GetID(), rules.AsTextLevelChecker(wc.MatchList))
}
