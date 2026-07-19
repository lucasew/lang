package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

type langCodeRule struct {
	id, code string
}

func (r langCodeRule) GetID() string           { return r.id }
func (r langCodeRule) GetLanguageCode() string { return r.code }

func TestAdaptSuggestionsFilter_LanguageRegistry(t *testing.T) {
	prev := languagetool.LanguageAdaptSuggestionByCode["ca"]
	t.Cleanup(func() {
		if prev == nil {
			delete(languagetool.LanguageAdaptSuggestionByCode, "ca")
		} else {
			languagetool.LanguageAdaptSuggestionByCode["ca"] = prev
		}
	})
	languagetool.LanguageAdaptSuggestionByCode["ca"] = func(s, _ string) string {
		if s == "a el" {
			return "al"
		}
		return s
	}
	sent := languagetool.AnalyzePlain("a el")
	m := NewRuleMatch(langCodeRule{id: "X", code: "ca"}, sent, 0, 4, "msg")
	m.SetSuggestedReplacement("a el")
	f := NewAdaptSuggestionsFilter(nil) // resolve via registry
	out := f.AcceptRuleMatch(m, nil, 0, nil, nil)
	require.Equal(t, []string{"al"}, out.GetSuggestedReplacements())
}
