package rules

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// LoadSimpleReplaceWords ports SimpleReplaceDataLoader.loadWords from a reader.
// Format: wrong=right|right2  (pipes for multi forms); # comments.
func LoadSimpleReplaceWords(r io.Reader) (map[string][]string, error) {
	m := make(map[string][]string)
	sc := bufio.NewScanner(r)
	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := sc.Text()
		if line == "" || line[0] == '#' {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("simple replace line %d: expected word=replacement", lineNo)
		}
		if strings.TrimSpace(parts[1]) == "" {
			return nil, fmt.Errorf("simple replace line %d: empty replacement", lineNo)
		}
		wrongForms := strings.Split(parts[0], "|")
		replacements := strings.Split(parts[1], "|")
		for _, w := range wrongForms {
			// copy replacements slice
			reps := append([]string(nil), replacements...)
			m[w] = reps
		}
	}
	return m, sc.Err()
}
