package commandline

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGolden_MultiwordIgnoresSpelling(t *testing.T) {
	// Soft multiword chunker marks tokens IgnoreSpelling; binary/map speller must skip them.
	texts := []string{
		"I saw the Taj Mahal.",
		"I live in New York.",
		"The status quo is fine.",
	}
	for _, text := range texts {
		t.Run(text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", f.Rule, "%+v", findings)
			}
		})
	}
}

func TestGolden_MultiwordStillSpellsOtherTypos(t *testing.T) {
	// Ensure we still flag real typos outside multiwords
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "I live in New York and recieve mail.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "MORFOLOGIK_RULE_EN_US" && f.Suggestion == "receive" {
			found = true
		}
	}
	// receive map/binary suggestion; if binary only without map still spell-flag
	if !found {
		for _, f := range findings {
			if f.Rule == "MORFOLOGIK_RULE_EN_US" {
				found = true
			}
		}
	}
	require.True(t, found, "%+v", findings)
}
