package en

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"sync"

	atticmorfo "github.com/lucasew/lang/internal/attic/morfologik"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// CommonDemoSpellerSuggestions is a soft map of frequent EN typos → fixes.
// Used with binary and map spellers for suggestion UX until full Morfologik
// suggestion generation is wired.
var CommonDemoSpellerSuggestions = map[string][]string{
	"teh":      {"the"},
	"recieve":  {"receive"},
	"seperate": {"separate"},
	"occured":  {"occurred"},
	"definately": {"definitely"},
	"accomodate": {"accommodate"},
	"untill":   {"until"},
	"wich":     {"which"},
	"thier":    {"their"},
}

// RegisterBinaryEnglishSpeller installs MORFOLOGIK_RULE_EN_US backed by a CFSA2
// en_US.dict (attic morfologik loader). Returns false if the dictionary cannot be opened.
// nearestKnown is an optional small word set for edit-distance suggestions (not the full dict).
// suggestions may be nil (uses CommonDemoSpellerSuggestions).
func RegisterBinaryEnglishSpeller(lt *languagetool.JLanguageTool, dictPath string, nearestKnown map[string]struct{}, suggestions map[string][]string) bool {
	if lt == nil || dictPath == "" {
		return false
	}
	d, err := atticmorfo.OpenDictionary(dictPath)
	if err != nil || d == nil {
		return false
	}
	// Java SpellingCheckRule: words in hunspell/prohibit.txt are errors even when
	// the dictionary accepts them (isProhibited).
	prohibited := englishProhibitedWords()
	isKnown := func(w string) bool {
		if isEnglishProhibited(prohibited, w) {
			return false
		}
		if d.Contains(w) {
			return true
		}
		low := strings.ToLower(w)
		if low != w && d.Contains(low) {
			return true
		}
		return false
	}
	if suggestions == nil {
		suggestions = CommonDemoSpellerSuggestions
	}
	// CFSA2 edit-1 candidates via Contains (no full dict scan); then soft nearest set.
	// Java also filters prohibited strings from suggestions.
	suggestFn := func(w string) []string {
		raw := d.SuggestEdits(w, 8)
		if len(prohibited) == 0 {
			return raw
		}
		out := make([]string, 0, len(raw))
		for _, s := range raw {
			if !isEnglishProhibited(prohibited, s) {
				out = append(out, s)
			}
		}
		return out
	}
	lt.AddRuleChecker("MORFOLOGIK_RULE_EN_US", languagetool.SimplePredicateSpellerChecker(
		"MORFOLOGIK_RULE_EN_US", isKnown, suggestions, nearestKnown, suggestFn,
	))
	return true
}

// Cached prohibit.txt (Java SpellingCheckRule.getProhibitFileName → en/hunspell/prohibit.txt).
var (
	enProhibitOnce sync.Once
	enProhibitSet  map[string]struct{}
)

func englishProhibitedWords() map[string]struct{} {
	enProhibitOnce.Do(func() {
		enProhibitSet = map[string]struct{}{}
		for _, p := range discoverEnglishProhibitPaths() {
			if loaded, err := loadProhibitWords(p); err == nil {
				for k := range loaded {
					enProhibitSet[k] = struct{}{}
				}
			}
		}
	})
	return enProhibitSet
}

func isEnglishProhibited(set map[string]struct{}, w string) bool {
	if len(set) == 0 || w == "" {
		return false
	}
	// Java isProhibited: exact set membership (case-sensitive).
	if _, ok := set[w]; ok {
		return true
	}
	return false
}

func discoverEnglishProhibitPaths() []string {
	var out []string
	seen := map[string]struct{}{}
	add := func(p string) {
		if p == "" {
			return
		}
		if _, ok := seen[p]; ok {
			return
		}
		if st, err := os.Stat(p); err != nil || !st.Mode().IsRegular() {
			return
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	wd, err := os.Getwd()
	if err != nil {
		return out
	}
	dir := wd
	for {
		for _, rel := range []string{
			filepath.Join("testdata", "upstream", "en", "resource", "hunspell", "prohibit.txt"),
			filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "en",
				"src", "main", "resources", "org", "languagetool", "resource", "en", "hunspell", "prohibit.txt"),
			// SpellingCheckRule additional prohibit_custom.txt
			filepath.Join("testdata", "upstream", "en", "resource", "hunspell", "prohibit_custom.txt"),
		} {
			add(filepath.Join(dir, rel))
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return out
}

func loadProhibitWords(path string) (map[string]struct{}, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	out := map[string]struct{}{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		if line == "" {
			continue
		}
		// Java expandLine may expand suffixes; soft path accepts whole lines as-is.
		out[line] = struct{}{}
	}
	return out, sc.Err()
}
