package rules

// WhitespaceCheckFilter ports org.languagetool.rules.WhitespaceCheckFilter.
// Keeps the match when the whitespace before the token at 1-based position
// is not equal to the expected whitespaceChar.
type WhitespaceCheckFilter struct{}

func NewWhitespaceCheckFilter() *WhitespaceCheckFilter {
	return &WhitespaceCheckFilter{}
}

// Accept returns true when the match should be kept (whitespace differs).
// position is 1-based into whitespaceBefore (same length as pattern tokens).
func (f *WhitespaceCheckFilter) Accept(whitespaceBefore []string, position int, whitespaceChar string) (keep bool, err string) {
	if position < 1 || position > len(whitespaceBefore) {
		return false, "Wrong position in WhitespaceCheckFilter"
	}
	return whitespaceBefore[position-1] != whitespaceChar, ""
}
