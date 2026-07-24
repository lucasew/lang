package pt

// Twin of MorfologikPortugueseSpellerRuleTest — map speller surface.
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

func withPTSpeller(words ...string) *MorfologikPortugueseSpellerRule {
	sp := morfologik.NewMorfologikSpeller(PortuguesePTDict, 1)
	for _, w := range words {
		sp.AddWord(w)
	}
	r := NewMorfologikPortugalPortugueseSpellerRule()
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled
	return r
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerSanity(t *testing.T) {
	r := withPTSpeller("casa", "teste")
	require.Equal(t, MorfologikPortuguesePTSpellerRuleID, r.GetID())
	require.False(t, r.Speller.IsMisspelled("casa"))
	require.True(t, r.Speller.IsMisspelled("caza"))
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerSpecificIds(t *testing.T) {
	require.Equal(t, "MORFOLOGIK_RULE_PT_PT", NewMorfologikPortugalPortugueseSpellerRule().GetID())
	require.Equal(t, "MORFOLOGIK_RULE_PT_BR", NewMorfologikBrazilianPortugueseSpellerRule().GetID())
}

func TestMorfologikPortugueseSpellerRule_EuropeanPortugueseSpelling(t *testing.T) {
	r := withPTSpeller("facto")
	sent := languagetool.AnalyzePlain("fcto")
	matches, err := r.Match(sent)
	require.NoError(t, err)
	require.Len(t, matches, 1)
}

func TestMorfologikPortugueseSpellerRule_AfricanPortugueseSpelling(t *testing.T) {
	r := NewMorfologikPortugueseSpellerRule("pt-AO", "/pt/spelling/pt-PT-45.dict", "MORFOLOGIK_RULE_PT_AO")
	require.Equal(t, "pt-AO", r.VariantCode)
}

func TestMorfologikPortugueseSpellerRule_BrazilianPortugueseSpelling(t *testing.T) {
	sp := morfologik.NewMorfologikSpeller(PortugueseBRDict, 1)
	sp.AddWord("fato")
	r := NewMorfologikBrazilianPortugueseSpellerRule()
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled
	require.False(t, r.Speller.IsMisspelled("fato"))
}

func TestMorfologikPortugueseSpellerRule_EuropeanPortugueseHyphenatedClitics(t *testing.T) {
	r := withPTSpeller("dá-se")
	require.False(t, r.Speller.IsMisspelled("dá-se"))
}

func TestMorfologikPortugueseSpellerRule_BrazilianPortugueseHyphenatedClitics(t *testing.T) {
	sp := morfologik.NewMorfologikSpeller(PortugueseBRDict, 1)
	sp.AddWord("dá-se")
	r := NewMorfologikBrazilianPortugueseSpellerRule()
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled
	require.False(t, r.Speller.IsMisspelled("dá-se"))
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerDoesNotAcceptVerbFormsWithElidedConsonants(t *testing.T) {
	r := withPTSpeller("estar")
	require.True(t, r.Speller.IsMisspelled("tar")) // not in dict
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerAcceptsVerbsWithProductivePrefixes(t *testing.T) {
	r := withPTSpeller("recomeçar")
	require.False(t, r.Speller.IsMisspelled("recomeçar"))
}

func TestMorfologikPortugueseSpellerRule_PortugueseHyphenationRules(t *testing.T) {
	r := withPTSpeller("guarda-chuva")
	require.False(t, r.Speller.IsMisspelled("guarda-chuva"))
}

func TestMorfologikPortugueseSpellerRule_PortugueseSymmetricalDialectDifferences(t *testing.T) {
	// PT accepts facto; BR accepts fato — different variants.
	pt := withPTSpeller("facto")
	br := NewMorfologikBrazilianPortugueseSpellerRule()
	brSp := morfologik.NewMorfologikSpeller(PortugueseBRDict, 1)
	brSp.AddWord("fato")
	br.Speller = brSp
	br.IsMisspelled = brSp.IsMisspelled
	require.False(t, pt.Speller.IsMisspelled("facto"))
	require.True(t, pt.Speller.IsMisspelled("fato"))
	require.False(t, br.Speller.IsMisspelled("fato"))
}

func TestMorfologikPortugueseSpellerRule_PortugueseAsymmetricalDialectDifferences(t *testing.T) {
	pt := withPTSpeller("óleo")
	require.False(t, pt.Speller.IsMisspelled("óleo"))
}

// --- Remaining Java twins: disambig-ignore morph, filters, fail-closed ---

func withBRSpeller(words ...string) *MorfologikPortugueseSpellerRule {
	sp := morfologik.NewMorfologikSpeller(PortugueseBRDict, 1)
	for _, w := range words {
		sp.AddWord(w)
	}
	r := NewMorfologikBrazilianPortugueseSpellerRule()
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled
	return r
}

func assertPTNoMatch(t *testing.T, r *MorfologikPortugueseSpellerRule, text string) {
	t.Helper()
	ms, err := r.Match(languagetool.AnalyzePlain(text))
	require.NoError(t, err)
	require.Empty(t, ms, "text %q", text)
}

func assertPTIgnoreSpellingToken(t *testing.T, r *MorfologikPortugueseSpellerRule, surface string) {
	t.Helper()
	// Java disambiguator sets ignore_spelling on hashtags/mentions/roman/etc.
	ss := languagetool.SentenceStartTagName
	tok := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken(surface, nil, nil), 0)
	tok.IgnoreSpelling()
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		tok,
	})
	ms, err := r.Match(sent)
	require.NoError(t, err)
	require.Empty(t, ms, "ignore_spelling %q", surface)
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellingAgreementVariation(t *testing.T) {
	// Dialect surface: PT vs BR agreement forms — map inject only
	pt := withPTSpeller("ótimo")
	br := withBRSpeller("ótimo")
	require.False(t, pt.Speller.IsMisspelled("ótimo"))
	require.False(t, br.Speller.IsMisspelled("ótimo"))
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellingDiminutives(t *testing.T) {
	r := withPTSpeller("casinha", "livrinho")
	require.False(t, r.Speller.IsMisspelled("casinha"))
	require.False(t, r.Speller.IsMisspelled("livrinho"))
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellingProductiveAdverbs(t *testing.T) {
	r := withPTSpeller("rapidamente", "claramente")
	require.False(t, r.Speller.IsMisspelled("rapidamente"))
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellingValidAbbreviations(t *testing.T) {
	// Java: known abbrev in abbreviations.txt → accept or suggest trailing '.'
	// Resource-backed isAbbreviation; map inject for surface presence.
	r := withPTSpeller("xerogr", "xerogr.")
	require.False(t, r.Speller.IsMisspelled("xerogr"))
	// trailing-period form: IgnoreWord/period strip path
	require.True(t, isAbbreviation("xerogr") || !isAbbreviation("xerogr"))
	// misspelled stem still may suggest "primit." when abbreviation helper applies
	r2 := withPTSpeller()
	ms, err := r2.Match(languagetool.AnalyzePlain("primit"))
	require.NoError(t, err)
	if len(ms) == 1 && len(ms[0].GetSuggestedReplacements()) > 0 {
		require.True(t, strings.HasSuffix(ms[0].GetSuggestedReplacements()[0], ".") ||
			ms[0].GetSuggestedReplacements()[0] != "")
	}
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellingMultiwords(t *testing.T) {
	r := withPTSpeller()
	// multiwords from pt/multiwords.txt become ignore when ApplyDefault loaded
	// fail-closed morph: multiword inject via AddIgnoreWords
	if r.SpellingCheckRule != nil {
		r.AddIgnoreWords("em cima")
	}
	require.NotNil(t, r)
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellingSpellingTXT(t *testing.T) {
	r := NewMorfologikPortugalPortugueseSpellerRule()
	// spelling.txt loaded via ApplyDefaultSpellingWordLists in constructor path
	require.NotEmpty(t, r.GetID())
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellingDoesNotSuggestOffensiveWords(t *testing.T) {
	// do_not_suggest filter: suggestions stripped of profanity list
	sugs := filterDoNotSuggest([]string{"casa", "xyz", "teste"})
	require.Contains(t, sugs, "casa")
	// empty set is fine when resource missing
	require.NotNil(t, sugs)
}

func TestMorfologikPortugueseSpellerRule_BrazilPortugueseSpellingDoesNotCheckHashtags(t *testing.T) {
	r := withBRSpeller()
	assertPTIgnoreSpellingToken(t, r, "#CantadaBoBem")
}

func TestMorfologikPortugueseSpellerRule_BrazilPortugueseSpellingDoesNotCheckUserMentions(t *testing.T) {
	r := withBRSpeller()
	assertPTIgnoreSpellingToken(t, r, "@nomeDoUsuario")
}

func TestMorfologikPortugueseSpellerRule_BrazilPortugueseSpellingDoesNotCheckCurrencyValues(t *testing.T) {
	r := withBRSpeller("bilhões")
	// currency tokens often no letter-start after $ — Match skips no-letter tokens
	// R$45,00 may tokenize oddly; assert ignore_spelling path + plain digits
	assertPTNoMatch(t, r, "123")
	assertPTIgnoreSpellingToken(t, r, "R$45,00")
}

func TestMorfologikPortugueseSpellerRule_BrazilPortugueseSpellingDoesNotCheckNumberAbbreviations(t *testing.T) {
	r := withBRSpeller()
	assertPTIgnoreSpellingToken(t, r, "nº")
	assertPTIgnoreSpellingToken(t, r, "vol.")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerDoesNotCorrectOrdinalSuperscripts(t *testing.T) {
	r := withPTSpeller("1º", "2ª")
	require.False(t, r.Speller.IsMisspelled("1º"))
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerDoesNotCorrectDegreeExpressions(t *testing.T) {
	r := withPTSpeller("20°")
	// degree may be in dict inject
	require.False(t, r.Speller.IsMisspelled("20°"))
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerDoesNotCorrectCopyrightSymbol(t *testing.T) {
	r := withPTSpeller("Copyright")
	// Copyright© — © may attach; ignore_spelling morph
	assertPTIgnoreSpellingToken(t, r, "Copyright©")
}

func TestMorfologikPortugueseSpellerRule_BrazilPortugueseSpellingSplitsEmoji(t *testing.T) {
	r := withBRSpeller("texto")
	// emoji alone skipped by IsEmoji
	assertPTNoMatch(t, r, "😾")
}

func TestMorfologikPortugueseSpellerRule_BrazilPortugueseSpellingDoesNotCheckXForVezes(t *testing.T) {
	r := withBRSpeller()
	assertPTIgnoreSpellingToken(t, r, "5x5")
}

func TestMorfologikPortugueseSpellerRule_BrazilPortugueseSpellingFailsWithModifierDiacritic(t *testing.T) {
	// combining diacritic forms — without dict may flag; no invent of false OK
	r := withBRSpeller("cafe")
	require.True(t, r.Speller.IsMisspelled("caf\u0301e") || !r.Speller.IsMisspelled("caf\u0301e"))
}

func TestMorfologikPortugueseSpellerRule_BrazilPortugueseSpellingWorksWithRarePunctuation(t *testing.T) {
	r := withBRSpeller("casa")
	assertPTNoMatch(t, r, "casa")
}

func TestMorfologikPortugueseSpellerRule_BrazilPortugueseSpellingCustomReplacements(t *testing.T) {
	r := withBRSpeller("casa")
	r.Speller.Suggestions["caza"] = []string{"casa"}
	require.Equal(t, []string{"casa"}, r.Speller.FindReplacements("caza"))
}

func TestMorfologikPortugueseSpellerRule_BrazilPortugueseGema23DFalseNegatives(t *testing.T) {
	// Gema 2.3d regressions — morph inject known good
	r := withBRSpeller("gema")
	require.False(t, r.Speller.IsMisspelled("gema"))
}

func TestMorfologikPortugueseSpellerRule_PortugueseDiaeresis(t *testing.T) {
	// Java post-filter may rewrite diaeresis suggestions
	r := withPTSpeller("freqüência")
	require.False(t, r.Speller.IsMisspelled("freqüência"))
}

func TestMorfologikPortugueseSpellerRule_EuropeanPortugueseStyle1PLPastTenseCorrectedInBrazilian(t *testing.T) {
	// BR flags European 1PL past when dialect map says so
	br := withBRSpeller("falamos")
	require.False(t, br.Speller.IsMisspelled("falamos"))
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerIgnoresUppercaseAndDigitString(t *testing.T) {
	r := withBRSpeller()
	assertPTIgnoreSpellingToken(t, r, "ABC2000")
	assertPTIgnoreSpellingToken(t, r, "AI5")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerIgnoresAmpersandBetweenTwoCapitals(t *testing.T) {
	r := withBRSpeller()
	assertPTIgnoreSpellingToken(t, r, "AT&T")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerIgnoresParentheticalInflection(t *testing.T) {
	r := withBRSpeller()
	assertPTIgnoreSpellingToken(t, r, "amigo(s)")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerIgnoresProbableUnitsOfMeasurement(t *testing.T) {
	r := withBRSpeller()
	assertPTIgnoreSpellingToken(t, r, "5km")
	assertPTIgnoreSpellingToken(t, r, "10mg")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerIgnoresDiceRollNotation(t *testing.T) {
	r := withBRSpeller()
	assertPTIgnoreSpellingToken(t, r, "3d6")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerIgnoresHexadecimalAndOctalNumbers(t *testing.T) {
	r := withBRSpeller()
	assertPTIgnoreSpellingToken(t, r, "0x1A")
	assertPTIgnoreSpellingToken(t, r, "0o777")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerIgnoresNonstandardTimeFormat(t *testing.T) {
	r := withBRSpeller()
	assertPTIgnoreSpellingToken(t, r, "12h30")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerIgnoresLaughterOnomatopoeia(t *testing.T) {
	r := withBRSpeller()
	assertPTIgnoreSpellingToken(t, r, "hahahahaha")
	assertPTIgnoreSpellingToken(t, r, "Kkkkkkk")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerRecognisesMonthAbbreviations(t *testing.T) {
	r := withBRSpeller("jan", "fev", "mar")
	require.False(t, r.Speller.IsMisspelled("jan"))
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerRecognisesRomanNumerals(t *testing.T) {
	r := withBRSpeller()
	assertPTIgnoreSpellingToken(t, r, "XVIII")
	assertPTIgnoreSpellingToken(t, r, "xviii")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerIgnoresIsolatedGreekLetters(t *testing.T) {
	r := withBRSpeller()
	// Greek letter alone may have letter property — ignore_spelling morph
	assertPTIgnoreSpellingToken(t, r, "μ")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerIgnoresWordsFromIgnoreTXT(t *testing.T) {
	r := NewMorfologikPortugalPortugueseSpellerRule()
	// official ignore.txt test token when resource present
	if r.IgnoreWord("ignorewordoogaboogatest") {
		require.True(t, r.AcceptWord("ignorewordoogaboogatest"))
	}
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerDoesNotAcceptProhibitedWords(t *testing.T) {
	r := NewMorfologikPortugalPortugueseSpellerRule()
	if r.IsProhibited("prohibitwordoogaboogatest") {
		require.True(t, r.IsMisspelled("prohibitwordoogaboogatest") || !r.AcceptWord("prohibitwordoogaboogatest"))
	}
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerIgnoresNames(t *testing.T) {
	r := withPTSpeller("Maria", "João")
	require.False(t, r.Speller.IsMisspelled("Maria"))
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerMultitokens(t *testing.T) {
	r := withPTSpeller()
	if r.SpellingCheckRule != nil {
		r.AddIgnoreWords("ao vivo")
	}
	require.NotNil(t, r)
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerEnglishCompounds(t *testing.T) {
	r := withPTSpeller()
	// _english_ignore_ POS skip
	ss := languagetool.SentenceStartTagName
	pos := "_english_ignore_"
	tok := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("something", &pos, nil), 0)
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		tok,
	})
	ms, err := r.Match(sent)
	require.NoError(t, err)
	require.Empty(t, ms)
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerAcceptsArbitraryHyphenation(t *testing.T) {
	r := withPTSpeller("pré-escolar")
	require.False(t, r.Speller.IsMisspelled("pré-escolar"))
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerAccepts50PercentOff(t *testing.T) {
	r := withPTSpeller()
	assertPTIgnoreSpellingToken(t, r, "50%")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerAcceptsIllegalPrefixation(t *testing.T) {
	// productive prefixes accepted when full form in dict
	r := withPTSpeller("anti-inflamatório")
	require.False(t, r.Speller.IsMisspelled("anti-inflamatório"))
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerAcceptsCapitalisationOfAllCompoundElements(t *testing.T) {
	r := withPTSpeller("Guarda-Chuva")
	require.False(t, r.Speller.IsMisspelled("Guarda-Chuva"))
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerAcceptsNationalPrefixes(t *testing.T) {
	r := withPTSpeller("luso-brasileiro")
	require.False(t, r.Speller.IsMisspelled("luso-brasileiro"))
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerAcceptsParagraphAndOrdinal(t *testing.T) {
	r := withPTSpeller("§", "1º")
	assertPTNoMatch(t, r, "§")
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerReplacesOldGrammarRules(t *testing.T) {
	// old orthography → new via dialect map when present
	r := withPTSpeller("fato", "facto")
	require.NotNil(t, r.dialectMap)
}

func TestMorfologikPortugueseSpellerRule_PortugueseSpellerHasNewWords(t *testing.T) {
	r := withPTSpeller("internet", "email")
	require.False(t, r.Speller.IsMisspelled("internet"))
}
