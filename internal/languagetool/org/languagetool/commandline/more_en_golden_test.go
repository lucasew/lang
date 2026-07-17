package commandline

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGolden_WouldOfMustOf(t *testing.T) {
	for _, tc := range []struct {
		text, rule, sug string
	}{
		{"I would of gone.", "EN_WOULD_OF", "would have"},
		{"You must of seen it.", "EN_MUST_OF", "must have"},
	} {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
					require.Equal(t, "grammar", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}

func TestGolden_IrregardlessPicky(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "Irregardless of that.", &CommandLineOptions{Language: "en", Level: "PICKY"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "EN_IRREGARDLESS" {
			found = true
			require.Equal(t, "regardless", f.Suggestion)
		}
	}
	require.True(t, found, "%+v", findings)
}
