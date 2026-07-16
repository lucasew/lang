package language

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGermanAndRomanceVariants(t *testing.T) {
	require.Equal(t, "German (Germany)", GermanyGerman.GetName())
	v, ok := GermanVariantByCode("de-at")
	require.True(t, ok)
	require.Equal(t, "de-AT", v.ShortCode)
	require.Len(t, AllFrenchVariants(), 4)
	require.Len(t, AllSpanishVariants(), 2)
	require.Len(t, AllPortugueseVariants(), 4)
	require.True(t, SpanishVoseo.Voseo)
}
