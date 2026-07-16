package translation

import (
	"regexp"
	"strings"
)

var (
	reBrackets   = regexp.MustCompile(`\[.*?\]`)
	reBraces     = regexp.MustCompile(`\{.*?\}`)
	reParens     = regexp.MustCompile(`\(.*?\)`)
	reAbbrev     = regexp.MustCompile(`/[A-Z]+/`)
	reAbbrevWord = regexp.MustCompile(` /[A-Z][a-z]+\.?/`)
	reAngle      = regexp.MustCompile(`<(.*)>`)
	reSpaces     = regexp.MustCompile(`\s+`)
)

// verbsWithTo — keep "to " after these prev words (Java BeoLingusTranslator.verbsWithTo subset).
var verbsWithTo = map[string]struct{}{
	"want": {}, "need": {}, "have": {}, "like": {}, "try": {},
}

// SplitAtSemicolon ports BeoLingusTranslator.splitAtSemicolon — split on "; " unless inside {...}.
func SplitAtSemicolon(s string) []string {
	parts := regexp.MustCompile(`;\s+`).Split(s, -1)
	var merged []string
	merging := false
	for _, item := range parts {
		item = strings.TrimSpace(item)
		openPos := strings.IndexByte(item, '{')
		closePos := strings.IndexByte(item, '}')
		if merging {
			if len(merged) == 0 {
				merged = append(merged, item)
			} else {
				merged[len(merged)-1] = merged[len(merged)-1] + "; " + item
			}
			if closePos >= 0 {
				merging = false
			}
			continue
		}
		if openPos > closePos {
			// ";" inside "{...}" — start merge
			merged = append(merged, item)
			merging = true
			continue
		}
		merged = append(merged, item)
	}
	return merged
}

// Split ports BeoLingusTranslator.split (currently just splitAtSemicolon).
func Split(s string) []string {
	return SplitAtSemicolon(s)
}

// CleanTranslationForReplace ports BeoLingusTranslator.cleanTranslationForReplace.
func CleanTranslationForReplace(s, prevWord string) string {
	clean := reBrackets.ReplaceAllString(s, "")
	clean = reBraces.ReplaceAllString(clean, "")
	clean = reParens.ReplaceAllString(clean, "")
	clean = strings.ReplaceAll(clean, "sth./sb.", "")
	clean = strings.ReplaceAll(clean, "sb./sth.", "")
	clean = strings.ReplaceAll(clean, "sth.", "")
	clean = strings.ReplaceAll(clean, "sb.", "")
	clean = reAbbrev.ReplaceAllString(clean, "")
	clean = reAbbrevWord.ReplaceAllString(clean, "")
	clean = strings.ReplaceAll(clean, "<> ", "")
	clean = reAngle.ReplaceAllString(clean, "")
	clean = reSpaces.ReplaceAllString(clean, " ")
	clean = strings.TrimSpace(clean)
	if prevWord == "to" && strings.HasPrefix(clean, "to ") {
		return clean[3:]
	}
	if prevWord != "to" && strings.HasPrefix(clean, "to ") {
		if _, ok := verbsWithTo[prevWord]; !ok {
			return clean[3:]
		}
	}
	return clean
}

// GetTranslationSuffix ports BeoLingusTranslator.getTranslationSuffix.
func GetTranslationSuffix(s string) string {
	var sb strings.Builder
	var lookingFor []byte
	contains := func(c byte) bool {
		for _, x := range lookingFor {
			if x == c {
				return true
			}
		}
		return false
	}
	remove := func(c byte) {
		out := lookingFor[:0]
		for _, x := range lookingFor {
			if x != c {
				out = append(out, x)
			}
		}
		lookingFor = out
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '[':
			lookingFor = append(lookingFor, ']')
		case ']':
			if contains(']') {
				sb.WriteByte(c)
				sb.WriteByte(' ')
				remove(']')
			}
		case '<':
			lookingFor = append(lookingFor, '>')
		case '>':
			if contains('>') {
				sb.WriteByte(c)
				sb.WriteByte(' ')
				remove('>')
			}
		case '(':
			lookingFor = append(lookingFor, ')')
		case ')':
			if contains(')') {
				sb.WriteByte(c)
				sb.WriteByte(' ')
				remove(')')
			}
		case '{':
			lookingFor = append(lookingFor, '}')
		case '}':
			if contains('}') {
				sb.WriteByte(c)
				sb.WriteByte(' ')
				remove('}')
			}
		}
		if len(lookingFor) > 0 {
			sb.WriteByte(c)
		}
	}
	return strings.TrimSpace(sb.String())
}
