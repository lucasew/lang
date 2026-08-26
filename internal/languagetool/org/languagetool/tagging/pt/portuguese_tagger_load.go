package pt

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	pttok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/pt"
)

var (
	ptTaggerOnce  sync.Once
	ptPOSDictPath string
	// DefaultPortugueseTagger is the process-wide PortugueseTagger backed by
	// portuguese.dict (Java: new PortugueseTagger()). Nil if dict is missing.
	DefaultPortugueseTagger *PortugueseTagger
)

// DiscoverPortuguesePOSDict finds portuguese.dict (Java resource /pt/portuguese.dict).
// Order: LANG_PORTUGUESE_DICT, third_party/portuguese-pos-dict, inspiration module path.
func DiscoverPortuguesePOSDict() string {
	if p := os.Getenv("LANG_PORTUGUESE_DICT"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	relPaths := []string{
		filepath.Join("third_party", "portuguese-pos-dict", "org", "languagetool", "resource", "pt", "portuguese.dict"),
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "pt",
			"src", "main", "resources", "org", "languagetool", "resource", "pt", "portuguese.dict"),
	}
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for i := 0; i < 14; i++ {
		for _, rel := range relPaths {
			p := filepath.Join(dir, rel)
			if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
				return p
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// EnsureDefaultPortugueseTagger loads portuguese.dict into DefaultPortugueseTagger
// and wires PortugueseWordTokenizer construction to real PortugueseTagger
// (Java: new PortugueseWordTokenizer() → tagger = new PortugueseTagger()).
// Idempotent; no-op if dict missing (fail closed).
// Java BaseTagger: overwriteWithManualTagger() → false → CombiningTagger(overwrite=false).
func EnsureDefaultPortugueseTagger() {
	ptTaggerOnce.Do(func() {
		p := DiscoverPortuguesePOSDict()
		if p == "" {
			return
		}
		ptPOSDictPath = p
		mt := tagging.OpenMorfologikTagger(p)
		if mt == nil {
			return
		}
		// Java BaseTagger.initWordTagger: CombiningTagger when added.txt present
		wt := tagging.WordTagger(mt)
		if manual := loadPTManualTagger(p); manual != nil {
			removal := loadPTManualRemoval(p)
			wt = tagging.NewCombiningTaggerWithRemoval(wt, manual, removal, false)
		}
		// dictLookup = DictionaryLookup(getDictionary()) for -mente / soto- prefixes
		DefaultPortugueseTagger = NewPortugueseTaggerWithDictLookup(wt, mt)
		// Java PortugueseWordTokenizer.wordsToAdd →
		// tagger.tag(Arrays.asList(...)).get(0).isTagged()
		// Instance field path (preferred): NewPortugueseWordTokenizer uses loader.
		pttok.RegisterPortugueseTaggerLoader(portugueseWordTokenizerTagger)
		// Package hook kept for older inject sites (speller / binary POS).
		pttok.IsTaggedPT = portugueseTaggerIsTagged
	})
}

// portugueseWordTokenizerTagger is Java "new PortugueseTagger()" for tokenizer ctor.
func portugueseWordTokenizerTagger() pttok.PortugueseHyphenTagger {
	tg := DefaultPortugueseTagger
	if tg == nil {
		return nil
	}
	return pttok.ATRTagFunc(func(tokens []string) []pttok.PortugueseHyphenReadings {
		atrs := tg.Tag(tokens)
		out := make([]pttok.PortugueseHyphenReadings, len(atrs))
		for i, a := range atrs {
			out[i] = a
		}
		return out
	})
}

// portugueseTaggerIsTagged ports PortugueseTagger.tag([s]).get(0).isTagged().
func portugueseTaggerIsTagged(s string) bool {
	if DefaultPortugueseTagger == nil {
		return false
	}
	atrs := DefaultPortugueseTagger.Tag([]string{s})
	if len(atrs) == 0 || atrs[0] == nil {
		return false
	}
	return atrs[0].IsTagged()
}

// PortuguesePOSDictPath returns the resolved portuguese.dict path after Ensure (may be empty).
func PortuguesePOSDictPath() string {
	EnsureDefaultPortugueseTagger()
	return ptPOSDictPath
}

func loadPTManualTagger(dictPath string) tagging.WordTagger {
	if wt := languagetool.LoadManualTaggerBesideDict(dictPath, []string{"added.txt", "added_custom.txt"}); wt != nil {
		return wt
	}
	if p := walkUpPTResourceDir(); p != "" {
		return languagetool.LoadManualTaggerFromDirs([]string{p}, []string{"added.txt", "added_custom.txt"})
	}
	return nil
}

func loadPTManualRemoval(dictPath string) tagging.WordTagger {
	if wt := languagetool.LoadManualTaggerBesideDict(dictPath, []string{"removed.txt", "removed_custom.txt"}); wt != nil {
		return wt
	}
	if p := walkUpPTResourceDir(); p != "" {
		return languagetool.LoadManualTaggerFromDirs([]string{p}, []string{"removed.txt", "removed_custom.txt"})
	}
	return nil
}

func walkUpPTResourceDir() string {
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "pt",
		"src", "main", "resources", "org", "languagetool", "resource", "pt")
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for i := 0; i < 14; i++ {
		p := filepath.Join(dir, rel)
		if st, err := os.Stat(p); err == nil && st.IsDir() {
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

func init() {
	// Java PortugueseWordTokenizer() always constructs PortugueseTagger.
	// Register a loader that loads official portuguese.dict on first tokenizer construction
	// (avoids import cycle: tokenizers/pt does not import tagging/pt).
	pttok.RegisterPortugueseTaggerLoader(func() pttok.PortugueseHyphenTagger {
		EnsureDefaultPortugueseTagger()
		return portugueseWordTokenizerTagger()
	})
}
