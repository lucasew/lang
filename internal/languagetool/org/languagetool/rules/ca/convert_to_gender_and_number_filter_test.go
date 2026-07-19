package ca

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func ptrCGN(s string) *string { return &s }

func atrCGN(token, pos, lemma string, start int) *languagetool.AnalyzedTokenReadings {
	return languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken(token, ptrCGN(pos), ptrCGN(lemma)), start)
}

func sentenceCGN(toks ...*languagetool.AnalyzedTokenReadings) *languagetool.AnalyzedSentence {
	start := languagetool.NewAnalyzedTokenReadings(
		languagetool.NewAnalyzedToken("", ptrCGN(languagetool.SentenceStartTagName), nil))
	return languagetool.NewAnalyzedSentence(append([]*languagetool.AnalyzedTokenReadings{start}, toks...))
}

func TestSplitGenderAndNumber(t *testing.T) {
	s := SplitGenderAndNumber("NCMS000")
	require.NotNil(t, s)
	require.Equal(t, "NC", s.Prefix)
	require.Equal(t, "M", s.Gender)
	require.Equal(t, "S", s.Number)

	s = SplitGenderAndNumber("VMP00SM0")
	require.NotNil(t, s)
	require.True(t, s.Prefix[0] == 'V')
	require.Equal(t, "M", s.Gender)
	require.Equal(t, "S", s.Number)
}

func TestDesiredPostag(t *testing.T) {
	f := NewConvertToGenderAndNumberFilter()
	s := SplitGenderAndNumber("NCMS000")
	got := f.DesiredPostag(s, "F", "P")
	require.Contains(t, got, "F")
	require.Contains(t, got, "P")
}

func TestBoToBonAndIgnore(t *testing.T) {
	require.Equal(t, "bon", BoToBon("bo"))
	require.True(t, ShouldIgnoreForm("mes"))
	require.True(t, IsPostagException("NP00000"))
}

func TestConvertToGenderAndNumberRegistered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.ca.ConvertToGenderAndNumberFilter"))
}

// Simple path: convert NCMS noun to feminine plural via synthesizer
func TestConvertToGenderAndNumberFilter_AcceptSimple(t *testing.T) {
	f := NewConvertToGenderAndNumberFilter()
	f.Synthesize = func(tok *languagetool.AnalyzedToken, postagRE string) []string {
		// NC[FC][PN]000 for amigues
		if tok.GetLemma() != nil && *tok.GetLemma() == "amic" {
			if postagRE == "NC[FC][PN]000" || postagRE == "NC[F][P]000" {
				return []string{"amigues"}
			}
			// pattern is NC[FC][PN]000 with C in gender class
			if postagRE == "NC[FC][PN]000" {
				return []string{"amigues"}
			}
			// actual DesiredPostag: NC[FC][PN]000 for F+P → prefix NC, [F C], [P N], 000
			if postagRE == "NC[FC][PN]000" {
				return []string{"amigues"}
			}
		}
		// match flexible
		if tok != nil && tok.GetLemma() != nil && *tok.GetLemma() == "amic" {
			return []string{"amigues"}
		}
		return nil
	}
	// Override to check actual postagRE
	f.Synthesize = func(tok *languagetool.AnalyzedToken, postagRE string) []string {
		require.Equal(t, "NC[FC][PN]000", postagRE)
		return []string{"amigues"}
	}

	// tokens: [0]SENT [1]dummy [2]amic — posWord not 0 for startPos>1 backwards
	dummy := atrCGN("veig", "VMIP1S00", "veure", 0)
	amic := atrCGN("amic", "NCMS000", "amic", 5)
	amic.SetWhitespaceBefore(true)
	sent := sentenceCGN(dummy, amic)
	m := rules.NewRuleMatch(nil, sent, 5, 9, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"lemmaSelect": "N.*",
		"gender":      "F",
		"number":      "P",
	}, 0, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"amigues"}, out.GetSuggestedReplacements())
}

func TestConvertToGenderAndNumberFilter_NoSynth(t *testing.T) {
	f := NewConvertToGenderAndNumberFilter()
	// without Synthesize, synthesize returns "" → ignoreThisSuggestion → empty → nil
	amic := atrCGN("amic", "NCMS000", "amic", 0)
	sent := sentenceCGN(amic)
	m := rules.NewRuleMatch(nil, sent, 0, 4, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{
		"lemmaSelect": "N.*",
		"gender":      "F",
		"number":      "S",
	}, 0, nil, nil))
}

func TestConvertToGenderAndNumberFilter_KeepOriginalWithDet(t *testing.T) {
	f := NewConvertToGenderAndNumberFilter()
	f.Synthesize = func(tok *languagetool.AnalyzedToken, postagRE string) []string {
		// DA forms for "el" → "la"
		if tok.GetLemma() != nil && *tok.GetLemma() == "el" {
			return []string{"la"}
		}
		if tok.GetLemma() != nil && *tok.GetLemma() == "amic" {
			return []string{"amiga"}
		}
		return nil
	}
	// [0]SENT [1]x [2]el [3]amic
	x := atrCGN("veig", "VMIP1S00", "veure", 0)
	el := atrCGN("el", "DA0MS0", "el", 5)
	el.SetWhitespaceBefore(true)
	amic := atrCGN("amic", "NCMS000", "amic", 8)
	amic.SetWhitespaceBefore(true)
	sent := sentenceCGN(x, el, amic)
	m := rules.NewRuleMatch(nil, sent, 8, 12, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"lemmaSelect": "N.*",
		"gender":      "F",
		"number":      "S",
	}, 0, nil, nil)
	require.NotNil(t, out)
	sugs := out.GetSuggestedReplacements()
	require.NotEmpty(t, sugs)
	// should include determiner rewrite path with "la" prefix via getPrepositionAndDeterminer
	// or synthesized "la"+"amiga"
	joined := sugs[0]
	require.Contains(t, joined, "amiga")
}

func TestPreserveCaseWordByWord(t *testing.T) {
	require.Equal(t, "Foo Bar", preserveCaseWordByWord("foo bar", "Aaa Bbb"))
	// length mismatch falls back to preserveCase whole
	require.NotEmpty(t, preserveCaseWordByWord("foo bar baz", "Aaa"))
}
