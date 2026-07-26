package en

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// enSynthResourceRel is the official LT resource/en layout (Java classpath /en/...).
const enSynthResourceRel = "inspiration/languagetool/languagetool-language-modules/en/src/main/resources/org/languagetool/resource/en"

// Official Maven english-pos-dict extract (same layout as EnglishTagger).
const enDictThirdPartySynth = "third_party/english-pos-dict/org/languagetool/resource/en/english_synth.dict"

var (
	instanceOnce sync.Once
	instance     *EnglishSynthesizer
)

// Instance returns EnglishSynthesizer.INSTANCE loaded from official LT resources.
// Nil if english_synth.dict cannot be opened (fail-closed).
func Instance() *EnglishSynthesizer {
	instanceOnce.Do(func() {
		if p := DiscoverEnglishSynthDict(); p != "" {
			instance = OpenEnglishSynthesizerFromDictPath(p)
			bindINSTANCE(instance)
		}
	})
	return instance
}

// DiscoverEnglishSynthDict finds english_synth.dict (Java RESOURCE_FILENAME /en/english_synth.dict).
// Order: LANG_EN_SYNTH_DICT / LANG_ENGLISH_SYNTH_DICT, sibling of LANG_ENGLISH_DICT,
// third_party/english-pos-dict, inspiration module path.
func DiscoverEnglishSynthDict() string {
	for _, env := range []string{"LANG_EN_SYNTH_DICT", "LANG_ENGLISH_SYNTH_DICT"} {
		if p := os.Getenv(env); p != "" {
			if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
				return p
			}
		}
	}
	// Sibling of POS dict when only LANG_ENGLISH_DICT is set.
	if p := os.Getenv("LANG_ENGLISH_DICT"); p != "" {
		cand := filepath.Join(filepath.Dir(p), "english_synth.dict")
		if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
			return cand
		}
	}
	relPaths := []string{
		enDictThirdPartySynth,
		filepath.Join(enSynthResourceRel, "english_synth.dict"),
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

// OpenEnglishSynthesizerFromDir loads Java EnglishSynthesizer resources from a
// resource/en directory containing english_synth.dict (+ english_tags.txt, added/removed,
// do-not-synthesize.txt, en.sor). Returns nil if the binary synth dict cannot be opened.
func OpenEnglishSynthesizerFromDir(resourceDir string) *EnglishSynthesizer {
	if resourceDir == "" {
		return nil
	}
	base := synthesis.OpenBaseSynthesizerFromDir("en", resourceDir)
	if base == nil {
		return nil
	}
	// Preserve Java RESOURCE_FILENAME / TAGS_FILE_NAME / SOR_FILE_NAME for diagnostics.
	base.ResourceFileName = EnglishSynthResource
	base.TagFileName = EnglishTagsFile
	base.SorFileName = EnglishSorFile
	// Java EnglishSynthesizer.isException → Base removeExceptions on synthesize paths.
	base.IsExceptionFn = englishIsException
	// Java TAGS_FILE_NAME is english_tags.txt — not english_synth.dict_tags.txt.
	tagsPath := filepath.Join(resourceDir, "english_tags.txt")
	if tags, err := synthesis.LoadPossibleTagsFile(tagsPath); err == nil && len(tags) > 0 {
		base.PossibleTags = tags
	}
	// Ensure number spellers (en.sor lives in inspiration when dict is under third_party).
	base.LoadNumberSpellersFromDir(resourceDir)
	// Java BaseSynthesizer also loads /en/added.txt + /en/removed.txt + /en/do-not-synthesize.txt
	// from the module tree. Dict may live under a jar extract; manuals live in inspiration resource/en.
	supplementManualFromInspiration(base, resourceDir)
	return &EnglishSynthesizer{BaseSynthesizer: base}
}

// OpenEnglishSynthesizerFromDictPath loads resources from the directory of english_synth.dict.
func OpenEnglishSynthesizerFromDictPath(dictPath string) *EnglishSynthesizer {
	if dictPath == "" {
		return nil
	}
	return OpenEnglishSynthesizerFromDir(filepath.Dir(dictPath))
}

// supplementManualFromInspiration loads added/removed/do-not-synthesize from inspiration module
// when missing beside the binary dict (Java classpath merges jar + language-module resources).
func supplementManualFromInspiration(base *synthesis.BaseSynthesizer, resourceDir string) {
	if base == nil {
		return
	}
	insp := findInspirationENResourceDir()
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
			// Java: removed.txt then do-not-synthesize.txt (two ManualSynthesizer streams).
			// OpenBase merges when both present in resourceDir; here merge from inspiration.
			m := loadMergedManual(
				filepath.Join(d, "removed.txt"),
				filepath.Join(d, "do-not-synthesize.txt"),
			)
			if m != nil {
				base.Removal = m
				break
			}
		}
	}
}

func findInspirationENResourceDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for i := 0; i < 14; i++ {
		cand := filepath.Join(dir, enSynthResourceRel)
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

// loadMergedManual concatenates ManualSynthesizer files (Java dual removal streams).
func loadMergedManual(paths ...string) *synthesis.ManualSynthesizer {
	var parts []byte
	any := false
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil || len(data) == 0 {
			continue
		}
		any = true
		parts = append(parts, data...)
		if data[len(data)-1] != '\n' {
			parts = append(parts, '\n')
		}
	}
	if !any {
		return nil
	}
	m, err := synthesis.NewManualSynthesizer(strings.NewReader(string(parts)))
	if err != nil {
		return nil
	}
	return m
}
