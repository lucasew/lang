package server

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/commandline"
	"github.com/stretchr/testify/require"
)

func TestApiV2_MotherTongueFalseFriends(t *testing.T) {
	// Official Java false-friends.xml only (checklist 1.19 — not testdata/*-soft.xml invent).
	// Unset force env so DiscoverFalseFriendsFile walks to inspiration resources.
	_ = os.Unsetenv("LANG_FALSEFRIENDS_FILE")
	path := commandline.DiscoverFalseFriendsFile(nil)
	if path == "" {
		// Walk from this package to inspiration if discover CWD differs.
		_, file, _, _ := runtime.Caller(0)
		root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../.."))
		cand := filepath.Join(root, "inspiration/languagetool/languagetool-core/src/main/resources/org/languagetool/rules/false-friends.xml")
		if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
			path = cand
		}
	}
	if path == "" {
		t.Skip("official false-friends.xml not found")
	}
	require.NotContains(t, path, "soft.xml")
	t.Setenv("LANG_FALSEFRIENDS_FILE", path)

	api := NewApiV2(nil, nil)
	// Java FalseFriendPatternRule is Tag.picky — only active at Level.PICKY.
	r, err := api.Handle("check", map[string]string{
		"language":     "en",
		"motherTongue": "de",
		"level":        "picky",
		"text":         "This is a gift for you.",
	})
	require.NoError(t, err)
	require.Contains(t, r.Body, "GIFT")
}
