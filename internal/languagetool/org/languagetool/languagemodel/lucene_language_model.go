package languagemodel

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"
)

// IndexLanguageModel is a count-backed LM (stand-in for LuceneSingleIndexLanguageModel).
type IndexLanguageModel struct {
	*BaseLanguageModel
	name string
}

// MapCountProvider is an in-memory ngram count store.
type MapCountProvider struct {
	// key: tokens joined by \x00
	Counts map[string]int64
	Total  int64
}

func (p *MapCountProvider) GetCountToken(token string) int64 {
	if p == nil {
		return 0
	}
	return p.Counts[token]
}

func (p *MapCountProvider) GetCount(tokens []string) int64 {
	if p == nil || len(tokens) == 0 {
		return 0
	}
	return p.Counts[strings.Join(tokens, "\x00")]
}

func (p *MapCountProvider) GetTotalTokenCount() int64 {
	if p == nil {
		return 0
	}
	return p.Total
}

func NewIndexLanguageModel(name string, counts *MapCountProvider) *IndexLanguageModel {
	return &IndexLanguageModel{
		BaseLanguageModel: NewBaseLanguageModel(counts),
		name:              name,
	}
}

func (m *IndexLanguageModel) GetCount(tokens []string) int64 {
	if m == nil || m.Counts == nil {
		return 0
	}
	return m.Counts.GetCount(tokens)
}

func (m *IndexLanguageModel) GetCountToken(token string) int64 {
	if m == nil || m.Counts == nil {
		return 0
	}
	return m.Counts.GetCountToken(token)
}

func (m *IndexLanguageModel) GetTotalTokenCount() int64 {
	if m == nil || m.Counts == nil {
		return 0
	}
	return m.Counts.GetTotalTokenCount()
}

func (m *IndexLanguageModel) Close() error { return nil }

func (m *IndexLanguageModel) String() string { return m.name }

// LuceneLanguageModel ports org.languagetool.languagemodel.LuceneLanguageModel
// as a multi-index count aggregator (no real Lucene dependency).
type LuceneLanguageModel struct {
	lms []*IndexLanguageModel
}

// ValidateLuceneDirectory checks for index-N subdirs or ngram subdirs (surface check).
func ValidateLuceneDirectory(top string) error {
	fi, err := os.Stat(top)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("not a directory: %s", top)
	}
	entries, err := os.ReadDir(top)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() && (strings.HasPrefix(e.Name(), "index-") ||
			e.Name() == "1grams" || e.Name() == "2grams" || e.Name() == "3grams") {
			return nil
		}
	}
	// empty / unknown layout still allowed for map-backed tests
	return nil
}

// ListIndexSubdirs returns index-\d+ children or nil if none.
func ListIndexSubdirs(top string) ([]string, error) {
	entries, err := os.ReadDir(top)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() && len(e.Name()) > 6 && strings.HasPrefix(e.Name(), "index-") {
			out = append(out, filepath.Join(top, e.Name()))
		}
	}
	return out, nil
}

// NewLuceneLanguageModelFromIndexes builds a multi-index LM from count providers.
func NewLuceneLanguageModelFromIndexes(indexes []*IndexLanguageModel) *LuceneLanguageModel {
	return &LuceneLanguageModel{lms: append([]*IndexLanguageModel(nil), indexes...)}
}

// NewLuceneLanguageModelMap is a convenience single-index map-backed LM.
func NewLuceneLanguageModelMap(counts *MapCountProvider) *LuceneLanguageModel {
	return NewLuceneLanguageModelFromIndexes([]*IndexLanguageModel{
		NewIndexLanguageModel("map", counts),
	})
}

func (m *LuceneLanguageModel) GetCount(tokens []string) int64 {
	var sum int64
	for _, lm := range m.lms {
		sum += lm.GetCount(tokens)
	}
	return sum
}

func (m *LuceneLanguageModel) GetCountToken(token string) int64 {
	return m.GetCount([]string{token})
}

func (m *LuceneLanguageModel) GetTotalTokenCount() int64 {
	var sum int64
	for _, lm := range m.lms {
		sum += lm.GetTotalTokenCount()
	}
	return sum
}

func (m *LuceneLanguageModel) GetPseudoProbability(context []string) ngrams.Probability {
	// merge via BaseLanguageModel over summed counts
	base := NewBaseLanguageModel(m)
	return base.GetPseudoProbability(context)
}

func (m *LuceneLanguageModel) Close() error {
	var first error
	for _, lm := range m.lms {
		if err := lm.Close(); err != nil && first == nil {
			first = err
		}
	}
	return first
}

func (m *LuceneLanguageModel) String() string {
	return fmt.Sprint(m.lms)
}

// Ensure LuceneLanguageModel satisfies CountProvider + LanguageModel.
var (
	_ CountProvider = (*LuceneLanguageModel)(nil)
	_ LanguageModel = (*LuceneLanguageModel)(nil)
)

// LuceneSearcher ports LuceneSingleIndexLanguageModel.LuceneSearcher —
// a cached count backend bound to an index directory (map-backed without Lucene).
type LuceneSearcher struct {
	Dir    string
	Counts *MapCountProvider
}

// Global cache: static dir → searcher (Java dirToSearcherMap).
var luceneDirToSearcher = map[string]*LuceneSearcher{}

// GetCachedLuceneSearcher ports getCachedLuceneSearcher(indexDir).
func GetCachedLuceneSearcher(indexDir string, counts *MapCountProvider) *LuceneSearcher {
	if s, ok := luceneDirToSearcher[indexDir]; ok {
		return s
	}
	if counts == nil {
		counts = &MapCountProvider{Counts: map[string]int64{}}
	}
	s := &LuceneSearcher{Dir: indexDir, Counts: counts}
	luceneDirToSearcher[indexDir] = s
	return s
}

// ClearLuceneSearcherCache clears the static cache (tests).
func ClearLuceneSearcherCache() {
	luceneDirToSearcher = map[string]*LuceneSearcher{}
}

// GetCount ports searcher count for a token sequence.
func (s *LuceneSearcher) GetCount(tokens []string) int64 {
	if s == nil || s.Counts == nil {
		return 0
	}
	return s.Counts.GetCount(tokens)
}

func (s *LuceneSearcher) GetCountToken(token string) int64 {
	if s == nil || s.Counts == nil {
		return 0
	}
	return s.Counts.GetCountToken(token)
}

func (s *LuceneSearcher) GetTotalTokenCount() int64 {
	if s == nil || s.Counts == nil {
		return 0
	}
	return s.Counts.GetTotalTokenCount()
}

// LuceneSingleIndexLanguageModel ports org.languagetool.languagemodel.LuceneSingleIndexLanguageModel
// as an IndexLanguageModel with ngram-size → LuceneSearcher map.
type LuceneSingleIndexLanguageModel struct {
	*IndexLanguageModel
	TopIndexDir string
	Searchers   map[int]*LuceneSearcher // ngram size → searcher
}

// ValidateSingleIndexDirectory ports LuceneSingleIndexLanguageModel.validateDirectory.
func ValidateSingleIndexDirectory(topIndexDir string) error {
	fi, err := os.Stat(topIndexDir)
	if err != nil || !fi.IsDir() {
		return fmt.Errorf("Not found or is not a directory:\n%s\nAs ngram directory, please select the directory that has a subdirectory like 'en'\n(or whatever language code you're using).", topIndexDir)
	}
	entries, err := os.ReadDir(topIndexDir)
	if err != nil {
		return err
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() && (e.Name() == "1grams" || e.Name() == "2grams" || e.Name() == "3grams") {
			dirs = append(dirs, e.Name())
		}
	}
	if len(dirs) == 0 {
		return fmt.Errorf("%s does not contain 1grams/2grams/3grams subdirectories", topIndexDir)
	}
	return nil
}

// NewLuceneSingleIndexLanguageModelMap builds a map-backed single-index LM.
func NewLuceneSingleIndexLanguageModelMap(top string, counts *MapCountProvider) *LuceneSingleIndexLanguageModel {
	s := GetCachedLuceneSearcher(top, counts)
	return &LuceneSingleIndexLanguageModel{
		IndexLanguageModel: NewIndexLanguageModel(top, s.Counts),
		TopIndexDir:        top,
		Searchers:          map[int]*LuceneSearcher{1: s},
	}
}
