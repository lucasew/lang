package de

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// German MultitokenSuggest settings match Java GermanRuleDisambiguator:
//
//	MultiWordChunker.getInstance("/de/multitoken-suggest.txt", true, true, false,
//	  MultiWordChunker.tagForNotAddingTags);
//	multitokenSpeller2.setIgnoreSpelling(true);
//
// Same flags as MultitokenIgnore (allowFirstCapitalized=true).
var germanMultitokenSuggestSettings = disambiguation.MultiWordChunkerSettings{
	AllowFirstCapitalized: true,
	AllowAllUppercase:     true,
	AllowTitlecase:        false,
	DefaultTag:            disambiguation.TagForNotAddingTags,
}

var (
	deMultitokenSuggestOnce sync.Once
	deMultitokenSuggestInst *disambiguation.MultiWordChunker
)

// OpenGermanMultitokenSuggest ports MultiWordChunker.getInstance for
// /de/multitoken-suggest.txt with GermanRuleDisambiguator constructor settings
// and setIgnoreSpelling(true).
func OpenGermanMultitokenSuggest(r io.Reader) (*disambiguation.MultiWordChunker, error) {
	c, err := disambiguation.NewMultiWordChunkerFromReader(r, germanMultitokenSuggestSettings)
	if err != nil {
		return nil, err
	}
	c.SetIgnoreSpelling(true)
	return c, nil
}

// LoadGermanMultitokenSuggestFromPath opens the official multitoken-suggest file
// and builds MultiWordChunker with GermanRuleDisambiguator MultitokenSuggest defaults.
func LoadGermanMultitokenSuggestFromPath(path string) (*disambiguation.MultiWordChunker, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return OpenGermanMultitokenSuggest(f)
}

// GermanMultitokenSuggest returns the process-cached MultiWordChunker for
// official /de/multitoken-suggest.txt (Java multitokenSpeller2 field).
// Nil if the official resource is not discoverable.
func GermanMultitokenSuggest() *disambiguation.MultiWordChunker {
	deMultitokenSuggestOnce.Do(func() {
		p := DiscoverGermanMultitokenSuggest()
		if p == "" {
			return
		}
		c, err := LoadGermanMultitokenSuggestFromPath(p)
		if err != nil || c == nil {
			return
		}
		deMultitokenSuggestInst = c
	})
	return deMultitokenSuggestInst
}

// DiscoverGermanMultitokenSuggest finds official de/multitoken-suggest.txt
// (Java resource /de/multitoken-suggest.txt used by GermanRuleDisambiguator.multitokenSpeller2).
func DiscoverGermanMultitokenSuggest() string {
	if p := os.Getenv("LANG_DE_MULTITOKEN_SUGGEST_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "de",
		"src", "main", "resources", "org", "languagetool", "resource", "de", "multitoken-suggest.txt")
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
