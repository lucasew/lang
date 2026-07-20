package rules

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// MisspelledFunc reports whether a string is misspelled.
// Used as SpellingCheckRule.isMisspelled(token) when IsMisspelledOverride is nil
// (parent Java tokenizes then checks each token).
type MisspelledFunc func(s string) bool

// TagReplacementsFunc tags replacement strings (pluggable tagger). Optional.
type TagReplacementsFunc func(replacements []string) []*languagetool.AnalyzedTokenReadings

// AbstractSuppressMisspelledSuggestionsFilter ports
// org.languagetool.rules.AbstractSuppressMisspelledSuggestionsFilter with hooks
// instead of Language/SpellingCheckRule/Tagger.
type AbstractSuppressMisspelledSuggestionsFilter struct {
	// IsMisspelled is SpellingCheckRule.isMisspelled for one token (parent Java path).
	// When nil and IsMisspelledOverride is nil: nothing is misspelled
	// (Java: null SpellingCheckRule → false).
	IsMisspelled MisspelledFunc
	// IsMisspelledOverride ports a subclass override of isMisspelled(s, language)
	// receiving the full suggestion (e.g. Catalan). When set, it replaces the
	// default tokenize+per-token path entirely (Java polymorphism).
	IsMisspelledOverride MisspelledFunc
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

// isMisspelled ports AbstractSuppressMisspelledSuggestionsFilter.isMisspelled.
// Subclass override (IsMisspelledOverride) receives the full suggestion string —
// do not invent tokenize-then-override-per-token (breaks CA incorrect-verb chunk).
func (f *AbstractSuppressMisspelledSuggestionsFilter) isMisspelled(s string) bool {
	if f != nil && f.IsMisspelledOverride != nil {
		return f.IsMisspelledOverride(s)
	}
	// Java base:
	//   SpellingCheckRule spellerRule = language.getDefaultSpellingRule();
	//   if (spellerRule == null) return false;
	//   for (String token : language.getWordTokenizer().tokenize(s))
	//     if (spellerRule.isMisspelled(token)) return true;
	if f == nil || f.IsMisspelled == nil {
		return false
	}
	for _, tok := range f.tokenizeSuggestion(s) {
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
