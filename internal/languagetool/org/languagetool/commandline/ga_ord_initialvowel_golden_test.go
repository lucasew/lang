package commandline

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGolden_GA_ORD_INITIALVOWEL(t *testing.T) {
	text := "Sa tríú alt, déan cur síos ar a bhfaca siad sa Spáinn."
	lt, err := configureCoreLT("ga", &CommandLineOptions{Language: "ga"})
	require.NoError(t, err)
	found := false
	for _, m := range lt.Check(text) {
		if m.RuleID == "ORD_INITIALVOWEL" {
			found = true
		}
	}
	require.True(t, found, "ORD_INITIALVOWEL should flag tríú alt")
}
