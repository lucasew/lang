package commandline

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

// Official COMBINE_TOGETHER: <match no="1"/> → soft \1 → "combined".
func TestGolden_UpstreamENCombineSuggestion(t *testing.T) {
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "Two things are combined together in this application.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule != "COMBINE_TOGETHER" {
			continue
		}
		found = true
		require.Equal(t, "combined", f.Suggestion, "%+v", f)
	}
	require.True(t, found, "%+v", findings)
}
