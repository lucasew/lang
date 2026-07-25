package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/RedundantModalOrAuxiliaryVerbTest.java
//
// Java setup:
//   Language lang = Languages.getLanguageForShortCode("de-DE");
//   JLanguageTool lt = new JLanguageTool(lang);
//   setUpRule(lt): disable all; add RedundantModalOrAuxiliaryVerb; enable
//   assertEquals(1|0, lt.check(sentence).size());
//
// Go: de-DE JLanguageTool + WireGermanTagWord (createDefaultTagger) so VER:MOD /
// VER:AUX / PRO:* / PA2 tags come from GermanTagger — no AnalyzePlain invent and no
// hand-injected POS for the Java testRule matrix. Primary twin is rule.Match on
// lt.GetAnalyzedSentence (same as Java check with only this rule enabled).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// setUpRedundantModalRule ports RedundantModalOrAuxiliaryVerbTest.setUpRule.
func setUpRedundantModalRule(t *testing.T) (*RedundantModalOrAuxiliaryVerb, *languagetool.JLanguageTool) {
	t.Helper()
	// Java: Languages.getLanguageForShortCode("de-DE"); new JLanguageTool(lang)
	lt := languagetool.NewJLanguageTool("de-DE")
	require.True(t, WireGermanTagWord(lt), "GermanTagger resources required for RedundantModalOrAuxiliaryVerb twin")
	// Java: for (Rule rule : lt.getAllRules()) { lt.disableRule(rule.getId()); }
	for _, id := range lt.GetAllRegisteredRuleIDs() {
		lt.DisableRule(id)
	}
	// Java: new RedundantModalOrAuxiliaryVerb(TestTools.getMessages(lang.getShortCode()))
	rule := NewRedundantModalOrAuxiliaryVerb(nil)
	// Only this rule: Match on GetAnalyzedSentence is the check path for a single
	// sentence-level rule (Java lt.check size with all others disabled).
	return rule, lt
}

// assertRedundantModalCount ports assertEquals(want, lt.check(text).size()).
func assertRedundantModalCount(t *testing.T, rule *RedundantModalOrAuxiliaryVerb, lt *languagetool.JLanguageTool, text string, want int) {
	t.Helper()
	sent := lt.GetAnalyzedSentence(text)
	require.NotNil(t, sent, "GetAnalyzedSentence(%q)", text)
	got := len(rule.Match(sent))
	require.Equal(t, want, got, "sentence %q", text)
}

// TestRedundantModalOrAuxiliaryVerb_Rule ports RedundantModalOrAuxiliaryVerbTest.testRule.
// Full Java matrix: 10 match + 20 no-match sentences, exact strings and expected sizes.
func TestRedundantModalOrAuxiliaryVerb_Rule(t *testing.T) {
	rule, lt := setUpRedundantModalRule(t)

	// matches (Java assertEquals(1, lt.check(...).size()))
	assertRedundantModalCount(t, rule, lt, "Erst werde ich die Preise vergleichen und erst dann werde ich entscheiden, ob ich die Kamera kaufe.", 1)
	assertRedundantModalCount(t, rule, lt, "Sie hat das Foto von mir als kleinem Jungen angeschaut und hat gelacht.", 1)
	assertRedundantModalCount(t, rule, lt, "Ich bin gern in eurer Mitte und ich bin gern zu Gast bei euch.", 1)
	assertRedundantModalCount(t, rule, lt, "Da ich nun einmal bin, wer ich bin und was ich bin, kann ich es nicht übers Herz bringen.", 1)
	assertRedundantModalCount(t, rule, lt, "Wann ist jemand kühn und wann ist jemand tollkühn?", 1)
	assertRedundantModalCount(t, rule, lt, "Tom ist um halb drei angekommen und Mary ist kurze Zeit später angekommen.", 1)
	assertRedundantModalCount(t, rule, lt, "Das Essen ist gut und der Service hier ist gut.", 1)
	assertRedundantModalCount(t, rule, lt, "Wir müssen wissen, was wir tun sollen und wie wir es tun sollen.", 1)
	assertRedundantModalCount(t, rule, lt, "Ich muss unbedingt wissen, was zu tun ist, wie es zu tun ist und wann es zu tun ist.", 1)
	assertRedundantModalCount(t, rule, lt, "Er erzählte ihr, eine böse Hexe habe ihn in einen Frosch verwandelt und nur sie allein habe ihn retten können und morgen würden sie gemeinsam in sein Reich fahren.", 1)

	// no match (Java assertEquals(0, lt.check(...).size()))
	assertRedundantModalCount(t, rule, lt, "„nicht dürfen“ und „nicht müssen“ dürfen nicht miteinander verwechselt werden.", 0)
	assertRedundantModalCount(t, rule, lt, "Ich mag Physik und Mathematik mag ich noch mehr.", 0)
	assertRedundantModalCount(t, rule, lt, "Aus Angst vor den Zeitungen sind Politiker langweilig und am Ende sind sie selbst für die Zeitungen zu langweilig.", 0)
	assertRedundantModalCount(t, rule, lt, "Ändert der Akkusativ wirklich die Bedeutung eines Satzes: Sie hat Aids oder Aids hat sie?", 0)
	assertRedundantModalCount(t, rule, lt, "Tom hat eine Tüte Äpfel gekauft und er hat an einem Tag ein Drittel davon gegessen.", 0)
	assertRedundantModalCount(t, rule, lt, "Mein Vaterland ist mein Stadtviertel und nun ist es beinahe nicht wiedererkennbar.", 0)
	assertRedundantModalCount(t, rule, lt, "Johnny hat Alice vorgeschlagen und sie hat akzeptiert.", 0)
	assertRedundantModalCount(t, rule, lt, "Unsere Zeit hier ist begrenzt und sie ist wertvoll.", 0)
	assertRedundantModalCount(t, rule, lt, "Manchmal ist die Wahrheit nützlich und manchmal ist sie unnütz.", 0)
	assertRedundantModalCount(t, rule, lt, "Treue ist irgendwo absolut oder sie ist gar nichts.", 0)
	assertRedundantModalCount(t, rule, lt, "Sie mag niemanden und niemand mag sie.", 0)
	assertRedundantModalCount(t, rule, lt, "Eines Tages wird die USA eine Frau als Präsidenten wählen und das wird kein schönes Schauspiel sein.", 0)
	assertRedundantModalCount(t, rule, lt, "Mein Name ist Mary und das ist Tom.", 0)
	assertRedundantModalCount(t, rule, lt, "Er hat seine Augen langsam geöffnet und dann hat sie ihn geküsst.", 0)
	assertRedundantModalCount(t, rule, lt, "Die Flaneure haben es prinzipiell nicht eilig und sind immer für ein Schwätzchen zu haben.", 0)
	assertRedundantModalCount(t, rule, lt, "Es sind Tendenzen erkennbar und diese Tendenzen sind leider nicht ermutigend.", 0)
	assertRedundantModalCount(t, rule, lt, "Ja ist ja und nein ist nein, das nennt man Offenheit.", 0)
	assertRedundantModalCount(t, rule, lt, "Wie werden einmal unsere Namen hinter den Erfindern des Fliegens und dergleichen vergessen werden?", 0)
	assertRedundantModalCount(t, rule, lt, "In den Chroniken ist über die Flut von 1342 zu lesen, dass im Mainzer Dom einem Mann das Wasser bis zur Brust gestanden habe und dass man in Köln mit Booten über die Stadtmauern habe fahren können.", 0)
	assertRedundantModalCount(t, rule, lt, "Sie stellt sicher, dass die gerissenen Bänder nicht belastet werden können und dass das Gelenk dennoch bewegt werden kann.", 0)
}

// TestRedundantModalOrAuxiliaryVerb_Meta ports Java ctor metadata (id, description, STYLE, default-off).
func TestRedundantModalOrAuxiliaryVerb_Meta(t *testing.T) {
	r := NewRedundantModalOrAuxiliaryVerb(nil)
	require.Equal(t, "REDUNDANT_MODAL_VERB", r.GetID())
	require.Equal(t, "Redundantes Modal- oder Hilfsverb", r.GetDescription())
	require.True(t, r.IsDefaultOff())
	require.NotNil(t, r.GetCategory())
	require.Equal(t, "STYLE", r.GetCategory().GetID().String())
	require.Equal(t, "style", string(r.GetLocQualityIssueType()))
	// Untagged plain analysis must not invent VER:MOD/VER:AUX matches (Java needs tagger).
	require.Equal(t, 0, len(r.Match(languagetool.AnalyzePlain("Erst werde ich die Preise vergleichen und erst dann werde ich entscheiden."))))
}

// Twin of Java MARKS_REGEX character-class range quirk (!-– includes letters/digits).
func TestRedundantMarksRE_JavaRangeQuirk(t *testing.T) {
	require.True(t, redundantMarksRE.MatchString(","))
	require.True(t, redundantMarksRE.MatchString("."))
	require.True(t, redundantMarksRE.MatchString("!"))
	require.True(t, redundantMarksRE.MatchString("a"))
	require.True(t, redundantMarksRE.MatchString("5"))
	require.True(t, redundantMarksRE.MatchString("—")) // em-dash after range
	require.True(t, redundantMarksRE.MatchString("»"))
	require.True(t, redundantMarksRE.MatchString("("))
	// multi-char tokens never full-match (.matches())
	require.False(t, redundantMarksRE.MatchString("Auto"))
	require.False(t, redundantMarksRE.MatchString("und"))
	require.False(t, redundantMarksRE.MatchString("Tom"))
	// bullet/ellipsis not in this Java pattern (unlike MissingComma)
	require.False(t, redundantMarksRE.MatchString("•"))
	require.False(t, redundantMarksRE.MatchString("…"))
}
