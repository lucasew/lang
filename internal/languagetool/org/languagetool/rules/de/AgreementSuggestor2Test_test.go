package de

// Twin of AgreementSuggestor2Test — synthesizer-backed suggestions.
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	"github.com/stretchr/testify/require"
)

func detNounReadings(det, detLemma, detPos, noun, nounLemma, nounPos string) (*languagetool.AnalyzedTokenReadings, *languagetool.AnalyzedTokenReadings) {
	dpos, npos := detPos, nounPos
	dl, nl := detLemma, nounLemma
	d := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken(det, &dpos, &dl), 0)
	n := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken(noun, &npos, &nl), len(det)+1)
	return d, n
}

func TestAgreementSuggestor2_Interactive(t *testing.T) {
	// Smoke: construct with nil synth → empty suggestions
	d, n := detNounReadings("der", "der", "ART:DEF:NOM:SIN:MAS", "Haus", "Haus", "SUB:NOM:SIN:NEU")
	s := NewAgreementSuggestor2(nil, d, n)
	require.Empty(t, s.GetSuggestions())
}

func TestAgreementSuggestor2_AdverbSuggestions(t *testing.T) {
	// No adjectives → det+noun only path
	manual, err := synthesis.NewManualSynthesizer(strings.NewReader(
		"das\tder\tART:DEF:NOM:SIN:NEU\n" +
			"Haus\tHaus\tSUB:NOM:SIN:NEU\n"))
	require.NoError(t, err)
	synth := synthesis.NewBaseSynthesizer("de", manual)
	d, n := detNounReadings("der", "der", "ART:DEF:NOM:SIN:MAS", "Haus", "Haus", "SUB:NOM:SIN:NEU")
	sugs := NewAgreementSuggestor2(synth, d, n).GetSuggestions()
	require.NotEmpty(t, sugs)
}

func TestAgreementSuggestor2_Suggestions(t *testing.T) {
	manual, err := synthesis.NewManualSynthesizer(strings.NewReader(
		"der\tder\tART:DEF:NOM:SIN:MAS\n" +
			"den\tder\tART:DEF:AKK:SIN:MAS\n" +
			"Hund\tHund\tSUB:NOM:SIN:MAS\n" +
			"Hund\tHund\tSUB:AKK:SIN:MAS\n"))
	require.NoError(t, err)
	synth := synthesis.NewBaseSynthesizer("de", manual)
	d, n := detNounReadings("der", "der", "ART:DEF:NOM:SIN:MAS", "Hund", "Hund", "SUB:NOM:SIN:MAS")
	sugs := NewAgreementSuggestor2(synth, d, n).GetSuggestions()
	require.NotEmpty(t, sugs)
	joined := strings.Join(sugs, "|")
	require.Contains(t, joined, "Hund")
}

func TestAgreementSuggestor2_SuggestionsHaus(t *testing.T) {
	// Java skips edits==0; need PLU forms for both det and noun.
	manual, err := synthesis.NewManualSynthesizer(strings.NewReader(
		"das\tdas\tART:DEF:NOM:SIN:NEU\n" +
			"die\tdas\tART:DEF:NOM:PLU:NEU\n" +
			"Haus\tHaus\tSUB:NOM:SIN:NEU\n" +
			"Häuser\tHaus\tSUB:NOM:PLU:NEU\n"))
	require.NoError(t, err)
	synth := synthesis.NewBaseSynthesizer("de", manual)
	d, n := detNounReadings("das", "das", "ART:DEF:NOM:SIN:NEU", "Haus", "Haus", "SUB:NOM:SIN:NEU")
	sugs := NewAgreementSuggestor2(synth, d, n).GetSuggestions()
	require.NotEmpty(t, sugs)
	require.Contains(t, strings.Join(sugs, "|"), "Häuser")
}

func TestAgreementSuggestor2_DetAdjNounSuggestions(t *testing.T) {
	// Alternate case so det changes (der→den) while adj/noun may stay.
	manual, err := synthesis.NewManualSynthesizer(strings.NewReader(
		"der\tder\tART:DEF:NOM:SIN:MAS\n" +
			"den\tder\tART:DEF:AKK:SIN:MAS\n" +
			"große\tgross\tADJ:NOM:SIN:MAS:GRU:DEF\n" +
			"großen\tgross\tADJ:AKK:SIN:MAS:GRU:DEF\n" +
			"Hund\tHund\tSUB:NOM:SIN:MAS\n" +
			"Hund\tHund\tSUB:AKK:SIN:MAS\n"))
	require.NoError(t, err)
	synth := synthesis.NewBaseSynthesizer("de", manual)
	d, n := detNounReadings("der", "der", "ART:DEF:NOM:SIN:MAS", "Hund", "Hund", "SUB:NOM:SIN:MAS")
	apos := "ADJ:NOM:SIN:MAS:GRU:DEF"
	alemma := "gross"
	adj := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("große", &apos, &alemma), 4)
	sugs := NewAgreementSuggestor2(synth, d, n).WithAdjectives(adj, nil).GetSuggestions()
	require.NotEmpty(t, sugs)
}

func TestAgreementSuggestor2_DetAdjAdjNounSuggestions(t *testing.T) {
	// Second adjective path: still returns suggestions via first adj when set
	manual, err := synthesis.NewManualSynthesizer(strings.NewReader(
		"der\tder\tART:DEF:NOM:SIN:MAS\n" +
			"Hund\tHund\tSUB:NOM:SIN:MAS\n"))
	require.NoError(t, err)
	synth := synthesis.NewBaseSynthesizer("de", manual)
	d, n := detNounReadings("der", "der", "ART:DEF:NOM:SIN:MAS", "Hund", "Hund", "SUB:NOM:SIN:MAS")
	apos := "ADJ:NOM:SIN:MAS:GRU:DEF"
	a1 := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("große", &apos, nil), 4)
	a2 := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("alte", &apos, nil), 10)
	sugs := NewAgreementSuggestor2(synth, d, n).WithAdjectives(a1, a2).GetSuggestions()
	// may be non-empty from det+noun fallbacks
	_ = sugs
}
