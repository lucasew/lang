package nl

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestDutchTagger_Tagger(t *testing.T) {
	wt := tagging.MapWordTagger{
		"huis": {tagging.NewTaggedWord("huis", "ZNW:EKV:HET")},
	}
	tagger := NewDutchTagger(wt)
	got := tagger.Tag([]string{"huis", "xyz"})
	require.Len(t, got, 2)
	require.NotEmpty(t, got[0].GetReadings())
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
	// unknown word still yields a null-POS reading
	require.NotEmpty(t, got[1].GetReadings())
	require.Nil(t, got[1].GetReadings()[0].GetPOSTag())
}

func TestDutchTagger_DictionaryPath(t *testing.T) {
	tagger := NewDutchTagger(nil)
	require.Equal(t, DutchDictPath, tagger.GetDictionaryPath())
}

func TestDutchTagger_ApostropheNormalize(t *testing.T) {
	wt := tagging.MapWordTagger{
		"d'r": {tagging.NewTaggedWord("d'r", "VNW")},
	}
	tagger := NewDutchTagger(wt)
	// curly apostrophe → typewriter before lookup
	got := tagger.Tag([]string{"d’r"})
	require.Len(t, got, 1)
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
	require.Equal(t, "VNW", *got[0].GetReadings()[0].GetPOSTag())
}

func TestDutchTagger_LowercaseFallback(t *testing.T) {
	wt := tagging.MapWordTagger{
		"huis": {tagging.NewTaggedWord("huis", "ZNW:EKV:HET")},
	}
	tagger := NewDutchTagger(wt)
	// sentence-start capital: not mixed → also lower tags
	got := tagger.Tag([]string{"Huis"})
	require.NotEmpty(t, got[0].GetReadings())
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
}

func TestDutchTagger_AllUpperFirstUpper(t *testing.T) {
	wt := tagging.MapWordTagger{
		"Huis": {tagging.NewTaggedWord("huis", "ZNW:EKV:HET")},
	}
	tagger := NewDutchTagger(wt)
	// all-upper only hits first-upper path when surface/lower empty
	got := tagger.Tag([]string{"HUIS"})
	require.NotEmpty(t, got[0].GetReadings())
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
}

func TestDutchTagger_AccentStrip(t *testing.T) {
	// Java: déúr → deur via accent patterns when surface missing
	wt := tagging.MapWordTagger{
		"deur": {tagging.NewTaggedWord("deur", "ZNW:EKV:DE_")},
	}
	tagger := NewDutchTagger(wt)
	got := tagger.Tag([]string{"déúr"})
	require.NotEmpty(t, got[0].GetReadings())
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
	require.Equal(t, "ZNW:EKV:DE_", *got[0].GetReadings()[0].GetPOSTag())
	// accent-normalized hit → ignore spelling
	require.True(t, got[0].IsIgnoredBySpeller())
}

func TestDutchTagger_GetPostags_NoCompound(t *testing.T) {
	wt := tagging.MapWordTagger{
		"huis": {tagging.NewTaggedWord("huis", "ZNW:EKV:HET")},
	}
	tagger := NewDutchTagger(wt)
	// GetPostags must not use compound path
	tagger.GetCompoundParts = func(word string) []string {
		t.Fatal("GetPostags must not call GetCompoundParts")
		return nil
	}
	toks := tagger.GetPostags("huis")
	require.Len(t, toks, 1)
	require.NotNil(t, toks[0].GetPOSTag())
	// unknown: empty (no null invent in getPostags list — Java returns empty from asAnalyzedTokenList)
	require.Empty(t, tagger.GetPostags("onbekendwoordxyz"))
}

func TestDutchTagger_CompoundParts(t *testing.T) {
	wt := tagging.MapWordTagger{
		"puzzel": {tagging.NewTaggedWord("puzzel", "ZNW:EKV:DE_")},
	}
	tagger := NewDutchTagger(wt)
	tagger.GetCompoundParts = func(word string) []string {
		if word == "straatpuzzel" {
			return []string{"straat", "puzzel"}
		}
		return nil
	}
	got := tagger.Tag([]string{"straatpuzzel"})
	require.NotEmpty(t, got[0].GetReadings())
	require.Equal(t, "ZNW:EKV:DE_", *got[0].GetReadings()[0].GetPOSTag())
	// lemma = part1lc + part2 lemma
	require.NotNil(t, got[0].GetReadings()[0].GetLemma())
	require.Equal(t, "straatpuzzel", *got[0].GetReadings()[0].GetLemma())
}

func TestDutchTagger_CompoundAlwaysNeedsHet(t *testing.T) {
	wt := tagging.MapWordTagger{
		"weer": {tagging.NewTaggedWord("weer", "ZNW:EKV:DE_")}, // dict tag overridden
	}
	tagger := NewDutchTagger(wt)
	tagger.GetCompoundParts = func(word string) []string {
		if word == "varkensweer" {
			return []string{"varkens", "weer"}
		}
		return nil
	}
	got := tagger.Tag([]string{"varkensweer"})
	require.NotEmpty(t, got[0].GetReadings())
	require.Equal(t, "ZNW:EKV:HET", *got[0].GetReadings()[0].GetPOSTag())
}

func TestDutchTagger_CompoundGeo(t *testing.T) {
	wt := tagging.MapWordTagger{
		"Turkije": {tagging.NewTaggedWord("Turkije", "ENM:LOC:PTS")},
	}
	tagger := NewDutchTagger(wt)
	tagger.GetCompoundParts = func(word string) []string {
		if word == "Zuidoost-Turkije" {
			return []string{"Zuidoost-", "Turkije"}
		}
		return nil
	}
	got := tagger.Tag([]string{"Zuidoost-Turkije"})
	require.NotEmpty(t, got[0].GetReadings())
	require.Equal(t, "ENM:LOC:PTS", *got[0].GetReadings()[0].GetPOSTag())
	require.Equal(t, "Turkije", *got[0].GetReadings()[0].GetLemma())
}

func TestDutchTagger_NoCompoundWithoutHook(t *testing.T) {
	// without GetCompoundParts, unknown long word stays null POS (fail-closed)
	tagger := NewDutchTagger(tagging.MapWordTagger{})
	got := tagger.Tag([]string{"straatpuzzel"})
	require.NotEmpty(t, got[0].GetReadings())
	require.Nil(t, got[0].GetReadings()[0].GetPOSTag())
}

// Twin of DutchTaggerTest.testDictionary — path/dict availability (full morfologik scan is N/A without dict bytes).
func TestDutchTagger_Dictionary(t *testing.T) {
	tagger := NewDutchTagger(nil)
	// Java TestTools.testDictionary walks the dict; we assert resource path wiring exists.
	require.NotEmpty(t, tagger.GetDictionaryPath())
}
