package pt

// Twin of languagetool-language-modules/pt/src/test/java/org/languagetool/rules/pt/PortugueseWordRepeatRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// ignoreAt ports PortugueseWordRepeatRuleTest.ignore:
//   AnalyzedSentence sentence = lt.getAnalyzedSentence(input);
//   AnalyzedTokenReadings[] tokens = sentence.getTokensWithoutWhitespace();
//   return rule.ignore(tokens, position);
//
// Go uses AnalyzePlain (WordTokenizer). Like Java PT v0.13, "coloquem-na" stays one
// token (core WordTokenizer does not split ASCII hyphen-minus).
func ignoreAt(t *testing.T, rule *PortugueseWordRepeatRule, input string, position int) bool {
	t.Helper()
	sent := languagetool.AnalyzePlain(input)
	tokens := sent.GetTokensWithoutWhitespace()
	return rule.Ignore(tokens, position)
}

// Java PortugueseWordRepeatRuleTest.testIgnore matrix (exact outcomes).
func TestPortugueseWordRepeatRule_Ignore(t *testing.T) {
	rule := NewPortugueseWordRepeatRule(map[string]string{"repetition": "Repetição"})

	require.False(t, ignoreAt(t, rule, "no repetition", 2))
	require.True(t, ignoreAt(t, rule, "blá blá", 2))
	require.True(t, ignoreAt(t, rule, "Aaptos aaptos", 2))
	require.True(t, ignoreAt(t, rule, "Logo logo vamos ao mercado", 2))
	// Java: as of the v0.13 dict, 'coloquem-na' is a single token so position 2 is
	// the first standalone "na" — not a hyphen-path pronoun ignore → false.
	// Go WordTokenizer keeps "Coloquem-na" whole → same surfaces at position 2.
	require.False(t, ignoreAt(t, rule, "Coloquem-na na sala.", 2))
}

// Case-sensitive wordRepetitionOf (Java Token.equals) for blá/sapiens/tuk.
// Note: "Se se" is also true via TAUTONYMS genus "Se" + species "se" (not only wordRep).
func TestPortugueseWordRepeatRule_WordRepetitionCaseSensitive(t *testing.T) {
	rule := NewPortugueseWordRepeatRule(map[string]string{"repetition": "Repetição"})
	// "Blá blá": not wordRepetitionOf("blá") (case), not redup adverb, not tautonym → false.
	require.False(t, ignoreAt(t, rule, "Blá blá", 2))
	require.True(t, ignoreAt(t, rule, "blá blá", 2))
	// "se se" via wordRepetitionOf("se"); "Se se" via genus/species Se|se.
	require.True(t, ignoreAt(t, rule, "se se", 2))
	require.True(t, ignoreAt(t, rule, "Se se", 2), "Java tautonym Se + se")
	require.True(t, ignoreAt(t, rule, "sapiens sapiens", 2))
	require.False(t, ignoreAt(t, rule, "Sapiens sapiens", 2))
	require.True(t, ignoreAt(t, rule, "tuk tuk", 2))
	require.False(t, ignoreAt(t, rule, "Tuk tuk", 2))
}

// REDUPLICATED_ADVERBS is CASE_INSENSITIVE; Java only checks tokens[position].
func TestPortugueseWordRepeatRule_ReduplicatedAdverbsCase(t *testing.T) {
	rule := NewPortugueseWordRepeatRule(map[string]string{"repetition": "Repetição"})
	require.True(t, ignoreAt(t, rule, "Logo logo vamos", 2))
	require.True(t, ignoreAt(t, rule, "logo Logo vamos", 2))
	require.True(t, ignoreAt(t, rule, "já já", 2))
	require.True(t, ignoreAt(t, rule, "Já Já", 2))
	require.True(t, ignoreAt(t, rule, "fácil fácil", 2))
	require.True(t, ignoreAt(t, rule, "Fácil Fácil", 2))
}

// atr builds a surface-only reading (token string, whitespaceBefore).
func atr(surface string, wsBefore bool) *languagetool.AnalyzedTokenReadings {
	tok := languagetool.NewAnalyzedToken(surface, nil, nil)
	r := languagetool.NewAnalyzedTokenReadings(tok)
	r.SetWhitespaceBefore(wsBefore)
	return r
}

// Hyphen + pronoun path (Java isHyphenated at the *second* word of the repeat):
//   tokens[position-2] == "-" && !tokens[position-1].isWhitespaceBefore()
// e.g. [SENT_START, Coloquem, -, na, na] at position 4.
// PRONOUNS is case-sensitive (no CASE_INSENSITIVE in Java).
func TestPortugueseWordRepeatRule_HyphenPronounIgnore(t *testing.T) {
	rule := NewPortugueseWordRepeatRule(nil)
	// Split form of "Coloquem-na na …" when hyphen is its own token.
	tokens := []*languagetool.AnalyzedTokenReadings{
		atr("", false),       // 0 SENT_START
		atr("Coloquem", false), // 1
		atr("-", false),      // 2
		atr("na", false),     // 3 clitic, no whitespace before
		atr("na", true),      // 4 following "na" with whitespace (the repeat)
	}
	require.True(t, rule.Ignore(tokens, 4))

	// Uppercase second "Na" must not match case-sensitive PRONOUNS.
	tokensNa := []*languagetool.AnalyzedTokenReadings{
		atr("", false),
		atr("Coloquem", false),
		atr("-", false),
		atr("na", false),
		atr("Na", true),
	}
	require.False(t, rule.Ignore(tokensNa, 4))

	// Whitespace before the clitic (position-1) → isHyphenated false.
	tokensWS := []*languagetool.AnalyzedTokenReadings{
		atr("", false),
		atr("Coloquem", false),
		atr("-", false),
		atr("na", true), // whitespace before clitic breaks isHyphenated
		atr("na", true),
	}
	require.False(t, rule.Ignore(tokensWS, 4))

	// Current Java/Go tokenizer: "Coloquem-na" one token → ignore false at pos 2.
	require.False(t, ignoreAt(t, rule, "Coloquem-na na sala.", 2))
}

// Match-level outcomes: ignored repeats produce no RuleMatch; real repeats do.
func TestPortugueseWordRepeatRule_Match(t *testing.T) {
	rule := NewPortugueseWordRepeatRule(map[string]string{
		"repetition":            "Repetição",
		"desc_repetition_short": "Repetição",
	})
	// Ignored
	require.Empty(t, rule.Match(languagetool.AnalyzePlain("blá blá")))
	require.Empty(t, rule.Match(languagetool.AnalyzePlain("Aaptos aaptos")))
	require.Empty(t, rule.Match(languagetool.AnalyzePlain("Logo logo vamos ao mercado")))
	require.Empty(t, rule.Match(languagetool.AnalyzePlain("Homo sapiens sapiens")))
	require.Empty(t, rule.Match(languagetool.AnalyzePlain("tuk tuk")))
	// Real repetition (Java example style)
	ms := rule.Match(languagetool.AnalyzePlain("Este é é apenas uma frase de exemplo."))
	require.Len(t, ms, 1)
	require.Equal(t, []string{"é"}, ms[0].GetSuggestedReplacements())
	// Bare "na na" without hyphen path is a real repetition (not ignored).
	ms2 := rule.Match(languagetool.AnalyzePlain("na na sala"))
	require.Len(t, ms2, 1)
}

// Java ctor metadata: id, REPETITIONS category, Duplication, example pair.
func TestPortugueseWordRepeatRule_Metadata(t *testing.T) {
	rule := NewPortugueseWordRepeatRule(nil)
	require.Equal(t, "PORTUGUESE_WORD_REPEAT_RULE", rule.GetID())
	require.Equal(t, 1, rule.EstimateContextForSureMatch())
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "REPETITIONS", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSDuplication, rule.GetLocQualityIssueType())
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "Este <marker>é é</marker> apenas uma frase de exemplo.", inc[0].GetExample())
	require.Equal(t, []string{"é"}, inc[0].GetCorrections())
	require.Equal(t, "Este <marker>é</marker> apenas uma frase de exemplo.", rule.GetCorrectExamples()[0].GetExample())
}
