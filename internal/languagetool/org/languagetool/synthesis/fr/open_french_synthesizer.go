package fr

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// frSynthResourceRel is the official LT resource/fr layout (Java classpath /fr/...).
const frSynthResourceRel = "inspiration/languagetool/languagetool-language-modules/fr/src/main/resources/org/languagetool/resource/fr"

// Official Maven french-pos-dict extract (same layout as FrenchTagger).
const frDictThirdPartySynth = "third_party/french-pos-dict/org/languagetool/resource/fr/french_synth.dict"

var (
	instanceOnce sync.Once
	instance     *FrenchSynthesizer
)

// Instance returns FrenchSynthesizer.INSTANCE loaded from official LT resources.
// Nil if french_synth.dict cannot be opened (fail-closed).
func Instance() *FrenchSynthesizer {
	instanceOnce.Do(func() {
		if p := DiscoverFrenchSynthDict(); p != "" {
			instance = OpenFrenchSynthesizerFromDictPath(p)
			bindINSTANCE(instance)
		}
	})
	return instance
}

// DiscoverFrenchSynthDict finds french_synth.dict (Java RESOURCE_FILENAME /fr/french_synth.dict).
// Order: LANG_FR_SYNTH_DICT / LANG_FRENCH_SYNTH_DICT, sibling of LANG_FRENCH_DICT,
// third_party/french-pos-dict, inspiration module path.
func DiscoverFrenchSynthDict() string {
	for _, env := range []string{"LANG_FR_SYNTH_DICT", "LANG_FRENCH_SYNTH_DICT"} {
		if p := os.Getenv(env); p != "" {
			if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
				return p
			}
		}
	}
	// Sibling of POS dict when only LANG_FRENCH_DICT is set.
	if p := os.Getenv("LANG_FRENCH_DICT"); p != "" {
		cand := filepath.Join(filepath.Dir(p), "french_synth.dict")
		if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
			return cand
		}
	}
	relPaths := []string{
		frDictThirdPartySynth,
		filepath.Join(frSynthResourceRel, "french_synth.dict"),
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

// OpenFrenchSynthesizerFromDir loads Java FrenchSynthesizer resources from a
// resource/fr directory containing french_synth.dict (+ french_tags.txt, added/removed, fr.sor).
// Returns nil if the binary synth dict cannot be opened (fail-closed).
func OpenFrenchSynthesizerFromDir(resourceDir string) *FrenchSynthesizer {
	if resourceDir == "" {
		return nil
	}
	base := synthesis.OpenBaseSynthesizerFromDir("fr", resourceDir)
	if base == nil {
		return nil
	}
	// Preserve Java RESOURCE_FILENAME / TAGS_FILE_NAME / SOR_FILE_NAME for diagnostics.
	base.ResourceFileName = FrenchSynthResource
	base.TagFileName = FrenchTagsFile
	base.SorFileName = FrenchSorFile
	// Java FrenchSynthesizer.isException → Base removeExceptions on synthesize paths.
	base.IsExceptionFn = frenchIsException
	// Java TAGS_FILE_NAME is french_tags.txt — not french_synth.dict_tags.txt
	// (OpenBase Glob may pick the latter when both exist beside the dict).
	tagsPath := filepath.Join(resourceDir, "french_tags.txt")
	if tags, err := synthesis.LoadPossibleTagsFile(tagsPath); err == nil && len(tags) > 0 {
		base.PossibleTags = tags
	}
	// Ensure number spellers (fr.sor lives in inspiration when dict is under third_party).
	// OpenBase already called LoadNumberSpellersFromDir; re-run after SorFileName is set
	// so walk-up finds /fr/fr.sor under the language module.
	base.LoadNumberSpellersFromDir(resourceDir)
	// Java BaseSynthesizer also loads /fr/added.txt + /fr/removed.txt from the module tree.
	// Dict may live under a jar extract; manual files live in inspiration resource/fr.
	supplementManualFromInspiration(base, resourceDir)
	return &FrenchSynthesizer{BaseSynthesizer: base}
}

// OpenFrenchSynthesizerFromDictPath loads resources from the directory of french_synth.dict.
func OpenFrenchSynthesizerFromDictPath(dictPath string) *FrenchSynthesizer {
	if dictPath == "" {
		return nil
	}
	return OpenFrenchSynthesizerFromDir(filepath.Dir(dictPath))
}

// supplementManualFromInspiration loads added/removed from inspiration module when missing
// beside the binary dict (Java classpath merges jar + language-module resources).
func supplementManualFromInspiration(base *synthesis.BaseSynthesizer, resourceDir string) {
	if base == nil {
		return
	}
	insp := findInspirationFRResourceDir()
	dirs := []string{}
	if insp != "" && insp != resourceDir {
		dirs = append(dirs, insp)
	}
	if base.Manual == nil {
		for _, d := range dirs {
			if m, err := synthesis.LoadManualSynthesizerFile(filepath.Join(d, "added.txt")); err == nil && m != nil {
				base.Manual = m
				break
			}
		}
	}
	if base.Removal == nil {
		for _, d := range dirs {
			// Java: removed.txt then do-not-synthesize.txt; OpenBase merges when present in resourceDir.
			if m, err := synthesis.LoadManualSynthesizerFile(filepath.Join(d, "removed.txt")); err == nil && m != nil {
				base.Removal = m
				break
			}
		}
	}
}

func findInspirationFRResourceDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for i := 0; i < 14; i++ {
		cand := filepath.Join(dir, frSynthResourceRel)
		if st, err := os.Stat(cand); err == nil && st.IsDir() {
			return cand
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}
