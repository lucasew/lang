package de

import (
	"bufio"
	"embed"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// Official jWordSplitter 4.5 resources (see jwordsplitter/SOURCE.txt).
//
//go:embed jwordsplitter/wordsGerman.txt jwordsplitter/exceptionsGerman.txt
var jwordsplitterFS embed.FS

var (
	embeddedWordsOnce sync.Once
	embeddedWords     map[string]struct{}
)

// embeddedGermanDictionaryWords ports EmbeddedGermanDictionary.getWords():
// the full official jWordSplitter word set (immutable after load; callers must copy).
func embeddedGermanDictionaryWords() map[string]struct{} {
	embeddedWordsOnce.Do(func() {
		f, err := jwordsplitterFS.Open("jwordsplitter/wordsGerman.txt")
		if err != nil {
			panic("jwordsplitter: cannot open wordsGerman.txt: " + err.Error())
		}
		defer f.Close()
		m := make(map[string]struct{}, 65536)
		sc := bufio.NewScanner(f)
		buf := make([]byte, 0, 64*1024)
		sc.Buffer(buf, 1024*1024)
		for sc.Scan() {
			w := tools.JavaStringTrim(sc.Text())
			if w == "" {
				continue
			}
			m[w] = struct{}{}
		}
		if err := sc.Err(); err != nil {
			panic("jwordsplitter: reading wordsGerman.txt: " + err.Error())
		}
		embeddedWords = m
	})
	return embeddedWords
}

// copyWordSet returns a mutable copy of the embedded dictionary.
func copyWordSet(src map[string]struct{}) map[string]struct{} {
	out := make(map[string]struct{}, len(src)+128)
	for w := range src {
		out[w] = struct{}{}
	}
	return out
}

// loadExceptionSplitsFromEmbedded ports ExceptionSplits(filename) with the
// classpath resource /de/danielnaber/jwordsplitter/exceptionsGerman.txt.
func loadExceptionSplitsFromEmbedded() *exceptionSplits {
	es := newExceptionSplits()
	data, err := jwordsplitterFS.ReadFile("jwordsplitter/exceptionsGerman.txt")
	if err != nil {
		panic("jwordsplitter: cannot read exceptionsGerman.txt: " + err.Error())
	}
	es.loadFromString(string(data))
	return es
}

// javaStringToLower mirrors Java String.toLowerCase() for German compound parts
// (BMP; default locale case is fine for this dictionary).
func javaStringToLower(s string) string {
	return strings.ToLower(s)
}
