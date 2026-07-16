package dumpcheck

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// MixingSentenceSource ports org.languagetool.dev.dumpcheck.MixingSentenceSource.
// Alternately returns sentences from different sentence sources.
type MixingSentenceSource struct {
	sources            []SentenceSource
	count              int
	sourceDistribution map[string]int
}

// NewMixingSentenceSource builds a mixer from already-constructed sources.
func NewMixingSentenceSource(sources []SentenceSource) *MixingSentenceSource {
	cp := make([]SentenceSource, len(sources))
	copy(cp, sources)
	return &MixingSentenceSource{
		sources:            cp,
		sourceDistribution: map[string]int{},
	}
}

// CreateMixingSentenceSource opens dump files by naming convention:
//   - *.xml → WikipediaSentenceSource
//   - tatoeba-* → TatoebaSentenceSource
//   - *.txt → PlainTextSentenceSource
//   - *.xz / *.cc → CommonCrawlSentenceSource (plain lines; xz decompression deferred)
func CreateMixingSentenceSource(dumpFileNames []string, langCode string) (*MixingSentenceSource, error) {
	var sources []SentenceSource
	for _, name := range dumpFileNames {
		base := filepath.Base(name)
		f, err := os.Open(name)
		if err != nil {
			return nil, err
		}
		// Note: callers own closing via process lifetime; green slice opens for duration of iteration.
		// For tests we use in-memory sources; file-backed path is for parity.
		switch {
		case strings.HasSuffix(base, ".xml"):
			sources = append(sources, &closingSource{f: f, inner: NewWikipediaSentenceSource(f, langCode)})
		case strings.HasPrefix(base, "tatoeba-"):
			sources = append(sources, &closingSource{f: f, inner: NewTatoebaSentenceSource(f)})
		case strings.HasSuffix(base, ".txt"):
			sources = append(sources, &closingSource{f: f, inner: NewPlainTextSentenceSource(f)})
		case strings.HasSuffix(base, ".xz") || strings.HasSuffix(base, ".cc"):
			// CommonCrawl: expect already-decompressed stream for green tests; .xz auto-decode deferred
			sources = append(sources, &closingSource{f: f, inner: NewCommonCrawlSentenceSource(f)})
		default:
			_ = f.Close()
			return nil, fmt.Errorf("could not find a source handler for %s - Wikipedia files must be named '*.xml', Tatoeba files must be named 'tatoeba-*', CommonCrawl files '*.xz', plain text files '*.txt'", name)
		}
	}
	return NewMixingSentenceSource(sources), nil
}

func (c *closingSource) HasNext() bool           { return c.inner.HasNext() }
func (c *closingSource) Next() (Sentence, error) { return c.inner.Next() }
func (c *closingSource) GetSource() string       { return c.inner.GetSource() }

func (m *MixingSentenceSource) GetSourceDistribution() map[string]int {
	out := make(map[string]int, len(m.sourceDistribution))
	for k, v := range m.sourceDistribution {
		out[k] = v
	}
	return out
}

func (m *MixingSentenceSource) HasNext() bool {
	for _, s := range m.sources {
		if s.HasNext() {
			return true
		}
	}
	return false
}

func (m *MixingSentenceSource) Next() (Sentence, error) {
	if len(m.sources) == 0 {
		return Sentence{}, fmt.Errorf("no such element")
	}
	sentenceSource := m.sources[m.count%len(m.sources)]
	for !sentenceSource.HasNext() {
		// remove exhausted source (Java mutates list)
		idx := m.count % len(m.sources)
		m.sources = append(m.sources[:idx], m.sources[idx+1:]...)
		if len(m.sources) == 0 {
			return Sentence{}, fmt.Errorf("no such element")
		}
		m.count++
		sentenceSource = m.sources[m.count%len(m.sources)]
	}
	m.count++
	next, err := sentenceSource.Next()
	if err != nil {
		return Sentence{}, err
	}
	m.updateDistribution(next)
	return next, nil
}

func (m *MixingSentenceSource) updateDistribution(next Sentence) {
	m.sourceDistribution[next.GetSource()]++
}

func (m *MixingSentenceSource) GetSource() string {
	parts := make([]string, 0, len(m.sources))
	for _, s := range m.sources {
		parts = append(parts, s.GetSource())
	}
	return strings.Join(parts, ", ")
}
