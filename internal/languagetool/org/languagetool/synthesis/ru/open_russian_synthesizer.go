package ru

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// ruSynthResourceRel is the official LT resource/ru directory layout (Java classpath /ru/...).
const ruSynthResourceRel = "inspiration/languagetool/languagetool-language-modules/ru/src/main/resources/org/languagetool/resource/ru"

var (
	instanceOnce sync.Once
	instance     *RussianSynthesizer
)

// Instance returns RussianSynthesizer.INSTANCE loaded from official LT resources.
// Nil if russian_synth.dict cannot be opened (fail-closed).
func Instance() *RussianSynthesizer {
	instanceOnce.Do(func() {
		if p := DiscoverRussianSynthDict(); p != "" {
			instance = OpenRussianSynthesizerFromDictPath(p)
			bindINSTANCE(instance)
		}
	})
	return instance
}

// DiscoverRussianSynthDict finds russian_synth.dict (Java RESOURCE_FILENAME /ru/russian_synth.dict).
// Order: LANG_RU_SYNTH_DICT / LANG_RUSSIAN_SYNTH_DICT, sibling of LANG_RUSSIAN_DICT,
// inspiration module path (dict ships in language-module resources; no third_party jar needed).
func DiscoverRussianSynthDict() string {
	for _, env := range []string{"LANG_RU_SYNTH_DICT", "LANG_RUSSIAN_SYNTH_DICT"} {
		if p := os.Getenv(env); p != "" {
			if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
				return p
			}
		}
	}
	// Sibling of POS dict when only LANG_RUSSIAN_DICT is set.
	if p := os.Getenv("LANG_RUSSIAN_DICT"); p != "" {
		cand := filepath.Join(filepath.Dir(p), "russian_synth.dict")
		if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
			return cand
		}
	}
	relPaths := []string{
		filepath.Join(ruSynthResourceRel, "russian_synth.dict"),
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

// OpenRussianSynthesizerFromDir loads Java RussianSynthesizer resources from a
// resource/ru directory containing russian_synth.dict (+ tags_russian.txt, added/removed).
// Returns nil if the binary synth dict cannot be opened (fail-closed).
func OpenRussianSynthesizerFromDir(resourceDir string) *RussianSynthesizer {
	if resourceDir == "" {
		return nil
	}
	base := synthesis.OpenBaseSynthesizerFromDir("ru", resourceDir)
	if base == nil {
		return nil
	}
	// Preserve Java RESOURCE_FILENAME / TAGS_FILE_NAME for diagnostics (OpenBase sets basename).
	base.ResourceFileName = ResourceFilename
	base.TagFileName = TagsFileName
	// Java TAGS_FILE_NAME is tags_russian.txt — OpenBase Glob looks for *tags.txt / *_tags.txt
	// and will not match tags_russian.txt (suffix is _russian.txt).
	tagsPath := filepath.Join(resourceDir, "tags_russian.txt")
	if tags, err := synthesis.LoadPossibleTagsFile(tagsPath); err == nil && len(tags) > 0 {
		base.PossibleTags = tags
	}
	// Java BaseSynthesizer also loads /ru/added.txt + /ru/removed.txt from the module tree.
	// Dict ships in inspiration; keep supplement for env/dict-path overrides that only point at the dict.
	supplementManualFromInspiration(base, resourceDir)
	return &RussianSynthesizer{BaseSynthesizer: base}
}

// OpenRussianSynthesizerFromDictPath loads resources from the directory of russian_synth.dict.
func OpenRussianSynthesizerFromDictPath(dictPath string) *RussianSynthesizer {
	if dictPath == "" {
		return nil
	}
	return OpenRussianSynthesizerFromDir(filepath.Dir(dictPath))
}

// supplementManualFromInspiration loads added/removed from inspiration module when missing
// beside the binary dict (Java classpath merges jar + language-module resources).
func supplementManualFromInspiration(base *synthesis.BaseSynthesizer, resourceDir string) {
	if base == nil {
		return
	}
	insp := findInspirationRUResourceDir()
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

func findInspirationRUResourceDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for i := 0; i < 14; i++ {
		cand := filepath.Join(dir, ruSynthResourceRel)
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
