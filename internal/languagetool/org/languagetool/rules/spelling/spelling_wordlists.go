package spelling

import (
	"bufio"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"os"
	"path/filepath"
	"strings"
)

// DiscoverLangHunspellWordList finds language shortCode + "/hunspell/" + name
// (Java SpellingCheckRule.getIgnoreFileName: "da/hunspell/ignore.txt").
// Fallbacks (language overrides of getIgnoreFileName / getSpellingFileName):
//   - shortCode + "/spelling/" + name  (NL: "/nl/spelling/ignore.txt")
//   - shortCode + "/" + name           (PT: "pt/ignore.txt", "pt/spelling.txt")
func DiscoverLangHunspellWordList(shortCode, name string) string {
	if shortCode == "" || name == "" {
		return ""
	}
	if p := discoverResourceRel(shortCode + "/hunspell/" + name); p != "" {
		return p
	}
	// NL (and similar) place ignore/spelling/prohibit under resource/{lang}/spelling/.
	if p := discoverResourceRel(shortCode + "/spelling/" + name); p != "" {
		return p
	}
	// PT MorfologikPortugueseSpellerRule: "pt/ignore.txt", "pt/prohibit.txt".
	return discoverResourceRel(shortCode + "/" + name)
}

// DiscoverSpellingGlobal finds official spelling_global.txt
// (Java SpellingCheckRule.GLOBAL_SPELLING_FILE / getAdditionalSpellingFileNames).
// Path: org/languagetool/resource/spelling_global.txt (core resources, all languages).
func DiscoverSpellingGlobal() string {
	return discoverResourceRel(GlobalSpellingFile)
}

// DiscoverSpellingResource finds a resource-dir relative path
// (e.g. "en/hunspell/spelling_en-US.txt", Java getLanguageVariantSpellingFileName).
func DiscoverSpellingResource(rel string) string {
	return discoverResourceRel(rel)
}

// LanguageVariantSpellingClasspath ports getLanguageVariantSpellingFileName for
// well-known Java overrides (EN locales + de-AT/de-CH). Empty when Java returns null
// (base SpellingCheckRule.SPELLING_FILE_VARIANT).
func LanguageVariantSpellingClasspath(langCode string) string {
	c := strings.ToLower(tools.JavaStringTrim(langCode))
	if c == "" {
		return ""
	}
	// English Morfologik*SpellerRule LANGUAGE_SPECIFIC_PLAIN_TEXT_DICT
	if strings.HasPrefix(c, "en") {
		switch {
		case strings.Contains(c, "gb"):
			return "en/hunspell/spelling_en-GB.txt"
		case strings.Contains(c, "-ca") || strings.HasSuffix(c, "_ca"):
			return "en/hunspell/spelling_en-CA.txt"
		case strings.Contains(c, "au"):
			return "en/hunspell/spelling_en-AU.txt"
		case strings.Contains(c, "nz"):
			return "en/hunspell/spelling_en-NZ.txt"
		case strings.Contains(c, "za"):
			return "en/hunspell/spelling_en-ZA.txt"
		default:
			// bare "en" and en-US → American (Java MorfologikAmericanSpellerRule)
			return "en/hunspell/spelling_en-US.txt"
		}
	}
	// German Austrian / Swiss only (Java AustrianGermanSpellerRule / SwissGermanSpellerRule)
	if strings.HasPrefix(c, "de") {
		switch {
		case strings.Contains(c, "at"):
			return "de/hunspell/spelling-de-AT.txt"
		case strings.Contains(c, "ch"):
			return "de/hunspell/spelling-de-CH.txt"
		}
	}
	return ""
}

// ReapplyDefaultSpellingWordLists clears ignore/prohibit/multi-word ignore state
// then runs ApplyDefaultSpellingWordLists. Call after flipping DisableTokenizeNewWords
// so spelling lists reload under the correct tokenizeNewWords mode.
func ReapplyDefaultSpellingWordLists(r *SpellingCheckRule) {
	if r == nil {
		return
	}
	r.IgnoreWords = map[string]struct{}{}
	r.ProhibitedWords = map[string]struct{}{}
	r.MultiWordIgnore = nil
	r.AntiPatterns = nil
	r.ignoreDictSorted = nil
	r.ignoreDictSortedFold = nil
	ApplyDefaultSpellingWordLists(r)
}

// ApplySpellingResourcePaths ports language-specific getIgnoreFileName /
// getSpellingFileName / getProhibitFileName (absolute Java resource paths).
// Empty paths are skipped; missing files fail-closed (no invent).
func ApplySpellingResourcePaths(r *SpellingCheckRule, ignoreRel, spellingRel, prohibitRel string) {
	if r == nil {
		return
	}
	loadIgnore := func(rel string) {
		if rel == "" {
			return
		}
		p := DiscoverSpellingResource(strings.TrimPrefix(rel, "/"))
		if p == "" {
			return
		}
		words, err := LoadSpellingWordListFile(p)
		if err != nil {
			return
		}
		r.AddIgnoreWords(words...)
	}
	loadProhibit := func(rel string) {
		if rel == "" {
			return
		}
		p := DiscoverSpellingResource(strings.TrimPrefix(rel, "/"))
		if p == "" {
			return
		}
		words, err := LoadSpellingWordListFile(p)
		if err != nil {
			return
		}
		// Java init: addProhibitedWords(expandLine(line))
		for _, line := range words {
			expanded := r.ExpandLine(line)
			if len(expanded) == 0 {
				continue
			}
			r.AddProhibitedWords(expanded...)
		}
	}
	loadIgnore(ignoreRel)
	loadIgnore(spellingRel)
	loadProhibit(prohibitRel)
}

// ApplyVariantSpellingFile loads Java getLanguageVariantSpellingFileName words
// into IgnoreWords (accepted spellings). Missing file is a no-op.
func ApplyVariantSpellingFile(r *SpellingCheckRule, relClasspath string) {
	if r == nil || relClasspath == "" {
		return
	}
	p := DiscoverSpellingResource(relClasspath)
	if p == "" {
		return
	}
	words, err := LoadSpellingWordListFile(p)
	if err != nil {
		return
	}
	r.AddIgnoreWords(words...)
}

func discoverResourceRel(rel string) string {
	rel = strings.TrimPrefix(rel, "/")
	if rel == "" {
		return ""
	}
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for {
		var candidates []string
		// Root-level core resources (e.g. spelling_global.txt).
		candidates = append(candidates,
			filepath.Join(dir, "inspiration", "languagetool", "languagetool-core",
				"src", "main", "resources", "org", "languagetool", "resource", rel),
			filepath.Join(dir, "third_party", "languagetool-dicts", "org", "languagetool", "resource", rel),
			filepath.Join(dir, "third_party", "english-pos-dict", "org", "languagetool", "resource", rel),
			filepath.Join(dir, "testdata", "upstream", rel),
		)
		// Language-module paths: "pl/hunspell/ignore.txt" → resource/pl/hunspell/ignore.txt
		if lang, rest, ok := strings.Cut(rel, "/"); ok && lang != "" && rest != "" {
			candidates = append(candidates,
				filepath.Join(dir, "inspiration", "languagetool", "languagetool-language-modules", lang,
					"src", "main", "resources", "org", "languagetool", "resource", lang, rest),
				filepath.Join(dir, "inspiration", "languagetool", "languagetool-language-modules", lang,
					"src", "main", "resources", "org", "languagetool", "resource", rel),
			)
		}
		for _, p := range candidates {
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

// LoadSpellingWordListFile ports CachingWordListLoader.loadWords for a filesystem path:
// skip empty lines and lines starting with #; strip trailing # comments
// (StringUtils.substringBefore(line.trim(), "#").trim()).
// Does NOT strip Hunspell flags after '/' — Java keeps "word/S" for expandLine.
func LoadSpellingWordListFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []string
	sc := bufio.NewScanner(f)
	// Large prohibit/spelling files (e.g. de/hunspell/prohibit.txt).
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = tools.JavaStringTrim(line)
		if i := strings.Index(line, "#"); i >= 0 {
			line = tools.JavaStringTrim(line[:i])
		}
		if line != "" {
			out = append(out, line)
		}
	}
	return out, sc.Err()
}

// ApplyDefaultSpellingWordLists ports SpellingCheckRule.init():
//
//	loadWords(getIgnoreFileName()) → addIgnoreWords
//	if getSpellingFileName() != null → loadWords → addIgnoreWords
//	for getAdditionalSpellingFileNames() if resourceExists → addIgnoreWords
//	loadWords(getProhibitFileName()) → expandLine → addProhibitedWords
//	for getAdditionalProhibitFileNames() → expandLine → addProhibitedWords
//
// Plus variant file from getLanguageVariantSpellingFileName (Java Morfologik initSpeller).
// Multi-token lines go to MultiWordIgnore (Java IGNORE_SPELLING antipatterns).
// Missing files are skipped (fail closed, no invent) — Java resourceExists false.
func ApplyDefaultSpellingWordLists(r *SpellingCheckRule) {
	if r == nil {
		return
	}
	loadIgnoreClasspath := func(rel string) {
		if rel == "" {
			return
		}
		// Prefer exact resource path; fall back to DiscoverLang hunspell/spelling/root layouts.
		p := DiscoverSpellingResource(strings.TrimPrefix(rel, "/"))
		if p == "" {
			// e.g. "en/hunspell/ignore.txt" → short "en", name "ignore.txt"
			rel2 := strings.TrimPrefix(rel, "/")
			if lang, rest, ok := strings.Cut(rel2, "/"); ok {
				// rest may be "hunspell/ignore.txt" or "ignore.txt"
				if strings.HasPrefix(rest, "hunspell/") {
					p = DiscoverLangHunspellWordList(lang, strings.TrimPrefix(rest, "hunspell/"))
				} else if strings.HasPrefix(rest, "spelling/") {
					p = DiscoverLangHunspellWordList(lang, strings.TrimPrefix(rest, "spelling/"))
				} else {
					p = DiscoverLangHunspellWordList(lang, rest)
				}
			}
		}
		if p == "" {
			return
		}
		words, err := LoadSpellingWordListFile(p)
		if err != nil {
			return
		}
		r.AddIgnoreWords(words...)
	}
	loadProhibitClasspath := func(rel string) {
		if rel == "" {
			return
		}
		p := DiscoverSpellingResource(strings.TrimPrefix(rel, "/"))
		if p == "" {
			rel2 := strings.TrimPrefix(rel, "/")
			if lang, rest, ok := strings.Cut(rel2, "/"); ok {
				if strings.HasPrefix(rest, "hunspell/") {
					p = DiscoverLangHunspellWordList(lang, strings.TrimPrefix(rest, "hunspell/"))
				} else {
					p = DiscoverLangHunspellWordList(lang, rest)
				}
			}
		}
		if p == "" {
			return
		}
		words, err := LoadSpellingWordListFile(p)
		if err != nil {
			return
		}
		// Java: addProhibitedWords(expandLine(prohibitedWord)) per line.
		for _, line := range words {
			expanded := r.ExpandLine(line)
			if len(expanded) == 0 {
				continue
			}
			r.AddProhibitedWords(expanded...)
		}
	}

	// Java init order
	loadIgnoreClasspath(r.GetIgnoreFileName())
	if name := r.GetSpellingFileName(); name != "" {
		loadIgnoreClasspath(name)
	}
	for _, name := range r.GetAdditionalSpellingFileNames() {
		loadIgnoreClasspath(name)
	}
	loadProhibitClasspath(r.GetProhibitFileName())
	for _, name := range r.GetAdditionalProhibitFileNames() {
		loadProhibitClasspath(name)
	}
	// Variant spelling file (Java MorfologikSpellerRule / language variant overrides).
	ApplyVariantSpellingFile(r, r.GetLanguageVariantSpellingFileName())
}
