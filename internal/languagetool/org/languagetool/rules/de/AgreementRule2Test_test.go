package de

// Twin of org.languagetool.rules.de.AgreementRule2Test
// Java: rule.match(lt.getAnalyzedSentence(s)) with GermanyGerman + AgreementRule2.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func ar2Setup(t *testing.T) (*languagetool.JLanguageTool, *AgreementRule2) {
	t.Helper()
	lt := languagetool.NewJLanguageTool("de-DE")
	if !WireGermanTagWord(lt) {
		t.Skip("GermanTagger resources unavailable")
	}
	_ = WireGermanDisambiguator(lt)
	rule := NewAgreementRule2(nil)
	// Java: GermanSynthesizer.INSTANCE used directly in getSuggestions
	if s := openDiscoveredGermanSynthesizer(); s != nil {
		rule = rule.WithSynth(s)
	}
	require.Equal(t, "DE_AGREEMENT2", rule.GetID())
	require.Contains(t, rule.GetDescription(), "Kongruenz von Adjektiv und Nomen")
	return lt, rule
}

// ar2Match ports rule.match(lt.getAnalyzedSentence(s)).
func ar2Match(lt *languagetool.JLanguageTool, rule *AgreementRule2, s string) []*rules.RuleMatch {
	sent := lt.GetAnalyzedSentence(s)
	if sent == nil {
		sents := lt.Analyze(s)
		if len(sents) == 0 {
			return nil
		}
		return rule.Match(sents[0])
	}
	return rule.Match(sent)
}

func ar2AssertGood(t *testing.T, lt *languagetool.JLanguageTool, rule *AgreementRule2, s string) {
	t.Helper()
	ms := ar2Match(lt, rule, s)
	require.Equal(t, 0, len(ms), "Found unexpected match in sentence %q: %v", s, ms)
}

func ar2AssertBad(t *testing.T, lt *languagetool.JLanguageTool, rule *AgreementRule2, s string) {
	t.Helper()
	ms := ar2Match(lt, rule, s)
	require.Equal(t, 1, len(ms), "Did not find one match in sentence %q (got %d)", s, len(ms))
}

func ar2AssertBadSuggestion(t *testing.T, lt *languagetool.JLanguageTool, rule *AgreementRule2, s, suggestion string) {
	t.Helper()
	ms := ar2Match(lt, rule, s)
	require.Equal(t, 1, len(ms), "Did not find one match in sentence %q (got %d)", s, len(ms))
	require.Contains(t, ms[0].GetSuggestedReplacements(), suggestion,
		"Got suggestions: %v, expected contains: %q", ms[0].GetSuggestedReplacements(), suggestion)
}

// Twin of AgreementRule2Test.testRule
func TestAgreementRule2_Rule(t *testing.T) {
	lt, rule := ar2Setup(t)

	ar2AssertBad(t, lt, rule, "Kleiner Haus am Waldesrand")
	ar2AssertGood(t, lt, rule, "Kleines Haus am Waldesrand")
	ar2AssertBad(t, lt, rule, "\"Kleiner Haus am Waldesrand\"")
	ar2AssertGood(t, lt, rule, "\"Kleines Haus am Waldesrand\"")
	ar2AssertBad(t, lt, rule, "Wirtschaftlich Wachstum kommt ins Stocken")
	ar2AssertGood(t, lt, rule, "Wirtschaftliches Wachstum kommt ins Stocken")
	ar2AssertGood(t, lt, rule, "Unter Berücksichtigung des Übergangs")

	ar2AssertGood(t, lt, rule, "Wirklich Frieden herrscht aber noch nicht")
	ar2AssertBad(t, lt, rule, "Deutscher Taschenbuch")
	ar2AssertGood(t, lt, rule, "Deutscher Taschenbuch Verlag expandiert")

	ar2AssertGood(t, lt, rule, "Wohl Anfang 1725 begegnete Bach dem Dichter.")
	ar2AssertGood(t, lt, rule, "Weniger Personal wird im ganzen Land gebraucht.")
	ar2AssertGood(t, lt, rule, "National Board of Review") // eng
	ar2AssertGood(t, lt, rule, "International Management") // eng.
	ar2AssertGood(t, lt, rule, "Gemeinsam Sportler anfeuern.")
	ar2AssertGood(t, lt, rule, "Viel Spaß beim Arbeiten")
	ar2AssertGood(t, lt, rule, "Ganz Europa stand vor einer Neuordnung.")
	ar2AssertGood(t, lt, rule, "Gesetzlich Versicherte sind davon ausgenommen.")
	ar2AssertGood(t, lt, rule, "Ausreichend Bananen essen.")
	ar2AssertGood(t, lt, rule, "Nachhaltig Yoga praktizieren")
	ar2AssertGood(t, lt, rule, "Überraschend Besuch bekommt er dann von ihr.")
	ar2AssertGood(t, lt, rule, "Ruhig Schlafen & Zentral Wohnen")
	ar2AssertGood(t, lt, rule, "Voller Mitleid")
	ar2AssertGood(t, lt, rule, "Voll Mitleid")
	ar2AssertGood(t, lt, rule, "Einzig Fernschüsse brachten Erfolgsaussichten.")
	ar2AssertGood(t, lt, rule, "Gelangweilt Dinge sortieren hilft als Ablenkung.")
	ar2AssertGood(t, lt, rule, "Ganzjährig Garten pflegen")
	ar2AssertGood(t, lt, rule, "Herzlich Willkommen bei unseren günstigen Rezepten!")
	ar2AssertGood(t, lt, rule, "10-tägiges Rückgaberecht")
	ar2AssertGood(t, lt, rule, "Angeblich Schüsse vor Explosionen gefallen")
	ar2AssertGood(t, lt, rule, "Dickes Danke auch an Elena")
	ar2AssertGood(t, lt, rule, "Dickes Dankeschön auch an Elena")
	ar2AssertGood(t, lt, rule, "Echt Scheiße")
	ar2AssertGood(t, lt, rule, "Entsprechende Automaten werden heute nicht mehr gebaut")
	ar2AssertGood(t, lt, rule, "Existenziell Bedrohte kriegen einen Taschenrechner")
	ar2AssertGood(t, lt, rule, "Flächendeckend Tempo 30")
	ar2AssertGood(t, lt, rule, "Frei Klavier spielen lernen")
	ar2AssertGood(t, lt, rule, "Ganz Eilige können es schaffen")
	ar2AssertGood(t, lt, rule, "Gering Gebildete laufen Gefahr ...")
	ar2AssertGood(t, lt, rule, "Ganz Ohr ist man hier")
	ar2AssertGood(t, lt, rule, "Gleichzeitig Muskeln aufbauen und Fett verlieren")
	ar2AssertGood(t, lt, rule, "Klar Schiff, Erster Offizier!")
	ar2AssertGood(t, lt, rule, "Kostenlos Bewegung schnuppern")
	ar2AssertGood(t, lt, rule, "Prinzipiell Anrecht auf eine Vertretung")
	ar2AssertGood(t, lt, rule, "Regelrecht Modell gestanden haben Michel")
	ar2AssertGood(t, lt, rule, "Weitgehend Konsens, auch über ...")
	ar2AssertGood(t, lt, rule, "Alarmierte Polizeibeamte nahmen den Mann fest.")
	ar2AssertGood(t, lt, rule, "Anderen Brot und Arbeit ermöglichen - das ist ihr Ziel")
	ar2AssertGood(t, lt, rule, "Diverse Unwesen, mit denen sich Hellboy beschäftigen muss, ...")
	ar2AssertGood(t, lt, rule, "Gut Qualifizierte bekommen Angebote")
	ar2AssertGood(t, lt, rule, "Liebe Mai, wie geht es dir?")
	ar2AssertGood(t, lt, rule, "Willkommen Simpsons-Fan!")
}

// Twin of AgreementRule2Test.testSuggestion
func TestAgreementRule2_Suggestion(t *testing.T) {
	lt, rule := ar2Setup(t)
	// commented in Java (false alarms with SOL categories path) — not asserted

	ar2AssertGood(t, lt, rule, "Kleinem Haus am Waldesrand ...")
	ar2AssertBadSuggestion(t, lt, rule, "Kleiner Haus am Waldesrand", "Kleines Haus")
	ar2AssertBadSuggestion(t, lt, rule, "Kleines Häuser am Waldesrand", "Kleine Häuser")
	ar2AssertBadSuggestion(t, lt, rule, "Kleinem Häuser am Waldesrand", "Kleine Häuser")
	// assertBad("Kleinen Häuser...") commented in Java

	ar2AssertBadSuggestion(t, lt, rule, "Kleines Tisch reicht auch", "Kleiner Tisch")
	ar2AssertGood(t, lt, rule, "Junger Frau geht das Geld aus")
	ar2AssertGood(t, lt, rule, "Junge Frau gewinnt im Lotto")
	ar2AssertBadSuggestion(t, lt, rule, "Junges Frau gewinnt im Lotto", "Junge Frau")
	// assertBad("Jungen Frau...") commented in Java
	ar2AssertBadSuggestion(t, lt, rule, "Jungem Frau gewinnt im Lotto", "Junge Frau")
	ar2AssertBadSuggestion(t, lt, rule, "Jung Frau gewinnt im Lotto", "Junge Frau")
	ar2AssertBadSuggestion(t, lt, rule, "Wirtschaftlich Wachstum kommt ins Stocken", "Wirtschaftliches Wachstum")
	ar2AssertBadSuggestion(t, lt, rule, "Wirtschaftlicher Wachstum kommt ins Stocken", "Wirtschaftliches Wachstum")
	// assertBad("Wirtschaftliche Wachstum...") commented in Java
}

// Twin metadata / anti-pattern surface (Java getAntiPatterns + estimateContext).
func TestAgreementRule2_MetaAndAntiPatterns(t *testing.T) {
	rule := NewAgreementRule2(nil)
	require.Equal(t, "DE_AGREEMENT2", rule.GetID())
	require.Equal(t, len(AgreementRule2AntiPatterns), len(rule.GetAntiPatterns()))
	max := 0
	for _, ap := range AgreementRule2AntiPatterns {
		if n := len(ap); n > max {
			max = n
		}
	}
	require.Equal(t, max, rule.EstimateContextForSureMatch())
	ex := rule.GetIncorrectExamples()
	require.NotEmpty(t, ex)
	require.Contains(t, ex[0].GetExample(), "Kleiner Haus")
}
