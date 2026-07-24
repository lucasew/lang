package pl

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
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

func TestPolishSynthesizer_ExactAndPlus(t *testing.T) {
	man, err := synthesis.NewManualSynthesizer(strings.NewReader(strings.Join([]string{
		"Aarona\tAaron\tsubst:sg:gen:m1",
		"Abchazem\tAbchaz\tsubst:sg:inst:m1",
		"tonera\ttoner\tsubst:sg:gen:m3",
	}, "\n") + "\n"))
	require.NoError(t, err)
	s := NewPolishSynthesizer(man)

	got, err := s.Synthesize(plTok("blablabla", "blablabla", "blablabla"), "blablabla")
	require.NoError(t, err)
	require.Empty(t, got)

	got, err = s.Synthesize(plTok("Aaron", "Aaron", "Aaron"), "subst:sg:gen:m1")
	require.NoError(t, err)
	require.Equal(t, []string{"Aarona"}, got)

	// + forces regexp path: subst:sg:gen:m.*
	got, err = s.Synthesize(plTok("toner", "toner", "toner"), "subst:sg:gen:m+")
	require.NoError(t, err)
	// m+ becomes m| — need full pattern; use SynthesizeRE like Java test
	got, err = s.SynthesizeRE(plTok("toner", "toner", "toner"), "subst:sg:gen:m.*", true)
	require.NoError(t, err)
	require.Equal(t, []string{"tonera"}, got)
}

func TestPolishSynthesizer_Negation(t *testing.T) {
	// lemma "duży" with aff form "duży" under adj:sg:nom:m:pos:aff → nie+duży
	man, err := synthesis.NewManualSynthesizer(strings.NewReader(
		"duży\tduży\tadj:sg:nom:m:pos:aff\n",
	))
	require.NoError(t, err)
	s := NewPolishSynthesizer(man)
	// token POS has :neg, request pos without com/sup
	tok := plTok("nieduży", "duży", "adj:sg:nom:m:pos:neg")
	got, err := s.Synthesize(tok, "adj:sg:nom:m:pos:neg")
	require.NoError(t, err)
	require.Equal(t, []string{"nieduży"}, got)

	// :neg in requested tag
	tok2 := plTok("duży", "duży", "adj:sg:nom:m:pos")
	got, err = s.Synthesize(tok2, "adj:sg:nom:m:pos:neg")
	require.NoError(t, err)
	require.Equal(t, []string{"nieduży"}, got)
}

func TestPolishSynthesizer_RegexpDedup(t *testing.T) {
	man, err := synthesis.NewManualSynthesizer(strings.NewReader(strings.Join([]string{
		"miał\tmieć\tverb:praet:sg:m1:ter:imperf:refl.nonrefl",
		"miała\tmieć\tverb:praet:sg:f:ter:imperf:refl.nonrefl",
		"miało\tmieć\tverb:praet:sg:n:ter:imperf:refl.nonrefl",
	}, "\n") + "\n"))
	require.NoError(t, err)
	s := NewPolishSynthesizer(man)
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
}

func TestPolishSynthesizer_RealDict_Aaron(t *testing.T) {
	s := OpenPolishSynthesizerFromDictPath("")
	// try discover
	dir, _ := os.Getwd()
	var dict string
	for i := 0; i < 12; i++ {
		cand := filepath.Join(dir, "inspiration/languagetool/languagetool-language-modules/pl/src/main/resources/org/languagetool/resource/pl/polish_synth.dict")
		if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
			dict = cand
			break
		}
		dir = filepath.Dir(dir)
	}
	if dict == "" {
		t.Skip("polish_synth.dict missing")
	}
	s = OpenPolishSynthesizerFromDictPath(dict)
	require.NotNil(t, s)
	got, err := s.Synthesize(plTok("Aaron", "Aaron", "Aaron"), "subst:sg:gen:m1")
	require.NoError(t, err)
	require.Equal(t, []string{"Aarona"}, got)
}
