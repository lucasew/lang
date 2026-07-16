package language

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestArabicMetadata(t *testing.T) {
	require.Equal(t, "ar", Arabic.GetShortCode())
	require.Equal(t, "Arabic", Arabic.GetName())
	require.Contains(t, Arabic.GetCountries(), "SA")
	require.Contains(t, ArabicRelevantRuleIDs(), "HUNSPELL_RULE_AR")
	require.Contains(t, ArabicMaintainers(), "Taha Zerrouki")
}
