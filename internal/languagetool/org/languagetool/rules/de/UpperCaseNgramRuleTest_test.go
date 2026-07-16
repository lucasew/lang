package de

// Twin of UpperCaseNgramRuleTest (surface Tage/Tagen heuristics).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestUpperCaseNgramRule_Rule(t *testing.T) {
	rule := NewUpperCaseNgramRule(nil)
	matchN := func(s string) int {
		return len(rule.Match(languagetool.AnalyzePlain(s)))
	}
	require.Equal(t, 0, matchN("Nach 5 Tagen war es aus."))
	require.Equal(t, 1, matchN("Nach 5 tagen war es aus."))
	require.Equal(t, 0, matchN("Sie tagen im Hotel."))
	require.Equal(t, 1, matchN("Sie Tagen im Hotel."))
}
