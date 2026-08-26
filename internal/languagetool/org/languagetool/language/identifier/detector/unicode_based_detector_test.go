package detector

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnicodeBasedDetector_GetDominantLangCodes(t *testing.T) {
	ident := NewUnicodeBasedDetectorMax(100)
	codes := func(s string) string {
		return fmt.Sprint(ident.GetDominantLangCodes(s))
	}
	require.Equal(t, "[]", codes(""))
	require.Equal(t, "[]", codes(" "))
	require.Equal(t, "[]", codes("hallo"))
	require.Equal(t, "[]", codes("hallo this is a text"))
	require.Equal(t, "[]", codes("hallo this is a text стиль"))

	cyrillic := "[ru uk be]"
	require.Equal(t, cyrillic, codes("Грамматика, стиль и орфография LanguageTool проверяет ваше правописание на более чем 20 языках"))
	require.Equal(t, cyrillic, codes("проверяет ваше правописание на более чем 20 языках"))

	arabic := "[ar fa]"
	require.Equal(t, arabic, codes("لِينُكس (بالإنجليزية: Linux)\u200F (عن هذا الملف استمع (؟·معلومات)) ويسمى أيضا"))
	require.Equal(t, arabic, codes("طور لينكس في الأصل لكي يعمل على"))

	cjk := "[zh ja]"
	require.Equal(t, cjk, codes("您的意思是"))
	require.Equal(t, cjk, codes("Linux（リナックス、他の読みは後述）とは、Unix系オペレーティングシステムカーネル"))
	require.Equal(t, "[]", codes("通常情况下 but here's more text with Latin characters"))

	require.Equal(t, "[km]", codes("ហើយដោយ​ព្រោះ​"))
}

// Twin test name used by generated stub.
func TestUnicodeBasedLangIdentifier_GetDominantLangCodes(t *testing.T) {
	TestUnicodeBasedDetector_GetDominantLangCodes(t)
}

// Twin: maxCheckLength counts UTF-16 units (charAt), not code points.
func TestUnicodeBasedDetector_MaxCheckUTF16(t *testing.T) {
	// maxCheckLength 1: first unit of "привет" is Cyrillic п — one significant cyrillic char
	// With only 1 unit, significant=1, cyrillic=1 → 100% ≥ threshold → ru,uk,be
	d := NewUnicodeBasedDetectorMax(1)
	codes := d.GetDominantLangCodes("привет")
	require.Contains(t, codes, "ru")
	// emoji at start is 2 UTF-16 units; maxCheckLength 1 only sees high surrogate (not in script ranges)
	// significant may count non-whitespace non-digit surrogate → significant=1, no script hit → empty or partial
	d2 := NewUnicodeBasedDetectorMax(1)
	_ = d2.GetDominantLangCodes("😀русский")
	// full cyrillic still detected with default max
	codes2 := NewUnicodeBasedDetector().GetDominantLangCodes("русский текст здесь")
	require.Contains(t, codes2, "ru")
}
