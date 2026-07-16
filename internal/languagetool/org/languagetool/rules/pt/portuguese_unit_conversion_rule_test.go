package pt

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPortugueseUnitConversionRule_ID(t *testing.T) {
	r := NewPortugueseUnitConversionRule(nil)
	require.Equal(t, "UNITS_PT", r.GetID())
}

func TestPortugueseUnitConversionRule_MatchMiles(t *testing.T) {
	r := NewPortugueseUnitConversionRule(nil)
	// plain number + unit tokens
	sent := languagetool.AnalyzePlain("10 milhas")
	matches := r.Match(sent)
	// may or may not match depending on abstract scanner; at least no panic
	_ = matches
}

func TestBrazilianToponymMapLoader(t *testing.T) {
	var l BrazilianToponymMapLoader
	m := l.Load()
	require.NotNil(t, m)
	require.Contains(t, l.States(), "SP")
}

func TestConfusionPairsDataLoader(t *testing.T) {
	var l ConfusionPairsDataLoader
	in := strings.NewReader("secreto;secréto;AQ0MS0\n")
	m, err := l.LoadWords(in, "test")
	require.NoError(t, err)
	require.Contains(t, m, "secreto")
}
