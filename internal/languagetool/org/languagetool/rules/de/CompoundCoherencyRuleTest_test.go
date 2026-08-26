package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/CompoundCoherencyRuleTest.java
//
// Java:
//   CompoundCoherencyRule rule = new CompoundCoherencyRule(TestTools.getEnglishMessages());
//   JLanguageTool lt = new JLanguageTool(Languages.getLanguageForShortCode("de"));
//   rule.match(Arrays.asList(lt.getAnalyzedSentence(s1), lt.getAnalyzedSentence(s2)));
//
// Go: de-DE JLanguageTool + WireGermanTagWord (createDefaultTagger) so Morfologik lemmas
// feed getLemma / hyphen reconstruct the same way as Java. No lemma invent for the
// main matrix. @Ignore testRuleInteractive is omitted (debug-only).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// newCompoundCoherencyFixture mirrors Java CompoundCoherencyRuleTest fields.
func newCompoundCoherencyFixture(t *testing.T) (*CompoundCoherencyRule, *languagetool.JLanguageTool) {
	t.Helper()
	rule := NewCompoundCoherencyRule(nil) // Java: TestTools.getEnglishMessages() — STYLE category only
	lt := languagetool.NewJLanguageTool("de-DE")
	require.True(t, WireGermanTagWord(lt), "GermanTagger resources required for CompoundCoherencyRule twin")
	return rule, lt
}

// getCompoundCoherencyMatches ports private getMatches(s1, s2).
func getCompoundCoherencyMatches(rule *CompoundCoherencyRule, lt *languagetool.JLanguageTool, s1, s2 string) []*rules.RuleMatch {
	sent1 := lt.GetAnalyzedSentence(s1)
	sent2 := lt.GetAnalyzedSentence(s2)
	return rule.MatchList([]*languagetool.AnalyzedSentence{sent1, sent2})
}

func TestCompoundCoherencyRule_Rule(t *testing.T) {
	rule, lt := newCompoundCoherencyFixture(t)

	assertOkay := func(s1, s2 string) {
		t.Helper()
		ms := getCompoundCoherencyMatches(rule, lt, s1, s2)
		require.Equal(t, 0, len(ms), "okay %q | %q got %v", s1, s2, ms)
	}
	assertError := func(s1, s2 string, fromPos, toPos int, suggestion *string) {
		t.Helper()
		ms := getCompoundCoherencyMatches(rule, lt, s1, s2)
		require.Equal(t, 1, len(ms), "error %q | %q got %d matches", s1, s2, len(ms))
		require.Equal(t, fromPos, ms[0].GetFromPos(), "fromPos %q | %q", s1, s2)
		require.Equal(t, toPos, ms[0].GetToPos(), "toPos %q | %q", s1, s2)
		if suggestion == nil {
			require.Empty(t, ms[0].GetSuggestedReplacements(), "no sugg %q | %q", s1, s2)
		} else {
			require.Equal(t, []string{*suggestion}, ms[0].GetSuggestedReplacements(), "sugg %q | %q", s1, s2)
		}
	}
	strp := func(s string) *string { return &s }

	// Java CompoundCoherencyRuleTest.testRule assertOkay:
	assertOkay("Ein Jugendfoto.", "Und ein Jugendfoto.")
	assertOkay("Ein Jugendfoto.", "Der Rahmen eines Jugendfotos.")
	assertOkay("Der Rahmen eines Jugendfotos.", "Ein Jugendfoto.")
	assertOkay("Der Zahn-Ärzte-Verband.", "Der Zahn-Ärzte-Verband.")
	assertOkay("Der Zahn-Ärzte-Verband.", "Des Zahn-Ärzte-Verbands.")
	assertOkay("Der Zahn-Ärzte-Verband.", "Des Zahn-Ärzte-Verbandes.")
	assertOkay("Es gibt E-Mail.", "Und es gibt E-Mails.")
	assertOkay("Es gibt E-Mails.", "Und es gibt E-Mail.")
	assertOkay("Ein Jugend-Foto.", "Der Rahmen eines Jugend-Fotos.")

	// Java assertError (positions + suggestions; null → empty replacements):
	assertError("Ein Jugendfoto.", "Und ein Jugend-Foto.", 23, 34, strp("Jugendfoto"))
	assertError("Ein Jugend-Foto.", "Und ein Jugendfoto.", 24, 34, strp("Jugend-Foto"))

	assertError("Viele Zahn-Ärzte.", "Oder Zahnärzte.", 22, 31, nil)
	assertError("Viele Zahn-Ärzte.", "Oder Zahnärzte.", 22, 31, nil) // Java duplicate line
	assertError("Viele Zahn-Ärzte.", "Oder Zahnärzten.", 22, 32, nil)

	assertError("Der Zahn-Ärzte-Verband.", "Der Zahn-Ärzteverband.", 27, 44, strp("Zahn-Ärzte-Verband"))
	assertError("Der Zahn-Ärzte-Verband.", "Der Zahnärzte-Verband.", 27, 44, strp("Zahn-Ärzte-Verband"))
	assertError("Der Zahn-Ärzte-Verband.", "Der Zahnärzteverband.", 27, 43, strp("Zahn-Ärzte-Verband"))
	assertError("Der Zahn-Ärzteverband.", "Der Zahn-Ärzte-Verband.", 26, 44, strp("Zahn-Ärzteverband"))
	assertError("Der Zahnärzte-Verband.", "Der Zahn-Ärzte-Verband.", 26, 44, strp("Zahnärzte-Verband"))
	assertError("Der Zahnärzteverband.", "Der Zahn-Ärzte-Verband.", 25, 43, strp("Zahnärzteverband"))

	assertError("Der Zahn-Ärzte-Verband.", "Des Zahn-Ärzteverbandes.", 27, 46, nil)
	assertError("Der Zahn-Ärzte-Verband.", "Des Zahnärzte-Verbandes.", 27, 46, nil)
	assertError("Der Zahn-Ärzte-Verband.", "Des Zahnärzteverbandes.", 27, 45, nil)

	// Constructor / TextLevelRule surface (Java STYLE + example pair + minToCheckParagraph -1)
	require.Equal(t, "DE_COMPOUND_COHERENCY", rule.GetID())
	require.Equal(t, "Einheitliche Schreibweise bei Komposita (mit oder ohne Bindestrich)", rule.GetDescription())
	require.Equal(t, -1, rule.MinToCheckParagraph())
	require.NotNil(t, rule.GetCategory())
	require.NotEmpty(t, rule.GetIncorrectExamples())
	require.NotEmpty(t, rule.GetCorrectExamples())
	require.Contains(t, rule.GetIncorrectExamples()[0].GetExample(), "Help-Desks")
	require.Contains(t, rule.GetCorrectExamples()[0].GetExample(), "Helpdesks")
}

func TestCompoundCoherencyRule_NumericBreak(t *testing.T) {
	// Java: StringUtils.isNumeric(normToken) → break token loop (avoid "2-3" vs "23")
	rule := NewCompoundCoherencyRule(nil)
	ms := rule.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Von 2-3."),
		languagetool.AnalyzePlain("Bis 23."),
	})
	require.Equal(t, 0, len(ms))
}

func TestCompoundCoherencyRule_GetLemmaHyphenReconstruct(t *testing.T) {
	// Java getLemma: lemma "Jugendfoto" for surface "Jugend-Fotos" → "Jugend-Foto"
	lemma := "Jugendfoto"
	surface := "Jugend-Fotos"
	pos := "SUB:NOM:SIN:NEU"
	atr := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken(surface, &pos, &lemma),
	}, 0)
	require.Equal(t, "Jugend-Foto", compoundCoherencyLemma(atr))

	// Multi-hyphen reconstruct (Zahnarzt + Zahn-Ärzte → Zahn-Arzt)
	lemma2 := "Zahnarzt"
	surface2 := "Zahn-Ärzte"
	atr2 := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken(surface2, &pos, &lemma2),
	}, 0)
	require.Equal(t, "Zahn-Arzt", compoundCoherencyLemma(atr2))

	// No reconstruct when lemma already has hyphen
	lemma3 := "E-Mail"
	surface3 := "E-Mails"
	atr3 := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken(surface3, &pos, &lemma3),
	}, 0)
	require.Equal(t, "E-Mail", compoundCoherencyLemma(atr3))

	// hasSameLemmas false / empty → no lemma (caller uses surface token)
	atr4 := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("x", &pos, nil),
	}, 0)
	require.Equal(t, "", compoundCoherencyLemma(atr4))
}
