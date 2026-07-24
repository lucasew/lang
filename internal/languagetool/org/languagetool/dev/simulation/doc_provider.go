package simulation

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
)

const maxVal = 20_000

// DocProvider ports org.languagetool.dev.simulation.DocProvider.
// Provides random-length documents with a production-like length distribution.
type DocProvider struct {
	mu   sync.Mutex
	docs []string
	rnd  *rand.Rand
}

func NewDocProvider(docs []string) *DocProvider {
	p := &DocProvider{docs: append([]string(nil), docs...)}
	p.Reset()
	return p
}

// Reset reseeds RNG to the Java fixed seed (120).
func (p *DocProvider) Reset() {
	p.rnd = rand.New(rand.NewSource(120))
}

// Remaining returns how many source snippets are left.
func (p *DocProvider) Remaining() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.docs)
}

// GetDoc consumes source snippets until the weighted target length is reached.
func (p *DocProvider) GetDoc() (string, error) {
	lenTarget := p.getWeightedRandomLength()
	p.mu.Lock()
	defer p.mu.Unlock()
	var appended strings.Builder
	paraSize := 0
	for appended.Len() < lenTarget {
		if len(p.docs) == 0 {
			return "", fmt.Errorf("not enough docs left to provide another document")
		}
		doc := p.docs[0]
		p.docs = p.docs[1:]
		appended.WriteString(doc)
		appended.WriteByte(' ')
		paraSize += len(doc)
		if paraSize > 250 && strings.HasSuffix(appended.String(), ". ") {
			appended.WriteString(doc)
			appended.WriteString("\n\n")
			paraSize = 0
		}
	}
	s := appended.String()
	if len(s) > lenTarget {
		// Java substring(0, len) is UTF-16-ish but for ASCII tests byte==char
		s = s[:lenTarget]
	}
	return s, nil
}

// GetWeightedRandomLength is exported for distribution tests.
func (p *DocProvider) GetWeightedRandomLength() int {
	return p.getWeightedRandomLength()
}

func (p *DocProvider) getWeightedRandomLength() int {
	max := p.getRandomMaxLength()
	min := max - 49
	if max == maxVal {
		min = 550
	}
	return min + p.rnd.Intn(max-min)
}

func (p *DocProvider) getRandomMaxLength() int {
	// Java: nextFloat()*100 then thresholds with fix=15.6
	rnd := p.rnd.Float64() * 100
	const fix = 15.6
	switch {
	case rnd < 32:
		return 49
	case rnd < 50+fix:
		return 99
	case rnd < 60+fix:
		return 149
	case rnd < 67+fix:
		return 199
	case rnd < 72+fix:
		return 249
	case rnd < 75+fix:
		return 299
	case rnd < 78+fix:
		return 349
	case rnd < 80+fix:
		return 399
	case rnd < 82+fix:
		return 449
	case rnd < 83+fix:
		return 499
	case rnd < 84+fix:
		return 549
	default:
		return maxVal
	}
}
