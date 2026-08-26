package de

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenGermanSynthesizerFromDir_Missing(t *testing.T) {
	require.Nil(t, OpenGermanSynthesizerFromDir(""))
	require.Nil(t, OpenGermanSynthesizerFromDir(t.TempDir()))
}

func TestStrictCompoundTokenizeGetStrictInstance(t *testing.T) {
	fn := strictCompoundTokenizeGetStrictInstance()
	require.NotNil(t, fn)
	require.Nil(t, fn(""))
	// WLAN-LAN-Kabel stays whole (hyphen path in getCompoundForms); embedded dict only.
	require.Equal(t, []string{"WLAN-LAN-Kabel"}, fn("WLAN-LAN-Kabel"))
}

func TestOpenGermanSynthesizerFromDir_WithSynthDict(t *testing.T) {
	const rel = "inspiration/languagetool/languagetool-language-modules/de/src/main/resources/org/languagetool/resource/de"
	wd, err := os.Getwd()
	require.NoError(t, err)
	dir := wd
	var root string
	for {
		cand := filepath.Join(dir, rel)
		if st, err := os.Stat(cand); err == nil && st.IsDir() {
			root = cand
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	if root == "" {
		t.Skip("no DE resource dir")
	}
	gs := OpenGermanSynthesizerFromDir(root)
	if gs == nil {
		t.Skip("no german_synth.dict")
	}
	require.NotNil(t, gs.StrictCompoundTokenize)
	require.NotNil(t, gs.Lookup)
	require.NotEmpty(t, gs.PossibleTags)
	require.NotNil(t, gs.NumberSpeller)
}
