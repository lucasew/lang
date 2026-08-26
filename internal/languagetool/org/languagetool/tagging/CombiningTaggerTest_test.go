package tagging

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.tagging.CombiningTaggerTest.

func xxResource(t *testing.T, name string) *os.File {
	t.Helper()
	// walk up to module root
	dir, err := os.Getwd()
	require.NoError(t, err)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
	p := filepath.Join(dir, "inspiration/languagetool/languagetool-core/src/test/resources/org/languagetool/resource/xx", name)
	f, err := os.Open(p)
	require.NoError(t, err)
	return f
}

func getCombiningTagger(t *testing.T, overwrite bool, removalName string) *CombiningTagger {
	t.Helper()
	f1 := xxResource(t, "added1.txt")
	defer f1.Close()
	f2 := xxResource(t, "added2.txt")
	defer f2.Close()
	tagger1, err := NewManualTagger(f1)
	require.NoError(t, err)
	tagger2, err := NewManualTagger(f2)
	require.NoError(t, err)
	var removal WordTagger
	if removalName != "" {
		fr := xxResource(t, removalName)
		defer fr.Close()
		rt, err := NewManualTagger(fr)
		require.NoError(t, err)
		removal = rt
	}
	return NewCombiningTaggerWithRemoval(tagger1, tagger2, removal, overwrite)
}

func getAsString(result []TaggedWord) string {
	var sb strings.Builder
	for _, tw := range result {
		sb.WriteString(tw.GetLemma())
		sb.WriteByte('/')
		sb.WriteString(tw.GetPosTag())
		sb.WriteByte('\n')
	}
	return sb.String()
}

func TestCombiningTagger_TagNoOverwrite(t *testing.T) {
	tagger := getCombiningTagger(t, false, "")
	require.Equal(t, 0, len(tagger.Tag("nosuchword")))
	result := tagger.Tag("fullform")
	require.Equal(t, 2, len(result))
	asString := getAsString(result)
	require.Contains(t, asString, "baseform1/POSTAG1")
	require.Contains(t, asString, "baseform2/POSTAG2")
}

func TestCombiningTagger_TagOverwrite(t *testing.T) {
	tagger := getCombiningTagger(t, true, "")
	require.Equal(t, 0, len(tagger.Tag("nosuchword")))
	result := tagger.Tag("fullform")
	require.Equal(t, 1, len(result))
	asString := getAsString(result)
	require.Contains(t, asString, "baseform2/POSTAG2")
}

func TestCombiningTagger_TagRemoval(t *testing.T) {
	tagger := getCombiningTagger(t, false, "removed.txt")
	require.Equal(t, 0, len(tagger.Tag("nosuchword")))
	result := tagger.Tag("fullform")
	asString := getAsString(result)
	require.False(t, strings.Contains(asString, "baseform1/POSTAG1"))
	require.Contains(t, asString, "baseform2/POSTAG2")
}

func TestCombiningTagger_InvalidFile(t *testing.T) {
	f := xxResource(t, "added-invalid.txt")
	defer f.Close()
	_, err := NewManualTagger(f)
	require.Error(t, err)
}
