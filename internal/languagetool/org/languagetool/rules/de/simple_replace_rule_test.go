package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSimpleReplaceRule(t *testing.T) {
	rule := NewSimpleReplaceRule(nil)
	matches := rule.Match(languagetool.AnalyzePlain("Jo Biden sprach gestern."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Joe Biden", matches[0].GetSuggestedReplacements()[0])

	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Joe Biden sprach gestern."))))
}
