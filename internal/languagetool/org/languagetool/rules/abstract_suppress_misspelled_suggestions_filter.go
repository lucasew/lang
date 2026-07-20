package rules

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// MisspelledFunc reports whether a suggestion token is misspelled (pluggable speller).
// Java SpellingCheckRule.isMisspelled receives each WordTokenizer token separately.
type MisspelledFunc func(s string) bool

// TagReplacementsFunc tags replacement strings (pluggable tagger). Optional.
type TagReplacementsFunc func(replacements []string) []*languagetool.AnalyzedTokenReadings

// AbstractSuppressMisspelledSuggestionsFilter ports
// org.languagetool.rules.AbstractSuppressMisspelledSuggestionsFilter with hooks
// instead of Language/SpellingCheckRule/Tagger.
type AbstractSuppressMisspelledSuggestionsFilter struct {
	// IsMisspelled required for filtering; when nil, nothing is treated as misspelled
	// (Java: null SpellingCheckRule → isMisspelled false).
	IsMisspelled MisspelledFunc
	// Tokenize ports language.getWordTokenizer().tokenize; nil → WordTokenizer.
	Tokenize func(s string) []string
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
	// Java getRequired("suppressMatch") — no soft default.
	suppressMatch, ok := arguments["suppressMatch"]
	if !ok {
		panic("Missing key 'suppressMatch'")
	}
	bSuppressMatch := true
	if strings.EqualFold(suppressMatch, "false") {
		bSuppressMatch = false
	}
	if len(newReplacements) == 0 && bSuppressMatch {
		return nil
	}
	match.SetSuggestedReplacements(newReplacements)
	return match
}

// isMisspelled ports AbstractSuppressMisspelledSuggestionsFilter.isMisspelled:
//
//	tokens = language.getWordTokenizer().tokenize(s)
//	for token : tokens { if spellerRule.isMisspelled(token) return true }
//	return false
//
// (null speller → false is handled by nil IsMisspelled.)
func (f *AbstractSuppressMisspelledSuggestionsFilter) isMisspelled(s string) bool {
	if f.IsMisspelled == nil {
		return false
	}
	tokens := f.tokenizeSuggestion(s)
	if len(tokens) == 0 {
		// Empty tokenize: Java would not enter the loop → not misspelled.
		return false
	}
	for _, tok := range tokens {
		if f.IsMisspelled(tok) {
			return true
		}
	}
	return false
}

func (f *AbstractSuppressMisspelledSuggestionsFilter) tokenizeSuggestion(s string) []string {
	if f != nil && f.Tokenize != nil {
		return f.Tokenize(s)
	}
	return tokenizers.NewWordTokenizer().Tokenize(s)
}
