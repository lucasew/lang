package de

import (
	"bufio"
	"io"
	"os"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

// SpellingAdjExpansion ports GermanTagger ExpansionInfos.adjInfos built from
// de/hunspell/spelling.txt lines ending /A or /P (adjective / participle endings).
// Implements tagging.WordTagger for CombiningTagger / GermanTagger.
type SpellingAdjExpansion struct {
	// fullform → readings (lemma = base word without ending suffix)
	byForm map[string][]tagging.TaggedWord
}

// LoadSpellingAdjExpansionFromFile loads /A and /P expansions from a spelling extras file.
func LoadSpellingAdjExpansionFromFile(path string) (*SpellingAdjExpansion, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return LoadSpellingAdjExpansion(f)
}

// LoadSpellingAdjExpansion ports initExpansionInfos adjective branch from a reader.
func LoadSpellingAdjExpansion(r io.Reader) (*SpellingAdjExpansion, error) {
	ex := &SpellingAdjExpansion{byForm: map[string][]tagging.TaggedWord{}}
	sc := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
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
		// Java rejects /PA or /AP combo with RuntimeException — skip invalid lines
		if strings.HasSuffix(line, "/PA") || strings.HasSuffix(line, "/AP") {
			continue
		}
		asPA2 := false
		switch {
		case strings.HasSuffix(line, "/P"):
			asPA2 = true
		case strings.HasSuffix(line, "/A"):
			// skip ste/A and er/A (Java: comparative/superlative would miss tagging)
			if strings.HasSuffix(line, "ste/A") || strings.HasSuffix(line, "er/A") {
				continue
			}
		default:
			continue
		}
		// Java: word = line.replaceFirst("/.*", "") — strip from first '/'
		base := line
		if i := strings.IndexByte(line, '/'); i >= 0 {
			base = line[:i]
		}
		base = strings.TrimSpace(base)
		if base == "" {
			continue
		}
		ex.addWordEndings(base, asPA2)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return ex, nil
}

func (e *SpellingAdjExpansion) addWordEndings(word string, asPA2 bool) {
	type pair struct {
		suffix string
		tags   []string
	}
	pairs := []pair{
		{"", TagsForAdj},
		{"e", TagsForAdjE},
		{"en", TagsForAdjEn},
		{"er", TagsForAdjEr},
		{"em", TagsForAdjEm},
		{"es", TagsForAdjEs},
	}
	for _, p := range pairs {
		tags := p.tags
		if asPA2 {
			tags = ToPA2(tags)
		}
		full := word + p.suffix
		// Java fillAdjInfos: adjInfos.put(fullform, l) — replace, do not append
		list := make([]tagging.TaggedWord, 0, len(tags))
		for _, tag := range tags {
			list = append(list, tagging.NewTaggedWord(word, tag))
		}
		e.byForm[full] = list
	}
}

// Tag implements tagging.WordTagger.
func (e *SpellingAdjExpansion) Tag(word string) []tagging.TaggedWord {
	if e == nil || len(e.byForm) == 0 {
		return nil
	}
	return append([]tagging.TaggedWord(nil), e.byForm[word]...)
}

// Size returns number of expanded fullforms.
func (e *SpellingAdjExpansion) Size() int {
	if e == nil {
		return 0
	}
	return len(e.byForm)
}

var _ tagging.WordTagger = (*SpellingAdjExpansion)(nil)
