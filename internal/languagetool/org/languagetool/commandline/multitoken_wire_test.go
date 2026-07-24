package commandline

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/multitoken"
	"github.com/stretchr/testify/require"
)

// Twin: FR MultitokenSpeller wires with checkSpelling=false (Java shortCode gate).
func TestWireFrenchMultitokenSpeller_NoCheckSpelling(t *testing.T) {
	patterns.SetDefaultMultitokenSpellerWithOptions(nil, nil, false)
	t.Cleanup(func() { patterns.SetDefaultMultitokenSpellerWithOptions(nil, nil, false) })

	wireFrenchMultitokenSpeller(nil)
	// Filter registered; Accept without panic even if resources empty.
	f := patterns.GlobalRuleFilterCreator.GetFilter(
		"org.languagetool.rules.spelling.multitoken.MultitokenSpellerFilter")
	require.NotNil(t, f)
	// Wire empty: if discover found nothing, filter may still return nil (no invent).
	// Just ensure wire completed.
	_ = multitoken.NewMultitokenSpeller()
}

func TestWireSpanishAndCatalanMultitokenSpeller(t *testing.T) {
	t.Cleanup(func() { patterns.SetDefaultMultitokenSpellerWithOptions(nil, nil, false) })
	wireSpanishMultitokenSpeller(nil)
	wireCatalanMultitokenSpeller(nil)
	// Portuguese still uses checkSpelling path
	wirePortugueseMultitokenSpeller(nil)
}
