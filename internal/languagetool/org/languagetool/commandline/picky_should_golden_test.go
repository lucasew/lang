package commandline

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGolden_PickyShouldOf(t *testing.T) {
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

func TestGolden_PickyLevelExtras(t *testing.T) {
	cases := []struct {
		text, rule, sug string
	}{
		{"I have alot of work.", "EN_A_LOT", "a lot"},
		{"Irregardless of that, we proceed.", "EN_IRREGARDLESS", "regardless"},
		{"Supposably that is true.", "EN_SUPPOSABLY", "supposedly"},
		{"I ordered an expresso.", "EN_EXPRESSO", "espresso"},
		{"They tried to excape.", "EN_EXCAPE", "escape"},
		{"That was nukeular power.", "EN_NUKEULAR", "nuclear"},
		{"Go to the libary.", "EN_LIBARY", "library"},
		{"A mischievious smile.", "EN_MISCHIEVOUS", "mischievous"},
		{"Please orientate yourself.", "EN_ORIENTATE", "orient"},
		{"Use preventative measures.", "EN_PREVENTATIVE", "preventive"},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{Language: "en", Level: "PICKY"})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					if tc.sug != "" {
						require.Equal(t, tc.sug, f.Suggestion)
					}
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}
