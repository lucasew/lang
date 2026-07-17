package commandline

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"runtime"
	"strings"
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
		{"en", "de", "I want to become a doctor.", "BECOME", "werden"},
		{"de", "en", "Ich will das Buch bekommen.", "BECOME", "get"},
		{"en", "es", "I felt embarrassed.", "EMBARRASSED", "avergonzado"},
		{"es", "en", "Está embarazada.", "EMBARRASSED", "pregnant"},
		{"en", "fr", "My parents live here.", "PARENTS", "parents (père et mère)"},
		{"fr", "en", "Mes parents sont là.", "PARENTS", "relatives"},
		{"en", "de", "She is sympathetic.", "SYMPATHIC", "mitfühlend"},
		{"de", "en", "Er ist sympathisch.", "SYMPATHIC", "likeable"},
		{"en", "fr", "Soft fabric only.", "FABRIC", "tissu"},
		{"fr", "en", "La fabrique est grande.", "FABRIC", "factory"},
		{"en", "fr", "A strong argument.", "ARGUMENT", "dispute"},
		{"fr", "en", "Un bon argument.", "ARGUMENT", "point / reason"},
		{"en", "es", "A sensible choice.", "SENSIBLE", "prudente / razonable"},
		{"es", "en", "Es muy sensible.", "SENSIBLE", "sensitive"},
		{"en", "fr", "Food preservative only.", "PRESERVATIVE", "conservateur (alimentaire)"},
		{"fr", "en", "Un préservatif.", "PRESERVATIVE", "condom"},
		{"en", "es", "Eventually we left.", "EVENTUALLY", "finalmente"},
		{"es", "en", "Eventualmente iremos.", "EVENTUALLY", "possibly / if necessary"},
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

func TestGolden_FalseFriends_WalkUpDiscover(t *testing.T) {
	// MotherTongue alone discovers testdata/false-friends-soft.xml via walk-up.
	var buf bytes.Buffer
	_, err := CoreGoldenHook(&buf, "A gift for you.", &CommandLineOptions{
		Language:     "en",
		MotherTongue: "de",
	})
	require.NoError(t, err)
	var findings []Finding
	require.NoError(t, json.Unmarshal(buf.Bytes(), &findings))
	found := false
	for _, f := range findings {
		if f.Rule == "GIFT" {
			found = true
			require.Equal(t, "Geschenk", f.Suggestion)
			require.Equal(t, "misspelling", f.Type)
		}
	}
	require.True(t, found, "%+v", findings)
}

func TestGolden_ApplySuggestions_FalseFriend(t *testing.T) {
	ff := softFalseFriendsPath(t)
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "-m", "de", "--falsefriends", ff, "--apply", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "A gift for you.", nil },
		Check:     CoreApplySuggestionsHook,
	}, &out, &errb)
	require.True(t, code == 0 || code == 1 || code == 2, "code=%d err=%s", code, errb.String())
	require.Equal(t, "A Geschenk for you.", strings.TrimSpace(out.String()))
}
