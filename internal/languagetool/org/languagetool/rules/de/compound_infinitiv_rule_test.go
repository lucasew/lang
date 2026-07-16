package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestCompoundInfinitivRule(t *testing.T) {
	rule := NewCompoundInfinitivRule(nil)
	matchN := func(s string) int {
		return len(rule.Match(languagetool.AnalyzePlain(s)))
	}
	require.Equal(t, 1, matchN("Ich brachte ihn dazu, mein Zimmer sauber zu machen."))
	require.Equal(t, 1, matchN("Du brauchst nicht bei mir vorbei zu kommen."))
	require.Equal(t, 1, matchN("Ich ging zur Seite, um die alte Dame vorbei zu lassen."))
	// goods (separable prefixes / exceptions — surface should not flag)
	require.Equal(t, 0, matchN("Seine Frau gab vor zu schlafen."))
	require.Equal(t, 0, matchN("Mein Herz hörte auf zu schlagen."))
	require.Equal(t, 0, matchN("Fang an zu zählen."))
	require.Equal(t, 0, matchN("Aber um auf Nummer sicher zu gehen, schrieb er es auf."))
}
