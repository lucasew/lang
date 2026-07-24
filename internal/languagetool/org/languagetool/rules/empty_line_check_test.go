package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestEmptyLineRule_ViaCheck(t *testing.T) {
	lt := languagetool.NewJLanguageTool("en")
	RegisterCoreEnglishRules(lt)
	// Java EmptyLineRule is default-off; enable for check coverage.
	lt.EnableRule("EMPTY_LINE")
	// default mode: four newlines marks empty line
	m := lt.Check("Hello world.\n\n\n\nNext para starts here.")
	var has bool
	for _, x := range m {
		if x.RuleID == "EMPTY_LINE" {
			has = true
		}
	}
	require.True(t, has, "matches: %+v", m)
}
