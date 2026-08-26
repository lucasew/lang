package patterns

// Twin of AbstractPatternRulePerformerTest
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of AbstractPatternRulePerformerTest.testTestAllReadings
func TestAbstractPatternRulePerformer_TestAllReadings(t *testing.T) {
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("hello", nil, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("world", nil, nil), 6),
	}
	sent := testSentence(toks...)
	rule := NewAbstractTokenBasedRule("R", "d", "en", []*PatternToken{Token("hello"), Token("world")})
	p := NewAbstractPatternRulePerformer(rule, nil)
	var hits int
	p.DoMatch(sent, func(tokenPositions []int, first, last, fm, lm int, _ []*languagetool.AnalyzedTokenReadings) {
		hits++
		// index 0 is SENT_START; content starts at 1
		require.Equal(t, 1, first)
		require.Equal(t, 2, last)
	})
	require.Equal(t, 1, hits)
}

// Port of AbstractPatternRulePerformerTest.testTestAllReadingsWithChunks
func TestAbstractPatternRulePerformer_TestAllReadingsWithChunks(t *testing.T) {
	// soft: multiword-like two-token pattern still matches
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("New", nil, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("York", nil, nil), 4),
	}
	sent := testSentence(toks...)
	rule := NewAbstractTokenBasedRule("R", "d", "en", []*PatternToken{Token("New"), Token("York")})
	p := NewAbstractPatternRulePerformer(rule, NewUnifier(nil, nil))
	var hits int
	p.DoMatch(sent, func(tokenPositions []int, first, last, fm, lm int, _ []*languagetool.AnalyzedTokenReadings) { hits++ })
	require.Equal(t, 1, hits)
}
