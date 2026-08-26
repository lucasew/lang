package language

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtendedSmallLangsAndNamed(t *testing.T) {
	all := AllExtendedSmallLangs()
	require.Greater(t, len(all), len(AllSmallLangs()))
	require.Equal(t, "km", NewKhmer().GetShortCode())
	require.Equal(t, "ml", NewMalayalam().GetShortCode())
	require.Equal(t, "tl", NewTagalog().GetShortCode())
	require.Equal(t, "en-US", NewAmericanEnglish().ShortCode)
	require.Equal(t, "fr", NewFrench().ShortCode)
	require.Equal(t, "de-DE", NewGermanyGerman().ShortCode)
	require.Equal(t, "es", NewSpanish().ShortCode)
	require.Equal(t, "pt-BR", NewBrazilianPortuguese().ShortCode)
	require.Equal(t, "zh", NewChinese().GetShortCode())
}
