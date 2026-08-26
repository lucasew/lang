package ca

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestAbstractSimpleReplaceLemmasRule(t *testing.T) {
	lemma := "anar"
	pos := "VMN0000"
	r := &AbstractSimpleReplaceLemmasRule{
		WrongLemmas: map[string][]string{"anar": {"caminar"}},
		Synthesize: func(lem, tag string) []string {
			if lem == "caminar" {
				return []string{"caminar"}
			}
			return nil
		},
	}
	tok := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("vaig", &pos, &lemma))
	tok.SetStartPos(0)
	ss := languagetool.SentenceStartTagName
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", &ss, nil)),
		tok,
	})
	matches := r.Match(sent)
	require.Len(t, matches, 1)
	require.Contains(t, matches[0].GetSuggestedReplacements(), "caminar")
}
