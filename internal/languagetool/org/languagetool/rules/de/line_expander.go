package de

import (
	"strings"
)

// LineExpander ports org.languagetool.rules.de.LineExpander for suffix flags
// (without German synthesizer for verb-prefix "_" lines).
type LineExpander struct{}

func NewLineExpander() *LineExpander { return &LineExpander{} }

func (e *LineExpander) ExpandLine(line string) []string {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return nil
	}
	if isLineWithVerbPrefix(line) {
		// without synthesizer: emit base pieces only
		return handleLineWithPrefixSimple(line)
	}
	if isLineWithFlag(line) {
		return handleLineWithFlags(line)
	}
	return []string{cleanTagsAndEscapeChars(line)}
}

func isLineWithFlag(line string) bool {
	idx := strings.IndexByte(line, '/')
	return idx > 0 && line[idx-1] != '\\'
}

func isLineWithVerbPrefix(line string) bool {
	idx := strings.IndexByte(line, '_')
	return idx > 0 && line[idx-1] != '\\'
}

func handleLineWithPrefixSimple(line string) []string {
	// "weiter_gehen" → weitergehen, zuweitergehen-style stubs without full synthesis
	clean := cleanTagsAndEscapeChars(line)
	parts := strings.Split(clean, "_")
	if len(parts) != 2 {
		return []string{strings.ReplaceAll(clean, "_", "")}
	}
	joined := parts[0] + parts[1]
	return []string{joined, parts[0] + "zu" + parts[1]}
}

func handleLineWithFlags(line string) []string {
	clean := cleanTagsAndEscapeChars(line)
	parts := strings.SplitN(clean, "/", 2)
	if len(parts) != 2 {
		return []string{clean}
	}
	word, suffix := parts[0], parts[1]
	var result []string
	add := func(w string) {
		for _, x := range result {
			if x == w {
				return
			}
		}
		result = append(result, w)
	}
	for _, c := range suffix {
		switch c {
		case 'S':
			add(word)
			add(word + "s")
		case 'N':
			add(word)
			add(word + "n")
		case 'E':
			add(word)
			add(word + "e")
		case 'F':
			add(word)
			add(word + "in")
		case 'T':
			add(word)
			if strings.HasSuffix(word, "straße") || strings.HasSuffix(word, "strasse") {
				add(strings.ReplaceAll(strings.ReplaceAll(word, "straße", "str."), "strasse", "str."))
			}
			if strings.HasSuffix(word, "Straße") || strings.HasSuffix(word, "Strasse") {
				add(strings.ReplaceAll(strings.ReplaceAll(word, "Straße", "Str."), "Strasse", "Str."))
			}
		case 'A', 'P':
			add(word)
			if strings.HasSuffix(word, "e") {
				add(word + "r")
				add(word + "s")
				add(word + "n")
				add(word + "m")
			} else {
				add(word + "e")
				add(word + "er")
				add(word + "es")
				add(word + "en")
				add(word + "em")
			}
		}
	}
	if len(result) == 0 {
		return []string{word}
	}
	return result
}

func cleanTagsAndEscapeChars(s string) string {
	if idx := strings.IndexByte(s, '#'); idx >= 0 {
		s = s[:idx]
	}
	return strings.TrimSpace(strings.ReplaceAll(s, "\\", ""))
}
