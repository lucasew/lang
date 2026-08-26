package el

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// elSynthResourceRel is the official LT resource/el directory layout (Java classpath /el/...).
const elSynthResourceRel = "inspiration/languagetool/languagetool-language-modules/el/src/main/resources/org/languagetool/resource/el"

var (
	instanceOnce sync.Once
	instance     *GreekSynthesizer
)

// Instance returns GreekSynthesizer.INSTANCE loaded from official LT resources.
// Nil if greek_synth.dict cannot be opened (fail-closed).
func Instance() *GreekSynthesizer {
	instanceOnce.Do(func() {
		if p := DiscoverGreekSynthDict(); p != "" {
			instance = OpenGreekSynthesizerFromDictPath(p)
			bindINSTANCE(instance)
		}
	})
	return instance
}

// DiscoverGreekSynthDict finds greek_synth.dict (Java RESOURCE_FILENAME /el/greek_synth.dict).
// Order: LANG_EL_SYNTH_DICT / LANG_GREEK_SYNTH_DICT, sibling of LANG_GREEK_DICT,
// inspiration module path (dict ships in language-module resources; no third_party jar needed).
func DiscoverGreekSynthDict() string {
	for _, env := range []string{"LANG_EL_SYNTH_DICT", "LANG_GREEK_SYNTH_DICT"} {
		if p := os.Getenv(env); p != "" {
			if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
				return p
			}
		}
	}
	// Sibling of POS dict when only LANG_GREEK_DICT is set.
	if p := os.Getenv("LANG_GREEK_DICT"); p != "" {
		cand := filepath.Join(filepath.Dir(p), "greek_synth.dict")
		if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
			return cand
		}
	}
	relPaths := []string{
		filepath.Join(elSynthResourceRel, "greek_synth.dict"),
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

// OpenGreekSynthesizerFromDir loads Java GreekSynthesizer resources from a
// resource/el directory containing greek_synth.dict (+ greek_tags.txt, added/removed).
// Returns nil if the binary synth dict cannot be opened (fail-closed).
func OpenGreekSynthesizerFromDir(resourceDir string) *GreekSynthesizer {
	if resourceDir == "" {
		return nil
	}
	base := synthesis.OpenBaseSynthesizerFromDir("el", resourceDir)
	if base == nil {
		return nil
	}
	// Preserve Java RESOURCE_FILENAME / TAGS_FILE_NAME for diagnostics (OpenBase sets basename).
	base.ResourceFileName = ResourceFilename
	base.TagFileName = TagsFileName
	// Java TAGS_FILE_NAME is greek_tags.txt — prefer exact name over glob ambiguity.
	tagsPath := filepath.Join(resourceDir, "greek_tags.txt")
	if tags, err := synthesis.LoadPossibleTagsFile(tagsPath); err == nil && len(tags) > 0 {
		base.PossibleTags = tags
	}
	// Java BaseSynthesizer also loads /el/added.txt + /el/removed.txt from the module tree.
	// Dict ships in inspiration; keep supplement for env/dict-path overrides that only point at the dict.
	supplementManualFromInspiration(base, resourceDir)
	return &GreekSynthesizer{BaseSynthesizer: base}
}

// OpenGreekSynthesizerFromDictPath loads resources from the directory of greek_synth.dict.
func OpenGreekSynthesizerFromDictPath(dictPath string) *GreekSynthesizer {
	if dictPath == "" {
		return nil
	}
	return OpenGreekSynthesizerFromDir(filepath.Dir(dictPath))
}

// supplementManualFromInspiration loads added/removed from inspiration module when missing
// beside the binary dict (Java classpath merges jar + language-module resources).
func supplementManualFromInspiration(base *synthesis.BaseSynthesizer, resourceDir string) {
	if base == nil {
		return
	}
	insp := findInspirationELResourceDir()
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

func findInspirationELResourceDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for i := 0; i < 14; i++ {
		cand := filepath.Join(dir, elSynthResourceRel)
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
