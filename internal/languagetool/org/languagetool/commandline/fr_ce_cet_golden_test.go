package commandline

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGolden_FR_CE_CET_AndMultiword(t *testing.T) {
	lt, err := configureCoreLT("fr", &CommandLineOptions{Language: "fr"})
	require.NoError(t, err)
	// CE_CET cases including a priori (was blocked by multiword A tags)
	for _, text := range []string{
		"Ce a priori.", "Ce arbre.", "Ce homme.", "Ce été.",
	} {
		found := false
		for _, m := range lt.Check(text) {
			if m.RuleID == "CE_CET" {
				found = true
			}
		}
		require.True(t, found, "CE_CET for %q", text)
	}
	// multiword home page still applies N f s / J f s
	var out string
	// tag via Analyze
	sents := lt.Analyze("La home page est belle.")
	require.NotEmpty(t, sents)
	mw := false
	for _, tok := range sents[0].GetTokensWithoutWhitespace() {
		if tok == nil {
			continue
		}
		for _, r := range tok.GetReadings() {
			if r != nil && r.GetPOSTag() != nil {
				p := *r.GetPOSTag()
				if p == "N f s" || p == "J f s" {
					mw = true
				}
			}
		}
	}
	require.True(t, mw, "home page multiword POS%s", out)
}
