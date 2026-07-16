package de

import (
	"regexp"
	"strings"
)

// InsertCommaFilter ports surface suggestion logic from InsertCommaFilter
// for simple 2-token replacements (always insert comma between parts).
type InsertCommaFilter struct{}

func NewInsertCommaFilter() *InsertCommaFilter {
	return &InsertCommaFilter{}
}

var insertCommaWS = regexp.MustCompile(`\s+`)

// Suggest rewrites a replacement string by inserting commas for short phrases.
func (f *InsertCommaFilter) Suggest(replacement string) []string {
	parts := insertCommaWS.Split(strings.TrimSpace(replacement), -1)
	var out []string
	switch len(parts) {
	case 2:
		out = append(out, parts[0]+", "+parts[1])
	case 3:
		// without POS: offer both common splits
		out = append(out, parts[0]+", "+parts[1]+" "+parts[2])
		out = append(out, parts[0]+" "+parts[1]+", "+parts[2])
	default:
		if replacement != "" {
			out = append(out, replacement)
		}
	}
	return out
}
