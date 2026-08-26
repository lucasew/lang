package gl

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// glSynthResourceRel is the official LT resource/gl directory layout (Java classpath /gl/...).
const glSynthResourceRel = "inspiration/languagetool/languagetool-language-modules/gl/src/main/resources/org/languagetool/resource/gl"

var (
	instanceOnce sync.Once
	instance     *GalicianSynthesizer
)

// Instance returns GalicianSynthesizer.INSTANCE loaded from official LT resources.
// Nil if galician_synth.dict cannot be opened (fail-closed).
func Instance() *GalicianSynthesizer {
	instanceOnce.Do(func() {
		if p := DiscoverGalicianSynthDict(); p != "" {
			instance = OpenGalicianSynthesizerFromDictPath(p)
			bindINSTANCE(instance)
		}
	})
	return instance
}

// DiscoverGalicianSynthDict finds galician_synth.dict (Java RESOURCE_FILENAME /gl/galician_synth.dict).
// Order: LANG_GL_SYNTH_DICT / LANG_GALICIAN_SYNTH_DICT, sibling of LANG_GALICIAN_DICT,
// inspiration module path (dict ships in language-module resources; no third_party jar needed).
func DiscoverGalicianSynthDict() string {
	for _, env := range []string{"LANG_GL_SYNTH_DICT", "LANG_GALICIAN_SYNTH_DICT"} {
		if p := os.Getenv(env); p != "" {
			if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
				return p
			}
		}
	}
	// Sibling of POS dict when only LANG_GALICIAN_DICT is set.
	if p := os.Getenv("LANG_GALICIAN_DICT"); p != "" {
		cand := filepath.Join(filepath.Dir(p), "galician_synth.dict")
		if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
			return cand
		}
	}
	relPaths := []string{
		filepath.Join(glSynthResourceRel, "galician_synth.dict"),
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

// OpenGalicianSynthesizerFromDir loads Java GalicianSynthesizer resources from a
// resource/gl directory containing galician_synth.dict (+ galician_tags.txt, added/removed).
// Returns nil if the binary synth dict cannot be opened (fail-closed).
func OpenGalicianSynthesizerFromDir(resourceDir string) *GalicianSynthesizer {
	if resourceDir == "" {
		return nil
	}
	base := synthesis.OpenBaseSynthesizerFromDir("gl", resourceDir)
	if base == nil {
		return nil
	}
	// Preserve Java RESOURCE_FILENAME / TAGS_FILE_NAME for diagnostics (OpenBase sets basename).
	base.ResourceFileName = ResourceFilename
	base.TagFileName = TagsFileName
	// Java TAGS_FILE_NAME is galician_tags.txt — prefer exact name over glob ambiguity.
	tagsPath := filepath.Join(resourceDir, "galician_tags.txt")
	if tags, err := synthesis.LoadPossibleTagsFile(tagsPath); err == nil && len(tags) > 0 {
		base.PossibleTags = tags
	}
	// Java BaseSynthesizer also loads /gl/added.txt + /gl/removed.txt from the module tree.
	// Dict ships in inspiration; keep supplement for env/dict-path overrides that only point at the dict.
	supplementManualFromInspiration(base, resourceDir)
	return &GalicianSynthesizer{BaseSynthesizer: base}
}

// OpenGalicianSynthesizerFromDictPath loads resources from the directory of galician_synth.dict.
func OpenGalicianSynthesizerFromDictPath(dictPath string) *GalicianSynthesizer {
	if dictPath == "" {
		return nil
	}
	return OpenGalicianSynthesizerFromDir(filepath.Dir(dictPath))
}

// supplementManualFromInspiration loads added/removed from inspiration module when missing
// beside the binary dict (Java classpath merges jar + language-module resources).
func supplementManualFromInspiration(base *synthesis.BaseSynthesizer, resourceDir string) {
	if base == nil {
		return
	}
	insp := findInspirationGLResourceDir()
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

func findInspirationGLResourceDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for i := 0; i < 14; i++ {
		cand := filepath.Join(dir, glSynthResourceRel)
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
