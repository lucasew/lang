package rules

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// MisspelledFunc reports whether a suggestion token is misspelled (pluggable speller).
type MisspelledFunc func(s string) bool

// TagReplacementsFunc tags replacement strings (pluggable tagger). Optional.
type TagReplacementsFunc func(replacements []string) []*languagetool.AnalyzedTokenReadings

// AbstractSuppressMisspelledSuggestionsFilter ports
// org.languagetool.rules.AbstractSuppressMisspelledSuggestionsFilter with hooks
// instead of Language/SpellingCheckRule/Tagger.
type AbstractSuppressMisspelledSuggestionsFilter struct {
	// IsMisspelled required for filtering; when nil, nothing is treated as misspelled.
	IsMisspelled MisspelledFunc
	// TagReplacements optional; used when SuppressPostag/FilterPostag args present.
	TagReplacements TagReplacementsFunc
}

// AcceptRuleMatch filters suggested replacements; returns nil when all dropped and suppressMatch.
func (f *AbstractSuppressMisspelledSuggestionsFilter) AcceptRuleMatch(
	match *RuleMatch, arguments map[string]string,
) *RuleMatch {
	if match == nil {
		return nil
	}
	replacements := match.GetSuggestedReplacements()
	var newReplacements []string
	suppressPostag := arguments["SuppressPostag"]
	filterPostag := arguments["FilterPostag"]
	var atrs []*languagetool.AnalyzedTokenReadings
	if f.TagReplacements != nil && (suppressPostag != "" || filterPostag != "") {
		atrs = f.TagReplacements(replacements)
	}
	for i, rep := range replacements {
		if f.isMisspelled(rep) {
			continue
		}
		add := true
		if atrs != nil && i < len(atrs) && atrs[i] != nil {
			if suppressPostag != "" && atrs[i].MatchesPosTagRegex(suppressPostag) {
				add = false
			}
			if filterPostag != "" && !atrs[i].MatchesPosTagRegex(filterPostag) {
				add = false
			}
		}
		if add {
			newReplacements = append(newReplacements, rep)
		}
	}
	bSuppressMatch := true
	if sm, ok := arguments["suppressMatch"]; ok && strings.EqualFold(sm, "false") {
		bSuppressMatch = false
	}
	if len(newReplacements) == 0 && bSuppressMatch {
		return nil
	}
	match.SetSuggestedReplacements(newReplacements)
	return match
}

func (f *AbstractSuppressMisspelledSuggestionsFilter) isMisspelled(s string) bool {
	if f.IsMisspelled == nil {
		return false
	}
	// Java tokenizes with language tokenizer; surface: check whole string.
	return f.IsMisspelled(s)
}
