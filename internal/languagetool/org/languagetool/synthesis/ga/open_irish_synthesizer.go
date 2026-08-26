package ga

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// gaSynthResourceRel is the official LT resource/ga directory layout (Java classpath /ga/...).
const gaSynthResourceRel = "inspiration/languagetool/languagetool-language-modules/ga/src/main/resources/org/languagetool/resource/ga"

// Official Maven languagetool-ga-dicts extract (same fetch tree as IrishTagger irish.dict).
const gaDictFetchSynth = "/tmp/ga-dict-fetch/org/languagetool/resource/ga/irish_synth.dict"

var (
	instanceOnce sync.Once
	instance     *IrishSynthesizer
)

// Instance returns IrishSynthesizer.INSTANCE loaded from official LT resources.
// Nil if irish_synth.dict cannot be opened (fail-closed).
func Instance() *IrishSynthesizer {
	instanceOnce.Do(func() {
		if p := DiscoverIrishSynthDict(); p != "" {
			instance = OpenIrishSynthesizerFromDictPath(p)
			bindINSTANCE(instance)
		}
	})
	return instance
}

// DiscoverIrishSynthDict finds irish_synth.dict (Java RESOURCE_FILENAME /ga/irish_synth.dict).
// Order: LANG_GA_SYNTH_DICT / LANG_IRISH_SYNTH_DICT, sibling of irish.dict (LANG_IRISH_DICT /
// LANG_GA_DICT), /tmp Maven extract, third_party, walk-up inspiration module path.
func DiscoverIrishSynthDict() string {
	for _, env := range []string{"LANG_GA_SYNTH_DICT", "LANG_IRISH_SYNTH_DICT"} {
		if p := os.Getenv(env); p != "" {
			if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
				return p
			}
		}
	}
	// Sibling of POS dict when only LANG_IRISH_DICT / LANG_GA_DICT is set.
	for _, env := range []string{"LANG_IRISH_DICT", "LANG_GA_DICT"} {
		if p := os.Getenv(env); p != "" {
			cand := filepath.Join(filepath.Dir(p), "irish_synth.dict")
			if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
				return cand
			}
		}
	}
	// Pre-extracted Maven languagetool-ga-dicts (official irish_synth.dict bytes).
	if st, err := os.Stat(gaDictFetchSynth); err == nil && st.Mode().IsRegular() {
		return gaDictFetchSynth
	}
	relPaths := []string{
		filepath.Join("third_party", "irish-pos-dict", "org", "languagetool", "resource", "ga", "irish_synth.dict"),
		filepath.Join("third_party", "ga-pos-dict", "org", "languagetool", "resource", "ga", "irish_synth.dict"),
		filepath.Join(gaSynthResourceRel, "irish_synth.dict"),
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

// OpenIrishSynthesizerFromDir loads Java IrishSynthesizer resources from a
// resource/ga directory containing irish_synth.dict (+ irish_tags.txt, added/removed).
// Returns nil if the binary synth dict cannot be opened (fail-closed).
func OpenIrishSynthesizerFromDir(resourceDir string) *IrishSynthesizer {
	if resourceDir == "" {
		return nil
	}
	base := synthesis.OpenBaseSynthesizerFromDir("ga", resourceDir)
	if base == nil {
		return nil
	}
	// Preserve Java RESOURCE_FILENAME / TAGS_FILE_NAME for diagnostics (OpenBase sets basename).
	base.ResourceFileName = ResourceFilename
	base.TagFileName = TagsFileName
	// Java TAGS_FILE_NAME is irish_tags.txt — not irish_synth.dict_tags.txt
	// (OpenBase Glob may pick the latter when both exist beside the dict).
	loadIrishTags(base, resourceDir)
	// Java BaseSynthesizer also loads /ga/added.txt + /ga/removed.txt from the module tree.
	// Dict may live under a Maven extract; manual files live in inspiration resource/ga.
	supplementManualFromInspiration(base, resourceDir)
	return &IrishSynthesizer{BaseSynthesizer: base}
}

// OpenIrishSynthesizerFromDictPath loads resources from the directory of irish_synth.dict.
func OpenIrishSynthesizerFromDictPath(dictPath string) *IrishSynthesizer {
	if dictPath == "" {
		return nil
	}
	return OpenIrishSynthesizerFromDir(filepath.Dir(dictPath))
}

// loadIrishTags prefers Java TAGS_FILE_NAME (irish_tags.txt) from resourceDir, then inspiration.
func loadIrishTags(base *synthesis.BaseSynthesizer, resourceDir string) {
	if base == nil {
		return
	}
	try := func(dir string) bool {
		if dir == "" {
			return false
		}
		tagsPath := filepath.Join(dir, "irish_tags.txt")
		if tags, err := synthesis.LoadPossibleTagsFile(tagsPath); err == nil && len(tags) > 0 {
			base.PossibleTags = tags
			return true
		}
		return false
	}
	if try(resourceDir) {
		return
	}
	// Maven extract may only ship the binary dict; tags ship in language-module resources.
	try(findInspirationGAResourceDir())
}

// supplementManualFromInspiration loads added/removed from inspiration module when missing
// beside the binary dict (Java classpath merges jar + language-module resources).
func supplementManualFromInspiration(base *synthesis.BaseSynthesizer, resourceDir string) {
	if base == nil {
		return
	}
	insp := findInspirationGAResourceDir()
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

func findInspirationGAResourceDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for i := 0; i < 14; i++ {
		cand := filepath.Join(dir, gaSynthResourceRel)
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
