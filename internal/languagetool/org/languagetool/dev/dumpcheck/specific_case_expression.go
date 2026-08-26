package dumpcheck

import (
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// SpecificCaseCounter collects multi-word expressions whose tokens all start uppercase
// (WikipediaSpecificCaseExpressionExtractor green slice).
type SpecificCaseCounter struct {
	counts map[string]int
}

func NewSpecificCaseCounter() *SpecificCaseCounter {
	return &SpecificCaseCounter{counts: map[string]int{}}
}

var multiWordSplit = regexp.MustCompile(`\s+`)

// ObserveSentence extracts capitalized multi-word runs and increments counters.
// A run is 2+ consecutive tokens that each start with an uppercase letter.
func (c *SpecificCaseCounter) ObserveSentence(sentence string) {
	tokens := multiWordSplit.Split(strings.TrimSpace(sentence), -1)
	var run []string
	flush := func() {
		if len(run) >= 2 {
			expr := strings.Join(run, " ")
			c.counts[expr]++
		}
		run = run[:0]
	}
	for _, tok := range tokens {
		tok = strings.Trim(tok, ".,;:!?\"'()[]")
		if tok == "" {
			flush()
			continue
		}
		r := []rune(tok)[0]
		if unicode.IsUpper(r) {
			run = append(run, tok)
		} else {
			flush()
		}
	}
	flush()
}

// ObserveSource drains a SentenceSource.
func (c *SpecificCaseCounter) ObserveSource(source SentenceSource) error {
	for source.HasNext() {
		s, err := source.Next()
		if err != nil {
			return err
		}
		c.ObserveSentence(s.GetText())
	}
	return nil
}

// Top returns the n most frequent expressions (stable by count desc, then name).
func (c *SpecificCaseCounter) Top(n int) []string {
	type pair struct {
		k string
		v int
	}
	var all []pair
	for k, v := range c.counts {
		all = append(all, pair{k, v})
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].v != all[j].v {
			return all[i].v > all[j].v
		}
		return all[i].k < all[j].k
	})
	if n > len(all) {
		n = len(all)
	}
	out := make([]string, n)
	for i := 0; i < n; i++ {
		out[i] = all[i].k
	}
	return out
}

func (c *SpecificCaseCounter) Count(expr string) int { return c.counts[expr] }
