package rules

import (
	"strings"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// AbstractSimpleReplaceRule ports org.languagetool.rules.AbstractSimpleReplaceRule
// for dictionary-based token replacements (checkLemmas path optional).
// Java ctor: MISC category; checkLemmas defaults true (callers must set true unless
// Java setCheckLemmas(false) — Go zero value is false, so do not leave unset invent).
type AbstractSimpleReplaceRule struct {
	Messages          map[string]string
	WrongWords        map[string][]string
	CaseSensitive     bool
	// CheckLemmas ports isCheckLemmas(); Java default true.
	CheckLemmas       bool
	IgnoreTaggedWords bool
	ID                string
	Description       string
	ShortMsg          string
	// LanguageCode ports Language for StringTools.toId (useSubRuleSpecificIds).
	LanguageCode string
	// SubRuleSpecificIDs ports useSubRuleSpecificIds() → SpecificIdRule per match.
	SubRuleSpecificIDs bool
	// Premium ports Rule.isPremium for SpecificIdRule.
	Premium bool
	// Tags ports Rule.getTags for SpecificIdRule.
	Tags []Tag
	// Category ports Rule.category (Java MISC).
	Category *Category
	// URL ports Rule.url (Java setUrl).
	URL string
	// IssueType ports getLocQualityIssueType.
	IssueType ITSIssueType
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []IncorrectExample
	correctExamples   []CorrectExample
	// MessageFn custom message; if nil uses default.
	// Called with replacements before title-case adjust (Java getMessage then casing).
	MessageFn func(tokenStr string, replacements []string) string
	// TokenException optional skip (ports isTokenException).
	TokenException func(token *languagetool.AnalyzedTokenReadings) bool
	// IsTagged optional override for ignoreTaggedWords (Java protected isTagged).
	// nil → tokenReadings.IsTagged().
	IsTagged func(token *languagetool.AnalyzedTokenReadings) bool
}

// UseSubRuleSpecificIDs ports AbstractSimpleReplaceRule.useSubRuleSpecificIds.
func (r *AbstractSimpleReplaceRule) UseSubRuleSpecificIDs() {
	if r != nil {
		r.SubRuleSpecificIDs = true
	}
}

// GetTags ports Rule.getTags.
func (r *AbstractSimpleReplaceRule) GetTags() []Tag {
	if r == nil {
		return nil
	}
	return r.Tags
}

// InitSimpleReplaceMeta applies Java AbstractSimpleReplaceRule constructor metadata.
func InitSimpleReplaceMeta(r *AbstractSimpleReplaceRule, messages map[string]string) {
	if r == nil {
		return
	}
	r.Messages = messages
	if r.Category == nil {
		r.Category = CatMisc.GetCategory(messages)
	}
}

func (r *AbstractSimpleReplaceRule) GetID() string {
	if r.ID != "" {
		return r.ID
	}
	return "SIMPLE_REPLACE"
}

func (r *AbstractSimpleReplaceRule) GetCategory() *Category {
	if r == nil {
		return nil
	}
	return r.Category
}

func (r *AbstractSimpleReplaceRule) GetDescription() string {
	if r != nil && r.Description != "" {
		return r.Description
	}
	return ""
}

// GetURL ports Rule.getUrl.
func (r *AbstractSimpleReplaceRule) GetURL() string {
	if r == nil {
		return ""
	}
	return r.URL
}

// SetURL ports Rule.setUrl.
func (r *AbstractSimpleReplaceRule) SetURL(u string) {
	if r != nil {
		r.URL = u
	}
}

// GetLocQualityIssueType ports Rule.getLocQualityIssueType.
func (r *AbstractSimpleReplaceRule) GetLocQualityIssueType() ITSIssueType {
	if r == nil || r.IssueType == "" {
		return ITSUncategorized
	}
	return r.IssueType
}

// AddExamplePair ports Rule.addExamplePair.
func (r *AbstractSimpleReplaceRule) AddExamplePair(incorrect IncorrectExample, correct CorrectExample) {
	if r == nil {
		return
	}
	appendExamplePair(&r.incorrectExamples, &r.correctExamples, incorrect, correct)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *AbstractSimpleReplaceRule) GetIncorrectExamples() []IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *AbstractSimpleReplaceRule) GetCorrectExamples() []CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

func (r *AbstractSimpleReplaceRule) cleanup(word string) string {
	if r.CaseSensitive {
		return word
	}
	return strings.ToLower(word)
}

// Match ports AbstractSimpleReplaceRule.match (sentence-level).
func (r *AbstractSimpleReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*RuleMatch {
	var ruleMatches []*RuleMatch
	for _, tokenReadings := range sentence.GetTokensWithoutWhitespace() {
		if tokenReadings.IsSentenceStart() || tokenReadings.IsImmunized() {
			continue
		}
		if r.TokenException != nil && r.TokenException(tokenReadings) {
			continue
		}
		if tokenReadings.IsIgnoredBySpeller() {
			continue
		}
		if r.IgnoreTaggedWords {
			tagged := tokenReadings.IsTagged()
			if r.IsTagged != nil {
				tagged = r.IsTagged(tokenReadings)
			}
			if tagged {
				continue
			}
		}
		matches := r.FindMatches(tokenReadings, sentence)
		ruleMatches = append(ruleMatches, matches...)
	}
	return ruleMatches
}

// FindMatches ports AbstractSimpleReplaceRule.findMatches (exported for language overrides).
func (r *AbstractSimpleReplaceRule) FindMatches(tokenReadings *languagetool.AnalyzedTokenReadings, sentence *languagetool.AnalyzedSentence) []*RuleMatch {
	return r.findMatches(tokenReadings, sentence)
}

func (r *AbstractSimpleReplaceRule) findMatches(tokenReadings *languagetool.AnalyzedTokenReadings, sentence *languagetool.AnalyzedSentence) []*RuleMatch {
	originalTokenStr := tokenReadings.GetToken()
	tokenString := r.cleanup(originalTokenStr)
	isAllUppercase := tools.IsAllUppercase(originalTokenStr)

	possibleReplacements := r.WrongWords[originalTokenStr]
	if possibleReplacements == nil {
		possibleReplacements = r.WrongWords[tokenString]
	}

	// Lemma path skipped when CheckLemmas is false (ContractionSpellingRule).
	if possibleReplacements == nil && r.CheckLemmas {
		// Without synthesizer: look up lemmas directly as wrong words.
		var found []string
		seen := map[string]bool{}
		for _, at := range tokenReadings.GetReadings() {
			if at.GetLemma() == nil {
				continue
			}
			lemma := r.cleanup(*at.GetLemma())
			if reps, ok := r.WrongWords[lemma]; ok {
				for _, rep := range reps {
					if !seen[rep] {
						seen[rep] = true
						found = append(found, rep)
					}
				}
			}
		}
		if len(found) > 0 {
			possibleReplacements = found
		}
	}

	if len(possibleReplacements) == 0 {
		return nil
	}

	var replacements []string
	if isAllUppercase {
		for _, s := range possibleReplacements {
			replacements = append(replacements, strings.ToUpper(s))
		}
	} else {
		replacements = append([]string(nil), possibleReplacements...)
	}
	// remove identity
	filtered := replacements[:0]
	for _, rep := range replacements {
		if rep != originalTokenStr {
			filtered = append(filtered, rep)
		}
	}
	replacements = filtered
	if len(replacements) == 0 {
		return nil
	}

	// createRuleMatch: message first (replacements not yet title-cased), then casing, then setSuggestedReplacements.
	return []*RuleMatch{r.createRuleMatch(tokenReadings, replacements, sentence, originalTokenStr)}
}

// createRuleMatch ports AbstractSimpleReplaceRule.createRuleMatch.
func (r *AbstractSimpleReplaceRule) createRuleMatch(
	tokenReadings *languagetool.AnalyzedTokenReadings,
	replacements []string,
	sentence *languagetool.AnalyzedSentence,
	originalTokenStr string,
) *RuleMatch {
	tokenString := tokenReadings.GetToken()
	pos := tokenReadings.GetStartPos()
	end := pos + utf16TokenLen(tokenString)

	msg := "Possible spelling mistake found."
	if r.MessageFn != nil {
		// Java getMessage(tokenString, replacements) before uppercaseFirstChar on suggestions.
		msg = r.MessageFn(tokenString, replacements)
	}

	short := r.ShortMsg
	if short == "" {
		short = "Spelling mistake"
	}

	var rm *RuleMatch
	if r.SubRuleSpecificIDs {
		// Java: StringTools.toId(getId() + "_" + originalTokenStr, language)
		// getDescription().replace("$match", originalTokenStr) — all occurrences
		id := tools.ToId(r.GetID()+"_"+originalTokenStr, r.LanguageCode)
		desc := strings.ReplaceAll(r.GetDescription(), "$match", originalTokenStr)
		idRule := NewSpecificIdRule(id, desc, r.Premium, r.GetCategory(), r.GetLocQualityIssueType(), r.GetTags())
		rm = NewRuleMatch(idRule, sentence, pos, end, msg)
	} else {
		rm = NewRuleMatch(r, sentence, pos, end, msg)
	}
	rm.ShortMessage = short

	// Java: after RuleMatch construction, title-case suggestions when !caseSensitive && capitalized token.
	if !r.CaseSensitive && tools.StartsWithUppercase(tokenString) {
		for i, rep := range replacements {
			replacements[i] = tools.UppercaseFirstChar(rep)
		}
	}
	rm.SetSuggestedReplacements(replacements)
	return rm
}

func utf16TokenLen(s string) int {
	n := 0
	for _, r := range s {
		n += len(utf16.Encode([]rune{r}))
	}
	return n
}
