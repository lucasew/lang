// Package ltgolden runs LanguageTool-style XML <example> cases against the Go engine.
package ltgolden

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lucasew/lang/internal/data"
	"github.com/lucasew/lang/internal/engine"
	"github.com/lucasew/lang/internal/finding"
)

// Case is one LT grammar example converted to a golden test.
type Case struct {
	RuleID string
	Text   string
	// Incorrect is true when correction attribute was set (bad example).
	Incorrect  bool
	Correction string
	HasMarker  bool
	MarkerFrom int // rune offset
	MarkerTo   int
	SourceFile string
}

// Result is the outcome of one case.
type Result struct {
	Case    Case
	Pass    bool
	Detail  string
	Matches []finding.Finding
}

// EnglishGrammarPaths lists en grammar XML under data root.
func EnglishGrammarPaths(dataRoot string) []string {
	base := filepath.Join(data.LanguageModules(dataRoot), "en", "src", "main", "resources",
		"org", "languagetool", "rules", "en")
	var paths []string
	_ = filepath.WalkDir(base, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.HasSuffix(d.Name(), ".xml") {
			paths = append(paths, path)
		}
		return nil
	})
	return paths
}

// RunCase checks one example against the engine.
// Incorrect: expect a match whose rule ID matches (base or full) and preferably overlaps the marker.
// Correct: expect no match for that rule ID.
func RunCase(c *engine.Checker, tc Case) Result {
	res := Result{Case: tc}
	opt := engine.Options{
		Language: "en-US",
		// Run all enabled rules; filtering by id is done after match.
		// Restricting to only the target rule can miss dependency/order effects, but
		// LT tests usually enable the rule under test among others. We run all.
	}
	out, err := c.Check("example.txt", tc.Text, opt)
	if err != nil {
		res.Detail = "engine error: " + err.Error()
		return res
	}
	var matches []finding.Finding
	for _, f := range out.Findings {
		if ruleMatch(tc.RuleID, f.Rule) {
			matches = append(matches, f)
		}
	}
	res.Matches = matches

	if !tc.Incorrect {
		// correct example: must not match this rule
		if len(matches) == 0 {
			res.Pass = true
			res.Detail = "no match (ok)"
		} else {
			res.Detail = fmt.Sprintf("false positive: %d match(es)", len(matches))
		}
		return res
	}

	// incorrect: need at least one match
	if len(matches) == 0 {
		res.Detail = "missed error (no match for rule)"
		return res
	}
	if !tc.HasMarker {
		res.Pass = true
		res.Detail = fmt.Sprintf("%d match(es)", len(matches))
		return res
	}
	// prefer span overlap with marker
	for _, m := range matches {
		if spansOverlap(m.Offset, m.EndOffset, tc.MarkerFrom, tc.MarkerTo) {
			res.Pass = true
			res.Detail = "match overlaps marker"
			return res
		}
	}
	// match exists but wrong span — still count as soft fail for 1:1
	res.Detail = fmt.Sprintf("%d match(es) but none overlap marker [%d,%d)", len(matches), tc.MarkerFrom, tc.MarkerTo)
	// For progress metric, treat as fail for 1:1
	return res
}

func ruleMatch(want, got string) bool {
	if want == got {
		return true
	}
	// want may be "ID" and got "ID[2]" or reverse
	if strings.HasPrefix(got, want+"[") {
		return true
	}
	if i := strings.IndexByte(want, '['); i > 0 {
		if want[:i] == got || strings.HasPrefix(got, want[:i]+"[") {
			return true
		}
	}
	if i := strings.IndexByte(got, '['); i > 0 && got[:i] == want {
		return true
	}
	return false
}

func spansOverlap(a0, a1, b0, b1 int) bool {
	return a0 < b1 && b0 < a1
}

// Summary counts.
type Summary struct {
	Total, Pass, Fail          int
	Incorrect, Correct         int
	IncorrectPass, CorrectPass int
}

func Summarize(results []Result) Summary {
	var s Summary
	for _, r := range results {
		s.Total++
		if r.Pass {
			s.Pass++
		} else {
			s.Fail++
		}
		if r.Case.Incorrect {
			s.Incorrect++
			if r.Pass {
				s.IncorrectPass++
			}
		} else {
			s.Correct++
			if r.Pass {
				s.CorrectPass++
			}
		}
	}
	return s
}
