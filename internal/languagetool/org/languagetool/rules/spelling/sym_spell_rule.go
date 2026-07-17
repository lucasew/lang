package spelling

import (
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// SymSpellRule ports org.languagetool.rules.spelling.SymSpellRule as a
// dictionary + edit-distance suggestion rule (full SymSpell index deferred).
type SymSpellRule struct {
	*SpellingCheckRule
	EditDistance int
	// Dictionary accepted words.
	Dictionary map[string]struct{}
	// Prohibited words always flagged.
	Prohibited map[string]struct{}
	mu         sync.RWMutex
}

func NewSymSpellRule(id, languageCode string) *SymSpellRule {
	r := &SymSpellRule{
		SpellingCheckRule: NewSpellingCheckRule(id, "Possible spelling mistake (SymSpell)", languageCode),
		EditDistance:      3,
		Dictionary:        map[string]struct{}{},
		Prohibited:        map[string]struct{}{},
	}
	r.IsMisspelled = r.isMisspelled
	return r
}

func (r *SymSpellRule) AddWords(words ...string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, w := range words {
		r.Dictionary[w] = struct{}{}
	}
}

func (r *SymSpellRule) isMisspelled(word string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if _, bad := r.Prohibited[word]; bad {
		return true
	}
	if _, ok := r.Dictionary[word]; ok {
		return false
	}
	low := strings.ToLower(word)
	if low != word {
		if _, ok := r.Dictionary[low]; ok {
			return false
		}
	}
	return true
}

// Suggestions returns dictionary words within EditDistance (small dicts).
func (r *SymSpellRule) Suggestions(word string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.Dictionary) == 0 || len(r.Dictionary) > 10000 {
		return nil
	}
	var out []string
	for w := range r.Dictionary {
		if levenshtein(word, w) <= r.EditDistance {
			out = append(out, w)
			if len(out) >= 10 {
				break
			}
		}
	}
	return out
}

// Match flags misspelled tokens.
func (r *SymSpellRule) Match(sentence *languagetool.AnalyzedSentence) ([]*rules.RuleMatch, error) {
	if sentence == nil || r == nil {
		return nil, nil
	}
	var out []*rules.RuleMatch
	for _, tok := range sentence.GetTokensWithoutWhitespace() {
		if tok == nil || tok.IsSentenceStart() || tok.IsSentenceEnd() {
			continue
		}
		if tok.IsIgnoredBySpeller() || tok.IsImmunized() {
			continue
		}
		w := tok.GetToken()
		if w == "" || utf8.RuneCountInString(w) > MaxTokenLength {
			continue
		}
		if r.AcceptWord(w) {
			continue
		}
		m := rules.NewRuleMatch(r, sentence, tok.GetStartPos(), tok.GetEndPos(),
			"Possible spelling mistake found")
		if sug := r.Suggestions(w); len(sug) > 0 {
			m.SetSuggestedReplacements(sug)
		}
		out = append(out, m)
	}
	return out, nil
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
