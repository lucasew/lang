package rules

import "strings"

// SuggestionFilter ports org.languagetool.rules.SuggestionFilter.
// Triggers reports whether applying template with "{}"→repl produces a rule error.
type SuggestionFilter struct {
	// Triggers is true when the filled template still matches the rule (drop suggestion).
	Triggers func(filledTemplate string) bool
}

func NewSuggestionFilter(triggers func(string) bool) *SuggestionFilter {
	return &SuggestionFilter{Triggers: triggers}
}

// Filter keeps replacements that do not re-trigger the rule under template.
func (f *SuggestionFilter) Filter(replacements []string, template string) []string {
	trig := f.Triggers
	if trig == nil {
		return replacements
	}
	var out []string
	for _, repl := range replacements {
		filled := strings.Replace(template, "{}", repl, 1)
		if !trig(filled) {
			out = append(out, repl)
		}
	}
	return out
}
