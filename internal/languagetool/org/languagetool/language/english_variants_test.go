package language

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnglishVariants(t *testing.T) {
	require.Equal(t, "English (US)", AmericanEnglish.GetName())
	require.Equal(t, []string{"US"}, AmericanEnglish.GetCountries())
	v, ok := EnglishVariantByCode("en-gb")
	require.True(t, ok)
	require.Equal(t, "en-GB", v.ShortCode)
	require.Len(t, AllEnglishVariants(), 6)
}
