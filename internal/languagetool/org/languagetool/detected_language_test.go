package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDetectedLanguage(t *testing.T) {
	d := NewDetectedLanguage("en", "de")
	require.Equal(t, "de", d.String())
	require.Equal(t, "en", d.GetGivenLanguageCode())
	require.Equal(t, "de", d.GetDetectedLanguageCode())
	require.Equal(t, float32(1), d.GetDetectionConfidence())
	require.Nil(t, d.GetDetectionSource())
	src := "cld2"
	d2 := NewDetectedLanguageFull("en", "fr-FR", 0.8, &src)
	require.Equal(t, "cld2", *d2.GetDetectionSource())
	require.Equal(t, float32(0.8), d2.GetDetectionConfidence())
	require.Equal(t, "fr-FR", d2.String())
}
