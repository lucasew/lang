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
		line := tools.JavaStringTrim(sc.Text())
		line = strings.TrimPrefix(line, "\ufeff")
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = tools.JavaStringTrim(line[:i])
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
			line := tools.JavaStringTrim(sc.Text())
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
		line := tools.JavaStringTrim(sc.Text())
		line = strings.TrimPrefix(line, "\ufeff")
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, "=")
		if len(parts) != 2 {
			continue
		}
		a, b := tools.JavaStringTrim(parts[0]), tools.JavaStringTrim(parts[1])
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
	// Java RuleMatch FromPos/ToPos are UTF-16 code units (String.substring).
	return rules.UTF16Substring(sent.GetText(), m.GetFromPos(), m.GetToPos())
}

// dialectAlternative ports MorfologikPortugueseSpellerRule.dialectAlternative:
// 1) surface lemma map hit; 2) tagger lemma → map → synthesizer form for same POS tag.
func (r *MorfologikPortugueseSpellerRule) dialectAlternative(word string) string {
	if r == nil || r.dialectMap == nil || word == "" {
		return ""
	}
	// Naive check: word is identical to lemma present on the TXT list
	if v, ok := r.dialectMap[strings.ToLower(word)]; ok {
		return v
	}
	// Annoying check: the word's lemma exists on the TXT list → synthesize form
	if r.TagLemma == nil || r.TagPOS == nil || r.SynthesizeForPosTags == nil {
		return ""
	}
	lemmas := r.TagLemma(word)
	tags := r.TagPOS(word)
	if len(lemmas) == 0 || len(tags) == 0 {
		return ""
	}
	// Java iterates readings; pair lemma+tag from same reading. We approximate with
	// all combinations (same as expanding each reading independently when tags/lemmas align).
	// Prefer zip by index when lengths match (typical single-token tag path).
	n := len(lemmas)
	if len(tags) < n {
		n = len(tags)
	}
	for i := 0; i < n; i++ {
		lemma := lemmas[i]
		tag := tags[i]
		if tag == "" || lemma == "" {
			continue
		}
		candidate, ok := r.dialectMap[strings.ToLower(lemma)]
		if !ok {
			continue
		}
		want := tag
		forms := r.SynthesizeForPosTags(candidate, func(tagStr string) bool {
			return tagStr == want
		})
		if len(forms) > 0 {
			return forms[0]
		}
	}
	// Fallback: any lemma in map with any tag of the word (Java only uses same reading tag).
	for _, lemma := range lemmas {
		candidate, ok := r.dialectMap[strings.ToLower(lemma)]
		if !ok {
			continue
		}
		for _, tag := range tags {
			if tag == "" {
				continue
			}
			want := tag
			forms := r.SynthesizeForPosTags(candidate, func(tagStr string) bool {
				return tagStr == want
			})
			if len(forms) > 0 {
				return forms[0]
			}
		}
	}
	return ""
}

// dialectAlternativeSurface is the surface-only arm (tests / fail-closed without tagger).
func (r *MorfologikPortugueseSpellerRule) dialectAlternativeSurface(word string) string {
	if r == nil || r.dialectMap == nil {
		return ""
	}
	if v, ok := r.dialectMap[strings.ToLower(word)]; ok {
		return v
	}
	return ""
}

// wordIsMisspelled uses Multi/speller1 first (Java speller1.isMisspelled), then IsMisspelled hook.
func (r *MorfologikPortugueseSpellerRule) wordIsMisspelled(word string) bool {
	if r == nil || word == "" {
		return false
	}
	if r.Multi != nil && len(r.Multi.Spellers) > 0 {
		return r.Multi.IsMisspelled(word)
	}
	if r.IsMisspelled != nil {
		return r.IsMisspelled(word)
	}
	if r.Speller != nil {
		return r.Speller.IsMisspelled(word)
	}
	return false
}

// wordSuggestions ports speller1.getSuggestions: Multi first, then map Speller / filter dict.
func (r *MorfologikPortugueseSpellerRule) wordSuggestions(word string) []string {
	if r == nil || word == "" {
		return nil
	}
	if r.Multi != nil {
		if sugs := r.Multi.GetSuggestions(word); len(sugs) > 0 {
			return sugs
		}
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
