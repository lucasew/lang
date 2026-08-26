package pt

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// ptSynthResourceRel is the official LT resource/pt layout (Java classpath /pt/...).
const ptSynthResourceRel = "inspiration/languagetool/languagetool-language-modules/pt/src/main/resources/org/languagetool/resource/pt"

// Official Maven portuguese-pos-dict extract (same layout as PortugueseTagger).
const ptDictThirdPartySynth = "third_party/portuguese-pos-dict/org/languagetool/resource/pt/portuguese_synth.dict"

var (
	instanceOnce sync.Once
	instance     *PortugueseSynthesizer
)

// Instance returns PortugueseSynthesizer.INSTANCE loaded from official LT resources.
// Nil if portuguese_synth.dict cannot be opened (fail-closed).
func Instance() *PortugueseSynthesizer {
	instanceOnce.Do(func() {
		if p := DiscoverPortugueseSynthDict(); p != "" {
			instance = OpenPortugueseSynthesizerFromDictPath(p)
			bindINSTANCE(instance)
		}
	})
	return instance
}

// DiscoverPortugueseSynthDict finds portuguese_synth.dict (Java RESOURCE_FILENAME /pt/portuguese_synth.dict).
// Order: LANG_PT_SYNTH_DICT / LANG_PORTUGUESE_SYNTH_DICT, sibling of LANG_PORTUGUESE_DICT,
// third_party/portuguese-pos-dict, inspiration module path.
func DiscoverPortugueseSynthDict() string {
	for _, env := range []string{"LANG_PT_SYNTH_DICT", "LANG_PORTUGUESE_SYNTH_DICT"} {
		if p := os.Getenv(env); p != "" {
			if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
				return p
			}
		}
	}
	// Sibling of POS dict when only LANG_PORTUGUESE_DICT is set.
	if p := os.Getenv("LANG_PORTUGUESE_DICT"); p != "" {
		cand := filepath.Join(filepath.Dir(p), "portuguese_synth.dict")
		if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
			return cand
		}
	}
	relPaths := []string{
		ptDictThirdPartySynth,
		filepath.Join(ptSynthResourceRel, "portuguese_synth.dict"),
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

// OpenPortugueseSynthesizerFromDir loads Java PortugueseSynthesizer resources from a
// resource/pt directory containing portuguese_synth.dict (+ portuguese_tags.txt, added/removed, pt.sor).
// Returns nil if the binary synth dict cannot be opened (fail-closed).
func OpenPortugueseSynthesizerFromDir(resourceDir string) *PortugueseSynthesizer {
	if resourceDir == "" {
		return nil
	}
	base := synthesis.OpenBaseSynthesizerFromDir("pt", resourceDir)
	if base == nil {
		return nil
	}
	// Preserve Java RESOURCE_FILENAME / TAGS_FILE_NAME / SOR_FILE_NAME for diagnostics.
	base.ResourceFileName = PortugueseSynthDict
	base.TagFileName = PortugueseTagsFile
	base.SorFileName = PortugueseSorFile
	// Java TAGS_FILE_NAME is portuguese_tags.txt — not portuguese_synth.dict_tags.txt
	// (OpenBase Glob may pick the latter when both exist beside the dict).
	tagsPath := filepath.Join(resourceDir, "portuguese_tags.txt")
	if tags, err := synthesis.LoadPossibleTagsFile(tagsPath); err == nil && len(tags) > 0 {
		base.PossibleTags = tags
	}
	// Ensure number spellers (pt.sor lives in inspiration when dict is under third_party).
	// OpenBase already called LoadNumberSpellersFromDir; re-run after SorFileName is set
	// so walk-up finds /pt/pt.sor under the language module.
	base.LoadNumberSpellersFromDir(resourceDir)
	// Java BaseSynthesizer also loads /pt/added.txt + /pt/removed.txt from the module tree.
	// Dict may live under a jar extract; manual files live in inspiration resource/pt.
	supplementManualFromInspiration(base, resourceDir)
	return &PortugueseSynthesizer{BaseSynthesizer: base}
}

// OpenPortugueseSynthesizerFromDictPath loads resources from the directory of portuguese_synth.dict.
func OpenPortugueseSynthesizerFromDictPath(dictPath string) *PortugueseSynthesizer {
	if dictPath == "" {
		return nil
	}
	return OpenPortugueseSynthesizerFromDir(filepath.Dir(dictPath))
}

// supplementManualFromInspiration loads added/removed from inspiration module when missing
// beside the binary dict (Java classpath merges jar + language-module resources).
func supplementManualFromInspiration(base *synthesis.BaseSynthesizer, resourceDir string) {
	if base == nil {
		return
	}
	insp := findInspirationPTResourceDir()
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

func findInspirationPTResourceDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for i := 0; i < 14; i++ {
		cand := filepath.Join(dir, ptSynthResourceRel)
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
