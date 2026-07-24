package en

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	entok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/en"
)

var (
	enTaggerOnce  sync.Once
	enTagWordFn   func(token string) []languagetool.TokenTag
	enPOSDictPath string
)

// DiscoverEnglishPOSDict finds english.dict (Java resource /en/english.dict).
// Order: LANG_ENGLISH_DICT, walk-up third_party + inspiration.
func DiscoverEnglishPOSDict() string {
	if p := os.Getenv("LANG_ENGLISH_DICT"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	relPaths := []string{
		filepath.Join("third_party", "english-pos-dict", "org", "languagetool", "resource", "en", "english.dict"),
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "en",
			"src", "main", "resources", "org", "languagetool", "resource", "en", "english.dict"),
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

func init() {
	// Java EnglishWordTokenizer imports EnglishTagger.INSTANCE (always on classpath).
	// Register lazy ensure so wordsToAdd uses real english.dict isTagged, not invent lists.
	entok.SetEnsureEnglishTagger(EnsureDefaultEnglishTagger)
}

// EnsureDefaultEnglishTagger loads english.dict into DefaultEnglishTagger and
// wires EnglishWordTokenizer.IsTaggedEN (Java EnglishTagger.INSTANCE).
// Idempotent; no-op if dict missing (fail closed).
func EnsureDefaultEnglishTagger() {
	enTaggerOnce.Do(func() {
		p := DiscoverEnglishPOSDict()
		if p == "" {
			return
		}
		enPOSDictPath = p
		mt := tagging.OpenMorfologikTagger(p)
		if mt == nil {
			return
		}
		// Java BaseTagger.initWordTagger: CombiningTagger when added.txt present
		wt := tagging.WordTagger(mt)
		if manual := loadENManualTagger(p); manual != nil {
			removal := loadENManualRemoval(p)
			wt = tagging.NewCombiningTaggerWithRemoval(wt, manual, removal, false)
		}
		DefaultEnglishTagger = NewEnglishTagger(wt)
		enTagWordFn = englishTagWordFromTagger(DefaultEnglishTagger)
		// Java EnglishWordTokenizer.wordsToAdd →
		// EnglishTagger.INSTANCE.tag(Arrays.asList(normalized)).get(0).isTagged()
		entok.IsTaggedEN = englishTaggerIsTagged
	})
}

// englishTaggerIsTagged ports EnglishTagger.INSTANCE.tag([s]).get(0).isTagged().
func englishTaggerIsTagged(s string) bool {
	if DefaultEnglishTagger == nil {
		return false
	}
	atrs := DefaultEnglishTagger.Tag([]string{s})
	if len(atrs) == 0 || atrs[0] == nil {
		return false
	}
	return atrs[0].IsTagged()
}

// EnglishTagWord returns the TagWord function for English POS analysis (nil if dict missing).
func EnglishTagWord() func(token string) []languagetool.TokenTag {
	EnsureDefaultEnglishTagger()
	return enTagWordFn
}

// EnglishPOSDictPath returns the resolved english.dict path after Ensure (may be empty).
func EnglishPOSDictPath() string {
	EnsureDefaultEnglishTagger()
	return enPOSDictPath
}

func englishTagWordFromTagger(t *EnglishTagger) func(token string) []languagetool.TokenTag {
	if t == nil {
		return nil
	}
	return func(token string) []languagetool.TokenTag {
		if token == "" {
			return nil
		}
		atrs := t.Tag([]string{token})
		if len(atrs) == 0 || atrs[0] == nil {
			return nil
		}
		var out []languagetool.TokenTag
		for _, r := range atrs[0].GetReadings() {
			if r == nil {
				continue
			}
			pos, lemma := "", ""
			if r.GetPOSTag() != nil {
				pos = *r.GetPOSTag()
			}
			if r.GetLemma() != nil {
				lemma = *r.GetLemma()
			}
			if pos == "" && lemma == "" {
				continue
			}
			if pos == languagetool.SentenceStartTagName || pos == languagetool.SentenceEndTagName {
				continue
			}
			out = append(out, languagetool.TokenTag{POS: pos, Lemma: lemma})
		}
		return out
	}
}

func loadENManualTagger(dictPath string) tagging.WordTagger {
	if wt := languagetool.LoadManualTaggerBesideDict(dictPath, []string{"added.txt", "added_custom.txt"}); wt != nil {
		return wt
	}
	if p := walkUpENResourceDir(); p != "" {
		return languagetool.LoadManualTaggerFromDirs([]string{p}, []string{"added.txt", "added_custom.txt"})
	}
	return nil
}

func loadENManualRemoval(dictPath string) tagging.WordTagger {
	if wt := languagetool.LoadManualTaggerBesideDict(dictPath, []string{"removed.txt", "removed_custom.txt"}); wt != nil {
		return wt
	}
	if p := walkUpENResourceDir(); p != "" {
		return languagetool.LoadManualTaggerFromDirs([]string{p}, []string{"removed.txt", "removed_custom.txt"})
	}
	return nil
}

func walkUpENResourceDir() string {
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "en",
		"src", "main", "resources", "org", "languagetool", "resource", "en")
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
