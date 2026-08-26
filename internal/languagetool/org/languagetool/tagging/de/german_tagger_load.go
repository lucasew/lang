package de

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	desynth "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/de"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	tokde "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/de"
)

var (
	deTaggerOnce  sync.Once
	dePOSDictPath string
	// DefaultGermanTagger is the process-wide GermanTagger backed by german.dict
	// (Java: GermanTagger.INSTANCE / new GermanTagger()). Nil if dict is missing.
	DefaultGermanTagger *GermanTagger
)

// DiscoverGermanPOSDict finds german.dict (Java resource /de/german.dict).
// Order: LANG_GERMAN_DICT, third_party/german-pos-dict, inspiration module path.
func DiscoverGermanPOSDict() string {
	if p := os.Getenv("LANG_GERMAN_DICT"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	relPaths := []string{
		// Prefer inspiration (same tree Java tests load) over older third_party snapshot.
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "de",
			"src", "main", "resources", "org", "languagetool", "resource", "de", "german.dict"),
		filepath.Join("third_party", "german-pos-dict", "org", "languagetool", "resource", "de", "german.dict"),
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

// DiscoverGermanSpellingTxt finds de/hunspell/spelling.txt (Java ExpansionInfos source).
func DiscoverGermanSpellingTxt() string {
	relPaths := []string{
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "de",
			"src", "main", "resources", "org", "languagetool", "resource", "de", "hunspell", "spelling.txt"),
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

// EnsureDefaultGermanTagger loads german.dict + added/removed manuals, wires
// GermanCompoundTokenizer.getStrictInstance(), and ExpansionInfos from
// spelling.txt (adj /A /P + prefix_verb via GermanSynthesizer.INSTANCE).
// Idempotent; no-op if dict missing (fail closed).
func EnsureDefaultGermanTagger() {
	deTaggerOnce.Do(func() {
		p := DiscoverGermanPOSDict()
		if p == "" {
			return
		}
		dePOSDictPath = p
		mt := tagging.OpenMorfologikTagger(p)
		if mt == nil {
			return
		}
		// Java BaseTagger.initWordTagger: CombiningTagger when added.txt present
		wt := tagging.WordTagger(mt)
		if manual := loadDEManualTagger(p); manual != nil {
			removal := loadDEManualRemoval(p)
			wt = tagging.NewCombiningTaggerWithRemoval(wt, manual, removal, false)
		}
		tagger := NewGermanTagger(wt)
		// Java: GermanCompoundTokenizer.getStrictInstance().tokenize
		tagger.SplitCompound = func(word string) []string {
			return tokde.GetStrictInstance().Tokenize(word)
		}
		// Java: expansionInfos = Suppliers.memoize(GermanTagger::initExpansionInfos)
		if sp := DiscoverGermanSpellingTxt(); sp != "" {
			if adj, err := LoadSpellingAdjExpansionFromFile(sp); err == nil && adj != nil {
				tagger.SetSpellingAdjExpansion(adj)
			}
			// Verb expansions: optional synthesizer forms (Java GermanSynthesizer.INSTANCE)
			var synth VerbFormSynth
			if inst := desynth.Instance(); inst != nil {
				synth = func(verbBase string) []string {
					return inst.SynthesizeForPosTags(verbBase, func(pos string) bool {
						return len(pos) >= 4 && pos[:4] == "VER:"
					})
				}
			}
			f, err := os.Open(sp)
			if err == nil {
				verb, err := LoadSpellingVerbExpansionWithSynth(f, synth)
				_ = f.Close()
				if err == nil && verb != nil {
					tagger.SetSpellingVerbExpansion(verb)
				}
			}
		}
		DefaultGermanTagger = tagger
	})
}

// GermanPOSDictPath returns the resolved german.dict path after Ensure (may be empty).
func GermanPOSDictPath() string {
	EnsureDefaultGermanTagger()
	return dePOSDictPath
}

func loadDEManualTagger(dictPath string) tagging.WordTagger {
	if wt := languagetool.LoadManualTaggerBesideDict(dictPath, []string{"added.txt", "added_custom.txt"}); wt != nil {
		return wt
	}
	if p := walkUpDEResourceDir(); p != "" {
		return languagetool.LoadManualTaggerFromDirs([]string{p}, []string{"added.txt", "added_custom.txt"})
	}
	return nil
}

func loadDEManualRemoval(dictPath string) tagging.WordTagger {
	if wt := languagetool.LoadManualTaggerBesideDict(dictPath, []string{"removed.txt", "removed_custom.txt"}); wt != nil {
		return wt
	}
	if p := walkUpDEResourceDir(); p != "" {
		return languagetool.LoadManualTaggerFromDirs([]string{p}, []string{"removed.txt", "removed_custom.txt"})
	}
	return nil
}

func walkUpDEResourceDir() string {
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "de",
		"src", "main", "resources", "org", "languagetool", "resource", "de")
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
