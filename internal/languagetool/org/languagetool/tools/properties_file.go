package tools

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// LoadJavaProperties loads a subset of Java .properties format (UTF-8 lines).
// Supports #/! comments, key=value and key:value; ignores empty keys.
func LoadJavaProperties(r io.Reader) (map[string]string, error) {
	if r == nil {
		return nil, fmt.Errorf("nil reader")
	}
	out := map[string]string{}
	sc := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := strings.TrimSpace(sc.Text())
		if line == "" || line[0] == '#' || line[0] == '!' {
			continue
		}
		// continuation lines with trailing \ not fully handled — LT files rarely need them
		sep := strings.IndexAny(line, "=:")
		if sep < 0 {
			// key with empty value
			out[line] = ""
			continue
		}
		key := strings.TrimSpace(line[:sep])
		val := strings.TrimSpace(line[sep+1:])
		if key == "" {
			return nil, fmt.Errorf("line %d: empty key", lineNo)
		}
		out[key] = val
	}
	return out, sc.Err()
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
