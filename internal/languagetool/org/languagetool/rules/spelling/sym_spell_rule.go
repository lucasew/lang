package spelling

import (
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/suggestions"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// SymSpellRuleID ports SymSpellRule.getId.
const SymSpellRuleID = "SYMSPELL_RULE"

// SymSpellRule ports org.languagetool.rules.spelling.SymSpellRule as a
// dictionary + edit-distance suggestion rule (full SymSpell index deferred;
// lookup uses in-memory Dictionary + edit distance like the Java SuggestItem path).
type SymSpellRule struct {
	*SpellingCheckRule
	EditDistance int
	// Dictionary accepted words (Java defaultDictSpeller surface).
	Dictionary map[string]struct{}
	// UserDictionary ports userDictSpeller accepted words.
	UserDictionary map[string]struct{}
	// Prohibited words always flagged (filterCandidates).
	Prohibited map[string]struct{}
	// Orderer ports SymSpellRule.orderer for addSuggestionsToRuleMatch (optional ML).
	Orderer suggestions.SuggestionsOrderer
	mu      sync.RWMutex
}

func NewSymSpellRule(id, languageCode string) *SymSpellRule {
	if id == "" {
		id = SymSpellRuleID
	}
	r := &SymSpellRule{
		SpellingCheckRule: NewSpellingCheckRule(id, "Spell checking rule using SymSpell algorithm", languageCode),
		EditDistance:      3,
		Dictionary:        map[string]struct{}{},
		UserDictionary:    map[string]struct{}{},
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

// AddUserWords ports user dictionary lines for userDictSpeller.
func (r *SymSpellRule) AddUserWords(words ...string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.UserDictionary == nil {
		r.UserDictionary = map[string]struct{}{}
	}
	for _, w := range words {
		r.UserDictionary[w] = struct{}{}
	}
}

func (r *SymSpellRule) isMisspelled(word string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if _, bad := r.Prohibited[word]; bad {
		return true
	}
	if r.inDictLocked(word) {
		return false
	}
	return true
}

// inDictLocked requires r.mu held for read.
func (r *SymSpellRule) inDictLocked(word string) bool {
	if _, ok := r.Dictionary[word]; ok {
		return true
	}
	if _, ok := r.UserDictionary[word]; ok {
		return true
	}
	low := strings.ToLower(word)
	if low != word {
		if _, ok := r.Dictionary[low]; ok {
			return true
		}
		if _, ok := r.UserDictionary[low]; ok {
			return true
		}
	}
	return false
}

// Suggestions returns default-dict words within EditDistance (small dicts).
func (r *SymSpellRule) Suggestions(word string) []string {
	return r.suggestionsFrom(word, false)
}

func (r *SymSpellRule) suggestionsFrom(word string, user bool) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	dict := r.Dictionary
	if user {
		dict = r.UserDictionary
	}
	if len(dict) == 0 || len(dict) > 10000 {
		return nil
	}
	var out []string
	for w := range dict {
		// filterCandidates: skip prohibited
		if _, bad := r.Prohibited[w]; bad {
			continue
		}
		if levenshtein(word, w) <= r.EditDistance {
			out = append(out, w)
			if len(out) >= 10 {
				break
			}
		}
	}
	return out
}

// Match ports SymSpellRule.match:
// skip sentence start / immunized / ignored / non-word;
// empty candidates → misspelling match;
// top candidate equals word → no match;
// else match + addSuggestionsToRuleMatch(user, default, orderer).
func (r *SymSpellRule) Match(sentence *languagetool.AnalyzedSentence) ([]*rules.RuleMatch, error) {
	if sentence == nil || r == nil {
		return nil, nil
	}
	var out []*rules.RuleMatch
	for _, tok := range sentence.GetTokensWithoutWhitespace() {
		// Java: token.isSentenceStart() || immunized || ignoredBySpeller || isNonWord
		if tok == nil || tok.IsSentenceStart() {
			continue
		}
		if tok.IsIgnoredBySpeller() || tok.IsImmunized() || tok.IsNonWord() {
			continue
		}
		w := tok.GetToken()
		if w == "" || tokenizers.UTF16Len(w) > MaxTokenLength {
			continue
		}
		// IgnoreWord list (Java ignoredWords.contains)
		if r.SpellingCheckRule != nil && r.IgnoreWord(w) {
			continue
		}
		candidates := r.suggestionsFrom(w, false)
		userCandidates := r.suggestionsFrom(w, true)
		// Java: candidates empty && user empty → match "Misspelling or unknown word!"
		// else if top candidate != word → match + suggestions
		var m *rules.RuleMatch
		if len(candidates) == 0 && len(userCandidates) == 0 {
			if !r.isMisspelled(w) {
				// in dict but no edit-distance sugs from small dict scan — accept
				continue
			}
			m = rules.NewRuleMatch(r, sentence, tok.GetStartPos(), tok.GetEndPos(),
				"Misspelling or unknown word!")
			m.SetType(rules.RuleMatchTypeUnknownWord)
		} else if !(len(candidates) > 0 && candidates[0] == w ||
			len(userCandidates) > 0 && userCandidates[0] == w) {
			m = rules.NewRuleMatch(r, sentence, tok.GetStartPos(), tok.GetEndPos(),
				"Misspelling!")
			m.SetType(rules.RuleMatchTypeUnknownWord)
			AddSuggestionsToRuleMatchStrings(w, userCandidates, candidates, r.Orderer, m)
		}
		if m != nil {
			out = append(out, m)
		}
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
