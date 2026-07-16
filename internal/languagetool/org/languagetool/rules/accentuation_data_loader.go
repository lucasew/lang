package rules

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// AccentuationDataLoader ports CA ConfusionPairsDataLoader / ES / PT PortugueseAccentuationDataLoader.
// File format: wrongForm;correctToken;POS (UTF-8), # comments and blank lines ignored.
//
// When allowMultiReadings is true (Catalan/ES loaders), additional lines for the same key
// append readings; when false (PortugueseAccentuationDataLoader), later lines replace.
type AccentuationDataLoader struct {
	AllowMultiReadings bool
}

func NewAccentuationDataLoader(allowMulti bool) *AccentuationDataLoader {
	return &AccentuationDataLoader{AllowMultiReadings: allowMulti}
}

// LoadWords parses stream into wrong-form → AnalyzedTokenReadings of the correct form.
func (l *AccentuationDataLoader) LoadWords(r io.Reader, pathHint string) (map[string]*languagetool.AnalyzedTokenReadings, error) {
	if pathHint == "" {
		pathHint = "accentuation data"
	}
	m := map[string]*languagetool.AnalyzedTokenReadings{}
	sc := bufio.NewScanner(r)
	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := strings.TrimSpace(sc.Text())
		if line == "" || line[0] == '#' {
			continue
		}
		parts := strings.Split(line, ";")
		if len(parts) != 3 {
			return nil, fmt.Errorf("format error in file %s, line: %s, expected 3 semicolon-separated parts, got %d",
				pathHint, line, len(parts))
		}
		key := parts[0]
		tok := parts[1]
		pos := parts[2]
		analyzed := languagetool.NewAnalyzedToken(tok, &pos, nil)
		if existing, ok := m[key]; ok && l.AllowMultiReadings {
			existing.AddReading(analyzed, "")
		} else {
			m[key] = languagetool.NewAnalyzedTokenReadings(analyzed)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return m, nil
}
