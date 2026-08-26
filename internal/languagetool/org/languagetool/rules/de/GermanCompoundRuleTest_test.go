package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/GermanCompoundRuleTest.java
//
// Java:
//   lt = new JLanguageTool(Languages.getLanguageForShortCode("de-DE"));
//   rule = new GermanCompoundRule(...); runTests();
//   rule = new SwissCompoundRule(...); runTests();
//   AbstractCompoundRuleTest.check → rule.match(lt.getAnalyzedSentence(text))
//
// Go: de-DE JLanguageTool + WireGermanRuntimeResourcesFor (tagger + de_DE.dict for
// isMisspelled via GermanyGerman default spelling rule). Full active Java matrix
// (no AnalyzePlain invent for the match path).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

type compoundMatcher interface {
	Match(*languagetool.AnalyzedSentence) []*rules.RuleMatch
}

// newGermanCompoundFixture mirrors Java GermanCompoundRuleTest fields (lt + rule).
func newGermanCompoundFixture(t *testing.T) *languagetool.JLanguageTool {
	t.Helper()
	// Java: new JLanguageTool(Languages.getLanguageForShortCode("de-DE"))
	lt := languagetool.NewJLanguageTool("de-DE")
	// createDefaultTagger + getDefaultSpellingRule dict (isMisspelled filter)
	WireGermanRuntimeResourcesFor(lt, "DE")
	return lt
}

func TestGermanCompoundRule_Rule(t *testing.T) {
	lt := newGermanCompoundFixture(t)
	// Java: rule = new GermanCompoundRule(...); runTests();
	runDECompoundTests(t, lt, NewGermanCompoundRule(nil))
	// Java: rule = new SwissCompoundRule(...); runTests(); same matrix
	runDECompoundTests(t, lt, NewSwissCompoundRule(nil))
}

// runDECompoundTests ports GermanCompoundRuleTest.runTests() full active matrix.
func runDECompoundTests(t *testing.T, lt *languagetool.JLanguageTool, rule compoundMatcher) {
	t.Helper()
	// Ports AbstractCompoundRuleTest.check(expectedErrors, text, expSuggestions...)
	check := func(expectedErrors int, text string, expSuggestions ...string) {
		t.Helper()
		// Java: rule.match(lt.getAnalyzedSentence(text))
		sent := lt.GetAnalyzedSentence(text)
		require.NotNil(t, sent, "GetAnalyzedSentence(%q)", text)
		matches := rule.Match(sent)
		require.Equal(t, expectedErrors, len(matches), "text %q got %v", text, formatDEMatches(matches))
		if len(expSuggestions) > 0 {
			require.Equal(t, 1, expectedErrors)
			require.Equal(t, expSuggestions, matches[0].GetSuggestedReplacements(), "text %q", text)
		}
	}

	// correct sentences:
	check(0, "Eine tolle CD-ROM")
	check(0, "Eine tolle CD-ROM.")
	check(0, "Ein toller CD-ROM-Test.")
	check(0, "Systemadministrator")
	check(0, "Eine Million Dollar")
	check(0, "Das System des Administrators")
	check(0, "Nur im Stand-by-Betrieb")
	check(0, "Start, Ziel, Sieg")
	check(0, "Roll-on-roll-off-Schiff")
	check(0, "Halswirbelsäule")
	check(0, "Castrop-Rauxel")
	check(0, "2-Zimmer-Wohnung")
	check(0, "3-Zimmer-Wohnung")
	check(0, "Hals-Wirbel-Säule")
	check(0, "Die Bürger konnten an die 900 Meter Kabel in Eigenregie verlegen.")
	check(0, "Die Bürger konnten ca. 900 Meter Kabel in Eigenregie verlegen.")
	check(0, "Aus dem Tank zapften die Diebe rund 250 Liter Diesel ab.")
	check(0, "Aus dem Tank zapften die Diebe 250 Liter Diesel ab.")
	check(0, "Lohnt sich die Werbung vom ausgegebenen Euro aus gedacht?")

	// incorrect sentences:
	check(1, "System Administrator", "Systemadministrator")
	check(1, "System-Administrator")
	check(1, "bla bla bla bla bla System Administrator bla bla bla bla bla")
	check(1, "System Administrator blubb")
	check(1, "Der System Administrator")
	check(1, "Der dumme System Administrator")
	check(1, "CD ROM", "CD-ROM")
	check(1, "Nur im Stand by Betrieb", "Stand-by-Betrieb")
	check(1, "Ein echter Start Ziel Sieg", "Start-Ziel-Sieg")
	check(1, "Ein echter Start Ziel Sieg.")
	check(1, "Ein Start Ziel Sieg")
	check(1, "Start Ziel Sieg")
	check(1, "Start Ziel Sieg!")
	check(1, "Doppler Effekt")
	check(1, "3 Tage Woche")
	check(1, "4 Tage Woche")
	check(1, "5 Tage Woche")
	check(1, "100 m Lauf")
	check(1, "200 m Lauf")
	check(1, "800 m Lauf")
	check(1, "1000 m Lauf")
	check(1, "2 Zimmer Wohnung")
	check(1, "3 Zimmer-Wohnung")
	check(1, "4-Zimmer Wohnung")
	check(1, "5 Zimmer Wohnung")
	check(1, "6 Zimmer Wohnung")
	check(1, "1000 Jahr Feier")
	check(1, "1000-Jahr Feier")
	check(1, "1000 Jahr-Feier")
	check(2, "Der dumme System Administrator legt die CD ROM")
	check(2, "Der dumme System Administrator legt die CD ROM.")
	check(2, "Der dumme System Administrator legt die CD ROM ein blah")
	check(2, "System Administrator CD ROM")
	// check(2, "Hals Wirbel Säule"); // Java commented
	check(1, "Und herum zu knobeln können.", "herumzuknobeln")
	check(1, "Castrop Rauxel", "Castrop-Rauxel")
	// check(1, "Roll on roll off Schiff", …); // Java commented — suggestion not on dictionary
	check(1, "Spin off")
	// no hyphen suggestion for some words:
	check(1, "Das ist Haar sträubend", "Haarsträubend")
	// Only hyphen suggestion for some words:
	// check(1, "Reality TV", "Reality-TV"); // Java commented
	check(1, "Spin off", "Spin-off")
	// also accept incorrect upper/lowercase spelling:
	// check(1, "Spin Off", …); // Java commented
	// check(1, "CW Wert", …); // Java commented
	// also detect an error if only some of the hyphens are missing:
	// check(1, "Roll-on-roll-off Schiff", …); // Java commented
	check(1, "Geräte Wahl", "Geräte-Wahl", "Gerätewahl")
	// first part is a single character:
	check(0, "x-mal")
	check(1, "x mal", "x-mal")
	check(0, "y-Achse")
	check(1, "y Achse", "y-Achse")

	// "Blu-ray-Brenner*"
	check(0, "Blu-ray-Brenner")
	check(1, "Blu ray Brenner", "Blu-ray-Brenner")
	// "nach-denken+"
	check(0, "Ich muss nachdenken")
	check(1, "Ich muss  nach denken", "nachdenken")
	// "Back-up$"
	check(0, "Backup")
	check(0, "Back-up")
	check(1, "Back up", "Back-up", "Backup")
	// "Aggregat-Zustand?"
	check(0, "Aggregatzustand")
	// check(0, "Aggregat-Zustand"); // Java commented
	check(1, "Aggregat-Zustand", "Aggregatzustand")
	check(1, "Aggregat Zustand", "Aggregatzustand")
	check(1, "Billard Kugel", "Billardkugel")
}

func TestGermanCompoundRule_AntiPatternsCount(t *testing.T) {
	// Java GermanCompoundRule.ANTI_PATTERNS has 8 entries.
	require.Equal(t, 8, len(GermanCompoundRuleAntiPatterns), "Java ANTI_PATTERNS 8/8")
	require.Equal(t, 8, len(deCompoundAntiPatterns()), "IMMUNIZE rules 8/8")
	require.Equal(t, 8, len(NewGermanCompoundRule(nil).GetAntiPatterns()))
	require.Equal(t, 8, len(NewSwissCompoundRule(nil).GetAntiPatterns()))
}

func TestGermanCompoundRule_AntiPatternImmunizeSurface(t *testing.T) {
	lt := newGermanCompoundFixture(t)
	// Java anti-pattern: an|um + die + \d+
	sent := lt.GetAnalyzedSentence("Die Bürger konnten an die 900 Meter Kabel in Eigenregie verlegen.")
	require.NotNil(t, sent)
	imm := getSentenceWithDECompoundImmunization(sent)
	anyImm := false
	for _, tok := range imm.GetTokensWithoutWhitespace() {
		if tok != nil && tok.IsImmunized() {
			anyImm = true
			break
		}
	}
	require.True(t, anyImm, "expected an die 900 anti-pattern to immunize")
	// Original sentence must not be mutated (Java clones).
	for _, tok := range sent.GetTokensWithoutWhitespace() {
		require.False(t, tok != nil && tok.IsImmunized(), "original must not be mutated")
	}
}

func TestGermanCompoundRule_Metadata(t *testing.T) {
	// Java getId / getDescription / example pair
	r := NewGermanCompoundRule(nil)
	require.Equal(t, "DE_COMPOUNDS", r.GetID())
	require.Equal(t, "Zusammenschreibung von Wörtern, z. B. 'CD-ROM' statt 'CD ROM'", r.GetDescription())
	require.Equal(t, "DE_CH_COMPOUNDS", NewSwissCompoundRule(nil).GetID())
}

func formatDEMatches(matches []*rules.RuleMatch) string {
	s := ""
	for i, m := range matches {
		if i > 0 {
			s += "; "
		}
		reps := m.GetSuggestedReplacements()
		if len(reps) == 0 {
			s += "?"
		} else {
			s += reps[0]
		}
	}
	return s
}
