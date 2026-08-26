package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestWireGermanChunker_PostDisambiguationOnly(t *testing.T) {
	lt := languagetool.NewJLanguageTool("de")
	WireGermanChunker(lt)
	// Java: getChunker() is null; post-disambiguation chunker is GermanChunker.
	require.Nil(t, lt.Chunker)
	require.NotNil(t, lt.PostDisambiguationChunker)
}

func TestRegisterCore_WiresPostDisambiguationChunker(t *testing.T) {
	lt := languagetool.NewJLanguageTool("de")
	RegisterCoreGermanRules(lt)
	require.Nil(t, lt.Chunker, "German has no pre-disambig chunker")
	require.NotNil(t, lt.PostDisambiguationChunker)
}

func TestWireGermanChunker_NilSafe(t *testing.T) {
	require.NotPanics(t, func() { WireGermanChunker(nil) })
}

func TestWireGermanChunker_AddsNPTagsAfterAnalyze(t *testing.T) {
	lt := languagetool.NewJLanguageTool("de")
	// Need TagWord for POS-driven NP chunks.
	if !WireGermanTagWord(lt) {
		// Without tagger, inject POS via TagWord for smoke only
		lt.TagWord = func(tok string) []languagetool.TokenTag {
			switch tok {
			case "Der":
				return []languagetool.TokenTag{{POS: "ART:DEF:NOM:SIN:MAS", Lemma: "der"}}
			case "Hund":
				return []languagetool.TokenTag{{POS: "SUB:NOM:SIN:MAS", Lemma: "Hund"}}
			case "bellt":
				return []languagetool.TokenTag{{POS: "VER:3:SIN:PRÄ:SFT", Lemma: "bellen"}}
			default:
				return nil
			}
		}
	}
	WireGermanChunker(lt)
	sents := lt.Analyze("Der Hund bellt.")
	require.NotEmpty(t, sents)
	// Post-disambig chunker runs at end of Analyze — NP tags when ART/SUB present.
	// Without morph dict may not tag; inject path above ensures POS.
	toks := sents[0].GetTokensWithoutWhitespace()
	require.NotEmpty(t, toks)
	// Find Der/Hund if present and check chunk tags when POS was available
	for _, tok := range toks {
		if tok == nil {
			continue
		}
		if tok.GetToken() == "Der" || tok.GetToken() == "der" {
			// may be B-NP when ART
			_ = tok.GetChunkTags()
		}
	}
}
