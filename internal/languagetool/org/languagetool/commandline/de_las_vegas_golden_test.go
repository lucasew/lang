package commandline

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Java disambiguation LAS_VEGAS uses case_sensitive="yes" so only "Las Vegas"
// is immunized; grammar LAS_VEGAS must still flag lowercase "las Vegas".
func TestGolden_DE_LasVegasCaseSensitiveImmunize(t *testing.T) {
	text := "Das ist jetzt der letzte Schrei in las Vegas."
	opts := &CommandLineOptions{Language: "de"}
	lt, err := configureCoreLT("de", opts)
	require.NoError(t, err)
	found := false
	for _, m := range lt.Check(text) {
		if m.RuleID == "LAS_VEGAS" {
			found = true
		}
	}
	require.True(t, found, "expected grammar LAS_VEGAS on lowercase las")
}
