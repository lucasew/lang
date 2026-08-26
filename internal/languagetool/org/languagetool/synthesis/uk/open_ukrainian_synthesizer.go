package uk

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// ukSynthResourceRel is the official LT resource/uk directory layout (Java classpath /uk/...).
const ukSynthResourceRel = "inspiration/languagetool/languagetool-language-modules/uk/src/main/resources/org/languagetool/resource/uk"

// Official Maven morfologik-ukrainian-lt extract (same as UkrainianTagger fetch path).
const ukDictFetchSynth = "/tmp/uk-dict-fetch/org/languagetool/resource/uk/ukrainian_synth.dict"

var (
	instanceOnce sync.Once
	instance     *UkrainianSynthesizer
)

// Instance returns UkrainianSynthesizer.INSTANCE loaded from official LT resources.
// Nil if ukrainian_synth.dict cannot be opened (fail-closed).
func Instance() *UkrainianSynthesizer {
	instanceOnce.Do(func() {
		if p := DiscoverUkrainianSynthDict(); p != "" {
			instance = OpenUkrainianSynthesizerFromDictPath(p)
			bindINSTANCE(instance)
		}
	})
	return instance
}

// DiscoverUkrainianSynthDict finds ukrainian_synth.dict (Java RESOURCE_FILENAME /uk/ukrainian_synth.dict).
// Order: LANG_UK_SYNTH_DICT, LANG_UKRAINIAN_DICT sibling, walk-up inspiration path, /tmp extract.
func DiscoverUkrainianSynthDict() string {
	if p := os.Getenv("LANG_UK_SYNTH_DICT"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	// Sibling of POS dict when only LANG_UKRAINIAN_DICT is set.
	if p := os.Getenv("LANG_UKRAINIAN_DICT"); p != "" {
		cand := filepath.Join(filepath.Dir(p), "ukrainian_synth.dict")
		if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
			return cand
		}
	}
	candidates := []string{
		filepath.Join(ukSynthResourceRel, "ukrainian_synth.dict"),
		ukDictFetchSynth,
	}
	wd, err := os.Getwd()
	if err != nil {
		if st, err := os.Stat(ukDictFetchSynth); err == nil && st.Mode().IsRegular() {
			return ukDictFetchSynth
		}
		return ""
	}
	dir := wd
	for i := 0; i < 14; i++ {
		for _, rel := range candidates {
			if filepath.IsAbs(rel) {
				if st, err := os.Stat(rel); err == nil && st.Mode().IsRegular() {
					return rel
				}
				continue
			}
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

// OpenUkrainianSynthesizerFromDir loads Java UkrainianSynthesizer resources from a
// resource/uk directory containing ukrainian_synth.dict (+ ukrainian_tags.txt, added/removed).
// Returns nil if the binary synth dict cannot be opened (fail-closed).
func OpenUkrainianSynthesizerFromDir(resourceDir string) *UkrainianSynthesizer {
	if resourceDir == "" {
		return nil
	}
	base := synthesis.OpenBaseSynthesizerFromDir("uk", resourceDir)
	if base == nil {
		return nil
	}
	// Preserve Java RESOURCE_FILENAME / TAGS_FILE_NAME for diagnostics (OpenBase sets basename).
	base.ResourceFileName = ResourceFilename
	base.TagFileName = TagsFileName
	// Java TAGS_FILE_NAME is ukrainian_tags.txt — not ukrainian_synth.dict_tags.txt
	// (OpenBase Glob may pick the latter when both exist beside the dict).
	tagsPath := filepath.Join(resourceDir, "ukrainian_tags.txt")
	if tags, err := synthesis.LoadPossibleTagsFile(tagsPath); err == nil && len(tags) > 0 {
		base.PossibleTags = tags
	}
	// Java BaseSynthesizer also loads /uk/added.txt + /uk/removed.txt from the module tree.
	// Dict may live under a jar extract; manual files live in inspiration resource/uk.
	supplementManualFromInspiration(base, resourceDir)
	return &UkrainianSynthesizer{BaseSynthesizer: base}
}

// OpenUkrainianSynthesizerFromDictPath loads resources from the directory of ukrainian_synth.dict.
func OpenUkrainianSynthesizerFromDictPath(dictPath string) *UkrainianSynthesizer {
	if dictPath == "" {
		return nil
	}
	return OpenUkrainianSynthesizerFromDir(filepath.Dir(dictPath))
}

// supplementManualFromInspiration loads added/removed from inspiration module when missing
// beside the binary dict (Java classpath merges jar + language-module resources).
func supplementManualFromInspiration(base *synthesis.BaseSynthesizer, resourceDir string) {
	if base == nil {
		return
	}
	insp := findInspirationUKResourceDir()
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

func findInspirationUKResourceDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for i := 0; i < 14; i++ {
		cand := filepath.Join(dir, ukSynthResourceRel)
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
