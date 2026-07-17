package en

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func findEnglishPOSDict(t *testing.T) string {
	t.Helper()
	wd, _ := os.Getwd()
	dir := wd
	for i := 0; i < 12; i++ {
		cand := filepath.Join(dir, "third_party", "english-pos-dict", "org", "languagetool", "resource", "en", "english.dict")
		if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
			return cand
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Skip("english.dict not found")
	return ""
}

func TestRegisterBinaryEnglishTagger(t *testing.T) {
	p := findEnglishPOSDict(t)
	lt := languagetool.NewJLanguageTool("en")
	require.True(t, RegisterBinaryEnglishTagger(lt, p))
	require.NotNil(t, lt.TagWord)
	tags := lt.TagWord("houses")
	require.NotEmpty(t, tags)
	var hasNNS bool
	for _, tg := range tags {
		if tg.POS == "NNS" && tg.Lemma == "house" {
			hasNNS = true
		}
	}
	require.True(t, hasNNS, "%+v", tags)
	// case fold
	tags = lt.TagWord("This")
	require.NotEmpty(t, tags)
}
