package tools

import (
	"unicode/utf16"
)

// TextMatch is the minimal match surface for CorrectTextFromMatches
// (avoids importing the rules package).
// FromPos/ToPos are UTF-16 code-unit offsets (Java String / RuleMatch positions).
type TextMatch struct {
	FromPos               int
	ToPos                 int
	SuggestedReplacements []string
}

// CorrectTextFromMatches ports Tools.correctTextFromMatches.
// Applies the first suggested replacement of each match, adjusting offsets as it goes.
// Positions and offset arithmetic use UTF-16 code units (Java StringBuilder indices).
func CorrectTextFromMatches(contents string, matches []TextMatch) string {
	if len(matches) == 0 {
		return contents
	}
	// Java StringBuilder is a sequence of UTF-16 code units.
	sb := utf16.Encode([]rune(contents))
	var errors []string
	for _, rm := range matches {
		if len(rm.SuggestedReplacements) == 0 {
			continue
		}
		if rm.FromPos < 0 || rm.ToPos > len(sb) || rm.FromPos > rm.ToPos {
			errors = append(errors, "")
			continue
		}
		errors = append(errors, string(utf16.Decode(sb[rm.FromPos:rm.ToPos])))
	}
	offset := 0
	counter := 0
	for _, rm := range matches {
		if len(rm.SuggestedReplacements) == 0 {
			continue
		}
		repl := rm.SuggestedReplacements[0]
		from := rm.FromPos - offset
		to := rm.ToPos - offset
		if from >= 0 && to >= from && to <= len(sb) &&
			counter < len(errors) && errors[counter] == string(utf16.Decode(sb[from:to])) {
			replU := utf16.Encode([]rune(repl))
			nb := make([]uint16, 0, len(sb)-(to-from)+len(replU))
			nb = append(nb, sb[:from]...)
			nb = append(nb, replU...)
			nb = append(nb, sb[to:]...)
			sb = nb
			// Java: offset += (toPos - fromPos) - replacements.get(0).length()
			// .length() is UTF-16 code units.
			offset += (rm.ToPos - rm.FromPos) - len(replU)
		}
		counter++
	}
	return string(utf16.Decode(sb))
}
