package bigdata

import (
	"bufio"
	"io"
	"strings"
	"unicode"
)

const MaxTokenLength = 20

// NgramCounts holds unigram/bigram/trigram frequency maps.
type NgramCounts struct {
	Unigram map[string]int64
	Bigram  map[string]int64
	Trigram map[string]int64
}

func NewNgramCounts() *NgramCounts {
	return &NgramCounts{
		Unigram: map[string]int64{},
		Bigram:  map[string]int64{},
		Trigram: map[string]int64{},
	}
}

// IndexSentence ports CommonCrawlToNgram.indexSentence (count only, no Lucene).
// Adds _START_ / _END_ sentinels like Google ngram format.
func (c *NgramCounts) IndexSentence(sentence string, tokenize func(string) []string) {
	if c == nil {
		return
	}
	if tokenize == nil {
		tokenize = simpleWordTokenize
	}
	tokens := tokenize(sentence)
	// prepend/append sentinels
	toks := make([]string, 0, len(tokens)+2)
	toks = append(toks, GoogleSentenceStart)
	for _, t := range tokens {
		if strings.TrimSpace(t) == "" {
			continue
		}
		toks = append(toks, t)
	}
	toks = append(toks, GoogleSentenceEnd)

	var prevPrev, prev string
	for _, token := range toks {
		if token == "" {
			continue
		}
		if len([]rune(token)) <= MaxTokenLength {
			c.Unigram[token]++
		}
		if prev != "" {
			if len([]rune(token)) <= MaxTokenLength && len([]rune(prev)) <= MaxTokenLength {
				c.Bigram[prev+" "+token]++
			}
		}
		if prevPrev != "" && prev != "" {
			if len([]rune(token)) <= MaxTokenLength &&
				len([]rune(prev)) <= MaxTokenLength &&
				len([]rune(prevPrev)) <= MaxTokenLength {
				c.Trigram[prevPrev+" "+prev+" "+token]++
			}
		}
		prevPrev = prev
		prev = token
	}
}

// IndexLines tokenizes each line as a sentence (or splits on .!? lightly).
func (c *NgramCounts) IndexLines(r io.Reader, tokenize func(string) []string) error {
	sc := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 4*1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		// green: whole line as one sentence
		c.IndexSentence(line, tokenize)
	}
	return sc.Err()
}

func simpleWordTokenize(s string) []string {
	var out []string
	var cur strings.Builder
	flush := func() {
		if cur.Len() > 0 {
			out = append(out, cur.String())
			cur.Reset()
		}
	}
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '\'' || r == '-' {
			cur.WriteRune(r)
		} else {
			flush()
			if !unicode.IsSpace(r) {
				out = append(out, string(r))
			}
		}
	}
	flush()
	return out
}
