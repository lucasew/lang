package tools

import (
	"bufio"
	"io"
	"strings"
)

// MultiKeyProperties ports org.languagetool.tools.MultiKeyProperties.
// Duplicate keys merge values into a list.
type MultiKeyProperties struct {
	properties map[string][]string
}

// LoadMultiKeyProperties parses property-style lines "key = value" (# comments, no multiline).
func LoadMultiKeyProperties(r io.Reader) *MultiKeyProperties {
	p := &MultiKeyProperties{properties: map[string][]string{}}
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// split on first = with optional spaces (Java: \\s*=\\s*)
		parts := splitProp(line)
		if len(parts) != 2 {
			continue
		}
		key, value := parts[0], parts[1]
		p.properties[key] = append(p.properties[key], value)
	}
	return p
}

func splitProp(line string) []string {
	i := strings.Index(line, "=")
	if i < 0 {
		return nil
	}
	key := strings.TrimSpace(line[:i])
	val := strings.TrimSpace(line[i+1:])
	if key == "" {
		return nil
	}
	return []string{key, val}
}

// GetProperty returns values for key, or nil if absent.
func (p *MultiKeyProperties) GetProperty(key string) []string {
	if p == nil {
		return nil
	}
	return p.properties[key]
}
