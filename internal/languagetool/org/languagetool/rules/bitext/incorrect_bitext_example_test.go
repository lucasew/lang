package bitext

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/bitext"
	"github.com/stretchr/testify/require"
)

func TestIncorrectBitextExample(t *testing.T) {
	p := bitext.NewStringPair("Hello", "Hallo")
	e := NewIncorrectBitextExampleWithCorrections(p, []string{"Hi"})
	require.Equal(t, "Hello", e.GetExample().GetSource())
	require.Equal(t, []string{"Hi"}, e.GetCorrections())
	require.Contains(t, e.String(), "Hallo")
}
