package filters

import (
	"embed"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// Official /ar/*.txt under rules/ar (same bytes as Java rules resources).
//
//go:embed data/arabic_adjective_exclamation.txt
//go:embed data/arabic_masdar_verb.txt
//go:embed data/arabic_verb_masdar.txt
var arRulesDataFS embed.FS

// Paths under inspiration (Java /ar/*.txt via rules dir) — used for other filter maps
// that are not yet embed-hardened in this sector.
const (
	relARRules = "inspiration/languagetool/languagetool-language-modules/ar/src/main/resources/org/languagetool/rules/ar"
)

func discoverARRulesFile(name string) string {
	rel := filepath.Join(relARRules, name)
	_, file, _, ok := runtime.Caller(0)
	if ok {
		// filters → ar → rules → languagetool → org → languagetool → internal → root (7)
		root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../../../"))
		p := filepath.Join(root, rel)
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
		cand := filepath.Join(dir, rel)
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

func loadARSimpleReplaceMap(name string) map[string][]string {
	path := discoverARRulesFile(name)
	if path == "" {
		return map[string][]string{}
	}
	f, err := os.Open(path)
	if err != nil {
		return map[string][]string{}
	}
	defer f.Close()
	m, err := rules.LoadSimpleReplaceWords(f)
	if err != nil || m == nil {
		return map[string][]string{}
	}
	out := make(map[string][]string, len(m))
	for k, v := range m {
		// Official AR files may pad keys with spaces (e.g. " جميل=…").
		k = tools.JavaStringTrim(k)
		if k == "" {
			continue
		}
		reps := make([]string, 0, len(v))
		for _, r := range v {
			r = tools.JavaStringTrim(r)
			if r != "" {
				reps = append(reps, r)
			}
		}
		if len(reps) == 0 {
			continue
		}
		out[k] = reps
	}
	return out
}

// loadOfficialMasdarVerbMap ports ArabicMasdarToVerbFilter.loadFromPath(FILE_NAME)
// → SimpleReplaceDataLoader.loadWords("/ar/arabic_masdar_verb.txt").
// Embedded official bytes only — no invent trim of keys (Java loadWords keeps them).
var (
	masdarMapOnce sync.Once
	masdarMapData map[string][]string
)

func loadOfficialMasdarVerbMap() map[string][]string {
	masdarMapOnce.Do(func() {
		f, err := arRulesDataFS.Open("data/arabic_masdar_verb.txt")
		if err != nil {
			// Java getFromRulesDirAsStream fails if resource missing.
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			// Java SimpleReplaceDataLoader throws RuntimeException on bad lines.
			panic(err)
		}
		if m == nil {
			m = map[string][]string{}
		}
		masdarMapData = m
	})
	return copyStringListMap(masdarMapData)
}

// loadOfficialVerbMasdarMap ports ArabicVerbToMafoulMutlaqFilter.loadFromPath(FILE_NAME)
// → SimpleReplaceDataLoader.loadWords("/ar/arabic_verb_masdar.txt").
// Embedded official bytes only — exact keys (no tashkeel-stripped dual map invent).
var (
	verbMasdarOnce sync.Once
	verbMasdarData map[string][]string
)

func loadOfficialVerbMasdarMap() map[string][]string {
	verbMasdarOnce.Do(func() {
		f, err := arRulesDataFS.Open("data/arabic_verb_masdar.txt")
		if err != nil {
			// Java getFromRulesDirAsStream fails if resource missing.
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			// Java SimpleReplaceDataLoader throws RuntimeException on bad lines.
			panic(err)
		}
		if m == nil {
			m = map[string][]string{}
		}
		verbMasdarData = m
	})
	return copyStringListMap(verbMasdarData)
}

// loadOfficialAdjExclamationMap ports ArabicAdjectiveToExclamationFilter.loadFromPath(FILE_NAME)
// → SimpleReplaceDataLoader.loadWords("/ar/arabic_adjective_exclamation.txt").
// Embedded official bytes only — no invent strip/trim of keys (Java loadWords keeps them).
var (
	adjExclOnce sync.Once
	adjExclData map[string][]string
)

func loadOfficialAdjExclamationMap() map[string][]string {
	adjExclOnce.Do(func() {
		f, err := arRulesDataFS.Open("data/arabic_adjective_exclamation.txt")
		if err != nil {
			// Java getFromRulesDirAsStream fails if resource missing.
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			// Java SimpleReplaceDataLoader throws RuntimeException on bad lines.
			panic(err)
		}
		if m == nil {
			m = map[string][]string{}
		}
		adjExclData = m
	})
	return copyStringListMap(adjExclData)
}

func copyStringListMap(m map[string][]string) map[string][]string {
	if m == nil {
		return map[string][]string{}
	}
	out := make(map[string][]string, len(m))
	for k, v := range m {
		out[k] = append([]string(nil), v...)
	}
	return out
}
