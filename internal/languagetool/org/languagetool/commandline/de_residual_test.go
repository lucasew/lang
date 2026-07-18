package commandline

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

// Upstream DE soft residual probe: examples copied from
// testdata/upstream/goldens/de-examples.json (inspiration/languagetool).
// Do not invent cases — only re-run official incorrect examples.
func TestGolden_UpstreamDEResiduals(t *testing.T) {
	cases := []struct{ rule, text string }{
		{"PREISE", "Steigende Priese belasten den Konsum."},
		{"PREISE", "Steigende Prise belasten den Konsum."},
		{"GESAGTE_HABEN", "Er vergaß was sie gesagte hatte."},
		{"IN_MEHREN", "Ich war bereits in mehren Ländern."},
		{"ZWEIFEL_SUBST", "Im zweifel für den Angeklagten."},
		{"TEILEN", "Aus beiden teilen wird ein neues Teil."},
		{"LATEINISCHE_TAENZE", "die Geschichte der lateinischen Tänze"},
		{"IM_VERGANGEN_JAHR", "Da war ich im vergangen Jahr."},
		{"SICH_AUSDRUCKEN", "Da habe ich mich ungeschickt ausgedruckt."},
		{"GUT_TUN_GUTTUN", "Die Kur hat mir wohl getan."},
		{"ALARM_AUSLOSEN_SPELLING_RULE", "Ich habe den Alarm ausgelost."},
		{"TAEGLICHER_ALLTAG", "Im täglichen Alltag kommt es häufig vor."},
		{"MITTELSTAENDIGES_UNTERNEHMEN", "Peter ist der Chef eines mittelständigen Unternehmens."},
		{"STELLT_FEST", "Später stellte man Fest, dass er krank war."},
		{"ERSTEN_MAIN", "Ich will noch mal auf den 18. Main hinweisen."},
		{"DATUM_VON_BIS", "Die Messe findet vom 12. bis 11. Januar statt."},
		{"WEHREND", "Hätte es sowas wehrend der Schulzeit gegeben!"},
		{"ZAHL_LANG_NOMEN", "Die 2 hohe Hürde wurde ihm zum Verhängnis."},
	}
	var miss []string
	for _, tc := range cases {
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
		if found {
			t.Logf("OK   %s", tc.rule)
		} else {
			t.Logf("MISS %s :: %q", tc.rule, tc.text)
			miss = append(miss, tc.rule)
		}
	}
	require.Empty(t, miss, "still missing: %v", miss)
}
