package pl

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// plSynthResourceRel is the official LT resource/pl directory layout (Java classpath /pl/...).
const plSynthResourceRel = "inspiration/languagetool/languagetool-language-modules/pl/src/main/resources/org/languagetool/resource/pl"

var (
	instanceOnce sync.Once
	instance     *PolishSynthesizer
)

// Instance returns PolishSynthesizer.INSTANCE loaded from official LT resources.
// Nil if polish_synth.dict cannot be opened (fail-closed).
func Instance() *PolishSynthesizer {
	instanceOnce.Do(func() {
		if p := DiscoverPolishSynthDict(); p != "" {
			instance = OpenPolishSynthesizerFromDictPath(p)
			bindINSTANCE(instance)
		}
	})
	return instance
}

// DiscoverPolishSynthDict finds polish_synth.dict (Java RESOURCE_FILENAME /pl/polish_synth.dict).
// Order: LANG_PL_SYNTH_DICT / LANG_POLISH_SYNTH_DICT, sibling of LANG_POLISH_DICT,
// inspiration module path (dict ships in language-module resources).
func DiscoverPolishSynthDict() string {
	for _, env := range []string{"LANG_PL_SYNTH_DICT", "LANG_POLISH_SYNTH_DICT"} {
		if p := os.Getenv(env); p != "" {
			if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
				return p
			}
		}
	}
	// Sibling of POS dict when only LANG_POLISH_DICT is set.
	if p := os.Getenv("LANG_POLISH_DICT"); p != "" {
		cand := filepath.Join(filepath.Dir(p), "polish_synth.dict")
		if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
			return cand
		}
	}
	relPaths := []string{
		filepath.Join(plSynthResourceRel, "polish_synth.dict"),
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

// OpenPolishSynthesizerFromDir loads Java PolishSynthesizer resources from a
// resource/pl directory containing polish_synth.dict (+ polish_tags.txt, added/removed).
// Returns nil if the binary synth dict cannot be opened (fail-closed).
// Note: Java PolishSynthesizer.getWordForms uses DictionaryLookup only; added/removed
// are loaded by Base but unused by Polish overrides (bug-for-bug).
func OpenPolishSynthesizerFromDir(resourceDir string) *PolishSynthesizer {
	if resourceDir == "" {
		return nil
	}
	base := synthesis.OpenBaseSynthesizerFromDir("pl", resourceDir)
	if base == nil {
		return nil
	}
	// Preserve Java RESOURCE_FILENAME / TAGS_FILE_NAME for diagnostics (OpenBase sets basename).
	base.ResourceFileName = ResourceFilename
	base.TagFileName = TagsFileName
	// Java TAGS_FILE_NAME is polish_tags.txt — OpenBase Glob *tags.txt may also match tagset.txt
	// (skipped) or other *tags files; pin the official file when present.
	tagsPath := filepath.Join(resourceDir, "polish_tags.txt")
	if tags, err := synthesis.LoadPossibleTagsFile(tagsPath); err == nil && len(tags) > 0 {
		base.PossibleTags = tags
	}
	return &PolishSynthesizer{BaseSynthesizer: base}
}

// OpenPolishSynthesizerFromDictPath loads resources from the directory of polish_synth.dict.
// Ports Language.createDefaultSynthesizer() → PolishSynthesizer (with getPosTagCorrection).
func OpenPolishSynthesizerFromDictPath(dictPath string) *PolishSynthesizer {
	if dictPath == "" {
		return nil
	}
	return OpenPolishSynthesizerFromDir(filepath.Dir(dictPath))
}
