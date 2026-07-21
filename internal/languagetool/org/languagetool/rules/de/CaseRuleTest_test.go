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
		"Testen": {"SUB:DAT:SIN:NEU:INF", "VER:INF:NON"}, "testen": {"VER:INF:NON", "SUB:DAT:SIN:NEU:INF"},
		"Laufen": {"SUB:NOM:SIN:NEU:INF", "VER:INF:NON"}, "laufen": {"VER:INF:NON"},
		"Fahren": {"SUB:NOM:SIN:NEU:INF", "VER:INF:NON"}, "fahren": {"VER:INF:NON"},
		"Vater": {"SUB:NOM:SIN:MAS"}, "Vaters": {"SUB:GEN:SIN:MAS"},
		// adjectives / adverbs (no pure SUB → can be uppercase errors when capitalized wrongly)
		"neu": {"ADJ:PRD:GRU"}, "neue": {"ADJ:NOM:SIN:NEU:GRU:DEF"},
		"Neue": {"ADJ:NOM:SIN:NEU:GRU:DEF", "SUB:NOM:SIN:NEU:ADJ"}, // :ADJ ignored for hasNounReading
		"heute": {"ADV"}, "Heute": {"ADV"},
		"groß": {"ADJ:PRD:GRU"}, "Groß": {"ADJ:PRD:GRU"},
		"mein": {"PRO:POS"}, "meines": {"PRO:POS:GEN:SIN:MAS"}, "Meines": {"PRO:POS:GEN:SIN:MAS"},
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
			"fahren": true, "neu": true, "neue": true, "heute": true, "groß": true,
			"mein": true, "meines": true, "vater": true, "vaters": true, "einfach": true,
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

	// untagged must not invent
	require.Equal(t, 0, len(NewCaseRule(nil).Match(languagetool.AnalyzePlain("Und das Neue Haus."))))
	require.Equal(t, 0, len(NewCaseRule(nil).Match(languagetool.AnalyzePlain("Ich habe Heute keine Zeit."))))
}
