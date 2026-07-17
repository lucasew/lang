package da

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// RegisterCoreDanishRules installs shared layout + language word-repeat + beginning.
func RegisterCoreDanishRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "da")
	wr := NewWordRepeatRule(map[string]string{"repetition": "Gentagelse"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	wrb := NewWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "Tre sætninger i træk begynder med samme ord.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))

	patterns.RegisterTokenSequences(lt, "da", []patterns.TokenSequenceSpec{
		{ID: "DA_I_I", Tokens: []string{"i", "i"}, Message: "Mulig gentagelse af 'i'.", Suggestion: "i"},
	})
}
