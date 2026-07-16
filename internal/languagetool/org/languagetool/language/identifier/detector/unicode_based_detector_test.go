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
