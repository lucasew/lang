package fa

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSimpleReplaceRule(t *testing.T) {
	rule := NewSimpleReplaceRule(nil)
	// Example from Java: حاظر → حاضر
	matches := rule.Match(languagetool.AnalyzePlain("وی حاظر به همکاری شد."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "حاضر", matches[0].GetSuggestedReplacements()[0])
}
