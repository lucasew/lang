package de

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// German MultitokenIgnore settings match Java GermanRuleDisambiguator:
//
//	MultiWordChunker.getInstance("/de/multitoken-ignore.txt", true, true, false,
//	  MultiWordChunker.tagForNotAddingTags);
//	multitokenSpeller.setIgnoreSpelling(true);
var germanMultitokenIgnoreSettings = disambiguation.MultiWordChunkerSettings{
	AllowFirstCapitalized: true,
	AllowAllUppercase:     true,
	AllowTitlecase:        false,
	DefaultTag:            disambiguation.TagForNotAddingTags,
}

var (
	deMultitokenIgnoreOnce sync.Once
	deMultitokenIgnoreInst *disambiguation.MultiWordChunker
)

// OpenGermanMultitokenIgnore ports MultiWordChunker.getInstance for
// /de/multitoken-ignore.txt with GermanRuleDisambiguator constructor settings
// and setIgnoreSpelling(true).
func OpenGermanMultitokenIgnore(r io.Reader) (*disambiguation.MultiWordChunker, error) {
	c, err := disambiguation.NewMultiWordChunkerFromReader(r, germanMultitokenIgnoreSettings)
	if err != nil {
		return nil, err
	}
	c.SetIgnoreSpelling(true)
	return c, nil
}

// LoadGermanMultitokenIgnoreFromPath opens the official multitoken-ignore file
// and builds MultiWordChunker with GermanRuleDisambiguator defaults.
func LoadGermanMultitokenIgnoreFromPath(path string) (*disambiguation.MultiWordChunker, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return OpenGermanMultitokenIgnore(f)
}

// GermanMultitokenIgnore returns the process-cached MultiWordChunker for
// official /de/multitoken-ignore.txt (Java multitokenSpeller field).
// Nil if the official resource is not discoverable.
func GermanMultitokenIgnore() *disambiguation.MultiWordChunker {
	deMultitokenIgnoreOnce.Do(func() {
		p := DiscoverGermanMultitokenIgnore()
		if p == "" {
			return
		}
		c, err := LoadGermanMultitokenIgnoreFromPath(p)
		if err != nil || c == nil {
			return
		}
		deMultitokenIgnoreInst = c
	})
	return deMultitokenIgnoreInst
}

// DiscoverGermanMultitokenIgnore finds official de/multitoken-ignore.txt
// (Java resource /de/multitoken-ignore.txt used by GermanRuleDisambiguator).
func DiscoverGermanMultitokenIgnore() string {
	if p := os.Getenv("LANG_DE_MULTITOKEN_IGNORE_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "de",
		"src", "main", "resources", "org", "languagetool", "resource", "de", "multitoken-ignore.txt")
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for i := 0; i < 14; i++ {
		p := filepath.Join(dir, rel)
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}
