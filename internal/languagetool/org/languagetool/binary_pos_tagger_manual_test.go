package languagetool

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// Java BaseTagger loads resource/{lang}/added.txt via CombiningTagger.
func TestRegisterBinaryPOSTagger_ManualAdded(t *testing.T) {
	// Walk to polish.dict under inspiration.
	wd, err := os.Getwd()
	require.NoError(t, err)
	var dictPath string
	dir := wd
	for i := 0; i < 12; i++ {
		cand := filepath.Join(dir, "inspiration", "languagetool", "languagetool-language-modules", "pl",
			"src", "main", "resources", "org", "languagetool", "resource", "pl", "polish.dict")
		if st, e := os.Stat(cand); e == nil && st.Mode().IsRegular() {
			dictPath = cand
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	if dictPath == "" {
		t.Skip("polish.dict not in tree")
	}
	// Official pl/added.txt has wieczora → wieczór subst:sg:gen:m3
	lt := NewJLanguageTool("pl")
	require.True(t, RegisterBinaryPOSTagger(lt, dictPath))
	require.NotNil(t, lt.TagWord)
	tags := lt.TagWord("wieczora")
	require.NotEmpty(t, tags, "manual added.txt reading for wieczora")
	found := false
	for _, tg := range tags {
		if tg.POS == "subst:sg:gen:m3" && tg.Lemma == "wieczór" {
			found = true
			break
		}
	}
	require.True(t, found, "tags=%v", tags)
}
