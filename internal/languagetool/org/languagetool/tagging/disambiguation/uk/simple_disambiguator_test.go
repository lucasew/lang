package uk

import (
	"regexp"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func atr(surface string, lemmasPOS [][2]string) *languagetool.AnalyzedTokenReadings {
	var rs []*languagetool.AnalyzedToken
	for _, lp := range lemmasPOS {
		l, p := lp[0], lp[1]
		rs = append(rs, languagetool.NewAnalyzedToken(surface, &p, &l))
	}
	return languagetool.NewAnalyzedTokenReadingsList(rs, 0)
}

func TestSimpleDisambiguator_RemoveRareForms(t *testing.T) {
	// inject: for "була" remove non-verb readings
	rm := map[string]*TokenMatcher{
		// Java MatcherEntry: full-string POS regex
		"була": {Entries: []MatcherEntry{{Lemma: "була", POS: regexp.MustCompile(`^noun.*$`)}}},
	}
	d := NewSimpleDisambiguatorWith(rm)
	sentStart := languagetool.NewAnalyzedTokenReadingsList(
		[]*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken("", strPtr("SENT_START"), nil)}, 0)
	// mark sentence start
	bula := atr("була", [][2]string{
		{"бути", "verb:imperf:past:f"},
		{"була", "noun:inanim:f:v_naz"},
	})
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{sentStart, bula})
	out := d.Disambiguate(sent)
	require.NotNil(t, out)
	tokens := out.GetTokensWithoutWhitespace()
	require.GreaterOrEqual(t, len(tokens), 2)
	require.True(t, tokens[1].HasPartialPosTag("verb"))
	require.False(t, tokens[1].HasPartialPosTag("noun"))

	// particle suffix: була-то uses base "була" matcher
	bulaTo := atr("була-то", [][2]string{
		{"бути", "verb:imperf:past:f"},
		{"була", "noun:inanim:f:v_naz"},
	})
	sent2 := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{sentStart, bulaTo})
	out2 := d.Disambiguate(sent2)
	require.False(t, out2.GetTokensWithoutWhitespace()[1].HasPartialPosTag("noun"))
}

func TestRemoveVmisReadings(t *testing.T) {
	a := atr("зв'язку", [][2]string{
		{"зв'язок", "noun:inanim:m:v_mis"},
		{"зв'язок", "noun:inanim:m:v_rod"},
		{"зв'язка", "noun:inanim:f:v_zna"},
	})
	RemoveVmisReadings(a)
	require.False(t, a.HasPartialPosTag("v_mis"))
	require.True(t, a.HasPartialPosTag("v_rod"))

	// only v_mis → keep
	only := atr("x", [][2]string{{"x", "noun:m:v_mis"}})
	RemoveVmisReadings(only)
	require.True(t, only.HasPartialPosTag("v_mis"))
}

func strPtr(s string) *string { return &s }

func TestRemoveDuplicateLemmas(t *testing.T) {
	dups := map[string][]string{
		"весь": {"ввесь"},
	}
	start := languagetool.NewAnalyzedTokenReadingsList(
		[]*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken("", strPtr("SENT_START"), nil)}, 0)
	a := atr("весь", [][2]string{
		{"весь", "adj:m:v_naz:pron:gen"},
		{"ввесь", "adj:m:v_naz:pron:gen"},
	})
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, a})
	d := NewSimpleDisambiguatorFull(nil, dups)
	out := d.Disambiguate(sent)
	tok := out.GetTokensWithoutWhitespace()[1]
	require.True(t, tok.HasAnyLemma("весь"))
	// ввесь removed
	require.False(t, tok.HasAnyLemma("ввесь"))
}

func TestLoadOfficialDisambigMaps(t *testing.T) {
	rm := LoadDisambigRemoveMap()
	require.NotEmpty(t, rm)
	// line: була булий adj:f:v_.*
	tm, ok := rm["була"]
	require.True(t, ok, "була in remove map")
	require.NotNil(t, tm)

	// reading with lemma булий and adj:f:v_naz should match
	l, p := "булий", "adj:f:v_naz"
	tok := languagetool.NewAnalyzedToken("була", &p, &l)
	require.True(t, tm.Matches(tok))

	dups := LoadDisambigDupsMap()
	require.NotEmpty(t, dups)
	require.Contains(t, dups, "весь")
	require.Contains(t, dups["весь"], "ввесь")

	// default constructor loads maps
	d := NewSimpleDisambiguator()
	require.NotEmpty(t, d.RemoveMap)
	require.NotEmpty(t, d.DupsMap)

	start := languagetool.NewAnalyzedTokenReadingsList(
		[]*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken("", strPtr("SENT_START"), nil)}, 0)
	bula := atr("була", [][2]string{
		{"бути", "verb:imperf:past:f"},
		{"булий", "adj:f:v_naz"},
	})
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, bula})
	out := d.Disambiguate(sent)
	require.True(t, out.GetTokensWithoutWhitespace()[1].HasPartialPosTag("verb"))
	require.False(t, out.GetTokensWithoutWhitespace()[1].HasPartialPosTag("adj"))
}
