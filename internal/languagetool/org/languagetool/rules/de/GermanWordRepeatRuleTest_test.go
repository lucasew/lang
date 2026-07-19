package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/GermanWordRepeatRuleTest.java
// Java uses JLanguageTool.getAnalyzedSentence (tagged). Token-only anti-patterns work with
// AnalyzePlain; POS / matchInflectedForms cases inject tags (no ignore-surface invent).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestGermanWordRepeatRule_Rule(t *testing.T) {
	rule := NewGermanWordRepeatRule(map[string]string{"repetition": "Word repetition"})
	assertN := func(s string, n int) {
		t.Helper()
		got := len(rule.Match(languagetool.AnalyzePlain(s)))
		require.Equal(t, n, got, "sentence %q", s)
	}
	assertSent := func(sent *languagetool.AnalyzedSentence, n int, label string) {
		t.Helper()
		got := len(rule.Match(sent))
		require.Equal(t, n, got, "sentence %q", label)
	}

	assertN("Das ist gut so.", 0)
	assertN("Das ist ist gut so.", 1)
	assertN("Der der Mann", 1)
	assertN("Warum fragen Sie sie nicht selbst?", 0)
	assertN("Er will nur sein Leben leben.", 0)

	// ", die die" — matchInflectedForms(csToken "der") needs lemma "der"
	assertSent(languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Wie", "ADV", "wie"),
		atrWithPOS("bei", "PRP:DAT", "bei"),
		atrWithPOS("Honda", "EIG:DAT:SIN:NEU", "Honda"),
		atrWithPOS(",", "PKT", ","),
		atrWithPOS("die", "ART:DEF:NOM:PLU:MAS", "der"),
		atrWithPOS("die", "ART:DEF:AKK:SIN:FEM", "der"),
		atrWithPOS("Bezahlung", "SUB:AKK:SIN:FEM", "Bezahlung"),
		atrWithPOS("erhöht", "VER:3:SIN:PRS:NON:NEB", "erhöhen"),
		atrWithPOS("haben", "VER:AUX:3:PLU:PRS:SFT", "haben"),
		atrWithPOS(".", "PKT", "."),
	)), 0, "Wie bei Honda, die die Bezahlung…")

	// warfen sie sie weg — VER:3 + ZUS
	assertSent(languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Dann", "ADV", "dann"),
		atrWithPOS("warfen", "VER:3:PLU:PRT:NON", "werfen"),
		atrWithPOS("sie", "PRO:PER:NOM:PLU:MAS", "sie"),
		atrWithPOS("sie", "PRO:PER:AKK:PLU:MAS", "sie"),
		atrWithPOS("weg", "ZUS", "weg"),
		atrWithPOS(".", "PKT", "."),
	)), 0, "Dann warfen sie sie weg.")

	// konnte sie sie sehen — VER:MOD:3 + exact VER:INF:NON (Java hasPosTag)
	assertSent(languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Dann", "ADV", "dann"),
		atrWithPOS("konnte", "VER:MOD:3:SIN:PRT:SFT", "können"),
		atrWithPOS("sie", "PRO:PER:NOM:SIN:FEM", "sie"),
		atrWithPOS("sie", "PRO:PER:AKK:SIN:FEM", "sie"),
		atrWithPOS("sehen", "VER:INF:NON", "sehen"),
		atrWithPOS(".", "PKT", "."),
	)), 0, "Dann konnte sie sie sehen.")

	// Java requires hasPosTag("VER:INF:NON") — other INF tags must not invent an ignore.
	assertSent(languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Dann", "ADV", "dann"),
		atrWithPOS("konnte", "VER:MOD:3:SIN:PRT:SFT", "können"),
		atrWithPOS("sie", "PRO:PER:NOM:SIN:FEM", "sie"),
		atrWithPOS("sie", "PRO:PER:AKK:SIN:FEM", "sie"),
		atrWithPOS("sehen", "VER:INF:SFT", "sehen"),
		atrWithPOS(".", "PKT", "."),
	)), 1, "VER:INF:SFT must not use VER:INF:NON ignore gate")

	assertN("Er muss sein Essen essen.", 0)

	// ist das das Problem — POS SUB:NEU on Problem
	assertSent(languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Wahrscheinlich", "ADV", "wahrscheinlich"),
		atrWithPOS("ist", "VER:3:SIN:PRS:SFT", "sein"),
		atrWithPOS("das", "PRO:DEM:NOM:SIN:NEU", "das"),
		atrWithPOS("das", "ART:DEF:NOM:SIN:NEU", "der"),
		atrWithPOS("Problem", "SUB:NOM:SIN:NEU", "Problem"),
		atrWithPOS(".", "PKT", "."),
	)), 0, "Wahrscheinlich ist das das Problem.")

	// wäre das das erste … Wirtschaftsmagazin
	assertSent(languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Dann", "ADV", "dann"),
		atrWithPOS("wäre", "VER:3:SIN:KJ1:SFT", "sein"),
		atrWithPOS("das", "PRO:DEM:NOM:SIN:NEU", "das"),
		atrWithPOS("das", "ART:DEF:NOM:SIN:NEU", "der"),
		atrWithPOS("erste", "ADJ:NOM:SIN:NEU:GRU:IND", "erst"),
		atrWithPOS("Wirtschaftsmagazin", "SUB:NOM:SIN:NEU", "Wirtschaftsmagazin"),
		atrWithPOS(".", "PKT", "."),
	)), 0, "Dann wäre das das erste Wirtschaftsmagazin…")

	// war das das Härteste — war + das + das + SUB:NEU
	assertSent(languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Für", "PRP:AKK", "für"),
		atrWithPOS("mich", "PRO:PER:AKK:SIN:MAS", "ich"),
		atrWithPOS("war", "VER:3:SIN:PRT:SFT", "sein"),
		atrWithPOS("das", "PRO:DEM:NOM:SIN:NEU", "das"),
		atrWithPOS("das", "ART:DEF:NOM:SIN:NEU", "der"),
		atrWithPOS("Härteste", "SUB:NOM:SIN:NEU", "Härteste"),
		atrWithPOS(".", "PKT", "."),
	)), 0, "Für mich war das das Härteste.")

	// token-only anti-pattern: wer wer
	assertN("Ich weiß, wer wer ist!", 0)

	// falls das das Problem ist — SUB + matchInflected sein|haben
	assertSent(languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Ich", "PRO:PER:NOM:SIN:MAS", "ich"),
		atrWithPOS("kann", "VER:MOD:1:SIN:PRS:SFT", "können"),
		atrWithPOS("das", "PRO:DEM:AKK:SIN:NEU", "das"),
		atrWithPOS("machen", "VER:INF:NON", "machen"),
		atrWithPOS(",", "PKT", ","),
		atrWithPOS("falls", "KON:UNT", "falls"),
		atrWithPOS("das", "PRO:DEM:NOM:SIN:NEU", "das"),
		atrWithPOS("das", "ART:DEF:NOM:SIN:NEU", "der"),
		atrWithPOS("Problem", "SUB:NOM:SIN:NEU", "Problem"),
		atrWithPOS("ist", "VER:3:SIN:PRS:SFT", "sein"),
		atrWithPOS(".", "PKT", "."),
	)), 0, "falls das das Problem ist")

	// Als ich das das erste Mal — PRO + ADJ:NEU + SUB:NEU
	assertSent(languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Als", "KON:UNT", "als"),
		atrWithPOS("ich", "PRO:PER:NOM:SIN:MAS", "ich"),
		atrWithPOS("das", "PRO:DEM:AKK:SIN:NEU", "das"),
		atrWithPOS("das", "ART:DEF:AKK:SIN:NEU", "der"),
		atrWithPOS("erste", "ADJ:AKK:SIN:NEU:GRU:IND", "erst"),
		atrWithPOS("Mal", "SUB:AKK:SIN:NEU", "Mal"),
		atrWithPOS("gehört", "VER:PA2:SFT", "hören"),
		atrWithPOS("habe", "VER:AUX:1:SIN:PRS:SFT", "haben"),
	)), 0, "Als ich das das erste Mal…")

	assertN("Hat sie sie", 1)
	assertN("Hat hat", 1)
	assertN("hat hat", 1)
	// anti-pattern token pairs
	assertN("Moin Moin", 0)
	assertN("Bora Bora", 0)
	assertN("Man kann nicht nicht kommunizieren.", 0)
	assertN("Ich mag mag das.", 1)
}
