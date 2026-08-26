package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func strPtrCTS(s string) *string { return &s }

func atrCTS(token string, startPos int, posTag *string, lemma *string, wsBefore bool) *languagetool.AnalyzedTokenReadings {
	at := languagetool.NewAnalyzedToken(token, posTag, lemma)
	r := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{at}, startPos)
	if wsBefore {
		r.SetWhitespaceBeforeToken(" ")
	}
	return r
}

func atrCTSNoTag(token string, startPos int, wsBefore bool) *languagetool.AnalyzedTokenReadings {
	// null POS → hasNoTag
	at := languagetool.NewAnalyzedToken(token, nil, nil)
	r := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{at}, startPos)
	if wsBefore {
		r.SetWhitespaceBeforeToken(" ")
	}
	return r
}

func atrCTSMultiReading(token string, startPos int, readings []*languagetool.AnalyzedToken, wsBefore bool) *languagetool.AnalyzedTokenReadings {
	r := languagetool.NewAnalyzedTokenReadingsList(readings, startPos)
	if wsBefore {
		r.SetWhitespaceBeforeToken(" ")
	}
	return r
}

func matchSpan(from, to int) *RuleMatch {
	return NewRuleMatch(NewFakeRule("SC"), nil, from, to, "msg")
}

func TestConvertToSentenceCaseFilter_HelloWorldLowerLemmas(t *testing.T) {
	// HELLO WORLD with lower lemmas → Hello world
	f := NewConvertToSentenceCaseFilter()
	pos := "NN"
	t1 := atrCTS("HELLO", 0, &pos, strPtrCTS("hello"), false)
	t2 := atrCTS("WORLD", 6, &pos, strPtrCTS("world"), true)
	m := matchSpan(0, 11)
	out := f.AcceptRuleMatch(m, nil, 0, []*languagetool.AnalyzedTokenReadings{t1, t2}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"Hello world"}, out.GetSuggestedReplacements())
}

func TestConvertToSentenceCaseFilter_AlreadySentenceCaseNil(t *testing.T) {
	// already sentence case → replacement equals original → null
	f := NewConvertToSentenceCaseFilter()
	pos := "NN"
	t1 := atrCTS("Hello", 0, &pos, strPtrCTS("hello"), false)
	t2 := atrCTS("world", 6, &pos, strPtrCTS("world"), true)
	m := matchSpan(0, 11)
	out := f.AcceptRuleMatch(m, nil, 0, []*languagetool.AnalyzedTokenReadings{t1, t2}, nil)
	require.Nil(t, out)
}

func TestConvertToSentenceCaseFilter_CorpDot(t *testing.T) {
	// corp. → Corp.
	f := NewConvertToSentenceCaseFilter()
	pos := "NN"
	t1 := atrCTS("corp", 0, &pos, strPtrCTS("corp"), false)
	t2 := atrCTS(".", 4, &pos, strPtrCTS("."), false)
	m := matchSpan(0, 5)
	out := f.AcceptRuleMatch(m, nil, 0, []*languagetool.AnalyzedTokenReadings{t1, t2}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"Corp."}, out.GetSuggestedReplacements())
}

func TestConvertToSentenceCaseFilter_SingleLetterBeforeDot(t *testing.T) {
	// single-char normalized before "." → uppercased ("a." → "A.")
	// first token is also first word → UppercaseFirstChar anyway; use mid-sentence via leading punct
	f := NewConvertToSentenceCaseFilter()
	pos := "NN"
	// ", a." → first real word is "a" which is capitalized as first; also single-letter before .
	// Better: "X a." with lower lemma for a so normalized is "a", then before "." → "A", but first is X
	t0 := atrCTS("X", 0, &pos, strPtrCTS("x"), false)
	t1 := atrCTS("a", 2, &pos, strPtrCTS("a"), true)
	t2 := atrCTS(".", 3, &pos, strPtrCTS("."), false)
	m := matchSpan(0, 4)
	out := f.AcceptRuleMatch(m, nil, 0, []*languagetool.AnalyzedTokenReadings{t0, t1, t2}, nil)
	require.NotNil(t, out)
	// first: UppercaseFirstChar("x")="X"; then space + single-letter upper "A" + "."
	require.Equal(t, []string{"X A."}, out.GetSuggestedReplacements())
}

func TestConvertToSentenceCaseFilter_PatternTokenPosOffset(t *testing.T) {
	// tokens before patternTokenPos are ignored
	f := NewConvertToSentenceCaseFilter()
	pos := "NN"
	// [SKIP THIS] HELLO WORLD — start at index 2
	pre1 := atrCTS("SKIP", 0, &pos, strPtrCTS("skip"), false)
	pre2 := atrCTS("THIS", 5, &pos, strPtrCTS("this"), true)
	t1 := atrCTS("HELLO", 10, &pos, strPtrCTS("hello"), true)
	t2 := atrCTS("WORLD", 16, &pos, strPtrCTS("world"), true)
	// match covers only HELLO WORLD
	m := matchSpan(10, 21)
	out := f.AcceptRuleMatch(m, nil, 2, []*languagetool.AnalyzedTokenReadings{pre1, pre2, t1, t2}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"Hello world"}, out.GetSuggestedReplacements())

	// if patternTokenPos were 0, pre tokens outside match span are also skipped by span check
	out0 := f.AcceptRuleMatch(m, nil, 0, []*languagetool.AnalyzedTokenReadings{pre1, pre2, t1, t2}, nil)
	require.NotNil(t, out0)
	require.Equal(t, []string{"Hello world"}, out0.GetSuggestedReplacements())

	// patternTokenPos past all tokens → empty builders equal → nil
	outPast := f.AcceptRuleMatch(m, nil, 4, []*languagetool.AnalyzedTokenReadings{pre1, pre2, t1, t2}, nil)
	require.Nil(t, outPast)

	// patternTokenPos=3 → only WORLD (first word capitalizes)
	out3 := f.AcceptRuleMatch(m, nil, 3, []*languagetool.AnalyzedTokenReadings{pre1, pre2, t1, t2}, nil)
	require.NotNil(t, out3)
	require.Equal(t, []string{"World"}, out3.GetSuggestedReplacements())
}

func TestConvertToSentenceCaseFilter_TokensOutsideMatchSpanSkipped(t *testing.T) {
	f := NewConvertToSentenceCaseFilter()
	pos := "NN"
	// BEFORE HELLO WORLD AFTER — match only HELLO WORLD
	before := atrCTS("BEFORE", 0, &pos, strPtrCTS("before"), false)
	t1 := atrCTS("HELLO", 7, &pos, strPtrCTS("hello"), true)
	t2 := atrCTS("WORLD", 13, &pos, strPtrCTS("world"), true)
	after := atrCTS("AFTER", 19, &pos, strPtrCTS("after"), true)
	m := matchSpan(7, 18)
	out := f.AcceptRuleMatch(m, nil, 0, []*languagetool.AnalyzedTokenReadings{before, t1, t2, after}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"Hello world"}, out.GetSuggestedReplacements())
}

func TestConvertToSentenceCaseFilter_LemmaNullOrNoTagCapitalize(t *testing.T) {
	f := NewConvertToSentenceCaseFilter()
	// hasNoTag / null lemma → capitalized form
	t1 := atrCTSNoTag("HELLO", 0, false)
	t2 := atrCTSNoTag("WORLD", 6, true)
	m := matchSpan(0, 11)
	out := f.AcceptRuleMatch(m, nil, 0, []*languagetool.AnalyzedTokenReadings{t1, t2}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"Hello World"}, out.GetSuggestedReplacements())
}

func TestConvertToSentenceCaseFilter_LemmaLower(t *testing.T) {
	f := NewConvertToSentenceCaseFilter()
	pos := "NN"
	t1 := atrCTS("HELLO", 0, &pos, strPtrCTS("hello"), false)
	t2 := atrCTS("WORLD", 6, &pos, strPtrCTS("world"), true)
	m := matchSpan(0, 11)
	out := f.AcceptRuleMatch(m, nil, 0, []*languagetool.AnalyzedTokenReadings{t1, t2}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"Hello world"}, out.GetSuggestedReplacements())
}

func TestConvertToSentenceCaseFilter_LemmaCapitalized(t *testing.T) {
	// capitalized lemma → tokenCapitalized for non-first as well
	f := NewConvertToSentenceCaseFilter()
	pos := "NNP"
	t1 := atrCTS("john", 0, &pos, strPtrCTS("John"), false)
	t2 := atrCTS("smith", 5, &pos, strPtrCTS("Smith"), true)
	m := matchSpan(0, 10)
	out := f.AcceptRuleMatch(m, nil, 0, []*languagetool.AnalyzedTokenReadings{t1, t2}, nil)
	// already "John Smith" after normalize? first: UppercaseFirstChar("john") with cap lemma = "John"
	// second: capitalized lemma → "Smith"; original "john smith" → suggestion "John Smith"
	require.NotNil(t, out)
	require.Equal(t, []string{"John Smith"}, out.GetSuggestedReplacements())
}

func TestConvertToSentenceCaseFilter_LemmaNeitherKeepSurface(t *testing.T) {
	// lemma neither all-lower nor capitalized (e.g. all-upper or mixed) → keep surface
	f := NewConvertToSentenceCaseFilter()
	pos := "NN"
	// lemma "NASA" is all upper: IsCapitalizedWord false, IsNotAllLowercase true → neither flag
	t1 := atrCTS("hello", 0, &pos, strPtrCTS("hello"), false) // first word capitalizes
	t2 := atrCTS("nasa", 6, &pos, strPtrCTS("NASA"), true)  // keep surface "nasa"
	m := matchSpan(0, 10)
	out := f.AcceptRuleMatch(m, nil, 0, []*languagetool.AnalyzedTokenReadings{t1, t2}, nil)
	// first lower lemma → "hello" then first capitalizes → "Hello"; second keep "nasa"
	// original "hello nasa" → "Hello nasa" differs
	require.NotNil(t, out)
	require.Equal(t, []string{"Hello nasa"}, out.GetSuggestedReplacements())

	// all-upper surface kept when lemma neither
	t3 := atrCTS("HELLO", 0, &pos, strPtrCTS("HELLO"), false) // neither → keep "HELLO"; first still UppercaseFirstChar of normalized
	// wait: neither returns atr.getToken() which is "HELLO"; firstDone applies UppercaseFirstChar("HELLO")="HELLO"
	t4 := atrCTS("NASA", 6, &pos, strPtrCTS("NASA"), true)
	m2 := matchSpan(0, 10)
	out2 := f.AcceptRuleMatch(m2, nil, 0, []*languagetool.AnalyzedTokenReadings{t3, t4}, nil)
	// replacement "HELLO NASA" == original → nil
	require.Nil(t, out2)
}

func TestConvertToSentenceCaseFilter_MultiwordLemmaFirstWord(t *testing.T) {
	// multi-word lemma: first word only for case flags
	f := NewConvertToSentenceCaseFilter()
	pos := "NN"
	// lemma "New York" → first word "New" is capitalized → capitalized branch
	t1 := atrCTS("new", 0, &pos, strPtrCTS("New York"), false)
	m := matchSpan(0, 3)
	out := f.AcceptRuleMatch(m, nil, 0, []*languagetool.AnalyzedTokenReadings{t1}, nil)
	// already "New" after cap → equals original "new"? UppercaseFirstChar → "New" ≠ "new"
	require.NotNil(t, out)
	require.Equal(t, []string{"New"}, out.GetSuggestedReplacements())
}

func TestConvertToSentenceCaseFilter_TypographicApostrophe(t *testing.T) {
	f := NewConvertToSentenceCaseFilter()
	pos := "NN"
	t1 := atrCTS("DON'T", 0, &pos, strPtrCTS("do"), false)
	// mark typographic apostrophe on readings
	t1.SetTypographicApostrophe(true)
	// second token so first is DON'T → first capitalizes normalized with ’ 
	// normalized: "don't".toLower with typo apos → "don" + ’ + "t" = "don’t"
	// first: UppercaseFirstChar → "Don’t"
	m := matchSpan(0, 5)
	out := f.AcceptRuleMatch(m, nil, 0, []*languagetool.AnalyzedTokenReadings{t1}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"Don’t"}, out.GetSuggestedReplacements())
}

func TestConvertToSentenceCaseFilter_TokenIsException(t *testing.T) {
	// exception returns lower form; as non-first stays lower
	f := NewConvertToSentenceCaseFilter()
	f.TokenIsException = func(s string) bool { return s == "me" }
	pos := "NN"
	t1 := atrCTS("CALL", 0, &pos, strPtrCTS("call"), false)
	t2 := atrCTS("ME", 5, &pos, strPtrCTS("I"), true) // exception → "me"
	t3 := atrCTS("LATER", 8, &pos, strPtrCTS("later"), true)
	m := matchSpan(0, 13)
	out := f.AcceptRuleMatch(m, nil, 0, []*languagetool.AnalyzedTokenReadings{t1, t2, t3}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"Call me later"}, out.GetSuggestedReplacements())

	// exception as first token: normalized lower "me", then first applies UppercaseFirstChar → "Me"
	t4 := atrCTS("ME", 0, &pos, strPtrCTS("I"), false)
	t5 := atrCTS("AND", 3, &pos, strPtrCTS("and"), true)
	t6 := atrCTS("YOU", 7, &pos, strPtrCTS("you"), true)
	m2 := matchSpan(0, 10)
	out2 := f.AcceptRuleMatch(m2, nil, 0, []*languagetool.AnalyzedTokenReadings{t4, t5, t6}, nil)
	require.NotNil(t, out2)
	require.Equal(t, []string{"Me and you"}, out2.GetSuggestedReplacements())
}

func TestConvertToSentenceCaseFilter_PunctuationSkipsFirstCapitalization(t *testing.T) {
	// leading punctuation does not count as first word
	f := NewConvertToSentenceCaseFilter()
	pos := "NN"
	// "\" HELLO" → quote is punct, HELLO is first real word
	t0 := atrCTS("\"", 0, &pos, strPtrCTS("\""), false)
	t1 := atrCTS("HELLO", 1, &pos, strPtrCTS("hello"), false) // no ws before in original quote attach
	m := matchSpan(0, 6)
	out := f.AcceptRuleMatch(m, nil, 0, []*languagetool.AnalyzedTokenReadings{t0, t1}, nil)
	require.NotNil(t, out)
	// punct branch: normalized of "\"" — lemma lower? "\"" lower, isPunctuation true → else branch without first
	// firstDone still false; HELLO becomes first → "Hello"
	// original: "\"HELLO" vs "\"Hello" or with spaces?
	// t0 no ws; else: no space + normalizedCase of quote
	// quote lemma "\"" — is it lower? IsNotAllLowercase("\"") → no letters → false → !false=true → lemmaIsLowercase
	// so normalized = "\""
	// else branch: no ws, append "\""
	// then HELLO firstDone path: UppercaseFirstChar("hello")="Hello"
	require.Equal(t, []string{"\"Hello"}, out.GetSuggestedReplacements())
}

func TestConvertToSentenceCaseFilter_NullPatternTokenPanics(t *testing.T) {
	// Java NPE on null patternTokens[i]
	f := NewConvertToSentenceCaseFilter()
	m := matchSpan(0, 1)
	require.Panics(t, func() {
		f.AcceptRuleMatch(m, nil, 0, []*languagetool.AnalyzedTokenReadings{nil}, nil)
	})
}

func TestConvertToSentenceCaseFilter_NullMatchPanics(t *testing.T) {
	// Java NPE on match.getFromPos()
	f := NewConvertToSentenceCaseFilter()
	pos := "NN"
	t1 := atrCTS("HELLO", 0, &pos, strPtrCTS("hello"), false)
	require.Panics(t, func() {
		f.AcceptRuleMatch(nil, nil, 0, []*languagetool.AnalyzedTokenReadings{t1}, nil)
	})
}

func TestConvertToSentenceCaseFilter_MixedLemmaFlagsPreferLower(t *testing.T) {
	// if any reading is lower lemma, lower wins (OR flags; lower checked first)
	f := NewConvertToSentenceCaseFilter()
	pos1, pos2 := "NN", "NNP"
	readings := []*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("WORLD", &pos1, strPtrCTS("world")), // lower
		languagetool.NewAnalyzedToken("WORLD", &pos2, strPtrCTS("World")), // capitalized
	}
	t1 := atrCTS("HELLO", 0, &pos1, strPtrCTS("hello"), false)
	t2 := atrCTSMultiReading("WORLD", 6, readings, true)
	m := matchSpan(0, 11)
	out := f.AcceptRuleMatch(m, nil, 0, []*languagetool.AnalyzedTokenReadings{t1, t2}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"Hello world"}, out.GetSuggestedReplacements())
}

func TestConvertToSentenceCaseFilter_IsPunctuationHelper(t *testing.T) {
	require.True(t, isPunctuationCTS("."))
	require.True(t, isPunctuationCTS(","))
	require.True(t, isPunctuationCTS("\""))
	require.False(t, isPunctuationCTS(""))
	require.False(t, isPunctuationCTS(".."))
	require.False(t, isPunctuationCTS("a"))
	require.False(t, isPunctuationCTS("a."))
}
