package srx

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Twin of GermanSRX "d. h." case: single-letter abbreviation chain needs overlapping
// beforebreak matches (FindAllStringIndex is non-overlapping and would skip "h.").
func TestGerman_DH_NoSplit(t *testing.T) {
	doc, err := DefaultDocument()
	require.NoError(t, err)
	text := "Er kannte eine Unmenge Quellen, aus denen er schöpfen konnte, d. h. natürlich, wo er durch Arbeit sich etwas verdienen konnte."
	parts := doc.Split(text, "de", "_two")
	require.Equal(t, []string{text}, parts)
}

func TestGerman_Bspw_ZB_NoSplit(t *testing.T) {
	doc, err := DefaultDocument()
	require.NoError(t, err)
	for _, s := range []string{
		"Dieser Satz ist bspw. okay so.",
		"Dieser Satz ist z.B. okay so.",
		"Dies ist, z. B., ein Satz.",
	} {
		parts := doc.Split(s, "de", "_two")
		require.Equal(t, 1, len(parts), s)
	}
}
