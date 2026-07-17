package commandline

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGolden_PhraseReplace(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "Guide tot he Galaxy", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "PHRASE_REPLACE" {
			found = true
			require.Equal(t, "to the", f.Suggestion)
		}
	}
	require.True(t, found, "%+v", findings)
}
