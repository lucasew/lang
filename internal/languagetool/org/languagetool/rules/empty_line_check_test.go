package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Optional EN full-check path for EmptyLineRule (not sole proof; see EmptyLineRuleTest_test.go).
// Java: EmptyLineRule is default-off; enable then check four-newline paragraph gap.
func TestEmptyLineRule_ViaCheck(t *testing.T) {
	lt := languagetool.NewJLanguageTool("en")
	RegisterCoreEnglishRules(lt)
	// Java EmptyLineRule is default-off; enable for check coverage.
	lt.EnableRule("EMPTY_LINE")
	// multi-line mode: four newlines marks empty line (Java en-US twin)
	text := "Hello world.\n\n\n\nNext para starts here."
	m := lt.Check(text)
	var hits []languagetool.LocalMatch
	for _, x := range m {
		if x.RuleID == "EMPTY_LINE" {
			hits = append(hits, x)
		}
	}
	require.Len(t, hits, 1, "matches: %+v", m)
	require.Equal(t, 11, hits[0].FromPos)
	require.Equal(t, 12, hits[0].ToPos)
}
