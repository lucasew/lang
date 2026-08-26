package patterns

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPatternTokenMatcher(t *testing.T) {
	m := NewPatternTokenMatcher(Token("Hello"))
	tok := languagetool.NewAnalyzedToken("hello", nil, nil)
	require.True(t, m.IsMatched(tok))
	require.False(t, m.IsMatched(languagetool.NewAnalyzedToken("world", nil, nil)))

	cs := NewPatternTokenMatcher(CsToken("Hello"))
	require.False(t, cs.IsMatched(tok))
	require.True(t, cs.IsMatched(languagetool.NewAnalyzedToken("Hello", nil, nil)))

	re := NewPatternTokenMatcher(TokenRegex("c.t"))
	require.True(t, re.IsMatched(languagetool.NewAnalyzedToken("cat", nil, nil)))

	pos := "NN"
	pm := NewPatternTokenMatcher(Pos("NN"))
	require.True(t, pm.IsMatched(languagetool.NewAnalyzedToken("dog", &pos, nil)))
}

func TestPatternTokenMatcher_TextMatcherAndGetTestToken(t *testing.T) {
	// Inflected: Java getTestToken uses lemma when non-null (not surface-then-lemma).
	pt := NewPatternToken("run", false, false, true)
	m := NewPatternTokenMatcher(pt)
	require.NotNil(t, m.textMatcher)
	// surface "running" with lemma "run" → match via lemma only
	lem := "run"
	tok := languagetool.NewAnalyzedToken("running", nil, &lem)
	require.True(t, m.IsMatched(tok))
	// surface equals pattern but lemma differs → no match when lemma non-null
	other := "ran"
	tok2 := languagetool.NewAnalyzedToken("run", nil, &other)
	require.False(t, m.IsMatched(tok2))
	// null lemma → fall back to surface
	tok3 := languagetool.NewAnalyzedToken("run", nil, nil)
	require.True(t, m.IsMatched(tok3))

	// StringMatcher required-substrings path for regexp tokens
	rePt := NewPatternToken("foo.*bar", false, true, false)
	require.True(t, NewPatternTokenMatcher(rePt).IsMatched(
		languagetool.NewAnalyzedToken("fooXbar", nil, nil)))
	require.False(t, NewPatternTokenMatcher(rePt).IsMatched(
		languagetool.NewAnalyzedToken("foXbar", nil, nil)))

	// whitespace-only pattern → no TEST_STRING_MASK (empty after normalize)
	ws := NewPatternToken("   ", false, false, false)
	require.False(t, hasStringThatMustMatch(ws))
}

func TestPatternTokenMatcher_AndGroupCheck(t *testing.T) {
	// Base surface + and-group POS on different readings (Java andGroupCheck OR).
	base := NewPatternToken("bank", false, false, false)
	andPOS := NewPatternToken("", false, false, false)
	andPOS.SetPosToken(PosToken{PosTag: "NN"})
	base.AddAndGroupElement(andPOS)
	m := NewPatternTokenMatcher(base)

	nn := "NN"
	vb := "VB"
	// reading0: bank/VB, reading1: bank/NN — base matches both; and POS needs NN
	atr := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("bank", &vb, nil),
		languagetool.NewAnalyzedToken("bank", &nn, nil),
	}, 0)
	require.True(t, m.IsMatchedReadings(atr))

	// only VB readings → and-group POS fails
	atrVB := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("bank", &vb, nil),
	}, 0)
	require.False(t, m.IsMatchedReadings(atrVB))
}

// Faithful: untagged tokens do not soft-accept open-class POS patterns.
// Ports PatternToken.isPosTokenMatched + PosToken.posUnknown.
func TestPatternTokenMatcher_UntaggedPOSStrict(t *testing.T) {
	nn := Pos("NN.*")
	nn.Pos.Regexp = true
	nm := NewPatternTokenMatcher(nn)
	require.False(t, nm.IsMatched(languagetool.NewAnalyzedToken("man", nil, nil)))

	unk := Pos("UNKNOWN")
	um := NewPatternTokenMatcher(unk)
	require.True(t, um.IsMatched(languagetool.NewAnalyzedToken("man", nil, nil)))

	// Regexp that accepts UNKNOWN also matches untagged (Java posUnknown).
	unkRE := NewPatternToken("", false, false, false)
	unkRE.SetPosToken(PosToken{PosTag: "UNKNOWN|NN", Regexp: true})
	require.True(t, NewPatternTokenMatcher(unkRE).IsMatched(
		languagetool.NewAnalyzedToken("man", nil, nil)))
	// Tagged NN still matches the alternation
	nnTag := "NN"
	require.True(t, NewPatternTokenMatcher(unkRE).IsMatched(
		languagetool.NewAnalyzedToken("man", &nnTag, nil)))
	// Tagged VB does not match UNKNOWN|NN and is not hasNoTag
	vb := "VB"
	require.False(t, NewPatternTokenMatcher(unkRE).IsMatched(
		languagetool.NewAnalyzedToken("run", &vb, nil)))

	// Exact non-UNKNOWN with null POS → false (not posUnknown)
	exact := Pos("NN")
	require.False(t, NewPatternTokenMatcher(exact).IsMatched(
		languagetool.NewAnalyzedToken("man", nil, nil)))
}

// Upstream EN NON_ENGLISH_CHARACTER_IN_A_WORD uses Java \uXXXX escapes.
func TestNormalizeJavaRegexpUnicode(t *testing.T) {
	pat := `[a-z]*(\u043E|\u0455|\u0435|\u0440|\u03BF)[a-z]*`
	got := normalizeJavaRegexp(pat)
	require.Contains(t, got, `\x{043e}`)
	require.Contains(t, got, `\x{0455}`)
	m := NewPatternTokenMatcher(NewPatternToken(pat, false, true, false))
	// U+0455 CYRILLIC SMALL LETTER DZE looks like Latin s
	require.True(t, m.IsMatched(languagetool.NewAnalyzedToken("ѕee", nil, nil)))
	require.False(t, m.IsMatched(languagetool.NewAnalyzedToken("see", nil, nil)))
}

// Java quotes non-alphabetic chars (\! \… \,); Go RE2 rejects those escapes.
// Upstream EN grammar: [\.\!\?\,\;\:\…]
func TestNormalizeJavaRegexpQuotedNonMeta(t *testing.T) {
	pat := `[\.\!\?\,\;\:\…]`
	got := normalizeJavaRegexp(pat)
	// Meta escapes kept; Java-quoted non-meta (! , ; : …) drop the backslash.
	require.Equal(t, `[\.!\?,;:…]`, got)
	// Still compiles and matches punctuation / ellipsis.
	m := NewStringMatcher(pat, true, true)
	require.True(t, m.Matches("…"))
	require.True(t, m.Matches("!"))
	require.True(t, m.Matches("."))
	require.False(t, m.Matches("a"))
	// Meta stays escaped; Java-quoted non-meta drops the backslash.
	require.Equal(t, `\.|!`, normalizeJavaRegexp(`\.|\!`))
}

func TestNormalizeJavaRegexp_PossessiveQuantifiers(t *testing.T) {
	require.Equal(t, `a{1,3}b`, normalizeJavaRegexp(`a{1,3}+b`))
	require.Equal(t, `a*b+c?`, normalizeJavaRegexp(`a*+b++c?+`))
	// Inside class, + is literal / not a possessive quantifier marker after {n}.
	require.Equal(t, `[a+]`, normalizeJavaRegexp(`[a+]`))
	// Unicode property braces are not quantifiers — keep trailing +.
	require.Equal(t, `\p{Ll}+`, normalizeJavaRegexp(`\p{Ll}+`))
	m := NewStringMatcher(`x{1,3}+`, true, true)
	require.True(t, m.Matches("xxx"))
	require.False(t, m.Matches("xxxx"))
}

// DE grammar.xml uses [a-z0-9\äöüß]+ — Java drops the backslash for non-ASCII letters.
func TestNormalizeJavaRegexp_NonASCIILetterEscape(t *testing.T) {
	// Java Pattern: \ä is literal ä (not an illegal alphabetic escape — only ASCII a-z).
	require.Equal(t, `[a-z0-9äöüß]+e`, normalizeJavaRegexp(`[a-z0-9\äöüß]+e`))
	require.Equal(t, `[a-z0-9äöüß]+en`, normalizeJavaRegexp(`[a-z0-9\äöüß]+en`))
	// Compiles and matches umlaut stems like DE grammar ADJ/PA tokens.
	m := NewStringMatcher(`[a-z0-9\äöüß]+e`, true, true)
	require.True(t, m.Matches("schöne"))
	require.True(t, m.Matches("große"))
	require.False(t, m.Matches("schön"))
	// ASCII unknown letter escape still kept (fail closed at compile).
	require.Equal(t, `\q`, normalizeJavaRegexp(`\q`))
}

// Java CaseRule emoji antipattern uses UTF-16 surrogate pair ranges.
func TestNormalizeJavaRegexp_UTF16SurrogatePairRange(t *testing.T) {
	pat := `[\ud83c\udc00-\ud83c\udfff]+|[\ud83d\udc00-\ud83d\udfff]+|[\u2600-\u27ff]+`
	got := normalizeJavaRegexp(pat)
	require.Contains(t, got, `\x{1f000}-\x{1f3ff}`)
	require.Contains(t, got, `\x{1f400}-\x{1f7ff}`)
	require.Contains(t, got, `\x{2600}-\x{27ff}`)
	// Compiles and matches a BMP symbol from the third alternative.
	m := NewStringMatcher(pat, true, true)
	require.True(t, m.Matches("☀")) // U+2600
	require.False(t, m.Matches("a"))
}

func TestIsMatchedReadings_ChunkTag(t *testing.T) {
	pt := NewPatternToken("house", false, false, false)
	pt.SetChunkTag("B-NP", false)
	m := NewPatternTokenMatcher(pt)
	pos := "NN"
	atr := languagetool.NewAnalyzedTokenReadingsList(
		[]*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken("house", &pos, nil)},
		0,
	)
	atr.SetChunkTags([]string{"B-NP"})
	require.True(t, m.IsMatchedReadings(atr))
	atr.SetChunkTags([]string{"I-VP"})
	require.False(t, m.IsMatchedReadings(atr))

	// Chunk-only (empty surface): chunk required when negation=false.
	ptChunkOnly := NewPatternToken("", false, false, false)
	ptChunkOnly.SetChunkTag("B-NP", false)
	mChunkOnly := NewPatternTokenMatcher(ptChunkOnly)
	atr.SetChunkTags([]string{"B-NP"})
	require.True(t, mChunkOnly.IsMatchedReadings(atr))
	atr.SetChunkTags([]string{"I-VP"})
	require.False(t, mChunkOnly.IsMatchedReadings(atr))

	// And-group chunk tags (Java testAllReadings and-group chunk loop).
	base := NewPatternToken("house", false, false, false)
	andC := NewPatternToken("", false, false, false)
	andC.SetChunkTag("B-NP", false)
	base.AddAndGroupElement(andC)
	mAnd := NewPatternTokenMatcher(base)
	atr.SetChunkTags([]string{"B-NP"})
	require.True(t, mAnd.IsMatchedReadings(atr), "and-group chunk B-NP required")
	atr.SetChunkTags([]string{"I-VP"})
	require.False(t, mAnd.IsMatchedReadings(atr), "and-group chunk mismatch fails")
}

// Java: anyMatched &= chunkMatch ^ getNegation() after surface/POS match.
func TestIsMatchedReadings_ChunkXORNegation(t *testing.T) {
	// Surface "x" with Negation: "y" matches ((false^true)=true); chunk B-NP ^ true
	// requires non-B-NP for overall match.
	pt := NewPatternToken("x", false, false, false)
	pt.SetNegation(true)
	pt.SetChunkTag("B-NP", false)
	m := NewPatternTokenMatcher(pt)
	pos := "NN"
	atr := languagetool.NewAnalyzedTokenReadingsList(
		[]*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken("y", &pos, nil)},
		0,
	)
	atr.SetChunkTags([]string{"I-VP"}) // chunkOK false → false^true → keep match
	require.True(t, m.IsMatchedReadings(atr), "negated surface non-match + non-chunk → match")
	atr.SetChunkTags([]string{"B-NP"}) // chunkOK true → true^true → fail
	require.False(t, m.IsMatchedReadings(atr), "negated surface + matching chunk → fail")
}

// Ensures NewStringMatcher does not panic on DE grammar.xml unnecessary umlaut escapes.
func TestStringMatcher_DEUmlautClassNoPanic(t *testing.T) {
	// From de/grammar.xml ADJ/PA surface patterns
	require.NotPanics(t, func() {
		m := NewStringMatcher(`[a-z0-9\äöüß]+e`, true, true)
		require.True(t, m.Matches("leise"))
		require.True(t, m.Matches("große"))
	})
}
