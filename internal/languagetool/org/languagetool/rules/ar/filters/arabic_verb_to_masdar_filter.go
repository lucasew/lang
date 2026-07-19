package filters

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// Official /ar/arabic_verb_masdar.txt (Java ArabicVerbToMafoulMutlaqFilter.FILE_NAME).
const arabicVerbMasdarRel = "inspiration/languagetool/languagetool-language-modules/ar/src/main/resources/org/languagetool/rules/ar/arabic_verb_masdar.txt"

var (
	verbMasdarOnce sync.Once
	verbMasdarData map[string][]string
)

func loadOfficialVerbMasdarMap() map[string][]string {
	verbMasdarOnce.Do(func() {
		verbMasdarData = map[string][]string{}
		path := discoverArabicVerbMasdar()
		if path == "" {
			return
		}
		f, err := os.Open(path)
		if err != nil {
			return
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil || m == nil {
			return
		}
		// Index by full form and tashkeel-stripped for lookup helpers.
		out := map[string][]string{}
		for verb, masdars := range m {
			out[verb] = append([]string(nil), masdars...)
			plain := tools.RemoveTashkeel(verb)
			if plain != verb {
				out[plain] = append(out[plain], masdars...)
			}
		}
		verbMasdarData = out
	})
	out := make(map[string][]string, len(verbMasdarData))
	for k, v := range verbMasdarData {
		out[k] = append([]string(nil), v...)
	}
	return out
}

func discoverArabicVerbMasdar() string {
	_, file, _, ok := runtime.Caller(0)
	if ok {
		root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../../../"))
		p := filepath.Join(root, arabicVerbMasdarRel)
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for {
		cand := filepath.Join(dir, arabicVerbMasdarRel)
		if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
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

// ArabicVerbToMasdarFilter ports verb→masdar lookup from official arabic_verb_masdar.txt
// (not invent reverse of a soft invent masdar map).
type ArabicVerbToMasdarFilter struct {
	Verb2Masdar map[string][]string
}

func NewArabicVerbToMasdarFilter() *ArabicVerbToMasdarFilter {
	return &ArabicVerbToMasdarFilter{Verb2Masdar: loadOfficialVerbMasdarMap()}
}

// SuggestMasdarsForVerb returns masdar lemmas for a verb lemma.
func (f *ArabicVerbToMasdarFilter) SuggestMasdarsForVerb(verbLemma string) []string {
	if f == nil {
		return nil
	}
	if v, ok := f.Verb2Masdar[verbLemma]; ok {
		return append([]string{}, v...)
	}
	if v, ok := f.Verb2Masdar[tools.RemoveTashkeel(verbLemma)]; ok {
		return append([]string{}, v...)
	}
	return nil
}
