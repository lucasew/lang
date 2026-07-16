package de

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// PotentialCompoundFilter ports PotentialCompoundFilter suggestion building
// without tagger/speller: prefers joined form when short, else hyphenated.
type PotentialCompoundFilter struct {
	// JoinedIsValid optional; nil treats joined words as valid when len <= 20.
	JoinedIsValid func(joined string) bool
}

func NewPotentialCompoundFilter() *PotentialCompoundFilter {
	return &PotentialCompoundFilter{}
}

// Suggestions returns replacement strings for part1+part2.
func (f *PotentialCompoundFilter) Suggestions(part1, part2 string) []string {
	p1cap := capitalizeIfUniform(part1)
	p2low, p2cap := part2, part2
	if !isMixedOrAllUpper(part2) {
		p2low = strings.ToLower(part2)
		p2cap = tools.UppercaseFirstChar(strings.ToLower(part2))
	}
	if !isMixedOrAllUpper(part1) {
		p1cap = tools.UppercaseFirstChar(strings.ToLower(part1))
	}
	joined := p1cap + p2low
	hyphenated := p1cap + "-" + p2cap
	valid := f.JoinedIsValid
	if valid == nil {
		valid = func(s string) bool { return utf8.RuneCountInString(s) <= 20 }
	}
	var out []string
	if valid(joined) {
		if utf8.RuneCountInString(joined) > 20 {
			out = append(out, hyphenated)
		}
		out = append(out, joined)
	} else {
		out = append(out, hyphenated)
	}
	return out
}

func isMixedOrAllUpper(s string) bool {
	hasLower, hasUpper := false, false
	for _, r := range s {
		if unicode.IsLower(r) {
			hasLower = true
		}
		if unicode.IsUpper(r) {
			hasUpper = true
		}
	}
	if hasUpper && !hasLower {
		return true // all upper
	}
	return hasUpper && hasLower
}

func capitalizeIfUniform(s string) string {
	if isMixedOrAllUpper(s) {
		return s
	}
	return tools.UppercaseFirstChar(strings.ToLower(s))
}
