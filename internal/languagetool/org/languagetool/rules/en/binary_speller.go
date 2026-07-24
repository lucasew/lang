package en

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"sync"

	atticmorfo "github.com/lucasew/lang/internal/attic/morfologik"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// CommonDemoSpellerSuggestions is a fixed typo map for LANG_DEMO_SPELLER only.
// Not used by the binary CFSA2 speller (that uses SuggestEdits — no invent map).
var CommonDemoSpellerSuggestions = map[string][]string{
	"teh":        {"the"},
	"recieve":    {"receive"},
	"seperate":   {"separate"},
	"occured":    {"occurred"},
	"definately": {"definitely"},
	"accomodate": {"accommodate"},
	"untill":     {"until"},
	"wich":       {"which"},
	"thier":      {"their"},
}

// EnglishVariantSpellerMeta ports Java Morfologik*SpellerRule.getId + getFileName basename
// for a language code (en, en-US, en-GB, …). Unknown/default → American (en_US).
func EnglishVariantSpellerMeta(langCode string) (ruleID, dictFile string) {
	code := strings.ToLower(langCode)
	switch {
	case strings.Contains(code, "gb"):
		return MorfologikBritishSpellerRuleID, "en_GB.dict"
	case strings.Contains(code, "-ca") || strings.HasSuffix(code, "_ca"):
		return MorfologikCanadianSpellerRuleID, "en_CA.dict"
	case strings.Contains(code, "au"):
		return MorfologikAustralianSpellerRuleID, "en_AU.dict"
	case strings.Contains(code, "nz"):
		return MorfologikNewZealandSpellerRuleID, "en_NZ.dict"
	case strings.Contains(code, "za"):
		return MorfologikSouthAfricanSpellerRuleID, "en_ZA.dict"
	default:
		return MorfologikAmericanSpellerRuleID, "en_US.dict"
	}
}

// DiscoverEnglishVariantDictFile walks for a hunspell dict basename (e.g. en_GB.dict)
// under third_party english-pos-dict and inspiration en resources. Empty if missing.
func DiscoverEnglishVariantDictFile(dictFile string) string {
	if dictFile == "" {
		return ""
	}
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for {
		for _, rel := range []string{
			filepath.Join("third_party", "english-pos-dict", "org", "languagetool", "resource", "en", "hunspell", dictFile),
			filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "en",
				"src", "main", "resources", "org", "languagetool", "resource", "en", "hunspell", dictFile),
		} {
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

// RegisterBinaryEnglishSpeller installs MORFOLOGIK_RULE_EN_US backed by a CFSA2
// en_US.dict (attic morfologik loader). Returns false if the dictionary cannot be opened.
// nearestKnown is an optional small word set for edit-distance suggestions (not the full dict).
// suggestions is an optional extra map merged with SuggestEdits; nil means dict-only (no invent).
func RegisterBinaryEnglishSpeller(lt *languagetool.JLanguageTool, dictPath string, nearestKnown map[string]struct{}, suggestions map[string][]string) bool {
	return RegisterBinaryEnglishSpellerID(lt, dictPath, MorfologikAmericanSpellerRuleID, nearestKnown, suggestions)
}

// RegisterBinaryEnglishSpellerID is RegisterBinaryEnglishSpeller with an explicit
// Java Morfologik*SpellerRule getId (e.g. MORFOLOGIK_RULE_EN_GB).
func RegisterBinaryEnglishSpellerID(lt *languagetool.JLanguageTool, dictPath, ruleID string, nearestKnown map[string]struct{}, suggestions map[string][]string) bool {
	if lt == nil || dictPath == "" {
		return false
	}
	if ruleID == "" {
		ruleID = MorfologikAmericanSpellerRuleID
	}
	d, err := atticmorfo.OpenDictionary(dictPath)
	if err != nil || d == nil {
		return false
	}
	// Locale for ignore/spelling/global/variant lists (Java SpellingCheckRule.init +
	// getLanguageVariantSpellingFileName).
	langCode := englishLocaleFromSpellerRuleID(ruleID)
	if lt.GetLanguageCode() != "" {
		langCode = lt.GetLanguageCode()
	}
	meta := spelling.NewSpellingCheckRule(ruleID, "Possible spelling mistake", langCode)
	spelling.ApplyDefaultSpellingWordLists(meta)
	// Java SpellingCheckRule: words in hunspell/prohibit.txt are errors even when
	// the dictionary accepts them (isProhibited). Merge EN-specific prohibit.txt loader.
	prohibited := englishProhibitedWords()
	isKnown := func(w string) bool {
		if isEnglishProhibited(prohibited, w) || meta.IsProhibited(w) {
			return false
		}
		if _, ok := meta.IgnoreWords[w]; ok {
			return true
		}
		if d.Contains(w) {
			return true
		}
		low := strings.ToLower(w)
		if low != w {
			if _, ok := meta.IgnoreWords[low]; ok {
				return true
			}
			if d.Contains(low) {
				return true
			}
		}
		return false
	}
	if suggestions == nil {
		suggestions = map[string][]string{}
	}
	// CFSA2 edit-distance suggestions (Java Morfologik). Filter prohibited.
	suggestFn := func(w string) []string {
		raw := d.SuggestEdits(w, 8)
		out := make([]string, 0, len(raw))
		for _, s := range raw {
			if isEnglishProhibited(prohibited, s) || meta.IsProhibited(s) {
				continue
			}
			out = append(out, s)
		}
		return out
	}
	// Wrap so multi-token IGNORE_SPELLING phrases (spelling_global / spelling.txt) apply.
	inner := languagetool.SimplePredicateSpellerChecker(
		ruleID, isKnown, suggestions, nearestKnown, suggestFn,
	)
	lt.AddRuleChecker(ruleID, func(s *languagetool.AnalyzedSentence) []languagetool.LocalMatch {
		meta.MarkMultiWordIgnoreSpelling(s)
		return inner(s)
	})
	return true
}

// englishLocaleFromSpellerRuleID maps Java Morfologik* getId → shortCodeWithCountry.
func englishLocaleFromSpellerRuleID(ruleID string) string {
	switch ruleID {
	case MorfologikBritishSpellerRuleID:
		return "en-GB"
	case MorfologikCanadianSpellerRuleID:
		return "en-CA"
	case MorfologikAustralianSpellerRuleID:
		return "en-AU"
	case MorfologikNewZealandSpellerRuleID:
		return "en-NZ"
	case MorfologikSouthAfricanSpellerRuleID:
		return "en-ZA"
	default:
		return "en-US"
	}
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
		line := tools.JavaStringTrim(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = tools.JavaStringTrim(line[:i])
		}
		if line == "" {
			continue
		}
		// Java expandLine may expand suffixes; soft path accepts whole lines as-is.
		out[line] = struct{}{}
	}
	return out, sc.Err()
}
