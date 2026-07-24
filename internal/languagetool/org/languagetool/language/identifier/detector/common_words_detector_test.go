package detector

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCommonWordsDetector(t *testing.T) {
	d := NewCommonWordsDetector()
	require.NoError(t, d.LoadWords("en", strings.NewReader("the\nand\n#c\n")))
	require.NoError(t, d.LoadWords("de", strings.NewReader("und\nder\n")))
	m := d.GetKnownWordsPerLanguage("the cat and the dog")
	require.GreaterOrEqual(t, m["en"], 2)
	// Spanish-ish
	m2 := d.GetKnownWordsPerLanguage("nación")
	require.Greater(t, m2["es"], 0)
}
