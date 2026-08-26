package gl

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

var (
	glTaggerOnce  sync.Once
	glPOSDictPath string
)

// DefaultGalicianTagger is the process singleton (Java new GalicianTagger()).
// Backed by MapWordTagger until EnsureDefaultGalicianTagger loads galician.dict.
var DefaultGalicianTagger = NewGalicianTagger(tagging.MapWordTagger{})

// DiscoverGalicianPOSDict finds galician.dict (Java resource /gl/galician.dict).
// Order: LANG_GALICIAN_DICT, walk-up inspiration module path.
func DiscoverGalicianPOSDict() string {
	if p := os.Getenv("LANG_GALICIAN_DICT"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	relPaths := []string{
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "gl",
			"src", "main", "resources", "org", "languagetool", "resource", "gl", "galician.dict"),
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

// EnsureDefaultGalicianTagger loads galician.dict into DefaultGalicianTagger.
// Idempotent; no-op if dict missing (fail closed).
// Java: overwriteWithManualTagger() → false → CombiningTagger(overwrite=false).
func EnsureDefaultGalicianTagger() {
	glTaggerOnce.Do(func() {
		p := DiscoverGalicianPOSDict()
		if p == "" {
			return
		}
		glPOSDictPath = p
		mt := tagging.OpenMorfologikTagger(p)
		if mt == nil {
			return
		}
		// Java BaseTagger.initWordTagger: CombiningTagger when added.txt present
		wt := tagging.WordTagger(mt)
		if manual := loadGLManualTagger(p); manual != nil {
			removal := loadGLManualRemoval(p)
			wt = tagging.NewCombiningTaggerWithRemoval(wt, manual, removal, false)
		}
		// dictLookup = DictionaryLookup(getDictionary()) for additionalTags
		DefaultGalicianTagger = NewGalicianTaggerWithDictLookup(wt, mt)
	})
}

// GalicianPOSDictPath returns the resolved galician.dict path after Ensure (may be empty).
func GalicianPOSDictPath() string {
	EnsureDefaultGalicianTagger()
	return glPOSDictPath
}

func loadGLManualTagger(dictPath string) tagging.WordTagger {
	if wt := languagetool.LoadManualTaggerBesideDict(dictPath, []string{"added.txt", "added_custom.txt"}); wt != nil {
		return wt
	}
	if p := walkUpGLResourceDir(); p != "" {
		return languagetool.LoadManualTaggerFromDirs([]string{p}, []string{"added.txt", "added_custom.txt"})
	}
	return nil
}

func loadGLManualRemoval(dictPath string) tagging.WordTagger {
	if wt := languagetool.LoadManualTaggerBesideDict(dictPath, []string{"removed.txt", "removed_custom.txt"}); wt != nil {
		return wt
	}
	if p := walkUpGLResourceDir(); p != "" {
		return languagetool.LoadManualTaggerFromDirs([]string{p}, []string{"removed.txt", "removed_custom.txt"})
	}
	return nil
}

func walkUpGLResourceDir() string {
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "gl",
		"src", "main", "resources", "org", "languagetool", "resource", "gl")
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for i := 0; i < 14; i++ {
		p := filepath.Join(dir, rel)
		if st, err := os.Stat(p); err == nil && st.IsDir() {
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
