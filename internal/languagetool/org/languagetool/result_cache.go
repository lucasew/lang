package languagetool

import (
	"sort"
	"strings"
	"sync"
)

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

// inputSentenceKey mirrors InputSentence.Equal dimensions used as a Guava cache key.
func inputSentenceKey(s InputSentence) string {
	text := ""
	if s.Analyzed != nil {
		text = s.Analyzed.GetText()
	}
	var b strings.Builder
	b.WriteString(text)
	b.WriteByte(0)
	b.WriteString(s.LanguageCode)
	b.WriteByte(0)
	b.WriteString(s.MotherTongueCode)
	b.WriteByte(0)
	b.WriteString(s.Mode)
	b.WriteByte(0)
	b.WriteString(string(s.Level))
	b.WriteByte(0)
	b.WriteString(sortedSetKey(s.DisabledRules))
	b.WriteByte(0)
	b.WriteString(sortedSetKey(s.DisabledCategories))
	b.WriteByte(0)
	b.WriteString(sortedSetKey(s.EnabledRules))
	b.WriteByte(0)
	b.WriteString(sortedSetKey(s.EnabledCategories))
	b.WriteByte(0)
	b.WriteString(strings.Join(s.AltLanguageCodes, ","))
	b.WriteByte(0)
	if s.TextSessionID != nil {
		b.WriteString(itoa64(*s.TextSessionID))
	}
	b.WriteByte(0)
	// tone tags
	if len(s.ToneTags) > 0 {
		keys := make([]string, 0, len(s.ToneTags))
		for t := range s.ToneTags {
			keys = append(keys, string(t))
		}
		sort.Strings(keys)
		b.WriteString(strings.Join(keys, ","))
	}
	return b.String()
}

func sortedSetKey(m map[string]struct{}) string {
	if len(m) == 0 {
		return ""
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, ",")
}

func itoa64(n int64) string {
	// small helper without strconv import pull for negatives
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
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

// remoteMatchKey is language-agnostic text key for remote-rule match caching
// (Java ResultCache remote matches keyed by analyzed sentence content).
func remoteMatchKey(sentenceText, ruleID string) string {
	return sentenceText + "\x00" + ruleID
}

// GetRemoteMatchesIfPresent returns cached remote matches for a sentence text + rule id.
func (c *ResultCache) GetRemoteMatchesIfPresent(sentenceText, ruleID string) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, ok := c.remoteMatches[remoteMatchKey(sentenceText, ruleID)]
	if ok {
		c.hits++
	} else {
		c.misses++
	}
	return v, ok
}

// PutRemoteMatches stores remote-rule match results for a sentence text + rule id.
func (c *ResultCache) PutRemoteMatches(sentenceText, ruleID string, matches any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.maxSize == 0 {
		return
	}
	if len(c.remoteMatches) >= c.maxSize {
		for k := range c.remoteMatches {
			delete(c.remoteMatches, k)
			break
		}
	}
	c.remoteMatches[remoteMatchKey(sentenceText, ruleID)] = matches
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
