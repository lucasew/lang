package patterns

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestFalseFriendPatternRule(t *testing.T) {
	r := NewFalseFriendPatternRule("FF", "en", []*PatternToken{Token("gift")}, "desc", "msg", "short")
	require.Equal(t, "FF", r.GetID())
	require.True(t, r.HasTag(rules.TagPicky))
	require.Len(t, r.Tokens, 1)
}

func TestPatternRule_EstimateContextForSureMatch(t *testing.T) {
	// skip sums (InsideMarker default true → no outside-marker count)
	a := Token("a")
	a.SetSkipNext(2)
	b := Token("b")
	pr := NewPatternRule("E", "en", []*PatternToken{a, b}, "d", "m", "")
	require.Equal(t, 2, pr.EstimateContextForSureMatch())

	// skip=-1 → -1
	c := Token("c")
	c.SetSkipNext(-1)
	pr2 := NewPatternRule("E2", "en", []*PatternToken{c}, "d", "m", "")
	require.Equal(t, -1, pr2.EstimateContextForSureMatch())

	// SENT_END postag adds 1
	end := Token("")
	end.SetPosToken(PosToken{PosTag: languagetool.SentenceEndTagName})
	pr3 := NewPatternRule("E3", "en", []*PatternToken{end}, "d", "m", "")
	require.Equal(t, 1, pr3.EstimateContextForSureMatch())

	// antipattern length + skip
	apTok := Token("x")
	apTok.SetSkipNext(3)
	ap := NewPatternRule("AP", "en", []*PatternToken{apTok, Token("y")}, "d", "m", "")
	base := NewPatternRule("BASE", "en", []*PatternToken{Token("z")}, "d", "m", "")
	base.AntiPatterns = []*PatternRule{ap}
	// extendAfterMarker=0 + max(2, 2+3)=5 → 5
	require.Equal(t, 5, base.EstimateContextForSureMatch())
}

func TestPatternRule_GetMatchType(t *testing.T) {
	pr := NewPatternRule("T", "en", []*PatternToken{Token("a")}, "d", "m", "")
	require.Equal(t, rules.RuleMatchTypeOther, pr.GetMatchType())

	pr.IssueType = string(rules.ITSStyle)
	require.Equal(t, rules.RuleMatchTypeHint, pr.GetMatchType())

	pr.IssueType = string(rules.ITSGrammar)
	require.Equal(t, rules.RuleMatchTypeOther, pr.GetMatchType())

	pr.SetMatchType(rules.RuleMatchTypeUnknownWord)
	require.Equal(t, rules.RuleMatchTypeUnknownWord, pr.GetMatchType())
}

func TestPatternRuleMatcher_SetsMatchType(t *testing.T) {
	pr := NewPatternRule("STYLE", "en", []*PatternToken{Token("foo")}, "d", "m", "")
	pr.IssueType = string(rules.ITSStyle)
	ms, err := pr.Match(testSentence(atr("foo", 0)))
	require.NoError(t, err)
	require.Len(t, ms, 1)
	require.Equal(t, rules.RuleMatchTypeHint, ms[0].GetType())
}
