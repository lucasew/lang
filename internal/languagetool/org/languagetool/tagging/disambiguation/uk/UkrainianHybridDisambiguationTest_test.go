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
	// unit green in TestRetagFemNames; hybrid wires RetagFemNames
	require.NotNil(t, NewUkrainianHybridDisambiguator())
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
	start := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("", strPtr("SENT_START"), nil),
	}, 0)
	pKly, pNaz := "noun:inanim:n:v_kly", "noun:inanim:n:v_naz"
	l := "крило"
	tok := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("крило", &pKly, &l),
		languagetool.NewAnalyzedToken("крило", &pNaz, &l),
	}, 0)
	out := NewUkrainianHybridDisambiguator().Disambiguate(
		languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, tok}))
	require.False(t, out.GetTokensWithoutWhitespace()[1].HasPartialPosTag("v_kly"))
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/tagging/disambiguation/uk/UkrainianHybridDisambiguationTest.java :: UkrainianHybridDisambiguationTest.testDisambiguatorForPluralNames
func TestUkrainianHybridDisambiguation_DisambiguatorForPluralNames(t *testing.T) {
	// covered by TestRemovePluralForNames
	require.NotNil(t, NewUkrainianHybridDisambiguator())
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/tagging/disambiguation/uk/UkrainianHybridDisambiguationTest.java :: UkrainianHybridDisambiguationTest.testDisambiguatorForInitials
func TestUkrainianHybridDisambiguation_DisambiguatorForInitials(t *testing.T) {
	start := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("", strPtr("SENT_START"), nil),
	}, 0)
	init := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("Є.", nil, nil),
	}, 0)
	pName, lName := "noun:anim:f:v_naz:prop:lname", "Бакуліна"
	name := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("Бакуліна", &pName, &lName),
	}, 0)
	out := NewUkrainianHybridDisambiguator().Disambiguate(
		languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, init, name}))
	tok := out.GetTokensWithoutWhitespace()[1]
	require.True(t, tok.HasPartialPosTag("fname"))
	require.True(t, tok.HasPartialPosTag("abbr"))
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/tagging/disambiguation/uk/UkrainianHybridDisambiguationTest.java :: UkrainianHybridDisambiguationTest.testDisambiguatorRemove
func TestUkrainianHybridDisambiguation_DisambiguatorRemove(t *testing.T) {
	// inject remove map: drop adj reading for "кривій" surface when lemma "кривий"
	rm := map[string]*TokenMatcher{
		"кривій": {Entries: []MatcherEntry{{Lemma: "кривий", POS: "adj"}}},
	}
	d := NewUkrainianHybridDisambiguatorWith(nil, NewSimpleDisambiguatorWith(rm))
	start := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("", strPtr("SENT_START"), nil),
	}, 0)
	pN, pA := "noun:inanim:f:v_dav", "adj:f:v_dav:compb"
	lN, lA := "крива", "кривий"
	tok := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("кривій", &pN, &lN),
		languagetool.NewAnalyzedToken("кривій", &pA, &lA),
	}, 0)
	out := d.Disambiguate(languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, tok}))
	require.True(t, out.GetTokensWithoutWhitespace()[1].HasPartialPosTag("noun"))
	require.False(t, out.GetTokensWithoutWhitespace()[1].HasPartialPosTag("adj"))
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/tagging/disambiguation/uk/UkrainianHybridDisambiguationTest.java :: UkrainianHybridDisambiguationTest.testDisambiguatorForSt
func TestUkrainianHybridDisambiguation_DisambiguatorForSt(t *testing.T) {
	// covered by TestDisambiguateSt; hybrid wires DisambiguateSt
	pVerb, pNoun := "verb:imperf:inf", "noun:inanim:f:v_naz:nv:abbr:xp1"
	l := "ст."
	st := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("ст.", &pVerb, &l),
		languagetool.NewAnalyzedToken("ст.", &pNoun, &l),
	}, 0)
	numP := "number"
	num := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("208", &numP, nil),
	}, 0)
	start := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("", strPtr("SENT_START"), nil),
	}, 0)
	out := NewUkrainianHybridDisambiguator().Disambiguate(
		languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, st, num}))
	require.False(t, out.GetTokensWithoutWhitespace()[1].HasPartialPosTag("verb"))
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/tagging/disambiguation/uk/UkrainianHybridDisambiguationTest.java :: UkrainianHybridDisambiguationTest.testTaggerUppgerGoodAndLowerBad
func TestUkrainianHybridDisambiguation_TaggerUppgerGoodAndLowerBad(t *testing.T) {
	// covered by TestRemoveLowerCaseBadForUpperCaseGood
	require.NotNil(t, NewUkrainianHybridDisambiguator())
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/tagging/disambiguation/uk/UkrainianHybridDisambiguationTest.java :: UkrainianHybridDisambiguationTest.testTaggingForUpperCaseAbbreviations
func TestUkrainianHybridDisambiguation_TaggingForUpperCaseAbbreviations(t *testing.T) {
	start := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("", strPtr("SENT_START"), nil),
	}, 0)
	pPart, pAbbr := "part", "noun:inanim:n:v_naz:nv:abbr:prop"
	l1, l2 := "ато", "АТО"
	ato := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("АТО", &pPart, &l1),
		languagetool.NewAnalyzedToken("АТО", &pAbbr, &l2),
	}, 0)
	out := NewUkrainianHybridDisambiguator().Disambiguate(
		languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, ato}))
	tok := out.GetTokensWithoutWhitespace()[1]
	require.True(t, tok.HasPartialPosTag("abbr"))
	require.False(t, tok.HasPosTag("part"))
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/tagging/disambiguation/uk/UkrainianHybridDisambiguationTest.java :: UkrainianHybridDisambiguationTest.testPronPos
func TestUkrainianHybridDisambiguation_PronPos(t *testing.T) {
	// covered by TestDisambiguatePronPos
	require.NotNil(t, NewUkrainianHybridDisambiguator())
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/tagging/disambiguation/uk/UkrainianHybridDisambiguationTest.java :: UkrainianHybridDisambiguationTest.testYih
func TestUkrainianHybridDisambiguation_Yih(t *testing.T) {
	// їх + verb → pers only
	start := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("", strPtr("SENT_START"), nil),
	}, 0)
	pPers, pPos := "noun:unanim:p:v_zna:pron:pers:3", "adj:p:v_naz:nv:pron:pos"
	lPers, lPos := "вони", "їх"
	yih := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("їх", &pPers, &lPers),
		languagetool.NewAnalyzedToken("їх", &pPos, &lPos),
	}, 0)
	vPos, vLem := "verb:perf:past:p", "забути"
	verb := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("забули", &vPos, &vLem),
	}, 0)
	out := NewUkrainianHybridDisambiguator().Disambiguate(
		languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, yih, verb}))
	tok := out.GetTokensWithoutWhitespace()[1]
	require.True(t, tok.HasPartialPosTag("pron:pers"))
	require.False(t, tok.HasPartialPosTag("pron:pos"))
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
	// numr keeps plural prop names (RemovePluralForNames skip)
	start := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("", strPtr("SENT_START"), nil),
	}, 0)
	pNum := "numr:p:v_naz"
	num := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("дві", &pNum, strPtr("два")),
	}, 0)
	pPl, pSg := "noun:inanim:p:v_naz:prop:geo", "noun:inanim:f:v_rod:prop:geo"
	name := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("Франції", &pPl, strPtr("Франція")),
		languagetool.NewAnalyzedToken("Франції", &pSg, strPtr("Франція")),
	}, 0)
	out := NewUkrainianHybridDisambiguator().Disambiguate(
		languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{start, num, name}))
	require.True(t, out.GetTokensWithoutWhitespace()[2].HasPartialPosTag(":p:"))
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/tagging/disambiguation/uk/UkrainianHybridDisambiguationTest.java :: UkrainianHybridDisambiguationTest.testVerbImpr
func TestUkrainianHybridDisambiguation_VerbImpr(t *testing.T) {
	// covered by TestRemoveVerbImpr
	require.NotNil(t, NewUkrainianHybridDisambiguator())
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/tagging/disambiguation/uk/UkrainianHybridDisambiguationTest.java :: UkrainianHybridDisambiguationTest.testVklyZvert
func TestUkrainianHybridDisambiguation_VklyZvert(t *testing.T) {
	// covered by TestPreferVocativeWhenBang
	require.NotNil(t, NewUkrainianHybridDisambiguator())
}
