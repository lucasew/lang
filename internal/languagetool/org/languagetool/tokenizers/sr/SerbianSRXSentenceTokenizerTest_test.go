package sr

// Twin of languagetool-language-modules/sr/src/test/java/org/languagetool/tokenizers/sr/SerbianSRXSentenceTokenizerTest.java
// Java: TestTools.testSplit — join parts, tokenize, expect same parts (incl. trailing spaces).
// stokenizer = new SRXSentenceTokenizer(new Serbian())  // short code "sr"
// No setSingleLineBreaksMarksParagraph — use SRX defaults for Serbian.
import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// testSplitSR mirrors Java private testSplit → TestTools.testSplit(sentences, stokenizer).
func testSplitSR(t *testing.T, parts ...string) {
	t.Helper()
	tok := NewSerbianSRXSentenceTokenizer()
	// default paragraph mode — do NOT invent flags unless Java sets them
	joined := strings.Join(parts, "")
	got := tok.Tokenize(joined)
	require.Equal(t, parts, got, "joined=%q", joined)
}

// Port of SerbianSRXSentenceTokenizerTest.testTokenize — all active cases, exact equality.
func TestSerbianSRXSentenceTokenizer_Tokenize(t *testing.T) {
	// NOTE: sentences here need to end with a space character so they
	// have correct whitespace when appended:
	testSplitSR(t, "Ово је једна реченица. ")
	testSplitSR(t, "Ово је једна реченица. ", "И још једна.")
	testSplitSR(t, "Једна реченица! ", "Још једна.")
	// commented out in Java: testSplit("Ein Satz... ", "Noch einer.");
	testSplitSR(t, "На адреси http://www.gov.rs станује српска влада.")
	testSplitSR(t, "Писмо је стигло 3.10. пре подне.")
	testSplitSR(t, "Писмо је стигло 31.1. пре подне.")
	testSplitSR(t, "Писмо је стигло 3.10.2000 поподне.")
	testSplitSR(t, "Србија је под Турцима била од 14. до 19. века.")

	// Testing (non-)segmentation after Roman numerals
	testSplitSR(t, "Петар I, познат и као Петар Ослободилац.")
	testSplitSR(t, "Петар II, познат и као Петар Изгнаник.")
	testSplitSR(t, "Петар III, принц наследник.")
	testSplitSR(t, "Александар V Обреновић.")

	// Testing (non-)segmentation in dates and times
	testSplitSR(t, "Данас је 13.12.2004.")
	testSplitSR(t, "Данас је 13. децембар.")
	testSplitSR(t, "Видећемо се 29. фебруара.")
	testSplitSR(t, "Јесен стиже 23.09. поподне.")
	testSplitSR(t, "Жена стиже тачно у 17:00 кући.")

	testSplitSR(t, "Ренесанса је почела у 13. веку и бла бла трућ.")
	testSplitSR(t, "Све је почело у 13. или 14. веку и бла бла трућ.")
	testSplitSR(t, "Трајало је све од 13. до 14. века и бла бла трућ.")

	testSplitSR(t, "Ово је једна(!) реченица.")
	testSplitSR(t, "Чуо сам једну(!!) реченицу.")
	testSplitSR(t, "Ово је једна(?) реченица.")
	testSplitSR(t, "Чујем ли само једну(???) реченицу.")

	testSplitSR(t, "„Ћуко је креп'о“, рече он")

	// commented out in Java: testSplit("Поштовани господине тј. госпођо.");
}
