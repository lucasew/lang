package ca

import (
	"embed"
	"sync"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/replace_anglicism.txt
var anglicismFS embed.FS

var (
	anglicismOnce sync.Once
	anglicismBase *rules.AbstractSimpleReplaceRule2

	anglicismGenderArgs = map[string]string{"lemmaSelect": "[NA].*"}
)

func loadAnglicism() *rules.AbstractSimpleReplaceRule2 {
	anglicismOnce.Do(func() {
		f, err := anglicismFS.Open("data/replace_anglicism.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		base := &rules.AbstractSimpleReplaceRule2{
			ID:                   "CA_SIMPLE_REPLACE_ANGLICISM",
			Description:          "Anglicismes innecessaris: $match",
			ShortMsg:             "Anglicisme innecessari",
			MessageTemplate:      "Anglicisme innecessari. Considereu fer servir una altra paraula.",
			SuggestionsSeparator: " o ",
			CaseSens:             rules.CaseInsensitive,
			LanguageCode:         "ca",
			SubRuleSpecificIDs:   true,
			// Java isTokenException: (NP* && len>1) || immunized || ignoredBySpeller.
			IsTokenException:     anglicismTokenException,
			IsRuleMatchException: anglicismRuleMatchException,
		}
		if err := base.LoadSimpleReplaceRule2Data(f, "/ca/replace_anglicism.txt"); err != nil {
			panic(err)
		}
		anglicismBase = base
	})
	return anglicismBase
}

// anglicismTokenException ports SimpleReplaceAnglicism.isTokenException.
func anglicismTokenException(atr *languagetool.AnalyzedTokenReadings) bool {
	if atr == nil {
		return false
	}
	if atr.IsImmunized() || atr.IsIgnoredBySpeller() {
		return true
	}
	// Java: hasPosTagStartingWith("NP") && getToken().length()>1
	if atr.HasPosTagStartingWith("NP") && utf8.RuneCountInString(atr.GetToken()) > 1 {
		return true
	}
	return false
}

// anglicismRuleMatchException ports SimpleReplaceAnglicism.isRuleMatchException
// (English-span _english_ignore_ tags; without tags fail closed → no exception).
func anglicismRuleMatchException(ruleMatch *rules.RuleMatch) bool {
	if ruleMatch == nil || ruleMatch.Sentence == nil {
		return false
	}
	tokens := ruleMatch.Sentence.GetTokensWithoutWhitespace()
	startIndex := 0
	for startIndex < len(tokens) && tokens[startIndex].GetStartPos() < ruleMatch.GetFromPos() {
		startIndex++
	}
	endIndex := startIndex
	for endIndex < len(tokens) && tokens[endIndex].GetEndPos() < ruleMatch.GetToPos() {
		endIndex++
	}
	if startIndex > 1 && tokens[startIndex].HasPosTag("_english_ignore_") &&
		tokens[startIndex-1].HasPosTag("_english_ignore_") {
		return true
	}
	if endIndex+1 < len(tokens) && tokens[endIndex].HasPosTag("_english_ignore_") &&
		tokens[endIndex+1].HasPosTag("_english_ignore_") {
		return true
	}
	return false
}

// SimpleReplaceAnglicism ports org.languagetool.rules.ca.SimpleReplaceAnglicism.
// ConvertToGenderAndNumberFilter runs when Filter.Tag is set; without Tag, surface
// suggestions only (multi-token matches skip the filter in Java).
type SimpleReplaceAnglicism struct {
	*rules.AbstractSimpleReplaceRule2
	// Filter optional ConvertToGenderAndNumberFilter (Java always constructs one).
	Filter *ConvertToGenderAndNumberFilter
}

func NewSimpleReplaceAnglicism(messages map[string]string) *SimpleReplaceAnglicism {
	base := loadAnglicism()
	r := *base
	r.Messages = messages
	return &SimpleReplaceAnglicism{AbstractSimpleReplaceRule2: &r}
}

// Match ports SimpleReplaceAnglicism.match (super + single-token gender/number filter).
func (r *SimpleReplaceAnglicism) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil {
		return nil
	}
	potential := r.AbstractSimpleReplaceRule2.Match(sentence)
	filter := r.Filter
	if filter == nil || filter.Tag == nil {
		return potential
	}
	var out []*rules.RuleMatch
	for _, m := range potential {
		if m == nil {
			continue
		}
		// multi-word: skip filter (Java isUnderlinedErrorSingleToken)
		if !m.IsUnderlinedErrorSingleToken() {
			out = append(out, m)
			continue
		}
		final := filter.AcceptRuleMatch(m, anglicismGenderArgs, 0, nil, nil)
		if final != nil {
			out = append(out, final)
		}
	}
	return out
}
