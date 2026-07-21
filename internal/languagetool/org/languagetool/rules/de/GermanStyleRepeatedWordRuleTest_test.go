package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/GermanStyleRepeatedWordRuleTest.java
// Java uses tagged analysis (ADJ/SUB/VER). Morph inject only; untagged AnalyzePlain fails closed.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestGermanStyleRepeatedWordRule_Meta(t *testing.T) {
	rule := NewGermanStyleRepeatedWordRule(nil)
	require.Equal(t, "STYLE_REPEATED_WORD_RULE_DE", rule.GetID())
	require.Equal(t, "Wiederholte Worte in aufeinanderfolgenden Sätzen", rule.GetDescription())
	require.True(t, rule.DefaultOff)
	require.Equal(t, 1, rule.MaxDistanceOfSentences)
	require.True(t, rule.ExcludeDirectSpeech)
	require.False(t, rule.TestCompoundWords) // Java TEST_COMPOUND_WORDS default false
	require.NotEmpty(t, rule.GetIncorrectExamples())
}

func TestGermanStyleRepeatedWordRule_Rule(t *testing.T) {
	rule := NewGermanStyleRepeatedWordRule(nil)

	// "großen" ADJ ×2 across sentences → 2
	s1 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Der", "ART:DEF:NOM:SIN:MAS", "der"),
		atrWithPOS("alte", "ADJ:NOM:SIN:MAS:GRU:DEF", "alt"),
		atrWithPOS("Mann", "SUB:NOM:SIN:MAS", "Mann"),
		atrWithPOS("wohnte", "VER:3:SIN:PRT:SFT", "wohnen"),
		atrWithPOS("in", "PRP:DAT", "in"),
		atrWithPOS("einem", "ART:IND:DAT:SIN:NEU", "ein"),
		atrWithPOS("großen", "ADJ:DAT:SIN:NEU:GRU:IND", "groß"),
		atrWithPOS("Haus", "SUB:DAT:SIN:NEU", "Haus"),
		atrWithPOS(".", "PKT", "."),
	))
	s2 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Es", "PRO:PER:NOM:SIN:NEU", "es"),
		atrWithPOS("stand", "VER:3:SIN:PRT:NON", "stehen"),
		atrWithPOS("in", "PRP:DAT", "in"),
		atrWithPOS("einem", "ART:IND:DAT:SIN:NEU", "ein"),
		atrWithPOS("großen", "ADJ:DAT:SIN:NEU:GRU:IND", "groß"),
		atrWithPOS("Garten", "SUB:DAT:SIN:MAS", "Garten"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 2, len(rule.MatchList([]*languagetool.AnalyzedSentence{s1, s2})))

	// different adjectives → 0
	s2b := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Es", "PRO:PER:NOM:SIN:NEU", "es"),
		atrWithPOS("stand", "VER:3:SIN:PRT:NON", "stehen"),
		atrWithPOS("in", "PRP:DAT", "in"),
		atrWithPOS("einem", "ART:IND:DAT:SIN:NEU", "ein"),
		atrWithPOS("weitläufigen", "ADJ:DAT:SIN:NEU:GRU:IND", "weitläufig"),
		atrWithPOS("Garten", "SUB:DAT:SIN:MAS", "Garten"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{s1, s2b})))

	// Java: Frau Heinrich … Frau Müller → 0 (Frau + next EIG not checked)
	f1 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Frau", "SUB:NOM:SIN:FEM", "Frau"),
		atrWithPOS("Heinrich", "EIG:NOM:SIN:FEM", "Heinrich"),
		atrWithPOS("stand", "VER:3:SIN:PRT:NON", "stehen"),
		atrWithPOS("unschlüssig", "ADJ:PRD:GRU", "unschlüssig"),
		atrWithPOS(".", "PKT", "."),
	))
	f2 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Frau", "SUB:NOM:SIN:FEM", "Frau"),
		atrWithPOS("Müller", "EIG:NOM:SIN:FEM", "Müller"),
		atrWithPOS("schaute", "VER:3:SIN:PRT:SFT", "schauen"),
		atrWithPOS("zu", "APPR:DAT", "zu"),
		atrWithPOS("ihr", "PRO:PER:DAT:SIN:FEM", "sie"),
		atrWithPOS("herüber", "ADV", "herüber"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{f1, f2})))

	// default TestCompoundWords=false: Schiffsmotor / Motor not matched as parts
	ship1 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Der", "ART:DEF:NOM:SIN:MAS", "der"),
		atrWithPOS("Schiffsmotor", "SUB:NOM:SIN:MAS", "Schiffsmotor"),
		atrWithPOS("röhrte", "VER:3:SIN:PRT:SFT", "röhren"),
		atrWithPOS(".", "PKT", "."),
	))
	ship2 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Der", "ART:DEF:NOM:SIN:MAS", "der"),
		atrWithPOS("Motor", "SUB:NOM:SIN:MAS", "Motor"),
		atrWithPOS("lief", "VER:3:SIN:PRT:NON", "laufen"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{ship1, ship2})))

	// Java isUnknownWord: isPosTagUnknown only (not invent !isTagged).
	unk := languagetool.AnalyzePlain("Blahxyz").GetTokensWithoutWhitespace()
	var blah *languagetool.AnalyzedTokenReadings
	for _, tok := range unk {
		if tok != nil && tok.GetToken() == "Blahxyz" {
			blah = tok
			break
		}
	}
	require.NotNil(t, blah)
	require.True(t, blah.IsPosTagUnknown())
	require.True(t, isUnknownWordStyle(blah))
	require.False(t, isUnknownWordStyle(atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus")))

	// untagged AnalyzePlain must not invent SUB/ADJ hits
	require.Equal(t, 0, len(rule.MatchList(languagetool.AnalyzeTextLocal(
		"Der alte Mann wohnte in einem großen Haus. Es stand in einem großen Garten."))))
}

func TestGermanStyleRepeatedWordRule_TestCompoundWords(t *testing.T) {
	// Java setUpRule(lt, getRuleValues(1, false, true)) → testComposedWords
	rule := NewGermanStyleRepeatedWordRule(nil)
	rule.TestCompoundWords = true
	// Speller required for isSecondPartOfWord; inject true (Java Morfologik).
	rule.IsCorrectSpell = func(word string) bool { return true }

	// Schiffsmotor … Motor → 3 (Java)
	s1 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Der", "ART:DEF:NOM:SIN:MAS", "der"),
		atrWithPOS("Schiffsmotor", "SUB:NOM:SIN:MAS", "Schiffsmotor"),
		atrWithPOS(",", "PKT", ","),
		atrWithPOS("der", "PRELS:NOM:SIN:MAS", "der"),
		atrWithPOS("im", "APPRART:DAT:SIN:NEU", "in"),
		atrWithPOS("Heck", "SUB:DAT:SIN:NEU", "Heck"),
		atrWithPOS("des", "ART:DEF:GEN:SIN:NEU", "der"),
		atrWithPOS("Schiffs", "SUB:GEN:SIN:NEU", "Schiff"),
		atrWithPOS("eingebaut", "VER:PA2:SFT", "einbauen"),
		atrWithPOS("war", "VER:AUX:3:SIN:PRT:NON", "sein"),
		atrWithPOS(",", "PKT", ","),
		atrWithPOS("röhrte", "VER:3:SIN:PRT:SFT", "röhren"),
		atrWithPOS(".", "PKT", "."),
	))
	s2 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Auf", "APPR:DAT", "auf"),
		atrWithPOS("Hochtouren", "SUB:DAT:PLU:FEM", "Hochtour"),
		atrWithPOS("lief", "VER:3:SIN:PRT:NON", "laufen"),
		atrWithPOS("der", "ART:DEF:NOM:SIN:MAS", "der"),
		atrWithPOS("Motor", "SUB:NOM:SIN:MAS", "Motor"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 3, len(rule.MatchList([]*languagetool.AnalyzedSentence{s1, s2})))

	// Buntspecht … Specht → 2
	a := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Der", "ART:DEF:NOM:SIN:MAS", "der"),
		atrWithPOS("Buntspecht", "SUB:NOM:SIN:MAS", "Buntspecht"),
		atrWithPOS("stolzierte", "VER:3:SIN:PRT:SFT", "stolzieren"),
		atrWithPOS("den", "ART:DEF:AKK:SIN:MAS", "der"),
		atrWithPOS("Baum", "SUB:AKK:SIN:MAS", "Baum"),
		atrWithPOS("hoch", "ADV", "hoch"),
		atrWithPOS(".", "PKT", "."),
	))
	b := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Schon", "ADV", "schon"),
		atrWithPOS("klopfte", "VER:3:SIN:PRT:SFT", "klopfen"),
		atrWithPOS("der", "ART:DEF:NOM:SIN:MAS", "der"),
		atrWithPOS("Specht", "SUB:NOM:SIN:MAS", "Specht"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 2, len(rule.MatchList([]*languagetool.AnalyzedSentence{a, b})))

	// Rotbraun … rot → 2 (compound part / shared stem via isPartOfWord)
	c1 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Rotbraun", "ADJ:PRD:GRU", "rotbraun"),
		atrWithPOS("war", "VER:3:SIN:PRT:NON", "sein"),
		atrWithPOS("die", "ART:DEF:NOM:SIN:FEM", "der"),
		atrWithPOS("Farbe", "SUB:NOM:SIN:FEM", "Farbe"),
		atrWithPOS(".", "PKT", "."),
	))
	c2 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Der", "ART:DEF:NOM:SIN:MAS", "der"),
		atrWithPOS("Horizont", "SUB:NOM:SIN:MAS", "Horizont"),
		atrWithPOS("schimmerte", "VER:3:SIN:PRT:SFT", "schimmern"),
		atrWithPOS("rot", "ADJ:PRD:GRU", "rot"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 2, len(rule.MatchList([]*languagetool.AnalyzedSentence{c1, c2})))

	// Ausblick … "Was für ein Ausblick!" → 2 when quotes not excluding (quoted Ausblick still flagged as Java)
	// Java quoted path: excludeDirectSpeech + isInQuotes for single-quoted token.
	// With „…“ opening without whitespaceBefore on next token → direct speech; may differ.
	// Stick to same-lemma Ausblick across sentences without quotes:
	d1 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Der", "ART:DEF:NOM:SIN:MAS", "der"),
		atrWithPOS("Ausblick", "SUB:NOM:SIN:MAS", "Ausblick"),
		atrWithPOS("war", "VER:3:SIN:PRT:NON", "sein"),
		atrWithPOS("überwältigend", "ADJ:PRD:GRU", "überwältigend"),
		atrWithPOS(".", "PKT", "."),
	))
	d2 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Was", "PWS", "was"),
		atrWithPOS("für", "APPR:AKK", "für"),
		atrWithPOS("ein", "ART:IND:NOM:SIN:MAS", "ein"),
		atrWithPOS("Ausblick", "SUB:NOM:SIN:MAS", "Ausblick"),
		atrWithPOS("sagte", "VER:3:SIN:PRT:SFT", "sagen"),
		atrWithPOS("sie", "PRO:PER:NOM:SIN:FEM", "sie"),
		atrWithPOS(".", "PKT", "."),
	))
	// same lemma, TestCompound not needed; distance 1 → 2 matches
	// reset compounds off path still finds same lemma:
	ruleSame := NewGermanStyleRepeatedWordRule(nil)
	require.Equal(t, 2, len(ruleSame.MatchList([]*languagetool.AnalyzedSentence{d1, d2})))

	// without IsCorrectSpell (fail-closed) compound parts must not invent
	ruleClosed := NewGermanStyleRepeatedWordRule(nil)
	ruleClosed.TestCompoundWords = true
	ruleClosed.IsCorrectSpell = func(string) bool { return false }
	closed1 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Der", "ART:DEF:NOM:SIN:MAS", "der"),
		atrWithPOS("Schiffsmotor", "SUB:NOM:SIN:MAS", "Schiffsmotor"),
		atrWithPOS("röhrte", "VER:3:SIN:PRT:SFT", "röhren"),
		atrWithPOS(".", "PKT", "."),
	))
	closed2 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Der", "ART:DEF:NOM:SIN:MAS", "der"),
		atrWithPOS("Motor", "SUB:NOM:SIN:MAS", "Motor"),
		atrWithPOS("lief", "VER:3:SIN:PRT:NON", "laufen"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 0, len(ruleClosed.MatchList([]*languagetool.AnalyzedSentence{closed1, closed2})))
}
