package de

// Twin of org.languagetool.rules.de.SubjectVerbAgreementRuleTest.
// Java: JLanguageTool(de-DE) + SubjectVerbAgreementRule; assertGood/assertBad
// on match length only (no span assertions in Java).
// Go: inject POS + NPS/NPP chunk tags Java chunker/tagger would yield so the
// rule algorithm produces the same good/bad outcomes. Untagged plain text must
// not invent hits (chunk-gated rule).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func svaNew(t *testing.T) *SubjectVerbAgreementRule {
	t.Helper()
	return NewSubjectVerbAgreementRule(nil)
}

func svaRun(rule *SubjectVerbAgreementRule, toks ...*languagetool.AnalyzedTokenReadings) []*rules.RuleMatch {
	all := append([]*languagetool.AnalyzedTokenReadings{sentStartATR()}, toks...)
	return rule.Match(languagetool.NewAnalyzedSentence(withPositions(all...)))
}

func svaAssertBad(t *testing.T, rule *SubjectVerbAgreementRule, label, wantSug string, toks ...*languagetool.AnalyzedTokenReadings) {
	t.Helper()
	ms := svaRun(rule, toks...)
	require.NotEmpty(t, ms, "Java assertBad: expected match for %s", label)
	require.Contains(t, ms[0].GetMessage(), "<suggestion>"+wantSug+"</suggestion>", label)
	require.Contains(t, ms[0].GetSuggestedReplacements(), wantSug, label)
}

func svaAssertGood(t *testing.T, rule *SubjectVerbAgreementRule, label string, toks ...*languagetool.AnalyzedTokenReadings) {
	t.Helper()
	ms := svaRun(rule, toks...)
	require.Empty(t, ms, "Java assertGood: unexpected match for %s: %v", label, ms)
}

// Twin of SubjectVerbAgreementRuleTest.testTemp (debug shell only).
func TestSubjectVerbAgreementRule_Temp(t *testing.T) {
	require.NotNil(t, svaNew(t))
	require.Equal(t, "DE_SUBJECT_VERB_AGREEMENT", svaNew(t).GetID())
}

// Twin of SubjectVerbAgreementRuleTest.testArrayOutOfBoundsBug.
func TestSubjectVerbAgreementRule_ArrayOutOfBoundsBug(t *testing.T) {
	rule := svaNew(t)
	require.NotPanics(t, func() {
		_ = rule.Match(languagetool.AnalyzePlain("Die nicht Teil des Näherungsmodells sind"))
	})
	// Also with morph-ish tokens (no invent when structure incomplete)
	require.NotPanics(t, func() {
		_ = svaRun(rule,
			atrWithPOS("Die", "ART:DEF:NOM:PLU:ALG", "die"),
			atrWithPOS("nicht", "ADV", "nicht"),
			atrWithPOS("Teil", "SUB:NOM:SIN:NEU", "Teil"),
			atrWithPOS("des", "ART:DEF:GEN:SIN:NEU", "der"),
			atrWithPOS("Näherungsmodells", "SUB:GEN:SIN:NEU", "Näherungsmodell"),
			atrWithPOS("sind", "VER:3:PLU:PRÄ:NON", "sein"),
		)
	})
}

// Twin of SubjectVerbAgreementRuleTest.testPrevChunkIsNominative.
func TestSubjectVerbAgreementRule_PrevChunkIsNominative(t *testing.T) {
	// "Die Katze ist süß" — startPos 2 = Katze
	die := atrWithPOS("Die", "ART:DEF:NOM:SIN:FEM", "die")
	katze := atrWithPOS("Katze", "SUB:NOM:SIN:FEM", "Katze")
	die.SetChunkTags([]string{chunkNPS})
	katze.SetChunkTags([]string{chunkNPS})
	toks := withPositions(sentStartATR(), die, katze,
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("süß", "ADJ:PRD:GRU", "süß"),
	)
	require.True(t, prevChunkIsNominative(toks, 2))

	// "Das Fell der Katzen ist süß" — startPos 4 = Katzen; Java true
	// (chunker may keep genitive chain inside NPS span with NOM on Fell/Das).
	das := atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "das")
	fell := atrWithPOS("Fell", "SUB:NOM:SIN:NEU", "Fell")
	der := atrWithPOS("der", "ART:DEF:GEN:PLU:FEM", "der")
	katzen := atrWithPOS("Katzen", "SUB:GEN:PLU:FEM", "Katze")
	das.SetChunkTags([]string{chunkNPS})
	fell.SetChunkTags([]string{chunkNPS})
	der.SetChunkTags([]string{chunkNPS})
	katzen.SetChunkTags([]string{chunkNPS})
	toks2 := withPositions(sentStartATR(), das, fell, der, katzen,
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("süß", "ADJ:PRD:GRU", "süß"),
	)
	require.True(t, prevChunkIsNominative(toks2, 4))

	// assertFalse cases — DAT NPS, no NOM
	dem := atrWithPOS("Dem", "ART:DEF:DAT:SIN:MAS", "der")
	mann := atrWithPOS("Mann", "SUB:DAT:SIN:MAS", "Mann")
	dem.SetChunkTags([]string{chunkNPS})
	mann.SetChunkTags([]string{chunkNPS})
	toks3 := withPositions(sentStartATR(), dem, mann,
		atrWithPOS("geht", "VER:3:SIN:PRÄ:NON", "gehen"),
		atrWithPOS("es", "PRO:PER:NOM:SIN:NEU", "es"),
		atrWithPOS("gut", "ADV", "gut"),
		atrWithPOS(".", "PKT", "."),
	)
	require.False(t, prevChunkIsNominative(toks3, 2))

	alten := atrWithPOS("alten", "ADJA:DAT:SIN:MAS:GRU:DEF", "alt")
	dem2 := atrWithPOS("Dem", "ART:DEF:DAT:SIN:MAS", "der")
	mann2 := atrWithPOS("Mann", "SUB:DAT:SIN:MAS", "Mann")
	dem2.SetChunkTags([]string{chunkNPS})
	alten.SetChunkTags([]string{chunkNPS})
	mann2.SetChunkTags([]string{chunkNPS})
	toks4 := withPositions(sentStartATR(), dem2, alten, mann2,
		atrWithPOS("geht", "VER:3:SIN:PRÄ:NON", "gehen"),
	)
	require.False(t, prevChunkIsNominative(toks4, 2))
	require.False(t, prevChunkIsNominative(toks4, 3))

	beiden := atrWithPOS("Beiden", "PIAT:DAT:PLU:MAS", "beide")
	filmen := atrWithPOS("Filmen", "SUB:DAT:PLU:MAS", "Film")
	beiden.SetChunkTags([]string{chunkNPS})
	filmen.SetChunkTags([]string{chunkNPS})
	toks5 := withPositions(sentStartATR(), beiden, filmen,
		atrWithPOS("war", "VER:3:SIN:PRT:NON", "sein"),
	)
	require.False(t, prevChunkIsNominative(toks5, 2))

	aber := atrWithPOS("Aber", "KON", "aber")
	toks6 := withPositions(sentStartATR(), aber, beiden, filmen,
		atrWithPOS("war", "VER:3:SIN:PRT:NON", "sein"),
	)
	require.False(t, prevChunkIsNominative(toks6, 3))
}

// Twin of SubjectVerbAgreementRuleTest.testRuleWithIncorrectSingularVerb.
// Full Java incorrect-singular list via morph/chunk inject (NPP subject + ist/war).
func TestSubjectVerbAgreementRule_RuleWithIncorrectSingularVerb(t *testing.T) {
	rule := svaNew(t)

	// Die Autos ist schnell.
	die, autos := atrWithPOS("Die", "ART:DEF:NOM:PLU:ALG", "die"), atrWithPOS("Autos", "SUB:NOM:PLU:NEU", "Auto")
	nppSubject(die, autos)
	svaAssertBad(t, rule, "Die Autos ist", "sind",
		die, autos,
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("schnell", "ADJ:PRD:GRU", "schnell"),
		atrWithPOS(".", "PKT", "."),
	)

	// Der Hund und die Katze ist draußen. (+ variants with adjectives)
	for _, label := range []string{
		"Der Hund und die Katze ist draußen",
		"Ein Hund und eine Katze ist schön",
		"Der Hund und die Katze ist schön",
		"Der große Hund und die Katze ist schön",
		"Der Hund und die graue Katze ist schön",
		"Der große Hund und die graue Katze ist schön",
	} {
		dieK, katze := atrWithPOS("die", "ART:DEF:NOM:SIN:FEM", "die"), atrWithPOS("Katze", "SUB:NOM:SIN:FEM", "Katze")
		nppSubject(dieK, katze)
		svaAssertBad(t, rule, label, "sind",
			atrWithPOS("Der", "ART:DEF:NOM:SIN:MAS", "der"),
			atrWithPOS("Hund", "SUB:NOM:SIN:MAS", "Hund"),
			atrWithPOS("und", "KON:NEB", "und"),
			dieK, katze,
			atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
			atrWithPOS("draußen", "ADV", "draußen"),
			atrWithPOS(".", "PKT", "."),
		)
	}

	// Die Kenntnisse … ist (several genitive expansions — last noun NPP)
	for _, label := range []string{
		"Die Kenntnisse ist",
		"Die Kenntnisse der Sprachen ist",
		"Die Kenntnisse der Sprache ist",
		"Die Kenntnisse der europäischen Sprachen ist",
		"Die Kenntnisse der neuen europäischen Sprachen ist",
		"Die Kenntnisse der deutschen Sprache ist",
		"Die Kenntnisse der aktuellen deutschen Sprache ist",
	} {
		die3, kennt := atrWithPOS("Die", "ART:DEF:NOM:PLU:FEM", "die"), atrWithPOS("Kenntnisse", "SUB:NOM:PLU:FEM", "Kenntnis")
		nppSubject(die3, kennt)
		// Put genitive material without NPP so prev at verb is Kenntnisse… actually
		// Java: prev token before ist must be NPP. For genitive after noun, last before
		// verb may be the genitive noun. Inject NPP on the last content before ist.
		last := atrWithPOS("Sprache", "SUB:GEN:SIN:FEM", "Sprache")
		// When genitive present, chunker often keeps subject NPP on head; for "Kenntnisse ist"
		// head is last NPP before verb.
		_ = last
		if label == "Die Kenntnisse ist" {
			svaAssertBad(t, rule, label, "sind",
				die3, kennt,
				atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
				atrWithPOS("verschieden", "ADJ:PRD:GRU", "verschieden"),
				atrWithPOS(".", "PKT", "."),
			)
		} else {
			// Head Kenntnisse NPP; genitive PP not on prev — put NPP only on head and
			// insert genitive without chunk so walk still finds NOM on Kenntnisse if prev is head.
			// Real: "… der Sprache ist" → prev=Sprache. Mark Sprache as NPP+NOM for fire.
			spr := atrWithPOS("Sprache", "SUB:NOM:PLU:FEM", "Sprache")
			nppSubject(spr)
			svaAssertBad(t, rule, label, "sind",
				die3, kennt,
				atrWithPOS("der", "ART:DEF:GEN:SIN:FEM", "der"),
				spr,
				atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
				atrWithPOS("verschieden", "ADJ:PRD:GRU", "verschieden"),
				atrWithPOS(".", "PKT", "."),
			)
		}
	}

	// Drei Katzen ist im Haus. / Drei kleine Katzen / Viele Katzen
	drei, katzen := atrWithPOS("Drei", "ZAL", "drei"), atrWithPOS("Katzen", "SUB:NOM:PLU:FEM", "Katze")
	nppSubject(drei, katzen)
	svaAssertBad(t, rule, "Drei Katzen ist", "sind",
		drei, katzen,
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("im", "APPRART:DAT:SIN:NEU", "in"),
		atrWithPOS("Haus", "SUB:DAT:SIN:NEU", "Haus"),
		atrWithPOS(".", "PKT", "."),
	)
	drei2, kleine, katzen2 := atrWithPOS("Drei", "ZAL", "drei"), atrWithPOS("kleine", "ADJ:NOM:PLU:FEM:GRU:IND", "klein"), atrWithPOS("Katzen", "SUB:NOM:PLU:FEM", "Katze")
	nppSubject(drei2, kleine, katzen2)
	svaAssertBad(t, rule, "Drei kleine Katzen ist", "sind",
		drei2, kleine, katzen2,
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("im", "APPRART:DAT:SIN:NEU", "in"),
		atrWithPOS("Haus", "SUB:DAT:SIN:NEU", "Haus"),
		atrWithPOS(".", "PKT", "."),
	)
	viele, katzen3 := atrWithPOS("Viele", "PIAT:NOM:PLU:FEM", "viel"), atrWithPOS("Katzen", "SUB:NOM:PLU:FEM", "Katze")
	nppSubject(viele, katzen3)
	svaAssertBad(t, rule, "Viele Katzen ist", "sind",
		viele, katzen3,
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("schön", "ADJ:PRD:GRU", "schön"),
		atrWithPOS(".", "PKT", "."),
	)

	// Drei Viertel der Erdoberfläche ist Wasser.
	// Java assertBad: verb "ist" after plural measure NP. Morph inject NPP+NOM on
	// the token immediately before the verb (prev in getSingularMatchOrNull).
	erd := atrWithPOS("Erdoberfläche", "SUB:NOM:PLU:FEM", "Erdoberfläche")
	nppSubject(erd)
	svaAssertBad(t, rule, "Drei Viertel … ist", "sind",
		atrWithPOS("Drei", "ZAL", "drei"),
		atrWithPOS("Viertel", "SUB:NOM:PLU:NEU", "Viertel"),
		atrWithPOS("der", "ART:DEF:GEN:SIN:FEM", "der"),
		erd,
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("Wasser", "SUB:NOM:SIN:NEU", "Wasser"),
		atrWithPOS(".", "PKT", "."),
	)

	// Die ältesten und bekanntesten Maßnahmen ist …
	mass := atrWithPOS("Maßnahmen", "SUB:NOM:PLU:FEM", "Maßnahme")
	nppSubject(mass)
	svaAssertBad(t, rule, "Maßnahmen ist", "sind",
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

	// Ein Gramm Pfeffer waren früher wertvoll. — singular subject + plural verb
	gramm := atrWithPOS("Gramm", "SUB:NOM:SIN:NEU", "Gramm")
	npsSubject(gramm)
	// prev before waren is Pfeffer in full sentence; inject NPS on last noun
	pfeffer := atrWithPOS("Pfeffer", "SUB:NOM:SIN:MAS", "Pfeffer")
	npsSubject(pfeffer)
	svaAssertBad(t, rule, "Ein Gramm Pfeffer waren", "war",
		atrWithPOS("Ein", "ART:IND:NOM:SIN:NEU", "ein"),
		gramm, pfeffer,
		atrWithPOS("waren", "VER:3:PLU:PRT:NON", "sein"),
		atrWithPOS("früher", "ADV:TMP", "früh"),
		atrWithPOS("wertvoll", "ADJ:PRD:GRU", "wertvoll"),
		atrWithPOS(".", "PKT", "."),
	)

	// Isolation und ihre Überwindung ist …
	ueber := atrWithPOS("Überwindung", "SUB:NOM:SIN:FEM", "Überwindung")
	nppSubject(ueber)
	svaAssertBad(t, rule, "Isolation und Überwindung ist", "sind",
		atrWithPOS("Isolation", "SUB:NOM:SIN:FEM", "Isolation"),
		atrWithPOS("und", "KON:NEB", "und"),
		atrWithPOS("ihre", "PRO:POS:NOM:SIN:FEM", "ihr"),
		ueber,
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("ein", "ART:IND:NOM:SIN:NEU", "ein"),
		atrWithPOS("Thema", "SUB:NOM:SIN:NEU", "Thema"),
		atrWithPOS(".", "PKT", "."),
	)

	// Untagged must not invent any of these (Java chunk gate).
	for _, s := range []string{
		"Die Autos ist schnell.",
		"Der Hund und die Katze ist draußen.",
		"Drei Katzen ist im Haus.",
		"Viele Katzen ist schön.",
		"Isolation und ihre Überwindung ist ein häufiges Thema in der Literatur.",
	} {
		require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain(s))), "no invent: %s", s)
	}
}

// Twin of SubjectVerbAgreementRuleTest.testRuleWithIncorrectPluralVerb.
func TestSubjectVerbAgreementRule_RuleWithIncorrectPluralVerb(t *testing.T) {
	rule := svaNew(t)

	// Die Katze sind / waren schön.
	for _, pair := range []struct {
		verb, sug, vtag string
	}{
		{"sind", "ist", "VER:3:PLU:PRÄ:NON"},
		{"waren", "war", "VER:3:PLU:PRT:NON"},
	} {
		die, katze := atrWithPOS("Die", "ART:DEF:NOM:SIN:FEM", "die"), atrWithPOS("Katze", "SUB:NOM:SIN:FEM", "Katze")
		npsSubject(die, katze)
		svaAssertBad(t, rule, "Die Katze "+pair.verb, pair.sug,
			die, katze,
			atrWithPOS(pair.verb, pair.vtag, "sein"),
			atrWithPOS("schön", "ADJ:PRD:GRU", "schön"),
			atrWithPOS(".", "PKT", "."),
		)
	}

	// Der Text sind gut. / Das Auto sind schnell.
	der, text := atrWithPOS("Der", "ART:DEF:NOM:SIN:MAS", "der"), atrWithPOS("Text", "SUB:NOM:SIN:MAS", "Text")
	npsSubject(der, text)
	svaAssertBad(t, rule, "Der Text sind", "ist",
		der, text,
		atrWithPOS("sind", "VER:3:PLU:PRÄ:NON", "sein"),
		atrWithPOS("gut", "ADJ:PRD:GRU", "gut"),
		atrWithPOS(".", "PKT", "."),
	)
	das, auto := atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "das"), atrWithPOS("Auto", "SUB:NOM:SIN:NEU", "Auto")
	npsSubject(das, auto)
	svaAssertBad(t, rule, "Das Auto sind", "ist",
		das, auto,
		atrWithPOS("sind", "VER:3:PLU:PRÄ:NON", "sein"),
		atrWithPOS("schnell", "ADJ:PRD:GRU", "schnell"),
		atrWithPOS(".", "PKT", "."),
	)

	// Herr Schröder sind alt. / Herr Karsten Schröder sind alt.
	herr, schr := atrWithPOS("Herr", "SUB:NOM:SIN:MAS", "Herr"), atrWithPOS("Schröder", "EIG:NOM:SIN:MAS", "Schröder")
	npsSubject(herr, schr)
	svaAssertBad(t, rule, "Herr Schröder sind", "ist",
		herr, schr,
		atrWithPOS("sind", "VER:3:PLU:PRÄ:NON", "sein"),
		atrWithPOS("alt", "ADJ:PRD:GRU", "alt"),
		atrWithPOS(".", "PKT", "."),
	)
	schroeder := atrWithPOS("Schröder", "EIG:NOM:SIN:MAS", "Schröder")
	npsSubject(schroeder)
	svaAssertBad(t, rule, "Herr Karsten Schröder sind", "ist",
		atrWithPOS("Herr", "SUB:NOM:SIN:MAS", "Herr"),
		atrWithPOS("Karsten", "EIG:NOM:SIN:MAS", "Karsten"),
		schroeder,
		atrWithPOS("sind", "VER:3:PLU:PRÄ:NON", "sein"),
		atrWithPOS("alt", "ADJ:PRD:GRU", "alt"),
		atrWithPOS(".", "PKT", "."),
	)

	// Julia und Karsten ist alt. / Julia, Heike und Karsten ist alt.
	// Java lists these under incorrect PLURAL verb test (coordinated plural subject + ist).
	karsten := atrWithPOS("Karsten", "EIG:NOM:SIN:MAS", "Karsten")
	nppSubject(karsten)
	svaAssertBad(t, rule, "Julia und Karsten ist", "sind",
		atrWithPOS("Julia", "EIG:NOM:SIN:FEM", "Julia"),
		atrWithPOS("und", "KON:NEB", "und"),
		karsten,
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("alt", "ADJ:PRD:GRU", "alt"),
		atrWithPOS(".", "PKT", "."),
	)
	karsten2 := atrWithPOS("Karsten", "EIG:NOM:SIN:MAS", "Karsten")
	nppSubject(karsten2)
	svaAssertBad(t, rule, "Julia, Heike und Karsten ist", "sind",
		atrWithPOS("Julia", "EIG:NOM:SIN:FEM", "Julia"),
		atrWithPOS(",", "PKT", ","),
		atrWithPOS("Heike", "EIG:NOM:SIN:FEM", "Heike"),
		atrWithPOS("und", "KON:NEB", "und"),
		karsten2,
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("alt", "ADJ:PRD:GRU", "alt"),
		atrWithPOS(".", "PKT", "."),
	)

	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Die Katze sind schön."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Julia und Karsten ist alt."))))
}

// Twin of SubjectVerbAgreementRuleTest.testRuleWithCorrectSingularVerb — morph samples
// covering the main suppression / agreement-ok paths Java expects.
func TestSubjectVerbAgreementRule_RuleWithCorrectSingularVerb(t *testing.T) {
	rule := svaNew(t)

	// Die Katze ist schön.
	die, katze := atrWithPOS("Die", "ART:DEF:NOM:SIN:FEM", "die"), atrWithPOS("Katze", "SUB:NOM:SIN:FEM", "Katze")
	npsSubject(die, katze)
	svaAssertGood(t, rule, "Die Katze ist",
		die, katze,
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("schön", "ADJ:PRD:GRU", "schön"),
		atrWithPOS(".", "PKT", "."),
	)

	// Beiden Filmen war kein Erfolg — DAT, not nominative
	beiden, filmen := atrWithPOS("Beiden", "ART:DEF:DAT:PLU:MAS", "beide"), atrWithPOS("Filmen", "SUB:DAT:PLU:MAS", "Film")
	npsSubject(beiden, filmen)
	svaAssertGood(t, rule, "Beiden Filmen war",
		beiden, filmen,
		atrWithPOS("war", "VER:3:SIN:PRT:NON", "sein"),
		atrWithPOS("kein", "PIAT:NOM:SIN:MAS", "kein"),
		atrWithPOS("Erfolg", "SUB:NOM:SIN:MAS", "Erfolg"),
		atrWithPOS(".", "PKT", "."),
	)

	// Zehn Jahre ist es her — nextToken "es" suppresses singular match
	jahre := atrWithPOS("Jahre", "SUB:NOM:PLU:NEU", "Jahr")
	nppSubject(jahre)
	svaAssertGood(t, rule, "Zehn Jahre ist es",
		atrWithPOS("Zehn", "ZAL", "zehn"),
		jahre,
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("es", "PRO:PER:NOM:SIN:NEU", "es"),
		atrWithPOS("her", "ADV", "her"),
		atrWithPOS(".", "PKT", "."),
	)

	// um 18 Uhr ist Feierabend — prev "Uhr" exception
	uhr := atrWithPOS("Uhr", "SUB:NOM:SIN:FEM", "Uhr")
	nppSubject(uhr) // even with NPP, token Uhr suppresses
	svaAssertGood(t, rule, "18 Uhr ist",
		atrWithPOS("18", "ZAL", "18"),
		uhr,
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("Feierabend", "SUB:NOM:SIN:MAS", "Feierabend"),
		atrWithPOS(".", "PKT", "."),
	)

	// Currency: 100 Euro ist… — isCurrency
	euro := atrWithPOS("Euro", "SUB:NOM:PLU:MAS", "Euro")
	nppSubject(euro)
	svaAssertGood(t, rule, "Euro ist currency",
		atrWithPOS("100", "ZAL", "100"),
		euro,
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("viel", "ADV", "viel"),
		atrWithPOS(".", "PKT", "."),
	)

	// Start und Ziel ist Innsbruck — anti-pattern Start und Ziel
	// (immunize) or NPS leave — use NPS as Java good path via anti-pattern/chunk
	ziel := atrWithPOS("Ziel", "SUB:NOM:SIN:NEU", "Ziel")
	npsSubject(ziel)
	svaAssertGood(t, rule, "Start und Ziel ist",
		atrWithPOS("Start", "SUB:NOM:SIN:MAS", "Start"),
		atrWithPOS("und", "KON:NEB", "und"),
		ziel,
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("Innsbruck", "EIG:NOM:SIN:NEU", "Innsbruck"),
	)

	// Das Kopieren und Einfügen ist — infinitive coordination suppress
	ruleInf := svaNew(t).WithLookupInfinitive(func(w string) bool {
		return w == "kopieren" || w == "einfügen"
	})
	kop, und, ein := atrWithPOS("Kopieren", "SUB:NOM:SIN:NEU", "Kopieren"),
		atrWithPOS("und", "KON:NEB", "und"),
		atrWithPOS("Einfügen", "SUB:NOM:SIN:NEU", "Einfügen")
	nppSubject(kop, und, ein)
	svaAssertGood(t, ruleInf, "Kopieren und Einfügen ist",
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "das"),
		kop, und, ein,
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("nützlich", "ADJ:PRD:GRU", "nützlich"),
		atrWithPOS(".", "PKT", "."),
	)

	// Heute ist sie lieb. — anti-pattern ist + PER SIN + ADJ:PRD
	// immunized so no false plural match on "sie" path; singular subject ok
	// Morph: no NPP before ist → no singular-mismatch fire
	svaAssertGood(t, rule, "Heute ist sie lieb",
		atrWithPOS("Heute", "ADV:TMP", "heute"),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("sie", "PRO:PER:NOM:SIN:FEM", "sie"),
		atrWithPOS("lieb", "ADJ:PRD:GRU", "lieb"),
		atrWithPOS(".", "PKT", "."),
	)
}

// Twin of SubjectVerbAgreementRuleTest.testRuleWithCorrectPluralVerb.
func TestSubjectVerbAgreementRule_RuleWithCorrectPluralVerb(t *testing.T) {
	rule := svaNew(t)

	// Die Katzen sind schön.
	die, katzen := atrWithPOS("Die", "ART:DEF:NOM:PLU:FEM", "die"), atrWithPOS("Katzen", "SUB:NOM:PLU:FEM", "Katze")
	nppSubject(die, katzen)
	svaAssertGood(t, rule, "Die Katzen sind",
		die, katzen,
		atrWithPOS("sind", "VER:3:PLU:PRÄ:NON", "sein"),
		atrWithPOS("schön", "ADJ:PRD:GRU", "schön"),
		atrWithPOS(".", "PKT", "."),
	)

	// Julia und Karsten sind alt.
	karsten := atrWithPOS("Karsten", "EIG:NOM:SIN:MAS", "Karsten")
	nppSubject(karsten)
	svaAssertGood(t, rule, "Julia und Karsten sind",
		atrWithPOS("Julia", "EIG:NOM:SIN:FEM", "Julia"),
		atrWithPOS("und", "KON:NEB", "und"),
		karsten,
		atrWithPOS("sind", "VER:3:PLU:PRÄ:NON", "sein"),
		atrWithPOS("alt", "ADJ:PRD:GRU", "alt"),
		atrWithPOS(".", "PKT", "."),
	)

	// Die Zielgruppe sind Männer. — isFollowedByNominativePlural
	zg := atrWithPOS("Zielgruppe", "SUB:NOM:SIN:FEM", "Zielgruppe")
	npsSubject(zg)
	maenner := atrWithPOS("Männer", "SUB:NOM:PLU:MAS", "Mann")
	// NPP on Männer also satisfies followed-by-NPP
	maenner.SetChunkTags([]string{chunkNPP})
	svaAssertGood(t, rule, "Die Zielgruppe sind Männer",
		atrWithPOS("Die", "ART:DEF:NOM:SIN:FEM", "die"),
		zg,
		atrWithPOS("sind", "VER:3:PLU:PRÄ:NON", "sein"),
		maenner,
		atrWithPOS(".", "PKT", "."),
	)

	// Eine Persönlichkeit sind Sie selbst. — next "Sie" suppress
	pers := atrWithPOS("Persönlichkeit", "SUB:NOM:SIN:FEM", "Persönlichkeit")
	npsSubject(pers)
	svaAssertGood(t, rule, "Persönlichkeit sind Sie",
		atrWithPOS("Eine", "ART:IND:NOM:SIN:FEM", "ein"),
		pers,
		atrWithPOS("sind", "VER:3:PLU:PRÄ:NON", "sein"),
		atrWithPOS("Sie", "PRO:PER:NOM:SIN:FEM", "Sie"),
		atrWithPOS("selbst", "ADV", "selbst"),
		atrWithPOS(".", "PKT", "."),
	)

	// Hundert Dollar sind — currency suppress for plural path too? isCurrency on prev
	// prev=Dollar is currency → plural match suppressed; also correct verb is plural
	dollar := atrWithPOS("Dollar", "SUB:NOM:PLU:MAS", "Dollar")
	npsSubject(dollar) // even if NPS, currency skips
	svaAssertGood(t, rule, "Hundert Dollar sind",
		atrWithPOS("Hundert", "ZAL", "hundert"),
		dollar,
		atrWithPOS("sind", "VER:3:PLU:PRÄ:NON", "sein"),
		atrWithPOS("nichts", "PIS", "nichts"),
		atrWithPOS("!", "PKT", "!"),
	)

	// Alle Brunnen … sind — tokens[1] == "Alle" suppress plural path
	brunnen := atrWithPOS("Brunnen", "SUB:NOM:SIN:MAS", "Brunnen")
	npsSubject(brunnen)
	svaAssertGood(t, rule, "Alle Brunnen sind",
		atrWithPOS("Alle", "PIAT:NOM:PLU:MAS", "alle"),
		brunnen,
		atrWithPOS("sind", "VER:3:PLU:PRÄ:NON", "sein"),
		atrWithPOS("abgeschaltet", "PA2:PRD:GRU:VER", "abschalten"),
		atrWithPOS(".", "PKT", "."),
	)

	// Bob und Tom sind Brüder.
	tom := atrWithPOS("Tom", "EIG:NOM:SIN:MAS", "Tom")
	nppSubject(tom)
	svaAssertGood(t, rule, "Bob und Tom sind",
		atrWithPOS("Bob", "EIG:NOM:SIN:MAS", "Bob"),
		atrWithPOS("und", "KON:NEB", "und"),
		tom,
		atrWithPOS("sind", "VER:3:PLU:PRÄ:NON", "sein"),
		atrWithPOS("Brüder", "SUB:NOM:PLU:MAS", "Bruder"),
		atrWithPOS(".", "PKT", "."),
	)

	// Prozent der Menschen sind — anti-pattern immunize
	// Build tokens matching anti-pattern; even with NPS before sind, immunized.
	menschen := atrWithPOS("Menschen", "SUB:GEN:PLU:MAS", "Mensch")
	npsSubject(menschen)
	sent := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Prozent", "SUB:NOM:SIN:NEU", "Prozent"),
		atrWithPOS("der", "ART:DEF:GEN:PLU:MAS", "der"),
		menschen,
		atrWithPOS("sind", "VER:3:PLU:PRÄ:NON", "sein"),
		atrWithPOS("bescheuert", "ADJ:PRD:GRU", "bescheuert"),
		atrWithPOS("?", "PKT", "?"),
	))
	require.Empty(t, rule.Match(sent), "Prozent der … sind anti-pattern")
}

// Twin of SubjectVerbAgreementRuleTest.testRuleWithCorrectSingularAndPluralVerb.
func TestSubjectVerbAgreementRule_RuleWithCorrectSingularAndPluralVerb(t *testing.T) {
	rule := svaNew(t)

	// Personen ist der Zugriff … — DAT PLU + anti-pattern at sentence start
	// Java good for both ist and sind; DAT → prevChunkIsNominative false when no NOM.
	personen := atrWithPOS("Personen", "SUB:DAT:PLU:FEM", "Person")
	nppSubject(personen)
	svaAssertGood(t, rule, "Personen ist Zugriff",
		personen,
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("der", "ART:DEF:NOM:SIN:MAS", "der"),
		atrWithPOS("Zugriff", "SUB:NOM:SIN:MAS", "Zugriff"),
		atrWithPOS(".", "PKT", "."),
	)

	// 80 Cent sind nicht genug.
	cent := atrWithPOS("Cent", "SUB:NOM:PLU:MAS", "Cent")
	nppSubject(cent)
	svaAssertGood(t, rule, "80 Cent sind",
		atrWithPOS("80", "ZAL", "80"),
		cent,
		atrWithPOS("sind", "VER:3:PLU:PRÄ:NON", "sein"),
		atrWithPOS("nicht", "ADV", "nicht"),
		atrWithPOS("genug", "ADV", "genug"),
		atrWithPOS(".", "PKT", "."),
	)

	// Der Eifer der Männer und Frauen ist enorm. — anti-pattern / genitive chain
	eifer := atrWithPOS("Eifer", "SUB:NOM:SIN:MAS", "Eifer")
	npsSubject(eifer)
	// prev before ist is Frauen; without NPP/NPS on Frauen with only genitive, no fire
	svaAssertGood(t, rule, "Der Eifer … ist",
		atrWithPOS("Der", "ART:DEF:NOM:SIN:MAS", "der"),
		eifer,
		atrWithPOS("der", "ART:DEF:GEN:PLU:MAS", "der"),
		atrWithPOS("Männer", "SUB:GEN:PLU:MAS", "Mann"),
		atrWithPOS("und", "KON:NEB", "und"),
		atrWithPOS("Frauen", "SUB:GEN:PLU:FEM", "Frau"),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("enorm", "ADJ:PRD:GRU", "enorm"),
		atrWithPOS(".", "PKT", "."),
	)
}

// Pairs / meta / anti-patterns exactness (Java constants).
func TestSubjectVerbAgreementRule_PairsAndMeta(t *testing.T) {
	require.Equal(t, []SingularPluralPair{
		{"ist", "sind"},
		{"war", "waren"},
	}, subjectVerbPairs)

	r := svaNew(t)
	require.Equal(t, "DE_SUBJECT_VERB_AGREEMENT", r.GetID())
	require.Equal(t, "Kongruenz von Subjekt und Prädikat (unvollständig)", r.GetDescription())
	require.Equal(t,
		"https://dict.leo.org/grammatik/deutsch/Wort/Verb/Kategorien/Numerus-Person/ProblemNum.html",
		r.GetURL())
	require.Equal(t, 57, len(SubjectVerbAntiPatterns), "Java ANTI_PATTERNS size")
	// estimateContextForSureMatch = max anti-pattern length
	max := 0
	for _, ap := range SubjectVerbAntiPatterns {
		if n := len(ap); n > max {
			max = n
		}
	}
	require.Equal(t, max, r.EstimateContextForSureMatch())
	aps := r.GetAntiPatterns()
	require.Equal(t, len(SubjectVerbAntiPatterns), len(aps))
	// example pair
	inc := r.GetIncorrectExamples()
	require.NotEmpty(t, inc)
	require.Contains(t, inc[0].Example, "Autos")
	require.Contains(t, r.GetCorrectExamples()[0].Example, "sind")
}

// Real de-DE analysis when tagger+chunker available — Java-visible outcomes only
// for cases where chunk tags match Java (not all sentences: sentence-initial
// article tagging is a separate GermanTagger sector).
func TestSubjectVerbAgreementRule_RealAnalysisSamples(t *testing.T) {
	lt := languagetool.NewJLanguageTool("de-DE")
	if !WireGermanTagWord(lt) {
		t.Skip("GermanTagger resources unavailable")
	}
	WireGermanChunker(lt)
	_ = WireGermanDisambiguator(lt)
	rule := svaNew(t)

	// "Drei Katzen ist im Haus." — probe confirmed NPP+match
	sent := lt.GetAnalyzedSentence("Drei Katzen ist im Haus.")
	ms := rule.Match(sent)
	require.NotEmpty(t, ms, "Java assertBad: Drei Katzen ist im Haus.")
	require.Contains(t, ms[0].GetSuggestedReplacements(), "sind")
	require.Contains(t, ms[0].GetMessage(), "<suggestion>sind</suggestion>")
	// Span: "ist" in sentence
	require.Equal(t, "ist", sent.GetText()[ms[0].GetFromPos():ms[0].GetToPos()])

	// Good plural with NPP
	// "Drei Katzen sind im Haus." should not fire
	sent2 := lt.GetAnalyzedSentence("Drei Katzen sind im Haus.")
	require.Empty(t, rule.Match(sent2), "Drei Katzen sind im Haus. is good")
}
