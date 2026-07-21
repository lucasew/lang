package commandline

import (
	"fmt"
	"io"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// LineChecker checks a single segment of text and returns matches.
type LineChecker func(segment string) ([]*rules.RuleMatch, error)

// SplitLines splits on \n (keeps empty lines as empty segments if keepEmpty).
func SplitLines(text string) []string {
	// normalize \r\n
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	return strings.Split(text, "\n")
}

// SplitParagraphs splits on blank lines (one or more).
// When singleLineBreakMarksParagraph is true, each non-empty line is a paragraph.
func SplitParagraphs(text string, singleLineBreakMarksParagraph bool) []string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	if singleLineBreakMarksParagraph {
		var out []string
		for _, line := range strings.Split(text, "\n") {
			// Java Main.isBreakPoint: "".equals(line) — exact empty only, not TrimSpace.
			if line == "" {
				continue
			}
			out = append(out, line)
		}
		return out
	}
	// double newline paragraphs — keep content; only drop exact-empty segments
	parts := strings.Split(text, "\n\n")
	var out []string
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// CheckLineByLine runs checker per line and prints with line offsets.
// Returns total match count.
func CheckLineByLine(w io.Writer, text string, check LineChecker) (int, error) {
	if w == nil {
		w = io.Discard
	}
	total := 0
	for i, line := range SplitLines(text) {
		// Java line-by-line: empty line is breakpoint only when "".equals(line);
		// whitespace-only lines are still content (not skipped via TrimSpace).
		if line == "" {
			continue
		}
		matches, err := check(line)
		if err != nil {
			return total, err
		}
		for _, m := range matches {
			if m == nil {
				continue
			}
			total++
			id := ruleIDOfMatch(m)
			fmt.Fprintf(w, "%d.) Line %d, column %d, Rule ID: %s\n",
				total, i+1, m.FromPos+1, id)
			if m.GetMessage() != "" {
				fmt.Fprintf(w, "Message: %s\n", m.GetMessage())
			}
		}
	}
	return total, nil
}

// CheckParagraphs runs checker per paragraph.
func CheckParagraphs(w io.Writer, text string, singleLineBreak bool, check LineChecker) (int, error) {
	if w == nil {
		w = io.Discard
	}
	total := 0
	paras := SplitParagraphs(text, singleLineBreak)
	for i, para := range paras {
		matches, err := check(para)
		if err != nil {
			return total, err
		}
		for _, m := range matches {
			if m == nil {
				continue
			}
			total++
			id := ruleIDOfMatch(m)
			fmt.Fprintf(w, "%d.) Paragraph %d, column %d, Rule ID: %s\n",
				total, i+1, m.FromPos+1, id)
			if m.GetMessage() != "" {
				fmt.Fprintf(w, "Message: %s\n", m.GetMessage())
			}
		}
	}
	return total, nil
}

func ruleIDOfMatch(m *rules.RuleMatch) string {
	if m == nil {
		return ""
	}
	if h, ok := m.GetRule().(interface{ GetID() string }); ok {
		return h.GetID()
	}
	return ""
}
