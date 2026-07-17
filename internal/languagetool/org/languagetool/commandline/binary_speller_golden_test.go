package commandline

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDiscoverEnglishUSDict(t *testing.T) {
	p := DiscoverEnglishUSDict(nil)
	if p == "" {
		t.Skip("en_US.dict not in tree (third_party/english-pos-dict)")
	}
	require.FileExists(t, p)
}

func TestGolden_BinaryEnglishSpeller(t *testing.T) {
	if DiscoverEnglishUSDict(nil) == "" {
		t.Skip("en_US.dict not available")
	}
	// Ensure demo env does not change outcome for map-backed path
	_ = os.Unsetenv("LANG_DEMO_SPELLER")

	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "I recieve the book.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "MORFOLOGIK_RULE_EN_US" {
			found = true
			require.Equal(t, "misspelling", f.Type)
			// common typo map should still suggest
			if f.Suggestion != "" {
				require.Equal(t, "receive", f.Suggestion)
			}
		}
	}
	require.True(t, found, "%+v", findings)

	// known word should not flag
	buf.Reset()
	_, err = CoreGoldenHook(&buf, "I receive the book.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	for _, f := range findings {
		require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
	}
}

func TestCoreDoctor_ReportsSpellerDict(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, CoreDoctor(&buf, nil))
	out := buf.String()
	// when dict is discoverable, doctor mentions it
	if p := DiscoverEnglishUSDict(nil); p != "" {
		require.Contains(t, out, "en_US.dict")
	}
}
