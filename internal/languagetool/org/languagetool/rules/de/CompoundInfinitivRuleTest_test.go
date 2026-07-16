package de

// Twin of CompoundInfinitivRuleTest (surface particle list).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestCompoundInfinitivRule_Rule(t *testing.T) {
	rule := NewCompoundInfinitivRule(nil)
	matchN := func(s string) int {
		return len(rule.Match(languagetool.AnalyzePlain(s)))
	}
	require.Equal(t, 1, matchN("Ich brachte ihn dazu, mein Zimmer sauber zu machen."))
	require.Equal(t, 1, matchN("Du brauchst nicht bei mir vorbei zu kommen."))
	require.Equal(t, 0, matchN("Fang an zu zählen."))
}
