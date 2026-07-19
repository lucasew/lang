package pt

import (
	"bufio"
	"os"
	"regexp"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// Portuguese speller extras from MorfologikPortugueseSpellerRule.
// Clitic-verb validity needs TagPOS (Java PortugueseTagger); fail-closed when unset.

var (
	doNotSuggestOnce sync.Once
	doNotSuggestSet  map[string]struct{}

	abbrevOnce sync.Once
	abbrevSet  map[string]struct{}

	// Java posTag.matches("V.+:P.+")
	cliticVerbPOS = regexp.MustCompile(`^V.+:P.+`)
)

func loadWordSetLower(rel string) map[string]struct{} {
	out := map[string]struct{}{}
	p := spelling.DiscoverSpellingResource(rel)
	if p == "" {
		return out
	}
	f, err := os.Open(p)
	if err != nil {
		return out
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		line = strings.TrimPrefix(line, "\ufeff")
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		if line != "" {
			out[strings.ToLower(line)] = struct{}{}
		}
	}
	return out
}

func getDoNotSuggestWords() map[string]struct{} {
	doNotSuggestOnce.Do(func() {
		doNotSuggestSet = loadWordSetLower("pt/do_not_suggest.txt")
	})
	return doNotSuggestSet
}

func getAbbreviations() map[string]struct{} {
	abbrevOnce.Do(func() {
		out := map[string]struct{}{}
		p := spelling.DiscoverSpellingResource("pt/abbreviations.txt")
		if p == "" {
			abbrevSet = out
			return
		}
		f, err := os.Open(p)
		if err != nil {
			abbrevSet = out
			return
		}
		defer f.Close()
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			line = strings.TrimPrefix(line, "\ufeff")
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			out[line] = struct{}{}
			out[strings.ToLower(line)] = struct{}{}
		}
		abbrevSet = out
	})
	return abbrevSet
}

// loadDialectAlternationMapping ports getDialectAlternationMapping.
// Lines formA=formB. pt-BR keys column1 → value column0; pt-PT keys column0 → value1.
func loadDialectAlternationMapping(variantCode string) map[string]string {
	col := -1
	switch variantCode {
	case "pt-BR":
		col = 1
	case "pt-PT":
		col = 0
	default:
		return map[string]string{}
	}
	p := spelling.DiscoverSpellingResource("pt/dialect_alternations.txt")
	if p == "" {
		return map[string]string{}
	}
	f, err := os.Open(p)
	if err != nil {
		return map[string]string{}
	}
	defer f.Close()
	out := map[string]string{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		line = strings.TrimPrefix(line, "\ufeff")
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, "=")
		if len(parts) != 2 {
			continue
		}
		a, b := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		if col == 1 {
			out[strings.ToLower(b)] = a
		} else {
			out[strings.ToLower(a)] = b
		}
	}
	return out
}

func checkDiaeresis(word string) string {
	if strings.Contains(word, "ü") {
		return strings.ReplaceAll(word, "ü", "u")
	}
	return ""
}

func checkEuropeanStyle1PLPastTense(variantCode, word string) string {
	if variantCode != "pt-BR" || !strings.HasSuffix(word, "ámos") {
		return ""
	}
	return strings.ReplaceAll(word, "á", "a")
}

func isTitlecasedHyphenatedWord(parts []string) bool {
	for _, part := range parts {
		if tools.IsMixedCase(part) {
			return false
		}
	}
	return true
}

func isAbbreviation(word string) bool {
	m := getAbbreviations()
	if len(m) == 0 {
		return false
	}
	if _, ok := m[word+"."]; ok {
		return true
	}
	if _, ok := m[strings.ToLower(word)+"."]; ok {
		return true
	}
	return false
}

func filterDoNotSuggest(sugs []string) []string {
	deny := getDoNotSuggestWords()
	if len(deny) == 0 || len(sugs) == 0 {
		return sugs
	}
	out := make([]string, 0, len(sugs))
	for _, s := range sugs {
		if _, bad := deny[strings.ToLower(s)]; !bad {
			out = append(out, s)
		}
	}
	return out
}

func startsWithUppercaseLetter(s string) bool {
	if s == "" {
		return false
	}
	r, _ := utf8.DecodeRuneInString(s)
	return unicode.IsUpper(r)
}

func matchSurface(m *rules.RuleMatch, sent *languagetool.AnalyzedSentence) string {
	if m == nil || sent == nil {
		return ""
	}
	text := sent.GetText()
	from, to := m.GetFromPos(), m.GetToPos()
	if from < 0 || from >= to {
		return ""
	}
	// AnalyzedTokenReadings positions are character offsets (Java char / Go runes for BMP).
	// Slicing UTF-8 by byte index truncates multi-byte letters (á, ü, …).
	runes := []rune(text)
	if to <= len(runes) {
		return string(runes[from:to])
	}
	// Fallback: treat as byte offsets when consistent with len(text).
	if to <= len(text) {
		return text[from:to]
	}
	return ""
}

func (r *MorfologikPortugueseSpellerRule) dialectAlternativeSurface(word string) string {
	if r == nil || r.dialectMap == nil {
		return ""
	}
	if v, ok := r.dialectMap[strings.ToLower(word)]; ok {
		return v
	}
	return ""
}

// wordIsMisspelled uses SpellingCheckRule.IsMisspelled when set (FilterDict path),
// else map Speller — same order Java multi-speller uses for isMisspelled.
func (r *MorfologikPortugueseSpellerRule) wordIsMisspelled(word string) bool {
	if r == nil || word == "" {
		return false
	}
	if r.IsMisspelled != nil {
		return r.IsMisspelled(word)
	}
	if r.Speller != nil {
		return r.Speller.IsMisspelled(word)
	}
	return false
}

// wordSuggestions ports speller1.getSuggestions: map Speller first, then wired CFSA2 dict.
func (r *MorfologikPortugueseSpellerRule) wordSuggestions(word string) []string {
	if r == nil || word == "" {
		return nil
	}
	if r.Speller != nil {
		if sugs := r.Speller.FindReplacements(word); len(sugs) > 0 {
			return sugs
		}
	}
	if FilterDictAvailable() {
		return FilterDictSuggest(word)
	}
	return nil
}

func (r *MorfologikPortugueseSpellerRule) checkCompoundElements(parts []string) string {
	if r == nil || len(parts) == 0 {
		return ""
	}
	suggested := make([]string, 0, len(parts))
	for _, part := range parts {
		if r.wordIsMisspelled(part) {
			sugs := r.wordSuggestions(part)
			if len(sugs) == 0 {
				suggested = append(suggested, part)
			} else {
				suggested = append(suggested, sugs[0])
			}
		} else {
			suggested = append(suggested, part)
		}
	}
	joined := strings.Join(suggested, "-")
	if joined == strings.Join(parts, "-") {
		return ""
	}
	return joined
}

// isValidCliticVerb ports isValidCliticVerb when TagPOS is set.
// Without TagPOS returns false (fail-closed).
func (r *MorfologikPortugueseSpellerRule) isValidCliticVerb(word string) bool {
	if r == nil || r.TagPOS == nil || word == "" {
		return false
	}
	tags := r.TagPOS(word)
	hasClitic := false
	for _, tag := range tags {
		if cliticVerbPOS.MatchString(tag) {
			hasClitic = true
			break
		}
	}
	if !hasClitic {
		return false
	}
	// Java: invalid if lemma is in dialectAlternationMapping (map keys are lowercased at load).
	if r.TagLemma != nil && r.dialectMap != nil {
		for _, lemma := range r.TagLemma(word) {
			if _, bad := r.dialectMap[strings.ToLower(lemma)]; bad {
				return false
			}
		}
	}
	return true
}
