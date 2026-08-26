package dev

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFilterFileByLanguage(t *testing.T) {
	in := "Hello world\nBonjour le monde\nAnother English line\n"
	var out strings.Builder
	skip, err := FilterFileByLanguage(strings.NewReader(in), &out, "en",
		func(line string) *DetectedLang {
			if strings.Contains(line, "Bonjour") {
				return &DetectedLang{ShortCode: "fr", Confidence: 0.99}
			}
			return &DetectedLang{ShortCode: "en", Confidence: 0.9}
		}, 0.95)
	require.NoError(t, err)
	require.Equal(t, 1, skip)
	require.Contains(t, out.String(), "Hello world")
	require.NotContains(t, out.String(), "Bonjour")
	require.Contains(t, out.String(), "Another English")
}
