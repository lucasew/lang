package commandline

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGolden_UpstreamDESoftFixes(t *testing.T) {
	cases := []struct{ rule, text string }{
		{"ALARM_AUSLOSEN_SPELLING_RULE", "Es wurde sofort Großalarm ausgelost."},
		{"SICH_AUSDRUCKEN", "Da habe ich mich nicht klar genug ausgedruckt."},
		{"GUT_TUN_GUTTUN", "Das hat mir gut getan."},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "de"})
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
