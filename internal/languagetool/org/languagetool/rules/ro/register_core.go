package ro

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// RegisterCoreRomanianRules installs shared layout + language word-repeat + beginning.
func RegisterCoreRomanianRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "ro")
	wr := NewWordRepeatRule(map[string]string{"repetition": "Repetiție"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	wrb := NewRomanianWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "Trei propoziții consecutive încep cu același cuvânt.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))
	patterns.RegisterTokenSequences(lt, "ro", []patterns.TokenSequenceSpec{
		{ID: "RO_DE_DE", Tokens: []string{"de", "de"}, Message: "Posibilă repetiție a prepoziției 'de'.", Suggestion: "de"},
	})
}
