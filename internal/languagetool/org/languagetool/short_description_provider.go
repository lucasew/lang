package languagetool

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"sync"
)

// ShortDescriptionProvider ports org.languagetool.ShortDescriptionProvider.
// Loads word→short definition maps (tab-separated) per language code.
type ShortDescriptionProvider struct {
	// LoadLines returns lines for path like "/en/word_definitions.txt".
	// When nil, GetShortDescription always returns "".
	// Java uses JLanguageTool.getDataBroker().resourceExists + getFromResourceDirAsLines.
	LoadLines func(path string) ([]string, error)

	mu    sync.Mutex
	cache map[string]map[string]string // lang short code → word → def
}

func NewShortDescriptionProvider() *ShortDescriptionProvider {
	return &ShortDescriptionProvider{cache: map[string]map[string]string{}}
}

// GetShortDescription returns a short definition for word in langCode, or "".
// Ports getShortDescription(String word, Language lang) — nullable String.
func (p *ShortDescriptionProvider) GetShortDescription(word, langCode string) string {
	if p == nil {
		return ""
	}
	m := p.allDescriptions(langCode)
	if m == nil {
		return ""
	}
	// Java Map.get returns null when missing; Go returns "".
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

// init ports ShortDescriptionProvider.init(Language).
// Throws (panics) on bad tab format like Java RuntimeException.
func (p *ShortDescriptionProvider) init(langCode string) map[string]string {
	if p.LoadLines == nil {
		return map[string]string{}
	}
	path := "/" + langCode + "/word_definitions.txt"
	lines, err := p.LoadLines(path)
	if err != nil {
		// Java: resourceExists false → empty map; IO errors from broker not thrown here.
		return map[string]string{}
	}
	if len(lines) == 0 {
		return map[string]string{}
	}
	out := map[string]string{}
	for _, line := range lines {
		// Java: if (line.startsWith("#") || line.trim().isEmpty()) continue;
		// Note: startsWith("#") is on the raw line (not trimmed).
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) != 2 {
			// Java: throw new RuntimeException("Format in " + path + " not expected...")
			panic(fmt.Sprintf("Format in %s not expected, expected 2 tab-separated columns: '%s'", path, line))
		}
		out[parts[0]] = parts[1]
	}
	return out
}

// ParseWordDefinitions parses a definitions stream (for tests / brokers).
// Soft on bad lines (helper, not a Java method).
func ParseWordDefinitions(r io.Reader) (map[string]string, error) {
	out := map[string]string{}
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
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
