package commandline

import (
	"io"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// ApplySuggestionsCheck runs checker and writes corrected text (first suggestion each match).
// Returns match count (like Check hooks).
func ApplySuggestionsCheck(w io.Writer, text string, matches []*rules.RuleMatch) (int, error) {
	tms := make([]tools.TextMatch, 0, len(matches))
	for _, m := range matches {
		if m == nil {
			continue
		}
		tms = append(tms, tools.TextMatch{
			FromPos:               m.FromPos,
			ToPos:                 m.ToPos,
			SuggestedReplacements: m.GetSuggestedReplacements(),
		})
	}
	corrected := tools.CorrectTextFromMatches(text, tms)
	if w != nil {
		_, err := io.WriteString(w, corrected)
		if err != nil {
			return len(matches), err
		}
		if len(corrected) == 0 || corrected[len(corrected)-1] != '\n' {
			_, _ = io.WriteString(w, "\n")
		}
	}
	return len(matches), nil
}
