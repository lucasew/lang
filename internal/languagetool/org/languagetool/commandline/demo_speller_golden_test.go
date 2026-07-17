package commandline

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGolden_DemoSpellerSuggestions(t *testing.T) {
	prev, had := os.LookupEnv("LANG_DEMO_SPELLER")
	require.NoError(t, os.Setenv("LANG_DEMO_SPELLER", "1"))
	t.Cleanup(func() {
		if had {
			_ = os.Setenv("LANG_DEMO_SPELLER", prev)
		} else {
			_ = os.Unsetenv("LANG_DEMO_SPELLER")
		}
	})

	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "I recieve teh book.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))

	got := map[string]string{}
	for _, f := range findings {
		if f.Rule == "MORFOLOGIK_RULE_EN_US" && f.Suggestion != "" {
			// offset text window is not needed; key by suggestion target
			got[f.Suggestion] = f.Type
			require.Equal(t, "misspelling", f.Type)
			require.Equal(t, "error", f.Severity)
		}
	}
	require.Equal(t, "misspelling", got["receive"], "findings=%+v", findings)
	require.Equal(t, "misspelling", got["the"], "findings=%+v", findings)
}
