package en

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVariantInfo(t *testing.T) {
	v := NewVariantInfo("American English", "colour")
	require.Equal(t, "American English", v.GetVariantName())
	require.Equal(t, "colour", v.GetOtherVariant())
}
