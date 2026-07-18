package commandline

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGolden_UpstreamENRemainingMisses(t *testing.T) {
	cases := []struct{ rule, text string }{
		{"LOCATED_ON_AT", "The company's name is Denver Inc. and it's located on 11056 Main street."},
		{"DT_PRP", "The dots in the my life."},
		{"BUY_TWO_GET_ONE_FREE", "Buy 2 Get 1 Free!"},
		{"FIGURE_HYPHEN", "He earns a 6 figure salary."},
		{"SEVERAL_OTHER", "Tom and several other did a great job."},
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
