package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/CompoundCoherencyRuleTest.java
// Surface forms via AnalyzePlain; inflected pairs use lemma inject (Java Morfologik lemmas).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestCompoundCoherencyRule_Rule(t *testing.T) {
	rule := NewCompoundCoherencyRule(nil)

	assertOkay := func(s1, s2 string) {
		t.Helper()
		ms := rule.MatchList([]*languagetool.AnalyzedSentence{
			languagetool.AnalyzePlain(s1),
			languagetool.AnalyzePlain(s2),
		})
		require.Equal(t, 0, len(ms), "okay %q | %q", s1, s2)
	}
	assertError := func(s1, s2 string, fromPos, toPos int, suggestion *string) {
		t.Helper()
		ms := rule.MatchList([]*languagetool.AnalyzedSentence{
			languagetool.AnalyzePlain(s1),
			languagetool.AnalyzePlain(s2),
		})
		require.Equal(t, 1, len(ms), "error %q | %q got %d", s1, s2, len(ms))
		require.Equal(t, fromPos, ms[0].GetFromPos(), "fromPos %q | %q", s1, s2)
		require.Equal(t, toPos, ms[0].GetToPos(), "toPos %q | %q", s1, s2)
		if suggestion == nil {
			require.Empty(t, ms[0].GetSuggestedReplacements(), "no sugg %q | %q", s1, s2)
		} else {
			require.Equal(t, []string{*suggestion}, ms[0].GetSuggestedReplacements(), "sugg %q | %q", s1, s2)
		}
	}
	strp := func(s string) *string { return &s }

	// Java assertOkay (same surface/hyphenation family without inconsistency):
	assertOkay("Ein Jugendfoto.", "Und ein Jugendfoto.")
	assertOkay("Der Zahn-Ärzte-Verband.", "Der Zahn-Ärzte-Verband.")
	assertOkay("Es gibt E-Mail.", "Und es gibt E-Mails.")
	assertOkay("Es gibt E-Mails.", "Und es gibt E-Mail.")
	assertOkay("Ein Jugend-Foto.", "Der Rahmen eines Jugend-Fotos.")

	// Java assertError surface (same uninflected strip → suggestion = first form):
	assertError("Ein Jugendfoto.", "Und ein Jugend-Foto.", 23, 34, strp("Jugendfoto"))
	assertError("Ein Jugend-Foto.", "Und ein Jugendfoto.", 24, 34, strp("Jugend-Foto"))

	assertError("Der Zahn-Ärzte-Verband.", "Der Zahn-Ärzteverband.", 27, 44, strp("Zahn-Ärzte-Verband"))
	assertError("Der Zahn-Ärzte-Verband.", "Der Zahnärzte-Verband.", 27, 44, strp("Zahn-Ärzte-Verband"))
	assertError("Der Zahn-Ärzte-Verband.", "Der Zahnärzteverband.", 27, 43, strp("Zahn-Ärzte-Verband"))
	assertError("Der Zahn-Ärzteverband.", "Der Zahn-Ärzte-Verband.", 26, 44, strp("Zahn-Ärzteverband"))
	assertError("Der Zahnärzte-Verband.", "Der Zahn-Ärzte-Verband.", 26, 44, strp("Zahnärzte-Verband"))
	assertError("Der Zahnärzteverband.", "Der Zahn-Ärzte-Verband.", 25, 43, strp("Zahnärzteverband"))

	require.Equal(t, "Einheitliche Schreibweise bei Komposita (mit oder ohne Bindestrich)", rule.GetDescription())
	require.Equal(t, "DE_COMPOUND_COHERENCY", rule.GetID())
	require.Equal(t, -1, rule.MinToCheckParagraph())
	require.NotEmpty(t, rule.GetIncorrectExamples())
}

// injectLemma replaces readings for a surface token so HasSameLemmas + getLemma path runs (Java tagger).
func injectLemma(s *languagetool.AnalyzedSentence, surface, lemma string) {
	for _, tok := range s.GetTokensWithoutWhitespace() {
		if tok != nil && tok.GetToken() == surface {
			pos := "SUB:NOM:PLU:MAS"
			l := lemma
			p := pos
			newTok := languagetool.NewAnalyzedToken(surface, &p, &l)
			*tok = *languagetool.NewAnalyzedTokenReadingsFromOld(tok, []*languagetool.AnalyzedToken{newTok}, "test")
		}
	}
}

// Inflected pairs: Java stores lemmas so suggestion is null when surfaces differ after strip.
func TestCompoundCoherencyRule_InflectedNullSuggestion(t *testing.T) {
	rule := NewCompoundCoherencyRule(nil)

	// Viele Zahn-Ärzte. / Oder Zahnärzte. — Java 22–31, no suggestion
	s1 := languagetool.AnalyzePlain("Viele Zahn-Ärzte.")
	s2 := languagetool.AnalyzePlain("Oder Zahnärzte.")
	injectLemma(s1, "Zahn-Ärzte", "Zahnarzt")
	injectLemma(s2, "Zahnärzte", "Zahnarzt")
	ms := rule.MatchList([]*languagetool.AnalyzedSentence{s1, s2})
	require.Equal(t, 1, len(ms))
	require.Equal(t, 22, ms[0].GetFromPos())
	require.Equal(t, 31, ms[0].GetToPos())
	require.Empty(t, ms[0].GetSuggestedReplacements())

	// Viele Zahn-Ärzte. / Oder Zahnärzten. — 22–32, no suggestion
	s1b := languagetool.AnalyzePlain("Viele Zahn-Ärzte.")
	s2b := languagetool.AnalyzePlain("Oder Zahnärzten.")
	injectLemma(s1b, "Zahn-Ärzte", "Zahnarzt")
	injectLemma(s2b, "Zahnärzten", "Zahnarzt")
	ms = rule.MatchList([]*languagetool.AnalyzedSentence{s1b, s2b})
	require.Equal(t, 1, len(ms))
	require.Equal(t, 22, ms[0].GetFromPos())
	require.Equal(t, 32, ms[0].GetToPos())
	require.Empty(t, ms[0].GetSuggestedReplacements())

	// Jugendfoto / Jugendfotos okay with shared lemma
	j1 := languagetool.AnalyzePlain("Ein Jugendfoto.")
	j2 := languagetool.AnalyzePlain("Der Rahmen eines Jugendfotos.")
	injectLemma(j1, "Jugendfoto", "Jugendfoto")
	injectLemma(j2, "Jugendfotos", "Jugendfoto")
	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{j1, j2})))
	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{j2, j1})))

	// Verbands / Verbandes same hyphenation → okay
	v1 := languagetool.AnalyzePlain("Der Zahn-Ärzte-Verband.")
	v2 := languagetool.AnalyzePlain("Des Zahn-Ärzte-Verbands.")
	injectLemma(v1, "Zahn-Ärzte-Verband", "Zahn-Ärzte-Verband")
	injectLemma(v2, "Zahn-Ärzte-Verbands", "Zahn-Ärzte-Verband")
	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{v1, v2})))
	v3 := languagetool.AnalyzePlain("Des Zahn-Ärzte-Verbandes.")
	injectLemma(v3, "Zahn-Ärzte-Verbandes", "Zahn-Ärzte-Verband")
	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{v1, v3})))

	assertInflectedError := func(s1, s2 string, from, to int, surface1, surface2, lemma string) {
		t.Helper()
		a := languagetool.AnalyzePlain(s1)
		b := languagetool.AnalyzePlain(s2)
		injectLemma(a, surface1, lemma)
		injectLemma(b, surface2, lemma)
		got := rule.MatchList([]*languagetool.AnalyzedSentence{a, b})
		require.Equal(t, 1, len(got), "%q | %q", s1, s2)
		require.Equal(t, from, got[0].GetFromPos(), "from %q | %q", s1, s2)
		require.Equal(t, to, got[0].GetToPos(), "to %q | %q", s1, s2)
		require.Empty(t, got[0].GetSuggestedReplacements(), "sugg %q | %q", s1, s2)
	}

	// Java genitive inconsistency (lemma unifies norm key; surface strip differs → no sugg)
	assertInflectedError("Der Zahn-Ärzte-Verband.", "Des Zahn-Ärzteverbandes.", 27, 46,
		"Zahn-Ärzte-Verband", "Zahn-Ärzteverbandes", "Zahnärzteverband")
	assertInflectedError("Der Zahn-Ärzte-Verband.", "Des Zahnärzte-Verbandes.", 27, 46,
		"Zahn-Ärzte-Verband", "Zahnärzte-Verbandes", "Zahnärzteverband")
	assertInflectedError("Der Zahn-Ärzte-Verband.", "Des Zahnärzteverbandes.", 27, 45,
		"Zahn-Ärzte-Verband", "Zahnärzteverbandes", "Zahnärzteverband")
}

func TestCompoundCoherencyRule_NumericBreak(t *testing.T) {
	// Java: isNumeric(normToken) breaks the token loop for that sentence
	rule := NewCompoundCoherencyRule(nil)
	ms := rule.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Von 2-3."),
		languagetool.AnalyzePlain("Bis 23."),
	})
	require.Equal(t, 0, len(ms))
}
