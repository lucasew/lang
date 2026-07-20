package chunking

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func ruTok(token, pos string, start int) *languagetool.AnalyzedTokenReadings {
	p := pos
	return languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken(token, &p, nil), start)
}

// Java REGEXES1: <posre='VB:.*:.*' & !posre='NN:.*'>* → B-VP
func TestRussianChunker_VP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		ruTok("спит", "VB:INDIC:IMPERF:PRES:SG:3", 0),
	}
	NewRussianChunker().AddChunkTags(tokens)
	require.Equal(t, []string{"B-VP"}, tokens[0].GetChunkTags())
}

// Java REGEXES1: ADJ:Posit + NN:Anim/Inanim (not R/D/T/P) → B-NP I-NP
func TestRussianChunker_AdjNoun_NP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		ruTok("большой", "ADJ:Posit:Masc:Nom:Sin", 0),
		ruTok("кот", "NN:Anim:Masc:Nom:Sin", 8),
	}
	NewRussianChunker().AddChunkTags(tokens)
	require.Equal(t, []string{"B-NP"}, tokens[0].GetChunkTags())
	require.Equal(t, []string{"I-NP"}, tokens[1].GetChunkTags())
}

// Java REGEXES2: <posre=NN:Name:.*> <и> <posre=NN:Name:.*> → B-NP-plural …
func TestRussianChunker_NamesAnd_NPP(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		ruTok("Маша", "NN:Name:Fem:Nom:Sin", 0),
		ruTok("и", "CONJ", 5),
		ruTok("Миша", "NN:Name:Masc:Nom:Sin", 7),
	}
	// REGEXES1 may also tag names; REGEXES2 NPP overwrite uses B-NP-plural
	NewRussianChunker().AddChunkTags(tokens)
	require.Contains(t, tokens[0].GetChunkTags(), "B-NP-plural")
	require.Contains(t, tokens[2].GetChunkTags(), "I-NP-plural")
}

// Java: <если> → SBAR
func TestRussianChunker_Esli_SBAR(t *testing.T) {
	tokens := []*languagetool.AnalyzedTokenReadings{
		ruTok("если", "CONJ", 0),
	}
	NewRussianChunker().AddChunkTags(tokens)
	require.Equal(t, []string{"SBAR"}, tokens[0].GetChunkTags())
}

// Soft invent POS→BIO removed: bare "NN:nom" / "V:ipf" do not match Java REGEXES1.
// Java leaves singleton O when no pattern hits.
func TestRussianChunker_NoInventSoftPOS(t *testing.T) {
	nn := "NN:nom:m"
	v := "V:ipf"
	tokens := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("кот", &nn, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("спит", &v, nil), 4),
	}
	NewRussianChunker().AddChunkTags(tokens)
	require.Equal(t, []string{"O"}, tokens[0].GetChunkTags())
	require.Equal(t, []string{"O"}, tokens[1].GetChunkTags())
	require.NotContains(t, tokens[0].GetChunkTags(), "B-NP")
	require.NotContains(t, tokens[1].GetChunkTags(), "B-VP")
}
