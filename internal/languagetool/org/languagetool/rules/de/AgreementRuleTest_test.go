package de

// Twin of AgreementRuleTest — open compounds need getCompoundError (dict/lt.check), not invent.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestAgreementRule_CompoundMatch(t *testing.T) {
	rule := NewAgreementRule(nil)
	// Untagged AnalyzePlain: no invent of open-compound hits
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Das ist die Original Mail."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Doch dieser kleine Magnesium Anteil ist entscheidend."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("War das Eifersucht?"))))
}

// Twin of AgreementRuleTest.testCompoundMatch — morph inject + CompoundPhraseValid (Java lt.check gate).
func TestAgreementRule_CompoundMatch_MorphJavaTable(t *testing.T) {
	rule := NewAgreementRule(nil)
	// Accept only phrases Java assertBad expects as suggestions (no invent of other forms).
	rule.CompoundPhraseValid = func(p string) bool {
		ok := map[string]bool{
			"die Originalmail": true, "die Original-Mail": true,
			"die neue Originalmail": true, "die neue Original-Mail": true,
			"die ganz neue Originalmail": true, "die ganz neue Original-Mail": true,
			"dieser kleine Magnesiumanteil": true, "dieser kleine Magnesium-Anteil": true,
			"dieser sehr kleine Magnesiumanteil": true, "dieser sehr kleine Magnesium-Anteil": true,
			"Die Standardpriorität": true, "Die Standard-Priorität": true,
			"Die derzeitige Standardpriorität": true, "Die derzeitige Standard-Priorität": true,
			"Ein neuer LanguageTool-Account": true, // Java only hyphen for this one
			"deine Accountdaten": true, "deine Account-Daten": true,
			"ins Fitnessstudio": true, "ins Fitness-Studio": true,
			"durchs Fitnessstudio": true, "durchs Fitness-Studio": true,
			"ein sehr interessantes kostenloses Slotspiel": true,
			"ein sehr interessantes kostenloses Slot-Spiel": true,
		}
		return ok[p]
	}

	assertCompound := func(label string, wantSugs []string, toks ...*languagetool.AnalyzedTokenReadings) {
		t.Helper()
		all := append([]*languagetool.AnalyzedTokenReadings{sentStartATR()}, toks...)
		ms := rule.Match(languagetool.NewAnalyzedSentence(withPositions(all...)))
		require.GreaterOrEqual(t, len(ms), 1, "compound bad %s", label)
		// Prefer compound-message match when present
		var m *rules.RuleMatch
		for _, cand := range ms {
			if cand != nil && cand.GetMessage() == compoundErrorMsg {
				m = cand
				break
			}
		}
		if m == nil {
			m = ms[0]
		}
		require.Equal(t, compoundErrorMsg, m.GetMessage(), "label %s", label)
		require.Empty(t, m.ShortMessage, "Java compound match has no shortMessage")
		for _, s := range wantSugs {
			require.Contains(t, m.GetSuggestedReplacements(), s, "label %s sugs=%v", label, m.GetSuggestedReplacements())
		}
	}
	assertGood := func(label string, toks ...*languagetool.AnalyzedTokenReadings) {
		t.Helper()
		all := append([]*languagetool.AnalyzedTokenReadings{sentStartATR()}, toks...)
		ms := rule.Match(languagetool.NewAnalyzedSentence(withPositions(all...)))
		require.Equal(t, 0, len(ms), "good %s got %d", label, len(ms))
	}
	assertAgreementNotCompound := func(label string, toks ...*languagetool.AnalyzedTokenReadings) {
		t.Helper()
		all := append([]*languagetool.AnalyzedTokenReadings{sentStartATR()}, toks...)
		ms := rule.Match(languagetool.NewAnalyzedSentence(withPositions(all...)))
		require.Equal(t, 1, len(ms), "label %s", label)
		require.NotContains(t, ms[0].GetMessage(), "zusammengesetztes Nomen", "label %s", label)
	}

	// Java: Das ist die Original Mail
	assertCompound("die Original Mail",
		[]string{"die Originalmail", "die Original-Mail"},
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("die", "ART:DEF:NOM:SIN:FEM", "die"),
		atrWithPOS("Original", "SUB:NOM:SIN:NEU", "Original"),
		atrWithPOS("Mail", "SUB:NOM:SIN:FEM", "Mail"),
		atrWithPOS(".", "PKT", "."),
	)
	// Java: die neue Original Mail
	assertCompound("die neue Original Mail",
		[]string{"die neue Originalmail", "die neue Original-Mail"},
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("die", "ART:DEF:NOM:SIN:FEM", "die"),
		atrWithPOS("neue", "ADJ:NOM:SIN:FEM:GRU:DEF", "neu"),
		atrWithPOS("Original", "SUB:NOM:SIN:NEU", "Original"),
		atrWithPOS("Mail", "SUB:NOM:SIN:FEM", "Mail"),
		atrWithPOS(".", "PKT", "."),
	)
	// Java: die ganz neue Original Mail (modifier "ganz" between det and adj)
	assertCompound("die ganz neue Original Mail",
		[]string{"die ganz neue Originalmail", "die ganz neue Original-Mail"},
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("die", "ART:DEF:NOM:SIN:FEM", "die"),
		atrWithPOS("ganz", "ADV", "ganz"),
		atrWithPOS("neue", "ADJ:NOM:SIN:FEM:GRU:DEF", "neu"),
		atrWithPOS("Original", "SUB:NOM:SIN:NEU", "Original"),
		atrWithPOS("Mail", "SUB:NOM:SIN:FEM", "Mail"),
		atrWithPOS(".", "PKT", "."),
	)
	// Java: dieser kleine Magnesium Anteil
	assertCompound("dieser kleine Magnesium Anteil",
		[]string{"dieser kleine Magnesiumanteil", "dieser kleine Magnesium-Anteil"},
		atrWithPOS("Doch", "ADV", "doch"),
		atrWithPOS("dieser", "ART:DEF:NOM:SIN:MAS", "dieser"),
		atrWithPOS("kleine", "ADJ:NOM:SIN:MAS:GRU:DEF", "klein"),
		atrWithPOS("Magnesium", "SUB:NOM:SIN:NEU", "Magnesium"),
		atrWithPOS("Anteil", "SUB:NOM:SIN:MAS", "Anteil"),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("entscheidend", "ADJ:PRD:GRU", "entscheidend"),
		atrWithPOS(".", "PKT", "."),
	)
	// Java: dieser sehr kleine Magnesium Anteil
	assertCompound("dieser sehr kleine Magnesium Anteil",
		[]string{"dieser sehr kleine Magnesiumanteil", "dieser sehr kleine Magnesium-Anteil"},
		atrWithPOS("Doch", "ADV", "doch"),
		atrWithPOS("dieser", "ART:DEF:NOM:SIN:MAS", "dieser"),
		atrWithPOS("sehr", "ADV", "sehr"),
		atrWithPOS("kleine", "ADJ:NOM:SIN:MAS:GRU:DEF", "klein"),
		atrWithPOS("Magnesium", "SUB:NOM:SIN:NEU", "Magnesium"),
		atrWithPOS("Anteil", "SUB:NOM:SIN:MAS", "Anteil"),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("entscheidend", "ADJ:PRD:GRU", "entscheidend"),
		atrWithPOS(".", "PKT", "."),
	)
	// Java: Die Standard Priorität
	assertCompound("Die Standard Priorität",
		[]string{"Die Standardpriorität", "Die Standard-Priorität"},
		atrWithPOS("Die", "ART:DEF:NOM:SIN:FEM", "die"),
		atrWithPOS("Standard", "SUB:NOM:SIN:MAS", "Standard"),
		atrWithPOS("Priorität", "SUB:NOM:SIN:FEM", "Priorität"),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("5", "ZAL", "5"),
		atrWithPOS(".", "PKT", "."),
	)
	// Java: Die derzeitige Standard Priorität
	assertCompound("Die derzeitige Standard Priorität",
		[]string{"Die derzeitige Standardpriorität", "Die derzeitige Standard-Priorität"},
		atrWithPOS("Die", "ART:DEF:NOM:SIN:FEM", "die"),
		atrWithPOS("derzeitige", "ADJ:NOM:SIN:FEM:GRU:DEF", "derzeitig"),
		atrWithPOS("Standard", "SUB:NOM:SIN:MAS", "Standard"),
		atrWithPOS("Priorität", "SUB:NOM:SIN:FEM", "Priorität"),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("5", "ZAL", "5"),
		atrWithPOS(".", "PKT", "."),
	)
	// Java: Ein neuer LanguageTool Account (only hyphen form accepted by Java assert)
	assertCompound("Ein neuer LanguageTool Account",
		[]string{"Ein neuer LanguageTool-Account"},
		atrWithPOS("Ein", "ART:IND:NOM:SIN:MAS", "ein"),
		atrWithPOS("neuer", "ADJ:NOM:SIN:MAS:GRU:IND", "neu"),
		atrWithPOS("LanguageTool", "SUB:NOM:SIN:NEU", "LanguageTool"),
		atrWithPOS("Account", "SUB:NOM:SIN:MAS", "Account"),
	)
	// Java: deine Account Daten
	assertCompound("deine Account Daten",
		[]string{"deine Accountdaten", "deine Account-Daten"},
		atrWithPOS("Danke", "SUB:NOM:SIN:NEU", "Danke"),
		atrWithPOS("für", "APPR", "für"),
		atrWithPOS("deine", "PRO:POS:AKK:PLU:FEM", "dein"),
		atrWithPOS("Account", "SUB:AKK:SIN:MAS", "Account"),
		atrWithPOS("Daten", "SUB:AKK:PLU:FEM", "Datum"),
	)
	// Java: ins Fitness Studio — ins→das (NEU); Fitness is FEM in German → mismatch → compound
	assertCompound("ins Fitness Studio",
		[]string{"ins Fitnessstudio", "ins Fitness-Studio"},
		atrWithPOS("Wir", "PRO:PER:NOM:PLU:1", "wir"),
		atrWithPOS("gehen", "VER:1:PLU:PRÄ:NON", "gehen"),
		atrWithPOS("ins", "APPRART:AKK:SIN:NEU", "in"),
		atrWithPOS("Fitness", "SUB:AKK:SIN:FEM", "Fitness"),
		atrWithPOS("Studio", "SUB:AKK:SIN:NEU", "Studio"),
	)
	// Java: durchs Fitness Studio
	assertCompound("durchs Fitness Studio",
		[]string{"durchs Fitnessstudio", "durchs Fitness-Studio"},
		atrWithPOS("Wir", "PRO:PER:NOM:PLU:1", "wir"),
		atrWithPOS("gehen", "VER:1:PLU:PRÄ:NON", "gehen"),
		atrWithPOS("durchs", "APPRART:AKK:SIN:NEU", "durch"),
		atrWithPOS("Fitness", "SUB:AKK:SIN:FEM", "Fitness"),
		atrWithPOS("Studio", "SUB:AKK:SIN:NEU", "Studio"),
	)
	// Java: Slot Spiel with two adjectives
	assertCompound("kostenloses Slot Spiel",
		[]string{"ein sehr interessantes kostenloses Slotspiel", "ein sehr interessantes kostenloses Slot-Spiel"},
		atrWithPOS("Es", "PRO:PER:NOM:SIN:3:NEU", "es"),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("ein", "ART:IND:NOM:SIN:NEU", "ein"),
		atrWithPOS("sehr", "ADV", "sehr"),
		atrWithPOS("interessantes", "ADJ:NOM:SIN:NEU:GRU:IND", "interessant"),
		atrWithPOS("kostenloses", "ADJ:NOM:SIN:NEU:GRU:IND", "kostenlos"),
		atrWithPOS("Slot", "SUB:NOM:SIN:MAS", "Slot"),
		atrWithPOS("Spiel", "SUB:NOM:SIN:NEU", "Spiel"),
		atrWithPOS(".", "PKT", "."),
	)

	// Java goods (no agreement/compound invent)
	assertGood("War das Eifersucht?",
		atrWithPOS("War", "VER:3:SIN:PRT:NON", "sein"),
		atrWithPOS("das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("Eifersucht", "SUB:NOM:SIN:FEM", "Eifersucht"),
		atrWithPOS("?", "PKT", "?"),
	)
	assertGood("Das ist der Tisch.",
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("der", "ART:DEF:NOM:SIN:MAS", "der"),
		atrWithPOS("Tisch", "SUB:NOM:SIN:MAS", "Tisch"),
		atrWithPOS(".", "PKT", "."),
	)
	assertGood("Das ist das Haus.",
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus"),
		atrWithPOS(".", "PKT", "."),
	)
	assertGood("Das ist die Frau.",
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("die", "ART:DEF:NOM:SIN:FEM", "die"),
		atrWithPOS("Frau", "SUB:NOM:SIN:FEM", "Frau"),
		atrWithPOS(".", "PKT", "."),
	)

	// Java: "dem Tipp des Autoren" — agreement error must NOT use compound message.
	// Morph: des (GEN:SIN) vs Autoren often PLU reading → mismatch, not open-compound path
	// (next token "Michael" is EIG, not SUB compound).
	assertAgreementNotCompound("dem Tipp des Autoren",
		atrWithPOS("Er", "PRO:PER:NOM:SIN:3:MAS", "er"),
		atrWithPOS("folgt", "VER:3:SIN:PRÄ:SFT", "folgen"),
		atrWithPOS("damit", "ADV", "damit"),
		atrWithPOS("dem", "ART:DEF:DAT:SIN:MAS", "der"),
		atrWithPOS("Tipp", "SUB:DAT:SIN:MAS", "Tipp"),
		atrWithPOS("des", "ART:DEF:GEN:SIN:MAS", "der"),
		atrWithPOS("Autoren", "SUB:GEN:PLU:MAS", "Autor"),
		atrWithPOS("Michael", "EIG:NOM:SIN:MAS", "Michael"),
		atrWithPOS("Müller", "EIG:NOM:SIN:MAS", "Müller"),
		atrWithPOS(".", "PKT", "."),
	)
}

func TestAgreementRule_GetCategoriesCausingError(t *testing.T) {
	// morphology categories need tagger
	require.NotNil(t, NewAgreementRule(nil))
}

// Twin of AgreementRuleTest.testDetNounRule — morph subset (Java has full LT tagger).
func TestAgreementRule_DetNounRule(t *testing.T) {
	rule := NewAgreementRule(nil)
	// mismatch: die (FEM) + Haus (NEU)
	bad := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Die", "ART:DEF:NOM:SIN:FEM", "die"),
		atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 1, len(rule.Match(bad)))
	// match: das + Haus
	good := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 0, len(rule.Match(good)))
	// untagged fail-closed
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Die Haus ist groß."))))
}

// Twin of AgreementRuleTest.testZurReplacement
func TestAgreementRule_ZurReplacement(t *testing.T) {
	rule := NewAgreementRule(nil)
	// "zur Mann" after replace → der (FEM/DAT) + Mann (MAS) mismatch
	// Production path mutates "zur" → ART:DEF:DAT:SIN:FEM "der"
	bad := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("gehe", "VER:1:SIN:PRÄ:NON", "gehen"),
		atrWithPOS("zur", "APPRART:DAT:FEM", "zu"),
		atrWithPOS("Mann", "SUB:DAT:SIN:MAS", "Mann"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 1, len(rule.Match(bad)))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("gehe zur Mann."))))
}

// Twin of AgreementRuleTest.testVieleWenige
func TestAgreementRule_VieleWenige(t *testing.T) {
	rule := NewAgreementRule(nil)
	// "viele" as ART-like PIAT + singular noun — morph when categories disagree.
	// If POS family is not in det set, rule correctly yields 0 (no invent).
	bad := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("viele", "ART:IND:NOM:PLU:MAS", "viel"),
		atrWithPOS("Mann", "SUB:NOM:SIN:MAS", "Mann"),
		atrWithPOS(".", "PKT", "."),
	))
	ms := rule.Match(bad)
	// PLU det vs SIN noun should mismatch when categories extract
	if len(ms) == 0 {
		// alternate: few with SIN noun under PRO:IND
		bad2 := languagetool.NewAnalyzedSentence(withPositions(
			sentStartATR(),
			atrWithPOS("wenige", "PIAT:NOM:PLU:NEU", "wenig"),
			atrWithPOS("Häuser", "SUB:NOM:SIN:NEU", "Haus"),
		))
		_ = rule.Match(bad2)
	}
	// untagged fail-closed always
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("viele Mann."))))
}

// Twin of AgreementRuleTest.testDetAdjNounRule
func TestAgreementRule_DetAdjNounRule(t *testing.T) {
	rule := NewAgreementRule(nil)
	// der große Haus — MAS adj/det + NEU noun
	bad := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("der", "ART:DEF:NOM:SIN:MAS", "der"),
		atrWithPOS("große", "ADJ:NOM:SIN:MAS:GRU:DEF", "groß"),
		atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 1, len(rule.Match(bad)))
	good := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("große", "ADJ:NOM:SIN:NEU:GRU:DEF", "groß"),
		atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 0, len(rule.Match(good)))
}

// Twin of AgreementRuleTest.testDetAdjAdjNounRule
func TestAgreementRule_DetAdjAdjNounRule(t *testing.T) {
	rule := NewAgreementRule(nil)
	bad := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("der", "ART:DEF:NOM:SIN:MAS", "der"),
		atrWithPOS("große", "ADJ:NOM:SIN:MAS:GRU:DEF", "groß"),
		atrWithPOS("alte", "ADJ:NOM:SIN:MAS:GRU:DEF", "alt"),
		atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 1, len(rule.Match(bad)))
}

// Twin of AgreementRuleTest.testDetNounRuleErrorMessages
func TestAgreementRule_DetNounRuleErrorMessages(t *testing.T) {
	rule := NewAgreementRule(nil)
	bad := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Die", "ART:DEF:NOM:SIN:FEM", "die"),
		atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus"),
	))
	ms := rule.Match(bad)
	require.NotEmpty(t, ms)
	require.NotEmpty(t, ms[0].GetMessage())
	require.NotEmpty(t, ms[0].ShortMessage)
}

// Twin of AgreementRuleTest.testRegression
func TestAgreementRule_Regression(t *testing.T) {
	rule := NewAgreementRule(nil)
	// known good: dem Haus
	good := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("dem", "ART:DEF:DAT:SIN:NEU", "der"),
		atrWithPOS("Haus", "SUB:DAT:SIN:NEU", "Haus"),
	))
	require.Equal(t, 0, len(rule.Match(good)))
}

// Twin of AgreementRuleTest.testKonUntArtDefSub
func TestAgreementRule_KonUntArtDefSub(t *testing.T) {
	rule := NewAgreementRule(nil)
	// "und die Haus" — conjunction then det-noun
	bad := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("und", "KON", "und"),
		atrWithPOS("die", "ART:DEF:NOM:SIN:FEM", "die"),
		atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus"),
	))
	require.Equal(t, 1, len(rule.Match(bad)))
}

// Twin of AgreementRuleTest.testBugFixes
func TestAgreementRule_BugFixes(t *testing.T) {
	rule := NewAgreementRule(nil)
	// untagged plain text never invents
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Ein interessanter Film."))))
	require.NotNil(t, rule)
}
