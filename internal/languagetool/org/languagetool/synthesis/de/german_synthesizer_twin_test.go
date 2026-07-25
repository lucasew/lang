package de

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	"github.com/stretchr/testify/require"
)

// Twin of org.languagetool.synthesis.GermanSynthesizerTest using official
// GermanSynthesizer.INSTANCE resources (german_synth.dict, german_tags.txt, de.sor).

func requireGermanInstance(t *testing.T) *GermanSynthesizer {
	t.Helper()
	s := Instance()
	if s == nil {
		t.Skip("german_synth.dict / official DE resources not available")
	}
	require.NotNil(t, s.Lookup)
	require.NotEmpty(t, s.PossibleTags)
	return s
}

func dummyToken(tokenStr string) *languagetool.AnalyzedToken {
	// Java: new AnalyzedToken(tokenStr, tokenStr, tokenStr)
	tag := tokenStr
	lem := tokenStr
	return languagetool.NewAnalyzedToken(tokenStr, &tag, &lem)
}

// arraysToString ports Java Arrays.toString(String[]) for twin assertions.
func arraysToString(forms []string) string {
	if forms == nil {
		return "[]"
	}
	return "[" + strings.Join(forms, ", ") + "]"
}

func synth(t *testing.T, s *GermanSynthesizer, word, posTag string) string {
	t.Helper()
	forms, err := s.Synthesize(dummyToken(word), posTag)
	require.NoError(t, err)
	return arraysToString(forms)
}

func synthRE(t *testing.T, s *GermanSynthesizer, word, posTag string) string {
	t.Helper()
	forms, err := s.SynthesizeRE(dummyToken(word), posTag, true)
	require.NoError(t, err)
	return arraysToString(forms)
}

// Port of GermanSynthesizerTest.testSynthesize
func TestGermanSynthesizer_Synthesize(t *testing.T) {
	s := requireGermanInstance(t)

	require.Equal(t, "[einhundertdreiundzwanzig]", synth(t, s, "123", "_spell_number_"))
	require.Equal(t, "[Zug]", synth(t, s, "Zug", "SUB:DAT:SIN:MAS"))
	require.Equal(t, "[Tisch]", synth(t, s, "Tisch", "SUB:DAT:SIN:MAS"))
	require.Equal(t, "[Buschfeuers]", synth(t, s, "Buschfeuer", "SUB:GEN:SIN:NEU"))
	require.Equal(t, "[Äußerungen]", synth(t, s, "Äußerung", "SUB:NOM:PLU:FEM"))
	require.Equal(t, "[]", synth(t, s, "Äußerung", "SUB:NOM:PLU:MAS"))
	require.Equal(t, "[Häuser]", synth(t, s, "Haus", "SUB:AKK:PLU:NEU"))
	require.Equal(t, "[Diabetes-Zentrums]", synthRE(t, s, "Diabetes-Zentrum", "SUB:GEN:SIN.*"))
	require.Equal(t, "[Diabetes-Zentren]", synthRE(t, s, "Diabetes-Zentrum", "SUB:GEN:PLU.*"))
	require.Equal(t, "[Xxxxxx-Zentren]", synthRE(t, s, "Xxxxxx-Zentrum", "SUB:GEN:PLU.*"))
	require.Equal(t, "[Xxxxxx-zentren]", synthRE(t, s, "Xxxxxx-zentrum", "SUB:GEN:PLU.*"))
	require.Equal(t, "[]", synthRE(t, s, "gelb-grün", "ADJ.*")) // should this also work?
	require.Equal(t, "[WLAN-LAN-Kabels]", synthRE(t, s, "WLAN-LAN-Kabel", "SUB:GEN:SIN.*"))
	require.Equal(t, "[Häuser, Haus, Häusern, Haus, Hause, Häuser, Hauses, Häuser, Haus]",
		synthRE(t, s, "Haus", ".*"))

	// Java: synthesizer.synthesize(new AnalyzedToken("fake", "FAKE", null), "FAKE", false).length == 0
	fakeTag := "FAKE"
	forms, err := s.SynthesizeRE(languagetool.NewAnalyzedToken("fake", &fakeTag, nil), "FAKE", false)
	require.NoError(t, err)
	require.Equal(t, 0, len(forms))
}

// Port of GermanSynthesizerTest.testSynthesizeCompounds
func TestGermanSynthesizer_SynthesizeCompounds(t *testing.T) {
	s := requireGermanInstance(t)

	require.Equal(t, "[Regelsysteme]", synth(t, s, "Regelsystem", "SUB:NOM:PLU:NEU"))
	require.Equal(t, "[Regelsystemen]", synth(t, s, "Regelsystem", "SUB:DAT:PLU:NEU"))
	require.Equal(t, "[Regelsysteme, Regelsystemen]", synthRE(t, s, "Regelsystem", ".*:PLU:.*"))
	require.Equal(t, "[Regelsysteme, Regelsystemen]", synthRE(t, s, "Regel-System", ".*:PLU:.*"))
	require.Equal(t, "[Kühlschrankversuche, Kühlschrankversuchen]",
		synthRE(t, s, "Kühlschrankversuch", ".*:PLU:.*"))
	require.Equal(t, "[Kühlschrankversuche, Kühlschrankversuchen]",
		synthRE(t, s, "Kühlschrank-Versuch", ".*:PLU:.*"))
}

// Port of GermanSynthesizerTest.testMorfologikBug
func TestGermanSynthesizer_MorfologikBug(t *testing.T) {
	s := requireGermanInstance(t)
	// https://github.com/languagetool-org/languagetool/issues/586
	require.Equal(t, "[anfragen]", synth(t, s, "anfragen", "VER:1:PLU:KJ1:SFT:NEB"))
}

// REMOVE + case gate (Java GermanSynthesizer.lookup / synthesize) — unit path with manual.
func TestGermanSynthesizer_RemoveAndCase(t *testing.T) {
	manual, err := synthesis.NewManualSynthesizer(strings.NewReader(
		"schloß\tschließen\tVER:3:SIN:PRT:NON:NEB\n" +
			"schloss\tschließen\tVER:3:SIN:PRT:NON:NEB\n" +
			"Häuser\thaus\tSUB:NOM:PLU:NEU\n" +
			"häuser\thaus\tSUB:NOM:PLU:NEU\n" +
			"Ihr\tmein\tPRO:PER:NOM:PLU:ALG\n",
	))
	require.NoError(t, err)
	s := NewGermanSynthesizer(manual)

	lem := "schließen"
	tok := languagetool.NewAnalyzedToken("schließen", nil, &lem)
	forms, err := s.Synthesize(tok, "VER:3:SIN:PRT:NON:NEB")
	require.NoError(t, err)
	require.Equal(t, []string{"schloss"}, forms)

	hl := "haus"
	htok := languagetool.NewAnalyzedToken("haus", nil, &hl)
	forms, err = s.Synthesize(htok, "SUB:NOM:PLU:NEU")
	require.NoError(t, err)
	require.Equal(t, []string{"häuser"}, forms)

	ml := "mein"
	mtok := languagetool.NewAnalyzedToken("mein", nil, &ml)
	forms, err = s.Synthesize(mtok, "PRO:PER:NOM:PLU:ALG")
	require.NoError(t, err)
	require.Equal(t, []string{"Ihr"}, forms)
}

// getCompoundForms: whole compound not in synth dict → split last part, rejoin (Java).
func TestGermanSynthesizer_GetCompoundForms(t *testing.T) {
	manual, err := synthesis.NewManualSynthesizer(strings.NewReader(
		"Boote\tBoot\tSUB:NOM:PLU:NEU\n" +
			"Booten\tBoot\tSUB:DAT:PLU:NEU\n",
	))
	require.NoError(t, err)
	s := NewGermanSynthesizer(manual)

	lemma := "Hausboot"
	tok := languagetool.NewAnalyzedToken("Hausboot", nil, &lemma)

	forms, err := s.Synthesize(tok, "SUB:NOM:PLU:NEU")
	require.NoError(t, err)
	require.Empty(t, forms)

	s.StrictCompoundTokenize = func(w string) []string {
		if w == "Hausboot" {
			return []string{"Haus", "Boot"}
		}
		if w == "Haus-Boot" {
			return []string{"Haus-Boot"}
		}
		return []string{w}
	}
	forms, err = s.Synthesize(tok, "SUB:NOM:PLU:NEU")
	require.NoError(t, err)
	require.Equal(t, []string{"Hausboote"}, forms)

	forms, err = s.SynthesizeRE(tok, ".*:PLU:.*", true)
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"Hausboote", "Hausbooten"}, forms)

	hyLemma := "Haus-Boot"
	hyTok := languagetool.NewAnalyzedToken("Haus-Boot", nil, &hyLemma)
	forms, err = s.Synthesize(hyTok, "SUB:NOM:PLU:NEU")
	require.NoError(t, err)
	require.Equal(t, []string{"Haus-Boote"}, forms)
}

// Twin of Java lookup: lemma "ich" allows cross-case only when form is not in REMOVE.
func TestGermanSynthesizer_IchRemoveInCaseGate(t *testing.T) {
	manual, err := synthesis.NewManualSynthesizer(strings.NewReader(
		"unsre\tich\tPRO:PER:NOM:PLU:ALG\n" +
			"Sie\tich\tPRO:PER:NOM:PLU:ALG\n" +
			"sie\tich\tPRO:PER:NOM:PLU:ALG\n",
	))
	require.NoError(t, err)
	s := NewGermanSynthesizer(manual)
	lem := "ich"
	tok := languagetool.NewAnalyzedToken("ich", nil, &lem)
	forms, err := s.Synthesize(tok, "PRO:PER:NOM:PLU:ALG")
	require.NoError(t, err)
	// unsre: same-case passes lookup then REMOVE filter drops it; Sie/sie remain.
	require.Equal(t, []string{"Sie", "sie"}, forms)
}
