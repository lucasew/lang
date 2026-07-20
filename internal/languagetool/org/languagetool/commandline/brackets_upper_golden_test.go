package commandline

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGolden_UnpairedBrackets(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "This is broken (yes.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		// Java EnglishUnpairedBracketsRule.getId() → EN_UNPAIRED_BRACKETS (not bare UNPAIRED_BRACKETS)
		if f.Rule == "EN_UNPAIRED_BRACKETS" {
			found = true
			require.Equal(t, "typographical", f.Type)
			require.Equal(t, "warning", f.Severity)
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_UppercaseSentenceStart(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "lowercase start is wrong.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "UPPERCASE_SENTENCE_START" {
			found = true
			require.Equal(t, "typographical", f.Type)
		}
	}
	require.True(t, found, "%+v", findings)
}
