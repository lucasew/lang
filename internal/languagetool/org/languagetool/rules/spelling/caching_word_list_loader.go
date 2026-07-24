package spelling

import (
	"bufio"
	"io"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// CachingWordListLoader ports org.languagetool.rules.spelling.CachingWordListLoader.
// Uses an in-memory path → lines cache; callers supply content via LoadWordsFromReader
// or Register/LoadWords after preloading.
type CachingWordListLoader struct {
	mu    sync.Mutex
	cache map[string][]string
}

func NewCachingWordListLoader() *CachingWordListLoader {
	return &CachingWordListLoader{cache: map[string][]string{}}
}

// ParseWordListLines parses classpath-style word list content (skip #/empty, strip trailing comments).
// Java: StringUtils.substringBefore(line.trim(), "#").trim() — String.trim, not Unicode TrimSpace.
func ParseWordListLines(r io.Reader) ([]string, error) {
	var result []string
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Apache StringUtils.substringBefore(line.trim(), "#").trim()
		line = tools.JavaStringTrim(line)
		if i := strings.Index(line, "#"); i >= 0 {
			line = line[:i]
		}
		line = tools.JavaStringTrim(line)
		if line != "" {
			result = append(result, line)
		}
	}
	return result, sc.Err()
}

// LoadWordsFromReader parses and caches under filePath.
func (l *CachingWordListLoader) LoadWordsFromReader(filePath string, r io.Reader) ([]string, error) {
	l.mu.Lock()
	if words, ok := l.cache[filePath]; ok {
		l.mu.Unlock()
		return words, nil
	}
	l.mu.Unlock()
	words, err := ParseWordListLines(r)
	if err != nil {
		return nil, err
	}
	l.mu.Lock()
	l.cache[filePath] = words
	l.mu.Unlock()
	return words, nil
}

// LoadWords returns cached words for path (nil/empty if never loaded).
func (l *CachingWordListLoader) LoadWords(filePath string) []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.cache[filePath]
}

// Register puts words into the cache under filePath.
func (l *CachingWordListLoader) Register(filePath string, words []string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.cache[filePath] = append([]string(nil), words...)
}
