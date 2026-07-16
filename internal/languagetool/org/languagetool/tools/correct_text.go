package tools

// TextMatch is the minimal match surface for CorrectTextFromMatches
// (avoids importing the rules package).
type TextMatch struct {
	FromPos               int
	ToPos                 int
	SuggestedReplacements []string
}

// CorrectTextFromMatches ports Tools.correctTextFromMatches.
// Applies the first suggested replacement of each match, adjusting offsets as it goes.
func CorrectTextFromMatches(contents string, matches []TextMatch) string {
	if len(matches) == 0 {
		return contents
	}
	// Work on runes? Java uses UTF-16 indices. Our callers use byte indices for ASCII tests;
	// use byte indices for parity with simple ASCII (testCorrectTextFromMatches).
	sb := []byte(contents)
	var errors []string
	for _, rm := range matches {
		if len(rm.SuggestedReplacements) == 0 {
			continue
		}
		if rm.FromPos < 0 || rm.ToPos > len(sb) || rm.FromPos > rm.ToPos {
			errors = append(errors, "")
			continue
		}
		errors = append(errors, string(sb[rm.FromPos:rm.ToPos]))
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
			counter < len(errors) && errors[counter] == string(sb[from:to]) {
			// replace sb[from:to] with repl
			nb := make([]byte, 0, len(sb)- (to-from)+len(repl))
			nb = append(nb, sb[:from]...)
			nb = append(nb, repl...)
			nb = append(nb, sb[to:]...)
			sb = nb
			offset += (rm.ToPos - rm.FromPos) - len(repl)
		}
		counter++
	}
	return string(sb)
}
