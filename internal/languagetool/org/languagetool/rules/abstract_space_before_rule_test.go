package rules

import (
	"regexp"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// spaceBeforeDouble is a thin test double of AbstractSpaceBeforeRule with a fixed
// conjunction Pattern (mirrors FA/KM subclasses without porting those languages here).
// Pattern is unanchored like Java Pattern.compile("and|or"); Match uses Matcher.matches().
type spaceBeforeDouble struct {
	*AbstractSpaceBeforeRule
}

func newSpaceBeforeDouble() *spaceBeforeDouble {
	// Java-style unanchored alternation; AbstractSpaceBeforeRule uses matches() full-string.
	base := &AbstractSpaceBeforeRule{
		Conjunctions: regexp.MustCompile(`and|or`),
	}
	InitSpaceBeforeMeta(base, nil)
	return &spaceBeforeDouble{AbstractSpaceBeforeRule: base}
}

func TestAbstractSpaceBeforeRule_Metadata(t *testing.T) {
	r := newSpaceBeforeDouble()
	require.Equal(t, "SPACE_BEFORE_CONJUNCTION", r.GetID())
	require.Equal(t, "Checks for missing space before some conjunctions", r.GetDescription())
	require.NotNil(t, r.GetCategory())
	require.True(t, r.GetCategory().GetID().Equals(CategoryMisc))
	require.False(t, r.IsDefaultOff())
	// Java defaults for getSuggestion / getShort
	require.Equal(t, "Missing white space before conjunction", r.getSuggestion())
	require.Equal(t, "Missing white space", r.getShort())
}

func TestAbstractSpaceBeforeRule_MissingSpaceBeforeConjunction(t *testing.T) {
	r := newSpaceBeforeDouble()
	// "and" at sentence start: previous is SENT_START "" → match (Java)
	sent := languagetool.AnalyzePlain("and foo")
	ms := r.Match(sent)
	require.Len(t, ms, 1)
	require.Equal(t, 0, ms[0].GetFromPos())
	require.Equal(t, 3, ms[0].GetToPos()) // token.length() of "and"
	require.Equal(t, "Missing white space before conjunction", ms[0].GetMessage())
	require.Equal(t, "Missing white space", ms[0].GetShortMessage())
	require.Equal(t, []string{" and"}, ms[0].GetSuggestedReplacements())
	require.Equal(t, r.AbstractSpaceBeforeRule, ms[0].GetRule())
}

func TestAbstractSpaceBeforeRule_AlreadyHasSpaceBefore(t *testing.T) {
	r := newSpaceBeforeDouble()
	// previous token is " " → no match
	ms := r.Match(languagetool.AnalyzePlain("x and y"))
	require.Empty(t, ms)
	ms = r.Match(languagetool.AnalyzePlain("x or y"))
	require.Empty(t, ms)
}

func TestAbstractSpaceBeforeRule_PreviousIsOpenParen(t *testing.T) {
	r := newSpaceBeforeDouble()
	// previous token is "(" → no match
	ms := r.Match(languagetool.AnalyzePlain("(and y"))
	require.Empty(t, ms)
	ms = r.Match(languagetool.AnalyzePlain("x (or y"))
	require.Empty(t, ms)
}

func TestAbstractSpaceBeforeRule_NonMatchingToken(t *testing.T) {
	r := newSpaceBeforeDouble()
	require.Empty(t, r.Match(languagetool.AnalyzePlain("foo")))
	// Substring must NOT match: Java Matcher.matches() is full-string only.
	// Pattern "and|or" must not fire on token "band" / "aand" / "andx".
	require.Empty(t, r.Match(languagetool.AnalyzePlain("band")))
	require.Empty(t, r.Match(languagetool.AnalyzePlain("aand")))
	require.Empty(t, r.Match(languagetool.AnalyzePlain("andx")))
	require.Empty(t, r.Match(languagetool.AnalyzePlain("xor")))
}

func TestAbstractSpaceBeforeRule_MultiMatch(t *testing.T) {
	r := newSpaceBeforeDouble()
	// "and,and": two conjunction tokens; prev of first is ""; prev of second is ","
	// neither is " " or "(" → two matches
	sent := languagetool.AnalyzePlain("and,and")
	ms := r.Match(sent)
	require.Len(t, ms, 2)
	require.Equal(t, 0, ms[0].GetFromPos())
	require.Equal(t, 3, ms[0].GetToPos())
	require.Equal(t, []string{" and"}, ms[0].GetSuggestedReplacements())
	require.Equal(t, 4, ms[1].GetFromPos())
	require.Equal(t, 7, ms[1].GetToPos())
	require.Equal(t, []string{" and"}, ms[1].GetSuggestedReplacements())

	// "and and": first matches (SENT_START), second has space before → one match
	ms = r.Match(languagetool.AnalyzePlain("and and"))
	require.Len(t, ms, 1)
	require.Equal(t, 0, ms[0].GetFromPos())
	require.Equal(t, 3, ms[0].GetToPos())
}

func TestAbstractSpaceBeforeRule_OrConjunctionAndOverrides(t *testing.T) {
	r := newSpaceBeforeDouble()
	// "or" as conjunction at start
	ms := r.Match(languagetool.AnalyzePlain("or bar"))
	require.Len(t, ms, 1)
	require.Equal(t, []string{" or"}, ms[0].GetSuggestedReplacements())
	require.Equal(t, 0, ms[0].GetFromPos())
	require.Equal(t, 2, ms[0].GetToPos())

	// Subclass-style message/id overrides via fields (Java method overrides)
	r.ID = "TEST_SPACE_BEFORE"
	r.Description = "test desc"
	r.ShortMsg = "short-x"
	r.Suggestion = "sug-x"
	require.Equal(t, "TEST_SPACE_BEFORE", r.GetID())
	require.Equal(t, "test desc", r.GetDescription())
	ms = r.Match(languagetool.AnalyzePlain("and"))
	require.Len(t, ms, 1)
	require.Equal(t, "sug-x", ms[0].GetMessage())
	require.Equal(t, "short-x", ms[0].GetShortMessage())
}

func TestAbstractSpaceBeforeRule_UTF16TokenLengthSpan(t *testing.T) {
	// Java token.length() counts UTF-16 code units; non-BMP is length 2.
	emoji := "😀"
	require.Equal(t, 2, utf16Len(emoji))
	base := &AbstractSpaceBeforeRule{
		Conjunctions: regexp.MustCompile(regexp.QuoteMeta(emoji)),
	}
	InitSpaceBeforeMeta(base, nil)
	// Manual sentence: SENT_START + emoji (no space before)
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", nil, nil)),
		func() *languagetool.AnalyzedTokenReadings {
			ar := languagetool.NewAnalyzedTokenReadingsAt(
				languagetool.NewAnalyzedToken(emoji, nil, nil), 0)
			return ar
		}(),
	})
	// Fix SENT_START start pos
	sent.GetTokens()[0].SetStartPos(0)

	ms := base.Match(sent)
	require.Len(t, ms, 1)
	require.Equal(t, 0, ms[0].GetFromPos())
	require.Equal(t, 2, ms[0].GetToPos(), "Java pos + token.length() with surrogate pair")
	require.Equal(t, []string{" " + emoji}, ms[0].GetSuggestedReplacements())
}

func TestJavaMatcherMatches_FullStringOnly(t *testing.T) {
	re := regexp.MustCompile(`and|or`)
	require.True(t, javaMatcherMatches(re, "and"))
	require.True(t, javaMatcherMatches(re, "or"))
	require.False(t, javaMatcherMatches(re, "band"))
	require.False(t, javaMatcherMatches(re, "andx"))
	require.False(t, javaMatcherMatches(re, "aand"))
	require.False(t, javaMatcherMatches(nil, "and"))
	// Anchored subclass-style patterns still work
	re2 := regexp.MustCompile(`^(و|به)$`)
	require.True(t, javaMatcherMatches(re2, "به"))
	require.False(t, javaMatcherMatches(re2, "xبه"))
}

// Java Matcher.matches() is full-region with Java regex alt/backtracking semantics.
// RE2 leftmost-first + FindStringIndex full-span is wrong for shorter-left alts.
func TestJavaMatcherMatches_ShorterLeftAlternation(t *testing.T) {
	// a|ab on "ab": Java matches() true (longer alt); old FindStringIndex span false.
	re := regexp.MustCompile(`a|ab`)
	require.True(t, javaMatcherMatches(re, "a"))
	require.True(t, javaMatcherMatches(re, "ab"), "Java matches() succeeds via longer alt ab")
	require.False(t, javaMatcherMatches(re, "abc"))
	require.False(t, javaMatcherMatches(re, "xa"))

	// FA-style unanchored body (Java PersianSpaceBeforeRule Pattern.compile):
	// "و|به|با|تا|زیرا|چون|بنابراین|چونکه" — چون before چونکه.
	faBody := regexp.MustCompile(`و|به|با|تا|زیرا|چون|بنابراین|چونکه`)
	require.True(t, javaMatcherMatches(faBody, "چون"))
	require.True(t, javaMatcherMatches(faBody, "چونکه"),
		"Java matches() on FA CONJUNCTIONS must accept چونکه despite shorter-left چون")
	require.True(t, javaMatcherMatches(faBody, "به"))
	require.False(t, javaMatcherMatches(faBody, "چونکex"))
	require.False(t, javaMatcherMatches(faBody, "xچونکه"))

	// Minimal FA-style prefix pair
	faPrefix := regexp.MustCompile(`چون|چونکه`)
	require.True(t, javaMatcherMatches(faPrefix, "چونکه"))
	require.True(t, javaMatcherMatches(faPrefix, "چون"))
	require.False(t, javaMatcherMatches(faPrefix, "چونک"))
}

func TestAbstractSpaceBeforeRule_ShorterLeftAltYieldsMatch(t *testing.T) {
	// Unanchored a|ab like Java Pattern.compile; Match uses Matcher.matches().
	base := &AbstractSpaceBeforeRule{
		Conjunctions: regexp.MustCompile(`a|ab`),
	}
	InitSpaceBeforeMeta(base, nil)
	// token "ab" at start: prev SENT_START "" → missing-space RuleMatch
	ms := base.Match(languagetool.AnalyzePlain("ab x"))
	require.Len(t, ms, 1)
	require.Equal(t, 0, ms[0].GetFromPos())
	require.Equal(t, 2, ms[0].GetToPos())
	require.Equal(t, []string{" ab"}, ms[0].GetSuggestedReplacements())

	// FA-style: چونکه must fire missing-space path
	fa := &AbstractSpaceBeforeRule{
		Conjunctions: regexp.MustCompile(`چون|چونکه`),
	}
	InitSpaceBeforeMeta(fa, nil)
	// Manual sentence: SENT_START + "چونکه" (no space before)
	tok := "چونکه"
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", nil, nil)),
		languagetool.NewAnalyzedTokenReadingsAt(
			languagetool.NewAnalyzedToken(tok, nil, nil), 0),
	})
	sent.GetTokens()[0].SetStartPos(0)
	ms = fa.Match(sent)
	require.Len(t, ms, 1, "FA-style چون|چونکه must match token چونکه and yield RuleMatch")
	require.Equal(t, 0, ms[0].GetFromPos())
	require.Equal(t, utf16Len(tok), ms[0].GetToPos())
	require.Equal(t, []string{" " + tok}, ms[0].GetSuggestedReplacements())

	// Still no match for partials under matches()
	require.Empty(t, base.Match(languagetool.AnalyzePlain("band"))) // a|ab does not full-match "band"
	partial := &AbstractSpaceBeforeRule{Conjunctions: regexp.MustCompile(`and|or`)}
	InitSpaceBeforeMeta(partial, nil)
	require.Empty(t, partial.Match(languagetool.AnalyzePlain("band")))
	require.Empty(t, partial.Match(languagetool.AnalyzePlain("andx")))
}
