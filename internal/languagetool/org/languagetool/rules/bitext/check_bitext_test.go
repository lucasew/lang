package bitext

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of ToolsTest.testBitextCheck
func TestCheckBitext(t *testing.T) {
	matches := CheckBitext(
		"This is a perfectly good sentence.",
		"To jest całkowicie prawidłowe zdanie.",
		nil,
	)
	require.NotNil(t, matches)

	same := CheckBitext("Hello world.", "Hello world.", nil)
	require.NotEmpty(t, same, "expected bitext matches for identical src/trg")
}

// Twin of Tools.checkBitext order: monolingual target matches first, then bitext.
func TestCheckBitextFull_MergeOrder(t *testing.T) {
	src := languagetool.AnalyzePlain("Hello world.")
	trg := languagetool.AnalyzePlain("Hello world.")
	mono := []languagetool.LocalMatch{
		{FromPos: 0, ToPos: 5, RuleID: "MONO_DEMO", Message: "mono"},
	}
	// SameTranslation builtin fires on identical src/trg
	out := CheckBitextFull(src, trg, "Hello world.", mono, nil)
	require.GreaterOrEqual(t, len(out), 2)
	require.Equal(t, "MONO_DEMO", out[0].RuleID, "Java adds monolingual matches first")
	foundBitext := false
	for _, m := range out[1:] {
		if m.RuleID == "SAME_TRANSLATION" {
			foundBitext = true
		}
	}
	require.True(t, foundBitext, "bitext after mono: %+v", out)
}
