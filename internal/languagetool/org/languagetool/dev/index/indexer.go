package index

import "sync"

// Indexer is an in-memory document index for green tests (Lucene deferred).
type Indexer struct {
	mu   sync.RWMutex
	docs map[string]string // id → text
}

func NewIndexer() *Indexer {
	return &Indexer{docs: map[string]string{}}
}

func (ix *Indexer) Add(id, text string) {
	if ix == nil {
		return
	}
	ix.mu.Lock()
	ix.docs[id] = text
	ix.mu.Unlock()
}

func (ix *Indexer) Get(id string) (string, bool) {
	if ix == nil {
		return "", false
	}
	ix.mu.RLock()
	defer ix.mu.RUnlock()
	t, ok := ix.docs[id]
	return t, ok
}

func (ix *Indexer) Size() int {
	if ix == nil {
		return 0
	}
	ix.mu.RLock()
	defer ix.mu.RUnlock()
	return len(ix.docs)
}

// SearchSubstring returns doc IDs whose text contains q (soft search).
func (ix *Indexer) SearchSubstring(q string) []string {
	if ix == nil || q == "" {
		return nil
	}
	ix.mu.RLock()
	defer ix.mu.RUnlock()
	var out []string
	for id, text := range ix.docs {
		if containsFold(text, q) {
			out = append(out, id)
		}
	}
	return out
}

func containsFold(text, q string) bool {
	return len(q) > 0 && (text == q || len(text) >= len(q) &&
		(indexFold(text, q) >= 0))
}

func indexFold(s, substr string) int {
	// simple case-sensitive first; soft
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
