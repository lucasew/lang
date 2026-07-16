package ar

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestArabicWordRepeatRule_Rule(t *testing.T) {
	rule := NewArabicWordRepeatRule(map[string]string{"repetition": "تكرار"})
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("نفذت التعليمات خطوة خطوة."))))
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("هذا فقط فقط مثال."))))
}
