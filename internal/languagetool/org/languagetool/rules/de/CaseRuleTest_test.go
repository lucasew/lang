package de

// Twin of CaseRuleTest — Java uses POS + tagger lookup (no surface invent).
import (
	"regexp"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestCaseRule_Rule(t *testing.T) {
	rule := NewCaseRule(nil)
	// untagged AnalyzePlain must not invent capitalization hits
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Und das Neue Haus."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Ich habe Heute keine Zeit."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Ein einfacher Satz zum Testen."))))

	// Morph: "das" + lowercase adjective reading as noun error is complex (needs Lookup hook).
	// Smoke: matchMorph runs without panic on tagged sentence start.
	sent := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Ein", "ART:IND:NOM:SIN:MAS", "ein"),
		atrWithPOS("einfacher", "ADJ:NOM:SIN:MAS:GRU:IND", "einfach"),
		atrWithPOS("Satz", "SUB:NOM:SIN:MAS", "Satz"),
		atrWithPOS(".", "PKT", "."),
	))
	require.NotPanics(t, func() { _ = rule.Match(sent) })
}

func TestCaseRuleCompareListsLegacy(t *testing.T) {
	s := languagetool.AnalyzePlain("Hier ein Test")
	toks := s.GetTokensWithoutWhitespace()
	// tokens: "", Hier, ein, Test, .
	require.True(t, CaseRuleCompareLists(toks, 1, 2,
		[]*regexp.Regexp{regexp.MustCompile("Hier"), regexp.MustCompile("ein")}))
	require.False(t, CaseRuleCompareLists(toks, 1, 2,
		[]*regexp.Regexp{regexp.MustCompile("Hier"), regexp.MustCompile("Test")}))
}

func TestCaseRule_EstimateContextAndURL(t *testing.T) {
	rule := NewCaseRule(nil)
	// Java: ANTI_PATTERNS.stream().mapToInt(List::size).max().orElse(0)
	require.Greater(t, rule.EstimateContextForSureMatch(), 0)
	require.Equal(t, "https://dict.leo.org/grammatik/deutsch/Rechtschreibung/Regeln/Gross-klein/index.html", rule.GetURL())
}

// Java: (prevTokenIsDas && …) || (i>1 && VER:AUX|MOD at i-2) continues the whole
// iteration — uppercase match is skipped even when prev is not "das".
// "Wird man Gehen nach …" with Gehen as VER:INF baseform must not flag DE_CASE.
func TestCaseRule_AuxModSkipIndependentOfDas(t *testing.T) {
	rule := NewCaseRule(nil)
	rule.IsMisspelled = func(string) bool { return false }
	toks := withPositions(
		sentStartATR(),
		atrWithPOS("Wird", "VER:AUX:3:SIN:PRÄ:SFT", "werden"),
		atrWithPOS("man", "PRO:PER:NOM:SIN:3:MAS", "man"),
		atrWithPOS("Gehen", "VER:INF:NON", "Gehen"),
		atrWithPOS("weiter", "ADV", "weiter"),
		atrWithPOS(".", "PKT", "."),
	)
	sent := languagetool.NewAnalyzedSentence(toks)
	// matchMorph (no anti-pattern immunization) so we isolate the Java operator-precedence continue.
	require.Equal(t, 0, len(rule.matchMorph(toks, sent)), "VER:AUX at i-2 must skip uppercase path (Java precedence)")
}

// Twin of CaseRuleTest.testRuleActivation
func TestCaseRule_RuleActivation(t *testing.T) {
	rule := NewCaseRule(nil)
	require.True(t, rule.SupportsLanguage("de-DE"))
	require.True(t, rule.SupportsLanguage("de"))
	require.True(t, rule.SupportsLanguage("de-AT"))
	require.False(t, rule.SupportsLanguage("en"))
	require.False(t, rule.SupportsLanguage("en-US"))
}

// Twin of CaseRuleTest.testCompareLists (Java indices include SENT_START at 0)
func TestCaseRule_CompareLists(t *testing.T) {
	// AnalyzePlain may not insert empty SENT_START; match Java by building tokens like LT.
	ss := languagetool.SentenceStartTagName
	toks := withPositions(
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		atrWithPOS("Hier", "ADV", "hier"),
		atrWithPOS("ein", "ART:IND:NOM:SIN:MAS", "ein"),
		atrWithPOS("Test", "SUB:NOM:SIN:MAS", "Test"),
	)
	// Java: start 0 end 2 with "", Hier, ein
	require.True(t, CaseRuleCompareLists(toks, 0, 2,
		[]*regexp.Regexp{regexp.MustCompile(""), regexp.MustCompile("Hier"), regexp.MustCompile("ein")}))
	require.True(t, CaseRuleCompareLists(toks, 1, 2,
		[]*regexp.Regexp{regexp.MustCompile("Hier"), regexp.MustCompile("ein")}))
	require.True(t, CaseRuleCompareLists(toks, 0, 3,
		[]*regexp.Regexp{regexp.MustCompile(""), regexp.MustCompile("Hier"), regexp.MustCompile("ein"), regexp.MustCompile("Test")}))
	require.False(t, CaseRuleCompareLists(toks, 0, 4,
		[]*regexp.Regexp{regexp.MustCompile(""), regexp.MustCompile("Hier"), regexp.MustCompile("ein"), regexp.MustCompile("Test")}))

	toks2 := withPositions(
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		atrWithPOS("das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("Heilige", "ADJ:NOM:SIN:NEU:GRU:DEF", "heilig"),
		atrWithPOS("Römische", "ADJ:NOM:SIN:NEU:GRU:DEF", "römisch"),
		atrWithPOS("Reich", "SUB:NOM:SIN:NEU", "Reich"),
	)
	require.True(t, CaseRuleCompareLists(toks2, 0, 4,
		[]*regexp.Regexp{regexp.MustCompile(""), regexp.MustCompile("das"), regexp.MustCompile("Heilige"), regexp.MustCompile("Römische"), regexp.MustCompile("Reich")}))
	require.False(t, CaseRuleCompareLists(toks2, 8, 11,
		[]*regexp.Regexp{regexp.MustCompile(""), regexp.MustCompile("das"), regexp.MustCompile("Heilige"), regexp.MustCompile("Römische"), regexp.MustCompile("Reich")}))
}

// Twin of CaseRuleTest.testPhraseExceptions (partial phrase not error)
func TestCaseRule_PhraseExceptions(t *testing.T) {
	rule := NewCaseRule(nil)
	// Partial "ohne wenn" is not a complete exception phrase → still good (no invent hit)
	// Full exception path needs morph: "ohne Wenn und Aber"
	// Good: complete phrase with correct casing
	good := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("gilt", "VER:3:SIN:PRÄ:SFT", "gelten"),
		atrWithPOS("ohne", "APPR", "ohne"),
		atrWithPOS("Wenn", "SUB:NOM:SIN:NEU", "Wenn"),
		atrWithPOS("und", "KON:NEB", "und"),
		atrWithPOS("Aber", "SUB:NOM:SIN:NEU", "Aber"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 0, len(rule.Match(good)))
	// Incomplete phrase "ohne wenn" alone — no DE_CASE invent without noun morph error
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Das gilt ohne wenn"))))
}

// Twin of CaseRuleTest.testSubstantivierteVerben (morph: Das + lower VER as noun error)
func TestCaseRule_SubstantivierteVerben(t *testing.T) {
	rule := NewCaseRule(nil)
	// Lookup returns VER for "fahren" so "Das fahren" is potential lowercased nominalization error
	rule.Lookup = func(word string) *languagetool.AnalyzedTokenReadings {
		w := word
		if word == "fahren" || word == "Fahren" {
			return languagetool.NewAnalyzedTokenReadingsAt(
				languagetool.NewAnalyzedToken(w, strPtr("VER:INF:NON"), strPtr("fahren")), 0)
		}
		if word == "laufen" || word == "Laufen" {
			return languagetool.NewAnalyzedTokenReadingsAt(
				languagetool.NewAnalyzedToken(w, strPtr("VER:INF:NON"), strPtr("laufen")), 0)
		}
		return nil
	}
	// Good: capitalized substantivized "Fahren"
	good := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("Fahren", "SUB:NOM:SIN:NEU:INF", "Fahren"),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("einfach", "ADJ:PRD:GRU", "einfach"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 0, len(rule.Match(good)))

	// Bad: "Das fahren" — lowercased after das (Java assertBad)
	bad := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("fahren", "VER:INF:NON", "fahren"),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("einfach", "ADJ:PRD:GRU", "einfach"),
		atrWithPOS(".", "PKT", "."),
	))
	ms := rule.Match(bad)
	require.Equal(t, 1, len(ms), "Das fahren → lowercase nominalization error")
	// untagged must not invent
	require.Equal(t, 0, len(NewCaseRule(nil).Match(languagetool.AnalyzePlain("Das fahren ist einfach."))))
}

// caseRuleTwinLookup returns morph dict twins for CaseRuleTest surfaces
// (Java GermanTagger.lookup when german.dict is present).
func caseRuleTwinLookup(word string) *languagetool.AnalyzedTokenReadings {
	// map surface → POS tags (non-ADJ SUB for nouns; VER for infinitives used in lowercase path)
	type tags []string
	m := map[string]tags{
		// nouns / substantivized (hasNounReading true)
		"Haus": {"SUB:NOM:SIN:NEU"}, "haus": {"SUB:NOM:SIN:NEU"},
		"Zeit": {"SUB:AKK:SIN:FEM"}, "zeit": {"SUB:AKK:SIN:FEM"},
		"Satz": {"SUB:NOM:SIN:MAS"}, "satz": {"SUB:NOM:SIN:MAS"},
		"Auto": {"SUB:NOM:SIN:NEU"}, "auto": {"SUB:NOM:SIN:NEU"},
		"Mann": {"SUB:NOM:SIN:MAS"}, "mann": {"SUB:NOM:SIN:MAS"},
		"Frage": {"SUB:NOM:SIN:FEM"}, "frage": {"SUB:NOM:SIN:FEM"},
		"Fahrrad": {"SUB:NOM:SIN:NEU"}, "fahrrad": {"SUB:NOM:SIN:NEU"},
		"Garage": {"SUB:DAT:SIN:FEM"}, "garage": {"SUB:DAT:SIN:FEM"},
		"Schreibweise": {"SUB:NOM:SIN:FEM"}, "schreibweise": {"SUB:NOM:SIN:FEM"},
		"Töne": {"SUB:NOM:PLU:MAS"}, "töne": {"SUB:NOM:PLU:MAS"},
		"Nöten": {"SUB:DAT:PLU:FEM"}, "nöten": {"SUB:DAT:PLU:FEM"},
		"Prozent": {"SUB:NOM:PLU:NEU"}, "prozent": {"SUB:NOM:PLU:NEU"},
		"Testen": {"SUB:DAT:SIN:NEU:INF", "VER:INF:NON"}, "testen": {"VER:INF:NON", "SUB:DAT:SIN:NEU:INF"},
		"Laufen": {"SUB:NOM:SIN:NEU:INF", "VER:INF:NON"}, "laufen": {"VER:INF:NON"},
		"Fahren": {"SUB:NOM:SIN:NEU:INF", "VER:INF:NON"}, "fahren": {"VER:INF:NON"},
		"Winseln": {"SUB:NOM:SIN:NEU:INF", "VER:INF:NON"}, "winseln": {"VER:INF:NON"},
		"Verhalten": {"SUB:NOM:SIN:NEU"}, "verhalten": {"VER:INF:NON", "SUB:NOM:SIN:NEU"},
		"Vater": {"SUB:NOM:SIN:MAS"}, "Vaters": {"SUB:GEN:SIN:MAS"},
		"März": {"SUB:DAT:SIN:MAS"}, "märz": {"SUB:DAT:SIN:MAS"},
		// adjectives / adverbs (no pure SUB → can be uppercase errors when capitalized wrongly)
		"neu": {"ADJ:PRD:GRU"}, "neue": {"ADJ:NOM:SIN:NEU:GRU:DEF"},
		"Neue": {"ADJ:NOM:SIN:NEU:GRU:DEF", "SUB:NOM:SIN:NEU:ADJ"}, // :ADJ ignored for hasNounReading
		"neues": {"ADJ:AKK:SIN:NEU:GRU:IND"}, "Neues": {"ADJ:AKK:SIN:NEU:GRU:IND", "SUB:AKK:SIN:NEU:ADJ"},
		"einfacher": {"ADJ:NOM:SIN:MAS:GRU:IND"}, "Einfacher": {"ADJ:NOM:SIN:MAS:GRU:IND"},
		"einfache": {"ADJ:NOM:SIN:FEM:GRU:IND"}, "Einfache": {"ADJ:NOM:SIN:FEM:GRU:IND"},
		"blaue": {"ADJ:NOM:SIN:NEU:GRU:DEF"}, "Blaue": {"ADJ:NOM:SIN:NEU:GRU:DEF", "SUB:NOM:SIN:NEU:ADJ"},
		"große": {"ADJ:NOM:SIN:NEU:GRU:DEF"}, "Große": {"ADJ:NOM:SIN:NEU:GRU:DEF", "SUB:NOM:SIN:NEU:ADJ"},
		"heute": {"ADV"}, "Heute": {"ADV"},
		"früher": {"ADV:TMP"}, "Früher": {"ADV:TMP"},
		"schneller": {"ADJ:PRD:KOM"}, "Schneller": {"ADJ:PRD:KOM"},
		"bald": {"ADV:TMP"}, "Bald": {"ADV:TMP"},
		"groß": {"ADJ:PRD:GRU"}, "Groß": {"ADJ:PRD:GRU"},
		"mein": {"PRO:POS"}, "meines": {"PRO:POS:GEN:SIN:MAS"}, "Meines": {"PRO:POS:GEN:SIN:MAS"},
		"anderen": {"ADJ:NOM:PLU:MAS:GRU:DEF"}, "Anderen": {"ADJ:NOM:PLU:MAS:GRU:DEF", "SUB:NOM:PLU:MAS:ADJ"},
		"dreißig": {"ZAL"}, "Dreißig": {"ZAL"},
		// prepositions wrongly capitalized
		"über": {"PRP:LOK+TMP+CAU:DAT+AKK"}, "Über": {"PRP:LOK+TMP+CAU:DAT+AKK"},
		"im": {"APPRART:DAT:SIN:MAS"}, "Im": {"APPRART:DAT:SIN:MAS"},
		// verbs
		"machen": {"VER:INF:NON"}, "Machen": {"VER:INF:NON"},
		"lernt": {"VER:3:SIN:PRÄ:SFT"}, "Lernt": {"VER:3:SIN:PRÄ:SFT"},
		"stört": {"VER:3:SIN:PRÄ:SFT"}, "Stört": {"VER:3:SIN:PRÄ:SFT"},
		"vertraute": {"VER:3:SIN:PRT:SFT"}, "Vertraute": {"VER:3:SIN:PRT:SFT", "SUB:NOM:SIN:FEM:ADJ"},
		"ein": {"ART:IND:NOM:SIN:MAS"}, "Ein": {"ART:IND:NOM:SIN:MAS"},
		"eine": {"ART:IND:NOM:SIN:FEM"}, "Eine": {"ART:IND:NOM:SIN:FEM"},
		"kein": {"PIAT:NOM:SIN:MAS"}, "Kein": {"PIAT:NOM:SIN:MAS"},
	}
	ts, ok := m[word]
	if !ok {
		return nil
	}
	var readings []*languagetool.AnalyzedToken
	for _, p := range ts {
		pp, ww := p, word
		readings = append(readings, languagetool.NewAnalyzedToken(ww, &pp, &ww))
	}
	if len(readings) == 0 {
		return nil
	}
	atr := languagetool.NewAnalyzedTokenReadingsAt(readings[0], 0)
	for _, rd := range readings[1:] {
		atr.AddReading(rd, "")
	}
	return atr
}

func newCaseRuleTwin() *CaseRule {
	r := NewCaseRule(nil)
	r.Lookup = caseRuleTwinLookup
	// known lowercased common words are not misspelled (Java speller)
	r.IsMisspelled = func(w string) bool {
		known := map[string]bool{
			"haus": true, "zeit": true, "satz": true, "testen": true, "laufen": true,
			"fahren": true, "neu": true, "neue": true, "neues": true, "heute": true, "groß": true,
			"große": true, "mein": true, "meines": true, "vater": true, "vaters": true, "einfach": true,
			"einfacher": true, "einfache": true, "blaue": true, "auto": true, "mann": true,
			"frage": true, "fahrrad": true, "garage": true, "schreibweise": true, "töne": true,
			"nöten": true, "prozent": true, "winseln": true, "verhalten": true, "märz": true,
			"früher": true, "schneller": true, "bald": true, "anderen": true, "dreißig": true,
			"über": true, "im": true, "machen": true, "lernt": true, "stört": true,
			"vertraute": true, "ein": true, "eine": true, "kein": true, "okay": true,
		}
		return !known[w]
	}
	return r
}

// Twin of CaseRuleTest.testRule good/bad morph samples (POS inject + Lookup twin).
func TestCaseRule_MorphJavaSamples(t *testing.T) {
	rule := newCaseRuleTwin()

	assertGood := func(label string, toks ...*languagetool.AnalyzedTokenReadings) {
		t.Helper()
		all := append([]*languagetool.AnalyzedTokenReadings{sentStartATR()}, toks...)
		ms := rule.Match(languagetool.NewAnalyzedSentence(withPositions(all...)))
		require.Equal(t, 0, len(ms), "good %s got %d", label, len(ms))
	}
	assertBad := func(label string, toks ...*languagetool.AnalyzedTokenReadings) {
		t.Helper()
		all := append([]*languagetool.AnalyzedTokenReadings{sentStartATR()}, toks...)
		ms := rule.Match(languagetool.NewAnalyzedSentence(withPositions(all...)))
		require.GreaterOrEqual(t, len(ms), 1, "bad %s", label)
	}

	// Java assertGood
	assertGood("Ein einfacher Satz zum Testen.",
		atrWithPOS("Ein", "ART:IND:NOM:SIN:MAS", "ein"),
		atrWithPOS("einfacher", "ADJ:NOM:SIN:MAS:GRU:IND", "einfach"),
		atrWithPOS("Satz", "SUB:NOM:SIN:MAS", "Satz"),
		atrWithPOS("zum", "APPRART:DAT:SIN:NEU", "zu"),
		atrWithPOS("Testen", "SUB:DAT:SIN:NEU:INF", "Testen"),
		atrWithPOS(".", "PKT", "."),
	)
	assertGood("Das Laufen fällt mir leicht.",
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("Laufen", "SUB:NOM:SIN:NEU:INF", "Laufen"),
		atrWithPOS("fällt", "VER:3:SIN:PRÄ:SFT", "fallen"),
		atrWithPOS("mir", "PRO:PER:DAT:SIN:1", "ich"),
		atrWithPOS("leicht", "ADV", "leicht"),
		atrWithPOS(".", "PKT", "."),
	)
	assertGood("Das Fahren ist einfach.",
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("Fahren", "SUB:NOM:SIN:NEU:INF", "Fahren"),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("einfach", "ADJ:PRD:GRU", "einfach"),
		atrWithPOS(".", "PKT", "."),
	)
	assertGood("Sein Verhalten war okay.",
		atrWithPOS("Sein", "PRO:POS:NOM:SIN:NEU", "sein"),
		atrWithPOS("Verhalten", "SUB:NOM:SIN:NEU", "Verhalten"),
		atrWithPOS("war", "VER:3:SIN:PRT:NON", "sein"),
		atrWithPOS("okay", "ADJ:PRD:GRU", "okay"),
		atrWithPOS(".", "PKT", "."),
	)
	assertGood("A) Das Haus",
		atrWithPOS("A", "ABK", "A"),
		atrWithPOS(")", "SONST", ")"),
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus"),
	)
	assertGood("Das Winseln stört.",
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("Winseln", "SUB:NOM:SIN:NEU:INF", "Winseln"),
		atrWithPOS("stört", "VER:3:SIN:PRÄ:SFT", "stören"),
		atrWithPOS(".", "PKT", "."),
	)

	// Java assertBad
	assertBad("Und das Neue Haus.",
		atrWithPOS("Und", "KON:NEB", "und"),
		atrWithPOS("das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("Neue", "ADJ:NOM:SIN:NEU:GRU:DEF", "neu"),
		atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus"),
		atrWithPOS(".", "PKT", "."),
	)
	assertBad("Ich habe Heute keine Zeit.",
		atrWithPOS("Ich", "PRO:PER:NOM:SIN:1", "ich"),
		atrWithPOS("habe", "VER:1:SIN:PRÄ:NON", "haben"),
		atrWithPOS("Heute", "ADV", "heute"),
		atrWithPOS("keine", "PIAT:AKK:SIN:FEM", "kein"),
		atrWithPOS("Zeit", "SUB:AKK:SIN:FEM", "Zeit"),
		atrWithPOS(".", "PKT", "."),
	)
	assertBad("Er ist Groß.",
		atrWithPOS("Er", "PRO:PER:NOM:SIN:3:MAS", "er"),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("Groß", "ADJ:PRD:GRU", "groß"),
		atrWithPOS(".", "PKT", "."),
	)
	assertBad("Das fahren ist einfach.",
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("fahren", "VER:INF:NON", "fahren"),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("einfach", "ADJ:PRD:GRU", "einfach"),
		atrWithPOS(".", "PKT", "."),
	)
	assertBad("Das Auto Meines Vaters.",
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("Auto", "SUB:NOM:SIN:NEU", "Auto"),
		atrWithPOS("Meines", "PRO:POS:GEN:SIN:MAS", "mein"),
		atrWithPOS("Vaters", "SUB:GEN:SIN:MAS", "Vater"),
		atrWithPOS(".", "PKT", "."),
	)
	assertBad("Ein Einfacher Satz zum Testen.",
		atrWithPOS("Ein", "ART:IND:NOM:SIN:MAS", "ein"),
		atrWithPOS("Einfacher", "ADJ:NOM:SIN:MAS:GRU:IND", "einfach"),
		atrWithPOS("Satz", "SUB:NOM:SIN:MAS", "Satz"),
		atrWithPOS("zum", "APPRART:DAT:SIN:NEU", "zu"),
		atrWithPOS("Testen", "SUB:DAT:SIN:NEU:INF", "Testen"),
		atrWithPOS(".", "PKT", "."),
	)
	assertBad("Eine Einfache Frage zum Testen?",
		atrWithPOS("Eine", "ART:IND:NOM:SIN:FEM", "ein"),
		atrWithPOS("Einfache", "ADJ:NOM:SIN:FEM:GRU:IND", "einfach"),
		atrWithPOS("Frage", "SUB:NOM:SIN:FEM", "Frage"),
		atrWithPOS("zum", "APPRART:DAT:SIN:NEU", "zu"),
		atrWithPOS("Testen", "SUB:DAT:SIN:NEU:INF", "Testen"),
		atrWithPOS("?", "PKT", "?"),
	)
	assertBad("Das Blaue Auto.",
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("Blaue", "ADJ:NOM:SIN:NEU:GRU:DEF", "blau"),
		atrWithPOS("Auto", "SUB:NOM:SIN:NEU", "Auto"),
		atrWithPOS(".", "PKT", "."),
	)
	assertBad("Er kam Früher als sonst.",
		atrWithPOS("Er", "PRO:PER:NOM:SIN:3:MAS", "er"),
		atrWithPOS("kam", "VER:3:SIN:PRT:NON", "kommen"),
		atrWithPOS("Früher", "ADV:TMP", "früh"),
		atrWithPOS("als", "KON:NEB", "als"),
		atrWithPOS("sonst", "ADV", "sonst"),
		atrWithPOS(".", "PKT", "."),
	)
	// Java: surface often untagged; lowercase lookup ADJ:PRD:KOM skips isAdjectiveAsNoun
	// (CaseRule.java comment: avoid false true for ADJ:PRD:KOM when posTag == null).
	assertBad("Er rennt Schneller als ich.",
		atrWithPOS("Er", "PRO:PER:NOM:SIN:3:MAS", "er"),
		atrWithPOS("rennt", "VER:3:SIN:PRÄ:SFT", "rennen"),
		atrWithPOS("Schneller", "", ""),
		atrWithPOS("als", "KON:NEB", "als"),
		atrWithPOS("ich", "PRO:PER:NOM:SIN:1", "ich"),
		atrWithPOS(".", "PKT", "."),
	)
	assertBad("Das Winseln Stört.",
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("Winseln", "SUB:NOM:SIN:NEU:INF", "Winseln"),
		atrWithPOS("Stört", "VER:3:SIN:PRÄ:SFT", "stören"),
		atrWithPOS(".", "PKT", "."),
	)
	assertBad("Sein verhalten war okay.",
		atrWithPOS("Sein", "PRO:POS:NOM:SIN:NEU", "sein"),
		atrWithPOS("verhalten", "VER:INF:NON", "verhalten"),
		atrWithPOS("war", "VER:3:SIN:PRT:NON", "sein"),
		atrWithPOS("okay", "ADJ:PRD:GRU", "okay"),
		atrWithPOS(".", "PKT", "."),
	)
	assertBad("Bis Bald!",
		atrWithPOS("Bis", "APPR", "bis"),
		atrWithPOS("Bald", "ADV:TMP", "bald"),
		atrWithPOS("!", "PKT", "!"),
	)
	assertBad("Das ist Ein Mann.",
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("Ein", "ART:IND:NOM:SIN:MAS", "ein"),
		atrWithPOS("Mann", "SUB:NOM:SIN:MAS", "Mann"),
		atrWithPOS(".", "PKT", "."),
	)
	assertBad("Das ist Eine Schreibweise.",
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("Eine", "ART:IND:NOM:SIN:FEM", "ein"),
		atrWithPOS("Schreibweise", "SUB:NOM:SIN:FEM", "Schreibweise"),
		atrWithPOS(".", "PKT", "."),
	)
	assertBad("Das machen der Töne ist schwierig.",
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("machen", "VER:INF:NON", "machen"),
		atrWithPOS("der", "ART:DEF:GEN:PLU:MAS", "der"),
		atrWithPOS("Töne", "SUB:GEN:PLU:MAS", "Ton"),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("schwierig", "ADJ:PRD:GRU", "schwierig"),
		atrWithPOS(".", "PKT", "."),
	)
	assertBad("Sie Vertraute niemandem.",
		atrWithPOS("Sie", "PRO:PER:NOM:SIN:3:FEM", "sie"),
		atrWithPOS("Vertraute", "VER:3:SIN:PRT:SFT", "vertrauen"),
		atrWithPOS("niemandem", "PRO:IND:DAT:SIN:MAS", "niemand"),
		atrWithPOS(".", "PKT", "."),
	)
	assertBad("Beten Lernt man in Nöten.",
		atrWithPOS("Beten", "VER:INF:NON", "beten"),
		atrWithPOS("Lernt", "VER:3:SIN:PRÄ:SFT", "lernen"),
		atrWithPOS("man", "PRO:IND:NOM:SIN:MAS", "man"),
		atrWithPOS("in", "APPR", "in"),
		atrWithPOS("Nöten", "SUB:DAT:PLU:FEM", "Not"),
		atrWithPOS(".", "PKT", "."),
	)
	assertBad("Ich habe ein Neues Fahrrad.",
		atrWithPOS("Ich", "PRO:PER:NOM:SIN:1", "ich"),
		atrWithPOS("habe", "VER:1:SIN:PRÄ:NON", "haben"),
		atrWithPOS("ein", "ART:IND:AKK:SIN:NEU", "ein"),
		atrWithPOS("Neues", "ADJ:AKK:SIN:NEU:GRU:IND", "neu"),
		atrWithPOS("Fahrrad", "SUB:AKK:SIN:NEU", "Fahrrad"),
		atrWithPOS(".", "PKT", "."),
	)
	assertBad("Das Große Auto wurde gewaschen.",
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("Große", "ADJ:NOM:SIN:NEU:GRU:DEF", "groß"),
		atrWithPOS("Auto", "SUB:NOM:SIN:NEU", "Auto"),
		atrWithPOS("wurde", "VER:AUX:3:SIN:PRT:SFT", "werden"),
		atrWithPOS("gewaschen", "PA2:PRD:GRU:VER", "waschen"),
		atrWithPOS(".", "PKT", "."),
	)
	assertBad("Das ist es: Kein Satz.",
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("es", "PRO:PER:NOM:SIN:3:NEU", "es"),
		atrWithPOS(":", "PKT", ":"),
		atrWithPOS("Kein", "PIAT:NOM:SIN:MAS", "kein"),
		atrWithPOS("Satz", "SUB:NOM:SIN:MAS", "Satz"),
		atrWithPOS(".", "PKT", "."),
	)
	assertBad("Er wohnt Über einer Garage.",
		atrWithPOS("Er", "PRO:PER:NOM:SIN:3:MAS", "er"),
		atrWithPOS("wohnt", "VER:3:SIN:PRÄ:SFT", "wohnen"),
		atrWithPOS("Über", "PRP:LOK+TMP+CAU:DAT+AKK", "über"),
		atrWithPOS("einer", "ART:IND:DAT:SIN:FEM", "ein"),
		atrWithPOS("Garage", "SUB:DAT:SIN:FEM", "Garage"),
		atrWithPOS(".", "PKT", "."),
	)
	assertBad("Er war dort Im März 2000.",
		atrWithPOS("Er", "PRO:PER:NOM:SIN:3:MAS", "er"),
		atrWithPOS("war", "VER:3:SIN:PRT:NON", "sein"),
		atrWithPOS("dort", "ADV", "dort"),
		atrWithPOS("Im", "APPRART:DAT:SIN:MAS", "in"),
		atrWithPOS("März", "SUB:DAT:SIN:MAS", "März"),
		atrWithPOS("2000", "ZAL", "2000"),
		atrWithPOS(".", "PKT", "."),
	)
	assertBad("Die Anderen 90 Prozent waren krank.",
		atrWithPOS("Die", "ART:DEF:NOM:PLU:MAS", "der"),
		atrWithPOS("Anderen", "ADJ:NOM:PLU:MAS:GRU:DEF", "ander"),
		atrWithPOS("90", "ZAL", "90"),
		atrWithPOS("Prozent", "SUB:NOM:PLU:NEU", "Prozent"),
		atrWithPOS("waren", "VER:3:PLU:PRT:NON", "sein"),
		atrWithPOS("krank", "ADJ:PRD:GRU", "krank"),
		atrWithPOS(".", "PKT", "."),
	)
	assertBad("Tom ist etwas über Dreißig.",
		atrWithPOS("Tom", "EIG:NOM:SIN:MAS", "Tom"),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("etwas", "ADV", "etwas"),
		atrWithPOS("über", "PRP:LOK+TMP+CAU:DAT+AKK", "über"),
		atrWithPOS("Dreißig", "ZAL", "dreißig"),
		atrWithPOS(".", "PKT", "."),
	)

	// untagged must not invent
	require.Equal(t, 0, len(NewCaseRule(nil).Match(languagetool.AnalyzePlain("Und das Neue Haus."))))
	require.Equal(t, 0, len(NewCaseRule(nil).Match(languagetool.AnalyzePlain("Ich habe Heute keine Zeit."))))
	require.Equal(t, 0, len(NewCaseRule(nil).Match(languagetool.AnalyzePlain("Ein Einfacher Satz zum Testen."))))
	require.Equal(t, 0, len(NewCaseRule(nil).Match(languagetool.AnalyzePlain("Das ist Ein Mann."))))
}

// Twin of CaseRuleTest.testSubstantivierteVerben — more Java good/bad infinitive nominalizations.
func TestCaseRule_SubstantivierteVerben_JavaTable(t *testing.T) {
	rule := newCaseRuleTwin()
	// extend lookup for more infinitives used in Java testSubstantivierteVerben
	base := caseRuleTwinLookup
	rule.Lookup = func(word string) *languagetool.AnalyzedTokenReadings {
		if r := base(word); r != nil {
			return r
		}
		inf := map[string]string{
			"gehen": "gehen", "Gehen": "gehen",
			"essen": "essen", "Essen": "essen",
			"lesen": "lesen", "Lesen": "lesen",
		}
		if lem, ok := inf[word]; ok {
			w := word
			if toolsStartsUpper(word) {
				return languagetool.NewAnalyzedTokenReadingsAt(
					languagetool.NewAnalyzedToken(w, strPtr("SUB:NOM:SIN:NEU:INF"), &lem), 0)
			}
			return languagetool.NewAnalyzedTokenReadingsAt(
				languagetool.NewAnalyzedToken(w, strPtr("VER:INF:NON"), &lem), 0)
		}
		return nil
	}
	known := map[string]bool{"gehen": true, "essen": true, "lesen": true, "fahren": true, "laufen": true}
	prevMiss := rule.IsMisspelled
	rule.IsMisspelled = func(w string) bool {
		if known[w] {
			return false
		}
		return prevMiss(w)
	}

	assertGood := func(label string, toks ...*languagetool.AnalyzedTokenReadings) {
		t.Helper()
		all := append([]*languagetool.AnalyzedTokenReadings{sentStartATR()}, toks...)
		ms := rule.Match(languagetool.NewAnalyzedSentence(withPositions(all...)))
		require.Equal(t, 0, len(ms), "good %s", label)
	}
	assertBad := func(label string, toks ...*languagetool.AnalyzedTokenReadings) {
		t.Helper()
		all := append([]*languagetool.AnalyzedTokenReadings{sentStartATR()}, toks...)
		ms := rule.Match(languagetool.NewAnalyzedSentence(withPositions(all...)))
		require.GreaterOrEqual(t, len(ms), 1, "bad %s", label)
	}

	assertGood("Denn das Fahren ist einfach.",
		atrWithPOS("Denn", "KON:NEB", "denn"),
		atrWithPOS("das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("Fahren", "SUB:NOM:SIN:NEU:INF", "Fahren"),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("einfach", "ADJ:PRD:GRU", "einfach"),
		atrWithPOS(".", "PKT", "."),
	)
	assertGood("Das Gehen fällt mir leicht.",
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("Gehen", "SUB:NOM:SIN:NEU:INF", "Gehen"),
		atrWithPOS("fällt", "VER:3:SIN:PRÄ:SFT", "fallen"),
		atrWithPOS("mir", "PRO:PER:DAT:SIN:1", "ich"),
		atrWithPOS("leicht", "ADV", "leicht"),
		atrWithPOS(".", "PKT", "."),
	)
	assertGood("Ich liebe das Lesen.",
		atrWithPOS("Ich", "PRO:PER:NOM:SIN:1", "ich"),
		atrWithPOS("liebe", "VER:1:SIN:PRÄ:SFT", "lieben"),
		atrWithPOS("das", "ART:DEF:AKK:SIN:NEU", "das"),
		atrWithPOS("Lesen", "SUB:AKK:SIN:NEU:INF", "Lesen"),
		atrWithPOS(".", "PKT", "."),
	)

	assertBad("Denn das fahren ist einfach.",
		atrWithPOS("Denn", "KON:NEB", "denn"),
		atrWithPOS("das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("fahren", "VER:INF:NON", "fahren"),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("einfach", "ADJ:PRD:GRU", "einfach"),
		atrWithPOS(".", "PKT", "."),
	)
	assertBad("Denn das laufen ist einfach.",
		atrWithPOS("Denn", "KON:NEB", "denn"),
		atrWithPOS("das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("laufen", "VER:INF:NON", "laufen"),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("einfach", "ADJ:PRD:GRU", "einfach"),
		atrWithPOS(".", "PKT", "."),
	)
	assertBad("Denn das essen ist einfach.",
		atrWithPOS("Denn", "KON:NEB", "denn"),
		atrWithPOS("das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("essen", "VER:INF:NON", "essen"),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("einfach", "ADJ:PRD:GRU", "einfach"),
		atrWithPOS(".", "PKT", "."),
	)
	assertBad("Denn das gehen ist einfach.",
		atrWithPOS("Denn", "KON:NEB", "denn"),
		atrWithPOS("das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("gehen", "VER:INF:NON", "gehen"),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("einfach", "ADJ:PRD:GRU", "einfach"),
		atrWithPOS(".", "PKT", "."),
	)
}

func toolsStartsUpper(s string) bool {
	// local helper for Lookup twin only (mirrors Character.isUpperCase first letter)
	if s == "" {
		return false
	}
	r := []rune(s)[0]
	return (r >= 'A' && r <= 'Z') || r == 'Ä' || r == 'Ö' || r == 'Ü'
}
