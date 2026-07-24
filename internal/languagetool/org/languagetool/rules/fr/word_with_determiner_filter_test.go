package fr

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func ptrWD(s string) *string { return &s }

func atrWD(token, pos, lemma string) *languagetool.AnalyzedTokenReadings {
	return languagetool.NewAnalyzedTokenReadings(
		languagetool.NewAnalyzedToken(token, ptrWD(pos), ptrWD(lemma)))
}

func TestWordWithDeterminerFilter_Helpers(t *testing.T) {
	f := NewWordWithDeterminerFilter()
	require.True(t, f.IsExceptionDeterminer("nouvels"))
	require.False(t, f.IsExceptionDeterminer("les"))
	require.True(t, f.MatchesDetPOS("D m s"))
	require.True(t, f.MatchesWordPOS("N m s"))
	require.Equal(t, "[NZ] ", f.NounAdjPrefix(true, false))
	require.Equal(t, "J ", f.NounAdjPrefix(false, true))
	require.Equal(t, "[ZNJ] ", f.NounAdjPrefix(true, true))
}

func TestWordWithDeterminerRegistered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.fr.WordWithDeterminerFilter"))
}

// le maison (wrong) → la maison etc via synth
func TestWordWithDeterminerAccept(t *testing.T) {
	f := NewWordWithDeterminerFilter()
	f.Synthesize = func(tok *languagetool.AnalyzedToken, postagRE string) []string {
		lem := ""
		if tok.GetLemma() != nil {
			lem = *tok.GetLemma()
		}
		// determiner lemma le/la
		if lem == "le" || tok.GetToken() == "le" || tok.GetToken() == "la" {
			// MS / FS / MP / FP
			switch {
			case containsWD(postagRE, "([me]) (s|sp)"):
				return []string{"le"}
			case containsWD(postagRE, "([fe]) (s|sp)"):
				return []string{"la"}
			case containsWD(postagRE, "([me]) (p|sp)"):
				return []string{"les"}
			case containsWD(postagRE, "([fe]) (p|sp)"):
				return []string{"les"}
			}
		}
		if lem == "maison" {
			switch {
			case containsWD(postagRE, "([me]) (s|sp)"):
				return []string{"maison"} // shouldn't really
			case containsWD(postagRE, "([fe]) (s|sp)"):
				return []string{"maison"}
			case containsWD(postagRE, "([me]) (p|sp)"):
				return []string{"maisons"}
			case containsWD(postagRE, "([fe]) (p|sp)"):
				return []string{"maisons"}
			}
		}
		return nil
	}
	det := atrWD("le", "D m s", "le")
	noun := atrWD("maison", "N f s", "maison")
	pattern := []*languagetool.AnalyzedTokenReadings{det, noun}
	m := rules.NewRuleMatch(nil, nil, 0, 10, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{
		"determinerFrom": "1",
		"wordFrom":       "2",
	}, 0, pattern, nil)
	require.NotNil(t, out)
	sugs := out.GetSuggestedReplacements()
	require.NotEmpty(t, sugs)
	require.Contains(t, sugs, "la maison")
	// exception list not applied to les
	require.Contains(t, sugs, "les maisons")
}

func containsWD(s, sub string) bool {
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

func TestWordWithDeterminerNoSynth_KeepsExisting(t *testing.T) {
	f := NewWordWithDeterminerFilter()
	det := atrWD("le", "D m s", "le")
	noun := atrWD("maison", "N f s", "maison")
	m := rules.NewRuleMatch(nil, nil, 0, 10, "msg")
	m.SetSuggestedReplacements([]string{"existing"})
	out := f.AcceptRuleMatch(m, map[string]string{
		"determinerFrom": "1",
		"wordFrom":       "2",
	}, 0, []*languagetool.AnalyzedTokenReadings{det, noun}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"existing"}, out.GetSuggestedReplacements())
}

func TestWordWithDeterminerValidateFilters(t *testing.T) {
	f := NewWordWithDeterminerFilter()
	f.Synthesize = func(tok *languagetool.AnalyzedToken, postagRE string) []string {
		if containsWD(postagRE, "([fe]) (s|sp)") {
			if tok.GetToken() == "le" || (tok.GetLemma() != nil && *tok.GetLemma() == "le") {
				return []string{"la"}
			}
			return []string{"maison"}
		}
		return []string{"x"}
	}
	f.ValidateSuggestion = func(s string) bool {
		return s != "x x"
	}
	det := atrWD("le", "D m s", "le")
	noun := atrWD("maison", "N f s", "maison")
	out := f.AcceptRuleMatch(rules.NewRuleMatch(nil, nil, 0, 5, "m"), map[string]string{
		"determinerFrom": "1", "wordFrom": "2",
	}, 0, []*languagetool.AnalyzedTokenReadings{det, noun}, nil)
	require.NotNil(t, out)
	for _, s := range out.GetSuggestedReplacements() {
		require.NotEqual(t, "x x", s)
	}
}

func TestWordWithDeterminerMissingArgs(t *testing.T) {
	f := NewWordWithDeterminerFilter()
	require.Panics(t, func() {
		_ = f.AcceptRuleMatch(rules.NewRuleMatch(nil, nil, 0, 1, "m"), map[string]string{}, 0, nil, nil)
	})
}
