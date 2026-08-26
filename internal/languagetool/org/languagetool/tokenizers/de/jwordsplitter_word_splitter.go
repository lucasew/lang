package de

import (
	"errors"
	"fmt"
)

// errInputTooLong ports de.danielnaber.jwordsplitter.InputTooLongException.
var errInputTooLong = errors.New("jwordsplitter: input too long")

// germanWordSplitter ports de.danielnaber.jwordsplitter.GermanWordSplitter
// (extends AbstractWordSplitter).
type germanWordSplitter struct {
	words                  map[string]struct{}
	hideInterfixCharacters bool
	exceptionSplits        *exceptionSplits
	disambiguator          *germanInterfixDisambiguator
	strictMode             bool
	minimumWordLength      int
	maximumWordLength      int
	interfixes             []string
}

// newGermanWordSplitter ports GermanWordSplitter(hideInterfix, words) + init().
func newGermanWordSplitter(hideInterfixCharacters bool, words map[string]struct{}) *germanWordSplitter {
	s := &germanWordSplitter{
		words:                  words,
		hideInterfixCharacters: hideInterfixCharacters,
		exceptionSplits:        newExceptionSplits(),
		strictMode:             true,
		minimumWordLength:      3, // getDefaultMinimumWordLength
		maximumWordLength:      70,
		interfixes:             []string{"s-", "s", "-"},
	}
	s.disambiguator = newGermanInterfixDisambiguator(words)
	// setExceptionFile(EXCEPTION_DICT)
	s.exceptionSplits = loadExceptionSplitsFromEmbedded()
	return s
}

func (s *germanWordSplitter) setMinimumWordLength(n int) { s.minimumWordLength = n }
func (s *germanWordSplitter) setMaximumWordLength(n int) { s.maximumWordLength = n }
func (s *germanWordSplitter) setStrictMode(v bool)       { s.strictMode = v }

func (s *germanWordSplitter) addException(completeWord string, wordParts []string) {
	s.exceptionSplits.addSplit(javaStringToLower(completeWord), wordParts)
}

func (s *germanWordSplitter) addWord(w string) {
	if s.words == nil {
		s.words = map[string]struct{}{}
	}
	s.words[javaStringToLower(w)] = struct{}{}
	// disambiguator shares the same map reference
	if s.disambiguator != nil {
		s.disambiguator.dictionary = s.words
	}
}

// getAllSplits ports AbstractWordSplitter.getAllSplits(String).
func (s *germanWordSplitter) getAllSplits(word string) ([][]string, error) {
	if javaCharLen(word) > s.maximumWordLength {
		return nil, fmt.Errorf("%w: Input too long (%d characters), maximum is %d characters: '%s'",
			errInputTooLong, javaCharLen(word), s.maximumWordLength, word)
	}
	result1 := s.getAllSplitsFrom(word, true)
	result2 := s.getAllSplitsFrom(word, false)
	result := append([][]string(nil), result1...)
	for _, split := range result2 {
		if !containsStringSlice(result, split) {
			result = append(result, split)
		}
	}
	return result, nil
}

func (s *germanWordSplitter) getAllSplitsFrom(word string, fromLeft bool) [][]string {
	var result [][]string
	var i int
	if fromLeft {
		i = s.minimumWordLength
	} else {
		i = javaCharLen(word) - s.minimumWordLength
	}
	for s.isLoopEnd(fromLeft, i, word) {
		left := javaSubstring(word, 0, i)
		right := javaSubstring(word, i, javaCharLen(word))
		relevantWord := right
		if fromLeft {
			relevantWord = left
		}
		if s.isSimpleWord(relevantWord) {
			result = append(result, []string{left, right})
			var otherSplits [][]string
			if fromLeft {
				otherSplits, _ = s.getAllSplits(right)
			} else {
				otherSplits, _ = s.getAllSplits(left)
			}
			// Note: Java calls getAllSplits (public, both directions) recursively.
			// We already have the recursive public method; ignore too-long on subparts.
			if len(otherSplits) > 0 {
				for _, otherSplit := range otherSplits {
					sub := make([]string, 0, 1+len(otherSplit))
					if fromLeft {
						sub = append(sub, left)
						sub = append(sub, otherSplit...)
					} else {
						sub = append(sub, otherSplit...)
						sub = append(sub, right)
					}
					result = append(result, sub)
				}
			}
		}
		if fromLeft {
			i++
		} else {
			i--
		}
	}
	return result
}

func (s *germanWordSplitter) isLoopEnd(fromLeft bool, i int, word string) bool {
	if fromLeft {
		return i < javaCharLen(word)-s.minimumWordLength
	}
	return i > s.minimumWordLength
}

// splitWord ports AbstractWordSplitter.splitWord(String) / splitWord(word, false).
func (s *germanWordSplitter) splitWord(word string) ([]string, error) {
	return s.splitWordCollect(word, false)
}

func (s *germanWordSplitter) splitWordCollect(word string, collectSubwords bool) ([]string, error) {
	// Java: null → empty list. Go has no null string; empty string falls through.
	if javaCharLen(word) > s.maximumWordLength {
		return nil, fmt.Errorf("%w: Input too long (%d characters), maximum is %d characters: '%s'",
			errInputTooLong, javaCharLen(word), s.maximumWordLength, word)
	}
	trimmedWord := javaTrim(word)
	if ex := s.exceptionSplits.getExceptionSplitOrNull(trimmedWord); ex != nil {
		return ex, nil
	}
	parts, ok := s.split(trimmedWord, false, collectSubwords)
	if !ok {
		return []string{trimmedWord}, nil
	}
	disambiguatedParts := s.disambiguator.disambiguate(parts)
	cleanLeadingAndTrailingHyphens(disambiguatedParts)
	return disambiguatedParts, nil
}

// cleanLeadingAndTrailingHyphens ports AbstractWordSplitter.cleanLeadingAndTrailingHyphens.
// Note Java quirk: both start/end checks use the original element, not the post-strip value.
func cleanLeadingAndTrailingHyphens(parts []string) {
	for i, element := range parts {
		if stringsHasPrefix(element, "-") {
			parts[i] = element[1:]
		}
		if stringsHasSuffix(element, "-") {
			// Uses original element length (Java does not re-read after startsWith strip).
			parts[i] = element[:len(element)-1]
		}
	}
}

// split ports AbstractWordSplitter.split. ok=false means Java null.
func (s *germanWordSplitter) split(word string, allowInterfixRemoval, collectSubwords bool) ([]string, bool) {
	if parts := s.exceptionSplits.getExceptionSplitOrNull(word); parts != nil {
		return parts, true
	}
	lcWord := javaStringToLower(word)
	removableInterfix := s.findInterfixOrNull(lcWord)
	wordWithoutInterfix := s.removeInterfix(word, removableInterfix)
	canInterfixBeRemoved := removableInterfix != "" && allowInterfixRemoval

	var parts []string
	var ok bool

	if s.isSimpleWord(word) && !collectSubwords {
		parts = []string{word}
		ok = true
	} else if canInterfixBeRemoved && s.isSimpleWord(wordWithoutInterfix) {
		if s.hideInterfixCharacters {
			parts = []string{wordWithoutInterfix}
		} else {
			parts = []string{wordWithoutInterfix, removableInterfix}
		}
		ok = true
	} else {
		parts, ok = s.splitFromRight(word, collectSubwords)

		if !ok && s.isSimpleWord(word) {
			parts = []string{word}
			ok = true
		} else if ok && s.isSimpleWord(word) && !containsString(parts, word) {
			parts = append(parts, word)
		}

		if !ok && s.endsWithInterfix(lcWord) {
			parts, ok = s.splitFromRight(wordWithoutInterfix, collectSubwords)
			if ok && !s.hideInterfixCharacters {
				parts = append(parts, removableInterfix)
			}
		}
	}
	return parts, ok
}

func (s *germanWordSplitter) splitFromRight(word string, collectSubwords bool) ([]string, bool) {
	if parts := s.exceptionSplits.getExceptionSplitOrNull(word); parts != nil {
		return parts, true
	}

	var parts []string
	ok := false
	n := javaCharLen(word)
	for i := n - s.minimumWordLength; i >= s.minimumWordLength; i-- {
		leftPart := javaSubstring(word, 0, i)
		rightPart := javaSubstring(word, i, n)
		if !s.strictMode {
			if ex := s.getExceptionSplitOrNull(rightPart, leftPart); ex != nil {
				return ex, true
			}
		}
		if s.isSimpleWord(rightPart) {
			leftPartParts, isLeftPartAWord := s.split(leftPart, true, collectSubwords)
			if isLeftPartAWord {
				if collectSubwords {
					if !ok {
						parts = []string{}
						ok = true
					}
					for _, leftPartPart := range leftPartParts {
						if !containsString(parts, leftPartPart) {
							parts = append(parts, leftPartPart)
						}
					}
					if !containsString(parts, rightPart) {
						parts = append(parts, rightPart)
					}
					if rightPartExceptions := s.exceptionSplits.getExceptionSplitOrNull(rightPart); rightPartExceptions != nil {
						for _, exception := range rightPartExceptions {
							if !containsString(parts, exception) {
								parts = append(parts, exception)
							}
						}
					}
				} else {
					parts = append([]string(nil), leftPartParts...)
					parts = append(parts, rightPart)
					ok = true
				}
			} else if !s.strictMode {
				parts = []string{leftPart, rightPart}
				ok = true
			}
		} else if !s.strictMode {
			if s.isSimpleWord(leftPart) {
				parts = []string{leftPart, rightPart}
				ok = true
			}
		}
	}
	return parts, ok
}

func (s *germanWordSplitter) getExceptionSplitOrNull(rightPart, leftPart string) []string {
	if exceptionSplit := s.exceptionSplits.getExceptionSplitOrNull(rightPart); exceptionSplit != nil {
		parts := make([]string, 0, 1+len(exceptionSplit))
		parts = append(parts, leftPart)
		parts = append(parts, exceptionSplit...)
		return parts
	}
	if exceptionSplit2 := s.exceptionSplits.getExceptionSplitOrNull(leftPart); exceptionSplit2 != nil {
		parts := make([]string, 0, len(exceptionSplit2)+1)
		parts = append(parts, exceptionSplit2...)
		parts = append(parts, rightPart)
		return parts
	}
	return nil
}

func (s *germanWordSplitter) findInterfixOrNull(word string) string {
	// word already lowercased by caller for findInterfixOrNull in Java uses
	// both word and lcWord — findInterfixOrNull(String word) lowercases again.
	lcWord := javaStringToLower(word)
	for _, interfix := range s.interfixes {
		if stringsHasSuffix(lcWord, interfix) {
			return interfix
		}
	}
	return ""
}

func (s *germanWordSplitter) endsWithInterfix(word string) bool {
	for _, interfix := range s.interfixes {
		if stringsHasSuffix(word, interfix) {
			return true
		}
	}
	return false
}

func (s *germanWordSplitter) removeInterfix(word, interfixOrNull string) string {
	if interfixOrNull != "" {
		// Remove by rune/char length of interfix (ASCII interfixes only).
		n := javaCharLen(word)
		m := javaCharLen(interfixOrNull)
		return javaSubstring(word, 0, n-m)
	}
	return word
}

func (s *germanWordSplitter) isSimpleWord(part string) bool {
	return javaCharLen(part) >= s.minimumWordLength && mapHas(s.words, javaStringToLower(part))
}

// --- small helpers matching Java String indexing (UTF-16 code units ≈ runes for BMP) ---

func javaCharLen(s string) int {
	return len([]rune(s))
}

func javaSubstring(s string, start, end int) string {
	rs := []rune(s)
	if start < 0 {
		start = 0
	}
	if end > len(rs) {
		end = len(rs)
	}
	if start >= end {
		return ""
	}
	return string(rs[start:end])
}

func javaTrim(s string) string {
	// Java String.trim: codepoints <= U+0020
	return toolsJavaStringTrim(s)
}

// toolsJavaStringTrim local alias to avoid circular issues — use tools package.
func toolsJavaStringTrim(s string) string {
	// Inline Java trim to keep this file independent of import churn.
	start := 0
	rs := []rune(s)
	for start < len(rs) && rs[start] <= 0x20 {
		start++
	}
	end := len(rs)
	for end > start && rs[end-1] <= 0x20 {
		end--
	}
	return string(rs[start:end])
}

func mapHas(m map[string]struct{}, k string) bool {
	if m == nil {
		return false
	}
	_, ok := m[k]
	return ok
}

func containsString(ss []string, v string) bool {
	for _, s := range ss {
		if s == v {
			return true
		}
	}
	return false
}

func containsStringSlice(list [][]string, v []string) bool {
	for _, item := range list {
		if stringSliceEqual(item, v) {
			return true
		}
	}
	return false
}

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func stringsHasPrefix(s, pfx string) bool {
	return len(s) >= len(pfx) && s[:len(pfx)] == pfx
}

func stringsHasSuffix(s, sfx string) bool {
	return len(s) >= len(sfx) && s[len(s)-len(sfx):] == sfx
}
