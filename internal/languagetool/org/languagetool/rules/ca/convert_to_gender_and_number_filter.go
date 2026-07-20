package ca

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// ConvertToGenderAndNumberFilter ports
// org.languagetool.rules.ca.ConvertToGenderAndNumberFilter (1:1 AcceptRuleMatch).
//
// Synthesize ports synthesizer.synthesize(token, postagRE, true).
// Tag ports tagger.tag(singleton list) for rewriting existing suggestions.
type ConvertToGenderAndNumberFilter struct {
	// Synthesize ports synth.synthesize(atr, postagRegex, true).
	Synthesize func(tok *languagetool.AnalyzedToken, postagRE string) []string
	// Tag ports tagger.tag([]string{word}); returns readings for the word.
	Tag func(word string) *languagetool.AnalyzedTokenReadings
}

func NewConvertToGenderAndNumberFilter() *ConvertToGenderAndNumberFilter {
	return &ConvertToGenderAndNumberFilter{}
}

// GenderAndNumberSplit holds pieces of a Catalan POS tag around gender/number slots.
type GenderAndNumberSplit struct {
	Prefix string
	Suffix string
	Gender string
	Number string
}

// Java Pattern.compile strings (Matcher.matches = full string).
var (
	splitGenderNumber          = regexp.MustCompile(`(N.|A..|V.P..|D..|PX.)(.)(.)(.*)`)
	splitGenderNumberNoNoun    = regexp.MustCompile(`(A..|V.P..|D..|PX.)(.)(.)(.*)`)
	splitGenderNumberAdjective = regexp.MustCompile(`(A..|V.P..|PX.)(.)(.)(.*)`)
	postagExceptionsGN         = regexp.MustCompile(`NP.*|AQ0CN0|SPS00|[CP].*`)
)

// FormsToIgnore are tokens that stop gender/number expansion.
var FormsToIgnore = map[string]struct{}{
	"mes": {}, "las": {},
}

// SplitGenderAndNumber parses gender/number slots from a POS tag string.
func SplitGenderAndNumber(pos string) *GenderAndNumberSplit {
	return splitGenderAndNumberFromTag(pos)
}

func splitGenderAndNumberFromTag(pos string) *GenderAndNumberSplit {
	if pos == "" {
		return nil
	}
	m := splitGenderNumber.FindStringSubmatch(pos)
	if m == nil || !cgnFullMatch(splitGenderNumber, pos) {
		// FindStringSubmatch still works with unanchored RE2 as leftmost match;
		// require full-string match like Java Matcher.matches().
		if m == nil {
			return nil
		}
		// re-check full
		loc := splitGenderNumber.FindStringIndex(pos)
		if loc == nil || loc[0] != 0 || loc[1] != len(pos) {
			return nil
		}
	}
	res := &GenderAndNumberSplit{
		Prefix: m[1],
		Suffix: m[4],
	}
	g2, g3 := m[2], m[3]
	if strings.HasPrefix(res.Prefix, "V") {
		res.Gender = g3
		res.Number = g2
	} else {
		res.Gender = g2
		res.Number = g3
	}
	return res
}

func splitGenderAndNumberToken(atr *languagetool.AnalyzedToken) *GenderAndNumberSplit {
	if atr == nil || atr.GetPOSTag() == nil {
		return nil
	}
	return splitGenderAndNumberFromTag(*atr.GetPOSTag())
}

// DesiredPostag builds a synthesizer postag pattern for desired gender/number.
func (f *ConvertToGenderAndNumberFilter) DesiredPostag(split *GenderAndNumberSplit, gender, number string) string {
	if split == nil {
		return ""
	}
	g, n := gender, number
	if strings.HasPrefix(split.Prefix, "V") {
		g, n = number, gender
	}
	addGender := "C"
	if strings.HasPrefix(split.Prefix, "DA") {
		addGender = ""
	}
	return split.Prefix + "[" + g + addGender + "]" + "[" + n + "N" + "]" + split.Suffix
}

// ShouldIgnoreForm reports tokens that stop expansion.
func ShouldIgnoreForm(token string) bool {
	_, ok := FormsToIgnore[strings.ToLower(token)]
	return ok
}

// IsPostagException reports POS tags excluded from gender/number rewrite.
func IsPostagException(pos string) bool {
	return cgnFullMatch(postagExceptionsGN, pos)
}

// BoToBon special-cases Catalan "bo" → "bon" before nouns.
func BoToBon(s string) string {
	if s == "bo" {
		return "bon"
	}
	return s
}

// MatchesSplitGenderNumber reports whether POS matches the main split pattern.
func MatchesSplitGenderNumber(pos string) bool {
	return cgnFullMatch(splitGenderNumber, pos)
}

// MatchesAdjectiveSplit reports adjective/participle POS splits.
func MatchesAdjectiveSplit(pos string) bool {
	return cgnFullMatch(splitGenderNumberAdjective, pos)
}

func cgnFullMatch(re *regexp.Regexp, s string) bool {
	if re == nil {
		return false
	}
	loc := re.FindStringIndex(s)
	return loc != nil && loc[0] == 0 && loc[1] == len(s)
}

func readingWithRE(tok *languagetool.AnalyzedTokenReadings, re *regexp.Regexp) *languagetool.AnalyzedToken {
	if tok == nil || re == nil {
		return nil
	}
	for _, r := range tok.GetReadings() {
		if r == nil {
			continue
		}
		posTag := "UNKNOWN"
		if pt := r.GetPOSTag(); pt != nil {
			posTag = *pt
		}
		if cgnFullMatch(re, posTag) {
			return r
		}
	}
	return nil
}

// AcceptRuleMatch ports ConvertToGenderAndNumberFilter.acceptRuleMatch.
func (f *ConvertToGenderAndNumberFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, patternTokenPos int,
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

	tokens := match.Sentence.GetTokensWithoutWhitespace()
	posWord := 0
	for posWord < len(tokens) &&
		(tokens[posWord].GetStartPos() < match.GetFromPos() || tokens[posWord].IsSentenceStart()) {
		posWord++
	}
	if posWord >= len(tokens) {
		return nil
	}

	desiredGenderOrigStr := patterns.GetOptionalDefault("gender", arguments, "")
	desiredNumberOrigStr := patterns.GetOptionalDefault("number", arguments, "")
	lemmaSelect := patterns.GetRequired("lemmaSelect", arguments)
	newLemma := patterns.GetOptionalDefault("newLemma", arguments, "")
	keepOriginal := strings.EqualFold(patterns.GetOptionalDefault("keepOriginal", arguments, "false"), "true")

	var suggestions []string
	var atrNounList []*languagetool.AnalyzedToken
	atrNounOrig := readingWithTagRegex(tokens[posWord], lemmaSelect)

	if newLemma != "" && atrNounOrig != nil {
		at := languagetool.NewAnalyzedToken(atrNounOrig.GetToken(), atrNounOrig.GetPOSTag(), &newLemma)
		atrNounList = append(atrNounList, at)
	} else if len(match.GetSuggestedReplacements()) > 0 {
		if f.Tag == nil {
			// need tagger for this path
			return nil
		}
		var splitNounOrigPostag *GenderAndNumberSplit
		if atrNounOrig != nil {
			splitNounOrigPostag = splitGenderAndNumberToken(atrNounOrig)
		}
		for _, suggestion := range match.GetSuggestedReplacements() {
			parts := strings.SplitN(suggestion, " ", 2)
			word := parts[0]
			remainder := ""
			if len(parts) > 1 {
				remainder = parts[1]
			}
			atrs := f.Tag(word)
			if atrs == nil {
				suggestions = append(suggestions, match.GetSuggestedReplacements()...)
				atrNounList = nil
				break
			}
			at := readingWithRE(atrs, splitGenderNumber)
			if at == nil || at.GetPOSTag() == nil || readingWithRE(atrs, postagExceptionsGN) != nil {
				suggestions = append(suggestions, match.GetSuggestedReplacements()...)
				atrNounList = nil
				break
			}
			splitPostag := splitGenderAndNumberToken(at)
			if splitPostag == nil {
				suggestions = append(suggestions, match.GetSuggestedReplacements()...)
				atrNounList = nil
				break
			}
			var newPostag strings.Builder
			newPostag.WriteString(splitPostag.Prefix)
			number := ""
			if desiredNumberOrigStr != "" {
				number = desiredNumberOrigStr
			} else if splitNounOrigPostag != nil && splitNounOrigPostag.Number != "N" {
				number = splitNounOrigPostag.Number
			} else {
				number = splitPostag.Number
			}
			gender := desiredGenderOrigStr
			if gender == "" {
				gender = splitPostag.Gender
			}
			if strings.HasPrefix(splitPostag.Prefix, "V") {
				newPostag.WriteString(number)
				newPostag.WriteString(gender)
			} else {
				newPostag.WriteString(gender)
				newPostag.WriteString(number)
			}
			newPostag.WriteString(splitPostag.Suffix)
			completeForm := ""
			if at.GetLemma() != nil {
				completeForm = *at.GetLemma()
			}
			if remainder != "" {
				completeForm = word + " " + remainder
			}
			posStr := newPostag.String()
			lemma := at.GetLemma()
			at2 := languagetool.NewAnalyzedToken(completeForm, &posStr, lemma)
			atrNounList = append(atrNounList, at2)
		}
	} else {
		if atrNounOrig == nil {
			return nil
		}
		atrNounList = append(atrNounList, atrNounOrig)
	}

	startPos := posWord
	endPos := posWord
	for _, atrNoun := range atrNounList {
		if atrNoun == nil {
			continue
		}
		startPos = posWord
		endPos = posWord
		splitPostag := splitGenderAndNumberToken(atrNoun)
		if splitPostag == nil {
			continue
		}
		desiredGenderStr := desiredGenderOrigStr
		if desiredGenderStr == "" {
			desiredGenderStr = splitPostag.Gender
		}
		desiredNumberStr := desiredNumberOrigStr
		if desiredNumberStr == "" {
			desiredNumberStr = splitPostag.Number
		}
		// if gender = C, look into the words before and after
		if desiredGenderStr == "C" && posWord-1 > 0 {
			splitPostag2 := splitGenderAndNumberToken(readingWithRE(tokens[posWord-1], splitGenderNumber))
			if splitPostag2 != nil && (splitPostag2.Gender == "F" || splitPostag2.Gender == "M") {
				desiredGenderStr = splitPostag2.Gender
			}
		}
		if desiredGenderStr == "C" && posWord+1 < len(tokens) {
			splitPostag2 := splitGenderAndNumberToken(readingWithRE(tokens[posWord+1], splitGenderNumber))
			if splitPostag2 != nil && (splitPostag2.Gender == "F" || splitPostag2.Gender == "M") {
				desiredGenderStr = splitPostag2.Gender
			}
		}
		// if number = N, look into the words before and after
		if desiredNumberStr == "N" && posWord-1 > 0 {
			splitPostag2 := splitGenderAndNumberToken(readingWithRE(tokens[posWord-1], splitGenderNumber))
			if splitPostag2 != nil && (splitPostag2.Number == "S" || splitPostag2.Number == "P") {
				desiredNumberStr = splitPostag2.Number
			}
		}
		if desiredNumberStr == "N" && posWord+1 < len(tokens) {
			splitPostag2 := splitGenderAndNumberToken(readingWithRE(tokens[posWord+1], splitGenderNumber))
			if splitPostag2 != nil && (splitPostag2.Number == "S" || splitPostag2.Number == "P") {
				desiredNumberStr = splitPostag2.Number
			}
		}
		// Prioritize gender and number in the original
		if desiredGenderStr != "" && strings.Contains(desiredGenderStr, splitPostag.Gender) {
			desiredGenderStr = splitPostag.Gender + strings.ReplaceAll(desiredGenderStr, splitPostag.Gender, "")
		}
		if desiredNumberStr != "" && strings.Contains(desiredNumberStr, splitPostag.Number) {
			desiredNumberStr = splitPostag.Number + strings.ReplaceAll(desiredNumberStr, splitPostag.Number, "")
		}

		for _, genderCh := range desiredGenderStr {
			for _, numberCh := range desiredNumberStr {
				desiredGender := string(genderCh)
				desiredNumber := string(numberCh)
				var suggestionBuilder strings.Builder
				ignoreThisSuggestion := false
				if !keepOriginal {
					s := f.synthesizeWithGenderAndNumber(atrNoun, splitPostag, desiredGender, desiredNumber)
					if s == "" {
						ignoreThisSuggestion = true
					}
					suggestionBuilder.WriteString(s)
				} else {
					suggestionBuilder.WriteString(atrNoun.GetToken())
				}
				// backwards
				stop := false
				i := posWord
				prepositionToAdd := ""
				addDeterminer := false
				addedDemonstrative := false
				var conditionalAddedString strings.Builder
				addTot := ""
				for !stop && i > 1 {
					i--
					atr := readingWithRE(tokens[i], splitGenderNumberNoNoun)
					if (!tokens[i].HasPosTagStartingWith("D") && // incloem l'article fins i tot si està marcat com a _GV_
						(tokens[i].HasPosTag("_perfet") || tokens[i].HasPosTag("_GV_") || hasChunkTag(tokens[i], "GV"))) ||
						ShouldIgnoreForm(tokens[i].GetToken()) {
						atr = nil
					}
					if atr != nil && atr.GetPOSTag() != nil && atr.GetLemma() != nil {
						if strings.HasPrefix(*atr.GetPOSTag(), "DA") {
							suggestionBuilder.WriteString("") // no-op; insert below
							// insert conditional at front
							s := suggestionBuilder.String()
							suggestionBuilder.Reset()
							suggestionBuilder.WriteString(conditionalAddedString.String())
							suggestionBuilder.WriteString(s)
							conditionalAddedString.Reset()
							addDeterminer = true
							startPos = i
						} else {
							if !addDeterminer && !addedDemonstrative {
								s := f.synthesizeWithGenderAndNumber(atr, splitGenderAndNumberToken(atr), desiredGender, desiredNumber)
								if s == "" {
									ignoreThisSuggestion = true
								}
								if s == "bo" {
									s = "bon"
								}
								cur := suggestionBuilder.String()
								suggestionBuilder.Reset()
								suggestionBuilder.WriteString(conditionalAddedString.String())
								conditionalAddedString.Reset()
								if i+1 < len(tokens) && tokens[i+1].IsWhitespaceBefore() {
									suggestionBuilder.WriteString(" ")
								}
								suggestionBuilder.WriteString(s)
								suggestionBuilder.WriteString(cur)
								startPos = i
								if strings.HasPrefix(*atr.GetPOSTag(), "DD") {
									addedDemonstrative = true
								}
								if strings.HasPrefix(*atr.GetPOSTag(), "D") && !strings.HasPrefix(*atr.GetPOSTag(), "DN") &&
									!addedDemonstrative && !strings.EqualFold(*atr.GetLemma(), "quant") {
									stop = true
								}
							} else {
								// only before "el/aquest/aquell...": tota l'estona
								if atr.GetLemma() != nil && *atr.GetLemma() == "tot" {
									s := f.synthesizeWithGenderAndNumber(atr, splitGenderAndNumberToken(atr), desiredGender, desiredNumber)
									if s != "" {
										addTot = s + " "
										startPos = i
									}
								}
								stop = true
							}
						}
					} else if tokens[i].HasPosTag("SPS00") || tokens[i].HasPosTag("LOC_PREP") {
						if addDeterminer {
							preposition := strings.ToLower(tokens[i].GetToken())
							if preposition == "pe" {
								preposition = "per"
							}
							if preposition == "d'" {
								preposition = "de"
							}
							if preposition == "a" || preposition == "de" || preposition == "per" {
								prepositionToAdd = preposition
								startPos = i
							}
						}
						stop = true
					} else if tokens[i].HasPosTag("_PUNCT_CONT") || tokens[i].HasPosTag("CC") {
						if (posWord-i == 1) ||
							(i > 1 && strings.EqualFold(tokens[i].GetToken(), "i") &&
								strings.EqualFold(tokens[i-1].GetToken(), "tot")) {
							stop = true
						} else {
							// insert at front of conditional
							cs := conditionalAddedString.String()
							conditionalAddedString.Reset()
							conditionalAddedString.WriteString(tokens[i].GetToken() + " ")
							conditionalAddedString.WriteString(cs)
						}
					} else if tokens[i].HasPosTagStartingWith("RG") &&
						(i <= 1 || readingWithRE(tokens[i-1], splitGenderNumber) != nil) &&
						(i >= len(tokens)-1 || readingWithRE(tokens[i+1], splitGenderNumber) != nil) {
						cs := conditionalAddedString.String()
						conditionalAddedString.Reset()
						conditionalAddedString.WriteString(tokens[i].GetToken() + " ")
						conditionalAddedString.WriteString(cs)
					} else {
						stop = true
					}
				}
				// forwards
				stop = false
				i = posWord
				conditionalAddedString.Reset()
				isThereConjunction := false
				for !stop && i < len(tokens)-1 {
					i++
					if ShouldIgnoreForm(tokens[i].GetToken()) {
						break
					}
					atr := readingWithRE(tokens[i], splitGenderNumberAdjective)
					if isThereConjunction && tokens[i].HasPosTagStartingWith("NC") {
						atr = nil
					}
					if atr != nil {
						s := f.synthesizeWithGenderAndNumber(atr, splitGenderAndNumberToken(atr), desiredGender, desiredNumber)
						if s == "" {
							ignoreThisSuggestion = true
						}
						suggestionBuilder.WriteString(conditionalAddedString.String())
						conditionalAddedString.Reset()
						suggestionBuilder.WriteString(" ")
						suggestionBuilder.WriteString(s)
						endPos = i
					} else if tokens[i].HasPosTagStartingWith("RG") {
						conditionalAddedString.WriteString(" ")
						conditionalAddedString.WriteString(tokens[i].GetToken())
					} else if tokens[i].HasPosTag("CC") {
						isThereConjunction = true
						conditionalAddedString.WriteString(" ")
						conditionalAddedString.WriteString(tokens[i].GetToken())
					} else if tokens[i].HasPosTag("_PUNCT_CONT") {
						conditionalAddedString.WriteString(tokens[i].GetToken())
					} else {
						stop = true
					}
				}
				if addDeterminer {
					body := suggestionBuilder.String()
					suggestionBuilder.Reset()
					suggestionBuilder.WriteString(GetPrepositionAndDeterminer(body, desiredGender+desiredNumber, prepositionToAdd))
					suggestionBuilder.WriteString(body)
				} else if prepositionToAdd != "" {
					body := suggestionBuilder.String()
					suggestionBuilder.Reset()
					suggestionBuilder.WriteString(prepositionToAdd + " ")
					suggestionBuilder.WriteString(body)
				}
				if addTot != "" {
					body := suggestionBuilder.String()
					suggestionBuilder.Reset()
					suggestionBuilder.WriteString(addTot)
					suggestionBuilder.WriteString(body)
				}
				text := match.Sentence.GetText()
				from := tokens[startPos].GetStartPos()
				to := tokens[endPos].GetEndPos()
				originalSpan := ""
				if from >= 0 && to <= len(text) && from <= to {
					originalSpan = text[from:to]
				}
				suggestion := preserveCaseWordByWord(suggestionBuilder.String(), originalSpan)
				if endPos == posWord && startPos == posWord && tokens[posWord].GetToken() == suggestion {
					continue
				}
				if !ignoreThisSuggestion {
					suggestions = append(suggestions, suggestion)
				}
			}
		}
	}

	if len(suggestions) == 0 {
		return nil
	}
	out := rules.NewRuleMatch(match.GetRule(), match.Sentence,
		tokens[startPos].GetStartPos(), tokens[endPos].GetEndPos(),
		match.GetMessage())
	out.ShortMessage = match.GetShortMessage()
	text := match.Sentence.GetText()
	from := tokens[startPos].GetStartPos()
	to := tokens[endPos].GetEndPos()
	originalStr := ""
	if from >= 0 && to <= len(text) && from <= to {
		originalStr = text[from:to]
	}
	for _, s := range suggestions {
		if s == originalStr {
			return nil
		}
	}
	out.SetSuggestedReplacements(suggestions)
	return out
}

func (f *ConvertToGenderAndNumberFilter) synthesizeWithGenderAndNumber(
	atr *languagetool.AnalyzedToken, splitPostag *GenderAndNumberSplit, gender, number string,
) string {
	if f == nil || f.Synthesize == nil || atr == nil || splitPostag == nil {
		return ""
	}
	parts := strings.SplitN(atr.GetToken(), " ", 2)
	remainder := ""
	if len(parts) > 1 {
		remainder = parts[1]
	}
	g, n := gender, number
	if strings.HasPrefix(splitPostag.Prefix, "V") {
		g, n = number, gender
	}
	addGender := "C"
	if strings.HasPrefix(splitPostag.Prefix, "DA") {
		addGender = ""
	}
	postagRE := splitPostag.Prefix + "[" + g + addGender + "]" + "[" + n + "N" + "]" + splitPostag.Suffix
	synthesized := f.Synthesize(atr, postagRE)
	if len(synthesized) == 0 {
		return ""
	}
	synthesizedSuggestion := synthesized[0]
	if remainder != "" {
		synthesizedSuggestion = synthesizedSuggestion + " " + remainder
	}
	return synthesizedSuggestion
}

// preserveCaseWordByWord delegates to tools.PreserveCaseWordByWord (StringTools).
func preserveCaseWordByWord(inputString, modelString string) string {
	return tools.PreserveCaseWordByWord(inputString, modelString)
}
