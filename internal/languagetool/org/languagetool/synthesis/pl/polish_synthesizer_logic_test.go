package pl

// Algorithm unit tests for PolishSynthesizer overrides (negation, +→RE, getPosTagCorrection
// edge cases). Official Java matrix lives in PolishSynthesizerTest_test.go on Instance().

import (
	"sort"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func plTok(surface, lemma, pos string) *languagetool.AnalyzedToken {
	var p *string
	if pos != "" {
		p = &pos
	}
	var l *string
	if lemma != "" {
		l = &lemma
	} else {
		l = &surface
	}
	return languagetool.NewAnalyzedToken(surface, p, l)
}

// mapLookup builds a DictionaryLookup stand-in for algorithm tests (Java uses dict only).
func mapLookup(entries map[string][]string) func(lemma, posTag string) []string {
	return func(lemma, posTag string) []string {
		return append([]string(nil), entries[lemma+"|"+posTag]...)
	}
}

func TestPolishSynthesizer_ExactAndPlus(t *testing.T) {
	s := NewPolishSynthesizer(nil)
	s.Lookup = mapLookup(map[string][]string{
		"Aaron|subst:sg:gen:m1":   {"Aarona"},
		"Abchaz|subst:sg:inst:m1": {"Abchazem"},
		"toner|subst:sg:gen:m3":   {"tonera"},
	})
	s.PossibleTags = []string{"subst:sg:gen:m1", "subst:sg:gen:m3", "subst:sg:inst:m1"}

	got, err := s.Synthesize(plTok("blablabla", "blablabla", "blablabla"), "blablabla")
	require.NoError(t, err)
	require.Empty(t, got)

	got, err = s.Synthesize(plTok("Aaron", "Aaron", "Aaron"), "subst:sg:gen:m1")
	require.NoError(t, err)
	require.Equal(t, []string{"Aarona"}, got)

	// + forces regexp path
	got, err = s.SynthesizeRE(plTok("toner", "toner", "toner"), "subst:sg:gen:m.*", true)
	require.NoError(t, err)
	require.Equal(t, []string{"tonera"}, got)
}

func TestPolishSynthesizer_Negation(t *testing.T) {
	// lemma "duży" with aff form "duży" under adj:sg:nom:m:pos:aff → nie+duży
	s := NewPolishSynthesizer(nil)
	s.Lookup = mapLookup(map[string][]string{
		"duży|adj:sg:nom:m:pos:aff": {"duży"},
	})
	// token POS has :neg, request pos with :neg
	tok := plTok("nieduży", "duży", "adj:sg:nom:m:pos:neg")
	got, err := s.Synthesize(tok, "adj:sg:nom:m:pos:neg")
	require.NoError(t, err)
	require.Equal(t, []string{"nieduży"}, got)

	// :neg in requested tag only (token POS without :neg still triggers via a)
	tok2 := plTok("duży", "duży", "adj:sg:nom:m:pos")
	got, err = s.Synthesize(tok2, "adj:sg:nom:m:pos:neg")
	require.NoError(t, err)
	require.Equal(t, []string{"nieduży"}, got)
}

func TestPolishSynthesizer_RegexpDedup(t *testing.T) {
	s := NewPolishSynthesizer(nil)
	s.Lookup = mapLookup(map[string][]string{
		"mieć|verb:praet:sg:m1:ter:imperf:refl.nonrefl": {"miał"},
		"mieć|verb:praet:sg:f:ter:imperf:refl.nonrefl":  {"miała"},
		"mieć|verb:praet:sg:n:ter:imperf:refl.nonrefl":  {"miało"},
	})
	s.PossibleTags = []string{
		"verb:praet:sg:m1:ter:imperf:refl.nonrefl",
		"verb:praet:sg:f:ter:imperf:refl.nonrefl",
		"verb:praet:sg:n:ter:imperf:refl.nonrefl",
	}
	got, err := s.SynthesizeRE(plTok("mieć", "mieć", "mieć"), ".*praet:sg.*", true)
	require.NoError(t, err)
	sort.Strings(got)
	require.Equal(t, []string{"miał", "miała", "miało"}, got)
}

func TestPolishSynthesizer_NullPosTag(t *testing.T) {
	s := NewPolishSynthesizer(nil)
	got, err := s.Synthesize(plTok("x", "x", "x"), "")
	require.NoError(t, err)
	require.Nil(t, got)
}

func TestIsNegatedPL_OperatorPrecedence(t *testing.T) {
	// token has :neg, posTag has com → not negated via B path; only if posTag has :neg
	tok := plTok("w", "w", "adj:sg:nom:m:com:neg")
	require.False(t, isNegatedPL(tok, "adj:sg:nom:m:com"))
	require.True(t, isNegatedPL(tok, "adj:sg:nom:m:pos:neg"))
	// Java: token POS null → isNegated stays false even if posTag has :neg
	tokNilPOS := languagetool.NewAnalyzedToken("x", nil, nil)
	require.False(t, isNegatedPL(tokNilPOS, "adj:sg:nom:m:pos:neg"))
}

func TestPolishSynthesizer_ResourcePaths(t *testing.T) {
	s := NewPolishSynthesizer(nil)
	require.Equal(t, ResourceFilename, s.ResourceFileName)
	require.Equal(t, TagsFileName, s.TagFileName)
	require.Equal(t, "pl", s.LangShortCode)
}
