package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestUpperCaseNgramRule(t *testing.T) {
	r := NewUpperCaseNgramRule(nil)
	sent := languagetool.AnalyzePlain("The Dog ran.")
	ms, err := r.Match(sent)
	require.NoError(t, err)
	// "Dog" mid-sentence titlecase
	require.NotEmpty(t, ms)
	require.Equal(t, "dog", ms[0].GetSuggestedReplacements()[0])
}
