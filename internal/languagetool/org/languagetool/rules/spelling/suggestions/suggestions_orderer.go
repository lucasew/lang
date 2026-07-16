package suggestions

import (
	"sort"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// SuggestionsOrderer ports org.languagetool.rules.spelling.suggestions.SuggestionsOrderer.
type SuggestionsOrderer interface {
	IsMlAvailable() bool
	OrderSuggestions(suggestions []string, word string, sentence *languagetool.AnalyzedSentence, startPos int) []*rules.SuggestedReplacement
}

// OrderSuggestionsUsingModel maps ordered replacements back to strings.
func OrderSuggestionsUsingModel(o SuggestionsOrderer, suggestions []string, word string, sentence *languagetool.AnalyzedSentence, startPos int) []string {
	if o == nil {
		return append([]string(nil), suggestions...)
	}
	ordered := o.OrderSuggestions(suggestions, word, sentence, startPos)
	out := make([]string, 0, len(ordered))
	for _, s := range ordered {
		if s != nil {
			out = append(out, s.Replacement)
		}
	}
	return out
}

// IdentitySuggestionsOrderer preserves input order (no ML).
type IdentitySuggestionsOrderer struct{}

func (IdentitySuggestionsOrderer) IsMlAvailable() bool { return false }

func (IdentitySuggestionsOrderer) OrderSuggestions(suggestions []string, word string, sentence *languagetool.AnalyzedSentence, startPos int) []*rules.SuggestedReplacement {
	out := make([]*rules.SuggestedReplacement, 0, len(suggestions))
	for _, s := range suggestions {
		out = append(out, rules.NewSuggestedReplacement(s))
	}
	return out
}

// EditDistanceSuggestionsOrderer ranks by Levenshtein distance then alphabetically.
type EditDistanceSuggestionsOrderer struct{}

func (EditDistanceSuggestionsOrderer) IsMlAvailable() bool { return false }

func (EditDistanceSuggestionsOrderer) OrderSuggestions(suggestions []string, word string, sentence *languagetool.AnalyzedSentence, startPos int) []*rules.SuggestedReplacement {
	type item struct {
		s    string
		dist int
	}
	items := make([]item, 0, len(suggestions))
	for _, s := range suggestions {
		items = append(items, item{s: s, dist: lev(word, s)})
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].dist != items[j].dist {
			return items[i].dist < items[j].dist
		}
		return strings.ToLower(items[i].s) < strings.ToLower(items[j].s)
	})
	out := make([]*rules.SuggestedReplacement, 0, len(items))
	for _, it := range items {
		out = append(out, rules.NewSuggestedReplacement(it.s))
	}
	return out
}

func lev(a, b string) int {
	ar, br := []rune(a), []rune(b)
	if len(ar) == 0 {
		return len(br)
	}
	if len(br) == 0 {
		return len(ar)
	}
	prev := make([]int, len(br)+1)
	cur := make([]int, len(br)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ar); i++ {
		cur[0] = i
		for j := 1; j <= len(br); j++ {
			cost := 1
			if ar[i-1] == br[j-1] {
				cost = 0
			}
			del := prev[j] + 1
			ins := cur[j-1] + 1
			sub := prev[j-1] + cost
			m := del
			if ins < m {
				m = ins
			}
			if sub < m {
				m = sub
			}
			cur[j] = m
		}
		prev, cur = cur, prev
	}
	return prev[len(br)]
}

var (
	_ SuggestionsOrderer = IdentitySuggestionsOrderer{}
	_ SuggestionsOrderer = EditDistanceSuggestionsOrderer{}
)
