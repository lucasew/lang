package tools

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode/utf16"
)

// LoadJavaProperties loads Java .properties format (UTF-8 lines) used by LT
// MessagesBundle_*.properties. Supports:
//   - # / ! comments and blank lines
//   - key=value and key:value (first unescaped = or :)
//   - line continuation via trailing \ (Java Properties)
//   - \uXXXX unicode escapes and common \\, \n, \t, \r, \f in values
// Does not invent keys; incomplete features (ISO-8859-1 default encoding path)
// are explicit — LT ships UTF-8 bundles.
func LoadJavaProperties(r io.Reader) (map[string]string, error) {
	if r == nil {
		return nil, fmt.Errorf("nil reader")
	}
	out := map[string]string{}
	sc := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	lineNo := 0
	var logical strings.Builder
	for sc.Scan() {
		lineNo++
		raw := sc.Text()
		// strip UTF-8 BOM on first line
		if lineNo == 1 {
			raw = strings.TrimPrefix(raw, "\ufeff")
		}
		// continuation: append without trailing \ and without leading whitespace of next
		if logical.Len() > 0 {
			// continued line: trim leading whitespace per Java Properties
			raw = strings.TrimLeft(raw, " \t\f")
		} else {
			// new logical line: trim leading space only for comment detection
			trimmed := strings.TrimLeft(raw, " \t\f")
			if trimmed == "" || trimmed[0] == '#' || trimmed[0] == '!' {
				continue
			}
			raw = trimmed
		}

		// odd number of trailing backslashes → continuation
		if endsWithOddBackslash(raw) {
			// drop final \
			logical.WriteString(raw[:len(raw)-1])
			continue
		}
		logical.WriteString(raw)
		line := logical.String()
		logical.Reset()

		key, val, ok := splitPropertiesKeyValue(line)
		if !ok {
			// key with empty value (no separator)
			key = unescapeJavaProperties(strings.TrimSpace(line))
			if key == "" {
				return nil, fmt.Errorf("line %d: empty key", lineNo)
			}
			out[key] = ""
			continue
		}
		key = unescapeJavaProperties(strings.TrimSpace(key))
		// Java keeps value leading space after separator unless escaped; LT files use TrimSpace-friendly values
		val = unescapeJavaProperties(strings.TrimSpace(val))
		if key == "" {
			return nil, fmt.Errorf("line %d: empty key", lineNo)
		}
		out[key] = val
	}
	if logical.Len() > 0 {
		// dangling continuation at EOF — treat as complete line
		line := logical.String()
		key, val, ok := splitPropertiesKeyValue(line)
		if !ok {
			key = unescapeJavaProperties(strings.TrimSpace(line))
			if key != "" {
				out[key] = ""
			}
		} else {
			key = unescapeJavaProperties(strings.TrimSpace(key))
			val = unescapeJavaProperties(strings.TrimSpace(val))
			if key != "" {
				out[key] = val
			}
		}
	}
	return out, sc.Err()
}

func endsWithOddBackslash(s string) bool {
	n := 0
	for i := len(s) - 1; i >= 0 && s[i] == '\\'; i-- {
		n++
	}
	return n%2 == 1
}

// splitPropertiesKeyValue finds first unescaped = or : as separator.
func splitPropertiesKeyValue(line string) (key, val string, ok bool) {
	for i := 0; i < len(line); i++ {
		c := line[i]
		if c == '\\' {
			i++ // skip escaped char
			continue
		}
		if c == '=' || c == ':' {
			return line[:i], line[i+1:], true
		}
	}
	return "", "", false
}

// unescapeJavaProperties handles \uXXXX and common escapes (Java Properties).
func unescapeJavaProperties(s string) string {
	if !strings.Contains(s, `\`) {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' || i+1 >= len(s) {
			b.WriteByte(s[i])
			continue
		}
		i++
		switch s[i] {
		case 'u', 'U':
			// \uXXXX (may be surrogate pair if two consecutive)
			if i+4 < len(s) {
				hex := s[i+1 : i+5]
				if v, err := strconv.ParseUint(hex, 16, 16); err == nil {
					r := rune(v)
					// high surrogate: try pair
					if utf16.IsSurrogate(r) && i+10 < len(s) && s[i+5] == '\\' && (s[i+6] == 'u' || s[i+6] == 'U') {
						hex2 := s[i+7 : i+11]
						if v2, err2 := strconv.ParseUint(hex2, 16, 16); err2 == nil {
							b.WriteRune(utf16.DecodeRune(r, rune(v2)))
							i += 10
							continue
						}
					}
					b.WriteRune(r)
					i += 4
					continue
				}
			}
			b.WriteByte('u')
		case 'n':
			b.WriteByte('\n')
		case 't':
			b.WriteByte('\t')
		case 'r':
			b.WriteByte('\r')
		case 'f':
			b.WriteByte('\f')
		case '\\', '=', ':', '#', '!', ' ':
			b.WriteByte(s[i])
		default:
			// Java: unknown escape leaves the char
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

// ValidateTranslationKeys checks every English key exists in langProps.
// Returns missing keys (empty if complete).
func ValidateTranslationKeys(englishKeys, langProps map[string]string) []string {
	var missing []string
	for k := range englishKeys {
		if _, ok := langProps[k]; !ok {
			missing = append(missing, k)
		}
	}
	return missing
}

// ValidateTranslationsNotEmpty returns lines that have empty values (key present, empty string).
func ValidateTranslationsNotEmpty(props map[string]string) []string {
	var empty []string
	for k, v := range props {
		if strings.TrimSpace(v) == "" {
			empty = append(empty, k)
		}
	}
	return empty
}
