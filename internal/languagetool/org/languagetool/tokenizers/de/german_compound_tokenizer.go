package de

import (
	"bufio"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

// GermanCompoundTokenizer ports tokenizers.de.GermanCompoundTokenizer
// without jWordSplitter — greedy longest-match dictionary split.
type GermanCompoundTokenizer struct {
	// Words is a set of known compound parts (lowercase).
	Words map[string]struct{}
	// Strict requires full coverage by dictionary parts.
	Strict bool
	// MinPartLen minimum part length (default 3).
	MinPartLen int
}

func NewGermanCompoundTokenizer(strict bool) *GermanCompoundTokenizer {
	return &GermanCompoundTokenizer{
		Words:      map[string]struct{}{},
		Strict:     strict,
		MinPartLen: 3,
	}
}

// LoadWords loads dictionary lines (one word per line).
func (t *GermanCompoundTokenizer) LoadWords(r io.Reader) error {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		w := strings.TrimSpace(sc.Text())
		if w == "" || strings.HasPrefix(w, "#") {
			continue
		}
		t.Words[strings.ToLower(w)] = struct{}{}
	}
	return sc.Err()
}

func (t *GermanCompoundTokenizer) AddWord(w string) {
	if t.Words == nil {
		t.Words = map[string]struct{}{}
	}
	t.Words[strings.ToLower(w)] = struct{}{}
}

// Tokenize splits a German noun compound into parts when possible.
func (t *GermanCompoundTokenizer) Tokenize(text string) []string {
	if text == "" {
		return nil
	}
	// skip short / non-letter
	if utf8.RuneCountInString(text) < 6 {
		return []string{text}
	}
	parts := t.split(strings.ToLower(text))
	if len(parts) <= 1 {
		return []string{text}
	}
	// restore capitalization of first letter if original was capitalized
	if unicode.IsUpper([]rune(text)[0]) {
		rs := []rune(parts[0])
		rs[0] = unicode.ToUpper(rs[0])
		parts[0] = string(rs)
	}
	return parts
}

func (t *GermanCompoundTokenizer) split(word string) []string {
	minL := t.MinPartLen
	if minL <= 0 {
		minL = 3
	}
	// DP: best split ending at i
	type node struct {
		prev int
		ok   bool
	}
	n := len(word)
	best := make([]node, n+1)
	best[0] = node{prev: -1, ok: true}
	for i := 0; i < n; i++ {
		if !best[i].ok {
			continue
		}
		// try ends j
		for j := i + minL; j <= n; j++ {
			part := word[i:j]
			if _, ok := t.Words[part]; ok {
				if !best[j].ok {
					best[j] = node{prev: i, ok: true}
				}
			}
		}
	}
	if !best[n].ok {
		if t.Strict {
			return []string{word}
		}
		return []string{word}
	}
	// reconstruct
	var rev []string
	for cur := n; cur > 0; {
		p := best[cur].prev
		rev = append(rev, word[p:cur])
		cur = p
	}
	// reverse
	for i, j := 0, len(rev)-1; i < j; i, j = i+1, j-1 {
		rev[i], rev[j] = rev[j], rev[i]
	}
	return rev
}
