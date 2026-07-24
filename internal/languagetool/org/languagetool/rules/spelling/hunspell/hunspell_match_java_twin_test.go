package hunspell

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/stretchr/testify/require"
)

// Java: Type.UnknownWord on spelling matches.
func TestHunspellMatch_UnknownWordType(t *testing.T) {
	dict := NewMapHunspellDictionary([]string{"hello"})
	r := NewHunspellRule("en", dict)
	ms, err := r.Match(languagetool.AnalyzePlain("helo"))
	require.NoError(t, err)
	require.NotEmpty(t, ms)
	require.Equal(t, rules.RuleMatchTypeUnknownWord, ms[0].GetType())
}

// Java: userConfig.isSuggestionsEnabled() == false → empty suggestions.
func TestHunspellMatch_SuggestionsDisabled(t *testing.T) {
	dict := NewMapHunspellDictionary([]string{"hello"})
	dict.SetSuggestions("helo", []string{"hello"})
	r := NewHunspellRule("en", dict)
	uc := languagetool.NewUserConfig()
	uc.SuggestionsEnabled = false
	r.SetUserConfig(uc)
	ms, err := r.Match(languagetool.AnalyzePlain("helo"))
	require.NoError(t, err)
	require.NotEmpty(t, ms)
	require.Empty(t, ms[0].GetSuggestedReplacements())
}

// Java: maxSpellingSuggestions gate → too_many_errors.
func TestHunspellMatch_MaxSpellingSuggestions(t *testing.T) {
	dict := NewMapHunspellDictionary([]string{"ok"})
	dict.SetSuggestions("aaa", []string{"ok"})
	dict.SetSuggestions("bbb", []string{"ok"})
	dict.SetSuggestions("ccc", []string{"ok"})
	r := NewHunspellRule("en", dict)
	uc := languagetool.NewUserConfig()
	uc.MaxSpellingSuggestions = 1
	r.SetUserConfig(uc)
	ms, err := r.Match(languagetool.AnalyzePlain("aaa bbb ccc"))
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(ms), 2)
	// First match (soFar=0) may have real sugs; later when soFar > max → too_many_errors.
	foundLimit := false
	for _, m := range ms {
		for _, s := range m.GetSuggestedReplacements() {
			if s == tooManyErrorsMsg {
				foundLimit = true
			}
		}
	}
	require.True(t, foundLimit, "expected too_many_errors among %+v", ms)
}

// Java: word.startsWith("-") and rest correct → no match.
func TestHunspellMatch_LeadingDashAcceptedRest(t *testing.T) {
	dict := NewMapHunspellDictionary([]string{"hello"})
	r := NewHunspellRule("en", dict)
	// "-hello" as single token: rest "hello" accepted → skip
	// AnalyzePlain may split; inject ATR.
	at := languagetool.NewAnalyzedToken("-hello", nil, nil)
	atr := languagetool.NewAnalyzedTokenReadingsAt(at, 0)
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{atr})
	ms, err := r.Match(sent)
	require.NoError(t, err)
	require.Empty(t, ms)
}

// Java isFirstItemHighConfidenceSuggestion for DE "HAus" → Haus.
func TestHunspellMatch_DEHighConfidence(t *testing.T) {
	dict := NewMapHunspellDictionary([]string{"Haus"})
	dict.SetSuggestions("HAus", []string{"Haus"})
	r := NewHunspellRule("de", dict)
	ms, err := r.Match(languagetool.AnalyzePlain("HAus"))
	require.NoError(t, err)
	require.NotEmpty(t, ms)
	objs := ms[0].GetSuggestedReplacementObjects()
	require.NotEmpty(t, objs)
	require.NotNil(t, objs[0].GetConfidence())
	require.InDelta(t, float64(rules.SpellingHighConfidence), float64(*objs[0].GetConfidence()), 0.001)
}

// Java wrong-split still emits current-word match after wrong-split match.
func TestHunspellMatch_WrongSplitPlusCurrentWord(t *testing.T) {
	dict := NewMapHunspellDictionary([]string{"thank", "you"})
	r := NewHunspellRule("en", dict)
	ms, err := r.Match(languagetool.AnalyzePlain("thanky ou"))
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(ms), 2, "wrong-split + current word (Java)")
	// first should be wrong-split spanning both
	require.Contains(t, ms[0].GetSuggestedReplacements(), "thank you")
	require.Equal(t, 0, ms[0].GetFromPos())
}

// PT common word "ou" suppresses wrong-split.
func TestHunspellMatch_PTIgnoreWrongSplitOnCommonWord(t *testing.T) {
	dict := NewMapHunspellDictionary([]string{"thank", "you", "ou"})
	// "thanky" misspelled, "ou" is common Portuguese → no wrong-split
	r := NewHunspellRule("pt", dict)
	require.True(t, r.ignoreWrongSplit("thanky", "ou"))
	require.False(t, r.ignoreWrongSplit("thanky", "xyzzq"))
}

// ForeignLanguageChecker on Hunspell Match (preferredLanguages ≥ 2).
func TestHunspellMatch_ForeignLanguage(t *testing.T) {
	dict := NewMapHunspellDictionary([]string{"the"})
	r := NewHunspellRule("en", dict)
	uc := languagetool.NewUserConfig()
	uc.SetPreferredLanguagesList([]string{"en", "de"})
	r.SetUserConfig(uc)
	r.ForeignDetect = func(sentence string, preferred []string, maxResults int) []spelling.DetectedLanguageScore {
		return []spelling.DetectedLanguageScore{{ShortCode: "de", Confidence: 0.88}}
	}
	ms, err := r.Match(languagetool.AnalyzePlain("xxx yyy zzz www vvv"))
	require.NoError(t, err)
	require.NotEmpty(t, ms)
	require.Contains(t, ms[0].GetNewLanguageMatches(), "de")
}

// isMisspelled: "--" never misspelled; prohibited forces misspell.
func TestHunspellIsMisspelled_DashAndProhibit(t *testing.T) {
	dict := NewMapHunspellDictionary([]string{"ok", "badword"})
	r := NewHunspellRule("en", dict)
	require.False(t, r.IsMisspelledWord("--"))
	r.AddProhibitedWords("badword")
	require.True(t, r.IsMisspelledWord("badword"))
}
