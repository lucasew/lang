package sk

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// skSynthResourceRel is the official LT resource/sk directory layout (Java classpath /sk/...).
const skSynthResourceRel = "inspiration/languagetool/languagetool-language-modules/sk/src/main/resources/org/languagetool/resource/sk"

var (
	instanceOnce sync.Once
	instance     *SlovakSynthesizer
)

// Instance returns SlovakSynthesizer.INSTANCE loaded from official LT resources.
// Nil if slovak_synth.dict cannot be opened (fail-closed).
func Instance() *SlovakSynthesizer {
	instanceOnce.Do(func() {
		if p := DiscoverSlovakSynthDict(); p != "" {
			instance = OpenSlovakSynthesizerFromDictPath(p)
			bindINSTANCE(instance)
		}
	})
	return instance
}

// DiscoverSlovakSynthDict finds slovak_synth.dict (Java RESOURCE_FILENAME /sk/slovak_synth.dict).
// Order: LANG_SK_SYNTH_DICT / LANG_SLOVAK_SYNTH_DICT, sibling of LANG_SLOVAK_DICT,
// inspiration module path (dict ships in language-module resources; no third_party jar needed).
func DiscoverSlovakSynthDict() string {
	for _, env := range []string{"LANG_SK_SYNTH_DICT", "LANG_SLOVAK_SYNTH_DICT"} {
		if p := os.Getenv(env); p != "" {
			if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
				return p
			}
		}
	}
	// Sibling of POS dict when only LANG_SLOVAK_DICT is set.
	if p := os.Getenv("LANG_SLOVAK_DICT"); p != "" {
		cand := filepath.Join(filepath.Dir(p), "slovak_synth.dict")
		if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
			return cand
		}
	}
	relPaths := []string{
		filepath.Join(skSynthResourceRel, "slovak_synth.dict"),
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

// OpenSlovakSynthesizerFromDir loads Java SlovakSynthesizer resources from a
// resource/sk directory containing slovak_synth.dict (+ slovak_tags.txt, added/removed).
// Returns nil if the binary synth dict cannot be opened (fail-closed).
func OpenSlovakSynthesizerFromDir(resourceDir string) *SlovakSynthesizer {
	if resourceDir == "" {
		return nil
	}
	base := synthesis.OpenBaseSynthesizerFromDir("sk", resourceDir)
	if base == nil {
		return nil
	}
	// Preserve Java RESOURCE_FILENAME / TAGS_FILE_NAME for diagnostics (OpenBase sets basename).
	base.ResourceFileName = ResourceFilename
	base.TagFileName = TagsFileName
	// Java TAGS_FILE_NAME is slovak_tags.txt — prefer exact name over glob ambiguity.
	tagsPath := filepath.Join(resourceDir, "slovak_tags.txt")
	if tags, err := synthesis.LoadPossibleTagsFile(tagsPath); err == nil && len(tags) > 0 {
		base.PossibleTags = tags
	}
	// Java BaseSynthesizer also loads /sk/added.txt + /sk/removed.txt from the module tree.
	// Dict ships in inspiration; keep supplement for env/dict-path overrides that only point at the dict.
	supplementManualFromInspiration(base, resourceDir)
	return &SlovakSynthesizer{BaseSynthesizer: base}
}

// OpenSlovakSynthesizerFromDictPath loads resources from the directory of slovak_synth.dict.
func OpenSlovakSynthesizerFromDictPath(dictPath string) *SlovakSynthesizer {
	if dictPath == "" {
		return nil
	}
	return OpenSlovakSynthesizerFromDir(filepath.Dir(dictPath))
}

// supplementManualFromInspiration loads added/removed from inspiration module when missing
// beside the binary dict (Java classpath merges jar + language-module resources).
func supplementManualFromInspiration(base *synthesis.BaseSynthesizer, resourceDir string) {
	if base == nil {
		return
	}
	insp := findInspirationSKResourceDir()
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

func findInspirationSKResourceDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for i := 0; i < 14; i++ {
		cand := filepath.Join(dir, skSynthResourceRel)
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
