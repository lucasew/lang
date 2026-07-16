package patterns

import "strings"

// NodeAccept constants mirror LSParserFilter accept codes.
const (
	FilterAccept short = 1
	FilterReject short = 2
	FilterSkip   short = 3
)

type short = int

// IgnoreWhitespaceFilter ports
// org.languagetool.rules.patterns.IgnoreWhitespaceFilter for text nodes.
// Rejects whitespace-only text; accepts other content.
type IgnoreWhitespaceFilter struct{}

// AcceptText returns FilterReject when text is empty/whitespace-only.
func (IgnoreWhitespaceFilter) AcceptText(text string) int {
	if strings.TrimSpace(text) == "" {
		return FilterReject
	}
	return FilterAccept
}

// AcceptElement always accepts (Java startElement always FILTER_ACCEPT).
func (IgnoreWhitespaceFilter) AcceptElement(_ string) int {
	return FilterAccept
}

// FilterWhitespaceNodes drops empty/whitespace-only strings from a node list.
func FilterWhitespaceNodes(nodes []string) []string {
	var out []string
	f := IgnoreWhitespaceFilter{}
	for _, n := range nodes {
		if f.AcceptText(n) == FilterAccept {
			out = append(out, n)
		}
	}
	return out
}
