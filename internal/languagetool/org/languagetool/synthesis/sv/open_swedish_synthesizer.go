package sv

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// svSynthResourceRel is the official LT resource/sv directory layout (Java classpath /sv/...).
const svSynthResourceRel = "inspiration/languagetool/languagetool-language-modules/sv/src/main/resources/org/languagetool/resource/sv"

var (
	instanceOnce sync.Once
	instance     *SwedishSynthesizer
)

// Instance returns SwedishSynthesizer.INSTANCE loaded from official LT resources.
// Nil if swedish_synth.dict cannot be opened (fail-closed).
func Instance() *SwedishSynthesizer {
	instanceOnce.Do(func() {
		if p := DiscoverSwedishSynthDict(); p != "" {
			instance = OpenSwedishSynthesizerFromDictPath(p)
			bindINSTANCE(instance)
		}
	})
	return instance
}

// DiscoverSwedishSynthDict finds swedish_synth.dict (Java RESOURCE_FILENAME /sv/swedish_synth.dict).
// Order: LANG_SV_SYNTH_DICT / LANG_SWEDISH_SYNTH_DICT, sibling of LANG_SWEDISH_DICT,
// inspiration module path (dict ships in language-module resources; no third_party jar needed).
func DiscoverSwedishSynthDict() string {
	for _, env := range []string{"LANG_SV_SYNTH_DICT", "LANG_SWEDISH_SYNTH_DICT"} {
		if p := os.Getenv(env); p != "" {
			if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
				return p
			}
		}
	}
	// Sibling of POS dict when only LANG_SWEDISH_DICT is set.
	if p := os.Getenv("LANG_SWEDISH_DICT"); p != "" {
		cand := filepath.Join(filepath.Dir(p), "swedish_synth.dict")
		if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
			return cand
		}
	}
	relPaths := []string{
		filepath.Join(svSynthResourceRel, "swedish_synth.dict"),
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

// OpenSwedishSynthesizerFromDir loads Java SwedishSynthesizer resources from a
// resource/sv directory containing swedish_synth.dict (+ swedish_synth.dict_tags.txt,
// added/removed, sv.sor). Returns nil if the binary synth dict cannot be opened (fail-closed).
func OpenSwedishSynthesizerFromDir(resourceDir string) *SwedishSynthesizer {
	if resourceDir == "" {
		return nil
	}
	base := synthesis.OpenBaseSynthesizerFromDir("sv", resourceDir)
	if base == nil {
		return nil
	}
	// Preserve Java RESOURCE_FILENAME / TAGS_FILE_NAME / SOR_FILE_NAME for diagnostics.
	base.ResourceFileName = ResourceFilename
	base.TagFileName = TagsFileName
	base.SorFileName = SorFileName
	// Java TAGS_FILE_NAME is swedish_synth.dict_tags.txt — prefer exact name over glob ambiguity.
	tagsPath := filepath.Join(resourceDir, "swedish_synth.dict_tags.txt")
	if tags, err := synthesis.LoadPossibleTagsFile(tagsPath); err == nil && len(tags) > 0 {
		base.PossibleTags = tags
	}
	// OpenBase sets /sv/sv.sor and LoadNumberSpellersFromDir; re-run after SorFileName
	// is set so walk-up finds sv.sor when dict path differs from inspiration module.
	base.LoadNumberSpellersFromDir(resourceDir)
	// Java BaseSynthesizer also loads /sv/added.txt + /sv/removed.txt from the module tree.
	// Dict ships in inspiration; keep supplement for env/dict-path overrides that only point at the dict.
	supplementManualFromInspiration(base, resourceDir)
	return &SwedishSynthesizer{BaseSynthesizer: base}
}

// OpenSwedishSynthesizerFromDictPath loads resources from the directory of swedish_synth.dict.
func OpenSwedishSynthesizerFromDictPath(dictPath string) *SwedishSynthesizer {
	if dictPath == "" {
		return nil
	}
	return OpenSwedishSynthesizerFromDir(filepath.Dir(dictPath))
}

// supplementManualFromInspiration loads added/removed from inspiration module when missing
// beside the binary dict (Java classpath merges jar + language-module resources).
func supplementManualFromInspiration(base *synthesis.BaseSynthesizer, resourceDir string) {
	if base == nil {
		return
	}
	insp := findInspirationSVResourceDir()
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

func findInspirationSVResourceDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for i := 0; i < 14; i++ {
		cand := filepath.Join(dir, svSynthResourceRel)
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
