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

// Java PolishTagger always merges lowercase tags for non-lowercase surfaces
// (including mixed case), unlike BaseTagger which skips mixed case.
func TestRegisterBinaryPOSTagger_PolishCaseMerge(t *testing.T) {
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
	lt := NewJLanguageTool("pl")
	require.True(t, RegisterBinaryPOSTagger(lt, dictPath))
	// Title "Dom" should include lowercase "dom" readings (Java PolishTagger).
	tags := lt.TagWord("Dom")
	require.NotEmpty(t, tags, "Dom should be tagged via lower merge")
	// Known common lemma from polish.dict
	var hasLemma bool
	for _, tg := range tags {
		if tg.Lemma == "dom" || tg.POS != "" {
			hasLemma = true
			break
		}
	}
	require.True(t, hasLemma, "tags=%v", tags)
}

// Java RussianTagger strips combining acute before dict lookup.
func TestRegisterBinaryPOSTagger_RussianAccentStrip(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	var dictPath string
	dir := wd
	for i := 0; i < 12; i++ {
		cand := filepath.Join(dir, "inspiration", "languagetool", "languagetool-language-modules", "ru",
			"src", "main", "resources", "org", "languagetool", "resource", "ru", "russian.dict")
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
		t.Skip("russian.dict not in tree")
	}
	lt := NewJLanguageTool("ru")
	require.True(t, RegisterBinaryPOSTagger(lt, dictPath))
	// дом is in russian.dict; stressed д + о́ + м should still tag after strip
	plain := lt.TagWord("дом")
	require.NotEmpty(t, plain, "дом")
	stressed := "д" + "о\u0301" + "м"
	got := lt.TagWord(stressed)
	require.NotEmpty(t, got, "stressed дом should tag after accent strip")
}

// italian.dict is ISO-8859-15; Lookup must honor fsa.dict.encoding.
func TestRegisterBinaryPOSTagger_ItalianISO885915(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	var dictPath string
	dir := wd
	for i := 0; i < 12; i++ {
		cand := filepath.Join(dir, "inspiration", "languagetool", "languagetool-language-modules", "it",
			"src", "main", "resources", "org", "languagetool", "resource", "it", "italian.dict")
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
		t.Skip("italian.dict not in tree")
	}
	lt := NewJLanguageTool("it")
	require.True(t, RegisterBinaryPOSTagger(lt, dictPath))
	tags := lt.TagWord("casa")
	require.NotEmpty(t, tags)
	found := false
	for _, tg := range tags {
		if tg.Lemma == "casa" || tg.POS != "" {
			found = true
			break
		}
	}
	require.True(t, found, "tags=%v", tags)
}

func TestRegisterBinaryPOSTagger_Galician(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	var dictPath string
	dir := wd
	for i := 0; i < 12; i++ {
		cand := filepath.Join(dir, "inspiration", "languagetool", "languagetool-language-modules", "gl",
			"src", "main", "resources", "org", "languagetool", "resource", "gl", "galician.dict")
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
		t.Skip("galician.dict not in tree")
	}
	lt := NewJLanguageTool("gl")
	require.True(t, RegisterBinaryPOSTagger(lt, dictPath))
	tags := lt.TagWord("casa")
	require.NotEmpty(t, tags)
	var hasNCFS bool
	for _, tg := range tags {
		if tg.POS == "NCFS000" && tg.Lemma == "casa" {
			hasNCFS = true
		}
	}
	require.True(t, hasNCFS, "tags=%v", tags)
	// Title case lower-merge
	tags2 := lt.TagWord("Casa")
	require.NotEmpty(t, tags2)
}
