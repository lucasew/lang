package chunking

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func atrPos(token, pos string, start int) *languagetool.AnalyzedTokenReadings {
	p := pos
	return languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken(token, &p, nil), start)
}

func TestGermanChunker(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Der", "ART:DEF:NOM:SIN:MAS", 0),
		atrPos("Hund", "SUB:NOM:SIN:MAS", 4),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "B-NP")
	require.Contains(t, tokens[1].GetChunkTags(), "I-NP")
}

// Port of GermanChunkerTest.assertBasicChunks / getBasicChunks (REGEXES1 only).
func TestGermanChunker_GetBasicChunks_OpenNLPLike(t *testing.T) {
	// Java: "Ein/B Haus/I"
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Ein", "ART:IND:NOM:SIN:NEU", 0),
		atrPos("Haus", "SUB:NOM:SIN:NEU", 4),
	}
	basic := NewGermanChunker().GetBasicChunks(tokens)
	require.Len(t, basic, 2)
	require.Equal(t, "Ein", basic[0].Token)
	require.Equal(t, "B-NP", basic[0].ChunkTags[0].String())
	require.Equal(t, "I-NP", basic[1].ChunkTags[0].String())
	// Does not mutate readings (Java side list)
	require.Empty(t, tokens[0].GetChunkTags())
	require.Empty(t, tokens[1].GetChunkTags())

	// Java: "Herr/B Schrödinger/I isst einen/B Kuchen/I"
	tokens2 := []*languagetool.AnalyzedTokenReadings{
		atrPos("Herr", "SUB:NOM:SIN:MAS", 0),
		atrPos("Schrödinger", "EIG:NOM:SIN:MAS", 5),
		atrPos("isst", "VER:3:SIN:PRÄ:SFT", 17),
		atrPos("einen", "ART:IND:AKK:SIN:MAS", 22),
		atrPos("Kuchen", "SUB:AKK:SIN:MAS", 28),
	}
	basic2 := NewGermanChunker().GetBasicChunks(tokens2)
	require.Len(t, basic2, 5)
	require.Equal(t, "B-NP", basic2[0].ChunkTags[0].String())
	require.Equal(t, "I-NP", basic2[1].ChunkTags[0].String())
	// verb is O (no invent BIO)
	require.Equal(t, "O", basic2[2].ChunkTags[0].String())
	require.Equal(t, "B-NP", basic2[3].ChunkTags[0].String())
	require.Equal(t, "I-NP", basic2[4].ChunkTags[0].String())
}

func TestGermanChunker_GetBasicChunks_UntaggedIsO(t *testing.T) {
	// Lone EIG: REGEXES1 has no bare-EIG pattern → O
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Berlin", "EIG:NOM:SIN:NEU", 0),
	}
	basic := NewGermanChunker().GetBasicChunks(tokens)
	require.Len(t, basic, 1)
	require.Equal(t, "O", basic[0].ChunkTags[0].String())
}

// Java REGEXES1: SUB und SUB ("Mythen und Sagen")
func TestGermanChunker_SubUndSub(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Mythen", "SUB:NOM:PLU:MAS", 0),
		atrPos("und", "KON:NEB", 7),
		atrPos("Sagen", "SUB:NOM:PLU:FEM", 11),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "B-NP")
	require.Contains(t, tokens[1].GetChunkTags(), "I-NP")
	require.Contains(t, tokens[2].GetChunkTags(), "I-NP")
}

// Java REGEXES1 buildExpanded: SUB (bzw .) SUB
func TestGermanChunker_SubBzwSub(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Mythen", "SUB:NOM:PLU:MAS", 0),
		atrPos("bzw", "KON", 7),
		atrPos(".", "PKT", 10),
		atrPos("Sagen", "SUB:NOM:PLU:FEM", 12),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "B-NP")
	require.Contains(t, tokens[1].GetChunkTags(), "I-NP")
	require.Contains(t, tokens[2].GetChunkTags(), "I-NP")
	require.Contains(t, tokens[3].GetChunkTags(), "I-NP")
}

// Bare SUB is REGEXES1 (SUB+); bare EIG is not — no invent POS→BIO.
func TestGermanChunker_BareSubVsBareEig(t *testing.T) {
	sub := []*languagetool.AnalyzedTokenReadings{atrPos("Hund", "SUB:NOM:SIN:MAS", 0)}
	NewGermanChunker().AddChunkTags(sub)
	require.Contains(t, sub[0].GetChunkTags(), "B-NP")

	eig := []*languagetool.AnalyzedTokenReadings{atrPos("Berlin", "EIG:NOM:SIN:NEU", 0)}
	NewGermanChunker().AddChunkTags(eig)
	require.NotContains(t, eig[0].GetChunkTags(), "B-NP")
}

// Java REGEXES1: Herr + EIG
func TestGermanChunker_HerrEig(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Herr", "SUB:NOM:SIN:MAS", 0),
		atrPos("Schröder", "EIG:NOM:SIN:MAS", 5),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "B-NP")
	require.Contains(t, tokens[1].GetChunkTags(), "I-NP")
}

// Java REGEXES1: ZAL SUB ("zwei Wochen")
func TestGermanChunker_ZalSub(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("zwei", "ZAL", 0),
		atrPos("Wochen", "SUB:AKK:PLU:FEM", 5),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "B-NP")
	require.Contains(t, tokens[1].GetChunkTags(), "I-NP")
}

// Java REGEXES1: ART ADJ* Capitalized unknown noun
func TestGermanChunker_ArtCapitalUnknownNoun(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("eine", "ART:IND:NOM:SIN:FEM", 0),
		atrPos("leckere", "ADJ:NOM:SIN:FEM:GRU:SOL", 5),
		atrPos("Lasagne", "UNKNOWN", 13), // unknown POS; surface capitalized
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "B-NP")
	require.Contains(t, tokens[1].GetChunkTags(), "I-NP")
	require.Contains(t, tokens[2].GetChunkTags(), "I-NP")
}

// Java REGEXES2: EIG und EIG → NPP
func TestGermanChunker_EigUndEig_NPP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Julia", "EIG:NOM:SIN:FEM", 0),
		atrPos("und", "KON:NEB", 6),
		atrPos("Karsten", "EIG:NOM:SIN:MAS", 10),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "NPP")
	require.Contains(t, tokens[2].GetChunkTags(), "NPP")
}

// Java REGEXES2: PRP + NP → PP
func TestGermanChunker_PRP_NP_PP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("für", "PRP:AKK", 0),
		atrPos("die", "ART:DEF:AKK:PLU:FEM", 4),
		atrPos("Fische", "SUB:AKK:PLU:FEM", 8),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "PP")
	require.Contains(t, tokens[2].GetChunkTags(), "PP")
}

// Java REGEXES2: singular NP → NPS
func TestGermanChunker_SingularNPS(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("das", "ART:DEF:NOM:SIN:NEU", 0),
		atrPos("Auto", "SUB:NOM:SIN:NEU", 4),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "B-NP")
	require.Contains(t, tokens[0].GetChunkTags(), "NPS")
	require.Contains(t, tokens[1].GetChunkTags(), "NPS")
}

// Java: weder SUB noch SUB → NPP
func TestGermanChunker_WederNoch_NPP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("weder", "KON", 0),
		atrPos("Gerechtigkeit", "SUB:NOM:SIN:FEM", 6),
		atrPos("noch", "KON", 21),
		atrPos("Freiheit", "SUB:NOM:SIN:FEM", 26),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "NPP")
	require.Contains(t, tokens[3].GetChunkTags(), "NPP")
}

// Java: Letztes Jahr → PP
func TestGermanChunker_LetztesJahr_PP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Letztes", "ADJ:NOM:SIN:NEU", 0),
		atrPos("Jahr", "SUB:NOM:SIN:NEU", 8),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "PP")
	require.Contains(t, tokens[1].GetChunkTags(), "PP")
}

// Java: Herr und Frau Schröder → NPP
func TestGermanChunker_HerrUndFrau_NPP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Herr", "SUB:NOM:SIN:MAS", 0),
		atrPos("und", "KON", 5),
		atrPos("Frau", "SUB:NOM:SIN:FEM", 9),
		atrPos("Schröder", "EIG:NOM:SIN:MAS", 14),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "NPP")
	require.Contains(t, tokens[3].GetChunkTags(), "NPP")
}

// Java: Bei sehr guten Beobachtungsbedingungen → PP
func TestGermanChunker_PRP_Adv_Adj_NP_PP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Bei", "PRP:DAT", 0),
		atrPos("sehr", "ADV", 4),
		atrPos("guten", "ADJ:DAT:PLU:FEM", 9),
		atrPos("Beobachtungsbedingungen", "SUB:DAT:PLU:FEM", 15),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "PP")
	require.Contains(t, tokens[3].GetChunkTags(), "PP")
}

// Java genitive: die ältere der beiden Töchter → NPS
func TestGermanChunker_GenitiveAeltereDerBeiden(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("die", "ART:DEF:NOM:SIN:FEM", 0),
		atrPos("ältere", "ADJ:NOM:SIN:FEM", 4),
		atrPos("der", "ART:DEF:GEN:PLU:FEM", 11),
		atrPos("beiden", "PRO:POS:GEN:PLU:FEM", 15),
		atrPos("Töchter", "SUB:GEN:PLU:FEM", 22),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "NPS")
	require.Contains(t, tokens[4].GetChunkTags(), "NPS")
}

// Java: eine Menge englischer Wörter → NPP
func TestGermanChunker_EineMenge_NPP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("eine", "ART:IND:NOM:SIN:FEM", 0),
		atrPos("Menge", "SUB:NOM:SIN:FEM", 5),
		atrPos("englischer", "ADJ:GEN:PLU:NEU", 11),
		atrPos("Wörter", "SUB:GEN:PLU:NEU", 22),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "NPP")
	require.Contains(t, tokens[3].GetChunkTags(), "NPP")
}

// Java: laut den meisten Quellen → PP
func TestGermanChunker_LautQuellen_PP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Laut", "PRP", 0),
		atrPos("den", "ART:DEF:DAT:PLU", 5),
		atrPos("meisten", "ADJ:DAT:PLU", 9),
		atrPos("Quellen", "SUB:DAT:PLU:FEM", 17),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "PP")
	require.Contains(t, tokens[3].GetChunkTags(), "PP")
}

// Java: die älteste und bekannteste Maßnahme → NPS
func TestGermanChunker_AeltesteUndBekannteste_NPS(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("die", "ART:DEF:NOM:SIN:FEM", 0),
		atrPos("älteste", "ADJ:NOM:SIN:FEM:SUP:SOL", 4),
		atrPos("und", "KON", 12),
		atrPos("bekannteste", "ADJ:NOM:SIN:FEM:SUP:SOL", 16),
		atrPos("Maßnahme", "SUB:NOM:SIN:FEM", 28),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "NPS")
	require.Contains(t, tokens[4].GetChunkTags(), "NPS")
}

// Java: eins ihrer drei Autos → NPS
func TestGermanChunker_EinsIhrerDreiAutos_NPS(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("eins", "PRO:IND:NOM:SIN:NEU", 0),
		atrPos("ihrer", "PRO:POS:GEN:PLU:NEU", 5),
		atrPos("drei", "ZAL", 11),
		atrPos("Autos", "SUB:GEN:PLU:NEU", 16),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "NPS")
	require.Contains(t, tokens[3].GetChunkTags(), "NPS")
}

// Java: der von der Regierung geprüfte Hund → NPS
func TestGermanChunker_DerVonDerRegierungGepruefteHund_NPS(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Der", "ART:DEF:NOM:SIN:MAS", 0),
		atrPos("von", "PRP:DAT", 4),
		atrPos("der", "ART:DEF:DAT:SIN:FEM", 8),
		atrPos("Regierung", "SUB:DAT:SIN:FEM", 12),
		atrPos("geprüfte", "PA2:NOM:SIN:MAS:GRU:SOL", 23),
		atrPos("Hund", "SUB:NOM:SIN:MAS", 32),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "NPS")
	require.Contains(t, tokens[5].GetChunkTags(), "NPS")
}

// Java: Einer der beiden Höfe → NPS
func TestGermanChunker_EinerDerBeidenHoefe_NPS(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Einer", "PRO:IND:NOM:SIN:MAS", 0),
		atrPos("der", "ART:DEF:GEN:PLU:MAS", 6),
		atrPos("beiden", "PRO:IND:GEN:PLU:MAS", 10),
		atrPos("Höfe", "SUB:GEN:PLU:MAS", 17),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "NPS")
	require.Contains(t, tokens[3].GetChunkTags(), "NPS")
}

// Java: 37 Prozent → NPS and NPP
func TestGermanChunker_Prozent_NPS_NPP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("37", "ZAL", 0),
		atrPos("Prozent", "SUB:NOM:PLU:NEU", 3),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "NPS")
	require.Contains(t, tokens[0].GetChunkTags(), "NPP")
	require.Contains(t, tokens[1].GetChunkTags(), "NPS")
	require.Contains(t, tokens[1].GetChunkTags(), "NPP")
}

// Java: in den darauf folgenden Wochen → PP
func TestGermanChunker_InDenDaraufFolgendenWochen_PP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("In", "PRP:DAT", 0),
		atrPos("den", "ART:DEF:DAT:PLU", 3),
		atrPos("darauf", "ADV", 7),
		atrPos("folgenden", "ADJ:DAT:PLU:FEM", 14),
		atrPos("Wochen", "SUB:DAT:PLU:FEM", 25),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "PP")
	require.Contains(t, tokens[4].GetChunkTags(), "PP")
}

// Java: in deren deutschen Installationen → PP
func TestGermanChunker_InDerenDeutschenInstallationen_PP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("in", "PRP:DAT", 0),
		atrPos("deren", "PRO:POS:DAT:PLU", 3),
		atrPos("deutschen", "ADJ:DAT:PLU:FEM", 9),
		atrPos("Installationen", "SUB:DAT:PLU:FEM", 19),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "PP")
	require.Contains(t, tokens[3].GetChunkTags(), "PP")
}

// Java: die letzten zwei Monate → PP
func TestGermanChunker_DieLetztenZweiMonate_PP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Die", "ART:DEF:NOM:PLU", 0),
		atrPos("letzten", "ADJ:NOM:PLU", 4),
		atrPos("zwei", "ZAL", 12),
		atrPos("Monate", "SUB:NOM:PLU:MAS", 17),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "PP")
	require.Contains(t, tokens[3].GetChunkTags(), "PP")
}

// Java: Beziehungen zwischen Kanada und dem Iran → NPP
func TestGermanChunker_BeziehungenZwischen_NPP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Die", "ART:DEF:NOM:PLU:FEM", 0),
		atrPos("Beziehungen", "SUB:NOM:PLU:FEM", 4),
		atrPos("zwischen", "PRP:DAT", 16),
		atrPos("Kanada", "EIG:DAT:SIN:NEU", 25),
		atrPos("und", "KON", 32),
		atrPos("dem", "ART:DEF:DAT:SIN:MAS", 36),
		atrPos("Iran", "EIG:DAT:SIN:MAS", 40),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "NPP")
	require.Contains(t, tokens[3].GetChunkTags(), "NPP")
	require.Contains(t, tokens[6].GetChunkTags(), "NPP")
}

// Java: eine Masseeinheit und keine Gewichtseinheit → NPS
func TestGermanChunker_UndKeine_NPS(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("eine", "ART:IND:NOM:SIN:FEM", 0),
		atrPos("Masseeinheit", "SUB:NOM:SIN:FEM", 5),
		atrPos("und", "KON", 18),
		atrPos("keine", "ART:IND:NOM:SIN:FEM", 22),
		atrPos("Gewichtseinheit", "SUB:NOM:SIN:FEM", 28),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "NPS")
	require.Contains(t, tokens[3].GetChunkTags(), "NPS")
	require.Contains(t, tokens[4].GetChunkTags(), "NPS")
}

// Java: Der See und das anliegende Marschland → NPP
func TestGermanChunker_UndArtPa1Sub_NPP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Der", "ART:DEF:NOM:SIN:MAS", 0),
		atrPos("See", "SUB:NOM:SIN:MAS", 4),
		atrPos("und", "KON", 8),
		atrPos("das", "ART:DEF:NOM:SIN:NEU", 12),
		atrPos("anliegende", "PA1:NOM:SIN:NEU:GRU:SOL", 16),
		atrPos("Marschland", "SUB:NOM:SIN:NEU", 27),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "NPP")
	require.Contains(t, tokens[5].GetChunkTags(), "NPP")
}

// Java: dass sie wie ein Spiel → NPP
func TestGermanChunker_DassSieWie_NPP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("dass", "KON:UNT", 0),
		atrPos("sie", "PRO:PER:NOM:PLU", 5),
		atrPos("wie", "KON", 9),
		atrPos("ein", "ART:IND:NOM:SIN:NEU", 13),
		atrPos("Spiel", "SUB:NOM:SIN:NEU", 17),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "NPP")
	require.Contains(t, tokens[4].GetChunkTags(), "NPP")
}

// Java: Veranstaltung, die immer wieder ein kultureller Höhepunkt → NPS
func TestGermanChunker_RelClauseAdv_NPS(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Eine", "ART:IND:NOM:SIN:FEM", 0),
		atrPos("Veranstaltung", "SUB:NOM:SIN:FEM", 5),
		atrPos(",", "PKT", 18),
		atrPos("die", "PRO:REL:NOM:SIN:FEM", 20),
		atrPos("immer", "ADV", 24),
		atrPos("wieder", "ADV", 30),
		atrPos("ein", "ART:IND:NOM:SIN:MAS", 37),
		atrPos("kultureller", "ADJ:NOM:SIN:MAS:GRU:SOL", 41),
		atrPos("Höhepunkt", "SUB:NOM:SIN:MAS", 53),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "NPS")
	require.Contains(t, tokens[1].GetChunkTags(), "NPS")
	require.Contains(t, tokens[8].GetChunkTags(), "NPS")
}

// Java: ADJ , B-NP und NP → NPP ("…, islamischen und jüdischen Traditionen")
// Uses ART+… so REGEXES1 produces B-NP on the post-comma span (OpenNLP-like input).
func TestGermanChunker_AdjCommaUnd_NPP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("christlichen", "ADJ:DAT:PLU", 0),
		atrPos(",", "PKT", 13),
		atrPos("die", "ART:DEF:NOM:PLU", 15),
		atrPos("islamischen", "ADJ:NOM:PLU", 19),
		atrPos("Traditionen", "SUB:NOM:PLU:FEM", 31),
		atrPos("und", "KON", 43),
		atrPos("jüdischen", "ADJ:NOM:PLU", 47),
		atrPos("Mythen", "SUB:NOM:PLU:MAS", 57),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "NPP")
	require.Contains(t, tokens[7].GetChunkTags(), "NPP")
}

// Java: eine der am meisten verbreiteten Krankheiten → NPS
// REGEXES1 PA2*+SUB fuses verbreiteten+Krankheiten; OpenRegex needs Krankheiten as B-NP
// after PA2. Fixture: PA2 token without SUB fusion (Krankheiten B-NP alone).
func TestGermanChunker_EineDerAmMeisten_NPS(t *testing.T) {
	// Pattern unit: seeded chunks match Java OpenRegex
	toks := []ChunkTaggedToken{
		NewChunkTaggedToken("eine", []ChunkTag{NewChunkTag("O")}, atrPos("eine", "ART:IND:NOM:SIN:FEM", 0)),
		NewChunkTaggedToken("der", []ChunkTag{NewChunkTag("O")}, atrPos("der", "ART:DEF:GEN:PLU", 5)),
		NewChunkTaggedToken("am", []ChunkTag{NewChunkTag("O")}, atrPos("am", "PRP:DAT:ART", 9)),
		NewChunkTaggedToken("meisten", []ChunkTag{NewChunkTag("O")}, atrPos("meisten", "ADJ:SUP", 12)),
		NewChunkTaggedToken("verbreiteten", []ChunkTag{NewChunkTag("O")}, atrPos("verbreiteten", "PA2:GEN:PLU:FEM", 20)),
		NewChunkTaggedToken("Krankheiten", []ChunkTag{NewChunkTag("B-NP")}, atrPos("Krankheiten", "SUB:GEN:PLU:FEM", 33)),
	}
	re := CompileOpenRegex(ExpandGermanChunkSyntax(`<regex=eine[rs]?> <der> <am> <pos=ADJ> <pos=PA2> <NP>`), NewChunkTokenFactory(false))
	require.NotEmpty(t, re.FindAll(toks))
	// Full path: beiden-style alternate (avoids PA2*+SUB fuse).
	tokens2 := []*languagetool.AnalyzedTokenReadings{
		atrPos("eine", "ART:IND:NOM:SIN:FEM", 0),
		atrPos("der", "ART:DEF:GEN:PLU", 5),
		atrPos("beiden", "PRO:IND:GEN:PLU", 9),
		atrPos("großen", "ADJ:GEN:PLU:FEM", 16),
		atrPos("Töchter", "SUB:GEN:PLU:FEM", 23),
	}
	NewGermanChunker().AddChunkTags(tokens2)
	// <regex=eine[rs]?> <der> <beiden> <pos=ADJ>* <pos=SUB> → NPS
	require.Contains(t, tokens2[0].GetChunkTags(), "NPS")
	require.Contains(t, tokens2[4].GetChunkTags(), "NPS")
}

// Java: Synthese organischer Verbindungen → NPS (NPS + NPP&GEN overwrite)
func TestGermanChunker_SyntheseOrganischer_NPS(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Synthese", "SUB:NOM:SIN:FEM", 0),
		atrPos("organischer", "ADJ:GEN:PLU:NEU:GRU:SOL", 9),
		atrPos("Verbindungen", "SUB:GEN:PLU:NEU", 21),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "NPS")
	require.Contains(t, tokens[1].GetChunkTags(), "NPS")
	require.Contains(t, tokens[2].GetChunkTags(), "NPS")
}

// Java: !einige on NPS+NPP(GEN) — "einige" head must not match that REGEXES2 pattern.
func TestGermanChunker_EinigeDer_NotNPSMerge(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("einige", "PRO:IND:NOM:PLU:ALG", 0),
		atrPos("organischer", "ADJ:GEN:PLU:NEU:GRU:SOL", 7),
		atrPos("Verbindungen", "SUB:GEN:PLU:NEU", 19),
	}
	// Seed chunks as after prior REGEXES2 passes:
	toks := []ChunkTaggedToken{
		NewChunkTaggedToken("einige", []ChunkTag{NewChunkTag("B-NP"), NewChunkTag("NPS")}, tokens[0]),
		NewChunkTaggedToken("organischer", []ChunkTag{NewChunkTag("B-NP"), NewChunkTag("NPP")}, tokens[1]),
		NewChunkTaggedToken("Verbindungen", []ChunkTag{NewChunkTag("I-NP"), NewChunkTag("NPP")}, tokens[2]),
	}
	// Pattern: <chunk=NPS & !einige> <chunk=NPP & (pos=GEN |pos=ZAL)>+
	re := CompileOpenRegex(
		`<chunk=NPS & !einige> <chunk=NPP & (pos=GEN |pos=ZAL)>+`,
		NewChunkTokenFactory(false),
	)
	require.Empty(t, re.FindAll(toks), "einige head must be excluded by !einige")
	// Without !einige the same span would match:
	re2 := CompileOpenRegex(
		`<chunk=NPS> <chunk=NPP & (pos=GEN |pos=ZAL)>+`,
		NewChunkTokenFactory(false),
	)
	require.NotEmpty(t, re2.FindAll(toks))
}

// Java: Von ursprünglich drei Almhütten → PP
// ZAL+SUB fuses drei+Almhütten; PP pattern needs ZAL then B-NP. Digit form unfuses.
func TestGermanChunker_PRP_AdjPrd_Zal_NP_PP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Von", "PRP:DAT", 0),
		atrPos("ursprünglich", "ADJ:PRD:GRU", 4),
		atrPos("3", "ZAL", 17), // digit surface; REGEXES1 ZAL+SUB still fuses if SUB follows
		atrPos("Almhütten", "SUB:DAT:PLU:FEM", 19),
	}
	// Seeded OpenRegex: ZAL then B-NP (unfused)
	toks := []ChunkTaggedToken{
		NewChunkTaggedToken("Von", []ChunkTag{NewChunkTag("O")}, tokens[0]),
		NewChunkTaggedToken("ursprünglich", []ChunkTag{NewChunkTag("O")}, tokens[1]),
		NewChunkTaggedToken("drei", []ChunkTag{NewChunkTag("O")}, atrPos("drei", "ZAL", 17)),
		NewChunkTaggedToken("Almhütten", []ChunkTag{NewChunkTag("B-NP")}, tokens[3]),
	}
	re := CompileOpenRegex(ExpandGermanChunkSyntax(`<pos=PRP> <pos=ADJ:PRD:GRU> <pos=ZAL> <NP>`), NewChunkTokenFactory(false))
	require.NotEmpty(t, re.FindAll(toks))
	// Full path with ZAL+SUB fuse → NPP via numeral list
	tokens2 := []*languagetool.AnalyzedTokenReadings{
		atrPos("Von", "PRP:DAT", 0),
		atrPos("ursprünglich", "ADJ:PRD:GRU", 4),
		atrPos("drei", "ZAL", 17),
		atrPos("Almhütten", "SUB:DAT:PLU:FEM", 22),
	}
	NewGermanChunker().AddChunkTags(tokens2)
	require.Contains(t, tokens2[2].GetChunkTags(), "NPP")
	require.Contains(t, tokens2[3].GetChunkTags(), "NPP")
}

// Java: sowohl Tom als auch Maria → NPP
func TestGermanChunker_SowohlEigAlsAuchEig_NPP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Sowohl", "KON", 0),
		atrPos("Tom", "EIG:NOM:SIN:MAS", 7),
		atrPos("als", "KON", 11),
		atrPos("auch", "ADV", 15),
		atrPos("Maria", "EIG:NOM:SIN:FEM", 20),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "NPP")
	require.Contains(t, tokens[1].GetChunkTags(), "NPP")
	require.Contains(t, tokens[4].GetChunkTags(), "NPP")
}

// Java: sowohl er als auch seine Schwester → NPP
func TestGermanChunker_SowohlPronAlsAuch_NPP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("sowohl", "KON", 0),
		atrPos("er", "PRO:PER:NOM:SIN:MAS", 7),
		atrPos("als", "KON", 10),
		atrPos("auch", "ADV", 14),
		atrPos("seine", "PRO:POS:NOM:SIN:FEM", 19),
		atrPos("Schwester", "SUB:NOM:SIN:FEM", 25),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "NPP")
	require.Contains(t, tokens[5].GetChunkTags(), "NPP")
}

// Java: sowohl sein Vater als auch seine Mutter → NPP
func TestGermanChunker_SowohlNPAlsAuchNP_NPP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("sowohl", "KON", 0),
		atrPos("sein", "PRO:POS:NOM:SIN:MAS", 7),
		atrPos("Vater", "SUB:NOM:SIN:MAS", 12),
		atrPos("als", "KON", 18),
		atrPos("auch", "ADV", 22),
		atrPos("seine", "PRO:POS:NOM:SIN:FEM", 27),
		atrPos("Mutter", "SUB:NOM:SIN:FEM", 33),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "NPP")
	require.Contains(t, tokens[2].GetChunkTags(), "NPP")
	require.Contains(t, tokens[6].GetChunkTags(), "NPP")
}

// Java SYNTAX_EXPANSION &prozent;: "37 Euro" → NPS+NPP
func TestGermanChunker_NumberEuro_NPS_NPP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("37", "ZAL", 0),
		atrPos("Euro", "SUB:NOM:PLU:MAS", 3),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "NPS")
	require.Contains(t, tokens[0].GetChunkTags(), "NPP")
}

// Java REGEXES2: <NP> <,> <NP> <,> <NP> → NPP
// assertFullChunks: "Kommentare/NPP ,/NPP Korrekturen/NPP ,/NPP Kritik/NPP …"
func TestGermanChunker_KommentareKorrekturenKritik_NPP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Kommentare", "SUB:NOM:PLU:MAS", 0),
		atrPos(",", "PKT", 10),
		atrPos("Korrekturen", "SUB:NOM:PLU:FEM", 12),
		atrPos(",", "PKT", 24),
		atrPos("Kritik", "SUB:NOM:SIN:FEM", 26),
	}
	NewGermanChunker().AddChunkTags(tokens)
	for i, want := range []string{"Kommentare", ",", "Korrekturen", ",", "Kritik"} {
		require.Equal(t, want, tokens[i].GetToken())
		require.Contains(t, tokens[i].GetChunkTags(), "NPP", "token %q tags %v", want, tokens[i].GetChunkTags())
	}
}

// Java REGEXES2: <dass> <sie> <wie> <NP> → NPP
func TestGermanChunker_DassSieWieNP_NPP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("dass", "KOUS", 0),
		atrPos("sie", "PRO:PER:NOM:PLU:*", 5),
		atrPos("wie", "KOKOM", 9),
		atrPos("ein", "ART:INDEF:NOM:SIN:NEU", 13),
		atrPos("Spiel", "SUB:NOM:SIN:NEU", 17),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "NPP")
	require.Contains(t, tokens[4].GetChunkTags(), "NPP")
}

// Java REGEXES2: <pos=PLU> <die> <Regel> → NPP
// assertFullChunks: "… Platzwunden/NPP die/NPP Regel/NPP …"
func TestGermanChunker_PLUDieRegel_NPP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Platzwunden", "SUB:NOM:PLU:FEM", 0),
		atrPos("die", "ART:DEF:NOM:SIN:FEM", 12),
		atrPos("Regel", "SUB:NOM:SIN:FEM", 16),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "NPP")
	require.Contains(t, tokens[2].GetChunkTags(), "NPP")
}

// Java REGEXES2: <eine> <menge> <NP>+ → NPP (overwrite)
// assertFullChunks: "Sie kennt eine/NPP Menge/NPP englischer/NPP Wörter/NPP"
func TestGermanChunker_EineMengeNP_NPP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("eine", "ART:INDEF:AKK:SIN:FEM", 0),
		atrPos("Menge", "SUB:AKK:SIN:FEM", 5),
		atrPos("englischer", "ADJ:GEN:PLU:NEU", 11),
		atrPos("Wörter", "SUB:GEN:PLU:NEU", 22),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "NPP")
	require.Contains(t, tokens[1].GetChunkTags(), "NPP")
	// trailing genitive NP may be NPP on full span
	require.Contains(t, tokens[3].GetChunkTags(), "NPP")
}

// Java REGEXES2: <pos=PRP> <pos=ADV> <regex=\d+> <NP> → PP
// assertFullChunks: "Mit/PP über/PP 1000/PP Handschriften/PP …"
func TestGermanChunker_MitUeber1000Handschriften_PP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Mit", "PRP:DAT", 0),
		atrPos("über", "ADV", 4),
		atrPos("1000", "CARD", 9),
		atrPos("Handschriften", "SUB:DAT:PLU:FEM", 14),
	}
	NewGermanChunker().AddChunkTags(tokens)
	for i := 0; i < 4; i++ {
		require.Contains(t, tokens[i].GetChunkTags(), "PP", "token %q tags %v", tokens[i].GetToken(), tokens[i].GetChunkTags())
	}
}

// Java REGEXES2: <pos=PRP> <pos=ART> <pos=ADV>* <pos=ADJ> <NP> → PP
// assertFullChunks: "Bei/PP den/PP sehr/PP niedrigen/PP Oberflächentemperaturen/PP …"
func TestGermanChunker_BeiSehrNiedrigen_PP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Bei", "PRP:DAT", 0),
		atrPos("den", "ART:DEF:DAT:PLU:FEM", 4),
		atrPos("sehr", "ADV", 8),
		atrPos("niedrigen", "ADJ:DAT:PLU:FEM", 13),
		atrPos("Oberflächentemperaturen", "SUB:DAT:PLU:FEM", 23),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "PP")
	require.Contains(t, tokens[4].GetChunkTags(), "PP")
}

// Java REGEXES2: <pos=PRP> <pos=ADJ> <und|oder|sowie> <NP> → PP
// assertFullChunks: "Nach/PP sachlichen/PP und/PP militärischen/PP Kriterien/PP …"
func TestGermanChunker_NachSachlichenUnd_PP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Nach", "PRP:DAT", 0),
		atrPos("sachlichen", "ADJ:DAT:PLU:NEU", 5),
		atrPos("und", "KON", 16),
		atrPos("militärischen", "ADJ:DAT:PLU:NEU", 20),
		atrPos("Kriterien", "SUB:DAT:PLU:NEU", 34),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "PP")
	require.Contains(t, tokens[4].GetChunkTags(), "PP")
}

// Java REGEXES2: <pos=PRP> <pos=PA1> <NP> → PP
// assertFullChunks: "… über/PP laufende/PP Sanierungsmaßnahmen/PP"
func TestGermanChunker_UeberLaufende_PP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("über", "PRP:AKK", 0),
		atrPos("laufende", "PA1:AKK:PLU:FEM:GRU:SOL", 5),
		atrPos("Sanierungsmaßnahmen", "SUB:AKK:PLU:FEM", 14),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "PP")
	require.Contains(t, tokens[2].GetChunkTags(), "PP")
}

// Java REGEXES2: <die> <pos=ADJ> <Jahre|…> → PP
// assertFullChunks: "Die/PP darauffolgenden/PP Jahre/PP war es kalt"
// Java assertChunks uses contains (not exclusive): plural B-NP also gets NPP earlier
// in REGEXES2, so both NPP and PP are present (assertFullChunks "waren" case expects NPP).
func TestGermanChunker_DieDarauffolgendenJahre_PP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Die", "ART:DEF:NOM:PLU:NEU", 0),
		atrPos("darauffolgenden", "ADJ:NOM:PLU:NEU", 4),
		atrPos("Jahre", "SUB:NOM:PLU:NEU", 20),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "PP")
	require.Contains(t, tokens[1].GetChunkTags(), "PP")
	require.Contains(t, tokens[2].GetChunkTags(), "PP")
}

// Java assertFullChunks: "Die/NPP darauffolgenden/NPP Jahre/NPP waren kalt"
// Same surface as PP case; dual tags (NPP from plural NP pass + PP from time-unit pattern).
func TestGermanChunker_DieDarauffolgendenJahre_NPP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Die", "ART:DEF:NOM:PLU:NEU", 0),
		atrPos("darauffolgenden", "ADJ:NOM:PLU:NEU", 4),
		atrPos("Jahre", "SUB:NOM:PLU:NEU", 20),
	}
	NewGermanChunker().AddChunkTags(tokens)
	for i, tok := range []string{"Die", "darauffolgenden", "Jahre"} {
		require.Equal(t, tok, tokens[i].GetToken())
		require.Contains(t, tokens[i].GetChunkTags(), "NPP", "token %q tags=%v", tok, tokens[i].GetChunkTags())
		// Java also assigns PP for die+ADJ+time-unit (contains check)
		require.Contains(t, tokens[i].GetChunkTags(), "PP", "token %q tags=%v", tok, tokens[i].GetChunkTags())
	}
}

// Java REGEXES2 genitive: <chunk=NPS>+ <und> <chunk=NP[SP] & GEN>+ → NPS
// assertFullChunks: "die/NPS Pyramide/NPS des/NPS Friedens/NPS und/NPS der/NPS Eintracht/NPS"
func TestGermanChunker_PyramideDesFriedensUndDerEintracht_NPS(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("die", "ART:DEF:NOM:SIN:FEM", 0),
		atrPos("Pyramide", "SUB:NOM:SIN:FEM", 4),
		atrPos("des", "ART:DEF:GEN:SIN:MAS", 13),
		atrPos("Friedens", "SUB:GEN:SIN:MAS", 17),
		atrPos("und", "KON", 26),
		atrPos("der", "ART:DEF:GEN:SIN:FEM", 30),
		atrPos("Eintracht", "SUB:GEN:SIN:FEM", 34),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "NPS")
	require.Contains(t, tokens[3].GetChunkTags(), "NPS")
	require.Contains(t, tokens[6].GetChunkTags(), "NPS")
}

// Java genitive: Autor/NPS der/NPS beiden/NPS Bücher/NPS
func TestGermanChunker_AutorDerBeidenBuecher_NPS(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Autor", "SUB:NOM:SIN:MAS", 0),
		atrPos("der", "ART:DEF:GEN:PLU:NEU", 6),
		atrPos("beiden", "PRO:POS:GEN:PLU:NEU", 10),
		atrPos("Bücher", "SUB:GEN:PLU:NEU", 17),
	}
	NewGermanChunker().AddChunkTags(tokens)
	for i := 0; i < len(tokens); i++ {
		require.Contains(t, tokens[i].GetChunkTags(), "NPS", "token %q tags %v", tokens[i].GetToken(), tokens[i].GetChunkTags())
	}
}

// Java genitive: Autor/NPS der/NPS ersten/NPS beiden/NPS Bücher/NPS
func TestGermanChunker_AutorDerErstenBeidenBuecher_NPS(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Autor", "SUB:NOM:SIN:MAS", 0),
		atrPos("der", "ART:DEF:GEN:PLU:NEU", 6),
		atrPos("ersten", "ADJ:GEN:PLU:NEU", 10),
		atrPos("beiden", "PRO:POS:GEN:PLU:NEU", 17),
		atrPos("Bücher", "SUB:GEN:PLU:NEU", 24),
	}
	NewGermanChunker().AddChunkTags(tokens)
	for i := 0; i < len(tokens); i++ {
		require.Contains(t, tokens[i].GetChunkTags(), "NPS", "token %q tags %v", tokens[i].GetToken(), tokens[i].GetChunkTags())
	}
}

// Java REGEXES2: <pos=PRP> <NP> <NP> <und|oder> <NP> → PP
// assertFullChunks: "durch/PP Einsatz/PP größerer/PP Maschinen/PP und/PP bessere/PP Kapazitätsplanung/PP"
func TestGermanChunker_DurchEinsatzMaschinenUnd_PP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("durch", "PRP:AKK", 0),
		atrPos("Einsatz", "SUB:AKK:SIN:MAS", 6),
		atrPos("größerer", "ADJ:GEN:PLU:FEM:GRU:SOL", 14),
		atrPos("Maschinen", "SUB:GEN:PLU:FEM", 23),
		atrPos("und", "KON", 33),
		atrPos("bessere", "ADJ:AKK:SIN:FEM:GRU:SOL", 37),
		atrPos("Kapazitätsplanung", "SUB:AKK:SIN:FEM", 45),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "PP")
	require.Contains(t, tokens[3].GetChunkTags(), "PP")
	require.Contains(t, tokens[6].GetChunkTags(), "PP")
}

// Java REGEXES2: <pos=PRP> (<NP>)+ → PP
// assertFullChunks: "… für/PP Ärzte/PP und/PP Ärztinnen/PP festgestellte/PP Risikoprofil/PP"
func TestGermanChunker_FuerAerzteUndAerztinnenFestgestellte_PP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("für", "PRP:AKK", 0),
		atrPos("Ärzte", "SUB:AKK:PLU:MAS", 4),
		atrPos("und", "KON", 10),
		atrPos("Ärztinnen", "SUB:AKK:PLU:FEM", 14),
		atrPos("festgestellte", "PA2:AKK:SIN:NEU:GRU:SOL", 24),
		atrPos("Risikoprofil", "SUB:AKK:SIN:NEU", 38),
	}
	NewGermanChunker().AddChunkTags(tokens)
	for i := 0; i < len(tokens); i++ {
		require.Contains(t, tokens[i].GetChunkTags(), "PP", "token %q tags %v", tokens[i].GetToken(), tokens[i].GetChunkTags())
	}
}

// Java assertFullChunks: "gute/NPS Laune/NPS in/PP chemischen/PP Komplexverbindungen/PP"
func TestGermanChunker_GuteLauneInChemischen_NPS_PP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("gute", "ADJ:NOM:SIN:FEM:GRU:SOL", 0),
		atrPos("Laune", "SUB:NOM:SIN:FEM", 5),
		atrPos("in", "PRP:DAT", 11),
		atrPos("chemischen", "ADJ:DAT:PLU:FEM", 14),
		atrPos("Komplexverbindungen", "SUB:DAT:PLU:FEM", 25),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "NPS")
	require.Contains(t, tokens[1].GetChunkTags(), "NPS")
	require.Contains(t, tokens[2].GetChunkTags(), "PP")
	require.Contains(t, tokens[4].GetChunkTags(), "PP")
}

// Java: die/NPP Arbeitsplätze + dass/NPP sie/NPP wie/NPP ein/NPP Spiel/NPP
func TestGermanChunker_ArbeitsplaetzeDassSieWieSpiel_NPP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("die", "ART:DEF:AKK:PLU:MAS", 0),
		atrPos("Arbeitsplätze", "SUB:AKK:PLU:MAS", 4),
		atrPos("so", "ADV", 18),
		atrPos("umgestalten", "VER:INF:NON", 21),
		atrPos(",", "PKT", 32),
		atrPos("dass", "KOUS", 34),
		atrPos("sie", "PRO:PER:NOM:PLU:*", 39),
		atrPos("wie", "KOKOM", 43),
		atrPos("ein", "ART:INDEF:NOM:SIN:NEU", 47),
		atrPos("Spiel", "SUB:NOM:SIN:NEU", 51),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "NPP")
	require.Contains(t, tokens[1].GetChunkTags(), "NPP")
	require.Contains(t, tokens[5].GetChunkTags(), "NPP")
	require.Contains(t, tokens[9].GetChunkTags(), "NPP")
}

// Java: die/NPS größte/NPS und/NPS erfolgreichste/NPS Erfindung/NPS
func TestGermanChunker_GroessteUndErfolgreichste_NPS(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("die", "ART:DEF:NOM:SIN:FEM", 0),
		atrPos("größte", "ADJ:NOM:SIN:FEM", 4),
		atrPos("und", "KON", 12),
		atrPos("erfolgreichste", "ADJ:NOM:SIN:FEM", 16),
		atrPos("Erfindung", "SUB:NOM:SIN:FEM", 31),
	}
	NewGermanChunker().AddChunkTags(tokens)
	for i := 0; i < len(tokens); i++ {
		require.Contains(t, tokens[i].GetChunkTags(), "NPS", "token %q tags %v", tokens[i].GetToken(), tokens[i].GetChunkTags())
	}
}

// Java: deren/NPS Bestimmung/NPS und/NPS Funktion/NPS
func TestGermanChunker_DerenBestimmungUndFunktion_NPS(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("deren", "PRO:POS:GEN:PLU:*", 0),
		atrPos("Bestimmung", "SUB:NOM:SIN:FEM", 6),
		atrPos("und", "KON", 17),
		atrPos("Funktion", "SUB:NOM:SIN:FEM", 21),
	}
	NewGermanChunker().AddChunkTags(tokens)
	// <deren> <B-NP !PLU> <und> <B-NP>* — bare SUB+ also tags Funktion as B-NP (overlap),
	// so B-NP* includes Funktion → full span NPS (Java assertFullChunks).
	require.Contains(t, tokens[0].GetChunkTags(), "NPS")
	require.Contains(t, tokens[1].GetChunkTags(), "NPS")
	require.Contains(t, tokens[2].GetChunkTags(), "NPS") // und
	require.Contains(t, tokens[3].GetChunkTags(), "NPS") // Funktion
	require.Contains(t, tokens[3].GetChunkTags(), "B-NP")
}

// Java: Rekonstruktionen/NPP oder/NPP der/NPP Wiederaufbau/NPP
func TestGermanChunker_RekonstruktionenOderWiederaufbau_NPP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Rekonstruktionen", "SUB:NOM:PLU:FEM", 0),
		atrPos("oder", "KON", 17),
		atrPos("der", "ART:DEF:NOM:SIN:MAS", 22),
		atrPos("Wiederaufbau", "SUB:NOM:SIN:MAS", 26),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "NPP")
	require.Contains(t, tokens[3].GetChunkTags(), "NPP")
}

// Java: die/NPP Kenntnisse/NPP der/NPP Sprache/NPP
func TestGermanChunker_KenntnisseDerSprache_NPP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("die", "ART:DEF:NOM:PLU:FEM", 0),
		atrPos("Kenntnisse", "SUB:NOM:PLU:FEM", 4),
		atrPos("der", "ART:DEF:GEN:SIN:FEM", 15),
		atrPos("Sprache", "SUB:GEN:SIN:FEM", 19),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "NPP")
	require.Contains(t, tokens[1].GetChunkTags(), "NPP")
	require.Contains(t, tokens[2].GetChunkTags(), "NPP")
	require.Contains(t, tokens[3].GetChunkTags(), "NPP")
}

// Java NPP + NPS&GEN requires pos=GEN — bare DAT "dem" must not invent-merge into NPP.
func TestGermanChunker_KenntnisseDemMann_NoInventGenOnDat(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("die", "ART:DEF:NOM:PLU:FEM", 0),
		atrPos("Kenntnisse", "SUB:NOM:PLU:FEM", 4),
		atrPos("dem", "ART:DEF:DAT:SIN:MAS", 15),
		atrPos("Mann", "SUB:DAT:SIN:MAS", 19),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "NPP")
	require.Contains(t, tokens[1].GetChunkTags(), "NPP")
	// dem/Mann are separate DAT NPS (or B-NP) — not absorbed via invent genitive NPP merge
	require.NotContains(t, tokens[2].GetChunkTags(), "NPP", "dem tags=%v", tokens[2].GetChunkTags())
	require.NotContains(t, tokens[3].GetChunkTags(), "NPP", "Mann tags=%v", tokens[3].GetChunkTags())
}

// Java REGEXES2 genitive extend is surface <der> only — not dem/den invent.
// "Autor dem Buch" must not become one NPS via the <der> NP path.
func TestGermanChunker_AutorDemBuch_NoInventDerPathOnDem(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Autor", "SUB:NOM:SIN:MAS", 0),
		atrPos("dem", "ART:DEF:DAT:SIN:NEU", 6),
		atrPos("Buch", "SUB:DAT:SIN:NEU", 10),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "NPS", "Autor tags=%v", tokens[0].GetChunkTags())
	// dem is not surface "der" → no genitive-der extend attaching dem as I-NP under Autor
	require.NotContains(t, tokens[0].GetChunkTags(), "I-NP")
	require.NotContains(t, tokens[1].GetChunkTags(), "I-NP", "dem tags=%v", tokens[1].GetChunkTags())
}

// Java: einschließlich/PP der/PP biologischen/PP und/PP sozialen/PP Grundlagen/PP
func TestGermanChunker_EinschliesslichBiologischenUnd_PP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("einschließlich", "PRP:GEN", 0),
		atrPos("der", "ART:DEF:GEN:PLU:FEM", 15),
		atrPos("biologischen", "ADJ:GEN:PLU:FEM", 19),
		atrPos("und", "KON", 32),
		atrPos("sozialen", "ADJ:GEN:PLU:FEM", 36),
		atrPos("Grundlagen", "SUB:GEN:PLU:FEM", 45),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "PP")
	require.Contains(t, tokens[5].GetChunkTags(), "PP")
}

// Java: für/PP die/PP Stadtteile/PP und/PP selbständigen/PP Ortsteile/PP
func TestGermanChunker_FuerStadtteileUndOrtsteile_PP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("für", "PRP:AKK", 0),
		atrPos("die", "ART:DEF:AKK:PLU:MAS", 4),
		atrPos("Stadtteile", "SUB:AKK:PLU:MAS", 8),
		atrPos("und", "KON", 19),
		atrPos("selbständigen", "ADJ:AKK:PLU:MAS", 23),
		atrPos("Ortsteile", "SUB:AKK:PLU:MAS", 37),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "PP")
	require.Contains(t, tokens[2].GetChunkTags(), "PP")
	require.Contains(t, tokens[5].GetChunkTags(), "PP")
}

// Java genitive ADV PA2: Teil/NPS der/NPS dort/NPS ausgestellten/NPS Bestände/NPS
func TestGermanChunker_TeilDerDortAusgestellten_NPS(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Teil", "SUB:NOM:SIN:MAS", 0),
		atrPos("der", "ART:DEF:GEN:PLU:MAS", 5),
		atrPos("dort", "ADV", 9),
		atrPos("ausgestellten", "PA2:GEN:PLU:MAS:GRU:DEF", 14),
		atrPos("Bestände", "SUB:GEN:PLU:MAS", 28),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "NPS")
	require.Contains(t, tokens[4].GetChunkTags(), "NPS")
}

// Java: Isolation/NPP und/NPP ihre/NPP Überwindung/NPP
// Uses B-NP und NP path (SUB und B-NP excludes ihre intentionally).
func TestGermanChunker_IsolationUndIhreUeberwindung_NPP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Isolation", "SUB:NOM:SIN:FEM", 0),
		atrPos("und", "KON", 10),
		atrPos("ihre", "PRO:POS:NOM:SIN:FEM", 14),
		atrPos("Überwindung", "SUB:NOM:SIN:FEM", 19),
	}
	NewGermanChunker().AddChunkTags(tokens)
	for i := 0; i < len(tokens); i++ {
		require.Contains(t, tokens[i].GetChunkTags(), "NPP", "token %q tags %v", tokens[i].GetToken(), tokens[i].GetChunkTags())
	}
}

// Java late REGEXES2: <chunk=NPS> <pos=PRO> <pos=ADJ> <pos=ADJ> <NP> → NPS
// Java GermanChunkerTest has this assertFullChunks commented with "//?" — REGEXES1 fuses
// "dieser relativ kleinen Verwaltungseinheiten" as one B-NP…I-NP span, so <NP> after two ADJs
// does not start at a B-NP and the extension pattern does not fire. Head stays NPS.
func TestGermanChunker_HoheZahlDieserRelativ_NPS(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("die", "ART:DEF:NOM:SIN:FEM", 0),
		atrPos("hohe", "ADJ:NOM:SIN:FEM", 4),
		atrPos("Zahl", "SUB:NOM:SIN:FEM", 9),
		atrPos("dieser", "PRO:DEM:GEN:PLU:FEM", 14),
		atrPos("relativ", "ADJ:PRD:GRU", 21),
		atrPos("kleinen", "ADJ:GEN:PLU:FEM", 29),
		atrPos("Verwaltungseinheiten", "SUB:GEN:PLU:FEM", 37),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "NPS")
	require.Contains(t, tokens[2].GetChunkTags(), "NPS")
	// Fused genitive tail: plural NPP (or I-NP), not invent full-span NPS
	require.NotContains(t, tokens[6].GetChunkTags(), "NPS", "tail tags %v", tokens[6].GetChunkTags())
}

// Java: In/PP den/PP alten/PP Religionen/PP ,/PP Mythen/PP und/PP Sagen/PP
func TestGermanChunker_InAltenReligionenMythenSagen_PP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("In", "PRP:DAT", 0),
		atrPos("den", "ART:DEF:DAT:PLU:FEM", 3),
		atrPos("alten", "ADJ:DAT:PLU:FEM", 7),
		atrPos("Religionen", "SUB:DAT:PLU:FEM", 13),
		atrPos(",", "PKT", 24),
		atrPos("Mythen", "SUB:DAT:PLU:MAS", 26),
		atrPos("und", "KON", 33),
		atrPos("Sagen", "SUB:DAT:PLU:FEM", 37),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "PP")
	require.Contains(t, tokens[3].GetChunkTags(), "PP")
	require.Contains(t, tokens[5].GetChunkTags(), "PP")
	require.Contains(t, tokens[7].GetChunkTags(), "PP")
}

// Java: Gesteigerte/B Effizienz/I durch/PP Einsatz/… (B-NP + multi-NP PP already covered for durch)
func TestGermanChunker_GesteigerteEffizienzDurchEinsatz_B_PP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Gesteigerte", "ADJ:NOM:SIN:FEM:GRU:SOL", 0),
		atrPos("Effizienz", "SUB:NOM:SIN:FEM", 12),
		atrPos("durch", "PRP:AKK", 22),
		atrPos("Einsatz", "SUB:AKK:SIN:MAS", 28),
		atrPos("größerer", "ADJ:GEN:PLU:FEM:GRU:SOL", 36),
		atrPos("Maschinen", "SUB:GEN:PLU:FEM", 45),
		atrPos("und", "KON", 55),
		atrPos("bessere", "ADJ:AKK:SIN:FEM:GRU:SOL", 59),
		atrPos("Kapazitätsplanung", "SUB:AKK:SIN:FEM", 67),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "B-NP")
	require.Contains(t, tokens[1].GetChunkTags(), "I-NP")
	require.Contains(t, tokens[2].GetChunkTags(), "PP")
	require.Contains(t, tokens[8].GetChunkTags(), "PP")
}

// Java assertFullChunks: "Geräte/B , deren/NPS Bestimmung/NPS und/NPS Funktion/NPS …"
func TestGermanChunker_GeraeteDerenBestimmung_B_NPS(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Geräte", "SUB:NOM:PLU:NEU", 0),
		atrPos(",", "PKT", 7),
		atrPos("deren", "PRO:POS:GEN:PLU:*", 9),
		atrPos("Bestimmung", "SUB:NOM:SIN:FEM", 15),
		atrPos("und", "KON", 26),
		atrPos("Funktion", "SUB:NOM:SIN:FEM", 30),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "B-NP")
	require.Contains(t, tokens[2].GetChunkTags(), "NPS") // deren
	require.Contains(t, tokens[3].GetChunkTags(), "NPS") // Bestimmung
	require.Contains(t, tokens[4].GetChunkTags(), "NPS") // und
	require.Contains(t, tokens[5].GetChunkTags(), "I-NP") // Funktion after SUB und SUB
}

// Java assertFullChunks tags "Stephen King/NPS" after full Morphy (names as SUB/EIG).
// REGEXES1 has no bare EIG+EIG joiner (only Herr|Frau EIG+). With SUB SIN tags each
// name is its own B-NP → NPS — not invent multi-token name NP.
func TestGermanChunker_StephenKing_NPS(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Stephen", "SUB:NOM:SIN:MAS", 0),
		atrPos("King", "SUB:NOM:SIN:MAS", 8),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "NPS")
	require.Contains(t, tokens[1].GetChunkTags(), "NPS")
}

// Java assertFullChunks: "Da sitzen drei/NPP Katzen/NPP"
// REGEXES2: <zwei|…|zwölf> <chunk=I-NP|B-NP span> → NPP
func TestGermanChunker_DreiKatzen_NPP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Da", "ADV", 0),
		atrPos("sitzen", "VER:3:PLU:PRÄ:SFT", 3),
		atrPos("drei", "ZAL", 10),
		atrPos("Katzen", "SUB:NOM:PLU:FEM", 15),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[2].GetChunkTags(), "NPP")
	require.Contains(t, tokens[3].GetChunkTags(), "NPP")
}

// Java assertFullChunks: "Da sind er/NPP und/NPP seine/NPP Schwester/NPP"
// REGEXES2: <ich|du|er|…> <und|oder|sowie> <NP> → NPP
func TestGermanChunker_ErUndSeineSchwester_NPP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Da", "ADV", 0),
		atrPos("sind", "VER:3:PLU:PRÄ:NON", 3),
		atrPos("er", "PRO:PER:NOM:SIN:MAS", 8),
		atrPos("und", "KON", 11),
		atrPos("seine", "PRO:POS:NOM:SIN:FEM", 15),
		atrPos("Schwester", "SUB:NOM:SIN:FEM", 21),
	}
	NewGermanChunker().AddChunkTags(tokens)
	for i, tok := range []string{"er", "und", "seine", "Schwester"} {
		idx := i + 2
		require.Equal(t, tok, tokens[idx].GetToken())
		require.Contains(t, tokens[idx].GetChunkTags(), "NPP", "token %q tags=%v", tok, tokens[idx].GetChunkTags())
	}
}

// Java assertFullChunks: "Ein/NPP Hund/NPP und/NPP eine/NPP Katze/NPP stehen dort"
// REGEXES2: B-NP und NP → NPP (coordinated NPs)
func TestGermanChunker_EinHundUndEineKatze_NPP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Ein", "ART:IND:NOM:SIN:MAS", 0),
		atrPos("Hund", "SUB:NOM:SIN:MAS", 4),
		atrPos("und", "KON", 9),
		atrPos("eine", "ART:IND:NOM:SIN:FEM", 13),
		atrPos("Katze", "SUB:NOM:SIN:FEM", 18),
		atrPos("stehen", "VER:3:PLU:PRÄ:SFT", 24),
		atrPos("dort", "ADV", 31),
	}
	NewGermanChunker().AddChunkTags(tokens)
	for i, tok := range []string{"Ein", "Hund", "und", "eine", "Katze"} {
		require.Equal(t, tok, tokens[i].GetToken())
		require.Contains(t, tokens[i].GetChunkTags(), "NPP", "token %q tags=%v", tok, tokens[i].GetChunkTags())
	}
}

// Java assertFullChunks:
// "Der/NPS letzte/NPS der/NPS vier/NPS großen/NPS Flüsse/NPS ist der/B Nil/I"
// (Nil stays B-NP/I-NP; genitive "der letzte der vier großen Flüsse" → NPS)
func TestGermanChunker_DerLetzteDerVierGrossenFluesse_NPS(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Der", "ART:DEF:NOM:SIN:MAS", 0),
		atrPos("letzte", "ADJ:NOM:SIN:MAS", 4),
		atrPos("der", "ART:DEF:GEN:PLU:MAS", 11),
		atrPos("vier", "ZAL", 15),
		atrPos("großen", "ADJ:GEN:PLU:MAS", 20),
		atrPos("Flüsse", "SUB:GEN:PLU:MAS", 27),
		atrPos("ist", "VER:3:SIN:PRÄ:NON", 34),
		atrPos("der", "ART:DEF:NOM:SIN:MAS", 38),
		atrPos("Nil", "SUB:NOM:SIN:MAS", 42),
	}
	NewGermanChunker().AddChunkTags(tokens)
	for i, tok := range []string{"Der", "letzte", "der", "vier", "großen", "Flüsse"} {
		require.Equal(t, tok, tokens[i].GetToken())
		require.Contains(t, tokens[i].GetChunkTags(), "NPS", "token %q tags=%v", tok, tokens[i].GetChunkTags())
	}
	// der Nil: B-NP / I-NP (Java assert Der/B Nil/I)
	require.Contains(t, tokens[7].GetChunkTags(), "B-NP")
	require.Contains(t, tokens[8].GetChunkTags(), "I-NP")
}

// Java assertFullChunks:
// "Der/B Nil/I ist der/NPS letzte/NPS der/NPS vier/NPS großen/NPS Flüsse/NPS"
// Nil first as ART+SUB B-NP/I-NP; genitive residual after ist is NPS.
func TestGermanChunker_DerNilIstDerLetzteDerVier_NPS(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Der", "ART:DEF:NOM:SIN:MAS", 0),
		atrPos("Nil", "SUB:NOM:SIN:MAS", 4),
		atrPos("ist", "VER:3:SIN:PRÄ:NON", 8),
		atrPos("der", "ART:DEF:NOM:SIN:MAS", 12),
		atrPos("letzte", "ADJ:NOM:SIN:MAS", 16),
		atrPos("der", "ART:DEF:GEN:PLU:MAS", 23),
		atrPos("vier", "ZAL", 27),
		atrPos("großen", "ADJ:GEN:PLU:MAS", 32),
		atrPos("Flüsse", "SUB:GEN:PLU:MAS", 39),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "B-NP")
	require.Contains(t, tokens[1].GetChunkTags(), "I-NP")
	for i, tok := range []string{"der", "letzte", "der", "vier", "großen", "Flüsse"} {
		idx := i + 3
		require.Equal(t, tok, tokens[idx].GetToken())
		require.Contains(t, tokens[idx].GetChunkTags(), "NPS", "token %q tags=%v", tok, tokens[idx].GetChunkTags())
	}
}

// Java assertFullChunks:
// "Die/NPS Krankheit/NPS unserer/NPS heutigen/NPS Städte/NPS und/NPS Siedlungen/NPS ist der/NPS Verkehr/NPS"
// REGEXES2: NPS + PRO:POS + ADJ + NP; NPS + und + GEN NP
func TestGermanChunker_KrankheitUnsererHeutigenStaedte_NPS(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Die", "ART:DEF:NOM:SIN:FEM", 0),
		atrPos("Krankheit", "SUB:NOM:SIN:FEM", 4),
		atrPos("unserer", "PRO:POS:GEN:PLU:FEM", 14),
		atrPos("heutigen", "ADJ:GEN:PLU:FEM", 22),
		atrPos("Städte", "SUB:GEN:PLU:FEM", 31),
		atrPos("und", "KON", 38),
		atrPos("Siedlungen", "SUB:GEN:PLU:FEM", 42),
		atrPos("ist", "VER:3:SIN:PRÄ:NON", 53),
		atrPos("der", "ART:DEF:NOM:SIN:MAS", 57),
		atrPos("Verkehr", "SUB:NOM:SIN:MAS", 61),
	}
	NewGermanChunker().AddChunkTags(tokens)
	for i, tok := range []string{"Die", "Krankheit", "unserer", "heutigen", "Städte", "und", "Siedlungen"} {
		require.Equal(t, tok, tokens[i].GetToken())
		require.Contains(t, tokens[i].GetChunkTags(), "NPS", "token %q tags=%v", tok, tokens[i].GetChunkTags())
	}
	require.Contains(t, tokens[8].GetChunkTags(), "NPS") // der
	require.Contains(t, tokens[9].GetChunkTags(), "NPS") // Verkehr
}

// Java assertFullChunks uses Morphy tags; with ZAL+SUB, REGEXES1 fuses zwei+Wochen as B-NP/I-NP.
// Faithful OpenRegex then tags NPP via <zwei|…> <chunk=I-NP>. PP needs digits/ZAL not fused into NP
// (Java Morphy often leaves numerals outside the NP for the ADV+ZAL+B-NP PP pattern).
func TestGermanChunker_InNurZweiWochen_PP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("In", "PRP:DAT", 0),
		atrPos("nur", "ADV", 3),
		atrPos("zwei", "ZAL", 7),
		atrPos("Wochen", "SUB:DAT:PLU:FEM", 12),
		atrPos("geht", "VER:3:SIN:PRÄ:SFT", 19),
		atrPos("es", "PRO:PER:NOM:SIN:NEU", 24),
		atrPos("los", "ADV", 27),
	}
	NewGermanChunker().AddChunkTags(tokens)
	// Numeral + fused NP → NPP (REGEXES2 surface zwei|drei|…)
	require.Contains(t, tokens[2].GetChunkTags(), "NPP")
	require.Contains(t, tokens[3].GetChunkTags(), "NPP")
	// With unfused NP (Wochen alone B-NP), PP pattern fires — fixture without ZAL on zwei:
	tokens2 := []*languagetool.AnalyzedTokenReadings{
		atrPos("In", "PRP:DAT", 0),
		atrPos("nur", "ADV", 3),
		atrPos("zwei", "ZAL", 7),
		atrPos("Wochen", "SUB:DAT:PLU:FEM", 12),
	}
	// Pre-tag like Morphy when ZAL SUB does not join (Wochen B-NP only): use CARD on zwei
	tokens2[2] = atrPos("zwei", "CARD", 7)
	NewGermanChunker().AddChunkTags(tokens2)
	// zwei CARD + Wochen SUB → Wochen B-NP only; PP needs ZAL on zwei — still no PP.
	// Use surface digits path: In ADV 2 Wochen
	tokens3 := []*languagetool.AnalyzedTokenReadings{
		atrPos("In", "PRP:DAT", 0),
		atrPos("nur", "ADV", 3),
		atrPos("2", "CARD", 7),
		atrPos("Wochen", "SUB:DAT:PLU:FEM", 9),
	}
	NewGermanChunker().AddChunkTags(tokens3)
	for i, tok := range []string{"In", "nur", "2", "Wochen"} {
		require.Equal(t, tok, tokens3[i].GetToken())
		require.Contains(t, tokens3[i].GetChunkTags(), "PP", "token %q tags=%v", tok, tokens3[i].GetChunkTags())
	}
}

// Java assertFullChunks: "Es sind Atome/NPP ,/NPP welche/NPP der/NPP Urstoff/NPP aller/NPP Körper/NPP sind"
// REGEXES2: <,> <die|welche> <NP> <chunk=NPS & pos=GEN>+ → NPP
func TestGermanChunker_AtomeWelcheDerUrstoff_NPP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Es", "PRO:PER:NOM:SIN:NEU", 0),
		atrPos("sind", "VER:3:PLU:PRÄ:NON", 3),
		atrPos("Atome", "SUB:NOM:PLU:NEU", 8),
		atrPos(",", "PKT", 13),
		atrPos("welche", "PRO:REL:NOM:PLU:NEU", 15),
		atrPos("der", "ART:DEF:NOM:SIN:MAS", 22),
		atrPos("Urstoff", "SUB:NOM:SIN:MAS", 26),
		atrPos("aller", "PRO:IND:GEN:PLU:MAS", 34),
		atrPos("Körper", "SUB:GEN:PLU:MAS", 40),
		atrPos("sind", "VER:3:PLU:PRÄ:NON", 47),
	}
	NewGermanChunker().AddChunkTags(tokens)
	// comma + welche + genitive NP span tagged NPP (Java assertFullChunks)
	for i, tok := range []string{",", "welche", "der", "Urstoff", "aller", "Körper"} {
		idx := i + 3
		require.Equal(t, tok, tokens[idx].GetToken())
		require.Contains(t, tokens[idx].GetChunkTags(), "NPP", "token %q", tok)
	}
	// Atome itself is NPP in Java assertFullChunks.
	require.Contains(t, tokens[2].GetChunkTags(), "NPP", "Atome tags=%v", tokens[2].GetChunkTags())
}

// Java assertFullChunks:
// "Teil/NPS der/NPS umfangreichen/NPS dort/NPS ausgestellten/NPS Bestände/NPS …"
// REGEXES2: <chunk=NPS>+ <der> <pos=ADJ> <pos=ADV> <pos=PA2> <NP>
func TestGermanChunker_TeilDerUmfangreichenDort_NPS(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Teil", "SUB:NOM:SIN:MAS", 0),
		atrPos("der", "ART:DEF:GEN:PLU:MAS", 5),
		atrPos("umfangreichen", "ADJ:GEN:PLU:MAS", 9),
		atrPos("dort", "ADV", 23),
		atrPos("ausgestellten", "PA2:GEN:PLU:MAS:GRU:DEF", 28),
		atrPos("Bestände", "SUB:GEN:PLU:MAS", 42),
	}
	NewGermanChunker().AddChunkTags(tokens)
	for i, tok := range []string{"Teil", "der", "umfangreichen", "dort", "ausgestellten", "Bestände"} {
		require.Equal(t, tok, tokens[i].GetToken())
		require.Contains(t, tokens[i].GetChunkTags(), "NPS", "token %q tags=%v", tok, tokens[i].GetChunkTags())
	}
}

// Java assertFullChunks:
// "Ein/NPS Teil/NPS der/NPS umfangreichen/NPS dort/NPS ausgestellten/NPS Bestände/NPS …"
func TestGermanChunker_EinTeilDerUmfangreichenDort_NPS(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Ein", "ART:INDEF:NOM:SIN:MAS", 0),
		atrPos("Teil", "SUB:NOM:SIN:MAS", 4),
		atrPos("der", "ART:DEF:GEN:PLU:MAS", 9),
		atrPos("umfangreichen", "ADJ:GEN:PLU:MAS", 13),
		atrPos("dort", "ADV", 27),
		atrPos("ausgestellten", "PA2:GEN:PLU:MAS:GRU:DEF", 32),
		atrPos("Bestände", "SUB:GEN:PLU:MAS", 46),
	}
	NewGermanChunker().AddChunkTags(tokens)
	for i, tok := range []string{"Ein", "Teil", "der", "umfangreichen", "dort", "ausgestellten", "Bestände"} {
		require.Equal(t, tok, tokens[i].GetToken())
		require.Contains(t, tokens[i].GetChunkTags(), "NPS", "token %q tags=%v", tok, tokens[i].GetChunkTags())
	}
}

// Java assertFullChunks:
// "Eine/NPP Menge/NPP englischer/NPP Wörter/NPP sind aus/PP dem/NPS Lateinischen/NPS abgeleitet."
// contains-semantics: dem/Lateinischen may also keep PP from <pos=PRP> <NP>.
func TestGermanChunker_EineMengeAusDemLateinischen_NPP_PP_NPS(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Eine", "ART:INDEF:NOM:SIN:FEM", 0),
		atrPos("Menge", "SUB:NOM:SIN:FEM", 5),
		atrPos("englischer", "ADJ:GEN:PLU:NEU", 11),
		atrPos("Wörter", "SUB:GEN:PLU:NEU", 22),
		atrPos("sind", "VER:3:PLU:PRÄ:NON", 29),
		atrPos("aus", "PRP:DAT", 34),
		atrPos("dem", "ART:DEF:DAT:SIN:NEU", 38),
		atrPos("Lateinischen", "SUB:DAT:SIN:NEU", 42),
		atrPos("abgeleitet", "PA2:PRD:GRU:VER", 55),
	}
	NewGermanChunker().AddChunkTags(tokens)
	for i, tok := range []string{"Eine", "Menge", "englischer", "Wörter"} {
		require.Equal(t, tok, tokens[i].GetToken())
		require.Contains(t, tokens[i].GetChunkTags(), "NPP", "token %q tags=%v", tok, tokens[i].GetChunkTags())
	}
	require.Contains(t, tokens[5].GetChunkTags(), "PP") // aus
	require.Contains(t, tokens[6].GetChunkTags(), "NPS") // dem
	require.Contains(t, tokens[7].GetChunkTags(), "NPS") // Lateinischen
}

// Java REGEXES2: <der|die|das> <pos=ADJ> <der> <pos=PA1> <pos=SUB> → NPS
// "Das letzte der teilnehmenden Länder"
func TestGermanChunker_DasLetzteDerTeilnehmendenLaender_NPS(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Das", "ART:DEF:NOM:SIN:NEU", 0),
		atrPos("letzte", "ADJ:NOM:SIN:NEU", 4),
		atrPos("der", "ART:DEF:GEN:PLU:NEU", 11),
		atrPos("teilnehmenden", "PA1:GEN:PLU:NEU:GRU:DEF", 15),
		atrPos("Länder", "SUB:GEN:PLU:NEU", 29),
	}
	NewGermanChunker().AddChunkTags(tokens)
	for i, tok := range []string{"Das", "letzte", "der", "teilnehmenden", "Länder"} {
		require.Equal(t, tok, tokens[i].GetToken())
		require.Contains(t, tokens[i].GetChunkTags(), "NPS", "token %q tags=%v", tok, tokens[i].GetChunkTags())
	}
}

// Java REGEXES2: <pos=SUB & pos=PLU> <der> <pos=PA1> <pos=SUB> → NPP
// "Ursachen der vorliegenden Durchblutungsstörung"
func TestGermanChunker_UrsachenDerVorliegenden_NPP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Ursachen", "SUB:NOM:PLU:FEM", 0),
		atrPos("der", "ART:DEF:GEN:SIN:FEM", 9),
		atrPos("vorliegenden", "PA1:GEN:SIN:FEM:GRU:DEF", 13),
		atrPos("Durchblutungsstörung", "SUB:GEN:SIN:FEM", 26),
	}
	NewGermanChunker().AddChunkTags(tokens)
	for i, tok := range []string{"Ursachen", "der", "vorliegenden", "Durchblutungsstörung"} {
		require.Equal(t, tok, tokens[i].GetToken())
		require.Contains(t, tokens[i].GetChunkTags(), "NPP", "token %q tags=%v", tok, tokens[i].GetChunkTags())
	}
}

// Java REGEXES2: <NP> <,> <NP> <,> <wie> <auch> <chunk=NPS>+ → NPP
// "Details, Dialoge, wie auch die Typologie der Charaktere"
func TestGermanChunker_DetailsDialogeWieAuch_NPP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Details", "SUB:NOM:PLU:NEU", 0),
		atrPos(",", "PKT", 7),
		atrPos("Dialoge", "SUB:NOM:PLU:MAS", 9),
		atrPos(",", "PKT", 16),
		atrPos("wie", "KON", 18),
		atrPos("auch", "ADV", 22),
		atrPos("die", "ART:DEF:NOM:SIN:FEM", 27),
		atrPos("Typologie", "SUB:NOM:SIN:FEM", 31),
		atrPos("der", "ART:DEF:GEN:PLU:MAS", 41),
		atrPos("Charaktere", "SUB:GEN:PLU:MAS", 45),
	}
	NewGermanChunker().AddChunkTags(tokens)
	for i, tok := range []string{"Details", ",", "Dialoge", ",", "wie", "auch", "die", "Typologie"} {
		require.Equal(t, tok, tokens[i].GetToken())
		require.Contains(t, tokens[i].GetChunkTags(), "NPP", "token %q tags=%v", tok, tokens[i].GetChunkTags())
	}
}

// Java REGEXES2: <fuer> <in> <pos=EIG> <pos=PA1> <pos=SUB> <und> <pos=SUB> -> PP (overwrite)
// "Fuer in Oesterreich lebende Afrikaner und Afrikanerinnen"
func TestGermanChunker_FuerInOesterreichLebende_PP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Fuer", "PRP:AKK", 0),
		atrPos("in", "PRP:DAT", 5),
		atrPos("Oesterreich", "EIG:DAT:SIN:NEU", 8),
		atrPos("lebende", "PA1:AKK:PLU:MAS:GRU:IND", 20),
		atrPos("Afrikaner", "SUB:AKK:PLU:MAS", 28),
		atrPos("und", "KON", 38),
		atrPos("Afrikanerinnen", "SUB:AKK:PLU:FEM", 42),
	}
	// Use real German surfaces with umlauts (Java pattern surface <fuer> is case-insensitive string match)
	tokens[0] = atrPos("Für", "PRP:AKK", 0)
	tokens[2] = atrPos("Österreich", "EIG:DAT:SIN:NEU", 7)
	NewGermanChunker().AddChunkTags(tokens)
	for i, tok := range []string{"Für", "in", "Österreich", "lebende", "Afrikaner", "und", "Afrikanerinnen"} {
		require.Equal(t, tok, tokens[i].GetToken())
		require.Contains(t, tokens[i].GetChunkTags(), "PP", "token %q tags=%v", tok, tokens[i].GetChunkTags())
	}
}


// Java REGEXES2: <pos=PRP> <pos=ADJ> <pos=PA1> <NP> → PP
// "Aufgrund stark schwankender Absatzmärkte"
func TestGermanChunker_AufgrundStarkSchwankender_PP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Aufgrund", "PRP:GEN", 0),
		atrPos("stark", "ADJ:PRD:GRU", 9),
		atrPos("schwankender", "PA1:GEN:PLU:MAS:GRU:IND", 15),
		atrPos("Absatzmärkte", "SUB:GEN:PLU:MAS", 28),
	}
	NewGermanChunker().AddChunkTags(tokens)
	for i, tok := range []string{"Aufgrund", "stark", "schwankender", "Absatzmärkte"} {
		require.Equal(t, tok, tokens[i].GetToken())
		require.Contains(t, tokens[i].GetChunkTags(), "PP", "token %q tags=%v", tok, tokens[i].GetChunkTags())
	}
}

// Java assertFullChunks: "… Knochenbrüche/NPP und/NPP Platzwunden/NPP die/NPP Regel/NPP …"
// REGEXES2: <pos=PLU> <die> <Regel> → NPP (and und-coordination of plural NPs)
func TestGermanChunker_KnochenbruecheUndPlatzwundenDieRegel_NPP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Knochenbrüche", "SUB:NOM:PLU:MAS", 0),
		atrPos("und", "KON", 14),
		atrPos("Platzwunden", "SUB:NOM:PLU:FEM", 18),
		atrPos("die", "ART:DEF:NOM:SIN:FEM", 30),
		atrPos("Regel", "SUB:NOM:SIN:FEM", 34),
	}
	NewGermanChunker().AddChunkTags(tokens)
	for i, tok := range []string{"Knochenbrüche", "und", "Platzwunden", "die", "Regel"} {
		require.Equal(t, tok, tokens[i].GetToken())
		require.Contains(t, tokens[i].GetChunkTags(), "NPP", "token %q tags=%v", tok, tokens[i].GetChunkTags())
	}
}


// Java assertFullChunks residual: "… im/PP Weg/NPS"
// contains-semantics: Weg may also keep PP from <pos=PRP> <NP>
func TestGermanChunker_ImWeg_PP_NPS(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("im", "PRP:DAT", 0),
		atrPos("Weg", "SUB:DAT:SIN:MAS", 3),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "PP")
	require.Contains(t, tokens[1].GetChunkTags(), "NPS")
}

// Java assertFullChunks: "Programme/B , in/PP deren/PP deutschen/PP Installationen/PP …"
func TestGermanChunker_ProgrammeInDerenInstallationen_B_PP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Programme", "SUB:NOM:PLU:NEU", 0),
		atrPos(",", "PKT", 9),
		atrPos("in", "PRP:DAT", 11),
		atrPos("deren", "PRO:POS:DAT:PLU:FEM", 14),
		atrPos("deutschen", "ADJ:DAT:PLU:FEM", 20),
		atrPos("Installationen", "SUB:DAT:PLU:FEM", 30),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "B-NP")
	for i, tok := range []string{"in", "deren", "deutschen", "Installationen"} {
		idx := i + 2
		require.Equal(t, tok, tokens[idx].GetToken())
		require.Contains(t, tokens[idx].GetChunkTags(), "PP", "token %q tags=%v", tok, tokens[idx].GetChunkTags())
	}
}

// Java assertFullChunks: "Funktionen/NPP des/NPP Körpers/NPP einschließlich/PP der/PP biologischen/PP und/PP sozialen/PP Grundlagen/PP"
func TestGermanChunker_FunktionenDesKoerpersEinschliesslich_NPP_PP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Funktionen", "SUB:NOM:PLU:FEM", 0),
		atrPos("des", "ART:DEF:GEN:SIN:MAS", 11),
		atrPos("Körpers", "SUB:GEN:SIN:MAS", 15),
		atrPos("einschließlich", "PRP:GEN", 23),
		atrPos("der", "ART:DEF:GEN:PLU:FEM", 38),
		atrPos("biologischen", "ADJ:GEN:PLU:FEM", 42),
		atrPos("und", "KON", 55),
		atrPos("sozialen", "ADJ:GEN:PLU:FEM", 59),
		atrPos("Grundlagen", "SUB:GEN:PLU:FEM", 68),
	}
	NewGermanChunker().AddChunkTags(tokens)
	for i, tok := range []string{"Funktionen", "des", "Körpers"} {
		require.Equal(t, tok, tokens[i].GetToken())
		require.Contains(t, tokens[i].GetChunkTags(), "NPP", "token %q tags=%v", tok, tokens[i].GetChunkTags())
	}
	for i, tok := range []string{"einschließlich", "der", "biologischen", "und", "sozialen", "Grundlagen"} {
		idx := i + 3
		require.Equal(t, tok, tokens[idx].GetToken())
		require.Contains(t, tokens[idx].GetChunkTags(), "PP", "token %q tags=%v", tok, tokens[idx].GetChunkTags())
	}
}


// Java assertFullChunks: "Das/NPS Dokument/NPS umfasst das für/PP Ärzte/PP und/PP Ärztinnen/PP festgestellte/PP Risikoprofil/PP"
// "das" before für stays O (not absorbed into PP without being NP head of PRP pattern start).
func TestGermanChunker_DokumentFuerAerzteFestgestellte_NPS_PP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Das", "ART:DEF:NOM:SIN:NEU", 0),
		atrPos("Dokument", "SUB:NOM:SIN:NEU", 4),
		atrPos("umfasst", "VER:3:SIN:PRÄ:SFT", 13),
		atrPos("das", "ART:DEF:AKK:SIN:NEU", 21),
		atrPos("für", "PRP:AKK", 25),
		atrPos("Ärzte", "SUB:AKK:PLU:MAS", 29),
		atrPos("und", "KON", 35),
		atrPos("Ärztinnen", "SUB:AKK:PLU:FEM", 39),
		atrPos("festgestellte", "PA2:AKK:SIN:NEU:GRU:DEF", 49),
		atrPos("Risikoprofil", "SUB:AKK:SIN:NEU", 63),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "NPS")
	require.Contains(t, tokens[1].GetChunkTags(), "NPS")
	for i, tok := range []string{"für", "Ärzte", "und", "Ärztinnen", "festgestellte", "Risikoprofil"} {
		idx := i + 4
		require.Equal(t, tok, tokens[idx].GetToken())
		require.Contains(t, tokens[idx].GetChunkTags(), "PP", "token %q tags=%v", tok, tokens[idx].GetChunkTags())
	}
}

// Java assertFullChunks tags "Aber/B" but Java source comments "Aber should not be tagged".
// REGEXES1 has no bare KON→B-NP; leave incomplete (no invent B-NP on conjunction).
func TestGermanChunker_AberKenntnisse_NoInventBOnAber(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Aber", "KON", 0),
		atrPos("die", "ART:DEF:NOM:PLU:FEM", 5),
		atrPos("Kenntnisse", "SUB:NOM:PLU:FEM", 9),
		atrPos("der", "ART:DEF:GEN:SIN:FEM", 20),
		atrPos("Sprache", "SUB:GEN:SIN:FEM", 24),
	}
	NewGermanChunker().AddChunkTags(tokens)
	// No invent: Aber is not B-NP without a REGEXES1 path for bare KON
	require.NotContains(t, tokens[0].GetChunkTags(), "B-NP")
	require.NotContains(t, tokens[0].GetChunkTags(), "NPP")
	for i, tok := range []string{"die", "Kenntnisse", "der", "Sprache"} {
		idx := i + 1
		require.Equal(t, tok, tokens[idx].GetToken())
		require.Contains(t, tokens[idx].GetChunkTags(), "NPP", "token %q tags=%v", tok, tokens[idx].GetChunkTags())
	}
}


// Java assertFullChunks: "Die/B Straße/I ist wichtig für/PP die/PP Stadtteile/PP und/PP selbständigen/PP Ortsteile/PP"
func TestGermanChunker_StrasseIstWichtigFuer_NPS_PP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Die", "ART:DEF:NOM:SIN:FEM", 0),
		atrPos("Straße", "SUB:NOM:SIN:FEM", 4),
		atrPos("ist", "VER:3:SIN:PRÄ:NON", 11),
		atrPos("wichtig", "ADJ:PRD:GRU", 15),
		atrPos("für", "PRP:AKK", 23),
		atrPos("die", "ART:DEF:AKK:PLU:MAS", 27),
		atrPos("Stadtteile", "SUB:AKK:PLU:MAS", 31),
		atrPos("und", "KON", 42),
		atrPos("selbständigen", "ADJ:AKK:PLU:MAS", 46),
		atrPos("Ortsteile", "SUB:AKK:PLU:MAS", 60),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "B-NP")
	require.Contains(t, tokens[1].GetChunkTags(), "I-NP")
	// singular NP also gets NPS in REGEXES2
	require.Contains(t, tokens[0].GetChunkTags(), "NPS")
	for i, tok := range []string{"für", "die", "Stadtteile", "und", "selbständigen", "Ortsteile"} {
		idx := i + 4
		require.Equal(t, tok, tokens[idx].GetToken())
		require.Contains(t, tokens[idx].GetChunkTags(), "PP", "token %q tags=%v", tok, tokens[idx].GetChunkTags())
	}
}

// Java assertFullChunks: "Es gab Beschwerden/NPP über/PP laufende/PP Sanierungsmaßnahmen/PP"
func TestGermanChunker_BeschwerdenUeberLaufende_NPP_PP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Es", "PRO:PER:NOM:SIN:NEU", 0),
		atrPos("gab", "VER:3:SIN:PRT:NON", 3),
		atrPos("Beschwerden", "SUB:AKK:PLU:FEM", 7),
		atrPos("über", "PRP:AKK", 19),
		atrPos("laufende", "PA1:AKK:PLU:FEM:GRU:IND", 24),
		atrPos("Sanierungsmaßnahmen", "SUB:AKK:PLU:FEM", 33),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[2].GetChunkTags(), "NPP")
	for i, tok := range []string{"über", "laufende", "Sanierungsmaßnahmen"} {
		idx := i + 3
		require.Equal(t, tok, tokens[idx].GetToken())
		require.Contains(t, tokens[idx].GetChunkTags(), "PP", "token %q tags=%v", tok, tokens[idx].GetChunkTags())
	}
}

// Java assertFullChunks: "Mit/PP über/PP 1000/PP Handschriften/PP ist es die/NPS größte/NPS Sammlung/NPS"
func TestGermanChunker_MitUeberHandschriftenDieGroessteSammlung_PP_NPS(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Mit", "PRP:DAT", 0),
		atrPos("über", "ADV", 4),
		atrPos("1000", "CARD", 9),
		atrPos("Handschriften", "SUB:DAT:PLU:FEM", 14),
		atrPos("ist", "VER:3:SIN:PRÄ:NON", 28),
		atrPos("es", "PRO:PER:NOM:SIN:NEU", 32),
		atrPos("die", "ART:DEF:NOM:SIN:FEM", 35),
		atrPos("größte", "ADJ:NOM:SIN:FEM", 39),
		atrPos("Sammlung", "SUB:NOM:SIN:FEM", 46),
	}
	NewGermanChunker().AddChunkTags(tokens)
	for i, tok := range []string{"Mit", "über", "1000", "Handschriften"} {
		require.Equal(t, tok, tokens[i].GetToken())
		require.Contains(t, tokens[i].GetChunkTags(), "PP", "token %q tags=%v", tok, tokens[i].GetChunkTags())
	}
	for i, tok := range []string{"die", "größte", "Sammlung"} {
		idx := i + 6
		require.Equal(t, tok, tokens[idx].GetToken())
		require.Contains(t, tokens[idx].GetChunkTags(), "NPS", "token %q tags=%v", tok, tokens[idx].GetChunkTags())
	}
}


// Java assertFullChunks: "Es herrscht gute/NPS Laune/NPS in/PP chemischen/PP Komplexverbindungen/PP"
func TestGermanChunker_EsHerrschtGuteLauneInChemischen_NPS_PP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Es", "PRO:PER:NOM:SIN:NEU", 0),
		atrPos("herrscht", "VER:3:SIN:PRÄ:SFT", 3),
		atrPos("gute", "ADJ:NOM:SIN:FEM", 12),
		atrPos("Laune", "SUB:NOM:SIN:FEM", 17),
		atrPos("in", "PRP:DAT", 23),
		atrPos("chemischen", "ADJ:DAT:PLU:FEM", 26),
		atrPos("Komplexverbindungen", "SUB:DAT:PLU:FEM", 37),
	}
	NewGermanChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[2].GetChunkTags(), "NPS")
	require.Contains(t, tokens[3].GetChunkTags(), "NPS")
	for i, tok := range []string{"in", "chemischen", "Komplexverbindungen"} {
		idx := i + 4
		require.Equal(t, tok, tokens[idx].GetToken())
		require.Contains(t, tokens[idx].GetChunkTags(), "PP", "token %q tags=%v", tok, tokens[idx].GetChunkTags())
	}
}

// Java REGEXES2: <regex=eine[rs]?> <seiner|ihrer> <pos=PA1> <pos=SUB> → NPS
// (no assertFullChunks example in GermanChunkerTest; pattern is live in REGEXES2).
func TestGermanChunker_EinerSeinerPA1Sub_NPS(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("einer", "ART:INDEF:NOM:SIN:MAS", 0),
		atrPos("seiner", "PRO:POS:GEN:PLU:MAS", 6),
		atrPos("liebsten", "PA1:GEN:PLU:MAS:GRU:DEF", 13),
		atrPos("Freunde", "SUB:GEN:PLU:MAS", 22),
	}
	NewGermanChunker().AddChunkTags(tokens)
	for i, tok := range []string{"einer", "seiner", "liebsten", "Freunde"} {
		require.Equal(t, tok, tokens[i].GetToken())
		require.Contains(t, tokens[i].GetChunkTags(), "NPS", "token %q tags=%v", tok, tokens[i].GetChunkTags())
	}
}

// Java REGEXES2: <regex=eine[rs]?> <seiner|ihrer> <pos=PA1> <pos=SUB> → NPS (ihrer form)
func TestGermanChunker_EinesIhrerPA1Sub_NPS(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("eines", "ART:INDEF:NOM:SIN:NEU", 0),
		atrPos("ihrer", "PRO:POS:GEN:PLU:NEU", 6),
		atrPos("geliebten", "PA1:GEN:PLU:NEU:GRU:DEF", 12),
		atrPos("Kinder", "SUB:GEN:PLU:NEU", 22),
	}
	NewGermanChunker().AddChunkTags(tokens)
	for i, tok := range []string{"eines", "ihrer", "geliebten", "Kinder"} {
		require.Equal(t, tok, tokens[i].GetToken())
		require.Contains(t, tokens[i].GetChunkTags(), "NPS", "token %q tags=%v", tok, tokens[i].GetChunkTags())
	}
}

// Java assertFullChunks:
// "Und Teil/B der/NPS dort/NPS ausgestellten/NPS Bestände/NPS wurde privat finanziert."
// contains-semantics: Teil must keep B-NP; genitive span is NPS.
func TestGermanChunker_UndTeilDerDortAusgestellten_B_NPS(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("Und", "KON:NEB", 0),
		atrPos("Teil", "SUB:NOM:SIN:MAS", 4),
		atrPos("der", "ART:DEF:GEN:PLU:MAS", 9),
		atrPos("dort", "ADV", 13),
		atrPos("ausgestellten", "PA2:GEN:PLU:MAS:GRU:DEF", 18),
		atrPos("Bestände", "SUB:GEN:PLU:MAS", 32),
	}
	NewGermanChunker().AddChunkTags(tokens)
	// bare KON: no invent B-NP on "Und" (Java comment: Aber should not be tagged either)
	require.NotContains(t, tokens[0].GetChunkTags(), "B-NP")
	require.Contains(t, tokens[1].GetChunkTags(), "B-NP", "Teil tags=%v", tokens[1].GetChunkTags())
	for i, tok := range []string{"der", "dort", "ausgestellten", "Bestände"} {
		idx := i + 2
		require.Equal(t, tok, tokens[idx].GetToken())
		require.Contains(t, tokens[idx].GetChunkTags(), "NPS", "token %q tags=%v", tok, tokens[idx].GetChunkTags())
	}
}

// Java REGEXES2: <er|sie|es> <und> <NP> <NP> → NPP — "sie und sein Sohn ein Paar"
func TestGermanChunker_SieUndSeinSohnEinPaar_NPP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("sie", "PRO:PER:NOM:SIN:FEM", 0),
		atrPos("und", "KON:NEB", 4),
		atrPos("sein", "PRO:POS:NOM:SIN:MAS", 8),
		atrPos("Sohn", "SUB:NOM:SIN:MAS", 13),
		atrPos("ein", "ART:INDEF:NOM:SIN:NEU", 18),
		atrPos("Paar", "SUB:NOM:SIN:NEU", 22),
	}
	NewGermanChunker().AddChunkTags(tokens)
	for i, tok := range []string{"sie", "und", "sein", "Sohn", "ein", "Paar"} {
		require.Equal(t, tok, tokens[i].GetToken())
		require.Contains(t, tokens[i].GetChunkTags(), "NPP", "token %q tags=%v", tok, tokens[i].GetChunkTags())
	}
}

// Java REGEXES2: <pos=PRP> <der> <chunk=NPP>+ → PP
// comment: "Das Bündnis zwischen der Sowjetunion und Kuba"
// Also: EIG und EIG → NPP must not be wiped by genitive und (requires GEN|ADV).
func TestGermanChunker_ZwischenDerSowjetunionUndKuba_PP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrPos("zwischen", "PRP:DAT", 0),
		atrPos("der", "ART:DEF:DAT:SIN:FEM", 9),
		atrPos("Sowjetunion", "EIG:DAT:SIN:FEM", 13),
		atrPos("und", "KON:NEB", 25),
		atrPos("Kuba", "EIG:DAT:SIN:NEU", 29),
	}
	NewGermanChunker().AddChunkTags(tokens)
	for i, tok := range []string{"zwischen", "der", "Sowjetunion", "und", "Kuba"} {
		require.Equal(t, tok, tokens[i].GetToken())
		require.Contains(t, tokens[i].GetChunkTags(), "PP", "token %q tags=%v", tok, tokens[i].GetChunkTags())
	}
	// EIG und EIG → NPP (contains); genitive und must not invent-overwrite without GEN|ADV
	require.Contains(t, tokens[2].GetChunkTags(), "NPP")
	require.Contains(t, tokens[3].GetChunkTags(), "NPP")
	require.Contains(t, tokens[4].GetChunkTags(), "NPP")
}

