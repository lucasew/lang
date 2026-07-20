package languagetool

import (
	"sort"
	"strings"
	"sync"
	"time"
)

// ResultCache ports org.languagetool.ResultCache with a simple bounded map.
// Match values are stored as any (typically []*rules.RuleMatch) to avoid an
// import cycle with the rules package. Guava CacheBuilder expire/stats are
// approximated (access-time expiry optional; hit/miss counters).
type ResultCache struct {
	mu            sync.Mutex
	maxSize       int
	expireAfter   time.Duration // 0 = no expiry (still size-bounded)
	matches       map[string]cacheEntry
	sentences     map[string]sentenceEntry
	remoteMatches map[string]cacheEntry
	// separate stats like Java matchesCache + sentenceCache for hitRate average
	matchHits, matchMisses, sentHits, sentMisses, remoteHits, remoteMisses int64
}

type cacheEntry struct {
	val      any
	lastRead time.Time
}

type sentenceEntry struct {
	val      *AnalyzedSentence
	lastRead time.Time
}

// NewResultCache ports ResultCache(long maxSize) — expire 5 minutes after access.
func NewResultCache(maxSize int) *ResultCache {
	return NewResultCacheExpire(maxSize, 5*time.Minute)
}

// NewResultCacheExpire ports ResultCache(maxSize, expireAfter, timeUnit).
func NewResultCacheExpire(maxSize int, expireAfter time.Duration) *ResultCache {
	if maxSize < 0 {
		panic("Result cache size must be >= 0: " + itoa64(int64(maxSize)))
	}
	return &ResultCache{
		maxSize:       maxSize,
		expireAfter:   expireAfter,
		matches:       map[string]cacheEntry{},
		sentences:     map[string]sentenceEntry{},
		remoteMatches: map[string]cacheEntry{},
	}
}

// MatchesWeigh ports MatchesWeigher.weigh.
// return 1 + sentence text length/75 + matches.size()
func MatchesWeigh(sentenceText string, matchCount int) int {
	return 1 + len(sentenceText)/75 + matchCount
}

// RemoteMatchesWeigh ports RemoteMatchesWeigher.weigh.
func RemoteMatchesWeigh(sentenceText string) int {
	return 1 + len(sentenceText)/75
}

// SentenceWeigh ports SentenceWeigher.weigh.
func SentenceWeigh(text string) int {
	return 1 + len(text)/75
}

// inputSentenceKey mirrors InputSentence.equal dimensions used as a Guava cache key.
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

func (c *ResultCache) expired(t time.Time) bool {
	if c.expireAfter <= 0 {
		return false
	}
	return time.Since(t) > c.expireAfter
}

// GetMatchesIfPresent returns the cached matches (any) if present.
func (c *ResultCache) GetMatchesIfPresent(key InputSentence) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	k := inputSentenceKey(key)
	e, ok := c.matches[k]
	if ok && !c.expired(e.lastRead) {
		e.lastRead = time.Now()
		c.matches[k] = e
		c.matchHits++
		return e.val, true
	}
	if ok {
		delete(c.matches, k)
	}
	c.matchMisses++
	return nil, false
}

// PutMatches stores match results for a sentence key.
func (c *ResultCache) PutMatches(key InputSentence, matches any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.maxSize == 0 {
		return
	}
	// Java maximumWeight = maxSize/2 for each cache
	limit := c.maxSize / 2
	if limit < 1 {
		limit = c.maxSize
	}
	for len(c.matches) >= limit {
		for k := range c.matches {
			delete(c.matches, k)
			break
		}
	}
	c.matches[inputSentenceKey(key)] = cacheEntry{val: matches, lastRead: time.Now()}
}

func (c *ResultCache) GetSentenceIfPresent(key SimpleInputSentence) (*AnalyzedSentence, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	k := simpleInputKey(key)
	e, ok := c.sentences[k]
	if ok && !c.expired(e.lastRead) {
		e.lastRead = time.Now()
		c.sentences[k] = e
		c.sentHits++
		return e.val, true
	}
	if ok {
		delete(c.sentences, k)
	}
	c.sentMisses++
	return nil, false
}

func (c *ResultCache) PutSentence(key SimpleInputSentence, a *AnalyzedSentence) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.maxSize == 0 {
		return
	}
	limit := c.maxSize / 2
	if limit < 1 {
		limit = c.maxSize
	}
	for len(c.sentences) >= limit {
		for k := range c.sentences {
			delete(c.sentences, k)
			break
		}
	}
	c.sentences[simpleInputKey(key)] = sentenceEntry{val: a, lastRead: time.Now()}
}

func remoteMatchKey(sentenceText, ruleID string) string {
	return sentenceText + "\x00" + ruleID
}

func (c *ResultCache) GetRemoteMatchesIfPresent(sentenceText, ruleID string) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	k := remoteMatchKey(sentenceText, ruleID)
	e, ok := c.remoteMatches[k]
	if ok && !c.expired(e.lastRead) {
		e.lastRead = time.Now()
		c.remoteMatches[k] = e
		c.remoteHits++
		return e.val, true
	}
	if ok {
		delete(c.remoteMatches, k)
	}
	c.remoteMisses++
	return nil, false
}

func (c *ResultCache) PutRemoteMatches(sentenceText, ruleID string, matches any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.maxSize == 0 {
		return
	}
	limit := c.maxSize / 2
	if limit < 1 {
		limit = c.maxSize
	}
	for len(c.remoteMatches) >= limit {
		for k := range c.remoteMatches {
			delete(c.remoteMatches, k)
			break
		}
	}
	c.remoteMatches[remoteMatchKey(sentenceText, ruleID)] = cacheEntry{val: matches, lastRead: time.Now()}
}

// HitCount ports hitCount — matches + sentence cache hit counts.
func (c *ResultCache) HitCount() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.matchHits + c.sentHits + c.remoteHits
}

// RequestCount ports requestCount.
func (c *ResultCache) RequestCount() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.matchHits + c.matchMisses + c.sentHits + c.sentMisses + c.remoteHits + c.remoteMisses
}

// HitRate ports hitRate — average of matches and sentence cache hit rates.
func (c *ResultCache) HitRate() float64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	var mRate, sRate float64
	if mt := c.matchHits + c.matchMisses; mt > 0 {
		mRate = float64(c.matchHits) / float64(mt)
	}
	if st := c.sentHits + c.sentMisses; st > 0 {
		sRate = float64(c.sentHits) / float64(st)
	}
	// Java always averages the two stats.hitRate() values (0 if empty)
	return (mRate + sRate) / 2.0
}
