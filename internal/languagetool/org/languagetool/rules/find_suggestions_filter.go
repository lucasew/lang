package rules

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// FindSuggestionsFilter ports surface suggestion assembly from
// AbstractFindSuggestionsFilter (speller/tagger hooks are pluggable).
type FindSuggestionsFilter struct {
	// SpellingSuggestions returns candidates for a token.
	SpellingSuggestions func(token string) []string
	// MatchesDesiredPOS reports whether a candidate has the desired POS regex.
	MatchesDesiredPOS func(candidate, desiredPostag string) bool
	// MaxSuggestions caps returned suggestions (Java default 10).
	MaxSuggestions int
}

func NewFindSuggestionsFilter() *FindSuggestionsFilter {
	return &FindSuggestionsFilter{MaxSuggestions: 10}
}

// Collect filters spelling suggestions by desired POS and case of original.
func (f *FindSuggestionsFilter) Collect(token, desiredPostag string, capitalize, allUpper bool) []string {
	if f.SpellingSuggestions == nil {
		return nil
	}
	max := f.MaxSuggestions
	if max <= 0 {
		max = 10
	}
	var out []string
	seen := map[string]struct{}{}
	for _, s := range f.SpellingSuggestions(token) {
		if s == token {
			continue
		}
		if f.MatchesDesiredPOS != nil && desiredPostag != "" && !f.MatchesDesiredPOS(s, desiredPostag) {
			continue
		}
		repl := s
		if allUpper {
			repl = strings.ToUpper(repl)
		} else if capitalize {
			repl = tools.UppercaseFirstChar(repl)
		}
		if _, ok := seen[repl]; ok {
			continue
		}
		seen[repl] = struct{}{}
		out = append(out, repl)
		if len(out) >= max {
			break
		}
	}
	return out
}

// ApplyTemplates expands {suggestion}/{Suggestion}/{SUGGESTION} in templates.
func ApplySuggestionTemplates(templates, suggestions []string) []string {
	var out []string
	seen := map[string]struct{}{}
	add := func(s string) {
		if _, ok := seen[s]; ok {
			return
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	used := false
	for _, s := range templates {
		switch {
		case strings.Contains(s, "{suggestion}"):
			used = true
			for _, s2 := range suggestions {
				add(strings.ReplaceAll(s, "{suggestion}", s2))
			}
		case strings.Contains(s, "{Suggestion}"):
			used = true
			for _, s2 := range suggestions {
				add(strings.ReplaceAll(s, "{Suggestion}", tools.UppercaseFirstChar(s2)))
			}
		case strings.Contains(s, "{SUGGESTION}"):
			used = true
			for _, s2 := range suggestions {
				add(strings.ReplaceAll(s, "{SUGGESTION}", strings.ToUpper(s2)))
			}
		default:
			add(s)
		}
	}
	if !used {
		for _, s := range suggestions {
			add(s)
		}
	}
	return out
}
