package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/GermanWordRepeatRuleTest.java
//
// Java setup:
//   Language german = Languages.getLanguageForShortCode("de");
//   GermanWordRepeatRule rule = new GermanWordRepeatRule(TestTools.getEnglishMessages(), german);
//   JLanguageTool lt = new JLanguageTool(german);
//   assertThat(rule.match(lt.getAnalyzedSentence(...)).length, is(N));
//
// Go: de JLanguageTool + WireGermanTagWord (createDefaultTagger) so POS/lemma
// anti-patterns and ignore() gates use real GermanTagger readings — no AnalyzePlain
// invent and no hand-injected POS for the Java matrix.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func germanWordRepeatMessages() map[string]string {
	// Mirror TestTools English ResourceBundle keys used by WordRepeatRule.
	return map[string]string{
		"repetition":            "Word repetition",
		"desc_repetition":       "Word repetition",
		"desc_repetition_short": "repetition",
		"category_misc":         "Miscellaneous",
		"category_redundancy":   "Redundant Phrases",
	}
}

// newGermanWordRepeatFixture mirrors Java GermanWordRepeatRuleTest fields.
func newGermanWordRepeatFixture(t *testing.T) (*GermanWordRepeatRule, *languagetool.JLanguageTool) {
	t.Helper()
	// Java: Languages.getLanguageForShortCode("de")
	lt := languagetool.NewJLanguageTool("de")
	require.True(t, WireGermanTagWord(lt), "GermanTagger resources required for GermanWordRepeatRule twin")
	// Java: new GermanWordRepeatRule(TestTools.getEnglishMessages(), german)
	rule := NewGermanWordRepeatRule(germanWordRepeatMessages())
	return rule, lt
}

// assertGWRMatches ports rule.match(lt.getAnalyzedSentence(s)).length == want
func assertGWRMatches(t *testing.T, rule *GermanWordRepeatRule, lt *languagetool.JLanguageTool, s string, want int) {
	t.Helper()
	sent := lt.GetAnalyzedSentence(s)
	require.NotNil(t, sent, "GetAnalyzedSentence(%q)", s)
	got := len(rule.Match(sent))
	require.Equal(t, want, got, "sentence %q", s)
}

// TestGermanWordRepeatRule_Rule ports GermanWordRepeatRuleTest.testRule full matrix.
func TestGermanWordRepeatRule_Rule(t *testing.T) {
	rule, lt := newGermanWordRepeatFixture(t)

	// Order and strings match Java GermanWordRepeatRuleTest.testRule.
	assertGWRMatches(t, rule, lt, "Das ist gut so.", 0)
	assertGWRMatches(t, rule, lt, "Das ist ist gut so.", 1)
	assertGWRMatches(t, rule, lt, "Der der Mann", 1)
	assertGWRMatches(t, rule, lt, "Warum fragen Sie sie nicht selbst?", 0)
	assertGWRMatches(t, rule, lt, "Er will nur sein Leben leben.", 0)
	assertGWRMatches(t, rule, lt, "Wie bei Honda, die die Bezahlung erhöht haben.", 0)
	assertGWRMatches(t, rule, lt, "Dann warfen sie sie weg.", 0)
	assertGWRMatches(t, rule, lt, "Dann konnte sie sie sehen.", 0)
	assertGWRMatches(t, rule, lt, "Er muss sein Essen essen.", 0)
	assertGWRMatches(t, rule, lt, "Wahrscheinlich ist das das Problem.", 0)
	assertGWRMatches(t, rule, lt, "Dann wäre das das erste Wirtschaftsmagazin mit mehr als 10.000 Lesern.", 0)
	assertGWRMatches(t, rule, lt, "Für mich war das das Härteste.", 0)
	assertGWRMatches(t, rule, lt, "Ich weiß, wer wer ist!", 0)
	assertGWRMatches(t, rule, lt, "Ich kann das machen, falls das das Problem ist.", 0)
	assertGWRMatches(t, rule, lt, "Als ich das das erste Mal gehört habe …", 0)
	assertGWRMatches(t, rule, lt, "Hat sie sie", 1) // used to crash, issue #1010
	assertGWRMatches(t, rule, lt, "Hat hat", 1)
	assertGWRMatches(t, rule, lt, "hat hat", 1)
}

// Metadata from Java GermanWordRepeatRule ctor (id, category, example pair).
func TestGermanWordRepeatRule_Metadata(t *testing.T) {
	rule := NewGermanWordRepeatRule(germanWordRepeatMessages())
	require.Equal(t, "GERMAN_WORD_REPEAT_RULE", rule.GetID())
	// Java: super.setCategory(Categories.REDUNDANCY)
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "REDUNDANCY", rule.GetCategory().GetID().String())
	// Java: ist ist → ist example pair
	inc := rule.GetIncorrectExamples()
	require.NotEmpty(t, inc)
	require.Contains(t, inc[0].GetExample(), "ist ist")
	cor := rule.GetCorrectExamples()
	require.NotEmpty(t, cor)
}
