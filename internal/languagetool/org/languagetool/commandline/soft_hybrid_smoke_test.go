package commandline

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestGolden_SoftHybrid_FR_MultiwordRemovePreviousTags checks Java FrenchHybridDisambiguator:
// multiwords.txt + setRemovePreviousTags(true) so "home page" → N f s / J f s (getNextPosTag).
func TestGolden_SoftHybrid_FR_MultiwordRemovePreviousTags(t *testing.T) {
	mw := DiscoverLanguageMultiwords(nil, "fr")
	xml := DiscoverLanguageSoftDisambiguationXML(nil, "fr")
	if mw == "" || xml == "" {
		t.Skip("FR multiwords or soft disambig XML missing")
	}
	var out bytes.Buffer
	err := CoreTagHook(&out, "La home page est belle.", &CommandLineOptions{Language: "fr"})
	require.NoError(t, err)
	s := out.String()
	// Expect multiword POS from French multiwords.txt (removePreviousTags form).
	require.True(t, strings.Contains(s, "N f s") || strings.Contains(s, "<N f s>"),
		"expected FR multiword tag N f s in tag dump:\n%s", s)
	// Next token of multiword span gets J… (Java getNextPosTag for "N …")
	require.True(t, strings.Contains(s, "J f s") || strings.Contains(s, "J "),
		"expected FR multiword next-tag J… after removePreviousTags:\n%s", s)
}

// TestGolden_SoftHybrid_PL_Installs ensures Polish soft hybrid (XML then multiwords) wires without panic.
func TestGolden_SoftHybrid_PL_Installs(t *testing.T) {
	if DiscoverLanguageMultiwords(nil, "pl") == "" && DiscoverLanguageSoftDisambiguationXML(nil, "pl") == "" {
		t.Skip("PL soft hybrid resources missing")
	}
	var out bytes.Buffer
	err := CoreTagHook(&out, "To jest test.", &CommandLineOptions{Language: "pl"})
	require.NoError(t, err)
	require.NotEmpty(t, out.String())
}
