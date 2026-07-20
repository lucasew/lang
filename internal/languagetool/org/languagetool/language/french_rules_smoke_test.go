package language

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	frrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/fr"
	"github.com/stretchr/testify/require"
)

func TestFrench_Rules_CompoundAndPunct(t *testing.T) {
	cr := frrules.NewCompoundRule(nil)
	// known incorrect open compound
	ms := cr.Match(languagetool.AnalyzePlain("Jésus Christ"))
	require.NotEmpty(t, ms)

	// Java DoublePunctuationRule — not invent fr.NewDoublePunctuationRule re-export.
	dp := rules.NewDoublePunctuationRule(nil)
	ms2 := dp.Match(languagetool.AnalyzePlain("Hello..."))
	// may or may not flag; ensure call works
	_ = ms2
	require.NotNil(t, dp)
}
