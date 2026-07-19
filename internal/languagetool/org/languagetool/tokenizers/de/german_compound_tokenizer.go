package de

import (
	"bufio"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

// GermanCompoundTokenizer ports tokenizers.de.GermanCompoundTokenizer
// without jWordSplitter — greedy longest-match dictionary split, plus
// Java addException overrides and ExtendedGermanWordSplitter.extendedList parts.
type GermanCompoundTokenizer struct {
	// Words is a set of known compound parts (lowercase).
	Words map[string]struct{}
	// Exceptions maps full surface (lowercase) → fixed part split (lowercase).
	Exceptions map[string][]string
	// Strict requires full coverage by dictionary parts.
	Strict bool
	// MinPartLen minimum part length (default 3).
	MinPartLen int
}

func NewGermanCompoundTokenizer(strict bool) *GermanCompoundTokenizer {
	t := &GermanCompoundTokenizer{
		Words:      map[string]struct{}{},
		Exceptions: map[string][]string{},
		Strict:     strict,
		MinPartLen: 3,
	}
	t.ApplyLanguageToolExtras()
	return t
}

// ApplyLanguageToolExtras ports ExtendedGermanWordSplitter.extendedList and
// GermanCompoundTokenizer constructor addException calls.
func (t *GermanCompoundTokenizer) ApplyLanguageToolExtras() {
	if t == nil {
		return
	}
	if t.Words == nil {
		t.Words = map[string]struct{}{}
	}
	if t.Exceptions == nil {
		t.Exceptions = map[string][]string{}
	}
	for _, w := range compoundTokenizerExtendedParts {
		t.Words[strings.ToLower(w)] = struct{}{}
	}
	for k, parts := range compoundTokenizerExceptions {
		cp := append([]string(nil), parts...)
		t.Exceptions[k] = cp
		// also register parts as dictionary words so other compounds can use them
		for _, p := range parts {
			t.Words[p] = struct{}{}
		}
	}
}

// LoadWords loads dictionary lines (one word per line).
func (t *GermanCompoundTokenizer) LoadWords(r io.Reader) error {
	sc := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		w := strings.TrimSpace(sc.Text())
		if w == "" || strings.HasPrefix(w, "#") {
			continue
		}
		t.AddWord(w)
	}
	return sc.Err()
}

// LoadHunspellDic loads surface forms from a Hunspell .dic (word before '/', skip header count).
func (t *GermanCompoundTokenizer) LoadHunspellDic(r io.Reader) error {
	sc := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	first := true
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if first {
			first = false
			// optional word-count header
			if isAllDigits(line) {
				continue
			}
		}
		// strip affix flags: word/ABC
		if i := strings.IndexByte(line, '/'); i >= 0 {
			line = line[:i]
		}
		// strip morph data tab
		if i := strings.IndexByte(line, '\t'); i >= 0 {
			line = line[:i]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		t.AddWord(line)
	}
	return sc.Err()
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func (t *GermanCompoundTokenizer) AddWord(w string) {
	if t.Words == nil {
		t.Words = map[string]struct{}{}
	}
	w = strings.TrimSpace(w)
	if w == "" {
		return
	}
	t.Words[strings.ToLower(w)] = struct{}{}
}

// AddException ports WordSplitter.addException (fixed split for a surface).
func (t *GermanCompoundTokenizer) AddException(surface string, parts []string) {
	if t == nil || surface == "" || len(parts) == 0 {
		return
	}
	if t.Exceptions == nil {
		t.Exceptions = map[string][]string{}
	}
	key := strings.ToLower(surface)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.ToLower(strings.TrimSpace(p))
		if p == "" {
			continue
		}
		out = append(out, p)
		t.AddWord(p)
	}
	if len(out) > 0 {
		t.Exceptions[key] = out
	}
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
	low := strings.ToLower(text)
	// exceptions first (Java WordSplitter)
	if t != nil && t.Exceptions != nil {
		if parts, ok := t.Exceptions[low]; ok && len(parts) > 0 {
			return restoreCapitalization(text, parts)
		}
	}
	parts := t.split(low)
	if len(parts) <= 1 {
		return []string{text}
	}
	return restoreCapitalization(text, parts)
}

// maxAllSplits caps AllSplits enumeration (jWordSplitter InputTooLong / explosion guard).
const maxAllSplits = 64

// AllSplits ports jWordSplitter GermanWordSplitter.getAllSplits for a dictionary-
// based tokenizer: every complete partition of the surface into lexicon parts.
// Exceptions yield a single fixed split. Short words and no-split return empty
// (caller treats as no compound candidates), matching Java empty list on failure.
func (t *GermanCompoundTokenizer) AllSplits(text string) [][]string {
	if t == nil || text == "" {
		return nil
	}
	if utf8.RuneCountInString(text) < 6 {
		return nil
	}
	low := strings.ToLower(text)
	if t.Exceptions != nil {
		if parts, ok := t.Exceptions[low]; ok && len(parts) > 1 {
			return [][]string{restoreCapitalization(text, parts)}
		}
	}
	raw := t.allSplits(low)
	if len(raw) == 0 {
		return nil
	}
	out := make([][]string, 0, len(raw))
	for _, parts := range raw {
		if len(parts) <= 1 {
			continue
		}
		out = append(out, restoreCapitalization(text, parts))
	}
	return out
}

// allSplits enumerates complete dictionary partitions of word (already lowercase).
func (t *GermanCompoundTokenizer) allSplits(word string) [][]string {
	minL := t.MinPartLen
	if minL <= 0 {
		minL = 3
	}
	n := len(word)
	// memo[i] = all ways to split word[i:]
	memo := make([][][]string, n+1)
	memo[n] = [][]string{{}} // one empty suffix split
	var dfs func(i int) [][]string
	dfs = func(i int) [][]string {
		if memo[i] != nil {
			return memo[i]
		}
		var res [][]string
		for j := i + minL; j <= n; j++ {
			part := word[i:j]
			if _, ok := t.Words[part]; !ok {
				continue
			}
			for _, rest := range dfs(j) {
				if len(res) >= maxAllSplits {
					break
				}
				cp := make([]string, 0, 1+len(rest))
				cp = append(cp, part)
				cp = append(cp, rest...)
				res = append(res, cp)
			}
			if len(res) >= maxAllSplits {
				break
			}
		}
		if res == nil {
			res = [][]string{} // mark computed empty
		}
		memo[i] = res
		return res
	}
	return dfs(0)
}

func restoreCapitalization(original string, parts []string) []string {
	out := append([]string(nil), parts...)
	if len(out) == 0 {
		return []string{original}
	}
	if unicode.IsUpper([]rune(original)[0]) {
		rs := []rune(out[0])
		if len(rs) > 0 {
			rs[0] = unicode.ToUpper(rs[0])
			out[0] = string(rs)
		}
	}
	return out
}

func (t *GermanCompoundTokenizer) split(word string) []string {
	minL := t.MinPartLen
	if minL <= 0 {
		minL = 3
	}
	// DP: best split ending at i — prefer longer coverage; among equals prefer fewer parts via first-found
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
		// try longer parts first for greedy quality closer to jWordSplitter
		for j := n; j >= i+minL; j-- {
			part := word[i:j]
			if _, ok := t.Words[part]; ok {
				if !best[j].ok {
					best[j] = node{prev: i, ok: true}
				}
			}
		}
	}
	if !best[n].ok {
		return []string{word}
	}
	// reconstruct
	var rev []string
	for cur := n; cur > 0; {
		p := best[cur].prev
		rev = append(rev, word[p:cur])
		cur = p
	}
	for i, j := 0, len(rev)-1; i < j; i, j = i+1, j-1 {
		rev[i], rev[j] = rev[j], rev[i]
	}
	return rev
}
