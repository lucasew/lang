package ca

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// SynthesizeWithAnyDeterminerFilter ports
// org.languagetool.rules.ca.SynthesizeWithAnyDeterminerFilter (1:1 AcceptRuleMatch).
//
// GetPossibleTags / Synthesize / SynthesizeRE port CatalanSynthesizer.
// Without synthesizer hooks, only the original lemmaSelect reading is used.
type SynthesizeWithAnyDeterminerFilter struct {
	GetPossibleTags func() []string
	// Synthesize ports synth.synthesize(token, postag) exact POS.
	Synthesize func(tok *languagetool.AnalyzedToken, postag string) []string
	// SynthesizeRE ports synth.synthesize(token, postag, true) regex POS (determiners).
	SynthesizeRE func(tok *languagetool.AnalyzedToken, postagRE string) []string
}

func NewSynthesizeWithAnyDeterminerFilter() *SynthesizeWithAnyDeterminerFilter {
	return &SynthesizeWithAnyDeterminerFilter{}
}

// GenderNumberList is MS/FS/MP/FP.
var GenderNumberList = []string{"MS", "FS", "MP", "FP"}

// Prepositions recognized before determiners.
var PrepositionsList = []string{"a", "de", "per", "pe"}

// SuggestAll builds det+form for each gender/number (unit helper).
func (f *SynthesizeWithAnyDeterminerFilter) SuggestAll(forms []struct{ Form, POS string }, preposition, preferGN, casingModel string) []string {
	gns := make([]string, 0, 4)
	if preferGN != "" {
		gns = append(gns, preferGN)
	}
	for _, gn := range GenderNumberList {
		if gn != preferGN {
			gns = append(gns, gn)
		}
	}
	var out []string
	seen := map[string]struct{}{}
	for _, gn := range gns {
		for _, fr := range forms {
			if fr.POS != "" && GenderNumberFromPOS(fr.POS) != "" && GenderNumberFromPOS(fr.POS) != gn {
				continue
			}
			s := GetPrepositionAndDeterminer(fr.Form, gn, preposition) + fr.Form
			if casingModel != "" {
				s = tools.PreserveCase(s, casingModel)
			}
			if _, ok := seen[s]; ok {
				continue
			}
			seen[s] = struct{}{}
			out = append(out, s)
		}
	}
	return out
}

// PrepositionKey maps a full preposition token to the first-letter key.
func PrepositionKey(prep string) string {
	prep = strings.ToLower(tools.JavaStringTrim(prep))
	if prep == "" {
		return ""
	}
	return string([]rune(prep)[0])
}

// IsPreposition reports whether token is in PrepositionsList.
func IsPreposition(token string) bool {
	t := strings.ToLower(token)
	for _, p := range PrepositionsList {
		if t == p {
			return true
		}
	}
	return false
}

// AcceptRuleMatch ports SynthesizeWithAnyDeterminerFilter.acceptRuleMatch.
func (f *SynthesizeWithAnyDeterminerFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, patternTokenPos int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	_ = patternTokenPos
	_ = patternTokens
	_ = tokenPositions
	if f == nil || match == nil || match.Sentence == nil {
		return nil
	}
	if arguments == nil {
		return nil
	}

	synthAllForms := strings.EqualFold(patterns.GetOptionalDefault("synthAllForms", arguments, "false"), "true")
	lemmaSelect := patterns.GetRequired("lemmaSelect", arguments)

	tokens := match.Sentence.GetTokensWithoutWhitespace()
	posWord := 0
	for posWord < len(tokens) &&
		(tokens[posWord].GetStartPos() < match.GetFromPos() || tokens[posWord].IsSentenceStart()) {
		posWord++
	}
	if posWord >= len(tokens) {
		posWord = len(tokens) - 1
	}
	if posWord < 0 {
		return nil
	}

	originalWord := tokens[posWord].GetToken()
	originalAT := readingWithTagRegex(tokens[posWord], lemmaSelect)
	if originalAT == nil {
		panic("Cannot find analyzed token readings with postag " + lemmaSelect + " in sentence" + match.Sentence.GetText())
	}

	secondGenderNumber := ""
	determinerType := ""
	var determinerReading *languagetool.AnalyzedToken
	var betweenToken *languagetool.AnalyzedTokenReadings
	preposition := ""
	done := false
	k := 1
	firstUnderlinedToken := posWord
	for posWord-k > 0 && !done {
		done = true
		if determinerReading == nil {
			determinerReading = readingWithTagRegex(tokens[posWord-k], `D.*`)
			if determinerReading != nil && determinerReading.GetPOSTag() != nil {
				tag := *determinerReading.GetPOSTag()
				if len(tag) >= 5 {
					secondGenderNumber = tag[3:5]
					determinerType = tag[0:2]
				}
				done = determinerType != "DA"
				firstUnderlinedToken = posWord - k
			}
		}
		if tokens[posWord-k].HasPosTag("_QM_OPEN") {
			betweenToken = tokens[posWord-k]
			done = false
		}
		tokLower := strings.ToLower(tokens[posWord-k].GetToken())
		if IsPreposition(tokLower) {
			// a=a d=de per=p
			r := []rune(tokens[posWord-k].GetToken())
			if len(r) > 0 {
				preposition = strings.ToLower(string(r[0]))
			}
			firstUnderlinedToken = posWord - k
		}
		k++
	}
	betweenString := ""
	if betweenToken != nil {
		betweenString = betweenToken.GetToken()
	}

	p, err := regexp.Compile(lemmaSelect)
	if err != nil {
		panic(err)
	}

	// original word form in the first place
	potentialSuggestions := []*languagetool.AnalyzedToken{originalAT}
	// second-best suggestion from the determiner
	if f.GetPossibleTags != nil && f.Synthesize != nil {
		for _, tag := range f.GetPossibleTags() {
			if !swadFullMatch(p, tag) {
				continue
			}
			for _, synthForm := range f.Synthesize(originalAT, tag) {
				if !synthAllForms && !strings.EqualFold(synthForm, originalWord) {
					continue
				}
				tagCopy := tag
				at := languagetool.NewAnalyzedToken(synthForm, &tagCopy, originalAT.GetLemma())
				if listContainsAnalyzedTokenLemmaPOS(potentialSuggestions, at) {
					continue
				}
				// Java: tag.contains(secondGenderNumber) || tag.contains(swap)
				// empty secondGenderNumber → contains("") is true for all tags
				prefer := strings.Contains(tag, secondGenderNumber)
				if !prefer && len(secondGenderNumber) >= 2 {
					prefer = strings.Contains(tag, string(secondGenderNumber[1])+string(secondGenderNumber[0]))
				}
				if prefer {
					potentialSuggestions = append(potentialSuggestions[:1],
						append([]*languagetool.AnalyzedToken{at}, potentialSuggestions[1:]...)...)
				} else {
					potentialSuggestions = append(potentialSuggestions, at)
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
		for _, genderNumber := range GenderNumberList {
			if !daFullMatch(genderNumberPatterns[genderNumber], posTag) {
				continue
			}
			if determinerType == "DA" || determinerType == "" {
				suggestion := tools.PreserveCase(
					GetPrepositionAndDeterminer(newForm, genderNumber, preposition),
					tokens[firstUnderlinedToken].GetToken()) + betweenString +
					tools.PreserveCase(newForm, originalWord)
				if firstUnderlinedToken == 1 {
					suggestion = tools.UppercaseFirstChar(suggestion)
				}
				if !containsString(suggestions, suggestion) {
					suggestions = append(suggestions, suggestion)
				}
			} else if determinerReading != nil && f.SynthesizeRE != nil && len(genderNumber) >= 2 {
				// determinerType + ".[C" + gender + "]" + number + "."
				postagRE := determinerType + ".[C" + string(genderNumber[0]) + "]" + string(genderNumber[1]) + "."
				for _, synthForm := range f.SynthesizeRE(determinerReading, postagRE) {
					suggestion := tools.PreserveCase(synthForm, tokens[firstUnderlinedToken].GetToken()) +
						" " + betweenString + tools.PreserveCase(newForm, originalWord)
					if firstUnderlinedToken == 1 {
						suggestion = tools.UppercaseFirstChar(suggestion)
					}
					if !containsString(suggestions, suggestion) {
						suggestions = append(suggestions, suggestion)
					}
				}
			}
		}
	}

	out := rules.NewRuleMatch(match.GetRule(), match.Sentence,
		tokens[firstUnderlinedToken].GetStartPos(), match.GetToPos(),
		match.GetMessage())
	out.ShortMessage = match.GetShortMessage()
	out.SetSuggestedReplacements(suggestions)
	return out
}

// listContainsAnalizedToken ports private helper (lemma + POS only).
func listContainsAnalyzedTokenLemmaPOS(list []*languagetool.AnalyzedToken, at *languagetool.AnalyzedToken) bool {
	if at == nil {
		return false
	}
	for _, item := range list {
		if item == nil {
			continue
		}
		if strPtrEqCA(item.GetLemma(), at.GetLemma()) && strPtrEqCA(item.GetPOSTag(), at.GetPOSTag()) {
			return true
		}
	}
	return false
}

func strPtrEqCA(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func swadFullMatch(re *regexp.Regexp, s string) bool {
	if re == nil {
		return false
	}
	loc := re.FindStringIndex(s)
	return loc != nil && loc[0] == 0 && loc[1] == len(s)
}
