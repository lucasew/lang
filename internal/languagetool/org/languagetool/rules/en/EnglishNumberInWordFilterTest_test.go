package en

// Twin of EnglishNumberInWordFilterTest.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestEnglishNumberInWordFilter_FailClosedWithoutDict(t *testing.T) {
	ClearEnglishFilterSpeller()
	f := NewEnglishNumberInWordFilter()
	// Soft invent removed: without speller dict, no candidates
	require.Empty(t, f.Suggestions("H0use"))
	m := rules.NewRuleMatch(rules.NewFakeRule("N"), nil, 0, 5, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{"word": "H0use"}, 0, nil, nil))
}

func TestEnglishNumberInWordFilter_WithInjectedSpeller(t *testing.T) {
	// Inject misspelled gate without full dict binary
	f := NewEnglishNumberInWordFilter()
	f.inner.IsMisspelled = func(w string) bool {
		// known good forms
		return w != "House" && w != "house"
	}
	f.inner.GetSuggestions = nil
	// Force available path by using Suggestions via inner only (Suggestions also checks FilterDictAvailable)
	// Call inner.Suggestions directly for unit gate logic
	got := f.inner.Suggestions("H0use")
	require.Contains(t, got, "House")
}
