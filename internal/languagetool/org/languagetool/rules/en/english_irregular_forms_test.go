package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

func TestFilterEnglishContainsToken(t *testing.T) {
	got := filterEnglishContainsToken([]string{"timezone s", "timezones", "foo ll"})
	require.Equal(t, []string{"timezones"}, got)
}

func TestEnglishIrregularForms_PastTense(t *testing.T) {
	// "goed" → base "go" + VBD → "went" if synthesizer returns and not misspelled
	synth := func(surface, lemma, pos string) []string {
		if lemma == "go" && pos == "VBD" {
			return []string{"went", "goed"}
		}
		return nil
	}
	isMiss := func(w string) bool { return w == "goed" } // only goed misspelled
	f := EnglishIrregularForms("goed", isMiss, synth)
	require.NotNil(t, f)
	require.Equal(t, "go", f.BaseForm)
	require.Equal(t, "verb", f.PosName)
	require.Equal(t, "past tense", f.FormName)
	require.Equal(t, []string{"went"}, f.Forms) // goed removed as self
}

func TestEnglishIrregularForms_NoSynth(t *testing.T) {
	require.Nil(t, EnglishIrregularForms("goed", func(string) bool { return false }, nil))
}

func TestMatch_IrregularAndVariant(t *testing.T) {
	r := NewAbstractEnglishSpellerRule("MORFOLOGIK_RULE_EN_US", "en-US", nil)
	sp := morfologik.NewMorfologikSpeller("/en/hunspell/en_US.dict", 1)
	sp.AddWord("ok")
	sp.AddWord("went")
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled
	r.Synthesize = func(surface, lemma, pos string) []string {
		if lemma == "go" && pos == "VBD" {
			return []string{"went"}
		}
		return nil
	}
	m, err := r.Match(languagetool.AnalyzePlain("goed"))
	require.NoError(t, err)
	require.Len(t, m, 1)
	require.Equal(t, []string{"went"}, m[0].GetSuggestedReplacements())
}

func TestMatch_OtherVariant(t *testing.T) {
	r := NewAbstractEnglishSpellerRule("MORFOLOGIK_RULE_EN_US", "en-US", nil)
	sp := morfologik.NewMorfologikSpeller("/en/hunspell/en_US.dict", 1)
	sp.AddWord("ok")
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled
	r.IsValidInOtherVariantFn = func(word string) *VariantInfo {
		if word == "colour" || word == "Colour" {
			v := NewVariantInfo("British English", "color")
			return &v
		}
		return nil
	}
	m, err := r.Match(languagetool.AnalyzePlain("colour"))
	require.NoError(t, err)
	require.Len(t, m, 1)
	require.Equal(t, []string{"color"}, m[0].GetSuggestedReplacements())
	m, err = r.Match(languagetool.AnalyzePlain("Colour"))
	require.NoError(t, err)
	require.Len(t, m, 1)
	require.Equal(t, []string{"Color"}, m[0].GetSuggestedReplacements())
}

func TestVariantSpeller_IsValidInOtherVariantFnWired(t *testing.T) {
	r := NewMorfologikAmericanSpellerRule()
	require.NotNil(t, r.IsValidInOtherVariantFn)
	// en-US-GB map auto-loaded when resource present
	if len(r.OtherVariant) == 0 {
		t.Skip("en-US-GB not loaded")
	}
	vi := r.IsValidInOtherVariant("colour")
	require.NotNil(t, vi)
	require.Equal(t, "British English", vi.GetVariantName())
	require.Equal(t, "color", vi.GetOtherVariant())
}
