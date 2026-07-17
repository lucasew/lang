package commandline

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGolden_DESoftDasDass(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "Ich denke das es so ist.", &CommandLineOptions{Language: "de"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "DE_SOFT_DAS_DASS" {
			found = true
			require.Equal(t, "grammar", f.Type)
			require.Equal(t, "error", f.Severity)
			require.Contains(t, f.URL, "lang=de")
			require.Contains(t, f.URL, "DE_SOFT_DAS_DASS")
		}
	}
	require.True(t, found, "%+v", findings)
}
