package gl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// RegisterCoreGalicianRules installs shared layout + language word-repeat + beginning.
func RegisterCoreGalicianRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "gl")
	wr := NewWordRepeatRule(map[string]string{"repetition": "Repetición"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	wrb := NewWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "Tres frases sucesivas comezan coa mesma palabra.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))
	patterns.RegisterTokenSequences(lt, "gl", []patterns.TokenSequenceSpec{
		{ID: "GL_A_O", Tokens: []string{"a", "o"}, Message: "Quizais 'ao'?", Suggestion: "ao"},
		{ID: "GL_DE_O", Tokens: []string{"de", "o"}, Message: "Quizais 'do'?", Suggestion: "do"},
	})

	// Official replace.txt (embedded from upstream).
	sr := NewSimpleReplaceRule(nil)
	lt.AddRuleChecker(sr.GetID(), rules.AsSentenceCheckerSimple(sr.Match))
}
