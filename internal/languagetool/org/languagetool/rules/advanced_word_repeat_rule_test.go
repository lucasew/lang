package rules

import (
	"regexp"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func strPtrAdv(s string) *string { return &s }

// test double implementing AdvancedWordRepeatRule abstract hooks.
func newAdvRepeatDouble() *AdvancedWordRepeatRule {
	r := &AdvancedWordRepeatRule{
		ExcludedWords: map[string]bool{
			"się": true,
			"i":   true,
		},
		// same shape as PL EXC_NONWORDS (Java Matcher.matches full-string)
		ExcludedNonWords: regexp.MustCompile(`&quot|&gt|&lt|&amp|[0-9].*|M*(D?C{0,3}|C[DM])(L?X{0,3}|X[LC])(V?I{0,3}|I[VX])$`),
		// same shape as PL EXC_POS
		ExcludedPos:  regexp.MustCompile(`prep:.*|ppron.*`),
		ID:           "ADV_REPEAT_TEST",
		Message:      "Powtórzony wyraz w zdaniu",
		ShortMessage: "Powtórzenie wyrazu",
	}
	InitAdvancedWordRepeatMeta(r, map[string]string{
		"category_misc": "Miscellaneous",
	})
	return r
}

func atr(token string, start int, readings ...*languagetool.AnalyzedToken) *languagetool.AnalyzedTokenReadings {
	if len(readings) == 0 {
		return languagetool.NewAnalyzedTokenReadingsAt(
			languagetool.NewAnalyzedToken(token, nil, nil), start)
	}
	if len(readings) == 1 {
		return languagetool.NewAnalyzedTokenReadingsAt(readings[0], start)
	}
	return languagetool.NewAnalyzedTokenReadingsList(readings, start)
}

func sentStart() *languagetool.AnalyzedTokenReadings {
	return languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("", strPtrAdv(languagetool.SentenceStartTagName), nil), 0)
}

func readingPosLemma(token string, pos, lemma *string) *languagetool.AnalyzedToken {
	return languagetool.NewAnalyzedToken(token, pos, lemma)
}

func TestAdvancedWordRepeatRule_Metadata(t *testing.T) {
	r := newAdvRepeatDouble()
	require.True(t, r.IsDefaultOff(), "Java setDefaultOff()")
	require.Equal(t, ITSStyle, r.GetLocQualityIssueType())
	require.NotNil(t, r.GetCategory())
	require.True(t, r.GetCategory().GetID().Equals(CategoryMisc))
}

// Java AdvancedWordRepeatRule: token.length() < 2 uses UTF-16 units.
// A single non-BMP code point has length 2 in Java and is still a "word".
func TestAdvancedWordRepeatRule_UTF16TokenLengthGate(t *testing.T) {
	r := newAdvRepeatDouble()
	emoji := "😀"
	require.Equal(t, 2, utf16LenAdv(emoji))
	// two identical emoji tokens → surface repeat when untagged (hasLemma false)
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		sentStart(),
		atr(emoji, 0),
		atr(emoji, 2),
	})
	ms := r.Match(sent)
	require.Len(t, ms, 1, "Java treats length-2 surrogate pair as word; repeated emoji matches")
	require.Equal(t, 2, ms[0].FromPos)
	require.Equal(t, 4, ms[0].ToPos) // start 2 + UTF-16 length 2
	require.Equal(t, r.Message, ms[0].Message)
	require.Equal(t, r.ShortMessage, ms[0].ShortMessage)

	// ASCII single letter still excluded (length 1)
	sent2 := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		sentStart(),
		atr("a", 0),
		atr("a", 1),
	})
	require.Empty(t, r.Match(sent2), "single-letter ASCII not a word (length < 2)")
}

func TestAdvancedWordRepeatRule_NoRepeatClean(t *testing.T) {
	r := newAdvRepeatDouble()
	// untagged distinct surfaces
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		sentStart(),
		atr("To", 0),
		atr("jest", 3),
		atr("zdanie", 8),
		atr(".", 14, readingPosLemma(".", strPtrAdv(languagetool.SentenceEndTagName), nil)),
	})
	require.Empty(t, r.Match(sent))
}

func TestAdvancedWordRepeatRule_SurfaceRepeatUntagged(t *testing.T) {
	r := newAdvRepeatDouble()
	// Java PL sample (surface path without tagger): "bowiem" twice → 1 match
	// "Był on bowiem pięknym strzelcem bowiem."
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		sentStart(),
		atr("Był", 0),
		atr("on", 4),
		atr("bowiem", 7),
		atr("pięknym", 14),
		atr("strzelcem", 22),
		atr("bowiem", 32),
		atr(".", 38),
	})
	ms := r.Match(sent)
	require.Len(t, ms, 1)
	require.Equal(t, 32, ms[0].FromPos)
	require.Equal(t, 32+utf16LenAdv("bowiem"), ms[0].ToPos)
	require.Equal(t, "Powtórzony wyraz w zdaniu", ms[0].Message)
	require.Equal(t, "Powtórzenie wyrazu", ms[0].ShortMessage)
}

func TestAdvancedWordRepeatRule_LemmaRepeatDifferentSurfaces(t *testing.T) {
	r := newAdvRepeatDouble()
	// Mówiła / mówić share lemma "mówić"; długo twice with same lemma → 2 matches
	// mirrors Java PolishWordRepeatRuleTest count for tagged path.
	posV := "verb:praet"
	posAdv := "adv"
	lemMowic := "mówić"
	lemDlugo := "długo"
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		sentStart(),
		atr("Mówiła", 0, readingPosLemma("Mówiła", &posV, &lemMowic)),
		atr("długo", 7, readingPosLemma("długo", &posAdv, &lemDlugo)),
		atr(",", 12),
		atr("żeby", 14),
		atr("tylko", 19),
		atr("mówić", 25, readingPosLemma("mówić", &posV, &lemMowic)),
		atr("długo", 31, readingPosLemma("długo", &posAdv, &lemDlugo)),
		atr(".", 36),
	})
	ms := r.Match(sent)
	require.Len(t, ms, 2, "lemma mówić + lemma długo each once as repeats")
	// first hit is second "mówić" (lemma already seen)
	require.Equal(t, 25, ms[0].FromPos)
	require.Equal(t, 25+utf16LenAdv("mówić"), ms[0].ToPos)
	// second hit is second "długo"
	require.Equal(t, 31, ms[1].FromPos)
	require.Equal(t, 31+utf16LenAdv("długo"), ms[1].ToPos)
}

func TestAdvancedWordRepeatRule_ExcludedLemmaSet(t *testing.T) {
	r := newAdvRepeatDouble()
	pos := "qub"
	lem := "się"
	// "się" twice with excluded lemma → isWord false → no match
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		sentStart(),
		atr("On", 0),
		atr("się", 3, readingPosLemma("się", &pos, &lem)),
		atr("umył", 7),
		atr("i", 12),
		atr("się", 14, readingPosLemma("się", &pos, &lem)),
		atr("ubrał", 18),
	})
	require.Empty(t, r.Match(sent), "excluded lemma się is not tracked")
}

func TestAdvancedWordRepeatRule_ExcludedPosFullStringMatch(t *testing.T) {
	r := newAdvRepeatDouble()
	// prep:loc matches ExcludedPos via Matcher.matches()
	pos := "prep:loc:nwok"
	lem := "na"
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		sentStart(),
		atr("był", 0),
		atr("na", 4, readingPosLemma("na", &pos, &lem)),
		atr("dyskotece", 7),
		atr("choć", 17),
		atr("na", 22, readingPosLemma("na", &pos, &lem)),
		atr("bani", 25),
	})
	require.Empty(t, r.Match(sent), "ExcludedPos prep:.* full-match excludes")

	// Java Matcher.matches() requires whole string: "xprep:loc" does NOT match prep:.*
	// (unanchored find would match "prep:loc" substring — that is NOT Java)
	badPos := "xprep:loc"
	lem2 := "na"
	sentFind := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		sentStart(),
		atr("był", 0),
		atr("na", 4, readingPosLemma("na", &badPos, &lem2)),
		atr("tu", 7),
		atr("na", 10, readingPosLemma("na", &badPos, &lem2)),
	})
	ms := r.Match(sentFind)
	require.Len(t, ms, 1, "substring-only POS must NOT exclude under Matcher.matches()")
}

func TestAdvancedWordRepeatRule_ExcludedNonWordsFullString(t *testing.T) {
	r := newAdvRepeatDouble()
	// digits: [0-9].* full-string match → not a word
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		sentStart(),
		atr("12", 0),
		atr("and", 3),
		atr("12", 7),
	})
	require.Empty(t, r.Match(sent), "numeric tokens excluded by non-words pattern")

	// Roman numeral XX
	sent2 := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		sentStart(),
		atr("XX", 0),
		atr("and", 3),
		atr("XX", 7),
	})
	require.Empty(t, r.Match(sent2), "roman numerals excluded")

	// partial: "x12" is not full match of [0-9].* under Matcher.matches()
	sent3 := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		sentStart(),
		atr("x12", 0),
		atr("and", 4),
		atr("x12", 8),
	})
	require.Len(t, r.Match(sent3), 1, "non-full numeric pattern must not exclude")
}

func TestAdvancedWordRepeatRule_EmptyPOSNotWord(t *testing.T) {
	r := newAdvRepeatDouble()
	empty := ""
	lem := "foo"
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		sentStart(),
		atr("foo", 0, readingPosLemma("foo", &empty, &lem)),
		atr("bar", 4),
		atr("foo", 8, readingPosLemma("foo", &empty, &lem)),
	})
	require.Empty(t, r.Match(sent), "empty POS → isWord=false (Java StringTools.isEmpty)")
}

func TestAdvancedWordRepeatRule_NullLemmaUsesSurface(t *testing.T) {
	r := newAdvRepeatDouble()
	pos := "NN"
	// null lemma → hasLemma false → track surface token
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		sentStart(),
		atr("Test", 0, readingPosLemma("Test", &pos, nil)),
		atr("word", 5),
		atr("Test", 10, readingPosLemma("Test", &pos, nil)),
	})
	ms := r.Match(sent)
	require.Len(t, ms, 1)
	require.Equal(t, 10, ms[0].FromPos)
}

func TestAdvancedWordRepeatRule_NullPOSUsesSurface(t *testing.T) {
	r := newAdvRepeatDouble()
	// null POS sets hasLemma false (Java) even if later readings had lemmas —
	// single null-POS reading: surface path.
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		sentStart(),
		atr("test", 0),
		atr("and", 5),
		atr("test", 9),
	})
	ms := r.Match(sent)
	require.Len(t, ms, 1)
	require.Equal(t, 9, ms[0].FromPos)
	require.Equal(t, 13, ms[0].ToPos)
}

func TestAdvancedWordRepeatRule_SentEndSkipsRepetitionFlag(t *testing.T) {
	r := newAdvRepeatDouble()
	// Java: notSentEnd |= SENT_END; only checks/adds when !notSentEnd.
	// Token with sole SENT_END reading: first iteration sets notSentEnd=true before
	// the hasLemma/surface check's condition allows setting repetition=true when
	// already in set — but else branch still adds surface when hasLemma false:
	// if (contains && !notSentEnd) repetition; else add.
	// So a pure SENT_END second occurrence does not set repetition=true.
	posEnd := languagetool.SentenceEndTagName
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		sentStart(),
		atr("word", 0),
		atr("word", 5, readingPosLemma("word", &posEnd, nil)),
	})
	// First loop: pos is SENT_END (non-empty), lemma nil → hasLemma=false, break.
	// isWord true (len>=2, not excluded).
	// Second loop: notSentEnd becomes true; surface branch: contains("word") && !notSentEnd → false;
	// else add. repetition stays false.
	require.Empty(t, r.Match(sent), "SENT_END reading prevents repetition flag (Java !notSentEnd)")
}

func TestAdvancedWordRepeatRule_SentEndAfterContentReadingStillTracksFirst(t *testing.T) {
	r := newAdvRepeatDouble()
	// Realistic: content reading then SENT_END on same token (SetSentEnd style).
	// First reading processes before notSentEnd flips; lemma/surface tracked.
	// Second occurrence of same surface without SENT_END matches.
	posNN := "NN"
	lem := "word"
	posEnd := languagetool.SentenceEndTagName
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		sentStart(),
		atr("word", 0, readingPosLemma("word", &posNN, &lem), readingPosLemma("word", &posEnd, &lem)),
		atr("and", 5),
		atr("word", 9, readingPosLemma("word", &posNN, &lem)),
	})
	ms := r.Match(sent)
	require.Len(t, ms, 1)
	require.Equal(t, 9, ms[0].FromPos)
}

func TestAdvancedWordRepeatRule_MultiLemmaSameTokenCurToken(t *testing.T) {
	r := newAdvRepeatDouble()
	// Same token two distinct lemmas: both added; curToken==i prevents self-repeat.
	pos := "x"
	l1, l2 := "alpha", "beta"
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		sentStart(),
		atr("ab", 0, readingPosLemma("ab", &pos, &l1), readingPosLemma("ab", &pos, &l2)),
		atr("xx", 3),
		atr("ab", 6, readingPosLemma("ab", &pos, &l1)),
	})
	ms := r.Match(sent)
	require.Len(t, ms, 1, "lemma alpha seen again at later token")
	require.Equal(t, 6, ms[0].FromPos)
}

func TestAdvancedWordRepeatRule_JavaRegexpMatchesHelper(t *testing.T) {
	re := regexp.MustCompile(`prep:.*|ppron.*`)
	require.True(t, javaRegexpMatches(re, "prep:loc"))
	require.True(t, javaRegexpMatches(re, "ppron:sg"))
	require.False(t, javaRegexpMatches(re, "xprep:loc"), "Matcher.matches not find")
	require.False(t, javaRegexpMatches(nil, "prep:loc"))
	// Go MatchString would be true for substring — prove divergence
	require.True(t, re.MatchString("xprep:loc"), "sanity: unanchored find differs")
}

func TestAdvancedWordRepeatRule_CaseSensitiveSurface(t *testing.T) {
	r := newAdvRepeatDouble()
	// Java getToken() exact surface — "Test" vs "test" are different keys
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		sentStart(),
		atr("Test", 0),
		atr("and", 5),
		atr("test", 9),
	})
	require.Empty(t, r.Match(sent))
}
