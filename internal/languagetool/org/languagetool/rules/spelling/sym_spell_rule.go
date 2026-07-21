package spelling

import (
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/suggestions"
	symimpl "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/symspell/implementation"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// SymSpellRuleID ports SymSpellRule.getId.
const SymSpellRuleID = "SYMSPELL_RULE"

// SymSpellRule ports org.languagetool.rules.spelling.SymSpellRule.
// When Speller is set, getSpellerMatches uses SymSpell.lookup (Java path).
// When Speller is nil, falls back to map Dictionary + edit-distance scan (tests).
type SymSpellRule struct {
	*SpellingCheckRule
	// EditDistance ports default editDistance (3 in Java initParameters unless experiment).
	EditDistance int
	// Verbosity ports SymSpell.Verbosity for lookup (default Closest like Java default
	// when experiment not set — Java uses Top in some experiment paths; Closest is safer default for tests).
	// Java match uses getSpellerMatches → lookup(word, verbosity, editDistance).
	Verbosity symimpl.Verbosity
	// Speller is the defaultDictSpeller (Java).
	Speller *symimpl.SymSpell
	// UserSpeller is the userDictSpeller (Java); nil → no user candidates.
	UserSpeller *symimpl.SymSpell
	// Dictionary accepted words — used when Speller is nil (map-inject tests).
	Dictionary map[string]struct{}
	// UserDictionary map-inject user words when UserSpeller is nil.
	UserDictionary map[string]struct{}
	// Prohibited words always filtered from candidates (filterCandidates).
	Prohibited map[string]struct{}
	// Orderer ports SymSpellRule.orderer for addSuggestionsToRuleMatch.
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
		Verbosity:         symimpl.VerbosityClosest,
		Dictionary:        map[string]struct{}{},
		UserDictionary:    map[string]struct{}{},
		Prohibited:        map[string]struct{}{},
	}
	r.IsMisspelled = r.isMisspelled
	return r
}

// SetSpeller wires the default SymSpell dictionary (Java defaultDictSpeller).
func (r *SymSpellRule) SetSpeller(sp *symimpl.SymSpell) {
	if r != nil {
		r.Speller = sp
	}
}

// SetUserSpeller wires the user SymSpell dictionary (Java userDictSpeller).
func (r *SymSpellRule) SetUserSpeller(sp *symimpl.SymSpell) {
	if r != nil {
		r.UserSpeller = sp
	}
}

func (r *SymSpellRule) AddWords(words ...string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, w := range words {
		r.Dictionary[w] = struct{}{}
		if r.Speller != nil {
			r.Speller.CreateDictionaryEntry(w, 1, nil)
		}
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
		if r.UserSpeller != nil {
			r.UserSpeller.CreateDictionaryEntry(w, 1, nil)
		}
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
	// When Speller is set: known if exact Lookup distance 0.
	if r.Speller != nil {
		got := r.Speller.LookupMax(word, symimpl.VerbosityTop, 0)
		if len(got) > 0 && got[0].Term == word && got[0].Distance == 0 {
			return false
		}
		// also try with edit distance 0 only — if empty, misspelled
		// Java isMisspelled throws not implemented on SymSpellRule — Match uses candidates instead.
		// For AcceptWord path: treat as misspelled if not exact in words map.
		// CreateDictionaryEntry puts into Speller.words — Lookup Top max 0 should find exact.
		return true
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

// Suggestions returns default-dict spelling suggestions (getSpellerMatches + filterCandidates).
func (r *SymSpellRule) Suggestions(word string) []string {
	return r.filterCandidates(r.getSpellerMatches(word, false))
}

// getSpellerMatches ports SymSpellRule.getSpellerMatches.
func (r *SymSpellRule) getSpellerMatches(word string, user bool) []string {
	if r == nil {
		return nil
	}
	// Real SymSpell path
	var sp *symimpl.SymSpell
	if user {
		sp = r.UserSpeller
	} else {
		sp = r.Speller
	}
	if sp != nil {
		maxEd := r.EditDistance
		if maxEd <= 0 {
			maxEd = sp.MaxDictionaryEditDistance()
		}
		// clamp to speller max (LookupMax panics if larger)
		if maxEd > sp.MaxDictionaryEditDistance() {
			maxEd = sp.MaxDictionaryEditDistance()
		}
		items := sp.LookupMax(word, r.Verbosity, maxEd)
		out := make([]string, 0, len(items))
		for _, it := range items {
			out = append(out, it.Term)
		}
		return out
	}
	// Map-inject fallback (tests without full SymSpell)
	return r.suggestionsFromMap(word, user)
}

func (r *SymSpellRule) suggestionsFromMap(word string, user bool) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	dict := r.Dictionary
	if user {
		dict = r.UserDictionary
	}
	if len(dict) == 0 || len(dict) > 10000 {
		return nil
	}
	maxEd := r.EditDistance
	if maxEd <= 0 {
		maxEd = 3
	}
	var out []string
	for w := range dict {
		if levenshtein(word, w) <= maxEd {
			out = append(out, w)
			if len(out) >= 10 {
				break
			}
		}
	}
	return out
}

// filterCandidates ports SymSpellRule.filterCandidates — drop ignore + prohibited.
func (r *SymSpellRule) filterCandidates(candidates []string) []string {
	if len(candidates) == 0 {
		return nil
	}
	out := make([]string, 0, len(candidates))
	for _, c := range candidates {
		if r.SpellingCheckRule != nil && r.IsInIgnoredSet(c) {
			continue
		}
		if _, bad := r.Prohibited[c]; bad {
			continue
		}
		out = append(out, c)
	}
	return out
}

// Match ports SymSpellRule.match.
func (r *SymSpellRule) Match(sentence *languagetool.AnalyzedSentence) ([]*rules.RuleMatch, error) {
	if sentence == nil || r == nil {
		return nil, nil
	}
	var out []*rules.RuleMatch
	for _, tok := range sentence.GetTokensWithoutWhitespace() {
		// Java: sentenceStart || immunized || ignoredBySpeller || isNonWord
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
		// Java ignoredWords.contains(word)
		if r.SpellingCheckRule != nil && r.IgnoreWord(w) {
			continue
		}
		candidates := r.filterCandidates(r.getSpellerMatches(w, false))
		userCandidates := r.filterCandidates(r.getSpellerMatches(w, true))

		var m *rules.RuleMatch
		if len(candidates) == 0 && len(userCandidates) == 0 {
			// unknown — flag (Java always creates match when both empty)
			// Skip if AcceptWord (map dict exact / ignore)
			if r.AcceptWord(w) {
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
