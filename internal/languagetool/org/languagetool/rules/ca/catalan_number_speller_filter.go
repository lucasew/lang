package ca

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// CatalanNumberSpellerFilter ports org.languagetool.rules.ca.CatalanNumberSpellerFilter.
// SpellNumber converts a digit string (optionally prefixed with "feminine ") to words
// (Java: CatalanSynthesizer.getSpelledNumber).
type CatalanNumberSpellerFilter struct {
	// SpellNumber is required for suggestions; nil suppresses all matches (fail-closed).
	SpellNumber func(strToSpell string) string
}

func NewCatalanNumberSpellerFilter(spell func(string) string) *CatalanNumberSpellerFilter {
	return &CatalanNumberSpellerFilter{SpellNumber: spell}
}

// AcceptRuleMatch ports CatalanNumberSpellerFilter.acceptRuleMatch.
// Args: number_to_spell (required), gender (required; "feminine" enables feminine spelling).
func (f *CatalanNumberSpellerFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string,
	patternTokenPos int, _ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if f == nil || match == nil {
		return nil
	}
	numberToSpell, ok := arguments["number_to_spell"]
	if !ok {
		panic("Missing key 'number_to_spell'")
	}
	gender, ok := arguments["gender"]
	if !ok {
		panic("Missing key 'gender'")
	}
	// Sentence-start capitalization (Java: patternTokenPos <= 1 || prev has SENT_START).
	sentenceStart := patternTokenPos <= 1
	if !sentenceStart && match.Sentence != nil {
		tokens := match.Sentence.GetTokensWithoutWhitespace()
		// Java: tokens[patternTokenPos - 1]
		if patternTokenPos-1 >= 0 && patternTokenPos-1 < len(tokens) && tokens[patternTokenPos-1] != nil {
			if tokens[patternTokenPos-1].HasPartialPosTag("SENT_START") {
				sentenceStart = true
			}
		}
	}
	spelled := f.Suggest(numberToSpell, gender, sentenceStart)
	if spelled == "" {
		return nil
	}
	out := rules.NewRuleMatch(match.GetRule(), match.Sentence, match.GetFromPos(), match.GetToPos(), match.GetMessage())
	out.ShortMessage = match.ShortMessage
	out.SetSuggestedReplacement(spelled)
	return out
}

// Suggest returns the spelled number, or "" to suppress.
// sentenceStart capitalises the first letter; gender "feminine" prefixes the request.
// wordCountMax is 4 (Java: split length < 4).
func (f *CatalanNumberSpellerFilter) Suggest(numberToSpell, gender string, sentenceStart bool) string {
	if f.SpellNumber == nil {
		return ""
	}
	str := strings.ReplaceAll(numberToSpell, ".", "")
	if gender == "feminine" {
		str = "feminine " + str
	}
	spelled := f.SpellNumber(str)
	if sentenceStart {
		spelled = tools.UppercaseFirstChar(spelled)
	}
	if spelled == "" {
		return ""
	}
	// Java: spelledNumber.replace("-i-", " ").replace("-", " ").split(" ").length < 4
	// split(" ") keeps empty mid-fields (double spaces); Fields would collapse them.
	norm := strings.ReplaceAll(spelled, "-i-", " ")
	norm = strings.ReplaceAll(norm, "-", " ")
	parts := strings.Split(norm, " ")
	if len(parts) >= 4 {
		return ""
	}
	return spelled
}
