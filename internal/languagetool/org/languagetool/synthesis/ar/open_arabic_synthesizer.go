package ar

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	artag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ar"
)

// arSynthResourceRel is the official LT resource/ar directory layout (Java classpath /ar/...).
const arSynthResourceRel = "inspiration/languagetool/languagetool-language-modules/ar/src/main/resources/org/languagetool/resource/ar"

var (
	instanceOnce sync.Once
	instance     *ArabicSynthesizer
)

// Instance returns ArabicSynthesizer.INSTANCE loaded from official LT resources.
// Nil if arabic_synth.dict cannot be opened (fail-closed).
func Instance() *ArabicSynthesizer {
	instanceOnce.Do(func() {
		if p := DiscoverArabicSynthDict(); p != "" {
			instance = OpenArabicSynthesizerFromDictPath(p)
			bindINSTANCE(instance)
		}
	})
	return instance
}

// DiscoverArabicSynthDict finds arabic_synth.dict (Java RESOURCE_FILENAME /ar/arabic_synth.dict).
// Order: LANG_AR_SYNTH_DICT / LANG_ARABIC_SYNTH_DICT, sibling of LANG_ARABIC_DICT,
// inspiration module path (dict ships in language-module resources).
func DiscoverArabicSynthDict() string {
	for _, env := range []string{"LANG_AR_SYNTH_DICT", "LANG_ARABIC_SYNTH_DICT"} {
		if p := os.Getenv(env); p != "" {
			if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
				return p
			}
		}
	}
	// Sibling of POS dict when only LANG_ARABIC_DICT is set.
	if p := os.Getenv("LANG_ARABIC_DICT"); p != "" {
		cand := filepath.Join(filepath.Dir(p), "arabic_synth.dict")
		if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
			return cand
		}
	}
	relPaths := []string{
		filepath.Join(arSynthResourceRel, "arabic_synth.dict"),
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

// OpenArabicSynthesizerFromDir loads Java ArabicSynthesizer resources from a
// resource directory containing arabic_synth.dict (+ tags/manual when present).
// Returns nil if the binary synth dict cannot be opened (fail-closed).
func OpenArabicSynthesizerFromDir(resourceDir string) *ArabicSynthesizer {
	if resourceDir == "" {
		return nil
	}
	base := synthesis.OpenBaseSynthesizerFromDir("ar", resourceDir)
	if base == nil {
		return nil
	}
	// Preserve Java RESOURCE_FILENAME / TAGS_FILE_NAME for diagnostics (OpenBase sets basename).
	base.ResourceFileName = ArabicSynthDict
	base.TagFileName = ArabicTagsFile
	// Java TAGS_FILE_NAME is /ar/arabic_tags.txt — prefer exact name over
	// arabic_synth.dict_tags.txt (glob in OpenBase can pick the dict_tags twin first).
	tagsPath := filepath.Join(resourceDir, "arabic_tags.txt")
	if tags, err := synthesis.LoadPossibleTagsFile(tagsPath); err == nil && len(tags) > 0 {
		base.PossibleTags = tags
	}
	// Java BaseSynthesizer also loads /ar/added.txt + /ar/removed.txt from the module tree.
	supplementManualFromInspiration(base, resourceDir)
	// Java: private final ArabicTagger tagger = new ArabicTagger();
	// Prefer dict-backed DefaultArabicTagger when available (same dict path as Java).
	artag.EnsureDefaultArabicTagger()
	tg := artag.DefaultArabicTagger
	if tg == nil {
		tg = artag.NewArabicTagger(nil)
	}
	return &ArabicSynthesizer{
		BaseSynthesizer: base,
		tagmanager:      nil, // lazy via tm()
		tagger:          tg,
	}
}

// OpenArabicSynthesizerFromDictPath loads resources from the directory of arabic_synth.dict.
// Ports Language.createDefaultSynthesizer() → ArabicSynthesizer (getPosTagCorrection).
func OpenArabicSynthesizerFromDictPath(dictPath string) *ArabicSynthesizer {
	if dictPath == "" {
		return nil
	}
	return OpenArabicSynthesizerFromDir(filepath.Dir(dictPath))
}

// supplementManualFromInspiration loads added/removed from inspiration module when missing
// beside the binary dict (Java classpath merges jar + language-module resources).
func supplementManualFromInspiration(base *synthesis.BaseSynthesizer, resourceDir string) {
	if base == nil {
		return
	}
	insp := findInspirationARResourceDir()
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
			if m, err := synthesis.LoadManualSynthesizerFile(filepath.Join(d, "removed.txt")); err == nil && m != nil {
				base.Removal = m
				break
			}
		}
	}
}

func findInspirationARResourceDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for i := 0; i < 14; i++ {
		cand := filepath.Join(dir, arSynthResourceRel)
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
