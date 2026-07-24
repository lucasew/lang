package chunking

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Behavior matrix for RussianChunker. Java has no RussianChunkerTest.java.
// POS fixtures use Russian Morphy inventory shapes from LT ru resources:
//   NN:Anim:Masc:Sin:Nom, ADJ:Posit:Masc:Nom, VB:Real:TRANS:IMPFV:Sin:P3,
//   PT:Past:TRANS:2PFV:STR:Masc:Nom, DPT:Past:TRANS:2PFV, PNN:Fem:R:P3, …
// Expected tags are Java-visible outcomes (getChunkTag + REGEXES1 then REGEXES2).

func ruTok(token, pos string, start int) *languagetool.AnalyzedTokenReadings {
	p := pos
	return languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken(token, &p, nil), start)
}

func requireChunkTags(t *testing.T, tok *languagetool.AnalyzedTokenReadings, want ...string) {
	t.Helper()
	require.Equal(t, want, tok.GetChunkTags())
}

func requireHasChunk(t *testing.T, tok *languagetool.AnalyzedTokenReadings, tag string) {
	t.Helper()
	require.Contains(t, tok.GetChunkTags(), tag)
}

// --- REGEXES1: name sequences ------------------------------------------------

// <posre='NN:(Name|Fam|Patr):.*'> <posre='NN:(Name|Fam|Patr):.*'>+ → NP overwrite
func TestRussianChunker_NameFamPatr_NP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		ruTok("Иванов", "NN:Fam:Masc:Sin:Nom", 0),
		ruTok("Иван", "NN:Name:Masc:Sin:Nom", 7),
		ruTok("Иванович", "NN:Patr:Masc:Sin:Nom", 12),
	}
	NewRussianChunker().AddChunkTags(tokens)
	requireChunkTags(t, tokens[0], "B-NP")
	requireChunkTags(t, tokens[1], "I-NP")
	requireChunkTags(t, tokens[2], "I-NP")
}

// <posre='NN:Fam:.*'> <regexCS=[А-ЯЁ]> <.> <regexCS=[А-ЯЁ]> <.> → NP
func TestRussianChunker_FamInitials_NP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		ruTok("Иванов", "NN:Fam:Masc:Sin:Nom", 0),
		ruTok("И", "ABR", 7),
		ruTok(".", "UNKNOWN", 8),
		ruTok("И", "ABR", 9),
		ruTok(".", "UNKNOWN", 10),
	}
	NewRussianChunker().AddChunkTags(tokens)
	requireChunkTags(t, tokens[0], "B-NP")
	requireChunkTags(t, tokens[1], "I-NP")
	requireChunkTags(t, tokens[2], "I-NP")
	requireChunkTags(t, tokens[3], "I-NP")
	requireChunkTags(t, tokens[4], "I-NP")
}

// <regexCS=[А-ЯЁ]> <.> <regexCS=[А-ЯЁ]> <.> <posre='NN:Fam:.*'> → NP
func TestRussianChunker_InitialsFam_NP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		ruTok("И", "ABR", 0),
		ruTok(".", "UNKNOWN", 1),
		ruTok("И", "ABR", 2),
		ruTok(".", "UNKNOWN", 3),
		ruTok("Иванов", "NN:Fam:Masc:Sin:Nom", 5),
	}
	NewRussianChunker().AddChunkTags(tokens)
	requireChunkTags(t, tokens[0], "B-NP")
	requireChunkTags(t, tokens[1], "I-NP")
	requireChunkTags(t, tokens[2], "I-NP")
	requireChunkTags(t, tokens[3], "I-NP")
	requireChunkTags(t, tokens[4], "I-NP")
}

// --- REGEXES1: VP / SBAR -----------------------------------------------------

// <posre='VB:.*:.*' & !posre='NN:.*'>* → B-VP / I-VP
func TestRussianChunker_VP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		ruTok("спит", "VB:Real:INTR:IMPFV:Sin:P3", 0),
	}
	NewRussianChunker().AddChunkTags(tokens)
	requireChunkTags(t, tokens[0], "B-VP")
}

func TestRussianChunker_VP_MultiVerb(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		ruTok("хочет", "VB:Real:TRANS:IMPFV:Sin:P3", 0),
		ruTok("спать", "VB:INF:INTR:IMPFV", 6),
	}
	NewRussianChunker().AddChunkTags(tokens)
	requireChunkTags(t, tokens[0], "B-VP")
	requireChunkTags(t, tokens[1], "I-VP")
}

// Pure NN does not match VB pattern → O (no invent).
func TestRussianChunker_BareNN_IsO(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		ruTok("кот", "NN:Anim:Masc:Sin:Nom", 0),
	}
	NewRussianChunker().AddChunkTags(tokens)
	requireChunkTags(t, tokens[0], "O")
}

// <если> / <поэтому> → SBAR
func TestRussianChunker_Esli_SBAR(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		ruTok("если", "CONJ", 0),
	}
	NewRussianChunker().AddChunkTags(tokens)
	requireChunkTags(t, tokens[0], "SBAR")
}

func TestRussianChunker_Poetomu_SBAR(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		ruTok("поэтому", "ADV", 0),
	}
	NewRussianChunker().AddChunkTags(tokens)
	requireChunkTags(t, tokens[0], "SBAR")
}

// --- REGEXES1: ADJ + NN → NP -------------------------------------------------

// ADJ:Posit + NN:Anim/Inanim ending not R|D|T|P → B-NP I-NP
func TestRussianChunker_AdjNoun_NP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		ruTok("большой", "ADJ:Posit:Masc:Nom", 0),
		ruTok("кот", "NN:Anim:Masc:Sin:Nom", 8),
	}
	NewRussianChunker().AddChunkTags(tokens)
	requireChunkTags(t, tokens[0], "B-NP")
	requireChunkTags(t, tokens[1], "I-NP")
}

func TestRussianChunker_AdjNoun_Inanim_NP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		ruTok("красный", "ADJ:Posit:Masc:Nom", 0),
		ruTok("дом", "NN:Inanim:Masc:Sin:Nom", 9),
	}
	NewRussianChunker().AddChunkTags(tokens)
	requireChunkTags(t, tokens[0], "B-NP")
	requireChunkTags(t, tokens[1], "I-NP")
}

// ADJ + NN excluded from 2-token NP when case is R|D|T|P (Morphy case is final field).
func TestRussianChunker_AdjNoun_GenitiveExcludedFromNP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		ruTok("большого", "ADJ:Posit:Masc:R", 0),
		ruTok("кота", "NN:Anim:Masc:Sin:R", 9),
	}
	NewRussianChunker().AddChunkTags(tokens)
	// !posre='NN:(Anim|Inanim):.*:(R|D|T|P)' excludes Sin:R → no NP.
	requireChunkTags(t, tokens[0], "O")
	requireChunkTags(t, tokens[1], "O")
}

// ADJ:Posit + NN (not R/D/T/P) + NN → NP (3-token)
func TestRussianChunker_AdjNounNoun_NP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		ruTok("большая", "ADJ:Posit:Fem:Nom", 0),
		ruTok("группа", "NN:Inanim:Fem:Sin:Nom", 8),
		ruTok("студентов", "NN:Anim:Masc:PL:R", 15),
	}
	NewRussianChunker().AddChunkTags(tokens)
	requireChunkTags(t, tokens[0], "B-NP")
	requireChunkTags(t, tokens[1], "I-NP")
	requireChunkTags(t, tokens[2], "I-NP")
}

// ADJ:Posit + NN(not Nom|V) + NN(Nom|V) → ADJP
// Middle ends D so 2-token NP exclusion also blocks; ADJP pattern wins.
func TestRussianChunker_AdjNounNoun_ADJP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		ruTok("большая", "ADJ:Posit:Fem:Nom", 0),
		ruTok("группе", "NN:Inanim:Fem:Sin:D", 8),
		ruTok("студент", "NN:Anim:Masc:Sin:Nom", 15),
	}
	NewRussianChunker().AddChunkTags(tokens)
	requireChunkTags(t, tokens[0], "B-ADJP")
	requireChunkTags(t, tokens[1], "I-ADJP")
	requireChunkTags(t, tokens[2], "I-ADJP")
}

// --- REGEXES1: DPT -----------------------------------------------------------

// <posre='DPT:.*:.*' & !pos='PREP'> → B-DPT
func TestRussianChunker_DPT_Alone(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		ruTok("читая", "DPT:Real:TRANS:IMPFV", 0),
	}
	NewRussianChunker().AddChunkTags(tokens)
	requireChunkTags(t, tokens[0], "B-DPT")
}

// DPT + NN ending R/D/T/P → B-DPT I-DPT (overwrite)
func TestRussianChunker_DPT_NounCase(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		ruTok("читая", "DPT:Real:TRANS:IMPFV", 0),
		ruTok("книги", "NN:Inanim:Fem:Sin:R", 7),
	}
	NewRussianChunker().AddChunkTags(tokens)
	requireChunkTags(t, tokens[0], "B-DPT")
	requireChunkTags(t, tokens[1], "I-DPT")
}

// DPT + PREP + NN ending R/D/T/P
func TestRussianChunker_DPT_Prep_Noun(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		ruTok("читая", "DPT:Real:TRANS:IMPFV", 0),
		ruTok("в", "PREP", 7),
		ruTok("книге", "NN:Inanim:Fem:Sin:P", 9),
	}
	NewRussianChunker().AddChunkTags(tokens)
	requireChunkTags(t, tokens[0], "B-DPT")
	requireChunkTags(t, tokens[1], "I-DPT")
	requireChunkTags(t, tokens[2], "I-DPT")
}

// --- REGEXES1: PT (participle) → ADJP ----------------------------------------

func TestRussianChunker_PT_Alone_ADJP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		ruTok("написанный", "PT:Past:TRANS:2PFV:STR:Masc:Nom", 0),
	}
	NewRussianChunker().AddChunkTags(tokens)
	requireChunkTags(t, tokens[0], "B-ADJP")
}

func TestRussianChunker_PT_ADV_ADJP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		ruTok("написанный", "PT:Past:TRANS:2PFV:STR:Masc:Nom", 0),
		ruTok("быстро", "ADV", 11),
	}
	NewRussianChunker().AddChunkTags(tokens)
	requireChunkTags(t, tokens[0], "B-ADJP")
	requireChunkTags(t, tokens[1], "I-ADJP")
}

func TestRussianChunker_PT_NounCase_ADJP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		ruTok("написанный", "PT:Past:TRANS:2PFV:STR:Masc:Nom", 0),
		ruTok("автором", "NN:Anim:Masc:Sin:T", 11),
	}
	NewRussianChunker().AddChunkTags(tokens)
	requireChunkTags(t, tokens[0], "B-ADJP")
	requireChunkTags(t, tokens[1], "I-ADJP")
}

func TestRussianChunker_PT_Prep_Noun_ADJP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		ruTok("написанный", "PT:Past:TRANS:2PFV:STR:Masc:Nom", 0),
		ruTok("в", "PREP", 11),
		ruTok("книге", "NN:Inanim:Fem:Sin:P", 13),
	}
	NewRussianChunker().AddChunkTags(tokens)
	requireChunkTags(t, tokens[0], "B-ADJP")
	requireChunkTags(t, tokens[1], "I-ADJP")
	requireChunkTags(t, tokens[2], "I-ADJP")
}

func TestRussianChunker_PT_Prep_Adj_Noun_ADJP(t *testing.T) {
	// ADJ:.*:.*:(R|D|T|P|V) — ADJ:Posit:Fem:P ends with P
	tokens := []*languagetool.AnalyzedTokenReadings{
		ruTok("написанный", "PT:Past:TRANS:2PFV:STR:Masc:Nom", 0),
		ruTok("в", "PREP", 11),
		ruTok("большой", "ADJ:Posit:Fem:P", 13),
		ruTok("книге", "NN:Inanim:Fem:Sin:P", 21),
	}
	NewRussianChunker().AddChunkTags(tokens)
	requireChunkTags(t, tokens[0], "B-ADJP")
	requireChunkTags(t, tokens[1], "I-ADJP")
	requireChunkTags(t, tokens[2], "I-ADJP")
	requireChunkTags(t, tokens[3], "I-ADJP")
}

// PT + NN(not Nom|V) + NN(Nom|V) → ADJP
func TestRussianChunker_PT_NounNoun_ADJP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		ruTok("написанный", "PT:Past:TRANS:2PFV:STR:Masc:Nom", 0),
		ruTok("группе", "NN:Inanim:Fem:Sin:D", 11),
		ruTok("студент", "NN:Anim:Masc:Sin:Nom", 18),
	}
	NewRussianChunker().AddChunkTags(tokens)
	requireChunkTags(t, tokens[0], "B-ADJP")
	requireChunkTags(t, tokens[1], "I-ADJP")
	requireChunkTags(t, tokens[2], "I-ADJP")
}

// PT + PNN(not Nom) + NN(Nom|V) → ADJP
// PNN:Fem:R:P3 does not match PNN:.*:Nom:.*
func TestRussianChunker_PT_PNN_Noun_ADJP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		ruTok("написанный", "PT:Past:TRANS:2PFV:STR:Masc:Nom", 0),
		ruTok("им", "PNN:Masc:T:P3", 11),
		ruTok("роман", "NN:Inanim:Masc:Sin:Nom", 14),
	}
	NewRussianChunker().AddChunkTags(tokens)
	requireChunkTags(t, tokens[0], "B-ADJP")
	requireChunkTags(t, tokens[1], "I-ADJP")
	requireChunkTags(t, tokens[2], "I-ADJP")
}

// PT + ADJ → ADJP (overwrite=false); PT alone also fires first
func TestRussianChunker_PT_ADJ_ADJP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		ruTok("написанный", "PT:Past:TRANS:2PFV:STR:Masc:Nom", 0),
		ruTok("хороший", "ADJ:Posit:Masc:Nom", 11),
	}
	NewRussianChunker().AddChunkTags(tokens)
	requireHasChunk(t, tokens[0], "B-ADJP")
	requireHasChunk(t, tokens[1], "I-ADJP")
}

// <тов> → B-NP (single-token NP)
func TestRussianChunker_Tov_NP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		ruTok("тов", "ABR", 0),
	}
	NewRussianChunker().AddChunkTags(tokens)
	requireChunkTags(t, tokens[0], "B-NP")
}

// --- REGEXES2 ----------------------------------------------------------------

// <posre=NN:Name:.*> <и> <posre=NN:Name:.*> → B-NP-plural …
func TestRussianChunker_NamesAnd_NPP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		ruTok("Маша", "NN:Name:Fem:Sin:Nom", 0),
		ruTok("и", "CONJ", 5),
		ruTok("Миша", "NN:Name:Masc:Sin:Nom", 7),
	}
	NewRussianChunker().AddChunkTags(tokens)
	// REGEXES1 does not fuse across "и"; REGEXES2 NPP assigns plural BIO.
	requireChunkTags(t, tokens[0], "B-NP-plural")
	requireChunkTags(t, tokens[1], "I-NP-plural")
	requireChunkTags(t, tokens[2], "I-NP-plural")
}

func TestRussianChunker_NamesOr_NPP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		ruTok("Маша", "NN:Name:Fem:Sin:Nom", 0),
		ruTok("или", "CONJ", 5),
		ruTok("Миша", "NN:Name:Masc:Sin:Nom", 9),
	}
	NewRussianChunker().AddChunkTags(tokens)
	requireChunkTags(t, tokens[0], "B-NP-plural")
	requireChunkTags(t, tokens[1], "I-NP-plural")
	requireChunkTags(t, tokens[2], "I-NP-plural")
}

// <не> <VB>* → VP (REGEXES2); REGEXES1 already tags VB as B-VP
func TestRussianChunker_Ne_VB_VP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		ruTok("не", "PART", 0),
		ruTok("спит", "VB:Real:INTR:IMPFV:Sin:P3", 3),
	}
	NewRussianChunker().AddChunkTags(tokens)
	// REGEXES1: спит → B-VP; REGEXES2: не+спит → B-VP on не, I-VP on спит
	requireChunkTags(t, tokens[0], "B-VP")
	require.Contains(t, tokens[1].GetChunkTags(), "B-VP")
	require.Contains(t, tokens[1].GetChunkTags(), "I-VP")
}

// --- Control flow / anti-invent ----------------------------------------------

// Soft invent POS→BIO removed: bare non-Morphy tags stay O.
func TestRussianChunker_NoInventSoftPOS(t *testing.T) {
	nn := "NN:nom:m"
	v := "V:ipf"
	tokens := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("кот", &nn, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("спит", &v, nil), 4),
	}
	NewRussianChunker().AddChunkTags(tokens)
	requireChunkTags(t, tokens[0], "O")
	requireChunkTags(t, tokens[1], "O")
	require.NotContains(t, tokens[0].GetChunkTags(), "B-NP")
	require.NotContains(t, tokens[1].GetChunkTags(), "B-VP")
}

// GetBasicChunks runs REGEXES1 only and does not mutate readings.
func TestRussianChunker_GetBasicChunks_NoMutate(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		ruTok("большой", "ADJ:Posit:Masc:Nom", 0),
		ruTok("кот", "NN:Anim:Masc:Sin:Nom", 8),
	}
	basic := NewRussianChunker().GetBasicChunks(tokens)
	require.Len(t, basic, 2)
	require.Equal(t, "B-NP", basic[0].ChunkTags[0].String())
	require.Equal(t, "I-NP", basic[1].ChunkTags[0].String())
	require.Empty(t, tokens[0].GetChunkTags())
	require.Empty(t, tokens[1].GetChunkTags())
}

// REGEXES2 not applied in GetBasicChunks — names+и stay O at basic stage.
func TestRussianChunker_GetBasicChunks_NoREGEXES2(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		ruTok("Маша", "NN:Name:Fem:Sin:Nom", 0),
		ruTok("и", "CONJ", 5),
		ruTok("Миша", "NN:Name:Masc:Sin:Nom", 7),
	}
	basic := NewRussianChunker().GetBasicChunks(tokens)
	require.Len(t, basic, 3)
	require.Equal(t, "O", basic[0].ChunkTags[0].String())
	require.Equal(t, "O", basic[1].ChunkTags[0].String())
	require.Equal(t, "O", basic[2].ChunkTags[0].String())
}

// Tokens with MayMissingYO chunk tag are skipped (Java getBasicChunks filter).
func TestRussianChunker_SkipMayMissingYO(t *testing.T) {
	tok := ruTok("ёлка", "NN:Inanim:Fem:Sin:Nom", 0)
	tok.SetChunkTags([]string{"MayMissingYO"})
	tokens := []*languagetool.AnalyzedTokenReadings{
		tok,
		ruTok("спит", "VB:Real:INTR:IMPFV:Sin:P3", 5),
	}
	basic := NewRussianChunker().GetBasicChunks(tokens)
	require.Len(t, basic, 1)
	require.Equal(t, "спит", basic[0].Token)
	require.Equal(t, "B-VP", basic[0].ChunkTags[0].String())
}

// Lone Name stays O (no invent bare-name BIO); REGEXES1 needs 2+ Name/Fam/Patr.
func TestRussianChunker_LoneName_IsO(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		ruTok("Маша", "NN:Name:Fem:Sin:Nom", 0),
	}
	NewRussianChunker().AddChunkTags(tokens)
	requireChunkTags(t, tokens[0], "O")
}

// Debug hooks exist (Java setDebug/isDebug).
func TestRussianChunker_DebugHooks(t *testing.T) {
	defer SetRussianChunkerDebug(false)
	require.False(t, IsRussianChunkerDebug())
	SetRussianChunkerDebug(true)
	require.True(t, IsRussianChunkerDebug())
	SetRussianChunkerDebug(false)
	require.False(t, IsRussianChunkerDebug())
}

// SYNTAX_EXPANSION ports (used if patterns reference <NP>/<VP>/…).
func TestRussianChunker_SyntaxExpansion(t *testing.T) {
	require.Equal(t, "<chunk=B-NP> <chunk=I-NP>*", ExpandRussianChunkSyntax("<NP>"))
	require.Equal(t, "<chunk=B-VP> <chunk=I-VP>*", ExpandRussianChunkSyntax("<VP>"))
	require.Equal(t, "<chunk=B-ADJP> <chunk=I-ADJP>*", ExpandRussianChunkSyntax("<ADJP>"))
	require.Equal(t, "<chunk=B-DPT> <chunk=I-DPT>*", ExpandRussianChunkSyntax("<DPT>"))
}

// REGEXES1/2 list sizes and overwrite flags (production parity snapshot).
func TestRussianChunker_RegexTableParity(t *testing.T) {
	require.Len(t, russianRegexes1, 21)
	require.Len(t, russianRegexes2, 3)
	require.True(t, russianRegexes1[0].overwrite)
	require.False(t, russianRegexes1[3].overwrite)  // VP
	require.False(t, russianRegexes1[4].overwrite)  // если
	require.False(t, russianRegexes1[20].overwrite) // тов
	require.True(t, russianRegexes2[0].overwrite)
	require.True(t, russianRegexes2[1].overwrite)
	require.False(t, russianRegexes2[2].overwrite)
}

// Case-sensitive regexCS must not match lowercase initials for Fam+initials pattern.
func TestRussianChunker_Initials_CaseSensitive(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		ruTok("Иванов", "NN:Fam:Masc:Sin:Nom", 0),
		ruTok("и", "ABR", 7), // lowercase — regexCS=[А-ЯЁ] fails
		ruTok(".", "UNKNOWN", 8),
		ruTok("и", "ABR", 9),
		ruTok(".", "UNKNOWN", 10),
	}
	NewRussianChunker().AddChunkTags(tokens)
	require.NotContains(t, tokens[0].GetChunkTags(), "B-NP")
	requireChunkTags(t, tokens[0], "O")
}
