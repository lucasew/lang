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

// testPOS builds AnalyzedTokenReadings with a POS tag at start.
func atrPOS(token, pos string, start int) *languagetool.AnalyzedTokenReadings {
	return languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken(token, strPtr(pos), strPtr(token)), start)
}

// TestPatternRuleMatcher_UnificationAgreement ports Java testUnification semantics:
// <unify> requires shared feature type across tokens; negate="yes" fires on mismatch.
func TestPatternRuleMatcher_UnificationAgreement(t *testing.T) {
	cfg := NewUnifierConfiguration()
	cfg.SetEquivalence("number", "sg", func() *PatternToken {
		pt := NewPatternToken("", false, false, false)
		pt.SetPosToken(PosToken{PosTag: "NN", Regexp: false})
		return pt
	}())
	cfg.SetEquivalence("number", "pl", func() *PatternToken {
		pt := NewPatternToken("", false, false, false)
		pt.SetPosToken(PosToken{PosTag: "NNS", Regexp: false})
		return pt
	}())

	mkTok := func(pos string, last, neg bool) *PatternToken {
		pt := NewPatternToken("", false, false, false)
		pt.SetPosToken(PosToken{PosTag: pos, Regexp: false})
		pt.SetUnification(map[string][]string{"number": {}})
		if last {
			pt.SetLastInUnification()
		}
		if neg {
			pt.SetUniNegation()
		}
		return pt
	}

	// Positive unify: match only when both tokens share number (both sg).
	t1, t2 := mkTok("NN", false, false), mkTok("NN", true, false)
	agree := NewPatternRule("AGREE", "en", []*PatternToken{t1, t2}, "d", "agree", "")
	agree.UnifierConfig = cfg

	// cat(NN) + sits(VBZ) — second token pattern is NN so won't surface-match; use two NN.
	sg1 := atrPOS("cat", "NN", 0)
	sg2 := atrPOS("dog", "NN", 4)
	sentAgree := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{sg1, sg2})
	ms, err := agree.Match(sentAgree)
	require.NoError(t, err)
	require.Len(t, ms, 1, "same number should unify and match")

	// cat(NN) + dogs(NNS) — second pattern token is NN, won't match surface.
	// Change second pattern to accept both via empty postag / separate patterns.
	// Use two POS-open tokens with only unify constraining number via equivalence
	// on readings: pattern postag matches both NN and NNS via regexp.
	p1 := NewPatternToken("", false, false, false)
	p1.SetPosToken(PosToken{PosTag: "NN.*", Regexp: true})
	p1.SetUnification(map[string][]string{"number": {}})
	p2 := NewPatternToken("", false, false, false)
	p2.SetPosToken(PosToken{PosTag: "NN.*", Regexp: true})
	p2.SetUnification(map[string][]string{"number": {}})
	p2.SetLastInUnification()
	agree2 := NewPatternRule("AGREE2", "en", []*PatternToken{p1, p2}, "d", "agree2", "")
	agree2.UnifierConfig = cfg

	pl := atrPOS("dogs", "NNS", 4)
	sentDisagree := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{sg1, pl})
	ms, err = agree2.Match(sentDisagree)
	require.NoError(t, err)
	require.Empty(t, ms, "different number must not unify without negate")

	ms, err = agree2.Match(sentAgree)
	require.NoError(t, err)
	require.Len(t, ms, 1, "same number still matches")

	// Negated unify: fire when tokens do NOT share number.
	n1 := NewPatternToken("", false, false, false)
	n1.SetPosToken(PosToken{PosTag: "NN.*", Regexp: true})
	n1.SetUnification(map[string][]string{"number": {}})
	n2 := NewPatternToken("", false, false, false)
	n2.SetPosToken(PosToken{PosTag: "NN.*", Regexp: true})
	n2.SetUnification(map[string][]string{"number": {}})
	n2.SetLastInUnification()
	n2.SetUniNegation()
	negRule := NewPatternRule("NEG", "en", []*PatternToken{n1, n2}, "d", "neg", "")
	negRule.UnifierConfig = cfg

	ms, err = negRule.Match(sentDisagree)
	require.NoError(t, err)
	require.Len(t, ms, 1, "negate unify should fire on number mismatch")

	ms, err = negRule.Match(sentAgree)
	require.NoError(t, err)
	require.Empty(t, ms, "negate unify must not fire when numbers agree")

	// Fail-closed: unify without config never matches.
	noCfg := NewPatternRule("NOCFG", "en", []*PatternToken{p1, p2}, "d", "x", "")
	ms, err = noCfg.Match(sentAgree)
	require.NoError(t, err)
	require.Empty(t, ms)
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
