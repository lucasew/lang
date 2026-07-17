package commandline

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

// Official upstream EN examples only (re-vendor + soft matcher fixes).
func TestGolden_UpstreamENSoftFixes(t *testing.T) {
	cases := []struct {
		rule, text string
	}{
		{"NON_ENGLISH_CHARACTER_IN_A_WORD", "Can you ѕee it?"},
		{"NON_ENGLISH_CHARACTER_IN_A_WORD", "Do nοt open the window."},
		{"ACCEDE_TO", "He acceded to our demands."},
		{"COMBINE_TOGETHER", "Two things are combined together in this application."},
		{"ALL_OF_SUDDEN", "All of sudden, a unicorn appeared."},
		{"ACCELERATE", "The car accelerated from traffic lights"},
		// Official HAD_HARD example (soft be→was lemma map).
		{"HAD_HARD", "It was really had to do it."},
	}
	for _, tc := range cases {
		tc := tc
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
					break
				}
			}
			require.True(t, found, "want %s in %+v", tc.rule, findings)
		})
	}
}
