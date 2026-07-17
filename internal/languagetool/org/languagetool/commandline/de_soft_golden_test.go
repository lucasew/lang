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

func TestGolden_DESoftExtra(t *testing.T) {
	cases := []struct {
		text, rule string
	}{
		{"Ich glaube das es stimmt.", "DE_SOFT_GLAUBE_DAS"},
		{"Seit ihr bereit?", "DE_SOFT_SEIT_SEID"},
		{"Wir wollen uns wider sehen.", "DE_SOFT_WIDER_WIEDER"},
		{"Er ist größer als wie sie.", "DE_SOFT_ALS_WIE"},
		{"Ich weiß das es geht.", "DE_SOFT_WEISS_DAS"},
		{"Wegen dem Wetter bleiben wir.", "DE_SOFT_WEGEN_DEM"},
		{"Trotz dem Regen gehen wir.", "DE_SOFT_TROTZ_DEM"},
		{"Das selbe Problem erneut.", "DE_SOFT_DAS_SELBE"},
		{"Das gleiche Problem erneut.", "DE_SOFT_DAS_GLEICHE"},
		{"Das ist der Standart.", "DE_SOFT_STANDART"},
		{"Für den selben Preis.", "DE_SOFT_DENSELBEN"},
		{"Er sagt das es stimmt.", "DE_SOFT_SAGT_DAS"},
		{"Sie meint das es geht.", "DE_SOFT_MEINT_DAS"},
		{"Ich hoffe das du kommst.", "DE_SOFT_HOFFEN_DAS"},
		{"Laut dem Bericht stimmt es.", "DE_SOFT_LAUT_DEM"},
		{"Hallo!!", "DE_SOFT_DOUBLE_BANG"},
	}
	for _, tc := range cases {
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
					if tc.rule == "DE_SOFT_DOUBLE_BANG" {
						require.Equal(t, "typographical", f.Type)
					} else {
						require.Equal(t, "grammar", f.Type)
					}
					require.Contains(t, f.URL, "lang=de")
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}
