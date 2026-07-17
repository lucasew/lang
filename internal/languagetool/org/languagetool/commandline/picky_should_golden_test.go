package commandline

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGolden_ShouldOf(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "You should of known better.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "EN_SHOULD_OF" {
			found = true
			require.Equal(t, "should have", f.Suggestion)
			require.Equal(t, "grammar", f.Type)
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_PickyAlot(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "I have alot of work.", &CommandLineOptions{Language: "en", Level: "PICKY"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "EN_A_LOT" {
			found = true
			require.Equal(t, "a lot", f.Suggestion)
		}
	}
	require.True(t, found, "%+v", findings)
}
