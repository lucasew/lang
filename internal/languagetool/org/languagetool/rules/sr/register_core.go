package sr

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/sr/ekavian"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/sr/jekavian"
)

// serbianJekavian reports country variants that use JekavianSerbian in Java
// (BA / HR / ME). Bare "sr" and "sr-RS" remain Ekavian (Java Serbian / SerbianSerbian).
// Note: Java JekavianSerbian with empty countries is also short-code "sr" and cannot
// be distinguished by language code alone — code-based dispatch uses ekavian for bare sr.
func serbianJekavian(langCode string) bool {
	lc := strings.ToLower(langCode)
	switch {
	case strings.Contains(lc, "-ba") || strings.HasSuffix(lc, "_ba"):
		return true
	case strings.Contains(lc, "-hr") || strings.HasSuffix(lc, "_hr"):
		return true
	case strings.Contains(lc, "-me") || strings.HasSuffix(lc, "_me"):
		return true
	default:
		return false
	}
}

// RegisterCoreSerbianRules installs shared layout + dialect replace tables + Morfologik speller.
// Java Serbian.getRelevantRules → Ekavian; JekavianSerbian / BA / HR / ME → Jekavian.
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

	if serbianJekavian(lt.GetLanguageCode()) {
		// Java JekavianSerbian.getRelevantRules: jekavian replace + MorfologikJekavianSpellerRule.
		gr := jekavian.NewSimpleGrammarJekavianReplaceRule(nil)
		lt.AddRuleChecker(gr.GetID(), rules.AsSentenceCheckerSimple(gr.Match))
		st := jekavian.NewSimpleStyleJekavianReplaceRule(nil)
		lt.AddRuleChecker(st.GetID(), rules.AsSentenceCheckerSimple(st.Match))
		// Always full Match; word lists from dictionary/jekavian/{ignored,spelling,prohibit}.txt
		sp := jekavian.NewMorfologikJekavianSpellerRule()
		lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))
		return
	}

	// Java Serbian.getRelevantRules: ekavian replace + MorfologikEkavianSpellerRule.
	gr := ekavian.NewSimpleGrammarEkavianReplaceRule(nil)
	lt.AddRuleChecker(gr.GetID(), rules.AsSentenceCheckerSimple(gr.Match))
	st := ekavian.NewSimpleStyleEkavianReplaceRule(nil)
	lt.AddRuleChecker(st.GetID(), rules.AsSentenceCheckerSimple(st.Match))
	// Always full Match; word lists from dictionary/ekavian/{ignored,spelling,prohibit}.txt
	// (binary dict optional — map Words fail-closed without invent).
	sp := ekavian.NewMorfologikEkavianSpellerRule()
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))
}
