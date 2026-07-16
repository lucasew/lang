package rules

import (
	"sort"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// BERTSuggestionRanking ports org.languagetool.rules.BERTSuggestionRanking
// with a pluggable scorer (remote BERT deferred).
type BERTSuggestionRanking struct {
	// Scorer ranks (suggestion, score); higher is better. If nil, identity order.
	Scorer func(word string, suggestions []string, sentence *languagetool.AnalyzedSentence, startPos int) []scoredSuggestion
	TopK   int
	mu     sync.Mutex
	cache  map[string][]string
}

type scoredSuggestion struct {
	text  string
	score float64
}

func NewBERTSuggestionRanking() *BERTSuggestionRanking {
	return &BERTSuggestionRanking{
		TopK:  10,
		cache: map[string][]string{},
	}
}

// RankSuggestions reorders suggestions using Scorer or leaves them unchanged.
func (b *BERTSuggestionRanking) RankSuggestions(word string, suggestions []string, sentence *languagetool.AnalyzedSentence, startPos int) []string {
	if b == nil || len(suggestions) == 0 {
		return suggestions
	}
	key := word + "\x00" + joinSugs(suggestions)
	b.mu.Lock()
	if cached, ok := b.cache[key]; ok {
		b.mu.Unlock()
		return append([]string(nil), cached...)
	}
	b.mu.Unlock()

	var out []string
	if b.Scorer != nil {
		scored := b.Scorer(word, suggestions, sentence, startPos)
		sort.SliceStable(scored, func(i, j int) bool {
			return scored[i].score > scored[j].score
		})
		topK := b.TopK
		if topK <= 0 || topK > len(scored) {
			topK = len(scored)
		}
		out = make([]string, 0, topK)
		for i := 0; i < topK; i++ {
			out = append(out, scored[i].text)
		}
	} else {
		out = append([]string(nil), suggestions...)
		if b.TopK > 0 && len(out) > b.TopK {
			out = out[:b.TopK]
		}
	}

	b.mu.Lock()
	b.cache[key] = append([]string(nil), out...)
	b.mu.Unlock()
	return out
}

// RankSuggestedReplacements rewrites SuggestedReplacement order.
func (b *BERTSuggestionRanking) RankSuggestedReplacements(word string, reps []*SuggestedReplacement, sentence *languagetool.AnalyzedSentence, startPos int) []*SuggestedReplacement {
	if len(reps) == 0 {
		return reps
	}
	texts := make([]string, 0, len(reps))
	byText := map[string]*SuggestedReplacement{}
	for _, r := range reps {
		if r == nil {
			continue
		}
		texts = append(texts, r.Replacement)
		byText[r.Replacement] = r
	}
	ordered := b.RankSuggestions(word, texts, sentence, startPos)
	out := make([]*SuggestedReplacement, 0, len(ordered))
	for _, t := range ordered {
		if r, ok := byText[t]; ok {
			out = append(out, r)
		}
	}
	return out
}

// EditDistanceBERTScorer ranks by inverse edit distance (local stand-in for BERT).
func EditDistanceBERTScorer(word string, suggestions []string, _ *languagetool.AnalyzedSentence, _ int) []scoredSuggestion {
	out := make([]scoredSuggestion, 0, len(suggestions))
	for _, s := range suggestions {
		d := levenshtein(word, s)
		score := 1.0 / float64(1+d)
		out = append(out, scoredSuggestion{text: s, score: score})
	}
	return out
}

func joinSugs(s []string) string {
	if len(s) == 0 {
		return ""
	}
	n := 0
	for _, x := range s {
		n += len(x) + 1
	}
	b := make([]byte, 0, n)
	for i, x := range s {
		if i > 0 {
			b = append(b, '|')
		}
		b = append(b, x...)
	}
	return string(b)
}

func levenshtein(a, b string) int {
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
