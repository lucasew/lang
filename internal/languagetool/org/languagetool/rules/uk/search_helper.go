package uk

import (
	"strings"
)

// SearchMatch is a simplified SearchHelper.Match for surface token sequences.
type SearchMatch struct {
	Targets       []string
	Limit         int  // max logical distance; -1 = unlimited
	IgnoreQuotes  bool // skip quote tokens
	IgnoreInserts bool // skip (...) spans
}

// NewSearchMatch builds a matcher for space-separated target tokens.
func NewSearchMatch(tokenLine string) *SearchMatch {
	line := strings.ReplaceAll(tokenLine, ",", " ,")
	parts := strings.Fields(line)
	return &SearchMatch{Targets: parts, Limit: -1, IgnoreQuotes: true}
}

// WithLimit sets the max steps scanned.
func (m *SearchMatch) WithLimit(n int) *SearchMatch {
	m.Limit = n
	return m
}

// MAfter finds targets in order starting at pos; returns start index of match or -1.
func (m *SearchMatch) MAfter(tokens []string, pos int) int {
	if len(m.Targets) == 0 {
		return -1
	}
	iCond := 0
	logical := 0
	start := -1
	for iCond < len(m.Targets) {
		if pos >= len(tokens) {
			return -1
		}
		if m.Limit > 0 && logical > m.Limit {
			return -1
		}
		logical++
		tok := tokens[pos]
		if m.IgnoreQuotes && QuotesPattern.MatchString(tok) {
			pos++
			continue
		}
		if m.IgnoreInserts && tok == "(" {
			// skip to matching )
			depth := 1
			pos++
			for pos < len(tokens) && depth > 0 {
				if tokens[pos] == "(" {
					depth++
				} else if tokens[pos] == ")" {
					depth--
				}
				pos++
			}
			continue
		}
		if !strings.EqualFold(tok, m.Targets[iCond]) {
			if start >= 0 {
				return -1 // broken sequence after first hit
			}
			pos++
			continue
		}
		if start < 0 {
			start = pos
		}
		iCond++
		pos++
	}
	return start
}

// MBefore finds targets in reverse order ending at pos; returns start index or -1.
func (m *SearchMatch) MBefore(tokens []string, pos int) int {
	if len(m.Targets) == 0 {
		return -1
	}
	iCond := len(m.Targets) - 1
	logical := 0
	end := pos
	for iCond >= 0 {
		if pos < 0 {
			return -1
		}
		if m.Limit > 0 && logical > m.Limit {
			return -1
		}
		logical++
		tok := tokens[pos]
		if m.IgnoreQuotes && QuotesPattern.MatchString(tok) {
			pos--
			continue
		}
		if !strings.EqualFold(tok, m.Targets[iCond]) {
			if iCond != len(m.Targets)-1 {
				return -1
			}
			pos--
			continue
		}
		iCond--
		if iCond < 0 {
			return pos
		}
		pos--
	}
	_ = end
	return -1
}
