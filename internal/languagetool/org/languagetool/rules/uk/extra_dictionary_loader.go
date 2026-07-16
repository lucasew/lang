package uk

import (
	"bufio"
	"io"
	"strings"
)

// ExtraDictionaryLoader ports org.languagetool.rules.uk.ExtraDictionaryLoader.
// Methods take an io.Reader instead of a resource path (pluggable data).
type ExtraDictionaryLoader struct{}

// LoadSet loads non-comment lines into a set.
func LoadSet(r io.Reader) (map[string]struct{}, error) {
	out := map[string]struct{}{}
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		out[line] = struct{}{}
	}
	return out, sc.Err()
}

// LoadMap loads lines as "key value..." maps (first space-separated field → rest or "").
func LoadMap(r io.Reader) (map[string]string, error) {
	set, err := LoadSet(r)
	if err != nil {
		return nil, err
	}
	m := make(map[string]string, len(set))
	for line := range set {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, " ")
		if len(parts) > 1 {
			m[parts[0]] = parts[1]
		} else {
			m[parts[0]] = ""
		}
	}
	return m, nil
}

// LoadSpacedLists loads "key a b|c" → key → [a,b,c] (space or | separators after key).
func LoadSpacedLists(r io.Reader) (map[string][]string, error) {
	result := map[string][]string{}
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}
		if i := strings.Index(line, "#"); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		// split on space or |
		fields := splitSpaceOrPipe(line)
		if len(fields) == 0 {
			continue
		}
		if len(fields) == 1 {
			result[fields[0]] = nil
			continue
		}
		result[fields[0]] = append([]string(nil), fields[1:]...)
	}
	return result, sc.Err()
}

// LoadLists loads "key = a|b|c" (or key=a|b) from a rules-dir style stream.
func LoadLists(r io.Reader) (map[string][]string, error) {
	result := map[string][]string{}
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}
		// split on " *= *" or "|"
		parts := splitListLine(line)
		if len(parts) == 0 {
			continue
		}
		if len(parts) == 1 {
			result[parts[0]] = nil
			continue
		}
		result[parts[0]] = append([]string(nil), parts[1:]...)
	}
	return result, sc.Err()
}

func splitSpaceOrPipe(line string) []string {
	// Java: line.split(" |\\|")
	var out []string
	var b strings.Builder
	flush := func() {
		if b.Len() > 0 {
			out = append(out, b.String())
			b.Reset()
		}
	}
	for _, r := range line {
		if r == ' ' || r == '|' {
			flush()
			continue
		}
		b.WriteRune(r)
	}
	flush()
	return out
}

func splitListLine(line string) []string {
	// Java: line.split(" *= *|\\|")
	// First find " = " style or "=" with optional spaces
	var parts []string
	// replace = surrounded by spaces as separator, and |
	rest := line
	// Find first = as key separator if present
	if idx := indexEqualsSep(rest); idx >= 0 {
		key := strings.TrimSpace(rest[:idx])
		rest = strings.TrimSpace(rest[idx+1:])
		parts = append(parts, key)
		// rest split on |
		if rest == "" {
			return parts
		}
		for _, p := range strings.Split(rest, "|") {
			parts = append(parts, p) // Java keeps empty? Split keeps empties; trim not in Java for |
		}
		return parts
	}
	// no = : split only on |
	return strings.Split(line, "|")
}

func indexEqualsSep(s string) int {
	// match " *= *" — find '=' and allow spaces around
	for i, r := range s {
		if r == '=' {
			return i
		}
	}
	return -1
}
