package ga

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestGAPartialPosTagFiltersRegistered(t *testing.T) {
	for _, class := range []string{
		"org.languagetool.rules.ga.IrishPartialPosTagFilter",
		"org.languagetool.rules.ga.NoDisambiguationIrishPartialPosTagFilter",
	} {
		require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(class), class)
	}
}
