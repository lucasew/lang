package eval

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/dev/dumpcheck"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// AfterTheDeadlineEvaluator ports org.languagetool.dev.eval.AfterTheDeadlineEvaluator
// (HTTP optional via Query inject).
type AfterTheDeadlineEvaluator struct {
	URLPrefix string
	// Query returns AtD XML for a sentence; nil → never finds errors.
	Query func(sentence string) (string, error)
}

func NewAfterTheDeadlineEvaluator(urlPrefix string) *AfterTheDeadlineEvaluator {
	return &AfterTheDeadlineEvaluator{URLPrefix: urlPrefix}
}

// IsExpectedErrorFound ports AfterTheDeadlineEvaluator.isExpectedErrorFound.
func (e *AfterTheDeadlineEvaluator) IsExpectedErrorFound(example rules.IncorrectExample, resultXML string) (bool, error) {
	matches, err := dumpcheck.ParseAtDResultXML(resultXML)
	if err != nil {
		return false, err
	}
	ex := example.GetExample()
	expectedStart := strings.Index(ex, "<marker>")
	if expectedStart < 0 {
		return false, nil
	}
	clean := rules.CleanMarkersInExample(ex)
	for _, m := range matches {
		errorStr := m.String
		if errorStr == "" {
			continue
		}
		for _, start := range startPositions(clean, errorStr) {
			end := start + len(errorStr)
			// Java compares to expectedStart from dirty string; start positions match for leading markers.
			expectedEnd := start + len(errorStr) // always equals end
			if start == expectedStart && end == expectedEnd {
				return true, nil
			}
		}
	}
	return false, nil
}

func startPositions(sentence, search string) []int {
	var out []int
	pos := 0
	for {
		i := strings.Index(sentence[pos:], search)
		if i < 0 {
			break
		}
		abs := pos + i
		out = append(out, abs)
		pos = abs + 1
		if pos >= len(sentence) {
			break
		}
	}
	return out
}

// QueryExample cleans markers and queries AtD (or inject).
func (e *AfterTheDeadlineEvaluator) QueryExample(example rules.IncorrectExample) (bool, error) {
	sentence := rules.CleanMarkersInExample(example.GetExample())
	if e.Query == nil {
		return false, nil
	}
	xml, err := e.Query(sentence)
	if err != nil {
		return false, err
	}
	return e.IsExpectedErrorFound(example, xml)
}
