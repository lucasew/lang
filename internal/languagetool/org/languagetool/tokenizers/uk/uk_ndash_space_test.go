package uk

import "testing"

// Extra N_DASH_SPACE_PATTERN edge not listed as its own Java assertEquals but required by the
// Java pattern (?!(та|чи|і|й)[\h\v]) — conjunction alone at EOS does not block the break.
// Core Java testDash n-dash cases live in TestUkrainianWordTokenizer_Dash.
func TestUkrainianWordTokenizer_NDashSpace(t *testing.T) {
	assertTok(t, "слово– та", "слово", "–", " ", "та")
}
