package sr

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// srSynthResourceRel is the official LT resource/sr directory (Java classpath /sr/...).
const srSynthResourceRel = "inspiration/languagetool/languagetool-language-modules/sr/src/main/resources/org/languagetool/resource/sr"

// Nested layout (Java EkavianSynthesizer):
//   dict:  resource/sr/dictionary/ekavian/serbian_synth.dict
//   tags:  resource/sr/dictionary/serbian_synth_tags.txt
//   added/removed (Base langShortCode "sr"): resource/sr/added.txt, removed.txt
// Plain OpenBaseSynthesizerFromDictPath only scans the dict's parent (ekavian/) for
// tags and would miss serbian_synth_tags.txt under dictionary/.

var (
	ekavianInstanceOnce  sync.Once
	ekavianInstance      *EkavianSynthesizer
	jekavianInstanceOnce sync.Once
	jekavianInstance     *JekavianSynthesizer
)

// INSTANCE ports a lazy official Ekavian load (Go convenience; Java has no field).
// Prefer Instance(); INSTANCE is bound on successful load.
var INSTANCE *EkavianSynthesizer

// JEKAVIAN_INSTANCE ports a lazy official Jekavian load (Go convenience; Java has no field).
// Prefer JekavianInstance(); JEKAVIAN_INSTANCE is bound on successful load.
var JEKAVIAN_INSTANCE *JekavianSynthesizer

func bindINSTANCE(s *EkavianSynthesizer) {
	INSTANCE = s
}

func bindJEKAVIAN_INSTANCE(s *JekavianSynthesizer) {
	JEKAVIAN_INSTANCE = s
}

// Instance returns EkavianSynthesizer loaded from official LT resources.
// Nil if serbian_synth.dict cannot be opened (fail-closed).
func Instance() *EkavianSynthesizer {
	ekavianInstanceOnce.Do(func() {
		if p := DiscoverEkavianSynthDict(); p != "" {
			ekavianInstance = OpenEkavianSynthesizerFromDictPath(p)
			bindINSTANCE(ekavianInstance)
		}
	})
	return ekavianInstance
}

// JekavianInstance returns JekavianSynthesizer loaded from official LT resources.
// Nil if jekavian serbian_synth.dict cannot be opened (fail-closed).
func JekavianInstance() *JekavianSynthesizer {
	jekavianInstanceOnce.Do(func() {
		if p := DiscoverJekavianSynthDict(); p != "" {
			jekavianInstance = OpenJekavianSynthesizerFromDictPath(p)
			bindJEKAVIAN_INSTANCE(jekavianInstance)
		}
	})
	return jekavianInstance
}

// DiscoverEkavianSynthDict finds ekavian serbian_synth.dict
// (Java RESOURCE_FILENAME /sr/dictionary/ekavian/serbian_synth.dict).
// Order: LANG_SR_SYNTH_DICT / LANG_EKAVIAN_SYNTH_DICT, sibling of LANG_SR_DICT,
// inspiration module nested path.
func DiscoverEkavianSynthDict() string {
	for _, env := range []string{"LANG_SR_SYNTH_DICT", "LANG_EKAVIAN_SYNTH_DICT"} {
		if p := os.Getenv(env); p != "" {
			if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
				return p
			}
		}
	}
	// Sibling of POS dict when only LANG_SR_DICT is set (often .../ekavian/serbian.dict).
	if p := os.Getenv("LANG_SR_DICT"); p != "" {
		cand := filepath.Join(filepath.Dir(p), "serbian_synth.dict")
		if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
			return cand
		}
	}
	return walkUpSRSynthDict(filepath.Join("dictionary", "ekavian", "serbian_synth.dict"))
}

// DiscoverJekavianSynthDict finds jekavian serbian_synth.dict
// (Java RESOURCE_FILENAME /sr/dictionary/jekavian/serbian_synth.dict).
// Order: LANG_JEKAVIAN_SYNTH_DICT, inspiration module nested path.
func DiscoverJekavianSynthDict() string {
	if p := os.Getenv("LANG_JEKAVIAN_SYNTH_DICT"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	return walkUpSRSynthDict(filepath.Join("dictionary", "jekavian", "serbian_synth.dict"))
}

// walkUpSRSynthDict walks parents of cwd for inspiration/.../resource/sr/<rel>.
func walkUpSRSynthDict(relUnderSR string) string {
	rel := filepath.Join(srSynthResourceRel, relUnderSR)
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for i := 0; i < 14; i++ {
		p := filepath.Join(dir, rel)
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// OpenEkavianSynthesizerFromDictPath loads official Ekavian synthesizer resources.
// dictPath is .../dictionary/ekavian/serbian_synth.dict (fail-closed if missing/unreadable).
func OpenEkavianSynthesizerFromDictPath(dictPath string) *EkavianSynthesizer {
	s := openSerbianSynthesizerFromDictPath(dictPath, EkavianSynthDict)
	if s == nil {
		return nil
	}
	return &EkavianSynthesizer{SerbianSynthesizer: s}
}

// OpenJekavianSynthesizerFromDictPath loads official Jekavian synthesizer resources.
// dictPath is .../dictionary/jekavian/serbian_synth.dict (fail-closed if missing/unreadable).
// Nested tags layout matches Ekavian: tags at dictionary/serbian_synth_tags.txt
// (Java TAGS_FILE_NAME), not beside the dict under jekavian/.
func OpenJekavianSynthesizerFromDictPath(dictPath string) *JekavianSynthesizer {
	s := openSerbianSynthesizerFromDictPath(dictPath, JekavianSynthDict)
	if s == nil {
		return nil
	}
	return &JekavianSynthesizer{SerbianSynthesizer: s}
}

// openSerbianSynthesizerFromDictPath opens morfologik synth + nested tags + /sr/ added/removed.
func openSerbianSynthesizerFromDictPath(dictPath, resourceFileName string) *SerbianSynthesizer {
	if dictPath == "" {
		return nil
	}
	lookup, err := synthesis.OpenMorfologikSynthLookup(dictPath)
	if err != nil || lookup == nil {
		return nil
	}

	// .../dictionary/ekavian/serbian_synth.dict
	ekavianDir := filepath.Dir(dictPath)       // .../ekavian
	dictionaryDir := filepath.Dir(ekavianDir)  // .../dictionary
	resourceDir := filepath.Dir(dictionaryDir) // .../sr  (Java /sr/added.txt root)

	// Prefer resource/sr when layout is the official nested one; fall back to dict dir.
	if filepath.Base(dictionaryDir) != "dictionary" {
		// Unexpected layout: tags may sit next to dict; resource root = parent of dict.
		dictionaryDir = ekavianDir
		resourceDir = ekavianDir
	}

	base := synthesis.NewBaseSynthesizer("sr", nil)
	base.ResourceFileName = resourceFileName
	base.TagFileName = TagsFileName
	base.SorFileName = "/sr/sr.sor"
	base.Lookup = lookup

	// Java TAGS_FILE_NAME = /sr/dictionary/serbian_synth_tags.txt
	tagsPath := filepath.Join(dictionaryDir, "serbian_synth_tags.txt")
	if tags, err := synthesis.LoadPossibleTagsFile(tagsPath); err == nil && len(tags) > 0 {
		base.PossibleTags = tags
	}

	// Java BaseSynthesizer: "/" + langShortCode + "/added.txt" etc. under resource/sr.
	if m, err := synthesis.LoadManualSynthesizerFile(filepath.Join(resourceDir, "added.txt")); err == nil && m != nil {
		base.Manual = m
	}
	base.Removal = loadMergedManualFiles(
		filepath.Join(resourceDir, "removed.txt"),
		filepath.Join(resourceDir, "do-not-synthesize.txt"),
	)

	// Number/Roman spellers: optional; load from resource/sr when present.
	base.LoadNumberSpellersFromDir(resourceDir)

	// If resource/sr has no added/removed (Java classpath also missing), leave nil —
	// do not invent by loading dictionary/ekavian/added.txt (tagger layout, not Base path).
	return &SerbianSynthesizer{BaseSynthesizer: base}
}

func loadMergedManualFiles(paths ...string) *synthesis.ManualSynthesizer {
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

