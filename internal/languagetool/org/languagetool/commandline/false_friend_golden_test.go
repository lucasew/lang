package commandline

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func softFalseFriendsPath(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	// commandline → … → repo root
	p := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../testdata/false-friends-soft.xml"))
	require.FileExists(t, p)
	return p
}

func TestGolden_FalseFriends(t *testing.T) {
	ff := softFalseFriendsPath(t)
	cases := []struct {
		lang, mother, text, rule, sug string
	}{
		{"en", "fr", "My ability is great.", "ABILITY", "aptitude"},
		{"en", "de", "A gift for you.", "GIFT", "Geschenk"},
		{"en", "es", "The actual problem.", "ACTUAL", "real"},
		{"en", "es", "Go to the library.", "LIBRARY", "biblioteca"},
		{"en", "es", "Eventual success.", "EVENTUAL", "final"},
		{"de", "en", "Gift ist giftig.", "GIFT", "poison"},
	}
	for _, tc := range cases {
		t.Run(tc.rule+"_"+tc.lang+"_"+tc.mother, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := CoreGoldenHook(&buf, tc.text, &CommandLineOptions{
				Language:         tc.lang,
				MotherTongue:     tc.mother,
				FalseFriendsFile: ff,
			})
			require.NoError(t, err)
			var findings []Finding
			require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
			found := false
			for _, f := range findings {
				if f.Rule == tc.rule {
					found = true
					require.Equal(t, "misspelling", f.Type, "%+v", f)
					require.Equal(t, "error", f.Severity)
					require.Equal(t, tc.sug, f.Suggestion)
				}
			}
			require.True(t, found, "%+v", findings)
		})
	}
}
