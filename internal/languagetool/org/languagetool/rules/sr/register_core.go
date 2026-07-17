package sr

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/sr/ekavian"
)

// RegisterCoreSerbianRules installs shared layout + ekavian official replace tables.
// Jekavian tables exist for dialect-specific tools; soft default uses ekavian.
func RegisterCoreSerbianRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "sr")
	wr := rules.NewWordRepeatRule(map[string]string{"repetition": "Понављање речи"})
	wr.IDOverride = "SR_WORD_REPEAT_RULE"
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	wrb := rules.NewWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "Три узастопне реченице почињу истом речју.",
	})
	wrb.IDOverride = "SR_WORD_REPEAT_BEGINNING_RULE"
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))

	// Official replace-grammar.txt / replace-style.txt (embedded from upstream ekavian).
	gr := ekavian.NewSimpleGrammarEkavianReplaceRule(nil)
	lt.AddRuleChecker(gr.GetID(), rules.AsSentenceCheckerSimple(gr.Match))
	st := ekavian.NewSimpleStyleEkavianReplaceRule(nil)
	lt.AddRuleChecker(st.GetID(), rules.AsSentenceCheckerSimple(st.Match))
}
