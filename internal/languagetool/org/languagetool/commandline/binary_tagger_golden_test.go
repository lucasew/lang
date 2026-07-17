package commandline

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDiscoverEnglishPOSDict(t *testing.T) {
	p := DiscoverEnglishPOSDict(nil)
	if p == "" {
		t.Skip("english.dict not in tree")
	}
	require.FileExists(t, p)
}

func TestCoreTagHook_BinaryPOS(t *testing.T) {
	if DiscoverEnglishPOSDict(nil) == "" {
		t.Skip("english.dict not available")
	}
	var out bytes.Buffer
	err := CoreTagHook(&out, "The houses are big.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	s := out.String()
	require.Contains(t, s, "The houses are big.")
	// surface/lemma/POS for houses
	require.Contains(t, s, "houses/house/")
	require.True(t, strings.Contains(s, "houses/house/NNS") || strings.Contains(s, "houses/house/VBZ"), s)
	// DT for the (case-folded)
	require.Contains(t, s, "The/the/DT")
}

func TestCoreDoctor_ReportsPOSDict(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, CoreDoctor(&buf, nil))
	if p := DiscoverEnglishPOSDict(nil); p != "" {
		require.Contains(t, buf.String(), "english.dict")
	}
}
