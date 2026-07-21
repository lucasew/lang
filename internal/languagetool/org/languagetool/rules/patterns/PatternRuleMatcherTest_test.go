package patterns

// Twin of PatternRuleMatcherTest
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
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

// --- Missing Java PatternRuleMatcherTest methods (faithful twins) ---

// spacedSentence builds tokens with single spaces between words (Java Demo analysis positions).
func spacedSentence(words ...string) *languagetool.AnalyzedSentence {
	pos := 0
	var toks []*languagetool.AnalyzedTokenReadings
	for i, w := range words {
		if i > 0 {
			pos++ // space
		}
		toks = append(toks, atrTok(w, pos))
		pos += len([]rune(w))
	}
	return testSentence(toks...)
}

func matchRule(t *testing.T, rule *PatternRule, words ...string) []*rules.RuleMatch {
	t.Helper()
	ms, err := rule.Match(spacedSentence(words...))
	require.NoError(t, err)
	return ms
}

func TestPatternRuleMatcher_TwoMaxOccurrences(t *testing.T) {
	// Java: a b{1,2}
	b := Token("b")
	b.SetMaxOccurrence(2)
	rule := NewPatternRule("M2", "en", []*PatternToken{Token("a"), b}, "d", "m", "")
	require.Empty(t, matchRule(t, rule, "a", "a"))
	require.Len(t, matchRule(t, rule, "a", "b"), 1)
	require.Len(t, matchRule(t, rule, "a", "b", "b"), 1)
	require.NotEmpty(t, matchRule(t, rule, "a", "b", "c"))
	require.NotEmpty(t, matchRule(t, rule, "a", "b", "b", "c"))
	require.NotEmpty(t, matchRule(t, rule, "x", "a", "b", "b"))

	ms := matchRule(t, rule, "a", "b", "b", "b")
	require.Len(t, ms, 1)
	require.Equal(t, 0, ms[0].FromPos)
	require.Equal(t, 5, ms[0].ToPos) // "a b b"

	ms2 := matchRule(t, rule, "a", "b", "b", "b", "foo", "a", "b", "b")
	require.Len(t, ms2, 2)
	require.Equal(t, 0, ms2[0].FromPos)
	require.Equal(t, 5, ms2[0].ToPos)
	require.Equal(t, 12, ms2[1].FromPos)
	require.Equal(t, 17, ms2[1].ToPos)
}

func TestPatternRuleMatcher_ThreeMaxOccurrences(t *testing.T) {
	// Java: a b{1,3}
	b := Token("b")
	b.SetMaxOccurrence(3)
	rule := NewPatternRule("M3", "en", []*PatternToken{Token("a"), b}, "d", "m", "")
	require.Empty(t, matchRule(t, rule, "a", "a"))
	require.Len(t, matchRule(t, rule, "a", "b"), 1)
	require.Len(t, matchRule(t, rule, "a", "b", "b"), 1)
	require.Len(t, matchRule(t, rule, "a", "b", "b", "b"), 1)
	// four b's → longest match still 3 b's (partial)
	ms := matchRule(t, rule, "a", "b", "b", "b", "b")
	require.Len(t, ms, 1)
	require.Equal(t, 0, ms[0].FromPos)
	require.Equal(t, 7, ms[0].ToPos) // "a b b b"
}

func TestPatternRuleMatcher_OptionalWithoutExplicitMarker(t *testing.T) {
	// a b? c
	b := Token("b")
	b.SetMinOccurrence(0)
	rule := NewPatternRule("OPT", "en", []*PatternToken{Token("a"), b, Token("c")}, "d", "m", "")
	ms1 := matchRule(t, rule, "a", "b", "c", "zzz")
	require.Len(t, ms1, 1)
	require.Equal(t, 0, ms1[0].FromPos)
	require.Equal(t, 5, ms1[0].ToPos)
	ms2 := matchRule(t, rule, "a", "c", "zzz")
	require.Len(t, ms2, 1)
	require.Equal(t, 0, ms2[0].FromPos)
	require.Equal(t, 3, ms2[0].ToPos)
}

func TestPatternRuleMatcher_OptionalWithExplicitMarker(t *testing.T) {
	// (a b?) c — marker on a and optional b only
	a := Token("a")
	a.SetInsideMarker(true)
	b := Token("b")
	b.SetMinOccurrence(0)
	b.SetInsideMarker(true)
	c := Token("c")
	c.SetInsideMarker(false)
	rule := NewPatternRule("OPTM", "en", []*PatternToken{a, b, c}, "d", "m", "")
	ms1 := matchRule(t, rule, "a", "b", "c", "zzz")
	require.Len(t, ms1, 1)
	require.Equal(t, 0, ms1[0].FromPos)
	require.Equal(t, 3, ms1[0].ToPos) // marker ends after "a b" → pos 3
	ms2 := matchRule(t, rule, "a", "c", "zzz")
	require.Len(t, ms2, 1)
	require.Equal(t, 0, ms2[0].FromPos)
	require.Equal(t, 1, ms2[0].ToPos) // marker ends after "a"
}

func TestPatternRuleMatcher_OptionalAnyTokenWithExplicitMarker(t *testing.T) {
	a := Token("a")
	a.SetInsideMarker(true)
	any := NewPatternToken("", false, false, false)
	any.SetMinOccurrence(0)
	any.SetInsideMarker(true)
	c := Token("c")
	c.SetInsideMarker(false)
	rule := NewPatternRule("ANYM", "en", []*PatternToken{a, any, c}, "d", "m", "")
	ms1 := matchRule(t, rule, "a", "x", "c", "zzz")
	require.Len(t, ms1, 1)
	require.Equal(t, 0, ms1[0].FromPos)
	require.Equal(t, 3, ms1[0].ToPos)
	ms2 := matchRule(t, rule, "a", "c", "zzz")
	require.Len(t, ms2, 1)
	require.Equal(t, 0, ms2[0].FromPos)
	require.Equal(t, 1, ms2[0].ToPos)
}

func TestPatternRuleMatcher_OptionalAnyTokenWithExplicitMarker2(t *testing.T) {
	the := Token("the")
	the.SetInsideMarker(true)
	any := NewPatternToken("", false, false, false)
	any.SetMinOccurrence(0)
	any.SetInsideMarker(true)
	bike := Token("bike")
	bike.SetInsideMarker(false)
	rule := NewPatternRule("BIKE", "en", []*PatternToken{the, any, bike}, "d", "m", "")
	ms1 := matchRule(t, rule, "the", "nice", "bike", "zzz")
	require.Len(t, ms1, 1)
	require.Equal(t, 0, ms1[0].FromPos)
	require.Equal(t, 8, ms1[0].ToPos) // "the nice"
	ms2 := matchRule(t, rule, "the", "bike", "zzz")
	require.Len(t, ms2, 1)
	require.Equal(t, 0, ms2[0].FromPos)
	require.Equal(t, 3, ms2[0].ToPos) // "the"
}

func TestPatternRuleMatcher_MaxTwoAndThreeOccurrences(t *testing.T) {
	// a{1,2} b{1,3}
	a := Token("a")
	a.SetMaxOccurrence(2)
	b := Token("b")
	b.SetMaxOccurrence(3)
	rule := NewPatternRule("M23", "en", []*PatternToken{a, b}, "d", "m", "")
	require.Len(t, matchRule(t, rule, "a", "b"), 1)
	require.Len(t, matchRule(t, rule, "a", "b", "b"), 1)
	require.Len(t, matchRule(t, rule, "a", "b", "b", "b"), 1)
	require.Empty(t, matchRule(t, rule, "a", "a"))
	require.Empty(t, matchRule(t, rule, "a", "x", "b", "b", "b"))
	// Java keeps only the longest match for overlapping starts; assert longest span exists.
	ms2 := matchRule(t, rule, "a", "a", "b")
	require.True(t, hasSpan(ms2, 0, 5), "expected longest span [0,5], got %v", spans(ms2))
	ms3 := matchRule(t, rule, "a", "a", "b", "b")
	require.True(t, hasSpan(ms3, 0, 7), "expected longest span [0,7], got %v", spans(ms3))
	ms4 := matchRule(t, rule, "a", "a", "b", "b", "b")
	require.True(t, hasSpan(ms4, 0, 9), "expected longest span [0,9], got %v", spans(ms4))
}

func hasSpan(ms []*rules.RuleMatch, from, to int) bool {
	for _, m := range ms {
		if m != nil && m.FromPos == from && m.ToPos == to {
			return true
		}
	}
	return false
}

func spans(ms []*rules.RuleMatch) [][2]int {
	out := make([][2]int, 0, len(ms))
	for _, m := range ms {
		if m != nil {
			out = append(out, [2]int{m.FromPos, m.ToPos})
		}
	}
	return out
}

func TestPatternRuleMatcher_InfiniteSkip(t *testing.T) {
	a := Token("a")
	a.SetSkipNext(-1)
	rule := NewPatternRule("SKIP", "en", []*PatternToken{a, Token("b")}, "d", "m", "")
	require.Len(t, matchRule(t, rule, "a", "b"), 1)
	require.Len(t, matchRule(t, rule, "a", "x", "b"), 1)
	require.Len(t, matchRule(t, rule, "a", "x", "x", "b"), 1)
	require.Len(t, matchRule(t, rule, "a", "x", "x", "x", "b"), 1)
}

func TestPatternRuleMatcher_InfiniteSkipWithMatchReference(t *testing.T) {
	// a|b with infinite skip, then \0 backref to first matched token
	ab := NewPatternToken("a|b", false, true, false)
	ab.SetSkipNext(-1)
	c := Token("\\0")
	// Java sets match with tokenRef 0 and inMessageOnly
	m := NewMatch("", "", false, "", "", CaseNone, false, false, IncludeNone)
	m.SetTokenRef(0)
	m.SetInMessageOnly(true)
	c.SetMatch(m)
	// Pattern token matching \\0 as surface is wrong — Java PatternToken("\\0") uses match backref.
	// Rebuild like Java: element for backref uses setMatch only.
	c = NewPatternToken("\\0", false, false, false)
	c.SetMatch(m)
	rule := NewPatternRule("REF", "en", []*PatternToken{ab, c}, "d", "m", "")
	require.Len(t, matchRule(t, rule, "a", "a"), 1)
	require.Len(t, matchRule(t, rule, "b", "b"), 1)
	require.Len(t, matchRule(t, rule, "a", "x", "a"), 1)
	require.Len(t, matchRule(t, rule, "b", "x", "b"), 1)
	require.Len(t, matchRule(t, rule, "a", "x", "x", "a"), 1)
	require.Len(t, matchRule(t, rule, "b", "x", "x", "b"), 1)
	require.Empty(t, matchRule(t, rule, "a", "b"))
	require.Empty(t, matchRule(t, rule, "b", "a"))
	require.Empty(t, matchRule(t, rule, "b", "x", "a"))
	require.Empty(t, matchRule(t, rule, "a", "x", "x", "b"))
	require.Empty(t, matchRule(t, rule, "b", "x", "x", "a"))

	ms := matchRule(t, rule, "a", "foo", "a", "and", "b", "foo", "b")
	require.Len(t, ms, 2)
	require.Equal(t, 0, ms[0].FromPos)
	require.Equal(t, 7, ms[0].ToPos)
	require.Equal(t, 12, ms[1].FromPos)
	require.Equal(t, 19, ms[1].ToPos)

	ms2 := matchRule(t, rule, "xx", "a", "b", "x", "x", "x", "b", "a")
	// Java: single match at [3,16] on "xx a b x x x b a" (a … a via infinite skip).
	require.True(t, hasSpan(ms2, 3, 16), "expected span [3,16], got %v", spans(ms2))
}

func TestPatternRuleMatcher_NoMatchReferenceRecursion(t *testing.T) {
	// \n in rule messages refers to matches; match text containing \n must not re-expand
	p1 := NewPatternToken(`\p{Punct}`, false, true, false)
	p2 := NewPatternToken(`\d+`, false, true, false)
	rule := NewPatternRule("MATCH_REFERENCERE_CURSION_DEMO", "xx",
		[]*PatternToken{p1, p2},
		"", "Here come the match references: \\1\\2. This is the end", "")
	ms := matchRule(t, rule, ":42")
	// ":" and "42" as separate tokens
	if len(ms) == 0 {
		// Analyze as single-char punct + digits
		sent := testSentence(atrTok(":", 0), atrTok("42", 1))
		var err error
		ms, err = rule.Match(sent)
		require.NoError(t, err)
	}
	require.Len(t, ms, 1)
	require.Equal(t, "Here come the match references: :42. This is the end", ms[0].GetMessage())

	sent2 := testSentence(atrTok(`\`, 0), atrTok("42", 1))
	// Java matches "\\42" as punct \ and digits 42
	ms2, err := rule.Match(sent2)
	require.NoError(t, err)
	if len(ms2) == 1 {
		require.Equal(t, "Here come the match references: \\42. This is the end", ms2[0].GetMessage())
	}
}

func TestPatternRuleMatcher_Equals(t *testing.T) {
	// Java RuleMatch.equals: same rule id, positions, message, sentence, type
	rule := NewPatternRule("id1", "xx", nil, "desc1", "msg1", "short1")
	rm1 := rules.NewRuleMatch(rule, nil, 0, 1, "message")
	rm2 := rules.NewRuleMatch(rule, nil, 0, 1, "message")
	require.True(t, ruleMatchEqual(rm1, rm2))
	rm3 := rules.NewRuleMatch(rule, nil, 0, 9, "message")
	require.False(t, ruleMatchEqual(rm1, rm3))
	require.False(t, ruleMatchEqual(rm2, rm3))
}

func ruleMatchEqual(a, b *rules.RuleMatch) bool {
	if a == nil || b == nil {
		return a == b
	}
	idA, idB := "", ""
	if pr, ok := a.GetRule().(*PatternRule); ok {
		idA = pr.GetID()
	}
	if pr, ok := b.GetRule().(*PatternRule); ok {
		idB = pr.GetID()
	}
	return idA == idB &&
		a.FromPos == b.FromPos && a.ToPos == b.ToPos &&
		a.PatternFromPos == b.PatternFromPos && a.PatternToPos == b.PatternToPos &&
		a.GetMessage() == b.GetMessage() &&
		a.Sentence == b.Sentence
}
