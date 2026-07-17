package ca

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// RegisterCoreCatalanRules installs shared layout + Catalan word-repeat + beginning.
func RegisterCoreCatalanRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "ca")
	wr := NewCatalanWordRepeatRule(map[string]string{"repetition": "Repetició de paraula"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	wrb := NewCatalanWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "Tres frases successives comencen amb la mateixa paraula.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))

	patterns.RegisterTokenSequences(lt, "ca", []patterns.TokenSequenceSpec{
		{ID: "CA_A_EL", Tokens: []string{"a", "el"}, Message: "Volíeu dir 'al'?", Suggestion: "al"},
		{ID: "CA_DE_EL", Tokens: []string{"de", "el"}, Message: "Volíeu dir 'del'?", Suggestion: "del"},
	})

	// Official replace + coherency tables (embedded from upstream).
	sr := NewSimpleReplaceRule(nil)
	lt.AddRuleChecker(sr.GetID(), rules.AsSentenceCheckerSimple(sr.Match))
	wc := NewWordCoherencyRule(nil)
	lt.AddTextLevelRuleChecker(wc.GetID(), rules.AsTextLevelChecker(wc.MatchList))
}
