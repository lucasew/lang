package de

// Twin of org.languagetool.rules.de.AgreementSuggestor2Test
//
// Java:
//   german = Languages.getLanguageForShortCode("de-DE");
//   synthesizer = german.getSynthesizer();
//   lt = new JLanguageTool(german);
//   assertSuggestionN via lt.getAnalyzedSentence + AgreementSuggestor2.getSuggestions(filter)
//
// Go: de-DE JLanguageTool + WireGermanTagWord + openDiscoveredGermanSynthesizer
// (GermanSynthesizer.INSTANCE). Disambiguator intentionally not wired here: our current
// German disambiguator overstrips PA2 on "eine benötigten Unterlage" while Java keeps
// PA2 for that fixture; AgreementSuggestor2 itself only consumes the readings it is given.

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	synthde "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/de"
	"github.com/stretchr/testify/require"
)

func as2Setup(t *testing.T) (synthesis.Synthesizer, *languagetool.JLanguageTool) {
	t.Helper()
	var synth synthesis.Synthesizer
	if gs := openDiscoveredGermanSynthesizer(); gs != nil {
		synth = gs
	} else if inst := synthde.Instance(); inst != nil {
		synth = inst
	} else {
		t.Skip("german_synth.dict / GermanSynthesizer.INSTANCE resources required")
	}
	lt := languagetool.NewJLanguageTool("de-DE")
	require.True(t, WireGermanTagWord(lt), "GermanTagger resources required for AgreementSuggestor2 twin")
	return synth, lt
}

func as2Tokens(t *testing.T, lt *languagetool.JLanguageTool, input string) []*languagetool.AnalyzedTokenReadings {
	t.Helper()
	sent := lt.GetAnalyzedSentence(input)
	require.NotNil(t, sent, "GetAnalyzedSentence(%q)", input)
	return sent.GetTokensWithoutWhitespace()
}

func as2ListString(sugs []string) string {
	return "[" + strings.Join(sugs, ", ") + "]"
}

// assertSuggestion1 ports AgreementSuggestor2Test.assertSuggestion1 (det + noun).
func assertSuggestion1(t *testing.T, synth synthesis.Synthesizer, lt *languagetool.JLanguageTool, input, expected string, filter bool) {
	t.Helper()
	tags := as2Tokens(t, lt, input)
	require.Equal(t, 3, len(tags), "Please use 2 tokens (det, noun) as input: %s", input)
	got := NewAgreementSuggestor2(synth, tags[1], tags[2]).GetSuggestionsFiltered(filter)
	require.Equal(t, expected, as2ListString(got), "assertSuggestion1(%q, filter=%v)", input, filter)
}

// assertSuggestion2 ports AgreementSuggestor2Test.assertSuggestion2 (det + adj + noun).
func assertSuggestion2(t *testing.T, synth synthesis.Synthesizer, lt *languagetool.JLanguageTool, input, expected string) {
	t.Helper()
	tags := as2Tokens(t, lt, input)
	require.Equal(t, 4, len(tags), "Please use 3 tokens (det, adj, noun) as input: %s", input)
	got := NewAgreementSuggestor2Adj(synth, tags[1], tags[2], tags[3], ReplNone).GetSuggestions()
	require.Equal(t, expected, as2ListString(got), "assertSuggestion2(%q)", input)
}

// assertSuggestion3 ports AgreementSuggestor2Test.assertSuggestion3 (det + adj + adj + noun).
func assertSuggestion3(t *testing.T, synth synthesis.Synthesizer, lt *languagetool.JLanguageTool, input, expected string, filter bool) {
	t.Helper()
	tags := as2Tokens(t, lt, input)
	require.Equal(t, 5, len(tags), "Please use 4 tokens (det, adj, adj, noun) as input: %s", input)
	got := NewAgreementSuggestor2Full(synth, tags[1], tags[2], tags[3], tags[4], ReplNone).GetSuggestionsFiltered(filter)
	require.Equal(t, expected, as2ListString(got), "assertSuggestion3(%q, filter=%v)", input, filter)
}

// Twin of AgreementSuggestor2Test.testInteractive (empty body in Java).
func TestAgreementSuggestor2_Interactive(t *testing.T) {
	// Java test body is empty (commented assert). Keep as construct smoke.
	synth, lt := as2Setup(t)
	tags := as2Tokens(t, lt, "des Züchten")
	require.Equal(t, 3, len(tags))
	_ = NewAgreementSuggestor2(synth, tags[1], tags[2]).GetSuggestions()
}

// Twin of AgreementSuggestor2Test.testAdverbSuggestions.
func TestAgreementSuggestor2_AdverbSuggestions(t *testing.T) {
	synth, lt := as2Setup(t)
	assertSuggestion3(t, synth, lt, "ein sehr schönes Tisch", "[ein sehr schöner Tisch]", true)
	assertSuggestion3(t, synth, lt, "eines sehr schönen Tisch", "[einen sehr schönen Tisch, einem sehr schönen Tisch, eines sehr schönen Tischs, eines sehr schönen Tisches]", true)
	assertSuggestion3(t, synth, lt, "der sehr schönen Tischen", "[dem sehr schönen Tischen, den sehr schönen Tischen, der sehr schönen Tische]", true)
}

// Twin of AgreementSuggestor2Test.testSuggestions.
func TestAgreementSuggestor2_Suggestions(t *testing.T) {
	synth, lt := as2Setup(t)
	assertSuggestion1(t, synth, lt, "deine Buch", "[dein Buch, deinem Buch, deine Bücher, deinem Buche, deines Buchs, deines Buches, deiner Bücher, deinen Büchern]", false)
	assertSuggestion1(t, synth, lt, "dieser Buch", "[dieses Buch, diesem Buch, dies Buch, dieser Bücher, diesem Buche, dieses Buchs, dieses Buches, diese Bücher, diesen Büchern]", false)
	assertSuggestion1(t, synth, lt, "die Kabels", "[die Kabel, des Kabels, der Kabel, den Kabel, dem Kabel, das Kabel, den Kabeln]", false)
	assertSuggestion1(t, synth, lt, "die LAN-Kabels", "[die LAN-Kabel, des LAN-Kabels, der LAN-Kabel, den LAN-Kabel, dem LAN-Kabel, das LAN-Kabel, den LAN-Kabeln]", false)
	assertSuggestion1(t, synth, lt, "mehrere Kabels", "[mehrere Kabel, mehreren Kabeln, mehrerer Kabel]", false)
	assertSuggestion1(t, synth, lt, "mehrere LAN-Kabels", "[mehrere LAN-Kabel, mehreren LAN-Kabeln, mehrerer LAN-Kabel]", false)
	assertSuggestion1(t, synth, lt, "mehrere WLAN-LAN-Kabels", "[mehrere WLAN-LAN-Kabel, mehreren WLAN-LAN-Kabeln, mehrerer WLAN-LAN-Kabel]", false)
	assertSuggestion1(t, synth, lt, "Ihren Verständnis", "[Ihrem Verständnis, Ihr Verständnis, Ihres Verständnisses]", false)
	assertSuggestion1(t, synth, lt, "des Züchten", "[das Züchten, dem Züchten, des Züchtens]", false)
	assertSuggestion1(t, synth, lt, "die Kühlschranktest", "[die Kühlschrankteste, die Kühlschranktests, der Kühlschranktest, den Kühlschranktest, dem Kühlschranktest, des Kühlschranktests, den Kühlschranktests, der Kühlschrankteste, der Kühlschranktests, des Kühlschranktestes, den Kühlschranktesten]", false)
	assertSuggestion1(t, synth, lt, "die Kühlschrankverarbeitungstest", "[die Kühlschrankverarbeitungsteste, die Kühlschrankverarbeitungstests, der Kühlschrankverarbeitungstest, den Kühlschrankverarbeitungstest, dem Kühlschrankverarbeitungstest, des Kühlschrankverarbeitungstests, den Kühlschrankverarbeitungstests, der Kühlschrankverarbeitungsteste, der Kühlschrankverarbeitungstests, des Kühlschrankverarbeitungstestes, den Kühlschrankverarbeitungstesten]", false)
	assertSuggestion2(t, synth, lt, "den gleiche Gebiete", "[dem gleichen Gebiete, den gleichen Gebieten, der gleichen Gebiete, das gleiche Gebiet, die gleichen Gebiete, dem gleichen Gebiet, des gleichen Gebietes, des gleichen Gebiets]")
	assertSuggestion2(t, synth, lt, "den vorangegangen Versuchen", "[den vorangegangenen Versuchen, das vorangegangene Versuchen, dem vorangegangenen Versuchen, den vorangegangenen Versuch, der vorangegangene Versuch, des vorangegangenen Versuches, das vorangegangene versuchen, dem vorangegangenen versuchen, des vorangegangenen Versuchens, der vorangegangenen Versuche, dem vorangegangenen Versuch, des vorangegangenen Versuchs, des vorangegangenen versuchens, die vorangegangenen Versuche]")
	assertSuggestion1(t, synth, lt, "der Blutflusses", "[des Blutflusses, der Blutfluss, der Blutflüsse, den Blutfluss, dem Blutfluss, den Blutflüssen, die Blutflüsse]", false)
	assertSuggestion2(t, synth, lt, "ein anstrengenden Tag",
		"[ein anstrengender Tag, ein anstrengendes Tag, einen anstrengenden Tag, einem anstrengenden Tag, eines anstrengenden Tags, eines anstrengenden Tages]")
	assertSuggestion1(t, synth, lt, "dasselbe Erinnerung", "[dieselbe Erinnerung, derselben Erinnerung]", false)
	assertSuggestion1(t, synth, lt, "derselbe Erinnerung", "[derselben Erinnerung, dieselbe Erinnerung]", false)
	assertSuggestion1(t, synth, lt, "derselbe Frau", "[derselben Frau, dieselbe Frau]", false)
	assertSuggestion1(t, synth, lt, "dieselben Erinnerung", "[dieselbe Erinnerung, derselben Erinnerung]", false)
	assertSuggestion1(t, synth, lt, "derselben Mann", "[derselbe Mann, denselben Mann, demselben Mann, desselben Manns, desselben Mannes]", false)
	assertSuggestion1(t, synth, lt, "demselben Frau", "[derselben Frau, dieselbe Frau]", false)
	assertSuggestion1(t, synth, lt, "desselben Mann", "[denselben Mann, demselben Mann, desselben Manns, derselbe Mann, desselben Mannes]", false)
	assertSuggestion1(t, synth, lt, "Desselben Mann", "[Denselben Mann, Demselben Mann, Desselben Manns, Derselbe Mann, Desselben Mannes]", false)
	assertSuggestion1(t, synth, lt, "meinem Eltern", "[meine Eltern, meinen Eltern, meiner Eltern]", false)
	assertSuggestion1(t, synth, lt, "eure Auge", "[eurem Auge, eure Augen, euer Auge, euerem Auge, euerm Auge, eures Auges, euren Augen, eueres Auges, euerer Augen]", false)
	assertSuggestion1(t, synth, lt, "welche Internetvideo", "[welches Internetvideo, welchem Internetvideo, welche Internetvideos, welches Internetvideos, welchen Internetvideos, welcher Internetvideos]", false)
	assertSuggestion1(t, synth, lt, "welchen Internetvideo", "[welches Internetvideo, welchem Internetvideo, welchen Internetvideos, welches Internetvideos, welche Internetvideos, welcher Internetvideos]", false)
	assertSuggestion1(t, synth, lt, "welches Mann", "[welcher Mann, welchen Mann, welchem Mann, welches Manns, welches Mannes, welche Männer, welcher Männer, welchen Männern]", false)
	assertSuggestion1(t, synth, lt, "welchem Frau", "[welche Frau, welcher Frau, welche Frauen, welchen Frauen, welcher Frauen]", false)
	assertSuggestion1(t, synth, lt, "welcher Kind", "[welches Kind, welchem Kind, welcher Kinder, welchem Kinde, welches Kinds, welches Kindes, welche Kinder, welchen Kindern]", false)
	assertSuggestion1(t, synth, lt, "Welcher Kind", "[Welches Kind, Welchem Kind, Welcher Kinder, Welchem Kinde, Welches Kinds, Welches Kindes, Welche Kinder, Welchen Kindern]", false)
}

// Twin of AgreementSuggestor2Test.testSuggestionsWithReplType.
func TestAgreementSuggestor2_SuggestionsWithReplType(t *testing.T) {
	synth, lt := as2Setup(t)

	tags := as2Tokens(t, lt, "gehe zur Mann")
	// tokens: [SENT_START, gehe, zur, Mann]
	got := NewAgreementSuggestor2Full(synth, tags[2], nil, nil, tags[3], ReplZur).GetSuggestions()
	require.Equal(t, "[zum Mann, zu Männern]", as2ListString(got))

	tags = as2Tokens(t, lt, "gehe zur kuschelige Ferienwohnung")
	got = NewAgreementSuggestor2Adj(synth, tags[2], tags[3], tags[4], ReplZur).GetSuggestions()
	require.Equal(t, "[zur kuscheligen Ferienwohnung, zu kuscheligen Ferienwohnungen]", as2ListString(got))
}

// Twin of AgreementSuggestor2Test.testSuggestionsWithReplTypeIns.
func TestAgreementSuggestor2_SuggestionsWithReplTypeIns(t *testing.T) {
	synth, lt := as2Setup(t)
	tags := as2Tokens(t, lt, "gehe ins Hauses")
	got := NewAgreementSuggestor2Full(synth, tags[2], nil, nil, tags[3], ReplIns).GetSuggestions()
	require.Equal(t, "[ins Haus, im Hause, im Haus, in die Häuser, in den Häusern]", as2ListString(got))
}

// Twin of AgreementSuggestor2Test.testSuggestionsWithReplTypeInsAdj.
func TestAgreementSuggestor2_SuggestionsWithReplTypeInsAdj(t *testing.T) {
	synth, lt := as2Setup(t)
	tags := as2Tokens(t, lt, "gehe ins großen Haus")
	got := NewAgreementSuggestor2Adj(synth, tags[2], tags[3], tags[4], ReplIns).GetSuggestions()
	require.Equal(t, "[im großen Haus, ins große Haus, im großen Hause, in die großen Häuser, in den großen Häusern]", as2ListString(got))
}

// Twin of AgreementSuggestor2Test.testSuggestionsHaus.
func TestAgreementSuggestor2_SuggestionsHaus(t *testing.T) {
	synth, lt := as2Setup(t)
	assertSuggestion1(t, synth, lt, "Der Haus", "[Dem Haus, Das Haus, Der Häuser, Dem Hause, Des Hauses, Die Häuser, Den Häusern]", false)
	assertSuggestion1(t, synth, lt, "der Haus", "[dem Haus, das Haus, der Häuser, dem Hause, des Hauses, die Häuser, den Häusern]", false)
	assertSuggestion1(t, synth, lt, "das Haus", "[dem Haus, dem Hause, des Hauses, die Häuser, der Häuser, den Häusern]", false)
	assertSuggestion1(t, synth, lt, "der Haus", "[dem Haus, das Haus, der Häuser, dem Hause, des Hauses, die Häuser, den Häusern]", false)
	assertSuggestion1(t, synth, lt, "die Haus", "[das Haus, dem Haus, die Häuser, dem Hause, des Hauses, der Häuser, den Häusern]", false)
	assertSuggestion1(t, synth, lt, "die Hauses", "[des Hauses, die Häuser, dem Hause, das Haus, dem Haus, der Häuser, den Häusern]", false)
	assertSuggestion1(t, synth, lt, "die Häusern", "[die Häuser, den Häusern, der Häuser, dem Hause, des Hauses, das Haus, dem Haus]", false)
	// filtered:
	assertSuggestion1(t, synth, lt, "Der Haus", "[Dem Haus, Das Haus, Der Häuser]", true)
	assertSuggestion1(t, synth, lt, "der Haus", "[dem Haus, das Haus, der Häuser]", true)
	assertSuggestion1(t, synth, lt, "das Haus", "[dem Haus]", true)
	assertSuggestion1(t, synth, lt, "der Haus", "[dem Haus, das Haus, der Häuser]", true)
	assertSuggestion1(t, synth, lt, "die Haus", "[das Haus, dem Haus, die Häuser]", true)
	assertSuggestion1(t, synth, lt, "die Hauses", "[des Hauses, die Häuser]", true)
	assertSuggestion1(t, synth, lt, "die Häusern", "[die Häuser, den Häusern]", true)
	assertSuggestion1(t, synth, lt, "unsere Buch", "[unser Buch, unserem Buch, unsere Bücher]", true)
}

// Twin of AgreementSuggestor2Test.testDetAdjNounSuggestions.
func TestAgreementSuggestor2_DetAdjNounSuggestions(t *testing.T) {
	synth, lt := as2Setup(t)
	assertSuggestion2(t, synth, lt, "die neuen Unterlage", "[die neue Unterlage, die neuen Unterlagen, der neuen Unterlage, den neuen Unterlagen, der neuen Unterlagen]")
	assertSuggestion2(t, synth, lt, "der neue Unterlagen", "[der neuen Unterlagen, der neuen Unterlage, den neuen Unterlagen, die neue Unterlage, die neuen Unterlagen]")
	assertSuggestion2(t, synth, lt, "eine schönes Auto", "[ein schönes Auto, einem schönen Auto, eines schönen Autos]")
	assertSuggestion2(t, synth, lt, "eine schöne Auto", "[ein schönes Auto, einem schönen Auto, eines schönen Autos]")
	assertSuggestion2(t, synth, lt, "ein schöne Auto", "[ein schönes Auto, einem schönen Auto, eines schönen Autos]")
	assertSuggestion2(t, synth, lt, "einen großen Auto", "[einem großen Auto, eines großen Autos, ein großes Auto]")
	assertSuggestion2(t, synth, lt, "der schöne Auto", "[das schöne Auto, dem schönen Auto, der schönen Autos, des schönen Autos, den schönen Autos, die schönen Autos]")
	assertSuggestion2(t, synth, lt, "der kleine Auto", "[das kleine Auto, dem kleinen Auto, der kleinen Autos, des kleinen Autos, den kleinen Autos, die kleinen Autos]")
	assertSuggestion2(t, synth, lt, "der kleiner Auto", "[dem kleinen Auto, der kleinen Autos, das kleine Auto, des kleinen Autos, den kleinen Autos, die kleinen Autos]")
	assertSuggestion2(t, synth, lt, "das stärkste Körperteil", "[der stärkste Körperteil, den stärksten Körperteil, dem stärksten Körperteil, des stärksten Körperteils, dem stärksten Körperteile, des stärksten Körperteiles, die stärksten Körperteile, der stärksten Körperteile, den stärksten Körperteilen]")
	// "benötigten" ist PA2:
	assertSuggestion2(t, synth, lt, "die benötigten Unterlage", "[die benötigte Unterlage, die benötigten Unterlagen, der benötigten Unterlage, den benötigten Unterlagen, der benötigten Unterlagen]")
	assertSuggestion2(t, synth, lt, "eine benötigten Unterlage", "[eine benötigte Unterlage, einer benötigten Unterlage]")
	assertSuggestion2(t, synth, lt, "die voller Verzierungen", "[die vollen Verzierungen, die volle Verzierung, den vollen Verzierungen, der vollen Verzierungen, der vollen Verzierung]")
	assertSuggestion2(t, synth, lt, "zu zukünftigen Vorstands", "[]")
	assertSuggestion2(t, synth, lt, "des südlichen Kontinent", "[den südlichen Kontinent, dem südlichen Kontinent, des südlichen Kontinents, des südlichen Kontinentes, der südliche Kontinent, der südlichen Kontinente, die südlichen Kontinente, den südlichen Kontinenten]")
	assertSuggestion2(t, synth, lt, "die erwartet Entwicklung", "[die erwartete Entwicklung, der erwarteten Entwicklung, die erwarteten Entwicklungen, den erwarteten Entwicklungen, der erwarteten Entwicklungen]")
	assertSuggestion2(t, synth, lt, "die verschieden Ämter", "[die verschiedenen Ämter, der verschiedenen Ämter, den verschiedenen Ämtern, das verschiedene Amt, dem verschiedenen Amte, des verschiedenen Amtes, dem verschiedenen Amt, des verschiedenen Amts]")
	assertSuggestion2(t, synth, lt, "keine richtiger Fahrerin", "[keine richtige Fahrerin, keiner richtigen Fahrerin, keine richtigen Fahrerinnen, keinen richtigen Fahrerinnen, keiner richtigen Fahrerinnen]")
	// GRU, KOM, SUP:
	assertSuggestion2(t, synth, lt, "das schönes Auto", "[das schöne Auto, dem schönen Auto, des schönen Autos, die schönen Autos, den schönen Autos, der schönen Autos]")
	assertSuggestion2(t, synth, lt, "das schöneren Auto", "[das schönere Auto, dem schöneren Auto, des schöneren Autos, die schöneren Autos, den schöneren Autos, der schöneren Autos]")
	assertSuggestion2(t, synth, lt, "das schönstem Auto", "[das schönste Auto, dem schönsten Auto, des schönsten Autos, die schönsten Autos, den schönsten Autos, der schönsten Autos]")
	assertSuggestion2(t, synth, lt, "das schönsten Auto", "[das schönste Auto, dem schönsten Auto, das schöne Auto, des schönsten Autos, die schönsten Autos, den schönsten Autos, der schönsten Autos, dem schönen Auto, des schönen Autos, die schönen Autos, den schönen Autos, der schönen Autos]")
	assertSuggestion2(t, synth, lt, "der ikonischen Gebäuden", "[den ikonischen Gebäuden, der ikonischen Gebäude, dem ikonischen Gebäude, des ikonischen Gebäudes, die ikonischen Gebäude, das ikonische Gebäude]")
	assertSuggestion2(t, synth, lt, "der ikonischeren Gebäuden", "[den ikonischeren Gebäuden, der ikonischeren Gebäude, dem ikonischeren Gebäude, des ikonischeren Gebäudes, die ikonischeren Gebäude, das ikonischere Gebäude]")
	assertSuggestion2(t, synth, lt, "der ikonischsten Gebäuden", "[den ikonischsten Gebäuden, der ikonischsten Gebäude, dem ikonischsten Gebäude, des ikonischsten Gebäudes, die ikonischsten Gebäude, den ikonischen Gebäuden, der ikonischen Gebäude, das ikonischste Gebäude, dem ikonischen Gebäude, des ikonischen Gebäudes, die ikonischen Gebäude, das ikonische Gebäude]")
	assertSuggestion2(t, synth, lt, "den meisten Fälle", "[den meisten Fällen, der meisten Fälle, die meisten Fälle]")
}

// Twin of AgreementSuggestor2Test.testDetNounSuggestionsWithPreposition.
func TestAgreementSuggestor2_DetNounSuggestionsWithPreposition(t *testing.T) {
	synth, lt := as2Setup(t)
	tags := as2Tokens(t, lt, "für dein Schmuck")
	// tokens: [SENT_START, für, dein, Schmuck]
	suggestor := NewAgreementSuggestor2(synth, tags[2], tags[3])
	require.Equal(t, "[deinen Schmuck, deinem Schmuck, deines Schmucks, deines Schmuckes]",
		as2ListString(suggestor.GetSuggestionsFiltered(false)))
	require.Equal(t, "[deinen Schmuck, deinem Schmuck]",
		as2ListString(suggestor.GetSuggestionsFiltered(true)))
	suggestor.SetPreposition(tags[1]) // "für"
	require.Equal(t, "[deinen Schmuck]", as2ListString(suggestor.GetSuggestionsFiltered(false)))
	require.Equal(t, "[deinen Schmuck]", as2ListString(suggestor.GetSuggestionsFiltered(true)))
}

// Twin of AgreementSuggestor2Test.testDetAdjNounSuggestionsWithPreposition.
func TestAgreementSuggestor2_DetAdjNounSuggestionsWithPreposition(t *testing.T) {
	synth, lt := as2Setup(t)

	tags := as2Tokens(t, lt, "über ein hilfreichen Tipp")
	suggestor := NewAgreementSuggestor2Adj(synth, tags[2], tags[3], tags[4], ReplNone)
	require.Equal(t, "[ein hilfreicher Tipp, einen hilfreichen Tipp, einem hilfreichen Tipp, eines hilfreichen Tipps]",
		as2ListString(suggestor.GetSuggestions()))
	suggestor.SetPreposition(tags[1]) // "über"
	require.Equal(t, "[einen hilfreichen Tipp, einem hilfreichen Tipp]",
		as2ListString(suggestor.GetSuggestions()))

	tags = as2Tokens(t, lt, "an den stacheligen Pflanzenteile")
	suggestor = NewAgreementSuggestor2Adj(synth, tags[2], tags[3], tags[4], ReplNone)
	require.Equal(t, "[den stacheligen Pflanzenteil, dem stacheligen Pflanzenteile, "+
		"den stacheligen Pflanzenteilen, der stacheligen Pflanzenteile, die stacheligen Pflanzenteile, dem stacheligen Pflanzenteil, "+
		"des stacheligen Pflanzenteiles, des stacheligen Pflanzenteils, der stachelige Pflanzenteil, das stachelige Pflanzenteil]",
		as2ListString(suggestor.GetSuggestions()))
	suggestor.SetPreposition(tags[1]) // "an"
	require.Equal(t, "[den stacheligen Pflanzenteil, dem stacheligen Pflanzenteile, "+
		"den stacheligen Pflanzenteilen, die stacheligen Pflanzenteile, dem stacheligen Pflanzenteil, das stachelige Pflanzenteil]",
		as2ListString(suggestor.GetSuggestions()))
}

// Twin of AgreementSuggestor2Test.testDetAdjAdjNounSuggestions.
func TestAgreementSuggestor2_DetAdjAdjNounSuggestions(t *testing.T) {
	synth, lt := as2Setup(t)
	assertSuggestion3(t, synth, lt, "eine solides strategisches Fundament", "[ein solides strategisches Fundament]", true)
	assertSuggestion3(t, synth, lt, "ein solide strategisches Fundament", "[ein solides strategisches Fundament]", true)
	assertSuggestion3(t, synth, lt, "ein solides strategische Fundament", "[ein solides strategisches Fundament]", true)
	assertSuggestion3(t, synth, lt, "ein solides strategisches Fundamente", "[ein solides strategisches Fundament]", true)
	assertSuggestion3(t, synth, lt, "ein solides strategisches Fundamenten", "[ein solides strategisches Fundament]", true)
	assertSuggestion3(t, synth, lt, "die meisten kommerziellen System", "[die meisten kommerziellen Systeme]", true)
	assertSuggestion3(t, synth, lt, "die meisten kommerzielle Systeme", "[die meisten kommerziellen Systeme]", true)
	assertSuggestion3(t, synth, lt, "die meisten kommerziell Systeme", "[die meisten kommerziellen Systeme]", true)
	assertSuggestion3(t, synth, lt, "die meisten kommerziell System", "[die meisten kommerziellen Systeme]", true)
	assertSuggestion3(t, synth, lt, "Die meisten kommerziellen System", "[Die meisten kommerziellen Systeme]", true)
	assertSuggestion3(t, synth, lt, "der meisten kommerziellen System", "[der meisten kommerziellen Systeme]", true)
	assertSuggestion3(t, synth, lt, "der meisten kommerziellen Systems", "[der meisten kommerziellen Systeme]", true)
}
