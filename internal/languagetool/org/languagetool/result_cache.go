package languagetool

import "sync"

// ResultCache ports org.languagetool.ResultCache with a simple bounded map.
// Match values are stored as any (typically []*rules.RuleMatch) to avoid an
// import cycle with the rules package.
type ResultCache struct {
	mu            sync.Mutex
	maxSize       int
	matches       map[string]any // InputSentence key → match list
	sentences     map[string]*AnalyzedSentence
	remoteMatches map[string]any
	hits, misses  int64
}

func NewResultCache(maxSize int) *ResultCache {
	if maxSize < 0 {
		panic("Result cache size must be >= 0")
	}
	return &ResultCache{
		maxSize:       maxSize,
		matches:       map[string]any{},
		sentences:     map[string]*AnalyzedSentence{},
		remoteMatches: map[string]any{},
	}
}

func inputSentenceKey(s InputSentence) string {
	text := ""
	if s.Analyzed != nil {
		text = s.Analyzed.GetText()
	}
	return text + "\x00" + s.LanguageCode + "\x00" + s.Mode + "\x00" + string(s.Level)
}

func simpleInputKey(s SimpleInputSentence) string {
	return s.Text + "\x00" + s.LanguageCode
}

// GetMatchesIfPresent returns the cached matches (any) if present.
func (c *ResultCache) GetMatchesIfPresent(key InputSentence) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, ok := c.matches[inputSentenceKey(key)]
	if ok {
		c.hits++
	} else {
		c.misses++
	}
	return v, ok
}

// PutMatches stores match results for a sentence key.
func (c *ResultCache) PutMatches(key InputSentence, matches any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.maxSize == 0 {
		return
	}
	if len(c.matches) >= c.maxSize {
		for k := range c.matches {
			delete(c.matches, k)
			break
		}
	}
	c.matches[inputSentenceKey(key)] = matches
}

func (c *ResultCache) GetSentenceIfPresent(key SimpleInputSentence) (*AnalyzedSentence, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, ok := c.sentences[simpleInputKey(key)]
	if ok {
		c.hits++
	} else {
		c.misses++
	}
	return v, ok
}

func (c *ResultCache) PutSentence(key SimpleInputSentence, a *AnalyzedSentence) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.maxSize == 0 {
		return
	}
	if len(c.sentences) >= c.maxSize {
		for k := range c.sentences {
			delete(c.sentences, k)
			break
		}
	}
	c.sentences[simpleInputKey(key)] = a
}

func (c *ResultCache) HitCount() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.hits
}

func (c *ResultCache) RequestCount() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.hits + c.misses
}

func (c *ResultCache) HitRate() float64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	total := c.hits + c.misses
	if total == 0 {
		return 0
	}
	return float64(c.hits) / float64(total)
}
