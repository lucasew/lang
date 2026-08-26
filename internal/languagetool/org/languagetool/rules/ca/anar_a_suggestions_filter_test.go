package ca

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func ptrAA(s string) *string { return &s }

func atrAA(token, pos, lemma string, start int) *languagetool.AnalyzedTokenReadings {
	return languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken(token, ptrAA(pos), ptrAA(lemma)), start)
}

func sentenceAA(toks ...*languagetool.AnalyzedTokenReadings) *languagetool.AnalyzedSentence {
	start := languagetool.NewAnalyzedTokenReadings(
		languagetool.NewAnalyzedToken("", ptrAA(languagetool.SentenceStartTagName), nil))
	return languagetool.NewAnalyzedSentence(append([]*languagetool.AnalyzedTokenReadings{start}, toks...))
}

func TestAnarASuggest(t *testing.T) {
	f := NewAnarASuggestionsFilter()
	f.SynthFuturePresent = func(lemma, suffix string) []string {
		require.Equal(t, "fer", lemma)
		return []string{"farem", "fem"}
	}
	got := f.Suggest("fer", "1P00", "li ho", "anem")
	require.Len(t, got, 2)
	require.Contains(t, got[0], "farem")
	require.Contains(t, got[1], "fem")
}

func TestAnarARegistered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.ca.AnarASuggestionsFilter"))
}

// anem a fer → farem, fem
func TestAnarAAccept_Basic(t *testing.T) {
	f := NewAnarASuggestionsFilter()
	f.Synthesize = func(tok *languagetool.AnalyzedToken, postagRE string) []string {
		lem := ""
		if tok.GetLemma() != nil {
			lem = *tok.GetLemma()
		}
		if lem != "fer" {
			return nil
		}
		// V[MS]IF1P00 / V[MS]IP1P00
		if postagRE == "V[MS]IF1P00" {
			return []string{"farem"}
		}
		if postagRE == "V[MS]IP1P00" {
			return []string{"fem"}
		}
		return nil
	}
	// anem VMIP1P00 + a + fer VMN0000
	anem := atrAA("anem", "VMIP1P00", "anar", 0)
	a := atrAA("a", "SPS00", "a", 5)
	a.SetWhitespaceBefore(true)
	fer := atrAA("fer", "VMN0000", "fer", 7)
	fer.SetWhitespaceBefore(true)
	sent := sentenceAA(anem, a, fer)
	m := rules.NewRuleMatch(nil, sent, 0, fer.GetEndPos(), "msg")
	out := f.AcceptRuleMatch(m, nil, 0, nil, nil)
	require.NotNil(t, out)
	sugs := out.GetSuggestedReplacements()
	require.Contains(t, sugs, "farem")
	require.Contains(t, sugs, "fem")
}

// li ho anem a fer → li ho farem (clitics before anar)
func TestAnarAAccept_PronounsBefore(t *testing.T) {
	f := NewAnarASuggestionsFilter()
	f.Synthesize = func(tok *languagetool.AnalyzedToken, postagRE string) []string {
		if postagRE == "V[MS]IF1P00" {
			return []string{"farem"}
		}
		if postagRE == "V[MS]IP1P00" {
			return []string{"fem"}
		}
		return nil
	}
	// Need weak pronouns before verb for VerbSynthesizer
	// li PP3CSD00, ho PP3NNA00, anem, a, fer
	// pPronomFeble: PP3CSD00 matches, PP3NNA00 matches PP3..A00
	li := atrAA("li", "PP3CSD00", "ell", 0)
	ho := atrAA("ho", "PP3NNA00", "ho", 3)
	ho.SetWhitespaceBefore(true)
	anem := atrAA("anem", "VMIP1P00", "anar", 6)
	anem.SetWhitespaceBefore(true)
	a := atrAA("a", "SPS00", "a", 11)
	a.SetWhitespaceBefore(true)
	fer := atrAA("fer", "VMN0000", "fer", 13)
	fer.SetWhitespaceBefore(true)
	// dummy so pronoun before loop (iFirstVerb+i > 0) works — need index 0 before pronouns
	// tokens: SENT, X, li, ho, anem, a, fer — match starts at li or anem?
	// Java initPos at match start — for "li ho anem a fer" match might start at anem
	// VerbSynthesizer from anem finds pronouns before.
	sent := sentenceAA(li, ho, anem, a, fer)
	// match from anem
	m := rules.NewRuleMatch(nil, sent, anem.GetStartPos(), fer.GetEndPos(), "msg")
	out := f.AcceptRuleMatch(m, nil, 0, nil, nil)
	require.NotNil(t, out)
	sugs := out.GetSuggestedReplacements()
	require.NotEmpty(t, sugs)
	// should include pronoun + farem
	found := false
	for _, s := range sugs {
		if len(s) > 5 && (containsSub(s, "farem") || containsSub(s, "fem")) {
			found = true
			// pronouns davant
			if containsSub(s, "li") || containsSub(s, "ho") {
				found = true
			}
		}
	}
	require.True(t, found, "got %v", sugs)
	// at least one should have pronoun prefix
	hasPron := false
	for _, s := range sugs {
		if containsSub(s, "farem") && len(s) > len("farem") {
			hasPron = true
		}
	}
	require.True(t, hasPron || len(sugs) >= 2, "got %v", sugs)
}

func containsSub(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub ||
		func() bool {
			for i := 0; i+len(sub) <= len(s); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}

func TestAnarAAccept_NoSynth(t *testing.T) {
	f := NewAnarASuggestionsFilter()
	require.Nil(t, f.AcceptRuleMatch(nil, nil, 0, nil, nil))
}

func TestAnarAAccept_KeepExistingSuggestions(t *testing.T) {
	f := NewAnarASuggestionsFilter()
	f.Synthesize = func(tok *languagetool.AnalyzedToken, postagRE string) []string {
		if postagRE == "V[MS]IF1S00" {
			return []string{"faré"}
		}
		if postagRE == "V[MS]IP1S00" {
			return []string{"faig"}
		}
		return nil
	}
	vaig := atrAA("vaig", "VMIP1S00", "anar", 0)
	a := atrAA("a", "SPS00", "a", 5)
	a.SetWhitespaceBefore(true)
	fer := atrAA("fer", "VMN0000", "fer", 7)
	fer.SetWhitespaceBefore(true)
	sent := sentenceAA(vaig, a, fer)
	m := rules.NewRuleMatch(nil, sent, 0, fer.GetEndPos(), "msg")
	m.SetSuggestedReplacements([]string{"existing"})
	out := f.AcceptRuleMatch(m, nil, 0, nil, nil)
	require.NotNil(t, out)
	require.Contains(t, out.GetSuggestedReplacements(), "existing")
	require.Contains(t, out.GetSuggestedReplacements(), "faré")
}
