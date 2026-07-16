package rules

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode"
)

// WordListValidator ports checks used by WordListValidatorTest:
// reject empty lines with content issues, tabs, trailing spaces, etc.
type WordListValidator struct {
	// AllowComments when true, # lines are ignored.
	AllowComments bool
}

func NewWordListValidator() *WordListValidator {
	return &WordListValidator{AllowComments: true}
}

// ValidateLines checks each non-comment line of a spelling/word list.
func (v *WordListValidator) ValidateLines(r io.Reader) []error {
	if r == nil {
		return []error{fmt.Errorf("nil reader")}
	}
	var errs []error
	sc := bufio.NewScanner(r)
	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := sc.Text()
		if v.AllowComments {
			trim := strings.TrimSpace(line)
			if trim == "" || strings.HasPrefix(trim, "#") {
				continue
			}
		}
		if line != strings.TrimSpace(line) {
			errs = append(errs, fmt.Errorf("line %d: leading/trailing whitespace: %q", lineNo, line))
		}
		if strings.Contains(line, "\t") {
			errs = append(errs, fmt.Errorf("line %d: contains tab: %q", lineNo, line))
		}
		// reject control chars except nothing — keep simple
		for _, r := range line {
			if unicode.IsControl(r) && r != '\t' {
				errs = append(errs, fmt.Errorf("line %d: control character in %q", lineNo, line))
				break
			}
		}
	}
	if err := sc.Err(); err != nil {
		errs = append(errs, err)
	}
	return errs
}

// LoadCustomSpellingWords loads spelling_custom.txt style lists (one word per line).
func LoadCustomSpellingWords(r io.Reader) ([]string, error) {
	if r == nil {
		return nil, fmt.Errorf("nil reader")
	}
	var words []string
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// strip trailing comments
		if i := strings.Index(line, "#"); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		if line != "" {
			words = append(words, line)
		}
	}
	return words, sc.Err()
}
