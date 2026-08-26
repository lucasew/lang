package language

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLanguageTwins(t *testing.T) {
	require.Equal(t, "it", ItalianShortCode())
	require.Equal(t, "pl", PolishShortCode())
	require.Equal(t, "ru", RussianShortCode())
	require.Equal(t, "nl", NewDutch().ShortCode)
	require.Equal(t, "nl-BE", NewBelgianDutch().ShortCode)
}
