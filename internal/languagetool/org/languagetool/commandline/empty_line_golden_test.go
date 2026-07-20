package commandline

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGolden_EmptyLine(t *testing.T) {
	var buf bytes.Buffer
	// Java EmptyLineRule is setDefaultOff; enable explicitly.
	// default SRX (_two): four newlines = empty line between paragraphs
	opts := &CommandLineOptions{Language: "en"}
	opts.EnabledRules = []string{"EMPTY_LINE"}
	_, err := CoreGoldenHook(&buf, "Hello world.\n\n\n\nNext para starts here.", opts)
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "EMPTY_LINE" {
			found = true
			require.Equal(t, "style", f.Type)
			require.Equal(t, "note", f.Severity)
		}
	}
	require.True(t, found, "%+v", findings)
}
