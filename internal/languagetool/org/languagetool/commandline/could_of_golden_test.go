package commandline

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGolden_CouldOf(t *testing.T) {
	var buf bytes.Buffer
	n, err := CoreGoldenHook(&buf, "I could of done better.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	require.GreaterOrEqual(t, n, 1)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "EN_COULD_OF" {
			found = true
			require.Equal(t, "could have", f.Suggestion)
			require.Equal(t, "grammar", f.Type)
		}
	}
	require.True(t, found, "%+v", findings)

	tmp := filepath.Join(t.TempDir(), "could.json")
	require.NoError(t, os.WriteFile(tmp, buf.Bytes(), 0o644))
	var cout bytes.Buffer
	n, err = CoreCompareHook(&cout, "I could of done better.", &CommandLineOptions{Language: "en", CompareGoldenPath: tmp})
	require.NoError(t, err)
	require.Equal(t, 0, n)
}
