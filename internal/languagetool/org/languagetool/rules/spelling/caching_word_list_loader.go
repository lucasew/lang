package spelling

import (
	"bufio"
	"io"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/broker"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// CachingWordListLoader ports org.languagetool.rules.spelling.CachingWordListLoader.
//
// Java holds a static Guava LoadingCache<path, List> with expireAfterWrite(10, MINUTES).
// Go maps that to a process-level path→[]string cache shared by all instances.
// TTL is omitted (pure performance mapping); same-path loads reuse the cached list
// without re-reading the broker until the process-level cache is cleared.
//
// Resource loading uses the package DataBroker, mapping JLanguageTool.getDataBroker()
// (Java static). Wire via SetDataBroker — the full JLanguageTool broker twin is a
// separate sector; this package-level hook is the leaf-safe injection point.
type CachingWordListLoader struct{}

// dataBroker maps JLanguageTool.getDataBroker() for word-list loads.
var (
	dataBrokerMu sync.RWMutex
	dataBroker   broker.ResourceDataBroker
)

// wordListCache maps Java's static LoadingCache (no expireAfterWrite TTL here).
var (
	wordListCacheMu sync.Mutex
	wordListCache   = map[string][]string{}
)

// SetDataBroker wires the ResourceDataBroker used by LoadWords
// (Java: JLanguageTool.setDataBroker / getDataBroker).
func SetDataBroker(b broker.ResourceDataBroker) {
	dataBrokerMu.Lock()
	defer dataBrokerMu.Unlock()
	dataBroker = b
}

// GetDataBroker returns the broker used by LoadWords
// (Java: JLanguageTool.getDataBroker — without auto-default DefaultResourceDataBroker).
// Nil means no resources are visible (resourceExists → false path).
func GetDataBroker() broker.ResourceDataBroker {
	dataBrokerMu.RLock()
	defer dataBrokerMu.RUnlock()
	return dataBroker
}

// NewCachingWordListLoader ports the Java default constructor.
// Cache is process-level (Java static), not per instance.
func NewCachingWordListLoader() *CachingWordListLoader {
	return &CachingWordListLoader{}
}

// clearWordListCache clears the process-level cache (test helper; no Java twin).
func clearWordListCache() {
	wordListCacheMu.Lock()
	wordListCache = map[string][]string{}
	wordListCacheMu.Unlock()
}

// LoadWords ports CachingWordListLoader.loadWords(String filePath).
// Java: return cache.getUnchecked(filePath) — loads via DataBroker on miss.
func (l *CachingWordListLoader) LoadWords(filePath string) []string {
	wordListCacheMu.Lock()
	if words, ok := wordListCache[filePath]; ok {
		wordListCacheMu.Unlock()
		return words
	}
	wordListCacheMu.Unlock()

	words := loadWordList(filePath)

	wordListCacheMu.Lock()
	if existing, ok := wordListCache[filePath]; ok {
		wordListCacheMu.Unlock()
		return existing
	}
	wordListCache[filePath] = words
	wordListCacheMu.Unlock()
	return words
}

// loadWordList ports CacheLoader.load(fileInClassPath).
//
//	if (!JLanguageTool.getDataBroker().resourceExists(fileInClassPath)) return empty;
//	lines = getFromResourceDirAsLines(path);
//	for each line: skip empty or startsWith("#"); else add substringBefore(trim, "#").trim();
//	return unmodifiable list
func loadWordList(fileInClassPath string) []string {
	b := GetDataBroker()
	if b == nil || !b.ResourceExists(fileInClassPath) {
		// Java returns a fresh empty ArrayList (not unmodifiable) when missing.
		return []string{}
	}
	lines, err := b.GetFromResourceDirAsLines(fileInClassPath)
	if err != nil {
		// Java CacheLoader throws IOException; getUnchecked wraps as unchecked.
		panic(err)
	}
	return parseWordListFromLines(lines)
}

// parseWordListFromLines is the per-line loop from CachingWordListLoader.CacheLoader.load.
//
// Bug-for-bug: empty / "#" checks use the raw line (before trim). Value uses
// String.trim (≤U+0020) then StringUtils.substringBefore(..., "#").trim().
// Empty results after that processing are still added (Java result.add always).
func parseWordListFromLines(lines []string) []string {
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Apache StringUtils.substringBefore(line.trim(), "#").trim()
		line = tools.JavaStringTrim(line)
		if i := strings.Index(line, "#"); i >= 0 {
			line = line[:i]
		}
		line = tools.JavaStringTrim(line)
		result = append(result, line)
	}
	return result
}

// ParseWordListLines parses classpath-style word list content from a reader.
// Same line rules as CachingWordListLoader.CacheLoader.load (see parseWordListFromLines).
// Java loads via getFromResourceDirAsLines; this helper is for unit tests of the parse loop.
func ParseWordListLines(r io.Reader) ([]string, error) {
	var lines []string
	sc := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return parseWordListFromLines(lines), nil
}
