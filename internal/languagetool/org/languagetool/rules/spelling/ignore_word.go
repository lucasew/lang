package spelling

import (
	"sort"
	"strings"
	"unicode"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// IgnoreWord ports SpellingCheckRule.ignoreWord(String).
// Tokens with no letters, oversize tokens, ignore set (optional case-fold),
// and ignoreWordsWithLength are accepted without dictionary lookup.
// All length checks use Java String.length() (UTF-16 code units).
func (r *SpellingCheckRule) IgnoreWord(word string) bool {
	if r == nil {
		return false
	}
	// Java: word.length() > MAX_TOKEN_LENGTH
	if javaStringLenSpell(word) > MaxTokenLength {
		return true
	}
	// Java considerIgnoreWords default true; when false, skip ignore-set path entirely.
	if !r.considerIgnoreWords() {
		return false
	}
	// Tokens with no letters cannot have spelling errors.
	// Java isLatinScript(): true → pHasNoLetterLatin (no Latin letters → ignore);
	// false → pHasNoLetter (no \p{L} → ignore).
	if r.NonLatinScript {
		if !wordHasLetter(word) {
			return true
		}
	} else if !wordHasLatinLetter(word) {
		return true
	}
	// Trailing period (e.g. sentence-end token) — check form without '.'
	// Java: word.substring(0, word.length()-1) when ends with '.'
	if strings.HasSuffix(word, ".") && !r.IsInIgnoredSet(word) {
		return r.IsIgnoredNoCase(javaSubstringSpell(word, 0, javaStringLenSpell(word)-1))
	}
	return r.IsIgnoredNoCase(word)
}

// IsInIgnoredSet ports isInIgnoredSet (exact membership of wordsToBeIgnored).
func (r *SpellingCheckRule) IsInIgnoredSet(word string) bool {
	if r == nil || word == "" || len(r.IgnoreWords) == 0 {
		return false
	}
	_, ok := r.IgnoreWords[word]
	return ok
}

// IsIgnoredNoCase ports isIgnoredNoCase.
func (r *SpellingCheckRule) IsIgnoredNoCase(word string) bool {
	if r == nil {
		return false
	}
	if r.IsInIgnoredSet(word) {
		return true
	}
	// Case conversion only when not mixed case and convertsCase (Java Morfologik setConvertsCase).
	if r.ConvertsCase && !tools.IsMixedCase(word) && r.IsInIgnoredSet(strings.ToLower(word)) {
		return true
	}
	// Java: ignoreWordsWithLength > 0 && word.length() <= ignoreWordsWithLength
	if r.IgnoreWordsWithLength > 0 && javaStringLenSpell(word) <= r.IgnoreWordsWithLength {
		return true
	}
	return false
}

// IgnoreToken ports SpellingCheckRule.ignoreToken(tokens, idx):
// default builds word list and calls ignoreWord(words.get(idx)).
func (r *SpellingCheckRule) IgnoreToken(tokens []*languagetool.AnalyzedTokenReadings, idx int) bool {
	if r == nil || idx < 0 || idx >= len(tokens) || tokens[idx] == nil {
		return false
	}
	// Optional language override hook.
	if r.IgnoreTokenFn != nil {
		return r.IgnoreTokenFn(tokens, idx)
	}
	return r.IgnoreWord(tokens[idx].GetToken())
}

// IgnorePotentiallyMisspelledWord ports SpellingCheckRule.ignorePotentiallyMisspelledWord.
// Java default is false; languages override via IgnorePotentiallyMisspelledWordFn
// (e.g. NL CompoundAcceptor, DE compound gender paths).
func (r *SpellingCheckRule) IgnorePotentiallyMisspelledWord(word string) bool {
	if r == nil || r.IgnorePotentiallyMisspelledWordFn == nil {
		return false
	}
	return r.IgnorePotentiallyMisspelledWordFn(word)
}

// StartsWithIgnoredWord ports startsWithIgnoredWord (longest ignored-word prefix length).
// Returns 0 when word length < 4 or no ignored prefix matches.
// Return value and comparisons use Java String.length() (UTF-16).
func (r *SpellingCheckRule) StartsWithIgnoredWord(word string, caseSensitive bool) int {
	if r == nil || javaStringLenSpell(word) < 4 || len(r.IgnoreWords) == 0 {
		return 0
	}
	arr := r.sortedIgnoreArray(caseSensitive)
	if len(arr) == 0 {
		return 0
	}
	w := word
	for w != "" {
		i := sort.Search(len(arr), func(i int) bool {
			if caseSensitive {
				return arr[i] >= w
			}
			return strings.ToLower(arr[i]) >= strings.ToLower(w)
		})
		if i < len(arr) && equalIgnore(arr[i], w, caseSensitive) {
			return javaStringLenSpell(w)
		}
		// Java: prev = -result - 2 after binarySearch miss
		prev := i - 1
		if prev < 0 {
			return 0
		}
		common := commonPrefixLenUTF16(w, arr[prev], caseSensitive)
		if common >= javaStringLenSpell(w) {
			// should not happen if not equal
			return 0
		}
		if common == 0 {
			return 0
		}
		// Java: word = caseSensitive ? commonPrefix : word.substring(0, commonPrefix.length())
		w = javaSubstringSpell(w, 0, common)
	}
	return 0
}

func (r *SpellingCheckRule) considerIgnoreWords() bool {
	if r == nil {
		return true
	}
	// ConsiderIgnoreWords defaults true (Java); false only when explicitly disabled.
	return !r.DisableConsiderIgnoreWords
}

func (r *SpellingCheckRule) sortedIgnoreArray(caseSensitive bool) []string {
	if r == nil {
		return nil
	}
	if caseSensitive {
		if r.ignoreDictSorted == nil {
			r.ignoreDictSorted = sortedKeys(r.IgnoreWords, true)
		}
		return r.ignoreDictSorted
	}
	if r.ignoreDictSortedFold == nil {
		r.ignoreDictSortedFold = sortedKeys(r.IgnoreWords, false)
	}
	return r.ignoreDictSortedFold
}

func sortedKeys(m map[string]struct{}, caseSensitive bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	if caseSensitive {
		sort.Strings(out)
	} else {
		sort.Slice(out, func(i, j int) bool {
			return strings.ToLower(out[i]) < strings.ToLower(out[j])
		})
	}
	return out
}

func equalIgnore(a, b string, caseSensitive bool) bool {
	if caseSensitive {
		return a == b
	}
	return strings.EqualFold(a, b)
}

// commonPrefixLenUTF16 ports Guava Strings.commonPrefix length on Java Strings (UTF-16 units).
func commonPrefixLenUTF16(a, b string, caseSensitive bool) int {
	ua, ub := javaCharsSpell(a), javaCharsSpell(b)
	if !caseSensitive {
		// Compare lowercased UTF-16 sequences (Java Locale.ROOT toLowerCase on full string then prefix).
		ua = javaCharsSpell(strings.ToLower(a))
		ub = javaCharsSpell(strings.ToLower(b))
	}
	n := len(ua)
	if len(ub) < n {
		n = len(ub)
	}
	i := 0
	for i < n && ua[i] == ub[i] {
		i++
	}
	return i
}

func javaStringLenSpell(s string) int {
	return len(utf16.Encode([]rune(s)))
}

func javaCharsSpell(s string) []uint16 {
	return utf16.Encode([]rune(s))
}

func javaSubstringSpell(s string, from, to int) string {
	u := javaCharsSpell(s)
	if from < 0 {
		from = 0
	}
	if to > len(u) {
		to = len(u)
	}
	if from >= to {
		return ""
	}
	return string(utf16.Decode(u[from:to]))
}

func wordHasLetter(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

// wordHasLatinLetter ports pHasNoLetterLatin inverted: any Latin-script letter.
func wordHasLatinLetter(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Latin, r) {
			return true
		}
	}
	return false
}
