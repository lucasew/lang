package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// SwissExpandLine ports SwissCompoundRule.SwissExpander — ß and ss variants.
func SwissExpandLine(line string) []string {
	if stringsContainsBeta(line) {
		return []string{line, replaceBeta(line)}
	}
	return []string{line}
}

func stringsContainsBeta(s string) bool {
	return len(s) > 0 && (indexBeta(s) >= 0)
}

func indexBeta(s string) int {
	for i, r := range s {
		if r == 'ß' {
			return i
		}
	}
	return -1
}

func replaceBeta(s string) string {
	out := make([]rune, 0, len(s)+4)
	for _, r := range s {
		if r == 'ß' {
			out = append(out, 's', 's')
		} else {
			out = append(out, r)
		}
	}
	return string(out)
}

// Ensure SwissExpandLine matches CompoundRuleData LineExpander signature.
var _ rules.LineExpander = SwissExpandLine
