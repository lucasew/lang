package commandline

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
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

func TestGolden_DemoSpellerEditDistance(t *testing.T) {
	// rely on edit-distance fallback (no explicit map entry for "tset")
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
	_, err := CoreGoldenHook(&buf, "This is a tset.", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "MORFOLOGIK_RULE_EN_US" {
			for _, s := range f.Suggestions {
				if s == "test" {
					found = true
				}
			}
			if f.Suggestion == "test" {
				found = true
			}
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_ApplySuggestions_AvsAn(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--apply", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "This is an test.", nil },
		Check:     CoreApplySuggestionsHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Equal(t, "This is a test.", strings.TrimSpace(out.String()))
}
