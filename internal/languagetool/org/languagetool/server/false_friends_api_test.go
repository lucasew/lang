package server

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApiV2_MotherTongueFalseFriends(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	// server → languagetool → org → languagetool → internal → module root (5)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../.."))
	path := filepath.Join(root, "testdata/false-friends-soft.xml")
	require.FileExists(t, path)
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
