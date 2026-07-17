package el

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// RegisterCoreGreekRules installs shared layout + language word-repeat + beginning.
func RegisterCoreGreekRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "el")
	wr := NewWordRepeatRule(map[string]string{"repetition": "Επανάληψη"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	wrb := NewGreekWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "Τρεις διαδοχικές προτάσεις αρχίζουν με την ίδια λέξη.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))
	patterns.RegisterTokenSequences(lt, "el", []patterns.TokenSequenceSpec{
		{ID: "EL_ΚΑΙ_ΚΑΙ", Tokens: []string{"και", "και"}, Message: "Πιθανή επανάληψη του «και».", Suggestion: "και"},
	})
}
