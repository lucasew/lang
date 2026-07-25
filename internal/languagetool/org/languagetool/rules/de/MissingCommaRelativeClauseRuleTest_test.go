package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/MissingCommaRelativeClauseRuleTest.java
//
// Java:
//   JLanguageTool lt = new JLanguageTool(Languages.getLanguageForShortCode("de-DE"));
//   MissingCommaRelativeClauseRule rule = new MissingCommaRelativeClauseRule(TestTools.getMessages("de"));
//   assertMatch / assertNoMatch via rule.match(lt.getAnalyzedSentence(input))
//   then rule = new MissingCommaRelativeClauseRule(messages, true) for behind constructor
//
// Go: de-DE analysis via GermanWordTokenizer + GermanTagger.Tag (sentence-start case)
// + GermanRuleDisambiguator + post chunker — same shape as Java getAnalyzedSentence.
// WireGermanTagWord alone is Lookup-only and leaves sentence-start "Alles" untagged;
// that would not match Java createDefaultTagger().tag. No hand-injected POS tables.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// newMissingCommaRelativeFixture mirrors Java MissingCommaRelativeClauseRuleTest setup.
func newMissingCommaRelativeFixture(t *testing.T) *languagetool.JLanguageTool {
	t.Helper()
	// Java: Languages.getLanguageForShortCode("de-DE")
	lt := languagetool.NewJLanguageTool("de-DE")
	// createDefaultTagger + disambiguator + chunker + synth (anti-pattern MatchInflectedForms)
	WireGermanRuntimeResourcesFor(lt, "DE")
	require.NotNil(t, DiscoveredGermanTagger(), "GermanTagger.INSTANCE resources required")
	require.True(t, WireGermanTagWord(lt) || lt.TagWord != nil, "GermanTagger resources required")
	return lt
}

// mcrcAnalyzedSentence ports lt.getAnalyzedSentence(input):
// GermanTagger.Tag (ignoreCase / sentence-start lowercasing) + Disambiguator + post chunker.
func mcrcAnalyzedSentence(t *testing.T, lt *languagetool.JLanguageTool, input string) *languagetool.AnalyzedSentence {
	t.Helper()
	tagger := DiscoveredGermanTagger()
	require.NotNil(t, tagger)
	sent := analyzeWithGermanTagger(tagger, input)
	require.NotNil(t, sent, "analyzeWithGermanTagger(%q)", input)
	if lt != nil && lt.Disambiguator != nil {
		if d := lt.Disambiguator.Disambiguate(sent); d != nil {
			sent = d
		}
	}
	if sent != nil && lt != nil && lt.PostDisambiguationChunker != nil {
		lt.PostDisambiguationChunker.AddChunkTags(sent.GetTokens())
	}
	return sent
}

// assertMCRCMatch ports MissingCommaRelativeClauseRuleTest.assertMatch.
func assertMCRCMatch(t *testing.T, input string, from, to int, rule *MissingCommaRelativeClauseRule, lt *languagetool.JLanguageTool) {
	t.Helper()
	matches := rule.Match(mcrcAnalyzedSentence(t, lt, input))
	require.Equal(t, 1, len(matches), "match count for %q", input)
	require.Equal(t, from, matches[0].GetFromPos(), "fromPos for %q", input)
	require.Equal(t, to, matches[0].GetToPos(), "toPos for %q", input)
}

// assertMCRCNoMatch ports MissingCommaRelativeClauseRuleTest.assertNoMatch.
func assertMCRCNoMatch(t *testing.T, input string, rule *MissingCommaRelativeClauseRule, lt *languagetool.JLanguageTool) {
	t.Helper()
	matches := rule.Match(mcrcAnalyzedSentence(t, lt, input))
	require.Equal(t, 0, len(matches), "expected 0 matches for %q, got %d", input, len(matches))
}

// TestMissingCommaRelativeClauseRule_Match ports MissingCommaRelativeClauseRuleTest.testMatch.
func TestMissingCommaRelativeClauseRule_Match(t *testing.T) {
	lt := newMissingCommaRelativeFixture(t)

	// --- front constructor (behind=false) ---
	rule := NewMissingCommaRelativeClauseRule(nil)
	require.Equal(t, "COMMA_IN_FRONT_RELATIVE_CLAUSE", rule.GetID())
	require.Equal(t, "Fehlendes Komma vor Relativsatz", rule.GetDescription())
	require.NotNil(t, rule.GetCategory())
	require.True(t, rule.GetCategory().GetID().Equals(rules.NewCategoryId("HILFESTELLUNG_KOMMASETZUNG")))
	require.Equal(t, 25, len(rule.GetAntiPatterns()))

	assertMCRCMatch(t, "Das Auto das am Straßenrand steht parkt im Halteverbot.", 4, 12, rule, lt)
	assertMCRCMatch(t, "Das Auto das am Straßenrand steht, parkt im Halteverbot.", 4, 12, rule, lt)
	assertMCRCMatch(t, "Das Auto in dem der Mann sitzt, parkt im Halteverbot.", 4, 15, rule, lt)
	assertMCRCMatch(t, "Das Auto in dem der Mann sitzt parkt im Halteverbot.", 4, 15, rule, lt)
	assertMCRCMatch(t, "Die Frau die vor dem Auto steht hat schwarze Haare.", 4, 12, rule, lt)
	assertMCRCMatch(t, "Die Frau die vor dem Auto steht, hat schwarze Haare.", 4, 12, rule, lt)
	assertMCRCMatch(t, "Alles was ich habe, ist ein Buch.", 0, 9, rule, lt)

	assertMCRCNoMatch(t, "Computer machen die Leute dumm.", rule, lt)
	assertMCRCNoMatch(t, "Die Unstimmigkeit zwischen den Geschichten der zwei Unfallbeteiligten war groß.", rule, lt)
	assertMCRCNoMatch(t, "Ebenso darf keine schwerere Strafe als die zum Zeitpunkt der Begehung der strafbaren Handlung angedrohte Strafe verhängt werden.", rule, lt)
	assertMCRCNoMatch(t, "Als dritte Gruppe lassen sich Aminosäuren fassen, die der Organismus anstelle dieser in Proteine einbaut.", rule, lt)
	assertMCRCNoMatch(t, "Selbst wenn das alles perfekt verlustfrei wäre, hätte ich nichts gewonnen.", rule, lt)
	assertMCRCNoMatch(t, "Die Studenten, deren Urteil am stärksten von dem der Profis abwich, waren sich sicher, einen guten von einem schlechten unterscheiden zu können.", rule, lt)
	assertMCRCNoMatch(t, "Die Studenten, deren Urteil am stärksten durch das der Profis beeinflusst wurde, waren sich sicher, einen guten von einem schlechten unterscheiden zu können.", rule, lt)

	// --- behind constructor (behind=true) ---
	rule = NewMissingCommaRelativeClauseRuleBehind(nil)
	require.Equal(t, "COMMA_BEHIND_RELATIVE_CLAUSE", rule.GetID())
	require.Equal(t, "Fehlendes Komma nach Relativsatz", rule.GetDescription())

	assertMCRCMatch(t, "Das Auto, das am Straßenrand steht parkt im Halteverbot.", 29, 40, rule, lt)
	assertMCRCMatch(t, "Das Auto, in dem der Mann sitzt parkt im Halteverbot.", 26, 37, rule, lt)
	assertMCRCMatch(t, "Die Frau, die vor dem Auto steht hat schwarze Haare.", 27, 36, rule, lt)
	assertMCRCMatch(t, "Alles, was ich habe ist ein Buch.", 15, 23, rule, lt)
	assertMCRCMatch(t, "In diesem Prozess sind aber Entwicklungsschritte ja integriert, die wir Psychiater glaube ich auch gut kennen.", 72, 93, rule, lt)
	assertMCRCMatch(t, "Alles, was du für die Spieße brauchst sind Tortellini, getrocknete Tomaten, Cherrytomaten und Mozzarellakugeln.", 29, 42, rule, lt)

	assertMCRCNoMatch(t, ".... das war eher so das, was alle im Browser deaktiviert hatten, weil man dachte, über Javascript wird irgendwie Schadsoftware eingeschleust.", rule, lt)
	assertMCRCNoMatch(t, "Ich habe einige Fehler begangen, die ich vermeiden hätte können sollen.", rule, lt)
	assertMCRCNoMatch(t, "Doch die Rolle, die Mohammed bin Salman in Riad spielt, ist zwiespältig.", rule, lt)
	assertMCRCNoMatch(t, "Laut Josef Peter Burg gehörte das Haus, aus dem er samt Familie geworfen wurde, Werner und Sigrid Bahlke, während er darin wohnte und darauf aufpasste.", rule, lt)
	assertMCRCNoMatch(t, "Darüber hinaus stellt die klassische Metaphysik eine Grundfrage, die sich etwa wie folgt formulieren lässt: Warum ist überhaupt Seiendes und nicht vielmehr Nichts?", rule, lt)
	assertMCRCNoMatch(t, "Wenn du alles, was du meinst nicht zu können, von anderen erledigen lässt, wirst du es niemals selbst lernen.", rule, lt)
	assertMCRCNoMatch(t, "Er hat einen Zeitraum durchlebt, in dem seine Gedanken verträumt auf den weiten Feldern der Mysterien umherirrten.", rule, lt)
	assertMCRCNoMatch(t, "Es ist die Wiederkehr der Panikmache, die der neue Nationalismus mit dem der Sprachreiniger verbindet und die Geschichte der Sprachreinigung zu einem Lehrstück macht.", rule, lt)
	assertMCRCNoMatch(t, "Gesuche können von Institutionen, Organisationen, Vereinen und Gruppierungen gestellt werden, die im Kanton Luzern domiziliert sind.", rule, lt)
	assertMCRCNoMatch(t, "Die Klausel kann zudem nur Gleichrang mit Verbindlichkeiten des Schuldners herstellen, die vom Gesetz in der Insolvenz nicht privilegiert sind, so dass die Klausel nichts an der gesetzlich vorgesehenen Rangfolge im Insolvenzverfahren ändert.", rule, lt)
	assertMCRCNoMatch(t, "Plan von Maßnahmen, mit denen das Ansteckungsrisiko während des Aufenthalts an einem Ort verringert werden soll", rule, lt)
	assertMCRCNoMatch(t, "Aus diesem Grund sind die Wörter nicht direkt übersetzt, stattdessen wird der Zustand oder die Situation beschrieben in der die Wörter benutzt werden.", rule, lt)
	assertMCRCNoMatch(t, "Kryptographisch sichere Verfahren sind dann solche, für die es keine bessere Methode zum Brechen der Sicherheit als das Faktorisieren einer großen Zahl gibt, insbesondere kann der private nicht aus dem öffentlichen Schlüssel errechnet werden.", rule, lt)
	assertMCRCNoMatch(t, "Wayne Jancik begrenzt seine One-Hit-Wonders in den USA auf die Billboard-Top-40-Pop-Hitparade mit einer „Ruhe-Periode“ von 5 Jahren, innerhalb derer kein weiterer Hit desselben Interpreten in die Top 40 gelangen darf.", rule, lt)
	assertMCRCNoMatch(t, "Nach dem Abzug russischer Truppen aus der Region Kiew sind in den ehemals besetzten und umkämpften ukrainischen Gebieten inzwischen Hunderte Leichen von Bewohnern gefunden worden.", rule, lt)
	assertMCRCNoMatch(t, "Bei der Installation werden Ihnen Sticker und Aufkleber zur Verfügung gestellt, die Sie anstelle der traditionellen Emojis verwenden können.", rule, lt)
	assertMCRCNoMatch(t, "»Alles, was wir dank dieses Projektes sehen werden, wird für uns neu sein«, so der renommierte Bienenforscher.", rule, lt)
}

// TestMissingCommaMarksRE_JavaRangeQuirk twins Java MARKS_REGEX character-class range
// quirk (!-– includes letters/digits as single-char separators when .matches()).
func TestMissingCommaMarksRE_JavaRangeQuirk(t *testing.T) {
	require.True(t, missingCommaMarksRE.MatchString(","))
	require.True(t, missingCommaMarksRE.MatchString("."))
	// Range U+0021..U+2013 includes ASCII letters/digits (Java quirk)
	require.True(t, missingCommaMarksRE.MatchString("a"))
	require.True(t, missingCommaMarksRE.MatchString("5"))
	require.True(t, missingCommaMarksRE.MatchString("—")) // em-dash after range
	require.True(t, missingCommaMarksRE.MatchString("•"))
	// Multi-char never matches the marks class itself
	require.False(t, missingCommaMarksRE.MatchString("Auto"))
	require.False(t, missingCommaMarksRE.MatchString("und"))
	// isSeparator still treats und/oder as separators via equals
	require.True(t, isSeparatorMissingComma("und"))
	require.True(t, isSeparatorMissingComma("oder"))
	require.False(t, isSeparatorMissingComma("Auto"))
}

// TestMissingCommaRelativeClauseRule_Constructors ports both Java constructors.
func TestMissingCommaRelativeClauseRule_Constructors(t *testing.T) {
	front := NewMissingCommaRelativeClauseRule(nil)
	require.False(t, front.Behind)
	require.Equal(t, "COMMA_IN_FRONT_RELATIVE_CLAUSE", front.GetID())

	behind := NewMissingCommaRelativeClauseRuleBehind(nil)
	require.True(t, behind.Behind)
	require.Equal(t, "COMMA_BEHIND_RELATIVE_CLAUSE", behind.GetID())

	flag := NewMissingCommaRelativeClauseRuleBehindFlag(nil, true)
	require.True(t, flag.Behind)
	require.Equal(t, "COMMA_BEHIND_RELATIVE_CLAUSE", flag.GetID())
	flag2 := NewMissingCommaRelativeClauseRuleBehindFlag(nil, false)
	require.False(t, flag2.Behind)
	require.Equal(t, "COMMA_IN_FRONT_RELATIVE_CLAUSE", flag2.GetID())
}
