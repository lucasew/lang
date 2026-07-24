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

// SplitParagraphs ports Main line-by-line paragraph accumulation for bulk splits.
// Java Main.runOnFile: readLine, append line+"\n", break when isBreakPoint(line).
// isBreakPoint: singleLineBreakMarksPara || "".equals(line) (exact empty, not TrimSpace).
// Each emitted paragraph includes the trailing '\n' after each read line (Java sb).
func SplitParagraphs(text string, singleLineBreakMarksParagraph bool) []string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	// Do not use strings.Split alone for the double-newline invent path — mirror Main's
	// BufferedReader.readLine loop so trailing newlines match Java StringBuilder content.
	var out []string
	var sb strings.Builder
	// strings.Split drops a trailing empty only when text ends with \n — readLine semantics
	// for "a\nb\n" yield lines "a","b"; for "a\nb" yield "a","b". Use Split and treat
	// trailing empty from a final \n as no extra readLine.
	rawLines := strings.Split(text, "\n")
	// If text ends with \n, Split yields a final "" that is not a readLine result.
	if strings.HasSuffix(text, "\n") && len(rawLines) > 0 {
		rawLines = rawLines[:len(rawLines)-1]
	}
	for _, line := range rawLines {
		sb.WriteString(line)
		sb.WriteByte('\n')
		// Java isBreakPoint: singleLineBreakMarksPara || "".equals(line)
		if singleLineBreakMarksParagraph || line == "" {
			if sb.Len() > 0 {
				out = append(out, sb.String())
				sb.Reset()
			}
		}
	}
	if sb.Len() > 0 {
		out = append(out, sb.String())
	}
	// Empty-line breakpoints produce a paragraph that is just "\n" (the empty line's
	// appended newline). Java still calls handleLine on that; for SplitParagraphs
	// consumers that want content paragraphs only, drop pure-empty-break segments
	// that are only newlines from empty breakpoints when they are sole content.
	// Keep Java-faithful: include all handleLine payloads (including "\n" only).
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
