package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/MissingVerbRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestMissingVerbRule_Test(t *testing.T) {
	rule := NewMissingVerbRule(nil)
	// Java isRealSentence requires PKT; untagged AnalyzePlain has no PKT → not a "real sentence".
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Da ist ein Verb, mal so zum testen."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Dieser Satz kein Verb."))))
}

func TestMissingVerbRule_MorphMissing(t *testing.T) {
	// All content tokens capitalized and tagged non-VER → missing verb.
	ss := languagetool.SentenceStartTagName
	pkt := "PKT"
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		atrWithPOS("In", "PRP:DAT:SIN", "in"),
		atrWithPOS("Diesem", "PRO:DEM:DAT:SIN:NEU", "dieser"),
		atrWithPOS("Satz", "SUB:DAT:SIN:MAS", "Satz"),
		atrWithPOS("Kein", "PIAT:NOM:SIN:NEU", "kein"), // capitalized non-VER
		atrWithPOS("Wort", "SUB:NOM:SIN:NEU", "Wort"),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken(".", &pkt, nil), 30),
	}
	// mark last as sentence end + PKT
	toks[len(toks)-1].SetSentEnd()
	// Need ≥5 non-ws tokens including SENT_START? Java MIN=5 on tokensWithoutWhitespace
	// tokens: START, In, Diesem, Satz, Kein, Wort, . = 7
	// Capitalized tagged non-VER — Kein is capitalized so !(!tagged && !cap) fails for all
	// Wait: "In" is capitalized and tagged PRP - not VER
	// All fail VER check; Kein is capitalized tagged - not (!tagged && !cap)
	// i==1 is "In" - verbAtSentenceStart without hook false
	sent := languagetool.NewAnalyzedSentence(toks)
	rule := NewMissingVerbRule(nil)
	ms := rule.Match(sent)
	require.NotEmpty(t, ms)
	// Java: message only, no invent shortMessage.
	require.Equal(t, "Dieser Satz scheint kein Verb zu enthalten", ms[0].GetMessage())
	require.Empty(t, ms[0].GetShortMessage())
	require.Equal(t, "Satz ohne Verb", rule.GetDescription())
}

func TestMissingVerbRule_MorphWithVerb(t *testing.T) {
	ss := languagetool.SentenceStartTagName
	pkt := "PKT"
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		atrWithPOS("Da", "ADV", "da"),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("ein", "ART:IND:NOM:SIN:NEU", "ein"),
		atrWithPOS("Verb", "SUB:NOM:SIN:NEU", "Verb"),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken(".", &pkt, nil), 20),
	}
	toks[len(toks)-1].SetSentEnd()
	sent := languagetool.NewAnalyzedSentence(toks)
	require.Empty(t, NewMissingVerbRule(nil).Match(sent))
}

func TestMissingVerbRule_SpecialCaseVielenDank(t *testing.T) {
	// Even without verb tags, special case
	require.Empty(t, NewMissingVerbRule(nil).Match(languagetool.AnalyzePlain("Vielen Dank.")))
}

func TestMissingVerbRule_ShortSentence(t *testing.T) {
	// fewer than MIN_TOKENS_FOR_ERROR
	ss := languagetool.SentenceStartTagName
	pkt := "PKT"
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		atrWithPOS("Hallo", "ITJ", "hallo"),
		atrWithPOS("Welt", "SUB:NOM:SIN:FEM", "Welt"),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken(".", &pkt, nil), 10),
	}
	toks[len(toks)-1].SetSentEnd()
	// 4 tokens < 5 → no error
	require.Empty(t, NewMissingVerbRule(nil).Match(languagetool.NewAnalyzedSentence(toks)))
}
