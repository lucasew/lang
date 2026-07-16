package languagetool

import (
	"fmt"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/broker"
)

// ShortDescriptionProvider ports org.languagetool.ShortDescriptionProvider.
// Loads word\tdefinition lines from /{lang}/word_definitions.txt via a ResourceDataBroker.
type ShortDescriptionProvider struct {
	mu     sync.Mutex
	cache  map[string]map[string]string // lang short code → word → def
	Broker broker.ResourceDataBroker
}

func NewShortDescriptionProvider(b broker.ResourceDataBroker) *ShortDescriptionProvider {
	return &ShortDescriptionProvider{
		cache:  map[string]map[string]string{},
		Broker: b,
	}
}

// GetShortDescription returns a short definition or empty if unknown.
func (p *ShortDescriptionProvider) GetShortDescription(word, langShortCode string) string {
	if p == nil {
		return ""
	}
	defs := p.allDescriptions(langShortCode)
	return defs[word]
}

func (p *ShortDescriptionProvider) allDescriptions(langShortCode string) map[string]string {
	p.mu.Lock()
	defer p.mu.Unlock()
	if m, ok := p.cache[langShortCode]; ok {
		return m
	}
	m := p.init(langShortCode)
	p.cache[langShortCode] = m
	return m
}

func (p *ShortDescriptionProvider) init(langShortCode string) map[string]string {
	if p.Broker == nil {
		return map[string]string{}
	}
	path := "/" + langShortCode + "/word_definitions.txt"
	if !p.Broker.ResourceExists(path) {
		// also try without leading slash style
		path2 := langShortCode + "/word_definitions.txt"
		if !p.Broker.ResourceExists(path2) {
			return map[string]string{}
		}
		path = path2
	}
	lines, err := p.Broker.GetFromResourceDirAsLines(path)
	if err != nil {
		return map[string]string{}
	}
	out := map[string]string{}
	for _, line := range lines {
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) != 2 {
			panic(fmt.Sprintf("Format in %s not expected, expected 2 tab-separated columns: '%s'", path, line))
		}
		out[parts[0]] = parts[1]
	}
	return out
}
