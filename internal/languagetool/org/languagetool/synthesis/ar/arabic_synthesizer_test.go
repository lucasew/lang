package ar

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"github.com/stretchr/testify/require"
)

func TestInflectMafoulAndTanwin(t *testing.T) {
	// Java ArabicSynthesizer.inflectMafoulMutlq / inflectAdjectiveTanwinNasb
	m := InflectMafoulMutlq("عمل")
	require.True(t, strings.HasPrefix(m, "عمل"))
	require.Contains(t, m, string(tools.ArabicFathatan))
	require.True(t, strings.HasSuffix(m, string(tools.ArabicAlef)))

	masc := InflectAdjectiveTanwinNasb("قوي", false)
	require.Contains(t, masc, string(tools.ArabicFathatan))
	fem := InflectAdjectiveTanwinNasb("قوي", true)
	require.Contains(t, fem, string(tools.ArabicTehMarbuta))
}

func TestArabicSynthesizer(t *testing.T) {
	man, err := synthesis.NewManualSynthesizer(strings.NewReader("كتب\tكتب\tNxx\n"))
	require.NoError(t, err)
	s := NewArabicSynthesizer(man)
	lemma := "كتب"
	tok := languagetool.NewAnalyzedToken("كتب", nil, &lemma)
	forms, err := s.Synthesize(tok, "Nxx")
	require.NoError(t, err)
	require.Equal(t, []string{"كتب"}, forms)
	require.Equal(t, ArabicSynthDict, s.ResourceFileName)
}
