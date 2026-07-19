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

func TestStrictCompoundTokenizeFromDir_FailClosedEmpty(t *testing.T) {
	fn := strictCompoundTokenizeFromDir("")
	require.NotNil(t, fn)
	// Only LT extras / exceptions may split; empty lemma → nil
	require.Nil(t, fn(""))
}

func TestOpenGermanSynthesizerFromDir_WithSynthDict(t *testing.T) {
	// Discover inspiration resource/de when present; skip if no german_synth.dict
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
	// often no german_synth.dict in checkout — fail-closed nil is correct
	gs := OpenGermanSynthesizerFromDir(root)
	if gs == nil {
		t.Skip("no german_synth.dict")
	}
	require.NotNil(t, gs.StrictCompoundTokenize)
}
