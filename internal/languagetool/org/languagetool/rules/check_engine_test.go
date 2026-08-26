package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestCheckEngine(t *testing.T) {
	eng := NewCheckEngine("en")
	eng.Rules = []SentenceRule{
		SentenceRuleFunc(func(s *languagetool.AnalyzedSentence) ([]*RuleMatch, error) {
			var out []*RuleMatch
			for _, tok := range s.GetTokensWithoutWhitespace() {
				if tok.GetToken() == "bar" {
					out = append(out, NewRuleMatch(NewFakeRule("DEMO"), s, tok.GetStartPos(), tok.GetEndPos(), "found bar"))
				}
			}
			return out, nil
		}),
	}
	ms, err := eng.CheckText("foo bar baz")
	require.NoError(t, err)
	require.Len(t, ms, 1)
	require.Equal(t, "found bar", ms[0].Message)
}
