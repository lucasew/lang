package uk

import "testing"

// Twin of UkrainianWordTokenizerTest n-dash cases (Java N_DASH_SPACE_PATTERN).
func TestUkrainianWordTokenizer_NDashSpace(t *testing.T) {
	// Java: "Стрий– " → break after letter (not followed by та|чи|і|й + whitespace)
	assertTok(t, "Стрий– ", "Стрий", "–", " ")
	// Java: "фіто– та термотерапії" → keep фіто– (lookahead blocks on "та ")
	assertTok(t, "фіто– та термотерапії", "фіто–", " ", "та", " ", "термотерапії")
	// Java: " –Виділено" (N_DASH_SPACE_PATTERN2)
	assertTok(t, " –Виділено", " ", "–", "Виділено")
	// Java: "Х–ХІ" (DASH_NUMBERS)
	assertTok(t, "Х–ХІ", "Х", "–", "ХІ")
	// Java negative lookahead needs (та|чи|і|й)[\h\v] — conjunction alone at EOS does not block
	assertTok(t, "слово– та", "слово", "–", " ", "та")
}
