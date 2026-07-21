package tagging

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// ManualTagger ports org.languagetool.tagging.ManualTagger (map-based; same lookups).
type ManualTagger struct {
	// form (as in file) → readings
	byForm map[string][]TaggedWord
}

func NewManualTagger(r io.Reader) (*ManualTagger, error) {
	mapping, err := loadMapping(r)
	if err != nil {
		return nil, err
	}
	return &ManualTagger{byForm: mapping}, nil
}

func loadMapping(r io.Reader) (map[string][]TaggedWord, error) {
	m := map[string][]TaggedWord{}
	sc := bufio.NewScanner(r)
	lineCount := 0
	separator := "\t"
	for sc.Scan() {
		// Java ManualTagger.loadMapping: line.trim() (String.trim, not Unicode TrimSpace).
		line := tools.JavaStringTrim(sc.Text())
		lineCount++
		if strings.HasPrefix(line, "#separatorRegExp=") {
			separator = strings.TrimPrefix(line, "#separatorRegExp=")
			continue
		}
		if tools.IsEmptyStr(line) || line[0] == '#' {
			continue
		}
		if strings.Contains(line, "\u00A0") {
			return nil, fmt.Errorf("Non-breaking space found in line #%d: '%s', please remove it", lineCount, line)
		}
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = tools.JavaStringTrim(line[:i])
		}
		parts := strings.Split(line, separator)
		if len(parts) != 3 {
			return nil, fmt.Errorf("Unknown line format in line %d when loading manual tagger dictionary, expected three tab-separated fields: '%s'", lineCount, line)
		}
		form := parts[0]
		lemma := parts[1]
		if lemma == form {
			lemma = form
		}
		// Java: parts[2].trim() for POS (if present in stream)
		tag := tools.JavaStringTrim(parts[2])
		m[form] = append(m[form], NewTaggedWord(lemma, tag))
	}
	return m, sc.Err()
}

// Tag looks up a word's lemma and POS (word is typically lowercased by adapter).
func (t *ManualTagger) Tag(word string) []TaggedWord {
	return append([]TaggedWord(nil), t.byForm[word]...)
}
