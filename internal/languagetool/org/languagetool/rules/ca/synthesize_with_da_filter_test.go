package ca

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func ptrDA(s string) *string { return &s }

func atrDA(token, pos, lemma string, start int) *languagetool.AnalyzedTokenReadings {
	return languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken(token, ptrDA(pos), ptrDA(lemma)), start)
}

func TestGenderNumberFromPOS(t *testing.T) {
	require.Equal(t, "MS", GenderNumberFromPOS("NCMS000"))
	require.Equal(t, "FS", GenderNumberFromPOS("NCFS000"))
	require.Equal(t, "MP", GenderNumberFromPOS("NCMP000"))
}

func TestSynthesizeWithDAFilter_Prefixed(t *testing.T) {
	f := NewSynthesizeWithDAFilter()
	require.Equal(t, "l'amic", f.PrefixedSuggestion("amic", "MS", ""))
	require.Equal(t, "de la casa", f.PrefixedSuggestion("casa", "FS", "de"))
}

func TestSynthesizeWithDARegistered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.ca.SynthesizeWithDAFilter"))
}

// Original form only (no synth tags): "amic" → "l'amic"
func TestSynthesizeWithDAFilter_AcceptOriginal(t *testing.T) {
	f := NewSynthesizeWithDAFilter()
	amic := atrDA("amic", "NCMS000", "amic", 0)
	pattern := []*languagetool.AnalyzedTokenReadings{amic}
	sent := languagetool.NewAnalyzedSentence(append([]*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", ptrDA(languagetool.SentenceStartTagName), nil)),
	}, pattern...))
	m := rules.NewRuleMatch(nil, sent, 0, 4, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"lemmaFrom":   "1",
		"lemmaSelect": "N.*",
	}, 0, pattern, nil)
	require.NotNil(t, out)
	sugs := out.GetSuggestedReplacements()
	require.NotEmpty(t, sugs)
	// Match at sentence start → Java uppercases first char
	require.Contains(t, sugs, "L'amic")
}

// With synthesizer expanding to FS form (not at sentence start)
func TestSynthesizeWithDAFilter_AcceptSynthAll(t *testing.T) {
	f := NewSynthesizeWithDAFilter()
	f.GetPossibleTags = func() []string { return []string{"NCMS000", "NCFS000"} }
	f.Synthesize = func(tok *languagetool.AnalyzedToken, postag string) []string {
		switch postag {
		case "NCMS000":
			return []string{"amic"}
		case "NCFS000":
			return []string{"amiga"}
		default:
			return nil
		}
	}
	prev := atrDA("veig", "VMIP1S00", "veure", 0)
	amic := atrDA("amic", "NCMS000", "amic", 5)
	amic.SetWhitespaceBefore(true)
	// patternTokens are only the matched tokens; sentence has previous word
	pattern := []*languagetool.AnalyzedTokenReadings{amic}
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", ptrDA(languagetool.SentenceStartTagName), nil)),
		prev, amic,
	})
	m := rules.NewRuleMatch(nil, sent, 5, 9, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"lemmaFrom":     "1",
		"lemmaSelect":   "N.*",
		"synthAllForms": "true",
	}, 0, pattern, nil)
	require.NotNil(t, out)
	sugs := out.GetSuggestedReplacements()
	require.Contains(t, sugs, "l'amic")
	require.Contains(t, sugs, "l'amiga")
}

func TestSynthesizeWithDAFilter_PrepositionDe(t *testing.T) {
	f := NewSynthesizeWithDAFilter()
	prev := atrDA("parlo", "VMIP1S00", "parlar", 0)
	casa := atrDA("casa", "NCFS000", "casa", 6)
	casa.SetWhitespaceBefore(true)
	pattern := []*languagetool.AnalyzedTokenReadings{casa}
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", ptrDA(languagetool.SentenceStartTagName), nil)),
		prev, casa,
	})
	m := rules.NewRuleMatch(nil, sent, 6, 10, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"lemmaFrom":       "1",
		"lemmaSelect":     "N.*",
		"prepositionFrom": "de",
	}, 0, pattern, nil)
	require.NotNil(t, out)
	require.Contains(t, out.GetSuggestedReplacements(), "de la casa")
}

func TestSynthesizeWithDAFilter_MissingOriginal(t *testing.T) {
	f := NewSynthesizeWithDAFilter()
	// POS doesn't match lemmaSelect
	tok := atrDA("x", "VMIP3S00", "x", 0)
	pattern := []*languagetool.AnalyzedTokenReadings{tok}
	sent := languagetool.NewAnalyzedSentence(pattern)
	m := rules.NewRuleMatch(nil, sent, 0, 1, "msg")
	require.Panics(t, func() {
		_ = f.AcceptRuleMatch(m, map[string]string{
			"lemmaFrom":   "1",
			"lemmaSelect": "N.*",
		}, 0, pattern, nil)
	})
}
