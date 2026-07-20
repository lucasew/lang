package tools

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRemoveDiacritics(t *testing.T) {
	require.Equal(t, "cafe", RemoveDiacritics("café"))
	require.Equal(t, "papa", RemoveDiacritics("papá"))
	require.Equal(t, "nino", RemoveDiacritics("niño"))
	require.Equal(t, "uber", RemoveDiacritics("über"))
	require.Equal(t, "casa", RemoveDiacritics("casa"))
	// NFD path: precomposed caron / other scripts not limited to Romance invent map
	require.Equal(t, "S", RemoveDiacritics("Š"))
	require.Equal(t, "c", RemoveDiacritics("č"))
	require.True(t, HasDiacritics("café"))
	require.False(t, HasDiacritics("cafe"))
	require.True(t, EqualsIgnoreCaseAndDiacritics("CAFÉ", "cafe"))
	require.False(t, EqualsIgnoreCaseAndDiacritics("casa", "caso"))
}
