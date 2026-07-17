package commandline

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGolden_WordRepeat(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "This is is wrong.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "ENGLISH_WORD_REPEAT_RULE" {
			found = true
			require.Equal(t, "duplication", f.Type)
			require.Equal(t, "warning", f.Severity)
		}
	}
	require.True(t, found, "%+v", findings)
}
