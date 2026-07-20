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

// registerSerbianBasicRules ports Serbian.getBasicRules (shared by Ekavian/Jekavian).
func registerSerbianBasicRules(lt *languagetool.JLanguageTool) {
	cw := rules.NewCommaWhitespaceRule(nil)
	lt.AddRuleChecker(cw.GetID(), rules.AsSentenceCheckerSimple(cw.Match))

	dp := rules.NewDoublePunctuationRule(nil)
	lt.AddRuleChecker(dp.GetID(), rules.AsSentenceCheckerSimple(dp.Match))

	// Java GenericUnpairedBracketsRule symbols for Serbian (id UNPAIRED_BRACKETS).
	start := []string{"[", "(", "{", "„", "„", "\""}
	end := []string{"]", ")", "}", "”", "“", "\""}
	ub := rules.NewGenericUnpairedBracketsRule(nil, start, end)
	lt.AddTextLevelRuleChecker(ub.GetID(), rules.AsTextLevelChecker(ub.MatchList))

	up := rules.NewUppercaseSentenceStartRule(nil, "sr")
	lt.AddRuleChecker(up.GetID(), rules.AsSentenceCheckerSimple(func(s *languagetool.AnalyzedSentence) []*rules.RuleMatch {
		return up.MatchList([]*languagetool.AnalyzedSentence{s})
	}))

	ws := rules.NewMultipleWhitespaceRule(map[string]string{
		"whitespace_repetition": "Possible typo: you repeated a whitespace",
	})
	lt.AddRuleChecker(ws.GetID(), rules.AsSentenceCheckerSimple(func(s *languagetool.AnalyzedSentence) []*rules.RuleMatch {
		return ws.Match([]*languagetool.AnalyzedSentence{s})
	}))

	sw := rules.NewSentenceWhitespaceRule(nil)
	lt.AddTextLevelRuleChecker(sw.GetID(), rules.AsTextLevelChecker(sw.MatchList))

	// Java WordRepeatRule default id WORD_REPEAT_RULE — no beginning, no invent SR_.
	wr := rules.NewWordRepeatRule(map[string]string{"repetition": "Понављање речи"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
}

// RegisterCoreSerbianRules ports Serbian / JekavianSerbian.getRelevantRules.
// Java basic layout + dialect replace + Morfologik — no invent SharedLayout extras.
func RegisterCoreSerbianRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	registerSerbianBasicRules(lt)

	if serbianJekavian(lt.GetLanguageCode()) {
		gr := jekavian.NewSimpleGrammarJekavianReplaceRule(nil)
		lt.AddRuleChecker(gr.GetID(), rules.AsSentenceCheckerSimple(gr.Match))
		st := jekavian.NewSimpleStyleJekavianReplaceRule(nil)
		lt.AddRuleChecker(st.GetID(), rules.AsSentenceCheckerSimple(st.Match))
		sp := jekavian.NewMorfologikJekavianSpellerRule()
		lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))
		return
	}

	gr := ekavian.NewSimpleGrammarEkavianReplaceRule(nil)
	lt.AddRuleChecker(gr.GetID(), rules.AsSentenceCheckerSimple(gr.Match))
	st := ekavian.NewSimpleStyleEkavianReplaceRule(nil)
	lt.AddRuleChecker(st.GetID(), rules.AsSentenceCheckerSimple(st.Match))
	sp := ekavian.NewMorfologikEkavianSpellerRule()
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))
}
