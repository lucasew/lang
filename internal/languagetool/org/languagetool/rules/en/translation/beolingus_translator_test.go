package translation

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBeoLingusTranslator(t *testing.T) {
	b := NewBeoLingusTranslator()
	require.NoError(t, b.LoadDict(strings.NewReader("#c\nHaus :: house\nHund|dog\n")))
	got, err := b.Translate("Haus", "de", "en")
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, []string{"house"}, got[0].GetL2())
	require.Contains(t, b.GetMessage(), "German")
}
