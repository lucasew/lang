package ml

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCoreMalayalamRules installs shared layout + WordRepeat + Morfologik speller.
// Java Malayalam.getRelevantRules includes WordRepeatRule (default WORD_REPEAT_RULE id).
func RegisterCoreMalayalamRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "ml")
	wr := rules.NewWordRepeatRule(map[string]string{"repetition": "Word repetition"})
	// Java WordRepeatRule default id WORD_REPEAT_RULE — do not invent ML_WORD_REPEAT_RULE.
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))

	// Java createDefaultSpellingRule / Morfologik getId; CFSA2 when dict present.
	// Empty map shell fails closed when binary resource is missing (no invent).
	// Always full Match (fail-closed map Words when binary dict missing; no invent).
	sp := NewMorfologikMalayalamSpellerRule()
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))
}
