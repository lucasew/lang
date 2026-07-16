package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/GermanCompoundRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

type compoundMatcher interface {
	Match(*languagetool.AnalyzedSentence) []*rules.RuleMatch
}

func TestGermanCompoundRule_Rule(t *testing.T) {
	runDECompoundTests(t, NewGermanCompoundRule(nil))
	runDECompoundTests(t, NewSwissCompoundRule(nil))
}

func runDECompoundTests(t *testing.T, rule compoundMatcher) {
	t.Helper()
	check := func(expectedErrors int, text string, expSuggestions ...string) {
		t.Helper()
		matches := rule.Match(languagetool.AnalyzePlain(text))
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
	check(1, "Und herum zu knobeln können.", "herumzuknobeln")
	check(1, "Castrop Rauxel", "Castrop-Rauxel")
	check(1, "Spin off")
	check(1, "Das ist Haar sträubend", "Haarsträubend")
	check(1, "Spin off", "Spin-off")
	check(1, "Geräte Wahl", "Geräte-Wahl", "Gerätewahl")
	check(0, "x-mal")
	check(1, "x mal", "x-mal")
	check(0, "y-Achse")
	check(1, "y Achse", "y-Achse")
	check(0, "Blu-ray-Brenner")
	check(1, "Blu ray Brenner", "Blu-ray-Brenner")
	check(0, "Ich muss nachdenken")
	check(1, "Ich muss  nach denken", "nachdenken")
	check(0, "Backup")
	check(0, "Back-up")
	check(1, "Back up", "Back-up", "Backup")
	check(0, "Aggregatzustand")
	check(1, "Aggregat-Zustand", "Aggregatzustand")
	check(1, "Aggregat Zustand", "Aggregatzustand")
	check(1, "Billard Kugel", "Billardkugel")
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
