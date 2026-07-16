package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDetectedLanguage(t *testing.T) {
	d := NewDetectedLanguage("en", "de")
	require.Equal(t, "de", d.String())
	require.Equal(t, float32(1), d.GetDetectionConfidence())
	src := "cld2"
	d2 := NewDetectedLanguageFull("en", "fr", 0.8, &src)
	require.Equal(t, "cld2", *d2.GetDetectionSource())
}
