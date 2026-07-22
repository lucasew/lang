package sv

// Twin of languagetool-language-modules/sv/src/test/java/org/languagetool/tokenizers/sv/SwedishSRXSentenceTokenizerTest.java
// Java: TestTools.testSplit — join parts, tokenize, expect same parts (incl. trailing spaces).
// stokenizer = new SRXSentenceTokenizer(new Swedish())  // short code "sv"
// No setSingleLineBreaksMarksParagraph — use SRX defaults for Swedish.
import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// testSplitSV mirrors Java private testSplit → TestTools.testSplit(sentences, stokenizer).
func testSplitSV(t *testing.T, parts ...string) {
	t.Helper()
	tok := NewSwedishSRXSentenceTokenizer()
	// default paragraph mode — do NOT invent flags unless Java sets them
	joined := strings.Join(parts, "")
	got := tok.Tokenize(joined)
	require.Equal(t, parts, got, "joined=%q", joined)
}

// Port of SwedishSRXSentenceTokenizerTest.testTokenize — all active cases, exact equality.
func TestSwedishSRXSentenceTokenizer_Tokenize(t *testing.T) {
	// NOTE: sentences here need to end with a space character so they
	// have correct whitespace when appended:
	testSplitSV(t,
		"Onkel Toms stuga är en roman skriven av Harriet Beecher Stowe, publicerad den 1852. ",
		"Den handlar om slaveriet i USA sett ur slavarnas perspektiv och bidrog starkt till att slaveriet avskaffades 1865 efter amerikanska inbördeskriget.",
	)

	// Second Java testSplit: many strings as SEPARATE sentences that must each stay unsplit
	// (abbrev list). Concatenated they re-split into the same parts.
	testSplitSV(t,
		"Vi kan leverera varorna alt. snabbt med expressfrakt.",
		"Art. handlar om klimatförändringar i Norden.",
		"Bil. innehåller mer detaljerad information.",
		"Rapporten innehåller bl.a. statistik och tabeller.",
		"Han kom sent, dvs. att mötet redan hade börjat.",
		"Vi behöver hammare, spik, etc. verktyg för arbetet.",
		"Hör av dig vid ev. problem med datorn.",
		"Den f.d. kollega arbetar nu på ett nytt företag.",
		"Fig. visar sambandet mellan tid och temperatur.",
		"Berättelsen slutar med texten “forts. följer”.",
		"Avtalet gäller fr.o.m. måndag.",
		"Priset anges inkl. moms.",
		"Kol. visar medelvärden för varje grupp.",
		"Resultatet är m.a.o. fel.",
		"Orig. version sparades i arkivet.",
		"Vi diskuterade mål, strategier, osv. detaljer.",
		"Han var frånvarande p.g.a. sjukdom.",
		"Vi har satt ett prel. schema för veckan.",
		"Ref. hänvisar till tidigare forskning.",
		"De har resp. ansvar för olika områden.",
		"Han kallar sig en s.k. expert.",
		"Varje exemplar st. kostar fem kronor.",
		"Jag gillar djur, t.ex. hundar och katter.",
		"Vi tar upp övr. frågor på nästa möte.",
	)
}
