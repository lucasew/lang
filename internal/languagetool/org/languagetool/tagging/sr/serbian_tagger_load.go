package sr

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

var (
	ekavianOnce      sync.Once
	jekavianOnce     sync.Once
	ekavianPOSPath   string
	jekavianPOSPath  string
)

// DefaultSerbianTagger is the process singleton for Java new SerbianTagger()
// (Ekavian default). Backed by MapWordTagger until EnsureDefaultSerbianTagger.
var DefaultSerbianTagger = NewSerbianTagger(tagging.MapWordTagger{})

// DefaultEkavianTagger is the process singleton for Java new EkavianTagger().
var DefaultEkavianTagger = NewEkavianTagger(tagging.MapWordTagger{})

// DefaultJekavianTagger is the process singleton for Java new JekavianTagger().
var DefaultJekavianTagger = NewJekavianTagger(tagging.MapWordTagger{})

// DiscoverEkavianPOSDict finds ekavian/serbian.dict
// (Java resource /sr/dictionary/ekavian/serbian.dict).
// Order: LANG_SERBIAN_EKAVIAN_DICT, walk-up inspiration module path.
func DiscoverEkavianPOSDict() string {
	if p := os.Getenv("LANG_SERBIAN_EKAVIAN_DICT"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	return walkUpSRFile(filepath.Join("dictionary", "ekavian", "serbian.dict"))
}

// DiscoverJekavianPOSDict finds jekavian/serbian.dict
// (Java resource /sr/dictionary/jekavian/serbian.dict).
// Order: LANG_SERBIAN_JEKAVIAN_DICT, walk-up inspiration module path.
func DiscoverJekavianPOSDict() string {
	if p := os.Getenv("LANG_SERBIAN_JEKAVIAN_DICT"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	return walkUpSRFile(filepath.Join("dictionary", "jekavian", "serbian.dict"))
}

// EnsureDefaultSerbianTagger loads ekavian/serbian.dict into DefaultSerbianTagger
// (Java SerbianTagger default is Ekavian). Idempotent; no-op if dict missing.
func EnsureDefaultSerbianTagger() {
	EnsureDefaultEkavianTagger()
	// SerbianTagger is the same resource path as EkavianTagger.
	if DefaultEkavianTagger != nil && DefaultEkavianTagger.SerbianTagger != nil {
		DefaultSerbianTagger = DefaultEkavianTagger.SerbianTagger
	}
}

// EnsureDefaultEkavianTagger loads ekavian/serbian.dict + ekavian manuals.
// Idempotent; no-op if dict missing (fail closed).
func EnsureDefaultEkavianTagger() {
	ekavianOnce.Do(func() {
		p := DiscoverEkavianPOSDict()
		if p == "" {
			return
		}
		ekavianPOSPath = p
		if tg := loadSerbianTaggerAt(p, EkavianDictionaryPath); tg != nil {
			DefaultEkavianTagger = &EkavianTagger{SerbianTagger: tg}
			DefaultSerbianTagger = tg
		}
	})
}

// EnsureDefaultJekavianTagger loads jekavian/serbian.dict + jekavian manuals.
// Idempotent; no-op if dict missing (fail closed).
func EnsureDefaultJekavianTagger() {
	jekavianOnce.Do(func() {
		p := DiscoverJekavianPOSDict()
		if p == "" {
			return
		}
		jekavianPOSPath = p
		if tg := loadSerbianTaggerAt(p, JekavianDictionaryPath); tg != nil {
			DefaultJekavianTagger = &JekavianTagger{SerbianTagger: tg}
		}
	})
}

// EkavianPOSDictPath returns the resolved ekavian serbian.dict path after Ensure.
func EkavianPOSDictPath() string {
	EnsureDefaultEkavianTagger()
	return ekavianPOSPath
}

// JekavianPOSDictPath returns the resolved jekavian serbian.dict path after Ensure.
func JekavianPOSDictPath() string {
	EnsureDefaultJekavianTagger()
	return jekavianPOSPath
}

// loadSerbianTaggerAt loads a SerbianTagger from a filesystem .dict path with the
// given Java resource dictionary path. Manual added/removed come from beside the
// dict (ekavian/ or jekavian/) matching Java EkavianTagger/JekavianTagger overrides.
func loadSerbianTaggerAt(fsDictPath, resourceDictPath string) *SerbianTagger {
	if fsDictPath == "" {
		return nil
	}
	mt := tagging.OpenMorfologikTagger(fsDictPath)
	if mt == nil {
		return nil
	}
	// Java BaseTagger.initWordTagger: CombiningTagger when added.txt present
	wt := tagging.WordTagger(mt)
	if manual := loadSRManualTagger(fsDictPath); manual != nil {
		removal := loadSRManualRemoval(fsDictPath)
		wt = tagging.NewCombiningTaggerWithRemoval(wt, manual, removal, false)
	}
	return NewSerbianTaggerWithPath(wt, resourceDictPath)
}

func loadSRManualTagger(dictPath string) tagging.WordTagger {
	// Java Ekavian/Jekavian: variant dir added.txt next to serbian.dict.
	// LoadManualTaggerBesideDict finds names beside the dict (ekavian/jekavian).
	if wt := languagetool.LoadManualTaggerBesideDict(dictPath, []string{"added.txt", "added_custom.txt"}); wt != nil {
		return wt
	}
	return nil
}

func loadSRManualRemoval(dictPath string) tagging.WordTagger {
	// Java Ekavian/Jekavian: variant dir removed.txt next to serbian.dict.
	if wt := languagetool.LoadManualTaggerBesideDict(dictPath, []string{"removed.txt", "removed_custom.txt"}); wt != nil {
		return wt
	}
	return nil
}

func walkUpSRFile(relUnderResourceSR string) string {
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "sr",
		"src", "main", "resources", "org", "languagetool", "resource", "sr", relUnderResourceSR)
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
