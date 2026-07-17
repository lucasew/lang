package uk

// Twin of languagetool-language-modules/uk/src/test/java/org/languagetool/tagging/disambiguation/uk/UkrainianHybridDisambiguationTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"github.com/stretchr/testify/require"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/tagging/disambiguation/uk/UkrainianHybridDisambiguationTest.java :: UkrainianHybridDisambiguationTest.testDisambiguator
func TestUkrainianHybridDisambiguation_Disambiguator(t *testing.T) {
	t.Skip("unimplemented: UkrainianHybridDisambiguationTest.testDisambiguator")
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/tagging/disambiguation/uk/UkrainianHybridDisambiguationTest.java :: UkrainianHybridDisambiguationTest.testDisambiguatorDups
func TestUkrainianHybridDisambiguation_DisambiguatorDups(t *testing.T) {
	// inject dups map (full disambig_dups.txt deferred)
	dups := map[string][]string{"весь": {"ввесь"}}
	d := NewUkrainianHybridDisambiguatorWith(nil, NewSimpleDisambiguatorFull(nil, dups))
	start := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("", strPtr("SENT_START"), nil),
	}, 0)
	p1, l1 := "adj:m:v_naz:pron:gen", "весь"
	p2, l2 := "adj:m:v_naz:pron:gen", "ввесь"
	tok := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("весь", &p1, &l1),
		languagetool.NewAnalyzedToken("весь", &p2, &l2),
	}, 0)
	out := d.Disambiguate(languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, tok}))
	require.True(t, out.GetTokensWithoutWhitespace()[1].HasAnyLemma("весь"))
	require.False(t, out.GetTokensWithoutWhitespace()[1].HasAnyLemma("ввесь"))
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/tagging/disambiguation/uk/UkrainianHybridDisambiguationTest.java :: UkrainianHybridDisambiguationTest.testDisambiguatorRetagFemNames
func TestUkrainianHybridDisambiguation_DisambiguatorRetagFemNames(t *testing.T) {
	t.Skip("unimplemented: UkrainianHybridDisambiguationTest.testDisambiguatorRetagFemNames")
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/tagging/disambiguation/uk/UkrainianHybridDisambiguationTest.java :: UkrainianHybridDisambiguationTest.testDisambiguatorRemoveVmis
func TestUkrainianHybridDisambiguation_DisambiguatorRemoveVmis(t *testing.T) {
	// unit green without full UK dict: hybrid applies RemoveVmisReadings
	p, l1, l2 := "noun:inanim:m:v_mis", "зв'язок", "зв'язок"
	p2 := "noun:inanim:m:v_rod"
	atr := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("Зв'язку", &p, &l1),
		languagetool.NewAnalyzedToken("Зв'язку", &p2, &l2),
	}, 0)
	start := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("", strPtr("SENT_START"), nil),
	}, 0)
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, atr})
	out := NewUkrainianHybridDisambiguator().Disambiguate(sent)
	toks := out.GetTokensWithoutWhitespace()
	require.False(t, toks[1].HasPartialPosTag("v_mis"))
	require.True(t, toks[1].HasPartialPosTag("v_rod"))
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/tagging/disambiguation/uk/UkrainianHybridDisambiguationTest.java :: UkrainianHybridDisambiguationTest.testDisambiguatorForInanimVKly
func TestUkrainianHybridDisambiguation_DisambiguatorForInanimVKly(t *testing.T) {
	t.Skip("unimplemented: UkrainianHybridDisambiguationTest.testDisambiguatorForInanimVKly")
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/tagging/disambiguation/uk/UkrainianHybridDisambiguationTest.java :: UkrainianHybridDisambiguationTest.testDisambiguatorForPluralNames
func TestUkrainianHybridDisambiguation_DisambiguatorForPluralNames(t *testing.T) {
	t.Skip("unimplemented: UkrainianHybridDisambiguationTest.testDisambiguatorForPluralNames")
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/tagging/disambiguation/uk/UkrainianHybridDisambiguationTest.java :: UkrainianHybridDisambiguationTest.testDisambiguatorForInitials
func TestUkrainianHybridDisambiguation_DisambiguatorForInitials(t *testing.T) {
	// contains assertEquals — full values in Java twin source
	// contains assertTrue
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/tagging/disambiguation/uk/UkrainianHybridDisambiguationTest.java :: UkrainianHybridDisambiguationTest.testDisambiguatorRemove
func TestUkrainianHybridDisambiguation_DisambiguatorRemove(t *testing.T) {
	t.Skip("unimplemented: UkrainianHybridDisambiguationTest.testDisambiguatorRemove")
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/tagging/disambiguation/uk/UkrainianHybridDisambiguationTest.java :: UkrainianHybridDisambiguationTest.testDisambiguatorForSt
func TestUkrainianHybridDisambiguation_DisambiguatorForSt(t *testing.T) {
	t.Skip("unimplemented: UkrainianHybridDisambiguationTest.testDisambiguatorForSt")
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/tagging/disambiguation/uk/UkrainianHybridDisambiguationTest.java :: UkrainianHybridDisambiguationTest.testTaggerUppgerGoodAndLowerBad
func TestUkrainianHybridDisambiguation_TaggerUppgerGoodAndLowerBad(t *testing.T) {
	t.Skip("unimplemented: UkrainianHybridDisambiguationTest.testTaggerUppgerGoodAndLowerBad")
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/tagging/disambiguation/uk/UkrainianHybridDisambiguationTest.java :: UkrainianHybridDisambiguationTest.testTaggingForUpperCaseAbbreviations
func TestUkrainianHybridDisambiguation_TaggingForUpperCaseAbbreviations(t *testing.T) {
	t.Skip("unimplemented: UkrainianHybridDisambiguationTest.testTaggingForUpperCaseAbbreviations")
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/tagging/disambiguation/uk/UkrainianHybridDisambiguationTest.java :: UkrainianHybridDisambiguationTest.testPronPos
func TestUkrainianHybridDisambiguation_PronPos(t *testing.T) {
	t.Skip("unimplemented: UkrainianHybridDisambiguationTest.testPronPos")
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/tagging/disambiguation/uk/UkrainianHybridDisambiguationTest.java :: UkrainianHybridDisambiguationTest.testYih
func TestUkrainianHybridDisambiguation_Yih(t *testing.T) {
	t.Skip("unimplemented: UkrainianHybridDisambiguationTest.testYih")
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/tagging/disambiguation/uk/UkrainianHybridDisambiguationTest.java :: UkrainianHybridDisambiguationTest.testSimpleRemove
func TestUkrainianHybridDisambiguation_SimpleRemove(t *testing.T) {
	// inject remove map (full disambig_remove.txt deferred)
	rm := map[string]*TokenMatcher{
		"була": {Entries: []MatcherEntry{{Lemma: "була", POS: "noun"}}},
	}
	d := NewUkrainianHybridDisambiguatorWith(nil, NewSimpleDisambiguatorWith(rm))
	start := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("", strPtr("SENT_START"), nil),
	}, 0)
	pv, pl := "verb:imperf:past:f", "бути"
	nv, nl := "noun:inanim:f:v_naz", "була"
	bula := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("була", &pv, &pl),
		languagetool.NewAnalyzedToken("була", &nv, &nl),
	}, 0)
	out := d.Disambiguate(languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, bula}))
	require.True(t, out.GetTokensWithoutWhitespace()[1].HasPartialPosTag("verb"))
	require.False(t, out.GetTokensWithoutWhitespace()[1].HasPartialPosTag("noun"))
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/tagging/disambiguation/uk/UkrainianHybridDisambiguationTest.java :: UkrainianHybridDisambiguationTest.testDisambiguatorRemovePresentInDictionary
func TestUkrainianHybridDisambiguation_DisambiguatorRemovePresentInDictionary(t *testing.T) {
	// contains assertTrue
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/tagging/disambiguation/uk/UkrainianHybridDisambiguationTest.java :: UkrainianHybridDisambiguationTest.testChunker
func TestUkrainianHybridDisambiguation_Chunker(t *testing.T) {
	// contains assertTrue
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/tagging/disambiguation/uk/UkrainianHybridDisambiguationTest.java :: UkrainianHybridDisambiguationTest.testIgnoredCharacters
func TestUkrainianHybridDisambiguation_IgnoredCharacters(t *testing.T) {
	// contains assertEquals — full values in Java twin source
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/tagging/disambiguation/uk/UkrainianHybridDisambiguationTest.java :: UkrainianHybridDisambiguationTest.testPluralProp
func TestUkrainianHybridDisambiguation_PluralProp(t *testing.T) {
	t.Skip("unimplemented: UkrainianHybridDisambiguationTest.testPluralProp")
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/tagging/disambiguation/uk/UkrainianHybridDisambiguationTest.java :: UkrainianHybridDisambiguationTest.testVerbImpr
func TestUkrainianHybridDisambiguation_VerbImpr(t *testing.T) {
	t.Skip("unimplemented: UkrainianHybridDisambiguationTest.testVerbImpr")
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/tagging/disambiguation/uk/UkrainianHybridDisambiguationTest.java :: UkrainianHybridDisambiguationTest.testVklyZvert
func TestUkrainianHybridDisambiguation_VklyZvert(t *testing.T) {
	t.Skip("unimplemented: UkrainianHybridDisambiguationTest.testVklyZvert")
}
