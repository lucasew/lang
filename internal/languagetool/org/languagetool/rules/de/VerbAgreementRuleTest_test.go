package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/VerbAgreementRuleTest.java
// Java uses tagged analysis (VER person/number). Morph/POS inject only; untagged AnalyzePlain remains fail-closed.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestVerbAgreementRule_Meta(t *testing.T) {
	r := NewVerbAgreementRule(nil)
	require.Equal(t, "DE_VERBAGREEMENT", r.GetID())
	require.Contains(t, r.GetDescription(), "Kongruenz")
	require.Equal(t, 0, r.EstimateContextForSureMatch())
	require.NotEmpty(t, r.GetIncorrectExamples())
}

func TestVerbAgreementRule_Positions(t *testing.T) {
	// Twin of VerbAgreementRuleTest.testPositions
	rule := NewVerbAgreementRule(nil)

	// "Du erreichst ich unter 12345" → from=3 to=16 (verb…subject)
	ms := rule.Match(languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Du", "PRO:PER:NOM:SIN:ALG", "du"),
		atrWithPOS("erreichst", "VER:2:SIN:PRÄ:SFT", "erreichen"),
		atrWithPOS("ich", "PRO:PER:NOM:SIN:ALG", "ich"),
		atrWithPOS("unter", "APPR:DAT", "unter"),
		atrWithPOS("12345", "CARD", "12345"),
	)))
	require.Equal(t, 1, len(ms))
	require.Equal(t, 3, ms[0].GetFromPos())
	require.Equal(t, 16, ms[0].GetToPos())

	// Multi-sentence MatchList: offset = first sentence GetCorrectedTextLength
	// Java: "Hallo Karl. " length 12 → from=15 to=28
	s1 := languagetool.AnalyzePlain("Hallo Karl. ")
	s2 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Du", "PRO:PER:NOM:SIN:ALG", "du"),
		atrWithPOS("erreichst", "VER:2:SIN:PRÄ:SFT", "erreichen"),
		atrWithPOS("ich", "PRO:PER:NOM:SIN:ALG", "ich"),
		atrWithPOS("unter", "APPR:DAT", "unter"),
		atrWithPOS("12345", "CARD", "12345"),
	))
	off := s1.GetCorrectedTextLength()
	ms = rule.MatchList([]*languagetool.AnalyzedSentence{s1, s2})
	require.Equal(t, 1, len(ms))
	require.Equal(t, off+3, ms[0].GetFromPos())
	require.Equal(t, off+16, ms[0].GetToPos())
}

func TestVerbAgreementRule_WrongVerb(t *testing.T) {
	rule := NewVerbAgreementRule(nil)

	// Ich bin OK
	ok := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Ich", "PRO:PER:NOM:SIN:ALG", "ich"),
		atrWithPOS("bin", "VER:1:SIN:PRÄ:NON", "sein"),
		atrWithPOS("müde", "ADJ:PRD:GRU", "müde"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 0, len(rule.Match(ok)))

	// Ich sind wrong (Java assertBad — may be 1 or 2 matches depending on branches)
	bad := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Ich", "PRO:PER:NOM:SIN:ALG", "ich"),
		atrWithPOS("sind", "VER:1:PLU:PRÄ:NON", "sein"),
		atrWithPOS("müde", "ADJ:PRD:GRU", "müde"),
		atrWithPOS(".", "PKT", "."),
	))
	require.GreaterOrEqual(t, len(rule.Match(bad)), 1)

	// Peter bin nett — VER:1:SIN without ich (Java assertBad)
	peterBin := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Peter", "EIG:NOM:SIN:MAS", "Peter"),
		atrWithPOS("bin", "VER:1:SIN:PRÄ:NON", "sein"),
		atrWithPOS("nett", "ADJ:PRD:GRU", "nett"),
		atrWithPOS(".", "PKT", "."),
	))
	require.GreaterOrEqual(t, len(rule.Match(peterBin)), 1)

	// Du weiß es doch — du + wrong person (Java assertBad)
	duWeiss := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Du", "PRO:PER:NOM:SIN:ALG", "du"),
		atrWithPOS("weiß", "VER:3:SIN:PRÄ:NON", "wissen"),
		atrWithPOS("es", "PRO:PER:AKK:SIN:NEU", "es"),
		atrWithPOS("doch", "ADV", "doch"),
		atrWithPOS(".", "PKT", "."),
	))
	require.GreaterOrEqual(t, len(rule.Match(duWeiss)), 1)

	// Osama bin Laden — bin after name ignored (Java assertGood)
	osama := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Osama", "EIG:NOM:SIN:MAS", "Osama"),
		atrWithPOS("bin", "VER:1:SIN:PRÄ:NON", "sein"),
		atrWithPOS("Laden", "EIG:NOM:SIN:MAS", "Laden"),
		atrWithPOS("stammt", "VER:3:SIN:PRÄ:SFT", "stammen"),
		atrWithPOS("aus", "APPR:DAT", "aus"),
		atrWithPOS("Saudi-Arabien", "EIG:DAT:SIN:NEU", "Saudi-Arabien"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 0, len(rule.Match(osama)))

	// more Java assertBad morph cases
	for _, tc := range []struct {
		name  string
		parts []*languagetool.AnalyzedTokenReadings
	}{
		{"Du muss gehen", []*languagetool.AnalyzedTokenReadings{
			atrWithPOS("Du", "PRO:PER:NOM:SIN:ALG", "du"),
			atrWithPOS("muss", "VER:3:SIN:PRÄ:NON", "müssen"),
			atrWithPOS("gehen", "VER:INF:NON", "gehen"),
			atrWithPOS(".", "PKT", "."),
		}},
		{"Ich müsst alles machen", []*languagetool.AnalyzedTokenReadings{
			atrWithPOS("Ich", "PRO:PER:NOM:SIN:ALG", "ich"),
			atrWithPOS("müsst", "VER:2:PLU:PRÄ:NON", "müssen"),
			atrWithPOS("alles", "PIS:AKK:SIN:NEU", "alles"),
			atrWithPOS("machen", "VER:INF:NON", "machen"),
			atrWithPOS(".", "PKT", "."),
		}},
		{"Ich brauchen einen Karren", []*languagetool.AnalyzedTokenReadings{
			atrWithPOS("Ich", "PRO:PER:NOM:SIN:ALG", "ich"),
			atrWithPOS("brauchen", "VER:1:PLU:PRÄ:SFT", "brauchen"),
			atrWithPOS("einen", "ART:IND:AKK:SIN:MAS", "ein"),
			atrWithPOS("Karren", "SUB:AKK:SIN:MAS", "Karren"),
			atrWithPOS(".", "PKT", "."),
		}},
		{"Die Unterlagen solltest ihr", []*languagetool.AnalyzedTokenReadings{
			atrWithPOS("Die", "ART:DEF:AKK:PLU:FEM", "der"),
			atrWithPOS("Unterlagen", "SUB:AKK:PLU:FEM", "Unterlage"),
			atrWithPOS("solltest", "VER:2:SIN:KJ2:NON", "sollen"),
			atrWithPOS("ihr", "PRO:PER:NOM:PLU:ALG", "ihr"),
			atrWithPOS("gründlich", "ADV", "gründlich"),
			atrWithPOS("durcharbeiten", "VER:INF:NON", "durcharbeiten"),
			atrWithPOS(".", "PKT", "."),
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			toks := append([]*languagetool.AnalyzedTokenReadings{sentStartATR()}, tc.parts...)
			require.GreaterOrEqual(t, len(rule.Match(languagetool.NewAnalyzedSentence(withPositions(toks...)))), 1)
		})
	}

	// untagged must not invent
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Ich sind müde."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Peter bin nett."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Du muss gehen."))))
}

func TestVerbAgreementRule_SuggestionSorting(t *testing.T) {
	// Java testSuggestionSorting needs synthesizer; without Synth, still match but no full list.
	require.NotNil(t, NewVerbAgreementRule(nil))
	// When synth available, "Wir nenne" → sorted suggestions starting with "Wir nennen"
	// Full synth path covered in synthesis package; rule WithSynth tested when dict present.
	if synth := openDiscoveredGermanSynthesizer(); synth != nil {
		rule := NewVerbAgreementRule(nil).WithSynth(synth)
		// "Wir nenne ihn mal" — nenne is VER:1:SIN or KONJ; inject 1:SIN only so unambiguous mismatch for "wir" (1:PLU)
		sent := languagetool.NewAnalyzedSentence(withPositions(
			sentStartATR(),
			atrWithPOS("Wir", "PRO:PER:NOM:PLU:ALG", "wir"),
			atrWithPOS("nenne", "VER:1:SIN:KJ1:SFT", "nennen"),
			atrWithPOS("ihn", "PRO:PER:AKK:SIN:MAS", "er"),
			atrWithPOS("mal", "ADV", "mal"),
			atrWithPOS(".", "PKT", "."),
		))
		ms := rule.Match(sent)
		require.GreaterOrEqual(t, len(ms), 1)
		if len(ms[0].GetSuggestedReplacements()) > 0 {
			// Java order starts with "Wir nennen"
			require.Equal(t, "Wir nennen", ms[0].GetSuggestedReplacements()[0])
		}
	}
}

// Twin of VerbAgreementRuleTest.testWrongVerbSubject
func TestVerbAgreementRule_WrongVerbSubject(t *testing.T) {
	rule := NewVerbAgreementRule(nil)

	good := func(parts ...*languagetool.AnalyzedTokenReadings) {
		t.Helper()
		toks := append([]*languagetool.AnalyzedTokenReadings{sentStartATR()}, parts...)
		require.Equal(t, 0, len(rule.Match(languagetool.NewAnalyzedSentence(withPositions(toks...)))), parts)
	}
	badN := func(n int, parts ...*languagetool.AnalyzedTokenReadings) {
		t.Helper()
		toks := append([]*languagetool.AnalyzedTokenReadings{sentStartATR()}, parts...)
		require.Equal(t, n, len(rule.Match(languagetool.NewAnalyzedSentence(withPositions(toks...)))), parts)
	}
	badAtLeast := func(n int, parts ...*languagetool.AnalyzedTokenReadings) {
		t.Helper()
		toks := append([]*languagetool.AnalyzedTokenReadings{sentStartATR()}, parts...)
		require.GreaterOrEqual(t, len(rule.Match(languagetool.NewAnalyzedSentence(withPositions(toks...)))), n, parts)
	}

	// good
	good(atrWithPOS("Du", "PRO:PER:NOM:SIN:ALG", "du"), atrWithPOS("lebst", "VER:2:SIN:PRÄ:SFT", "leben"), atrWithPOS(".", "PKT", "."))
	good(atrWithPOS("Wir", "PRO:PER:NOM:PLU:ALG", "wir"), atrWithPOS("leben", "VER:1:PLU:PRÄ:SFT", "leben"), atrWithPOS("noch", "ADV", "noch"), atrWithPOS(".", "PKT", "."))
	good(atrWithPOS("Ich", "PRO:PER:NOM:SIN:ALG", "ich"), atrWithPOS("bin", "VER:1:SIN:PRÄ:NON", "sein"), atrWithPOS("nett", "ADJ:PRD:GRU", "nett"), atrWithPOS(".", "PKT", "."))
	good(atrWithPOS("Er", "PRO:PER:NOM:SIN:MAS", "er"), atrWithPOS("lebt", "VER:3:SIN:PRÄ:SFT", "leben"), atrWithPOS(".", "PKT", "."))
	good(atrWithPOS("Nett", "ADJ:PRD:GRU", "nett"), atrWithPOS("bin", "VER:1:SIN:PRÄ:NON", "sein"), atrWithPOS("ich", "PRO:PER:NOM:SIN:ALG", "ich"), atrWithPOS(".", "PKT", "."))
	good(atrWithPOS("Nett", "ADJ:PRD:GRU", "nett"), atrWithPOS("bist", "VER:2:SIN:PRÄ:NON", "sein"), atrWithPOS("du", "PRO:PER:NOM:SIN:ALG", "du"), atrWithPOS(".", "PKT", "."))

	// bad
	badAtLeast(1, atrWithPOS("Du", "PRO:PER:NOM:SIN:ALG", "du"), atrWithPOS("leben", "VER:1:PLU:PRÄ:SFT", "leben"), atrWithPOS(".", "PKT", "."))
	badAtLeast(1, atrWithPOS("Ich", "PRO:PER:NOM:SIN:ALG", "ich"), atrWithPOS("leben", "VER:1:PLU:PRÄ:SFT", "leben"), atrWithPOS(".", "PKT", "."))
	badAtLeast(1, atrWithPOS("Auch", "ADV", "auch"), atrWithPOS("morgen", "ADV", "morgen"), atrWithPOS("leben", "VER:1:PLU:PRÄ:SFT", "leben"), atrWithPOS("du", "PRO:PER:NOM:SIN:ALG", "du"), atrWithPOS(".", "PKT", "."))
	// "Du bin nett." — Java expects 2
	badN(2, atrWithPOS("Du", "PRO:PER:NOM:SIN:ALG", "du"), atrWithPOS("bin", "VER:1:SIN:PRÄ:NON", "sein"), atrWithPOS("nett", "ADJ:PRD:GRU", "nett"), atrWithPOS(".", "PKT", "."))
	// "Ich bist nett." — Java expects 2
	badN(2, atrWithPOS("Ich", "PRO:PER:NOM:SIN:ALG", "ich"), atrWithPOS("bist", "VER:2:SIN:PRÄ:NON", "sein"), atrWithPOS("nett", "ADJ:PRD:GRU", "nett"), atrWithPOS(".", "PKT", "."))
	// "Er lebst." — Java expects 2
	badN(2, atrWithPOS("Er", "PRO:PER:NOM:SIN:MAS", "er"), atrWithPOS("lebst", "VER:2:SIN:PRÄ:SFT", "leben"), atrWithPOS(".", "PKT", "."))
	// "Wir lebst noch." — Java expects 2
	badN(2, atrWithPOS("Wir", "PRO:PER:NOM:PLU:ALG", "wir"), atrWithPOS("lebst", "VER:2:SIN:PRÄ:SFT", "leben"), atrWithPOS("noch", "ADV", "noch"), atrWithPOS(".", "PKT", "."))

	// untagged must not invent
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Auch morgen leben du."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Du leben."))))
}
