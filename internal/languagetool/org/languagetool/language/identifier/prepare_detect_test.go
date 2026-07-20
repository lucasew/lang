package identifier

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrepareDetectLanguage(t *testing.T) {
	p := PrepareDetectLanguage("hello", []string{"nb"}, []string{"en"}, nil)
	require.NotNil(t, p)
	require.Contains(t, p.AdditionalLangs, "no") // nb→no
	require.Contains(t, p.PreferredLangs, "en")
}

func TestPrepareDetectLanguage_VariantPanic(t *testing.T) {
	require.Panics(t, func() {
		_ = PrepareDetectLanguage("x", []string{}, []string{"en-US"}, nil)
	})
}

func TestPrepareDetectLanguage_DominantUnsupported(t *testing.T) {
	p := PrepareDetectLanguage("x", []string{}, []string{"en"}, func(string) []string { return []string{"th"} })
	require.Nil(t, p)
}

func TestGetHighestScoringResult(t *testing.T) {
	code, sc := GetHighestScoringResult(map[string]float64{"en": 0.2, "de": 0.9})
	require.Equal(t, "de", code)
	require.Equal(t, 0.9, sc)
}
