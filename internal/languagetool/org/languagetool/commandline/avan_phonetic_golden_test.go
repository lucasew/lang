package commandline

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGolden_AvsAn_PhoneticExceptions(t *testing.T) {
	// correct usages should not flag EN_A_VS_AN
	good := []string{
		"This is an hour.",
		"This is a university.",
		"This is a European car.",
		"This is a one-time offer.",
		"He is an honest man.",
	}
	for _, text := range good {
		t.Run("good_"+text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			for _, f := range findings {
				require.NotEqual(t, "EN_A_VS_AN", f.Rule, "%+v", findings)
			}
		})
	}

	// wrong usages should flag with the right suggestion
	bad := []struct{ text, sug string }{
		{"This is a hour.", "an"},
		{"This is an university.", "a"},
		{"This is an European car.", "a"},
		{"This is an one-time offer.", "a"},
		{"He is a honest man.", "an"},
	}
	for _, tc := range bad {
		t.Run("bad_"+tc.text, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == "EN_A_VS_AN" {
					found = true
					require.Equal(t, tc.sug, f.Suggestion)
					// Java AvsAnRule: ITSIssueType.Misspelling
			require.Equal(t, "misspelling", f.Type)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}
