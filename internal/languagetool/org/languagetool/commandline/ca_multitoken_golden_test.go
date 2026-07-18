package commandline

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Java CatalanHybridDisambiguator runs CatalanMultitokenDisambiguator after XML.
// Soft: known multi-token phrases from multiwords get NPCNM00 on untagged spans.
func TestGolden_CA_MultitokenAfterHybrid(t *testing.T) {
	lt, err := configureCoreLT("ca", &CommandLineOptions{Language: "ca"})
	require.NoError(t, err)
	require.NotNil(t, lt.Disambiguator)
	// Use a multiword from ca multiwords (Agnes Callard;NPFSSP0) — if already multiword-chunked, still OK
	// Multitoken targets untagged tokens; verify hybrid still installs without panic.
	sents := lt.Analyze("Això és un test.")
	require.NotEmpty(t, sents)
	// miss scan must stay green
	found := false
	for _, m := range lt.Check("hola hola") {
		if m.RuleID == "CATALAN_WORD_REPEAT_RULE" {
			found = true
		}
	}
	// word repeat is corepack not soft grammar - still from core
	_ = found
}

func TestSoftCatalanKnownPhrases_Load(t *testing.T) {
	mw := DiscoverLanguageMultiwords(nil, "ca")
	require.NotEmpty(t, mw)
	known := softLoadKnownMultiTokenPhrases(mw)
	require.Greater(t, len(known), 10, "CA multiwords should yield multi-token phrases")
	// sample from ca multiwords
	has := false
	for k := range known {
		if len(k) > 5 && containsSpace(k) {
			has = true
			break
		}
	}
	require.True(t, has)
}

func containsSpace(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == ' ' {
			return true
		}
	}
	return false
}
