package server

import (
	"sort"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// UserDictionary is an in-process stand-in for the premium DB-backed word list
// used by /v2/words, /v2/words/add, and /v2/words/delete (Java server API surface).
// Not an invent soft pack — incomplete vs DB persistence.
type UserDictionary struct {
	mu    sync.RWMutex
	words map[string]map[string]struct{} // username → word set
}

func NewUserDictionary() *UserDictionary {
	return &UserDictionary{words: map[string]map[string]struct{}{}}
}

func (d *UserDictionary) key(username string) string {
	// Java DB path: username.trim().isEmpty() → reject; in-memory stand-in maps empty→anon.
	u := tools.JavaStringTrim(username)
	if u == "" {
		return "anon"
	}
	return strings.ToLower(u)
}

// List returns sorted dictionary words for username (empty username → anon).
func (d *UserDictionary) List(username string, offset, limit int) []string {
	if d == nil {
		return nil
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	set := d.words[d.key(username)]
	out := make([]string, 0, len(set))
	for w := range set {
		out = append(out, w)
	}
	sort.Strings(out)
	if offset < 0 {
		offset = 0
	}
	if offset > len(out) {
		return nil
	}
	out = out[offset:]
	if limit > 0 && limit < len(out) {
		out = out[:limit]
	}
	return out
}

// Add inserts word; returns true if newly added.
func (d *UserDictionary) Add(username, word string) bool {
	if d == nil {
		return false
	}
	// Java: word == null || word.trim().isEmpty()
	word = tools.JavaStringTrim(word)
	if word == "" {
		return false
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	k := d.key(username)
	if d.words[k] == nil {
		d.words[k] = map[string]struct{}{}
	}
	if _, ok := d.words[k][word]; ok {
		return false
	}
	d.words[k][word] = struct{}{}
	return true
}

// Delete removes word; returns true if it was present.
func (d *UserDictionary) Delete(username, word string) bool {
	if d == nil {
		return false
	}
	word = tools.JavaStringTrim(word)
	if word == "" {
		return false
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	k := d.key(username)
	set := d.words[k]
	if set == nil {
		return false
	}
	if _, ok := set[word]; !ok {
		return false
	}
	delete(set, word)
	return true
}

// All returns all words for username (no pagination).
func (d *UserDictionary) All(username string) []string {
	return d.List(username, 0, 0)
}
