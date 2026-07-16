package languagetool

import (
	"bufio"
	"io"
	"strings"
	"sync"
)

// ShortDescriptionProvider ports org.languagetool.ShortDescriptionProvider.
// Loads word→short definition maps (tab-separated) per language code.
type ShortDescriptionProvider struct {
	// LoadLines returns lines for path like "/en/word_definitions.txt".
	// When nil, GetShortDescription always returns "".
	LoadLines func(path string) ([]string, error)

	mu   sync.Mutex
	cache map[string]map[string]string // lang short code → word → def
}

func NewShortDescriptionProvider() *ShortDescriptionProvider {
	return &ShortDescriptionProvider{cache: map[string]map[string]string{}}
}

// GetShortDescription returns a short definition for word in langCode, or "".
func (p *ShortDescriptionProvider) GetShortDescription(word, langCode string) string {
	if p == nil || word == "" || langCode == "" {
		return ""
	}
	m := p.allDescriptions(langCode)
	return m[word]
}

func (p *ShortDescriptionProvider) allDescriptions(langCode string) map[string]string {
	p.mu.Lock()
	defer p.mu.Unlock()
	if m, ok := p.cache[langCode]; ok {
		return m
	}
	m := p.init(langCode)
	p.cache[langCode] = m
	return m
}

func (p *ShortDescriptionProvider) init(langCode string) map[string]string {
	if p.LoadLines == nil {
		return map[string]string{}
	}
	path := "/" + langCode + "/word_definitions.txt"
	lines, err := p.LoadLines(path)
	if err != nil || len(lines) == 0 {
		return map[string]string{}
	}
	out := map[string]string{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line[0] == '#' {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) != 2 {
			continue // skip bad lines rather than panic in production port
		}
		out[parts[0]] = parts[1]
	}
	return out
}

// ParseWordDefinitions parses a definitions stream (for tests / brokers).
func ParseWordDefinitions(r io.Reader) (map[string]string, error) {
	out := map[string]string{}
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || line[0] == '#' {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) != 2 {
			continue
		}
		out[parts[0]] = parts[1]
	}
	return out, sc.Err()
}
