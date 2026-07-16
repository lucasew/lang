package patterns

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestAbstractTokenBasedRule_CanBeIgnored(t *testing.T) {
	r := NewAbstractTokenBasedRule("ID", "desc", "en", []*PatternToken{Token("hello")})
	require.NotEmpty(t, r.TokenHints)
	// sentence without hello → ignore
	sent := languagetool.AnalyzePlain("world there")
	require.True(t, r.CanBeIgnoredFor(sent))
	sent2 := languagetool.AnalyzePlain("say hello now")
	require.False(t, r.CanBeIgnoredFor(sent2))
}
