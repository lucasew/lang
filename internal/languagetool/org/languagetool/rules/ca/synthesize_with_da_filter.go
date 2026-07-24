package ca

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// SynthesizeWithDAFilter ports
// org.languagetool.rules.ca.SynthesizeWithDAFilter (1:1 AcceptRuleMatch).
//
// GetPossibleTags / Synthesize port CatalanSynthesizer; when nil, only the
// original lemmaSelect reading is used (still produces det+form suggestions).
type SynthesizeWithDAFilter struct {
	// GetPossibleTags ports CatalanSynthesizer.getPossibleTags.
	GetPossibleTags func() []string
	// Synthesize ports synth.synthesize(token, tag) without POS-regex.
	Synthesize func(tok *languagetool.AnalyzedToken, postag string) []string
}

func NewSynthesizeWithDAFilter() *SynthesizeWithDAFilter {
	return &SynthesizeWithDAFilter{}
}

// GenderNumberList for DA filter.
var daGenderNumberList = []string{"MS", "FS", "MP", "FP"}

// GenderNumber patterns — Java Pattern.compile strings (Matcher.matches = full).
var genderNumberPatterns = map[string]*regexp.Regexp{
	"MS": regexp.MustCompile(`(N|A.).[MC][SN].*|V.P.*SM.`),
	"FS": regexp.MustCompile(`(N|A.).[FC][SN].*|V.P.*SF.`),
	"MP": regexp.MustCompile(`(N|A.).[MC][PN].*|V.P.*PM.`),
	"FP": regexp.MustCompile(`(N|A.).[FC][PN].*|V.P.*PF.`),
}

// GenderNumberFromPOS returns MS/FS/MP/FP for a POS tag, or "".
func GenderNumberFromPOS(pos string) string {
	for _, gn := range daGenderNumberList {
		if daFullMatch(genderNumberPatterns[gn], pos) {
			return gn
		}
	}
	return ""
}

func daFullMatch(re *regexp.Regexp, s string) bool {
	if re == nil {
		return false
	}
	loc := re.FindStringIndex(s)
	return loc != nil && loc[0] == 0 && loc[1] == len(s)
}

// PrefixedSuggestion builds determiner/preposition + form for a gender/number.
func (f *SynthesizeWithDAFilter) PrefixedSuggestion(form, genderNumber, preposition string) string {
	det := GetPrepositionAndDeterminer(form, genderNumber, preposition)
	return det + form
}

// FilterForms keeps forms whose POS matches desired gender/number when set.
func (f *SynthesizeWithDAFilter) FilterForms(forms []struct{ Form, POS string }, wantGN string) []string {
	var out []string
	for _, fr := range forms {
		if wantGN != "" {
			if !daFullMatch(genderNumberPatterns[wantGN], fr.POS) {
				continue
			}
		}
		out = append(out, fr.Form)
	}
	return out
}

// PreferGenderNumber moves forms matching secondGenderNumber earlier (unit helper).
func PreferGenderNumber(forms []struct{ Form, POS string }, secondGenderNumber string) []struct{ Form, POS string } {
	if secondGenderNumber == "" || len(forms) < 2 {
		return forms
	}
	var preferred, rest []struct{ Form, POS string }
	swap := ""
	if len(secondGenderNumber) >= 2 {
		swap = string(secondGenderNumber[1]) + string(secondGenderNumber[0])
	}
	for i, fr := range forms {
		if i == 0 {
			preferred = append(preferred, fr)
			continue
		}
		if strings.Contains(fr.POS, secondGenderNumber) || (swap != "" && strings.Contains(fr.POS, swap)) {
			preferred = append(preferred, fr)
		} else {
			rest = append(rest, fr)
		}
	}
	return append(preferred, rest...)
}

// AcceptRuleMatch ports SynthesizeWithDAFilter.acceptRuleMatch.
func (f *SynthesizeWithDAFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, patternTokenPos int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	_ = patternTokenPos
	_ = tokenPositions
	if f == nil || match == nil || len(patternTokens) == 0 {
		return nil
	}
	if arguments == nil {
		return nil
	}

	lemmaFromStr := patterns.GetRequired("lemmaFrom", arguments)
	lemmaSelect := patterns.GetRequired("lemmaSelect", arguments)
	synthAllForms := strings.EqualFold(patterns.GetOptionalDefault("synthAllForms", arguments, "false"), "true")
	prepositionFromStr := patterns.GetOptionalDefault("prepositionFrom", arguments, "")

	lemmaFrom := daGetPosition(lemmaFromStr, patternTokens, match)
	preposition := ""
	if prepositionFromStr != "" && isNumeric(prepositionFromStr) {
		prepositionFrom := daGetPosition(prepositionFromStr, patternTokens, match)
		tok := patternTokens[prepositionFrom].GetToken()
		if tok != "" {
			// first letter lowercased
			r := []rune(tok)
			preposition = strings.ToLower(string(r[0]))
		}
	} else if prepositionFromStr != "" {
		r := []rune(prepositionFromStr)
		preposition = strings.ToLower(string(r[0]))
	}

	originalWord := patternTokens[lemmaFrom].GetToken()
	p, err := regexp.Compile(lemmaSelect)
	if err != nil {
		panic(err)
	}
	isSentenceStart := daIsMatchAtSentenceStart(match.Sentence.GetTokensWithoutWhitespace(), match)

	// original word form in the first place
	originalAT := readingWithTagRegex(patternTokens[lemmaFrom], lemmaSelect)
	if originalAT == nil {
		// Java throws RuntimeException
		panic("Cannot find analyzed token readings with postag " + lemmaSelect + " in sentence" + match.Sentence.GetText())
	}
	potentialSuggestions := []*languagetool.AnalyzedToken{originalAT}

	// second-best suggestion gender/number from previous determiner
	secondGenderNumber := ""
	if lemmaFrom-1 > 0 {
		reading := readingWithTagRegex(patternTokens[lemmaFrom-1], `D.*`)
		if reading != nil && reading.GetPOSTag() != nil {
			tag := *reading.GetPOSTag()
			if len(tag) >= 5 {
				secondGenderNumber = tag[3:5]
			}
		}
	}
	if f.GetPossibleTags != nil && f.Synthesize != nil {
		for _, tag := range f.GetPossibleTags() {
			if !daFullMatch(p, tag) {
				continue
			}
			synthForms := f.Synthesize(originalAT, tag)
			for _, synthForm := range synthForms {
				if !synthAllForms && !strings.EqualFold(synthForm, originalWord) {
					continue
				}
				tagCopy := tag
				at := languagetool.NewAnalyzedToken(synthForm, &tagCopy, originalAT.GetLemma())
				if !daContainsToken(potentialSuggestions, at) {
					if secondGenderNumber != "" && (strings.Contains(tag, secondGenderNumber) ||
						(len(secondGenderNumber) >= 2 && strings.Contains(tag,
							string(secondGenderNumber[1])+string(secondGenderNumber[0])))) {
						// insert at index 1
						potentialSuggestions = append(potentialSuggestions[:1], append([]*languagetool.AnalyzedToken{at}, potentialSuggestions[1:]...)...)
					} else {
						potentialSuggestions = append(potentialSuggestions, at)
					}
				}
			}
		}
	}

	var suggestions []string
	for _, potentialSuggestion := range potentialSuggestions {
		newForm := potentialSuggestion.GetToken()
		posTag := ""
		if potentialSuggestion.GetPOSTag() != nil {
			posTag = *potentialSuggestion.GetPOSTag()
		}
		for _, genderNumber := range daGenderNumberList {
			if daFullMatch(genderNumberPatterns[genderNumber], posTag) {
				suggestion := GetPrepositionAndDeterminer(newForm, genderNumber, preposition) +
					tools.PreserveCase(newForm, originalWord)
				if isSentenceStart {
					suggestion = tools.UppercaseFirstChar(suggestion)
				}
				if !containsString(suggestions, suggestion) {
					suggestions = append(suggestions, suggestion)
				}
			}
		}
	}
	// Java match.addSuggestedReplacements
	existing := match.GetSuggestedReplacements()
	match.SetSuggestedReplacements(append(existing, suggestions...))
	return match
}

func daContainsToken(list []*languagetool.AnalyzedToken, at *languagetool.AnalyzedToken) bool {
	for _, x := range list {
		if x != nil && x.Equals(at) {
			return true
		}
	}
	return false
}

func containsString(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// daGetPosition ports RuleFilter.getPosition (1-based marker/index → 0-based).
func daGetPosition(fromStr string, patternTokens []*languagetool.AnalyzedTokenReadings, match *rules.RuleMatch) int {
	var i int
	if strings.HasPrefix(fromStr, "marker") {
		i = 0
		for i < len(patternTokens) && (patternTokens[i].GetStartPos() < match.GetFromPos() || patternTokens[i].IsSentenceStart()) {
			i++
		}
		i++
		if len(fromStr) > 6 {
			off, err := strconv.Atoi(strings.ReplaceAll(fromStr, "marker", ""))
			if err != nil {
				panic(err)
			}
			i += off
		}
	} else {
		n, err := strconv.Atoi(fromStr)
		if err != nil {
			panic(err)
		}
		i = n
	}
	if i < 1 || i > len(patternTokens) {
		id := ""
		if match != nil && match.GetRule() != nil {
			id = "rule"
		}
		panic("RuleFilter: Index out of bounds in " + id + ", value: " + fromStr)
	}
	return i - 1
}

// daIsMatchAtSentenceStart ports RuleFilter.isMatchAtSentenceStart.
func daIsMatchAtSentenceStart(tokens []*languagetool.AnalyzedTokenReadings, match *rules.RuleMatch) bool {
	i := 0
	for i < len(tokens) && tokens[i].GetStartPos() < match.GetFromPos() {
		i++
	}
	for i > 0 && tools.IsPunctuationMark(tokens[i].GetToken()) {
		i--
	}
	return i == 0
}

func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
