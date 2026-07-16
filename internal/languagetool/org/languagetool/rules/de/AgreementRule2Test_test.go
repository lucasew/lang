package de

// Twin of AgreementRule2Test (surface ADJ+SUB at sentence start).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestAgreementRule2_Rule(t *testing.T) {
	rule := NewAgreementRule2(nil)
	matchN := func(s string) int {
		return len(rule.Match(languagetool.AnalyzePlain(s)))
	}
	require.Equal(t, 1, matchN("Kleiner Haus am Waldesrand"))
	require.Equal(t, 0, matchN("Kleines Haus am Waldesrand"))
	require.Equal(t, 1, matchN("Wirtschaftlich Wachstum kommt ins Stocken"))
	require.Equal(t, 0, matchN("Wirtschaftliches Wachstum kommt ins Stocken"))
	require.Equal(t, 0, matchN("Deutscher Taschenbuch Verlag expandiert"))
	require.Equal(t, 1, matchN("Deutscher Taschenbuch"))
}
