package tools

import (
	"bufio"
	"io"
	"regexp"
	"strings"
)

// MultiKeyProperties ports org.languagetool.tools.MultiKeyProperties.
// Duplicate keys merge values into a list.
type MultiKeyProperties struct {
	properties map[string][]string
}

// multiKeyEqSplit ports MultiKeyProperties: line.split("\\s*=\\s*") without UNICODE_CHARACTER_CLASS.
var multiKeyEqSplit = regexp.MustCompile(`[ \t\n\v\f\r]*=[ \t\n\v\f\r]*`)

// LoadMultiKeyProperties parses property-style lines "key = value" (# comments, no multiline).
// Java: line = scanner.nextLine().trim(); line.split("\\s*=\\s*"); require length 2.
func LoadMultiKeyProperties(r io.Reader) *MultiKeyProperties {
	p := &MultiKeyProperties{properties: map[string][]string{}}
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := JavaStringTrim(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := multiKeyEqSplit.Split(line, -1)
		// Java String.split limit 0 drops trailing empties; require exactly 2 parts.
		if len(parts) != 2 {
			continue
		}
		key, value := parts[0], parts[1]
		if key == "" {
			continue
		}
		p.properties[key] = append(p.properties[key], value)
	}
	return p
}

// GetProperty returns values for key, or nil if absent.
func (p *MultiKeyProperties) GetProperty(key string) []string {
	if p == nil {
		return nil
	}
	return p.properties[key]
}
