package tagging

// Twin of languagetool-core/.../tagging/MorfologikTaggerTest.java
// Resource: .../tagging/test.dict (+ test.info)

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

// testDictPath finds inspiration/.../tagging/test.dict by walking up from this file.
func testDictPath(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	dir := filepath.Dir(file)
	const rel = "inspiration/languagetool/languagetool-core/src/test/resources/org/languagetool/tagging/test.dict"
	for i := 0; i < 14; i++ {
		cand := filepath.Join(dir, rel)
		if _, err := os.Stat(cand); err == nil {
			return cand
		}
		dir = filepath.Dir(dir)
	}
	t.Fatalf("official test.dict not found (walk-up for %s)", rel)
	return ""
}

// TestMorfologikTagger_Tag twins MorfologikTaggerTest.testTag against official test.dict.
func TestMorfologikTagger_Tag(t *testing.T) {
	dictPath := testDictPath(t)
	tagger := OpenMorfologikTagger(dictPath)
	require.NotNil(t, tagger, "OpenMorfologikTagger(%s)", dictPath)

	result1 := tagger.Tag("lowercase")
	require.Len(t, result1, 2)
	require.Equal(t, "lclemma", result1[0].GetLemma())
	require.Equal(t, "POS1", result1[0].GetPosTag())
	require.Equal(t, "lclemma2", result1[1].GetLemma())
	require.Equal(t, "POS1a", result1[1].GetPosTag())

	result2 := tagger.Tag("Lowercase")
	require.Len(t, result2, 0)

	result3 := tagger.Tag("schön")
	require.Len(t, result3, 1)
	require.Equal(t, "testlemma", result3[0].GetLemma())
	require.Equal(t, "POSTEST", result3[0].GetPosTag())

	noResult := tagger.Tag("noSuchWord")
	require.Len(t, noResult, 0)
}
