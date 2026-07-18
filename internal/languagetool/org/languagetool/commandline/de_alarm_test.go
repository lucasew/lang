package commandline

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

// Official ALARM_AUSLOSEN_SPELLING_RULE goldens (ausgelost ← auslosen).
func TestGolden_UpstreamDEAlarmAuslosen(t *testing.T) {
	for _, text := range []string{
		"Es wurde sofort Großalarm ausgelost.",
		"Ich habe den Alarm ausgelost.",
	} {
		var buf bytes.Buffer
		_, err := CoreGoldenHook(&buf, text, &CommandLineOptions{Language: "de"})
		require.NoError(t, err)
		var findings []Finding
		require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
		found := false
		for _, f := range findings {
			if f.Rule == "ALARM_AUSLOSEN_SPELLING_RULE" {
				found = true
				break
			}
		}
		require.True(t, found, "text=%q findings=%+v", text, findings)
	}
}
