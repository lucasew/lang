package en

import (
	"os"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// official EN multiword resources (Java EnglishHybridDisambiguator).
const (
	enMultiwordsRel     = "en/multiwords.txt"
	enSpellingGlobalRel = "spelling_global.txt"
)

var (
	englishHybridOnce sync.Once
	englishHybridInst *EnglishHybridDisambiguator
)

// DefaultEnglishHybridDisambiguator returns a process singleton with official
// MultiWordChunkers (ignore-spelling), matching Java EnglishHybridDisambiguator.
// XML rule disambiguator is optional and left nil until wired (chunkers alone
// cover multiword IGNORE_SPELLING for the speller).
func DefaultEnglishHybridDisambiguator() *EnglishHybridDisambiguator {
	englishHybridOnce.Do(func() {
		englishHybridInst = loadEnglishHybridDisambiguator()
	})
	return englishHybridInst
}

func loadEnglishHybridDisambiguator() *EnglishHybridDisambiguator {
	d := NewEnglishHybridDisambiguator()
	// Java: chunkerGlobal first (spelling_global.txt, tagForNotAddingTags), then multiwords.txt
	if p := spelling.DiscoverSpellingResource(enSpellingGlobalRel); p != "" {
		if c, err := openENMultiWordChunker(p, disambiguation.TagForNotAddingTags); err == nil && c != nil {
			c.SetIgnoreSpelling(true)
			d.GlobalChunker = c
		}
	}
	if p := spelling.DiscoverSpellingResource(enMultiwordsRel); p != "" {
		if c, err := openENMultiWordChunker(p, ""); err == nil && c != nil {
			c.SetIgnoreSpelling(true)
			c.SetRemovePreviousTags(true)
			d.Chunker = c
		}
	}
	return d
}

// openENMultiWordChunker loads MultiWordChunker from a multiwords-style file.
// defaultTag empty → phrase\ttag lines; non-empty → phrase-only with fixed tag (global).
func openENMultiWordChunker(path, defaultTag string) (*disambiguation.MultiWordChunker, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	settings := disambiguation.MultiWordChunkerSettings{
		DefaultTag:            defaultTag,
		AllowFirstCapitalized: true,
		AllowAllUppercase:     true,
	}
	// Java MultiWordChunker.getInstance(..., true, true, false):
	// caseSensitive, allowFirstCapitalized — third false is not ignoreSpelling (set after).
	c, err := disambiguation.NewMultiWordChunkerFromReader(f, settings)
	if err != nil {
		return nil, err
	}
	return c, nil
}
