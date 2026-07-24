package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/VerbAgreementRuleTest.java
// Java uses tagged analysis (VER person/number). Morph/POS inject only; untagged AnalyzePlain remains fail-closed.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// atrWithPOSMulti builds a token with multiple POS readings (Java Morphy ambiguity).
func atrWithPOSMulti(token, lemma string, tags ...string) *languagetool.AnalyzedTokenReadings {
	if len(tags) == 0 {
		return atrWithPOS(token, "", lemma)
	}
	atr := atrWithPOS(token, tags[0], lemma)
	for _, tag := range tags[1:] {
		tt, ll := tag, lemma
		atr.AddReading(languagetool.NewAnalyzedToken(token, &tt, &ll), "")
	}
	return atr
}

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

	// Java match4 FP fixed: "Mir ist bewusst, dass viele Menschen wie du empfinden."
	// morph: du is not subject of empfinden (wie-clause) → no invent hit
	ms = rule.Match(languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Mir", "PRO:PER:DAT:SIN:1", "ich"),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("bewusst", "ADJ:PRD:GRU", "bewusst"),
		atrWithPOS(",", "PKT", ","),
		atrWithPOS("dass", "KOUS", "dass"),
		atrWithPOS("viele", "PIAT:NOM:PLU:MAS", "viel"),
		atrWithPOS("Menschen", "SUB:NOM:PLU:MAS", "Mensch"),
		atrWithPOS("wie", "KOKOM", "wie"),
		atrWithPOS("du", "PRO:PER:NOM:SIN:ALG", "du"),
		atrWithPOS("empfinden", "VER:3:PLU:PRÄ:SFT", "empfinden"),
		atrWithPOS(".", "PKT", "."),
	)))
	require.Equal(t, 0, len(ms), "wie du empfinden must not invent du/empfinden agreement")
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

	// more Java assertBad morph cases (testWrongVerb)
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
		// Als Borcarbid weißt es …
		{"Als Borcarbid weißt es", []*languagetool.AnalyzedTokenReadings{
			atrWithPOS("Als", "KOUS", "als"),
			atrWithPOS("Borcarbid", "SUB:NOM:SIN:NEU", "Borcarbid"),
			atrWithPOS("weißt", "VER:2:SIN:PRÄ:SFT", "wissen"),
			atrWithPOS("es", "PRO:PER:NOM:SIN:NEU", "es"),
			atrWithPOS("eine", "ART:IND:AKK:SIN:FEM", "ein"),
			atrWithPOS("hohe", "ADJ:AKK:SIN:FEM:GRU:IND", "hoch"),
			atrWithPOS("Härte", "SUB:AKK:SIN:FEM", "Härte"),
			atrWithPOS("auf", "PTKVZ", "auf"),
			atrWithPOS(".", "PKT", "."),
		}},
		// Die Eisenbahn dienst …
		{"Die Eisenbahn dienst", []*languagetool.AnalyzedTokenReadings{
			atrWithPOS("Die", "ART:DEF:NOM:SIN:FEM", "die"),
			atrWithPOS("Eisenbahn", "SUB:NOM:SIN:FEM", "Eisenbahn"),
			atrWithPOS("dienst", "VER:2:SIN:PRÄ:SFT", "dienen"),
			atrWithPOS("überwiegend", "ADV", "überwiegend"),
			atrWithPOS("dem", "ART:DEF:DAT:SIN:MAS", "der"),
			atrWithPOS("Güterverkehr", "SUB:DAT:SIN:MAS", "Güterverkehr"),
			atrWithPOS(".", "PKT", "."),
		}},
		// Weiter befindest sich …
		{"Weiter befindest sich", []*languagetool.AnalyzedTokenReadings{
			atrWithPOS("Weiter", "ADV", "weiter"),
			atrWithPOS("befindest", "VER:2:SIN:PRÄ:SFT", "befinden"),
			atrWithPOS("sich", "PRF:AKK:SIN:3", "sich"),
			atrWithPOS("im", "APPRART:DAT:SIN:MAS", "in"),
			atrWithPOS("Osten", "SUB:DAT:SIN:MAS", "Osten"),
			atrWithPOS("die", "ART:DEF:NOM:SIN:FEM", "die"),
			atrWithPOS("Gemeinde", "SUB:NOM:SIN:FEM", "Gemeinde"),
			atrWithPOS(".", "PKT", "."),
		}},
		// Ich geht jetzt nach Hause
		{"Ich geht jetzt", []*languagetool.AnalyzedTokenReadings{
			atrWithPOS("Ich", "PRO:PER:NOM:SIN:ALG", "ich"),
			atrWithPOS("geht", "VER:3:SIN:PRÄ:SFT", "gehen"),
			atrWithPOS("jetzt", "ADV", "jetzt"),
			atrWithPOS("nach", "APPR", "nach"),
			atrWithPOS("Hause", "SUB:DAT:SIN:NEU", "Haus"),
			atrWithPOS(".", "PKT", "."),
		}},
		// Ich setzet mich …
		{"Ich setzet mich", []*languagetool.AnalyzedTokenReadings{
			atrWithPOS("Ich", "PRO:PER:NOM:SIN:ALG", "ich"),
			atrWithPOS("setzet", "VER:2:PLU:PRÄ:SFT", "setzen"),
			atrWithPOS("mich", "PRO:PER:AKK:SIN:1", "ich"),
			atrWithPOS("auf", "APPR", "auf"),
			atrWithPOS("den", "ART:DEF:AKK:SIN:MAS", "der"),
			atrWithPOS("Teppich", "SUB:AKK:SIN:MAS", "Teppich"),
			atrWithPOS(".", "PKT", "."),
		}},
		// Ich haben meinen Ohrring …
		{"Ich haben meinen", []*languagetool.AnalyzedTokenReadings{
			atrWithPOS("Ich", "PRO:PER:NOM:SIN:ALG", "ich"),
			atrWithPOS("haben", "VER:1:PLU:PRÄ:NON", "haben"),
			atrWithPOS("meinen", "PRO:POS:AKK:SIN:MAS", "mein"),
			atrWithPOS("Ohrring", "SUB:AKK:SIN:MAS", "Ohrring"),
			atrWithPOS("fallen", "VER:INF:NON", "fallen"),
			atrWithPOS("lassen", "VER:INF:NON", "lassen"),
			atrWithPOS(".", "PKT", "."),
		}},
		// Ich stehen Ihnen gerne …
		{"Ich stehen Ihnen", []*languagetool.AnalyzedTokenReadings{
			atrWithPOS("Ich", "PRO:PER:NOM:SIN:ALG", "ich"),
			atrWithPOS("stehen", "VER:1:PLU:PRÄ:NON", "stehen"),
			atrWithPOS("Ihnen", "PRO:PER:DAT:PLU:2", "Sie"),
			atrWithPOS("gerne", "ADV", "gerne"),
			atrWithPOS("zur", "APPRART:DAT:SIN:FEM", "zu"),
			atrWithPOS("Verfügung", "SUB:DAT:SIN:FEM", "Verfügung"),
			atrWithPOS(".", "PKT", "."),
		}},
		// Das greift … bist auf die Zeit …
		{"bist auf die Zeit", []*languagetool.AnalyzedTokenReadings{
			atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "das"),
			atrWithPOS("greift", "VER:3:SIN:PRÄ:SFT", "greifen"),
			atrWithPOS("auf", "APPR", "auf"),
			atrWithPOS("Vorläuferinstitutionen", "SUB:AKK:PLU:FEM", "Vorläuferinstitution"),
			atrWithPOS("bist", "VER:2:SIN:PRÄ:NON", "sein"),
			atrWithPOS("auf", "APPR", "auf"),
			atrWithPOS("die", "ART:DEF:AKK:SIN:FEM", "die"),
			atrWithPOS("Zeit", "SUB:AKK:SIN:FEM", "Zeit"),
			atrWithPOS(".", "PKT", "."),
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			toks := append([]*languagetool.AnalyzedTokenReadings{sentStartATR()}, tc.parts...)
			require.GreaterOrEqual(t, len(rule.Match(languagetool.NewAnalyzedSentence(withPositions(toks...)))), 1)
		})
	}

	// Java goods that must not invent (morph)
	good := func(label string, parts ...*languagetool.AnalyzedTokenReadings) {
		t.Helper()
		toks := append([]*languagetool.AnalyzedTokenReadings{sentStartATR()}, parts...)
		require.Equal(t, 0, len(rule.Match(languagetool.NewAnalyzedSentence(withPositions(toks...)))), label)
	}
	// Weder Peter noch ich wollen das.
	// Morphy: wollen is 1:PLU and 3:PLU — not unambiguous 1:PLU (would false-fire without "wir").
	good("Weder … noch ich wollen",
		atrWithPOS("Weder", "KON", "weder"),
		atrWithPOS("Peter", "EIG:NOM:SIN:MAS", "Peter"),
		atrWithPOS("noch", "KON", "noch"),
		atrWithPOS("ich", "PRO:PER:NOM:SIN:ALG", "ich"),
		atrWithPOSMulti("wollen", "wollen", "VER:1:PLU:PRÄ:NON", "VER:3:PLU:PRÄ:NON"),
		atrWithPOS("das", "PDS:AKK:SIN:NEU", "das"),
		atrWithPOS(".", "PKT", "."),
	)
	// Max und ich sollten das machen. — sollten KJ2 1:PLU + 3:PLU
	good("Max und ich sollten",
		atrWithPOS("Max", "EIG:NOM:SIN:MAS", "Max"),
		atrWithPOS("und", "KON:NEB", "und"),
		atrWithPOS("ich", "PRO:PER:NOM:SIN:ALG", "ich"),
		atrWithPOSMulti("sollten", "sollen", "VER:1:PLU:KJ2:NON", "VER:3:PLU:KJ2:NON"),
		atrWithPOS("das", "PDS:AKK:SIN:NEU", "das"),
		atrWithPOS("machen", "VER:INF:NON", "machen"),
		atrWithPOS(".", "PKT", "."),
	)
	// Bin gleich wieder da. (imperative-like / omitted subject)
	good("Bin gleich wieder da",
		atrWithPOS("Bin", "VER:1:SIN:PRÄ:NON", "sein"),
		atrWithPOS("gleich", "ADV", "gleich"),
		atrWithPOS("wieder", "ADV", "wieder"),
		atrWithPOS("da", "ADV", "da"),
		atrWithPOS(".", "PKT", "."),
	)
	// Die Jagd nach bin Laden.
	good("Die Jagd nach bin Laden",
		atrWithPOS("Die", "ART:DEF:NOM:SIN:FEM", "die"),
		atrWithPOS("Jagd", "SUB:NOM:SIN:FEM", "Jagd"),
		atrWithPOS("nach", "APPR", "nach"),
		atrWithPOS("bin", "VER:1:SIN:PRÄ:NON", "sein"),
		atrWithPOS("Laden", "EIG:DAT:SIN:MAS", "Laden"),
		atrWithPOS(".", "PKT", "."),
	)
	// /usr/bin/firefox — path, bin not finite subject verb
	good("/usr/bin/firefox",
		atrWithPOS("/usr/bin/firefox", "XY", "/usr/bin/firefox"),
	)
	// soft hyphen surface: so tes\u00ADtest Du das
	good("soft hyphen testest Du",
		atrWithPOS("So", "ADV", "so"),
		atrWithPOS("tes\u00ADtest", "VER:2:SIN:PRÄ:SFT", "testen"),
		atrWithPOS("Du", "PRO:PER:NOM:SIN:ALG", "du"),
		atrWithPOS("das", "PDS:AKK:SIN:NEU", "das"),
		atrWithPOS(".", "PKT", "."),
	)

	// untagged must not invent
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Ich sind müde."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Peter bin nett."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Du muss gehen."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Ich geht jetzt nach Hause."))))
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

	// good — plural finite forms use Morphy-like 1:PLU+3:PLU ambiguity
	lebenPlu := func() *languagetool.AnalyzedTokenReadings {
		return atrWithPOSMulti("leben", "leben", "VER:1:PLU:PRÄ:SFT", "VER:3:PLU:PRÄ:SFT")
	}
	good(atrWithPOS("Du", "PRO:PER:NOM:SIN:ALG", "du"), atrWithPOS("lebst", "VER:2:SIN:PRÄ:SFT", "leben"), atrWithPOS(".", "PKT", "."))
	good(atrWithPOS("Wir", "PRO:PER:NOM:PLU:ALG", "wir"), lebenPlu(), atrWithPOS("noch", "ADV", "noch"), atrWithPOS(".", "PKT", "."))
	good(atrWithPOS("Ich", "PRO:PER:NOM:SIN:ALG", "ich"), atrWithPOS("bin", "VER:1:SIN:PRÄ:NON", "sein"), atrWithPOS("nett", "ADJ:PRD:GRU", "nett"), atrWithPOS(".", "PKT", "."))
	good(atrWithPOS("Er", "PRO:PER:NOM:SIN:MAS", "er"), atrWithPOS("lebt", "VER:3:SIN:PRÄ:SFT", "leben"), atrWithPOS(".", "PKT", "."))
	good(atrWithPOS("Nett", "ADJ:PRD:GRU", "nett"), atrWithPOS("bin", "VER:1:SIN:PRÄ:NON", "sein"), atrWithPOS("ich", "PRO:PER:NOM:SIN:ALG", "ich"), atrWithPOS(".", "PKT", "."))
	good(atrWithPOS("Nett", "ADJ:PRD:GRU", "nett"), atrWithPOS("bist", "VER:2:SIN:PRÄ:NON", "sein"), atrWithPOS("du", "PRO:PER:NOM:SIN:ALG", "du"), atrWithPOS(".", "PKT", "."))

	// bad (testWrongVerbSubject)
	badAtLeast(1, atrWithPOS("Du", "PRO:PER:NOM:SIN:ALG", "du"), atrWithPOS("leben", "VER:1:PLU:PRÄ:SFT", "leben"), atrWithPOS(".", "PKT", "."))
	badAtLeast(1, atrWithPOS("Ich", "PRO:PER:NOM:SIN:ALG", "ich"), atrWithPOS("leben", "VER:1:PLU:PRÄ:SFT", "leben"), atrWithPOS(".", "PKT", "."))
	badAtLeast(1, atrWithPOS("Auch", "ADV", "auch"), atrWithPOS("morgen", "ADV", "morgen"), atrWithPOS("leben", "VER:1:PLU:PRÄ:SFT", "leben"), atrWithPOS("du", "PRO:PER:NOM:SIN:ALG", "du"), atrWithPOS(".", "PKT", "."))
	// last token is "du" — no segfault / still bad
	badAtLeast(1, atrWithPOS("Auch", "ADV", "auch"), atrWithPOS("morgen", "ADV", "morgen"), atrWithPOS("leben", "VER:1:PLU:PRÄ:SFT", "leben"), atrWithPOS("du", "PRO:PER:NOM:SIN:ALG", "du"))
	badAtLeast(1, atrWithPOS("Auch", "ADV", "auch"), atrWithPOS("morgen", "ADV", "morgen"), atrWithPOS("leben", "VER:1:PLU:PRÄ:SFT", "leben"), atrWithPOS("er", "PRO:PER:NOM:SIN:MAS", "er"), atrWithPOS(".", "PKT", "."))
	badAtLeast(1, atrWithPOS("Auch", "ADV", "auch"), atrWithPOS("morgen", "ADV", "morgen"), atrWithPOS("leben", "VER:1:PLU:PRÄ:SFT", "leben"), atrWithPOS("ich", "PRO:PER:NOM:SIN:ALG", "ich"), atrWithPOS(".", "PKT", "."))
	badAtLeast(1, atrWithPOS("Auch", "ADV", "auch"), atrWithPOS("morgen", "ADV", "morgen"), atrWithPOS("lebte", "VER:3:SIN:PRT:SFT", "leben"), atrWithPOS("wir", "PRO:PER:NOM:PLU:ALG", "wir"), atrWithPOS("noch", "ADV", "noch"), atrWithPOS(".", "PKT", "."))
	// "Du bin nett." — Java expects 2
	badN(2, atrWithPOS("Du", "PRO:PER:NOM:SIN:ALG", "du"), atrWithPOS("bin", "VER:1:SIN:PRÄ:NON", "sein"), atrWithPOS("nett", "ADJ:PRD:GRU", "nett"), atrWithPOS(".", "PKT", "."))
	// "Ich bist nett." — Java expects 2
	badN(2, atrWithPOS("Ich", "PRO:PER:NOM:SIN:ALG", "ich"), atrWithPOS("bist", "VER:2:SIN:PRÄ:NON", "sein"), atrWithPOS("nett", "ADJ:PRD:GRU", "nett"), atrWithPOS(".", "PKT", "."))
	// "Er lebst." — Java expects 2
	badN(2, atrWithPOS("Er", "PRO:PER:NOM:SIN:MAS", "er"), atrWithPOS("lebst", "VER:2:SIN:PRÄ:SFT", "leben"), atrWithPOS(".", "PKT", "."))
	// "Wir lebst noch." — Java expects 2
	badN(2, atrWithPOS("Wir", "PRO:PER:NOM:PLU:ALG", "wir"), atrWithPOS("lebst", "VER:2:SIN:PRÄ:SFT", "leben"), atrWithPOS("noch", "ADV", "noch"), atrWithPOS(".", "PKT", "."))
	// "Er bin nett." — Java 2
	badN(2, atrWithPOS("Er", "PRO:PER:NOM:SIN:MAS", "er"), atrWithPOS("bin", "VER:1:SIN:PRÄ:NON", "sein"), atrWithPOS("nett", "ADJ:PRD:GRU", "nett"), atrWithPOS(".", "PKT", "."))
	// "Wir bin nett." — Java 2
	badN(2, atrWithPOS("Wir", "PRO:PER:NOM:PLU:ALG", "wir"), atrWithPOS("bin", "VER:1:SIN:PRÄ:NON", "sein"), atrWithPOS("nett", "ADJ:PRD:GRU", "nett"), atrWithPOS(".", "PKT", "."))
	// "Nett bist ich nicht." — Java 2
	badN(2, atrWithPOS("Nett", "ADJ:PRD:GRU", "nett"), atrWithPOS("bist", "VER:2:SIN:PRÄ:NON", "sein"), atrWithPOS("ich", "PRO:PER:NOM:SIN:ALG", "ich"), atrWithPOS("nicht", "ADV", "nicht"), atrWithPOS(".", "PKT", "."))
	// "Nett sind du."
	badAtLeast(1, atrWithPOS("Nett", "ADJ:PRD:GRU", "nett"), atrWithPOS("sind", "VER:1:PLU:PRÄ:NON", "sein"), atrWithPOS("du", "PRO:PER:NOM:SIN:ALG", "du"), atrWithPOS(".", "PKT", "."))
	// "Du wünscht dir so viel." — 3:SIN wünscht vs du 2:SIN
	badAtLeast(1, atrWithPOS("Du", "PRO:PER:NOM:SIN:ALG", "du"), atrWithPOS("wünscht", "VER:3:SIN:PRÄ:SFT", "wünschen"), atrWithPOS("dir", "PRO:PER:DAT:SIN:2", "du"), atrWithPOS("so", "ADV", "so"), atrWithPOS("viel", "PIS", "viel"), atrWithPOS(".", "PKT", "."))
	// "Wünscht du dir mehr Zeit?"
	badAtLeast(1, atrWithPOS("Wünscht", "VER:3:SIN:PRÄ:SFT", "wünschen"), atrWithPOS("du", "PRO:PER:NOM:SIN:ALG", "du"), atrWithPOS("dir", "PRO:PER:DAT:SIN:2", "du"), atrWithPOS("mehr", "PIAT", "mehr"), atrWithPOS("Zeit", "SUB:AKK:SIN:FEM", "Zeit"), atrWithPOS("?", "PKT", "?"))
	// "Lebe du?"
	badAtLeast(1, atrWithPOS("Lebe", "VER:1:SIN:KJ1:SFT", "leben"), atrWithPOS("du", "PRO:PER:NOM:SIN:ALG", "du"), atrWithPOS("?", "PKT", "?"))
	// "Leben du?"
	badAtLeast(1, atrWithPOS("Leben", "VER:1:PLU:PRÄ:SFT", "leben"), atrWithPOS("du", "PRO:PER:NOM:SIN:ALG", "du"), atrWithPOS("?", "PKT", "?"))
	// "Du können heute leider nicht kommen."
	badAtLeast(1, atrWithPOS("Du", "PRO:PER:NOM:SIN:ALG", "du"), atrWithPOS("können", "VER:1:PLU:PRÄ:NON", "können"), atrWithPOS("heute", "ADV", "heute"), atrWithPOS("leider", "ADV", "leider"), atrWithPOS("nicht", "ADV", "nicht"), atrWithPOS("kommen", "VER:INF:NON", "kommen"), atrWithPOS(".", "PKT", "."))
	// "Ich kannst heute leider nicht kommen." — Java 2
	badN(2, atrWithPOS("Ich", "PRO:PER:NOM:SIN:ALG", "ich"), atrWithPOS("kannst", "VER:2:SIN:PRÄ:NON", "können"), atrWithPOS("heute", "ADV", "heute"), atrWithPOS("leider", "ADV", "leider"), atrWithPOS("nicht", "ADV", "nicht"), atrWithPOS("kommen", "VER:INF:NON", "kommen"), atrWithPOS(".", "PKT", "."))
	// "Wir könnt heute leider nicht kommen."
	badAtLeast(1, atrWithPOS("Wir", "PRO:PER:NOM:PLU:ALG", "wir"), atrWithPOS("könnt", "VER:2:PLU:PRÄ:NON", "können"), atrWithPOS("heute", "ADV", "heute"), atrWithPOS("leider", "ADV", "leider"), atrWithPOS("nicht", "ADV", "nicht"), atrWithPOS("kommen", "VER:INF:NON", "kommen"), atrWithPOS(".", "PKT", "."))
	// "Nett warst wir." — Java 2
	badN(2, atrWithPOS("Nett", "ADJ:PRD:GRU", "nett"), atrWithPOS("warst", "VER:2:SIN:PRT:NON", "sein"), atrWithPOS("wir", "PRO:PER:NOM:PLU:ALG", "wir"), atrWithPOS(".", "PKT", "."))

	// more goods from Java subject list
	good(atrWithPOS("Auch", "ADV", "auch"), atrWithPOS("morgen", "ADV", "morgen"), atrWithPOS("lebe", "VER:1:SIN:PRÄ:SFT", "leben"), atrWithPOS("ich", "PRO:PER:NOM:SIN:ALG", "ich"), atrWithPOS(".", "PKT", "."))
	good(atrWithPOS("Auch", "ADV", "auch"), atrWithPOS("morgen", "ADV", "morgen"), atrWithPOS("lebst", "VER:2:SIN:PRÄ:SFT", "leben"), atrWithPOS("du", "PRO:PER:NOM:SIN:ALG", "du"), atrWithPOS(".", "PKT", "."))
	good(atrWithPOS("Auch", "ADV", "auch"), atrWithPOS("morgen", "ADV", "morgen"), atrWithPOS("lebt", "VER:3:SIN:PRÄ:SFT", "leben"), atrWithPOS("er", "PRO:PER:NOM:SIN:MAS", "er"), atrWithPOS(".", "PKT", "."))
	good(atrWithPOS("Auch", "ADV", "auch"), atrWithPOS("morgen", "ADV", "morgen"), lebenPlu(), atrWithPOS("wir", "PRO:PER:NOM:PLU:ALG", "wir"), atrWithPOS("noch", "ADV", "noch"), atrWithPOS(".", "PKT", "."))
	// Er und du/ich leben — coordinated subject; verb not unambiguous 1:PLU (Morphy 1+3 PLU)
	good(atrWithPOS("Er", "PRO:PER:NOM:SIN:MAS", "er"), atrWithPOS("und", "KON:NEB", "und"), atrWithPOS("du", "PRO:PER:NOM:SIN:ALG", "du"), lebenPlu(), atrWithPOS(".", "PKT", "."))
	good(atrWithPOS("Er", "PRO:PER:NOM:SIN:MAS", "er"), atrWithPOS("und", "KON:NEB", "und"), atrWithPOS("ich", "PRO:PER:NOM:SIN:ALG", "ich"), lebenPlu(), atrWithPOS(".", "PKT", "."))
	good(atrWithPOS("Nett", "ADJ:PRD:GRU", "nett"), atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"), atrWithPOS("er", "PRO:PER:NOM:SIN:MAS", "er"), atrWithPOS(".", "PKT", "."))
	// sind is 1:PLU and 3:PLU
	good(atrWithPOS("Nett", "ADJ:PRD:GRU", "nett"), atrWithPOSMulti("sind", "sein", "VER:1:PLU:PRÄ:NON", "VER:3:PLU:PRÄ:NON"), atrWithPOS("wir", "PRO:PER:NOM:PLU:ALG", "wir"), atrWithPOS(".", "PKT", "."))
	// Das lyrische Ich ist verzweifelt.
	good(
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("lyrische", "ADJ:NOM:SIN:NEU:GRU:DEF", "lyrisch"),
		atrWithPOS("Ich", "PRO:PER:NOM:SIN:ALG", "ich"),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("verzweifelt", "ADJ:PRD:GRU", "verzweifelt"),
		atrWithPOS(".", "PKT", "."),
	)
	// Wenn ich du wäre … (anti-pattern + KJ2 gates)
	good(
		atrWithPOS("Wenn", "KOUS", "wenn"),
		atrWithPOS("ich", "PRO:PER:NOM:SIN:ALG", "ich"),
		atrWithPOS("du", "PRO:PER:NOM:SIN:ALG", "du"),
		atrWithPOSMulti("wäre", "sein", "VER:1:SIN:KJ2:NON", "VER:3:SIN:KJ2:NON"),
		atrWithPOS(",", "PKT", ","),
		atrWithPOS("würde", "VER:1:SIN:KJ2:NON", "werden"),
		atrWithPOS("ich", "PRO:PER:NOM:SIN:ALG", "ich"),
		atrWithPOS("das", "PDS:AKK:SIN:NEU", "das"),
		atrWithPOS("nicht", "ADV", "nicht"),
		atrWithPOS("machen", "VER:INF:NON", "machen"),
		atrWithPOS(".", "PKT", "."),
	)

	// untagged must not invent
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Auch morgen leben du."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Du leben."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Ich bist nett."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Nett warst wir."))))
}
