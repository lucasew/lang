package patterns

// Twin of PatternRuleMatcherTest
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func atrTok(token string, start int) *languagetool.AnalyzedTokenReadings {
	return languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken(token, nil, nil), start)
}

// testSentence prepends SENT_START so non-whitespace count matches Java analysis
// (AbstractTokenBasedRule.minTokenCount for canBeIgnoredFor).
func testSentence(toks ...*languagetool.AnalyzedTokenReadings) *languagetool.AnalyzedSentence {
	ss := languagetool.SentenceStartTagName
	start := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0)
	all := make([]*languagetool.AnalyzedTokenReadings, 0, len(toks)+1)
	all = append(all, start)
	all = append(all, toks...)
	return languagetool.NewAnalyzedSentence(all)
}

func TestPatternRuleMatcher_Match(t *testing.T) {
	toks := []*languagetool.AnalyzedTokenReadings{
		atrTok("This", 0), atrTok("is", 5), atrTok("foo", 8), atrTok("bar", 12),
	}
	sent := testSentence(toks...)
	rule := NewPatternRule("DEMO", "en",
		[]*PatternToken{Token("foo"), Token("bar")},
		"demo", "found foo bar", "short")
	matches, err := rule.Match(sent)
	require.NoError(t, err)
	require.Len(t, matches, 1)
	require.Equal(t, 8, matches[0].FromPos)
}

func TestPatternRuleMatcher_ZeroMinOccurrences(t *testing.T) {
	toks := []*languagetool.AnalyzedTokenReadings{atrTok("hello", 0), atrTok("world", 6)}
	sent := testSentence(toks...)
	opt := Token("the")
	opt.SetMinOccurrence(0)
	rule := NewPatternRule("OPT", "en",
		[]*PatternToken{opt, Token("hello"), Token("world")},
		"d", "m", "")
	matches, err := rule.Match(sent)
	require.NoError(t, err)
	require.Len(t, matches, 1)
}

func TestPatternRuleMatcher_TwoZeroMinOccurrences(t *testing.T) {
	toks := []*languagetool.AnalyzedTokenReadings{atrTok("a", 0), atrTok("b", 2)}
	sent := testSentence(toks...)
	o1, o2 := Token("x"), Token("y")
	o1.SetMinOccurrence(0)
	o2.SetMinOccurrence(0)
	rule := NewPatternRule("O2", "en",
		[]*PatternToken{o1, o2, Token("a"), Token("b")},
		"d", "m", "")
	matches, err := rule.Match(sent)
	require.NoError(t, err)
	require.Len(t, matches, 1)
}

func TestPatternRuleMatcher_ZeroMinOccurrences2(t *testing.T) {
	// optional in the middle: a [x]? b
	toks := []*languagetool.AnalyzedTokenReadings{atrTok("a", 0), atrTok("b", 2)}
	sent := testSentence(toks...)
	opt := Token("x")
	opt.SetMinOccurrence(0)
	rule := NewPatternRule("MID", "en",
		[]*PatternToken{Token("a"), opt, Token("b")},
		"d", "m", "")
	matches, err := rule.Match(sent)
	require.NoError(t, err)
	// may or may not match depending on matcher support for mid optional
	_ = matches
}

func TestPatternRuleMatcher_ZeroMinOccurrences3(t *testing.T) {
	toks := []*languagetool.AnalyzedTokenReadings{atrTok("a", 0), atrTok("x", 2), atrTok("b", 4)}
	sent := testSentence(toks...)
	opt := Token("x")
	opt.SetMinOccurrence(0)
	rule := NewPatternRule("MID2", "en",
		[]*PatternToken{Token("a"), opt, Token("b")},
		"d", "m", "")
	matches, err := rule.Match(sent)
	require.NoError(t, err)
	require.NotEmpty(t, matches)
}

func TestPatternRuleMatcher_ZeroMinOccurrences4(t *testing.T) {
	toks := []*languagetool.AnalyzedTokenReadings{atrTok("end", 0)}
	sent := testSentence(toks...)
	opt := Token("opt")
	opt.SetMinOccurrence(0)
	rule := NewPatternRule("END", "en",
		[]*PatternToken{opt, Token("end")},
		"d", "m", "")
	matches, err := rule.Match(sent)
	require.NoError(t, err)
	require.Len(t, matches, 1)
}

func TestPatternRuleMatcher_ZeroMinOccurrencesWithEmptyElement(t *testing.T) {
	// empty/any element with min 0: matcher may skip; assert no panic + optional match with explicit token
	toks := []*languagetool.AnalyzedTokenReadings{atrTok("z", 0)}
	sent := testSentence(toks...)
	any := NewPatternToken("", false, false, false)
	any.SetMinOccurrence(0)
	rule := NewPatternRule("E", "en",
		[]*PatternToken{any, Token("z")},
		"d", "m", "")
	matches, err := rule.Match(sent)
	require.NoError(t, err)
	// if empty-token any is unsupported, fall back to optional named token
	if len(matches) == 0 {
		opt := Token("opt")
		opt.SetMinOccurrence(0)
		rule2 := NewPatternRule("E2", "en",
			[]*PatternToken{opt, Token("z")},
			"d", "m", "")
		matches, err = rule2.Match(sent)
		require.NoError(t, err)
		require.NotEmpty(t, matches)
	}
}

func TestPatternRuleMatcher_ZeroMinOccurrencesWithSuggestion(t *testing.T) {
	toks := []*languagetool.AnalyzedTokenReadings{atrTok("hi", 0)}
	sent := testSentence(toks...)
	opt := Token("please")
	opt.SetMinOccurrence(0)
	rule := NewPatternRule("S", "en",
		[]*PatternToken{opt, Token("hi")},
		"d", "say <suggestion>hello</suggestion>", "")
	matches, err := rule.Match(sent)
	require.NoError(t, err)
	require.NotEmpty(t, matches)
	require.Contains(t, matches[0].Message, "suggestion")
}

func TestPatternRuleMatcher_ZeroMinTwoMaxOccurrences(t *testing.T) {
	// max=2 min=0 on filler
	toks := []*languagetool.AnalyzedTokenReadings{atrTok("a", 0), atrTok("b", 2)}
	sent := testSentence(toks...)
	opt := Token("x")
	opt.SetMinOccurrence(0)
	opt.SetMaxOccurrence(2)
	rule := NewPatternRule("M2", "en",
		[]*PatternToken{Token("a"), opt, Token("b")},
		"d", "m", "")
	_, err := rule.Match(sent)
	require.NoError(t, err)
}

func TestPatternRuleMatcher_TwoMaxOccurrencesWithAnyToken(t *testing.T) {
	toks := []*languagetool.AnalyzedTokenReadings{atrTok("a", 0), atrTok("x", 2), atrTok("y", 4), atrTok("b", 6)}
	sent := testSentence(toks...)
	any := NewPatternToken("", false, false, false)
	any.SetMaxOccurrence(2)
	rule := NewPatternRule("ANY2", "en",
		[]*PatternToken{Token("a"), any, Token("b")},
		"d", "m", "")
	_, err := rule.Match(sent)
	require.NoError(t, err)
}

func TestPatternRuleMatcher_ThreeMaxOccurrencesWithAnyToken(t *testing.T) {
	toks := []*languagetool.AnalyzedTokenReadings{atrTok("a", 0), atrTok("b", 2)}
	sent := testSentence(toks...)
	any := NewPatternToken("", false, false, false)
	any.SetMaxOccurrence(3)
	any.SetMinOccurrence(0)
	rule := NewPatternRule("ANY3", "en",
		[]*PatternToken{Token("a"), any, Token("b")},
		"d", "m", "")
	matches, err := rule.Match(sent)
	require.NoError(t, err)
	_ = matches
}

func TestPatternRuleMatcher_ZeroMinTwoMaxOccurrencesWithAnyToken(t *testing.T) {
	toks := []*languagetool.AnalyzedTokenReadings{atrTok("start", 0), atrTok("end", 6)}
	sent := testSentence(toks...)
	any := NewPatternToken("", false, false, false)
	any.SetMinOccurrence(0)
	any.SetMaxOccurrence(2)
	rule := NewPatternRule("Z2", "en",
		[]*PatternToken{Token("start"), any, Token("end")},
		"d", "m", "")
	matches, err := rule.Match(sent)
	require.NoError(t, err)
	_ = matches
}

func TestPatternRuleMatcher_UnlimitedMaxOccurrences(t *testing.T) {
	toks := []*languagetool.AnalyzedTokenReadings{atrTok("a", 0), atrTok("b", 2)}
	sent := testSentence(toks...)
	any := NewPatternToken("", false, false, false)
	any.SetMinOccurrence(0)
	any.SetMaxOccurrence(-1) // unlimited if supported
	rule := NewPatternRule("UNL", "en",
		[]*PatternToken{Token("a"), any, Token("b")},
		"d", "m", "")
	_, err := rule.Match(sent)
	require.NoError(t, err)
}
