package de

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// exceptionSplits ports de.danielnaber.jwordsplitter.ExceptionSplits.
type exceptionSplits struct {
	exceptionMap map[string][]string
}

func newExceptionSplits() *exceptionSplits {
	return &exceptionSplits{exceptionMap: map[string][]string{}}
}

// loadFromString parses exceptionsGerman.txt contents (UTF-8, one rule per line).
func (e *exceptionSplits) loadFromString(content string) {
	if e.exceptionMap == nil {
		e.exceptionMap = map[string][]string{}
	}
	for _, raw := range strings.Split(content, "\n") {
		line := tools.JavaStringTrim(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Java: String[] parts = line.replace("/NS", "").split("\\|");
		stripped := strings.ReplaceAll(line, "/NS", "")
		parts := strings.Split(stripped, "|")
		// Java: String completeWord = line.replace("|", "");
		completeWord := strings.ReplaceAll(line, "|", "")
		list := append([]string(nil), parts...)
		if strings.Contains(completeWord, "/") {
			if strings.HasSuffix(completeWord, "/NS") {
				realWord := javaStringToLower(strings.ReplaceAll(completeWord, "/NS", ""))
				e.exceptionMap[realWord] = list
				e.exceptionMap[realWord+"n"] = addToLastPart(list, "n")
				e.exceptionMap[realWord+"s"] = addToLastPart(list, "s")
			} else {
				panic("jwordsplitter: unknown suffix in exception line: " + line)
			}
		} else {
			e.exceptionMap[javaStringToLower(completeWord)] = list
		}
	}
}

func addToLastPart(list []string, suffix string) []string {
	result := append([]string(nil), list...)
	if len(result) == 0 {
		return result
	}
	last := len(result) - 1
	result[last] = result[last] + suffix
	return result
}

func (e *exceptionSplits) getExceptionSplitOrNull(word string) []string {
	if e == nil || e.exceptionMap == nil {
		return nil
	}
	lcWord := javaStringToLower(word)
	result, ok := e.exceptionMap[lcWord]
	if !ok || result == nil {
		return nil
	}
	// Preserve case of original word when recombined lowercase matches.
	check := strings.Join(result, "")
	if lcWord == javaStringToLower(check) {
		return splitEqually(result, word)
	}
	// Java returns the stored list as-is when check doesn't match (rare).
	return append([]string(nil), result...)
}

func splitEqually(splitted []string, original string) []string {
	// Index by Java char (= rune for BMP German text).
	rs := []rune(original)
	list := make([]string, 0, len(splitted))
	offset := 0
	for _, s := range splitted {
		length := len([]rune(s))
		if offset+length > len(rs) {
			// Defensive: fall back to stored pieces.
			return append([]string(nil), splitted...)
		}
		list = append(list, string(rs[offset:offset+length]))
		offset += length
	}
	return list
}

func (e *exceptionSplits) addSplit(word string, wordParts []string) {
	if e.exceptionMap == nil {
		e.exceptionMap = map[string][]string{}
	}
	cp := append([]string(nil), wordParts...)
	e.exceptionMap[javaStringToLower(word)] = cp
}
