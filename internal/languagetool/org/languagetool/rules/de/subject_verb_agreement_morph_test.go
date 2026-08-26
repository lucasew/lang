package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSubjectVerbAgreementRule_MorphPluralSubjectSingularVerb(t *testing.T) {
	// Die Autos ist — NPP-like SUB:PLU + ist
	ss := languagetool.SentenceStartTagName
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		atrWithPOS("Die", "ART:DEF:NOM:PLU:ALG", "die"),
		atrWithPOS("Autos", "SUB:NOM:PLU:NEU", "Auto"),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("schnell", "ADJ:PRD:GRU", "schnell"),
	}
	// chunk NPP on subject noun (as German chunker would)
	toks[2].SetChunkTags([]string{chunkNPP})
	toks[1].SetChunkTags([]string{chunkNPP})
	toks[1].SetStartPos(0)
	toks[2].SetStartPos(4)
	toks[3].SetStartPos(10)
	sent := languagetool.NewAnalyzedSentence(toks)
	ms := NewSubjectVerbAgreementRule(nil).Match(sent)
	require.NotEmpty(t, ms)
	require.Contains(t, ms[0].GetSuggestedReplacements(), "sind")
	// Java message embeds <suggestion>sind</suggestion>; no invent shortMessage.
	require.Contains(t, ms[0].GetMessage(), "<suggestion>sind</suggestion>")
	require.Empty(t, ms[0].GetShortMessage())
}

func TestSubjectVerbAgreementRule_Meta(t *testing.T) {
	r := NewSubjectVerbAgreementRule(nil)
	require.Equal(t, "Kongruenz von Subjekt und Prädikat (unvollständig)", r.GetDescription())
	require.Greater(t, r.EstimateContextForSureMatch(), 0)
	require.Equal(t,
		"https://dict.leo.org/grammatik/deutsch/Wort/Verb/Kategorien/Numerus-Person/ProblemNum.html",
		r.GetURL())
}

func TestSubjectVerbAgreementRule_MorphSingularSubjectPluralVerb(t *testing.T) {
	// Das Auto sind — NPS + sind
	ss := languagetool.SentenceStartTagName
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("Auto", "SUB:NOM:SIN:NEU", "Auto"),
		atrWithPOS("sind", "VER:3:PLU:PRÄ:NON", "sein"),
		atrWithPOS("schnell", "ADJ:PRD:GRU", "schnell"),
	}
	toks[1].SetChunkTags([]string{chunkNPS})
	toks[2].SetChunkTags([]string{chunkNPS})
	toks[1].SetStartPos(0)
	toks[2].SetStartPos(4)
	toks[3].SetStartPos(9)
	sent := languagetool.NewAnalyzedSentence(toks)
	ms := NewSubjectVerbAgreementRule(nil).Match(sent)
	require.NotEmpty(t, ms)
	require.Contains(t, ms[0].GetSuggestedReplacements(), "ist")
}

func TestSubjectVerbAgreementRule_MorphOK(t *testing.T) {
	ss := languagetool.SentenceStartTagName
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		atrWithPOS("Die", "ART:DEF:NOM:SIN:FEM", "die"),
		atrWithPOS("Katze", "SUB:NOM:SIN:FEM", "Katze"),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("schön", "ADJ:PRD:GRU", "schön"),
	}
	toks[1].SetChunkTags([]string{chunkNPS})
	toks[2].SetChunkTags([]string{chunkNPS})
	sent := languagetool.NewAnalyzedSentence(toks)
	ms := NewSubjectVerbAgreementRule(nil).Match(sent)
	require.Empty(t, ms)
}

func TestSubjectVerbAgreementRule_NoChunksNoInvent(t *testing.T) {
	// Java: only chunkTags.contains(NPP/NPS) — POS-only SUB:PLU must not invent a hit.
	ss := languagetool.SentenceStartTagName
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		atrWithPOS("Autos", "SUB:NOM:PLU:NEU", "Auto"),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
	}
	toks[1].SetStartPos(0)
	toks[2].SetStartPos(6)
	// no SetChunkTags
	sent := languagetool.NewAnalyzedSentence(toks)
	ms := NewSubjectVerbAgreementRule(nil).Match(sent)
	require.Empty(t, ms, "without NPP/NPS chunks Java does not match")
}

func TestSubjectVerbAgreementRule_PrevChunkIsNominative_NoChunkFalse(t *testing.T) {
	ss := languagetool.SentenceStartTagName
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		atrWithPOS("Autos", "SUB:NOM:PLU:NEU", "Auto"),
	}
	// NOM alone without NPS/NPP must not invent nominative chunk span
	require.False(t, prevChunkIsNominative(toks, 1))
}

func TestSubjectVerbAgreementRule_PrevChunkIsNominativeMorph(t *testing.T) {
	ss := languagetool.SentenceStartTagName
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0),
		atrWithPOS("Autos", "SUB:NOM:PLU:NEU", "Auto"),
	}
	toks[1].SetChunkTags([]string{chunkNPP})
	require.True(t, prevChunkIsNominative(toks, 1))
}

// nppSubject marks the last subject token (and optionally more) as NPP nominative chunk.
func nppSubject(toks ...*languagetool.AnalyzedTokenReadings) {
	for _, t := range toks {
		if t != nil {
			t.SetChunkTags([]string{chunkNPP})
		}
	}
}

func npsSubject(toks ...*languagetool.AnalyzedTokenReadings) {
	for _, t := range toks {
		if t != nil {
			t.SetChunkTags([]string{chunkNPS})
		}
	}
}

// Twin of SubjectVerbAgreementRuleTest.testRuleWithIncorrectSingularVerb — morph/chunk inject samples.
func TestSubjectVerbAgreementRule_IncorrectSingularVerb_JavaTable(t *testing.T) {
	rule := NewSubjectVerbAgreementRule(nil)
	assertBad := func(label string, toks ...*languagetool.AnalyzedTokenReadings) {
		t.Helper()
		all := append([]*languagetool.AnalyzedTokenReadings{sentStartATR()}, toks...)
		ms := rule.Match(languagetool.NewAnalyzedSentence(withPositions(all...)))
		require.GreaterOrEqual(t, len(ms), 1, "bad %s", label)
		require.Contains(t, ms[0].GetSuggestedReplacements(), "sind", "label %s", label)
	}

	// Die Autos ist schnell.
	die, autos := atrWithPOS("Die", "ART:DEF:NOM:PLU:ALG", "die"), atrWithPOS("Autos", "SUB:NOM:PLU:NEU", "Auto")
	nppSubject(die, autos)
	assertBad("Die Autos ist",
		die, autos,
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("schnell", "ADJ:PRD:GRU", "schnell"),
		atrWithPOS(".", "PKT", "."),
	)

	// Der Hund und die Katze ist draußen. — last subject chunk NPP (coordinated plural)
	die2, katze := atrWithPOS("die", "ART:DEF:NOM:SIN:FEM", "die"), atrWithPOS("Katze", "SUB:NOM:SIN:FEM", "Katze")
	nppSubject(die2, katze)
	assertBad("Der Hund und die Katze ist",
		atrWithPOS("Der", "ART:DEF:NOM:SIN:MAS", "der"),
		atrWithPOS("Hund", "SUB:NOM:SIN:MAS", "Hund"),
		atrWithPOS("und", "KON:NEB", "und"),
		die2, katze,
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("draußen", "ADV", "draußen"),
		atrWithPOS(".", "PKT", "."),
	)

	// Die Kenntnisse ist je nach Bildungsgrad verschieden.
	die3, kennt := atrWithPOS("Die", "ART:DEF:NOM:PLU:FEM", "die"), atrWithPOS("Kenntnisse", "SUB:NOM:PLU:FEM", "Kenntnis")
	nppSubject(die3, kennt)
	assertBad("Die Kenntnisse ist",
		die3, kennt,
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("je", "ADV", "je"),
		atrWithPOS("nach", "APPR", "nach"),
		atrWithPOS("Bildungsgrad", "SUB:DAT:SIN:MAS", "Bildungsgrad"),
		atrWithPOS("verschieden", "ADJ:PRD:GRU", "verschieden"),
		atrWithPOS(".", "PKT", "."),
	)

	// Drei Katzen ist im Haus.
	drei, katzen := atrWithPOS("Drei", "ZAL", "drei"), atrWithPOS("Katzen", "SUB:NOM:PLU:FEM", "Katze")
	nppSubject(drei, katzen)
	assertBad("Drei Katzen ist",
		drei, katzen,
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("im", "APPRART:DAT:SIN:NEU", "in"),
		atrWithPOS("Haus", "SUB:DAT:SIN:NEU", "Haus"),
		atrWithPOS(".", "PKT", "."),
	)

	// Viele Katzen ist schön.
	viele, katzen2 := atrWithPOS("Viele", "PIAT:NOM:PLU:FEM", "viel"), atrWithPOS("Katzen", "SUB:NOM:PLU:FEM", "Katze")
	nppSubject(viele, katzen2)
	assertBad("Viele Katzen ist",
		viele, katzen2,
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("schön", "ADJ:PRD:GRU", "schön"),
		atrWithPOS(".", "PKT", "."),
	)

	// Die ältesten und bekanntesten Maßnahmen ist …
	mass := atrWithPOS("Maßnahmen", "SUB:NOM:PLU:FEM", "Maßnahme")
	nppSubject(mass)
	assertBad("Maßnahmen ist",
		atrWithPOS("Die", "ART:DEF:NOM:PLU:FEM", "die"),
		atrWithPOS("ältesten", "ADJ:NOM:PLU:FEM:SUP:DEF", "alt"),
		atrWithPOS("und", "KON:NEB", "und"),
		atrWithPOS("bekanntesten", "ADJ:NOM:PLU:FEM:SUP:DEF", "bekannt"),
		mass,
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("die", "ART:DEF:NOM:SIN:FEM", "die"),
		atrWithPOS("Einrichtung", "SUB:NOM:SIN:FEM", "Einrichtung"),
		atrWithPOS(".", "PKT", "."),
	)

	// Isolation und ihre Überwindung ist — coordinated NPP on last noun
	ueber := atrWithPOS("Überwindung", "SUB:NOM:SIN:FEM", "Überwindung")
	nppSubject(ueber)
	assertBad("Isolation und Überwindung ist",
		atrWithPOS("Isolation", "SUB:NOM:SIN:FEM", "Isolation"),
		atrWithPOS("und", "KON:NEB", "und"),
		atrWithPOS("ihre", "PRO:POS:NOM:SIN:FEM", "ihr"),
		ueber,
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("ein", "ART:IND:NOM:SIN:NEU", "ein"),
		atrWithPOS("Thema", "SUB:NOM:SIN:NEU", "Thema"),
		atrWithPOS(".", "PKT", "."),
	)

	// untagged must not invent any of these
	for _, s := range []string{
		"Die Autos ist schnell.",
		"Der Hund und die Katze ist draußen.",
		"Drei Katzen ist im Haus.",
	} {
		require.Equal(t, 0, len(NewSubjectVerbAgreementRule(nil).Match(languagetool.AnalyzePlain(s))), s)
	}
}

// Twin of SubjectVerbAgreementRuleTest.testRuleWithIncorrectPluralVerb — morph samples.
func TestSubjectVerbAgreementRule_IncorrectPluralVerb_JavaTable(t *testing.T) {
	rule := NewSubjectVerbAgreementRule(nil)
	assertBad := func(label, wantSug string, toks ...*languagetool.AnalyzedTokenReadings) {
		t.Helper()
		all := append([]*languagetool.AnalyzedTokenReadings{sentStartATR()}, toks...)
		ms := rule.Match(languagetool.NewAnalyzedSentence(withPositions(all...)))
		require.GreaterOrEqual(t, len(ms), 1, "bad %s", label)
		require.Contains(t, ms[0].GetSuggestedReplacements(), wantSug, "label %s sugs=%v", label, ms[0].GetSuggestedReplacements())
	}

	// Die Katze sind schön.
	die, katze := atrWithPOS("Die", "ART:DEF:NOM:SIN:FEM", "die"), atrWithPOS("Katze", "SUB:NOM:SIN:FEM", "Katze")
	npsSubject(die, katze)
	assertBad("Die Katze sind", "ist",
		die, katze,
		atrWithPOS("sind", "VER:3:PLU:PRÄ:NON", "sein"),
		atrWithPOS("schön", "ADJ:PRD:GRU", "schön"),
		atrWithPOS(".", "PKT", "."),
	)

	// Die Katze waren schön.
	die2, katze2 := atrWithPOS("Die", "ART:DEF:NOM:SIN:FEM", "die"), atrWithPOS("Katze", "SUB:NOM:SIN:FEM", "Katze")
	npsSubject(die2, katze2)
	assertBad("Die Katze waren", "war",
		die2, katze2,
		atrWithPOS("waren", "VER:3:PLU:PRT:NON", "sein"),
		atrWithPOS("schön", "ADJ:PRD:GRU", "schön"),
		atrWithPOS(".", "PKT", "."),
	)

	// Der Fisch sind gut.
	der, fisch := atrWithPOS("Der", "ART:DEF:NOM:SIN:MAS", "der"), atrWithPOS("Fisch", "SUB:NOM:SIN:MAS", "Fisch")
	npsSubject(der, fisch)
	assertBad("Der Fisch sind", "ist",
		der, fisch,
		atrWithPOS("sind", "VER:3:PLU:PRÄ:NON", "sein"),
		atrWithPOS("gut", "ADJ:PRD:GRU", "gut"),
		atrWithPOS(".", "PKT", "."),
	)

	// Das Auto sind schnell.
	das, auto := atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "das"), atrWithPOS("Auto", "SUB:NOM:SIN:NEU", "Auto")
	npsSubject(das, auto)
	assertBad("Das Auto sind", "ist",
		das, auto,
		atrWithPOS("sind", "VER:3:PLU:PRÄ:NON", "sein"),
		atrWithPOS("schnell", "ADJ:PRD:GRU", "schnell"),
		atrWithPOS(".", "PKT", "."),
	)

	// Herr Schröder sind alt.
	herr, schr := atrWithPOS("Herr", "SUB:NOM:SIN:MAS", "Herr"), atrWithPOS("Schröder", "EIG:NOM:SIN:MAS", "Schröder")
	npsSubject(herr, schr)
	assertBad("Herr Schröder sind", "ist",
		herr, schr,
		atrWithPOS("sind", "VER:3:PLU:PRÄ:NON", "sein"),
		atrWithPOS("alt", "ADJ:PRD:GRU", "alt"),
		atrWithPOS(".", "PKT", "."),
	)

	// Julia und Karsten ist alt. — coordinated subject last chunk NPP + singular verb
	karsten := atrWithPOS("Karsten", "EIG:NOM:SIN:MAS", "Karsten")
	nppSubject(karsten)
	assertBad("Julia und Karsten ist", "sind",
		atrWithPOS("Julia", "EIG:NOM:SIN:FEM", "Julia"),
		atrWithPOS("und", "KON:NEB", "und"),
		karsten,
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("alt", "ADJ:PRD:GRU", "alt"),
		atrWithPOS(".", "PKT", "."),
	)

	require.Equal(t, 0, len(NewSubjectVerbAgreementRule(nil).Match(languagetool.AnalyzePlain("Die Katze sind schön."))))
	require.Equal(t, 0, len(NewSubjectVerbAgreementRule(nil).Match(languagetool.AnalyzePlain("Julia und Karsten ist alt."))))
}

// Twin of SubjectVerbAgreementRuleTest correct singular/plural morph samples.
func TestSubjectVerbAgreementRule_CorrectSingularPlural_JavaSamples(t *testing.T) {
	rule := NewSubjectVerbAgreementRule(nil)
	assertGood := func(label string, toks ...*languagetool.AnalyzedTokenReadings) {
		t.Helper()
		all := append([]*languagetool.AnalyzedTokenReadings{sentStartATR()}, toks...)
		ms := rule.Match(languagetool.NewAnalyzedSentence(withPositions(all...)))
		require.Equal(t, 0, len(ms), "good %s", label)
	}

	// Die Katze ist schön.
	die, katze := atrWithPOS("Die", "ART:DEF:NOM:SIN:FEM", "die"), atrWithPOS("Katze", "SUB:NOM:SIN:FEM", "Katze")
	npsSubject(die, katze)
	assertGood("Die Katze ist",
		die, katze,
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("schön", "ADJ:PRD:GRU", "schön"),
		atrWithPOS(".", "PKT", "."),
	)

	// Die Katzen sind schön.
	die2, katzen := atrWithPOS("Die", "ART:DEF:NOM:PLU:FEM", "die"), atrWithPOS("Katzen", "SUB:NOM:PLU:FEM", "Katze")
	nppSubject(die2, katzen)
	assertGood("Die Katzen sind",
		die2, katzen,
		atrWithPOS("sind", "VER:3:PLU:PRÄ:NON", "sein"),
		atrWithPOS("schön", "ADJ:PRD:GRU", "schön"),
		atrWithPOS(".", "PKT", "."),
	)

	// Beiden Filmen war kein Erfolg — dative NPP/NPS should not trigger (prev not nominative)
	beiden, filmen := atrWithPOS("Beiden", "ART:DEF:DAT:PLU:MAS", "beide"), atrWithPOS("Filmen", "SUB:DAT:PLU:MAS", "Film")
	// DAT not NOM — even with NPS chunk, prevChunkIsNominative requires NOM
	npsSubject(beiden, filmen)
	assertGood("Beiden Filmen war",
		beiden, filmen,
		atrWithPOS("war", "VER:3:SIN:PRT:NON", "sein"),
		atrWithPOS("kein", "PIAT:NOM:SIN:MAS", "kein"),
		atrWithPOS("Erfolg", "SUB:NOM:SIN:MAS", "Erfolg"),
		atrWithPOS(".", "PKT", "."),
	)

	// Ein Gramm Pfeffer war früher wertvoll. — currency-like measure: Gramm may be currency skip
	// Java good with singular; inject NPS on Gramm
	gramm := atrWithPOS("Gramm", "SUB:NOM:SIN:NEU", "Gramm")
	npsSubject(gramm)
	assertGood("Ein Gramm war",
		atrWithPOS("Ein", "ART:IND:NOM:SIN:NEU", "ein"),
		gramm,
		atrWithPOS("Pfeffer", "SUB:NOM:SIN:MAS", "Pfeffer"),
		atrWithPOS("war", "VER:3:SIN:PRT:NON", "sein"),
		atrWithPOS("früher", "ADV:TMP", "früh"),
		atrWithPOS("wertvoll", "ADJ:PRD:GRU", "wertvoll"),
		atrWithPOS(".", "PKT", "."),
	)

	// Start und Ziel is Innsbruck — coordinated last noun NPP but verb "ist" singular is GOOD in Java
	// (special und-coordination list). Morph: if chunk is NPP on Ziel, singular match would fire —
	// Java chunker may leave NPS. Mirror Java good with NPS on last noun.
	ziel := atrWithPOS("Ziel", "SUB:NOM:SIN:NEU", "Ziel")
	npsSubject(ziel)
	assertGood("Start und Ziel ist",
		atrWithPOS("Start", "SUB:NOM:SIN:MAS", "Start"),
		atrWithPOS("und", "KON:NEB", "und"),
		ziel,
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("Innsbruck", "EIG:NOM:SIN:NEU", "Innsbruck"),
	)

	// Java testRuleWithCorrectPluralVerb morph samples
	// Julia und Karsten sind alt. — last subject NPP + plural verb
	karsten := atrWithPOS("Karsten", "EIG:NOM:SIN:MAS", "Karsten")
	nppSubject(karsten)
	assertGood("Julia und Karsten sind",
		atrWithPOS("Julia", "EIG:NOM:SIN:FEM", "Julia"),
		atrWithPOS("und", "KON:NEB", "und"),
		karsten,
		atrWithPOS("sind", "VER:3:PLU:PRÄ:NON", "sein"),
		atrWithPOS("alt", "ADJ:PRD:GRU", "alt"),
		atrWithPOS(".", "PKT", "."),
	)
	// Bob und Tom sind Brüder.
	tom := atrWithPOS("Tom", "EIG:NOM:SIN:MAS", "Tom")
	nppSubject(tom)
	assertGood("Bob und Tom sind",
		atrWithPOS("Bob", "EIG:NOM:SIN:MAS", "Bob"),
		atrWithPOS("und", "KON:NEB", "und"),
		tom,
		atrWithPOS("sind", "VER:3:PLU:PRÄ:NON", "sein"),
		atrWithPOS("Brüder", "SUB:NOM:PLU:MAS", "Bruder"),
		atrWithPOS(".", "PKT", "."),
	)
	// Die USA sind …
	usa := atrWithPOS("USA", "EIG:NOM:PLU:NEU", "USA")
	nppSubject(usa)
	assertGood("Die USA sind",
		atrWithPOS("Die", "ART:DEF:NOM:PLU:NEU", "die"),
		usa,
		atrWithPOS("sind", "VER:3:PLU:PRÄ:NON", "sein"),
		atrWithPOS("ein", "ART:IND:NOM:SIN:MAS", "ein"),
		atrWithPOS("Staat", "SUB:NOM:SIN:MAS", "Staat"),
		atrWithPOS(".", "PKT", "."),
	)
	// Hundert Dollar sind doch gar nichts!
	dollar := atrWithPOS("Dollar", "SUB:NOM:PLU:MAS", "Dollar")
	nppSubject(dollar)
	assertGood("Hundert Dollar sind",
		atrWithPOS("Hundert", "ZAL", "hundert"),
		dollar,
		atrWithPOS("sind", "VER:3:PLU:PRÄ:NON", "sein"),
		atrWithPOS("doch", "ADV", "doch"),
		atrWithPOS("gar", "ADV", "gar"),
		atrWithPOS("nichts", "PIS", "nichts"),
		atrWithPOS("!", "PKT", "!"),
	)
	// Einzelne Atome sind klein.
	atome := atrWithPOS("Atome", "SUB:NOM:PLU:NEU", "Atom")
	nppSubject(atome)
	assertGood("Einzelne Atome sind",
		atrWithPOS("Einzelne", "ADJ:NOM:PLU:NEU:GRU:IND", "einzeln"),
		atome,
		atrWithPOS("sind", "VER:3:PLU:PRÄ:NON", "sein"),
		atrWithPOS("klein", "ADJ:PRD:GRU", "klein"),
		atrWithPOS(".", "PKT", "."),
	)

	// Java testRuleWithCorrectSingularAndPluralVerb — both SIN and PLU OK
	// Personen ist/sind der Zugriff …
	personen := atrWithPOS("Personen", "SUB:DAT:PLU:FEM", "Person")
	// DAT: no nominative chunk → singular/plural verb path skipped
	npsSubject(personen) // DAT tag means prevChunkIsNominative false
	assertGood("Personen ist Zugriff",
		personen,
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("der", "ART:DEF:NOM:SIN:MAS", "der"),
		atrWithPOS("Zugriff", "SUB:NOM:SIN:MAS", "Zugriff"),
		atrWithPOS(".", "PKT", "."),
	)
	// 80 Cent ist/sind — measure; inject NPS singular-ish
	cent := atrWithPOS("Cent", "SUB:NOM:PLU:MAS", "Cent")
	nppSubject(cent)
	// Plural subject + plural verb OK
	assertGood("80 Cent sind",
		atrWithPOS("80", "ZAL", "80"),
		cent,
		atrWithPOS("sind", "VER:3:PLU:PRÄ:NON", "sein"),
		atrWithPOS("nicht", "ADV", "nicht"),
		atrWithPOS("genug", "ADV", "genug"),
		atrWithPOS(".", "PKT", "."),
	)
}

// Twin remaining incorrect plural morph: Julia, Heike und Karsten ist / Herr Karsten Schröder sind
func TestSubjectVerbAgreementRule_IncorrectPluralVerb_MoreJava(t *testing.T) {
	rule := NewSubjectVerbAgreementRule(nil)
	assertBad := func(label, want string, toks ...*languagetool.AnalyzedTokenReadings) {
		t.Helper()
		all := append([]*languagetool.AnalyzedTokenReadings{sentStartATR()}, toks...)
		ms := rule.Match(languagetool.NewAnalyzedSentence(withPositions(all...)))
		require.GreaterOrEqual(t, len(ms), 1, "bad %s", label)
		require.Contains(t, ms[0].GetSuggestedReplacements(), want, label)
	}
	// Julia, Heike und Karsten ist alt. — coordinated NPP + singular verb
	karsten := atrWithPOS("Karsten", "EIG:NOM:SIN:MAS", "Karsten")
	nppSubject(karsten)
	assertBad("Julia, Heike und Karsten ist", "sind",
		atrWithPOS("Julia", "EIG:NOM:SIN:FEM", "Julia"),
		atrWithPOS(",", "PKT", ","),
		atrWithPOS("Heike", "EIG:NOM:SIN:FEM", "Heike"),
		atrWithPOS("und", "KON:NEB", "und"),
		karsten,
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("alt", "ADJ:PRD:GRU", "alt"),
		atrWithPOS(".", "PKT", "."),
	)
	// Herr Karsten Schröder sind alt.
	schroeder := atrWithPOS("Schröder", "EIG:NOM:SIN:MAS", "Schröder")
	npsSubject(schroeder)
	assertBad("Herr Karsten Schröder sind", "ist",
		atrWithPOS("Herr", "SUB:NOM:SIN:MAS", "Herr"),
		atrWithPOS("Karsten", "EIG:NOM:SIN:MAS", "Karsten"),
		schroeder,
		atrWithPOS("sind", "VER:3:PLU:PRÄ:NON", "sein"),
		atrWithPOS("alt", "ADJ:PRD:GRU", "alt"),
		atrWithPOS(".", "PKT", "."),
	)
}
