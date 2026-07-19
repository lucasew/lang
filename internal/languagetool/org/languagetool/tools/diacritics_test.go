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
}
