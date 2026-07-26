package pt

// Twin of languagetool-language-modules/pt/src/test/java/org/languagetool/rules/pt/AccentuationCheckRuleTest._java
//
// Upstream ._java is a Catalan leftover (JLanguageTool(Catalan.getInstance()), Catalan
// sentences, testPositions constructs non-PT AccentuationCheckRule). This twin ports
// the same assertion shape (correct / incorrect / multi-match / positions) against the
// real PortugueseAccentuationCheckRule + official /pt/verbos_sem_acento_*.txt maps.
//
// Analysis: hand-built AnalyzedTokenReadings with FreeLing-style PT tags (same tag
// alphabet the rule's POS regexes expect). Surface-only AnalyzePlain is not enough —
// Match inspects POS via matchPostagRegexp.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestPortugueseAccentuationCheckRule_Metadata(t *testing.T) {
	r := NewPortugueseAccentuationCheckRule(nil)
	require.Equal(t, "ACCENTUATION_CHECK_PT", r.GetID())
	require.Equal(t, "Confusão com acentos gráficos", r.GetDescription())
	require.True(t, r.IsDefaultOff())
	require.True(t, r.DefaultOff)
	require.Equal(t, rules.ITSMisspelling, r.GetLocQualityIssueType())
	require.NotNil(t, r.GetCategory())
	require.Equal(t, "CONFUSED_WORDS", r.GetCategory().GetID().String())
	// Official maps (Java static PortugueseAccentuationDataLoader paths)
	require.Contains(t, r.relevantWords, "copia")
	require.Equal(t, "cópia", r.relevantWords["copia"].GetToken())
	require.True(t, r.relevantWords["copia"].MatchesPosTagRegex(`NC[FC][SN].*`))
	require.Contains(t, r.relevantWords2, "obvia")
	require.Equal(t, "óbvia", r.relevantWords2["obvia"].GetToken())
	require.Contains(t, r.relevantWords, "renuncia")
	require.Equal(t, "renúncia", r.relevantWords["renuncia"].GetToken())
	require.Contains(t, r.relevantWords2, "magnifica")
	require.Equal(t, "magnífica", r.relevantWords2["magnifica"].GetToken())
	require.Contains(t, r.relevantWords, "circunstancia")
	require.Equal(t, "circunstância", r.relevantWords["circunstancia"].GetToken())
}

// Correct sentences: map miss, already-noun GN tag, reflexive pronoun before verb.
func TestPortugueseAccentuationCheckRule_Correct(t *testing.T) {
	rule := NewPortugueseAccentuationCheckRule(nil)

	assertCorrect(t, rule, sent(
		tok("", languagetool.SentenceStartTagName, 0),
		tok("A", "DA0FS0", 0),
		tok("casa", "NCFS000", 2),
	))
	// accented surface is not a map key
	assertCorrect(t, rule, sent(
		tok("", languagetool.SentenceStartTagName, 0),
		tok("a", "DA0FS0", 0),
		tok("cópia", "NCFS000", 2),
	))
	// GN on current token → noun/adj branches skipped
	assertCorrect(t, rule, sent(
		tok("", languagetool.SentenceStartTagName, 0),
		tok("a", "DA0FS0", 0),
		tok("copia", "NCFS000", 2),
	))
	// reflexive pronoun before (PRONOME_PESSOAL + not hyphen-prefixed)
	assertCorrect(t, rule, sent(
		tok("", languagetool.SentenceStartTagName, 0),
		tok("se", "P0MS000", 0),
		tok("copia", "VMIP3S0", 3),
	))
}

// Determiner + matching-gender noun reading (Java "aquestes renuncies" / "a copia").
func TestPortugueseAccentuationCheckRule_DetNoun(t *testing.T) {
	rule := NewPortugueseAccentuationCheckRule(nil)
	s := sent(
		tok("", languagetool.SentenceStartTagName, 0),
		tok("a", "DA0FS0", 0),
		tok("copia", "VMIP3S0", 2),
	)
	matches := rule.Match(s)
	require.Len(t, matches, 1)
	require.Equal(t, []string{"cópia"}, matches[0].GetSuggestedReplacements())
	require.Equal(t, portugueseAccentuationMessage, matches[0].GetMessage())
	require.Equal(t, portugueseAccentuationShort, matches[0].GetShortMessage())
	require.Equal(t, 2, matches[0].GetFromPos())
	require.Equal(t, 7, matches[0].GetToPos())
}

// Preposition SPS00 before (Java "amb renuncies").
func TestPortugueseAccentuationCheckRule_PrepNoun(t *testing.T) {
	rule := NewPortugueseAccentuationCheckRule(nil)
	assertIncorrect(t, rule, sent(
		tok("", languagetool.SentenceStartTagName, 0),
		tok("com", "SPS00", 0),
		tok("renuncia", "VMIP3S0", 4),
	), "renúncia")
}

// Article surface + preposition de (Java "la renuncia del president").
func TestPortugueseAccentuationCheckRule_ArticleDe(t *testing.T) {
	rule := NewPortugueseAccentuationCheckRule(nil)
	// "A renuncia de presidente" — det branch also matches DA0FS0+NCFS; same replacement.
	assertIncorrect(t, rule, sent(
		tok("", languagetool.SentenceStartTagName, 0),
		tok("A", "DA0FS0", 0),
		tok("renuncia", "VMIP3S0", 2),
		tok("de", "SPS00", 11),
		tok("presidente", "NCMS000", 14),
	), "renúncia")
}

// Noun reading + following adjective (Java "circunstancias extraordinàries").
func TestPortugueseAccentuationCheckRule_NounThenAdj(t *testing.T) {
	rule := NewPortugueseAccentuationCheckRule(nil)
	// singular FS form in official map: circunstancia
	assertIncorrect(t, rule, sent(
		tok("", languagetool.SentenceStartTagName, 0),
		tok("Uma", "DI0FS0", 0),
		tok("circunstancia", "VMIP3S0", 4),
		tok("extraordinária", "AQ0FS0", 18),
	), "circunstância")
}

// de maneira obvia → óbvia (verb→adj; maneira branch).
func TestPortugueseAccentuationCheckRule_AdjManner(t *testing.T) {
	rule := NewPortugueseAccentuationCheckRule(nil)
	assertIncorrect(t, rule, sent(
		tok("", languagetool.SentenceStartTagName, 0),
		tok("De", "SPS00", 0),
		tok("maneira", "NCFS000", 3),
		tok("obvia", "VMIP3S0", 12),
	), "óbvia")
}

// a magnifica conservação → magnífica (article + adj + following noun).
func TestPortugueseAccentuationCheckRule_AdjArticleNoun(t *testing.T) {
	rule := NewPortugueseAccentuationCheckRule(nil)
	assertIncorrect(t, rule, sent(
		tok("", languagetool.SentenceStartTagName, 0),
		tok("a", "DA0FS0", 0),
		tok("magnifica", "VMIP3S0", 2),
		tok("conservação", "NCFS000", 12),
	), "magnífica")
}

// uma nova formula que → fórmula (det + adj + noun + que).
func TestPortugueseAccentuationCheckRule_FormulaQue(t *testing.T) {
	rule := NewPortugueseAccentuationCheckRule(nil)
	assertIncorrect(t, rule, sent(
		tok("", languagetool.SentenceStartTagName, 0),
		tok("uma", "DI0FS0", 0),
		tok("nova", "AQ0FS0", 4),
		tok("formula", "VMIP3S0", 9),
		tok("que", "PR0CN000", 17),
	), "fórmula")
}

// de positiva influencia → influência (SPS00 + adj + noun overwrite branch).
func TestPortugueseAccentuationCheckRule_DePositivaInfluencia(t *testing.T) {
	rule := NewPortugueseAccentuationCheckRule(nil)
	assertIncorrect(t, rule, sent(
		tok("", languagetool.SentenceStartTagName, 0),
		tok("de", "SPS00", 0),
		tok("positiva", "AQ0FS0", 3),
		tok("influencia", "VMIP3S0", 12),
	), "influência")
}

// de + adj-key + following noun (BEFORE_ADJECTIVE + next noun).
// "especifica" is in adj map; "disputa" is noun.
func TestPortugueseAccentuationCheckRule_AdjBeforeNoun(t *testing.T) {
	rule := NewPortugueseAccentuationCheckRule(nil)
	require.Contains(t, rule.relevantWords2, "especifica")
	assertIncorrect(t, rule, sent(
		tok("", languagetool.SentenceStartTagName, 0),
		tok("de", "SPS00", 0),
		tok("especifica", "VMIP3S0", 3),
		tok("disputa", "NCFS000", 14),
	), "específica")
}

// Article + noun + que (Java "les circumstancies que ens envolten").
func TestPortugueseAccentuationCheckRule_ArticleQue(t *testing.T) {
	rule := NewPortugueseAccentuationCheckRule(nil)
	assertIncorrect(t, rule, sent(
		tok("", languagetool.SentenceStartTagName, 0),
		tok("A", "DA0FS0", 0),
		tok("circunstancia", "VMIP3S0", 2),
		tok("que", "PR0CN000", 16),
		tok("nos", "PP1CP000", 20),
		tok("envolve", "VMIP3S0", 24),
	), "circunstância")
}

// Two matches in one sentence (Java multi-match assertEquals(2, …)).
func TestPortugueseAccentuationCheckRule_TwoMatches(t *testing.T) {
	rule := NewPortugueseAccentuationCheckRule(nil)
	// "A circunstancia que X e a distancia de Y."
	// 1) article+que on circunstancia
	// 2) article+de on distancia (or det+noun)
	s := sent(
		tok("", languagetool.SentenceStartTagName, 0),
		tok("A", "DA0FS0", 0),
		tok("circunstancia", "VMIP3S0", 2),
		tok("que", "PR0CN000", 16),
		tok("surge", "VMIP3S0", 20),
		tok("e", "CC", 26),
		tok("a", "DA0FS0", 28),
		tok("distancia", "VMIP3S0", 30),
		tok("de", "SPS00", 40),
		tok("casa", "NCFS000", 43),
	)
	matches := rule.Match(s)
	require.Len(t, matches, 2)
	require.Equal(t, []string{"circunstância"}, matches[0].GetSuggestedReplacements())
	require.Equal(t, []string{"distância"}, matches[1].GetSuggestedReplacements())
}

// Positions twin of AccentuationCheckRuleTest.testPositions shape.
func TestPortugueseAccentuationCheckRule_Positions(t *testing.T) {
	rule := NewPortugueseAccentuationCheckRule(nil)
	// "São circunstancia extraordinária." — key starts at UTF-16 index 4 ("São ")
	const start = 4
	key := "circunstancia"
	s := sent(
		tok("", languagetool.SentenceStartTagName, 0),
		tok("São", "VMIP3P0", 0),
		tok(key, "VMIP3S0", start),
		tok("extraordinária", "AQ0FS0", start+len(key)+1),
	)
	matches := rule.Match(s)
	require.Len(t, matches, 1)
	require.Equal(t, start, matches[0].GetFromPos())
	require.Equal(t, start+len(key), matches[0].GetToPos())
	require.Equal(t, []string{"circunstância"}, matches[0].GetSuggestedReplacements())
}

func TestPortugueseAccentuationCheckRule_NoMap(t *testing.T) {
	r := NewPortugueseAccentuationCheckRule(nil)
	require.Empty(t, r.Match(languagetool.AnalyzePlain("a casa")))
}

// First real token is lowercased for map lookup (Java i==1 → toLowerCase).
func TestPortugueseAccentuationCheckRule_FirstTokenLowercased(t *testing.T) {
	rule := NewPortugueseAccentuationCheckRule(nil)
	// "Copia de algo extra" — broad branch (i < len-2, prev not NOT_IN_PREV/RG)
	s := sent(
		tok("", languagetool.SentenceStartTagName, 0),
		tok("Copia", "VMIP3S0", 0),
		tok("de", "SPS00", 6),
		tok("algo", "PI0CS000", 9),
		tok("extra", "AQ0MS0", 14),
	)
	matches := rule.Match(s)
	require.Len(t, matches, 1)
	require.Equal(t, []string{"cópia"}, matches[0].GetSuggestedReplacements())
}

// Conjugated verb + article surface + noun (Java "fem la copia").
func TestPortugueseAccentuationCheckRule_VerbArticleNoun(t *testing.T) {
	rule := NewPortugueseAccentuationCheckRule(nil)
	assertIncorrect(t, rule, sent(
		tok("", languagetool.SentenceStartTagName, 0),
		tok("Fazemos", "VMIP1P0", 0),
		tok("a", "DA0FS0", 8),
		tok("copia", "VMIP3S0", 10),
	), "cópia")
}

// Adj before noun reading (Java "les seves contraries") — prev adj + noun map entry.
func TestPortugueseAccentuationCheckRule_PrevAdjNoun(t *testing.T) {
	rule := NewPortugueseAccentuationCheckRule(nil)
	// "nova formula" without "que" and without det two back — uses adj-prev branch
	// NOME_FS + ADJETIVO_FS on prev; current not VERBO_3S
	assertIncorrect(t, rule, sent(
		tok("", languagetool.SentenceStartTagName, 0),
		tok("boa", "AQ0FS0", 0),
		tok("formula", "VMIP1S0", 4), // 1S so VERBO_3S fails
	), "fórmula")
}

// --- helpers ---

func tok(surface, pos string, start int) *languagetool.AnalyzedTokenReadings {
	var p *string
	if pos != "" {
		pp := pos
		p = &pp
	}
	return languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken(surface, p, nil), start)
}

func sent(tokens ...*languagetool.AnalyzedTokenReadings) *languagetool.AnalyzedSentence {
	return languagetool.NewAnalyzedSentence(tokens)
}

func assertCorrect(t *testing.T, rule *PortugueseAccentuationCheckRule, s *languagetool.AnalyzedSentence) {
	t.Helper()
	matches := rule.Match(s)
	require.Equal(t, 0, len(matches), "expected 0 matches, got %d for text with %d tokens", len(matches), len(s.GetTokensWithoutWhitespace()))
}

func assertIncorrect(t *testing.T, rule *PortugueseAccentuationCheckRule, s *languagetool.AnalyzedSentence, wantSug string) {
	t.Helper()
	matches := rule.Match(s)
	require.Equal(t, 1, len(matches), "expected 1 match, got %d", len(matches))
	require.Equal(t, []string{wantSug}, matches[0].GetSuggestedReplacements())
}
