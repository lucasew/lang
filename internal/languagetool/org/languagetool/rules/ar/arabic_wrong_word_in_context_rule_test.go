package ar

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestArabicWrongWordInContextRule(t *testing.T) {
	// Upstream wrongWordInContext.txt is currently header-only; construction must not panic.
	rule := NewArabicWrongWordInContextRule(nil)
	require.Equal(t, "ARABIC_WRONG_WORD_IN_CONTEXT", rule.GetID())
	_ = rule.Match(languagetool.AnalyzePlain("من سوء الضن بالله ترك الأمر بالمعروف."))
}
