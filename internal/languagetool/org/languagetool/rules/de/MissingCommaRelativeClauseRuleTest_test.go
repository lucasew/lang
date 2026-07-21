package de

// Twin of MissingCommaRelativeClauseRuleTest — morph/POS inject (no surface invent).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestMissingCommaRelativeClauseRule_Match(t *testing.T) {
	// Untagged AnalyzePlain must not invent relative-comma hits
	rule := NewMissingCommaRelativeClauseRule(nil)
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Das Auto das am Straßenrand steht parkt im Halteverbot."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Computer machen die Leute dumm."))))

	behind := NewMissingCommaRelativeClauseRuleBehind(nil)
	require.Equal(t, 0, len(behind.Match(languagetool.AnalyzePlain("Das Auto, das am Straßenrand steht parkt im Halteverbot."))))
	require.Equal(t, 0, len(behind.Match(languagetool.AnalyzePlain("Das Auto, das am Straßenrand steht, parkt im Halteverbot."))))

	require.Equal(t, "COMMA_IN_FRONT_RELATIVE_CLAUSE", rule.GetID())
	require.Equal(t, "COMMA_BEHIND_RELATIVE_CLAUSE", behind.GetID())
}

// Twin of MissingCommaRelativeClauseRuleTest.testMatch front cases (morph).
func TestMissingCommaRelativeClauseRule_FrontMorphJava(t *testing.T) {
	rule := NewMissingCommaRelativeClauseRule(nil)
	// "Das Auto das am Straßenrand steht parkt im Halteverbot." → 4-12
	s := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("Auto", "SUB:NOM:SIN:NEU", "Auto"),
		atrWithPOS("das", "PRELS:NOM:SIN:NEU", "das"),
		atrWithPOS("am", "APPRART:DAT:SIN:MAS", "an"),
		atrWithPOS("Straßenrand", "SUB:DAT:SIN:MAS", "Straßenrand"),
		atrWithPOS("steht", "VER:3:SIN:PRS:SFT", "stehen"),
		atrWithPOS("parkt", "VER:3:SIN:PRS:SFT", "parken"),
		atrWithPOS("im", "APPRART:DAT:SIN:NEU", "in"),
		atrWithPOS("Halteverbot", "SUB:DAT:SIN:NEU", "Halteverbot"),
		atrWithPOS(".", "PKT", "."),
	))
	ms := rule.Match(s)
	require.Equal(t, 1, len(ms))
	require.Equal(t, 4, ms[0].GetFromPos())
	require.Equal(t, 12, ms[0].GetToPos())

	// with comma already after Auto — still front match at relative pronoun (Java)
	s2 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("Auto", "SUB:NOM:SIN:NEU", "Auto"),
		atrWithPOS("das", "PRELS:NOM:SIN:NEU", "das"),
		atrWithPOS("am", "APPRART:DAT:SIN:MAS", "an"),
		atrWithPOS("Straßenrand", "SUB:DAT:SIN:MAS", "Straßenrand"),
		atrWithPOS("steht", "VER:3:SIN:PRS:SFT", "stehen"),
		atrWithPOS(",", "PKT", ","),
		atrWithPOS("parkt", "VER:3:SIN:PRS:SFT", "parken"),
		atrWithPOS("im", "APPRART:DAT:SIN:NEU", "in"),
		atrWithPOS("Halteverbot", "SUB:DAT:SIN:NEU", "Halteverbot"),
		atrWithPOS(".", "PKT", "."),
	))
	ms2 := rule.Match(s2)
	require.Equal(t, 1, len(ms2))
	require.Equal(t, 4, ms2[0].GetFromPos())
	require.Equal(t, 12, ms2[0].GetToPos())

	// no relative clause
	good := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Computer", "SUB:NOM:PLU:MAS", "Computer"),
		atrWithPOS("machen", "VER:3:PLU:PRS:SFT", "machen"),
		atrWithPOS("die", "ART:DEF:AKK:PLU:*", "die"),
		atrWithPOS("Leute", "SUB:AKK:PLU:MAS", "Leute"),
		atrWithPOS("dumm", "ADJ:PRD:GRU", "dumm"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Empty(t, rule.Match(good))

	// "Die Frau die vor dem Auto steht hat schwarze Haare." → 4-12
	frau := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Die", "ART:DEF:NOM:SIN:FEM", "die"),
		atrWithPOS("Frau", "SUB:NOM:SIN:FEM", "Frau"),
		atrWithPOS("die", "PRELS:NOM:SIN:FEM", "die"),
		atrWithPOS("vor", "PRP:DAT", "vor"),
		atrWithPOS("dem", "ART:DEF:DAT:SIN:NEU", "der"),
		atrWithPOS("Auto", "SUB:DAT:SIN:NEU", "Auto"),
		atrWithPOS("steht", "VER:3:SIN:PRS:SFT", "stehen"),
		atrWithPOS("hat", "VER:3:SIN:PRS:SFT", "haben"),
		atrWithPOS("schwarze", "ADJ:AKK:PLU:NEU:GRU:IND", "schwarz"),
		atrWithPOS("Haare", "SUB:AKK:PLU:NEU", "Haar"),
		atrWithPOS(".", "PKT", "."),
	))
	msF := rule.Match(frau)
	require.Equal(t, 1, len(msF))
	require.Equal(t, 4, msF[0].GetFromPos())
	require.Equal(t, 12, msF[0].GetToPos())

	// "Das Auto in dem der Mann sitzt, parkt im Halteverbot." → 4-15
	// dem refers to Auto → SIN:NEU (Java LT tags)
	indem := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("Auto", "SUB:NOM:SIN:NEU", "Auto"),
		atrWithPOS("in", "PRP:DAT", "in"),
		atrWithPOS("dem", "PRELS:DAT:SIN:NEU", "der"),
		atrWithPOS("der", "ART:DEF:NOM:SIN:MAS", "der"),
		atrWithPOS("Mann", "SUB:NOM:SIN:MAS", "Mann"),
		atrWithPOS("sitzt", "VER:3:SIN:PRS:SFT", "sitzen"),
		atrWithPOS(",", "PKT", ","),
		atrWithPOS("parkt", "VER:3:SIN:PRS:SFT", "parken"),
		atrWithPOS("im", "APPRART:DAT:SIN:NEU", "in"),
		atrWithPOS("Halteverbot", "SUB:DAT:SIN:NEU", "Halteverbot"),
		atrWithPOS(".", "PKT", "."),
	))
	msIn := rule.Match(indem)
	require.Equal(t, 1, len(msIn))
	require.Equal(t, 4, msIn[0].GetFromPos())
	require.Equal(t, 15, msIn[0].GetToPos())

	// "Alles was ich habe, ist ein Buch." → 0-9
	// was often ungendered; Alles as PRO:DEM matches matchesGender empty-gender path
	alles := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Alles", "PRO:DEM:NOM:SIN:NEU", "alles"),
		atrWithPOS("was", "PRELS", "was"),
		atrWithPOS("ich", "PPER:NOM:SIN:1", "ich"),
		atrWithPOS("habe", "VER:1:SIN:PRS:SFT", "haben"),
		atrWithPOS(",", "PKT", ","),
		atrWithPOS("ist", "VER:3:SIN:PRS:NON", "sein"),
		atrWithPOS("ein", "ART:IND:NOM:SIN:NEU", "ein"),
		atrWithPOS("Buch", "SUB:NOM:SIN:NEU", "Buch"),
		atrWithPOS(".", "PKT", "."),
	))
	msA := rule.Match(alles)
	require.Equal(t, 1, len(msA))
	require.Equal(t, 0, msA[0].GetFromPos())
	require.Equal(t, 9, msA[0].GetToPos())
}

// Twin of MissingCommaRelativeClauseRuleTest behind (second constructor).
func TestMissingCommaRelativeClauseRule_BehindMorphJava(t *testing.T) {
	rule := NewMissingCommaRelativeClauseRuleBehind(nil)
	// "Das Auto, das am Straßenrand steht parkt im Halteverbot." → 29-40
	s := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("Auto", "SUB:NOM:SIN:NEU", "Auto"),
		atrWithPOS(",", "PKT", ","),
		atrWithPOS("das", "PRELS:NOM:SIN:NEU", "das"),
		atrWithPOS("am", "APPRART:DAT:SIN:MAS", "an"),
		atrWithPOS("Straßenrand", "SUB:DAT:SIN:MAS", "Straßenrand"),
		atrWithPOS("steht", "VER:3:SIN:PRS:SFT", "stehen"),
		atrWithPOS("parkt", "VER:3:SIN:PRS:SFT", "parken"),
		atrWithPOS("im", "APPRART:DAT:SIN:NEU", "in"),
		atrWithPOS("Halteverbot", "SUB:DAT:SIN:NEU", "Halteverbot"),
		atrWithPOS(".", "PKT", "."),
	))
	ms := rule.Match(s)
	require.Equal(t, 1, len(ms))
	require.Equal(t, 29, ms[0].GetFromPos())
	require.Equal(t, 40, ms[0].GetToPos())

	// comma after relative clause → no match
	good := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("Auto", "SUB:NOM:SIN:NEU", "Auto"),
		atrWithPOS(",", "PKT", ","),
		atrWithPOS("das", "PRELS:NOM:SIN:NEU", "das"),
		atrWithPOS("am", "APPRART:DAT:SIN:MAS", "an"),
		atrWithPOS("Straßenrand", "SUB:DAT:SIN:MAS", "Straßenrand"),
		atrWithPOS("steht", "VER:3:SIN:PRS:SFT", "stehen"),
		atrWithPOS(",", "PKT", ","),
		atrWithPOS("parkt", "VER:3:SIN:PRS:SFT", "parken"),
		atrWithPOS("im", "APPRART:DAT:SIN:NEU", "in"),
		atrWithPOS("Halteverbot", "SUB:DAT:SIN:NEU", "Halteverbot"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Empty(t, rule.Match(good))

	// "Die Frau, die vor dem Auto steht hat schwarze Haare." → 27-36
	frauB := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Die", "ART:DEF:NOM:SIN:FEM", "die"),
		atrWithPOS("Frau", "SUB:NOM:SIN:FEM", "Frau"),
		atrWithPOS(",", "PKT", ","),
		atrWithPOS("die", "PRELS:NOM:SIN:FEM", "die"),
		atrWithPOS("vor", "PRP:DAT", "vor"),
		atrWithPOS("dem", "ART:DEF:DAT:SIN:NEU", "der"),
		atrWithPOS("Auto", "SUB:DAT:SIN:NEU", "Auto"),
		atrWithPOS("steht", "VER:3:SIN:PRS:SFT", "stehen"),
		atrWithPOS("hat", "VER:3:SIN:PRS:SFT", "haben"),
		atrWithPOS("schwarze", "ADJ:AKK:PLU:NEU:GRU:IND", "schwarz"),
		atrWithPOS("Haare", "SUB:AKK:PLU:NEU", "Haar"),
		atrWithPOS(".", "PKT", "."),
	))
	msFB := rule.Match(frauB)
	require.Equal(t, 1, len(msFB))
	require.Equal(t, 27, msFB[0].GetFromPos())
	require.Equal(t, 36, msFB[0].GetToPos())

	// "Alles, was ich habe ist ein Buch." → 15-23
	allesB := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Alles", "PRO:DEM:NOM:SIN:NEU", "alles"),
		atrWithPOS(",", "PKT", ","),
		atrWithPOS("was", "PRELS", "was"),
		atrWithPOS("ich", "PPER:NOM:SIN:1", "ich"),
		atrWithPOS("habe", "VER:1:SIN:PRS:SFT", "haben"),
		atrWithPOS("ist", "VER:3:SIN:PRS:NON", "sein"),
		atrWithPOS("ein", "ART:IND:NOM:SIN:NEU", "ein"),
		atrWithPOS("Buch", "SUB:NOM:SIN:NEU", "Buch"),
		atrWithPOS(".", "PKT", "."),
	))
	msAB := rule.Match(allesB)
	require.Equal(t, 1, len(msAB))
	require.Equal(t, 15, msAB[0].GetFromPos())
	require.Equal(t, 23, msAB[0].GetToPos())
}

func TestWithPositions_UTF16AndPunctSpacing(t *testing.T) {
	// "Auto," then "das" must match Java offsets with ß in Straßenrand
	toks := withPositions(
		sentStartATR(),
		atrWithPOS("Das", "ART", "das"),
		atrWithPOS("Auto", "SUB", "Auto"),
		atrWithPOS(",", "PKT", ","),
		atrWithPOS("das", "PRELS", "das"),
		atrWithPOS("am", "PRP", "an"),
		atrWithPOS("Straßenrand", "SUB", "Straßenrand"),
		atrWithPOS("steht", "VER", "stehen"),
	)
	// Auto 4-8, comma 8-9, das 10-13, am 14-16, Straßenrand 17-28, steht 29-34
	require.Equal(t, 4, toks[2].GetStartPos())  // Auto
	require.Equal(t, 8, toks[3].GetStartPos())  // comma
	require.Equal(t, 10, toks[4].GetStartPos()) // das
	require.Equal(t, 17, toks[6].GetStartPos()) // Straßenrand
	require.Equal(t, 29, toks[7].GetStartPos()) // steht
	require.Equal(t, 11, utf16LenDE("Straßenrand"))
}
