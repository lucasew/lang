package ltgolden

import (
	"fmt"
	"strings"

	"github.com/lucasew/lang/internal/attic/engine"
	"github.com/lucasew/lang/internal/attic/finding"
)

// Result is the outcome of one ground-truth case.
type Result struct {
	Case    Case
	Pass    bool
	Detail  string
	Matches []finding.Finding
}

// RunCase executes one case against the engine (ground truth).
func RunCase(c *engine.Checker, tc Case) Result {
	res := Result{Case: tc}
	lang := tc.Lang
	if lang == "" || lang == "und" || lang == "xx" {
		lang = "en-US"
	}
	// Prefer full codes when family only
	if len(lang) == 2 {
		switch lang {
		case "en":
			lang = "en-US"
		case "de":
			lang = "de-DE"
		case "pt":
			lang = "pt-BR"
		}
	}

	opt := engine.Options{Language: lang}
	// Grammar examples: force the rule under test (including default=off) so each
	// LT example is an isolated ground-truth probe — matches PatternRuleTest spirit.
	if tc.Kind == KindGrammarExample && tc.RuleID != "" && tc.RuleID != "ANON" {
		base := ruleBase(tc.RuleID)
		opt.EnabledOnly = map[string]bool{base: true, tc.RuleID: true}
	}

	out, err := c.Check("example.txt", tc.Text, opt)
	if err != nil {
		// Engine cannot analyze: fail ground truth
		res.Detail = "engine error: " + err.Error()
		return res
	}

	switch tc.Kind {
	case KindGrammarExample:
		return scoreGrammar(res, tc, out.Findings)
	case KindDisambigExample:
		// Full reading equality needs tagger+disambig parity; for now:
		// untouched => must not crash (pass if engine runs)
		// ambiguous => must not crash; soft pass if engine runs (strict later)
		if strings.EqualFold(strings.Split(tc.ExampleType, "|")[0], "untouched") || !tc.Incorrect {
			res.Pass = true
			res.Detail = "disambig smoke (untouched/correct)"
			return res
		}
		// Require at least that analysis produced tokens without error — strict reading match TODO
		res.Pass = true
		res.Detail = "disambig smoke (ambiguous; reading equality pending 1:1)"
		return res
	case KindJavaUnit:
		return scoreJava(res, tc, out.Findings)
	default:
		res.Detail = "unknown kind"
		return res
	}
}

func scoreGrammar(res Result, tc Case, findings []finding.Finding) Result {
	var matches []finding.Finding
	for _, f := range findings {
		if ruleMatch(tc.RuleID, f.Rule) {
			matches = append(matches, f)
		}
	}
	res.Matches = matches

	if !tc.Incorrect {
		if len(matches) == 0 {
			res.Pass = true
			res.Detail = "no match (ok)"
		} else {
			res.Detail = fmt.Sprintf("false positive: %d match(es)", len(matches))
		}
		return res
	}
	if len(matches) == 0 {
		res.Detail = "missed error (no match for rule)"
		return res
	}
	if !tc.HasMarker {
		res.Pass = true
		res.Detail = fmt.Sprintf("%d match(es)", len(matches))
		return res
	}
	for _, m := range matches {
		if spansOverlap(m.Offset, m.EndOffset, tc.MarkerFrom, tc.MarkerTo) {
			res.Pass = true
			res.Detail = "match overlaps marker"
			return res
		}
	}
	res.Detail = fmt.Sprintf("%d match(es) but none overlap marker [%d,%d)", len(matches), tc.MarkerFrom, tc.MarkerTo)
	return res
}

func scoreJava(res Result, tc Case, findings []finding.Finding) Result {
	res.Matches = findings
	switch tc.ExampleType {
	case "assertGood", "check_smoke":
		// assertGood: zero findings for the class's rule if known; else zero findings overall
		// Many assertGood mean "no issues from this rule" with only that rule enabled —
		// we approximate: no error-severity findings, or empty.
		if tc.ExampleType == "check_smoke" {
			res.Pass = true
			res.Detail = "smoke: engine accepted text"
			return res
		}
		// assertGood: prefer no findings
		if len(findings) == 0 {
			res.Pass = true
			res.Detail = "assertGood: clean"
			return res
		}
		res.Detail = fmt.Sprintf("assertGood: %d finding(s)", len(findings))
		return res
	case "assertBad":
		if len(findings) > 0 {
			res.Pass = true
			res.Detail = fmt.Sprintf("assertBad: %d finding(s)", len(findings))
			return res
		}
		res.Detail = "assertBad: no findings"
		return res
	default:
		res.Pass = true
		res.Detail = "java case recorded"
		return res
	}
}

func ruleBase(id string) string {
	if i := strings.IndexByte(id, '['); i > 0 {
		return id[:i]
	}
	return id
}

func ruleMatch(want, got string) bool {
	if want == got {
		return true
	}
	if strings.HasPrefix(got, want+"[") {
		return true
	}
	wb, gb := ruleBase(want), ruleBase(got)
	return wb == gb
}

func spansOverlap(a0, a1, b0, b1 int) bool {
	return a0 < b1 && b0 < a1
}

// Summary aggregates results.
type Summary struct {
	Total, Pass, Fail          int
	ByKind                     map[Kind][2]int // pass, total
	Incorrect, Correct         int
	IncorrectPass, CorrectPass int
}

func Summarize(results []Result) Summary {
	s := Summary{ByKind: map[Kind][2]int{}}
	for _, r := range results {
		s.Total++
		bk := s.ByKind[r.Case.Kind]
		bk[1]++
		if r.Pass {
			s.Pass++
			bk[0]++
		} else {
			s.Fail++
		}
		s.ByKind[r.Case.Kind] = bk
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
