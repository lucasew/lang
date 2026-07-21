package hunspell

import (
	"strings"
	"unicode"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// maxCompoundSuggestions ports CompoundAwareHunspellRule.MAX_SUGGESTIONS.
const maxCompoundSuggestions = 20

// CompoundAwareHunspellRule ports
// org.languagetool.rules.spelling.hunspell.CompoundAwareHunspellRule —
// combines Hunspell with compound splitting + Morfologik multi-speller suggestions.
type CompoundAwareHunspellRule struct {
	*HunspellRule
	CompoundSplitter tokenizers.CompoundWordTokenizer
	MorfoSpeller     *morfologik.MorfologikMultiSpeller
	// FilterForLanguage ports abstract filterForLanguage (language-specific).
	// Nil → identity.
	FilterForLanguage func(suggestions []string) []string
	// GetCandidatesFn optional override for getCandidates(String) (DE GermanSpellerRule).
	// Nil → compoundSplitter.Tokenize(word) (Java base).
	GetCandidatesFn func(word string) []string
	// GetFilteredSuggestionsFn ports getFilteredSuggestions (default identity).
	GetFilteredSuggestionsFn func(words []string) []string
	// SortSuggestionByQualityFn optional override (DE extends base run-on preference).
	// Nil → base sortSuggestionByQuality.
	SortSuggestionByQualityFn func(misspelling string, suggestions []string) []string
	// MaxSuggestions caps returned suggestions (default MAX_SUGGESTIONS=20).
	MaxSuggestions int
}

func NewCompoundAwareHunspellRule(
	languageCode string,
	dict HunspellDictionary,
	splitter tokenizers.CompoundWordTokenizer,
	morfo *morfologik.MorfologikMultiSpeller,
) *CompoundAwareHunspellRule {
	if splitter == nil {
		splitter = tokenizers.NewSimpleCompoundWordTokenizer()
	}
	r := &CompoundAwareHunspellRule{
		HunspellRule:     NewHunspellRule(languageCode, dict),
		CompoundSplitter: splitter,
		MorfoSpeller:     morfo,
		MaxSuggestions:   maxCompoundSuggestions,
	}
	// Wire SuggestFn so HunspellRule.Match/calcSuggestions use compound-aware getSuggestions
	// (Java CompoundAwareHunspellRule overrides getSuggestions; Go embedding is not virtual).
	r.HunspellRule.SuggestFn = func(word string) []string {
		return r.GetSuggestions(word)
	}
	return r
}

// SpellingFilePaths ports getSpellingFilePaths.
func SpellingFilePaths(langCode string) []string {
	return []string{
		"/" + langCode + "/hunspell/spelling.txt",
		"/" + langCode + "/hunspell/spelling_custom.txt",
		"/" + langCode + "/multitoken-suggest.txt",
		"spelling_global.txt",
	}
}

// Suggest is the public surface for getSuggestions.
func (r *CompoundAwareHunspellRule) Suggest(word string) []string {
	return r.GetSuggestions(word)
}

// GetSuggestions ports CompoundAwareHunspellRule.getSuggestions.
func (r *CompoundAwareHunspellRule) GetSuggestions(word string) []string {
	if r == nil {
		return nil
	}
	// Java: candidates = getCandidates(word); simple = getCorrectWords(candidates)
	candidates := r.getCandidates(word)
	simple := r.getCorrectWords(candidates)
	simple = r.getFilteredSuggestions(simple)

	var noSplit []string
	var noSplitLower []string
	if r.MorfoSpeller != nil {
		noSplit = append(noSplit, r.MorfoSpeller.GetSuggestions(word)...)
		// handleWordEndPunctuation for "." and "..."
		handleWordEndPunctuation(".", word, &noSplit, r.MorfoSpeller)
		handleWordEndPunctuation("...", word, &noSplit, r.MorfoSpeller)
		if tools.StartsWithUppercase(word) && !tools.IsAllUppercase(word) {
			noSplitLower = r.MorfoSpeller.GetSuggestions(strings.ToLower(word))
		}
	}

	// Interleave: noSplit, uppercaseFirst(noSplitLower), simple (Java rotating mix).
	suggestions := interleaveThree(noSplit, upperFirstAll(noSplitLower), simple)
	suggestions = filterDupes(suggestions)
	if r.FilterForLanguage != nil {
		suggestions = r.FilterForLanguage(suggestions)
	}

	sorted := r.sortSuggestionByQuality(word, suggestions)
	max := r.MaxSuggestions
	if max <= 0 {
		max = maxCompoundSuggestions
	}
	if len(sorted) > max {
		sorted = sorted[:max]
	}
	return sorted
}

// getCandidates ports getCandidates(String word) — base returns compoundSplitter.tokenize.
func (r *CompoundAwareHunspellRule) getCandidates(word string) []string {
	if r == nil {
		return nil
	}
	if r.GetCandidatesFn != nil {
		return r.GetCandidatesFn(word)
	}
	if r.CompoundSplitter == nil {
		return []string{word}
	}
	return r.CompoundSplitter.Tokenize(word)
}

// GetCandidatesFromParts ports getCandidates(List<String> parts) for language overrides
// (DE GermanSpellerRule uses this for per-part Morfologik rebuild + Hunspell filter).
func (r *CompoundAwareHunspellRule) GetCandidatesFromParts(parts []string) []string {
	if r == nil || r.Dict == nil || r.MorfoSpeller == nil || len(parts) == 0 {
		return nil
	}
	var candidates []string
	for partCount, part := range parts {
		if r.Dict.Spell(part) {
			continue
		}
		// assume noun, so use uppercase for non-first parts
		doUpperCase := partCount > 0 && !tools.StartsWithUppercase(part)
		probe := part
		if doUpperCase {
			probe = tools.UppercaseFirstChar(part)
		}
		sugs := r.MorfoSpeller.GetSuggestions(probe)
		if len(sugs) == 0 {
			if doUpperCase {
				sugs = r.MorfoSpeller.GetSuggestions(tools.LowercaseFirstChar(part))
			} else {
				sugs = r.MorfoSpeller.GetSuggestions(part)
			}
		}
		appendS := false
		if doUpperCase && strings.HasSuffix(part, "s") {
			// maybe infix-s
			base := strings.TrimSuffix(part, "s")
			sugs = append(sugs, r.MorfoSpeller.GetSuggestions(base)...)
			appendS = true
		}
		for _, suggestion := range sugs {
			partsCopy := append([]string(nil), parts...)
			s := suggestion
			if appendS {
				s = s + "s"
			}
			if partCount > 0 && strings.HasPrefix(parts[partCount], "-") && utf16LenHun(parts[partCount]) > 1 {
				// partsCopy.set(partCount, "-" + uppercaseFirstChar(suggestion.substring(1)));
				rest := s
				if u := utf16.Encode([]rune(rest)); len(u) > 1 {
					rest = string(utf16.Decode(u[1:]))
				} else if len(u) == 1 {
					rest = ""
				}
				partsCopy[partCount] = "-" + tools.UppercaseFirstChar(rest)
			} else if partCount > 0 && !strings.HasSuffix(parts[partCount-1], "-") {
				partsCopy[partCount] = strings.ToLower(s)
			} else {
				partsCopy[partCount] = s
			}
			candidate := strings.Join(partsCopy, "")
			if !r.IsMisspelledWord(candidate) {
				candidates = append(candidates, candidate)
			}
			// Arbeidszimmer → Arbeitszimmer (infix-s + trailing "-" on suggestion)
			if partCount < len(parts)-1 && strings.HasSuffix(part, "s") && strings.HasSuffix(suggestion, "-") {
				partsCopy2 := append([]string(nil), parts...)
				// Java: suggestion.substring(0, suggestion.length()-1) UTF-16
				su := utf16.Encode([]rune(suggestion))
				trimmed := suggestion
				if len(su) >= 1 {
					trimmed = string(utf16.Decode(su[:len(su)-1]))
				}
				partsCopy2[partCount] = trimmed
				infixCandidate := strings.Join(partsCopy2, "")
				if !r.IsMisspelledWord(infixCandidate) {
					candidates = append(candidates, infixCandidate)
				}
			}
		}
	}
	return candidates
}

// getCorrectWords ports getCorrectWords — keep phrases where every token spells in Hunspell.
func (r *CompoundAwareHunspellRule) getCorrectWords(wordsOrPhrases []string) []string {
	if r == nil || r.Dict == nil {
		return nil
	}
	var result []string
	for _, wordOrPhrase := range wordsOrPhrases {
		// Java: tokenizeText(wordOrPhrase) uses nonWordPattern
		var words []string
		if r.HunspellRule != nil {
			words = r.HunspellRule.TokenizeText(wordOrPhrase)
		} else {
			words = tokenizeTextHun(wordOrPhrase)
		}
		ok := true
		for _, w := range words {
			if w == "" {
				continue
			}
			if !r.Dict.Spell(w) {
				ok = false
				break
			}
		}
		if ok {
			result = append(result, wordOrPhrase)
		}
	}
	return result
}

func (r *CompoundAwareHunspellRule) getFilteredSuggestions(words []string) []string {
	if r != nil && r.GetFilteredSuggestionsFn != nil {
		return r.GetFilteredSuggestionsFn(words)
	}
	return words
}

// sortSuggestionByQuality ports CompoundAwareHunspellRule.sortSuggestionByQuality:
// prefer run-on words (space removed equals misspelling) unless a single letter is split off.
func (r *CompoundAwareHunspellRule) sortSuggestionByQuality(misspelling string, suggestions []string) []string {
	if r != nil && r.SortSuggestionByQualityFn != nil {
		return r.SortSuggestionByQualityFn(misspelling, suggestions)
	}
	var result []string
	for _, suggestion := range suggestions {
		if strings.ReplaceAll(suggestion, " ", "") == misspelling && !hasSingleLetterToken(suggestion) {
			// prefer run-on words
			result = append([]string{suggestion}, result...)
		} else {
			result = append(result, suggestion)
		}
	}
	return result
}

// hasSingleLetterToken ports
// Arrays.stream(StringUtils.split(suggestion, ' ')).anyMatch(k -> k.length() == 1)
// — only ASCII space splits (not tabs/newlines); length is UTF-16.
func hasSingleLetterToken(s string) bool {
	for _, p := range splitASCIISpaceOmitEmpty(s) {
		if utf16LenHun(p) == 1 {
			return true
		}
	}
	return false
}

// splitASCIISpaceOmitEmpty ports org.apache.commons.lang3.StringUtils.split(s, ' ').
func splitASCIISpaceOmitEmpty(s string) []string {
	if s == "" {
		return nil
	}
	raw := strings.Split(s, " ")
	out := make([]string, 0, len(raw))
	for _, p := range raw {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// handleWordEndPunctuation ports private handleWordEndPunctuation.
func handleWordEndPunctuation(punct, word string, noSplit *[]string, morfo *morfologik.MorfologikMultiSpeller) {
	if morfo == nil || noSplit == nil || !strings.HasSuffix(word, punct) {
		return
	}
	// Java: word.substring(0, word.length()-punct.length()) — UTF-16 length
	wU := utf16.Encode([]rune(word))
	pU := utf16.Encode([]rune(punct))
	if len(wU) < len(pU) {
		return
	}
	base := string(utf16.Decode(wU[:len(wU)-len(pU)]))
	for _, s := range morfo.GetSuggestions(base) {
		*noSplit = append(*noSplit, s+punct)
	}
}

// tokenizeTextHun ports HunspellRule.tokenizeText with NON_ALPHABETIC default:
// split on non-letters (Java nonWordPattern default "[^\\p{L}]").
func tokenizeTextHun(sentence string) []string {
	if sentence == "" {
		return nil
	}
	var out []string
	var b strings.Builder
	flush := func() {
		if b.Len() > 0 {
			out = append(out, b.String())
			b.Reset()
		}
	}
	for _, r := range sentence {
		if unicode.IsLetter(r) {
			b.WriteRune(r)
		} else {
			flush()
		}
	}
	flush()
	return out
}

func upperFirstAll(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, len(in))
	for i, s := range in {
		out[i] = tools.UppercaseFirstChar(s)
	}
	return out
}

// interleaveThree ports the rotating mix of three suggestion lists.
func interleaveThree(a, b, c []string) []string {
	max := len(a)
	if len(b) > max {
		max = len(b)
	}
	if len(c) > max {
		max = len(c)
	}
	var out []string
	for i := 0; i < max; i++ {
		if i < len(a) {
			out = append(out, a[i])
		}
		if i < len(b) {
			out = append(out, b[i])
		}
		if i < len(c) {
			out = append(out, c[i])
		}
	}
	return out
}

func filterDupes(in []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, s := range in {
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
