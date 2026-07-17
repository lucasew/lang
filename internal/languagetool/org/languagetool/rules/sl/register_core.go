package sl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCoreSlovenianRules installs shared layout + Slovenian word-repeat id.
func RegisterCoreSlovenianRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "sl")
	wr := rules.NewWordRepeatRule(map[string]string{"repetition": "Ponovitev besede"})
	wr.IDOverride = "SL_WORD_REPEAT_RULE"
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
}
