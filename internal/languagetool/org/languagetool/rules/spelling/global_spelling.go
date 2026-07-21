package spelling

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// Prohibited global-spelling entries (GlobalSpellingTest.avoidSomeWords).
var (
	prohibitedGlobalExpressions = []string{"Dnipro", "Dnepr"}
	prohibitedGlobalTokens      = []string{"Tolstoi", "Tolstoy", "Dostoevsky"}
)

// ValidateGlobalSpellingLines ports GlobalSpellingTest.avoidSomeWords checks
// over already-loaded spelling_global.txt lines.
func ValidateGlobalSpellingLines(lines []string) error {
	for _, line := range lines {
		parts := strings.Split(line, "#")
		if len(parts) == 0 {
			continue
		}
		// Java String.trim / typical word-list load — not Unicode TrimSpace.
		entry := tools.JavaStringTrim(parts[0])
		if entry == "" {
			continue
		}
		for _, p := range prohibitedGlobalExpressions {
			if entry == p {
				return fmt.Errorf("Do not use '%s' in global_spelling.txt. It is not a valid spelling for all languages.", entry)
			}
		}
		// Split on single ASCII space like phrase tokens in spelling files.
		for _, token := range strings.Split(entry, " ") {
			if token == "" {
				continue
			}
			for _, p := range prohibitedGlobalTokens {
				if token == p {
					return fmt.Errorf("Do not use '%s' in global_spelling.txt. It is not a valid spelling for all languages.", token)
				}
			}
		}
	}
	return nil
}

// ValidateGlobalSpellingReader reads lines from r and validates them.
func ValidateGlobalSpellingReader(r io.Reader) error {
	if r == nil {
		return fmt.Errorf("nil reader")
	}
	var lines []string
	sc := bufio.NewScanner(r)
	// increase buffer for long phrase lines
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	if err := sc.Err(); err != nil {
		return err
	}
	return ValidateGlobalSpellingLines(lines)
}
