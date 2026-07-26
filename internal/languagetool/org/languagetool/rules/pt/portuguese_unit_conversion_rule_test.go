package pt

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPortugueseUnitConversionRule_ID(t *testing.T) {
	r := NewPortugueseUnitConversionRule(nil)
	// Java PortugueseUnitConversionRule.getId
	require.Equal(t, "UNIDADES_METRICAS", r.GetID())
}

func TestPortugueseUnitConversionRule_MatchMiles(t *testing.T) {
	r := NewPortugueseUnitConversionRule(nil)
	// Java assertMatches: "A via tem 100 milhas de comprimento." → suggestion with quilômetros
	sent := languagetool.AnalyzePlain("A via tem 100 milhas de comprimento.")
	matches := r.Match(sent)
	require.NotEmpty(t, matches)
	require.NotEmpty(t, matches[0].GetSuggestedReplacements())
	joined := strings.Join(matches[0].GetSuggestedReplacements(), " ")
	require.Contains(t, joined, "quilômetros")
}

func TestPortugueseUnitConversionRule_MatchPounds(t *testing.T) {
	r := NewPortugueseUnitConversionRule(nil)
	// Java: "A carga é de 10.000 libras." → toneladas (Locale.GERMANY thousands)
	sent := languagetool.AnalyzePlain("A carga é de 10.000 libras.")
	matches := r.Match(sent)
	require.NotEmpty(t, matches)
	joined := strings.Join(matches[0].GetSuggestedReplacements(), " ")
	require.Contains(t, joined, "toneladas")
}

func TestBrazilianToponymMapLoader_Smoke(t *testing.T) {
	// Full twins live in brazilian_toponym_map_loader_test.go.
	m := NewBrazilianToponymMapLoader().BuildMap()
	require.Equal(t, 27, len(m))
	require.Contains(t, m["SP"], "são paulo")
}

func TestConfusionPairsDataLoader(t *testing.T) {
	var l ConfusionPairsDataLoader
	in := strings.NewReader("secreto;secréto;AQ0MS0\n")
	m, err := l.LoadWords(in, "test")
	require.NoError(t, err)
	require.Contains(t, m, "secreto")
}
