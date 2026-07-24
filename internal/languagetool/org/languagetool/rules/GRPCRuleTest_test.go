package rules

// Twin of GRPCRuleTest
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestGRPCRule_Match(t *testing.T) {
	cfg := NewRemoteRuleConfig()
	cfg.RuleID = "EXAMPLE"
	cfg.Type = GRPCRuleConfigType
	g := NewGRPCRule("en", cfg)
	g.MatchSentences = func(sentences []*languagetool.AnalyzedSentence) [][]*RuleMatch {
		out := make([][]*RuleMatch, len(sentences))
		for i, s := range sentences {
			out[i] = []*RuleMatch{NewRuleMatch(NewFakeRule("EXAMPLE"), s, 0, 1, "hit")}
		}
		return out
	}
	sent := languagetool.AnalyzePlain("x")
	ms := g.MatchRemote([]*languagetool.AnalyzedSentence{sent})
	require.Len(t, ms, 1)
	require.Equal(t, 0, ms[0].FromPos)
}

func TestGRPCRule_MaxLength(t *testing.T) {
	cfg := NewRemoteRuleConfig()
	cfg.RuleID = "LONG"
	cfg.BaseTimeoutMilliseconds = 100
	g := NewGRPCRule("en", cfg)
	// Batching: many short sentences still processed
	g.BatchSize = 2
	n := 0
	g.MatchSentences = func(sentences []*languagetool.AnalyzedSentence) [][]*RuleMatch {
		n += len(sentences)
		out := make([][]*RuleMatch, len(sentences))
		return out
	}
	var sents []*languagetool.AnalyzedSentence
	for i := 0; i < 5; i++ {
		sents = append(sents, languagetool.AnalyzePlain("a"))
	}
	_ = g.MatchRemote(sents)
	require.Equal(t, 5, n)
}
