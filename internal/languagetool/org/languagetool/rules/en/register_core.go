package en

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCoreEnglishLanguageRules installs shared layout + EN-specific word-repeat + a/an + phrases.
func RegisterCoreEnglishLanguageRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "en")
	wr := NewEnglishWordRepeatRule(map[string]string{"repetition": "Word repetition"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	lt.AddRuleChecker("EN_A_VS_AN", languagetool.SimpleAvsAnChecker())
	lt.AddRuleChecker("PHRASE_REPLACE", languagetool.SimplePhraseReplaceChecker("PHRASE_REPLACE", map[string]string{
		"tot he": "to the",
	}))
}
