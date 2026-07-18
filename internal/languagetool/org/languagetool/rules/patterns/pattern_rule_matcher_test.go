package patterns

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func atr(token string, start int) *languagetool.AnalyzedTokenReadings {
	r := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken(token, nil, nil), start)
	return r
}

func TestPatternRuleMatcherSimple(t *testing.T) {
	// tokens: This(0) is(5) foo(8) bar(12)
	toks := []*languagetool.AnalyzedTokenReadings{
		atr("This", 0),
		atr("is", 5),
		atr("foo", 8),
		atr("bar", 12),
	}
	// fix end positions roughly
	sent := languagetool.NewAnalyzedSentence(toks)
	rule := NewPatternRule("DEMO", "en",
		[]*PatternToken{Token("foo"), Token("bar")},
		"demo", "found foo bar", "short")
	matches, err := rule.Match(sent)
	require.NoError(t, err)
	require.Len(t, matches, 1)
	require.Equal(t, 8, matches[0].FromPos)
	// GetEndPos of bar
	require.Equal(t, toks[3].GetEndPos(), matches[0].ToPos)
	require.Equal(t, "found foo bar", matches[0].Message)
}

func TestPatternRuleMatcherOptional(t *testing.T) {
	toks := []*languagetool.AnalyzedTokenReadings{
		atr("hello", 0),
		atr("world", 6),
	}
	sent := languagetool.NewAnalyzedSentence(toks)
	opt := Token("the")
	opt.SetMinOccurrence(0)
	rule := NewPatternRule("OPT", "en",
		[]*PatternToken{opt, Token("hello"), Token("world")},
		"d", "m", "")
	matches, err := rule.Match(sent)
	require.NoError(t, err)
	require.Len(t, matches, 1)
}

func TestPatternRuleMatcherNoMatch(t *testing.T) {
	toks := []*languagetool.AnalyzedTokenReadings{atr("hello", 0)}
	sent := languagetool.NewAnalyzedSentence(toks)
	rule := NewPatternRule("X", "en", []*PatternToken{Token("bye")}, "d", "m", "")
	matches, err := rule.Match(sent)
	require.NoError(t, err)
	require.Empty(t, matches)
}

func TestRepeatedAndConsistencyTransformers(t *testing.T) {
	a1 := NewAbstractPatternRule("STYLE_A_feat1", "d", "en", nil, false)
	a2 := NewAbstractPatternRule("STYLE_A_feat2", "d", "en", nil, false)
	b := NewAbstractPatternRule("OTHER", "d", "en", nil, false)
	ct := NewConsistencyPatternRuleTransformer("en")
	rem, tr := TransformPatternRules([]*AbstractPatternRule{a1, a2, b}, ct)
	require.Len(t, rem, 1)
	require.Equal(t, "OTHER", rem[0].ID)
	require.Len(t, tr, 1)

	require.Equal(t, "STYLE_A", GetMainRuleId("STYLE_A_feat1"))
	require.Equal(t, "feat1", GetFeature("STYLE_A_feat1"))

	r1 := NewAbstractPatternRule("REP", "d", "en", nil, false)
	r1.DistanceTokens = 10
	r2 := NewAbstractPatternRule("REP", "d", "en", nil, false)
	rt := NewRepeatedPatternRuleTransformer("en")
	rem2, tr2 := TransformPatternRules([]*AbstractPatternRule{r1, r2}, rt)
	require.Empty(t, rem2)
	require.Len(t, tr2, 1)
}

// Soft path: optional min=0 must backtrack when a later element needs the token
// (NL FULL_SENTENCE_001 style: adj? noun after "de").
