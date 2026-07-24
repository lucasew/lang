package ca

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// OblidarseSugestionsFilter ports
// org.languagetool.rules.ca.OblidarseSugestionsFilter (1:1 AcceptRuleMatch).
//
// Synthesize ports Synthesizer.synthesize(token, postag) without POS-regex.
// When nil, Accept returns nil (fail-closed).
type OblidarseSugestionsFilter struct {
	// Synthesize ports getSynthesizerFromRuleMatch(...).synthesize(token, postag).
	Synthesize func(tok *languagetool.AnalyzedToken, postag string) []string
}

func NewOblidarseSugestionsFilter() *OblidarseSugestionsFilter {
	return &OblidarseSugestionsFilter{}
}

// Reflexive prefix tables — Java static HashMaps.
var addReflexiveVowel = map[string]string{
	"1S": "m'",
	"2S": "t'",
	"3S": "s'",
	"1P": "ens ",
	"2P": "us ",
	"3P": "s'",
}
var addReflexiveConsonant = map[string]string{
	"1S": "em ",
	"2S": "et ",
	"3S": "es ",
	"1P": "ens ",
	"2P": "us ",
	"3P": "es ",
}
var addReflexiveEnVowel = map[string]string{
	"1S": "me n'",
	"2S": "te n'",
	"3S": "se n'",
	"1P": "ens n'",
	"2P": "us n'",
	"3P": "se n'",
}
var addReflexiveEnConsonant = map[string]string{
	"1S": "me'n ",
	"2S": "te'n ",
	"3S": "se'n ",
	"1P": "ens en ",
	"2P": "us en ",
	"3P": "se'n ",
}

// Java Pattern.CASE_INSENSITIVE on h?[aeiouàèéíòóú].*
var pApostropheNeededOblidar = regexp.MustCompile(`(?i)^h?[aeiouàèéíòóú].*`)

// ReflexivePrefix returns the weak-pronoun prefix for personNumber (e.g. "1S").
func (f *OblidarseSugestionsFilter) ReflexivePrefix(personNumber string, nextNeedsApos, withEn bool) string {
	if withEn {
		if nextNeedsApos {
			return addReflexiveEnVowel[personNumber]
		}
		return addReflexiveEnConsonant[personNumber]
	}
	if nextNeedsApos {
		return addReflexiveVowel[personNumber]
	}
	return addReflexiveConsonant[personNumber]
}

// NeedsApostrophe reports vowel-initial following words.
func (f *OblidarseSugestionsFilter) NeedsApostrophe(nextWord string) bool {
	return pApostropheNeededOblidar.MatchString(nextWord)
}

// AcceptRuleMatch ports OblidarseSugestionsFilter.acceptRuleMatch.
func (f *OblidarseSugestionsFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, patternTokenPos int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	_ = arguments
	_ = patternTokenPos
	_ = patternTokens
	_ = tokenPositions
	if f == nil || match == nil || match.Sentence == nil || f.Synthesize == nil {
		return nil
	}

	tokens := match.Sentence.GetTokensWithoutWhitespace()
	posWord := 0
	for posWord < len(tokens) &&
		(tokens[posWord].GetStartPos() < match.GetFromPos() || tokens[posWord].IsSentenceStart()) {
		posWord++
	}
	// Need posWord+1 (pronom) and posWord+2 (first verb of group)
	if posWord+2 >= len(tokens) {
		return nil
	}

	pr := readingWithTagRegex(tokens[posWord+1], `P.*`)
	if pr == nil || pr.GetPOSTag() == nil {
		return nil
	}
	pronomPostag := *pr.GetPOSTag()
	if len(pronomPostag) < 5 {
		return nil
	}
	pronomGenderNumber := pronomPostag[2:3] + pronomPostag[4:5]

	indexMainVerb := posWord + 2
	for indexMainVerb < len(tokens) &&
		!tokens[indexMainVerb].HasAnyLemma("oblidar", "descuidar", "passar") {
		indexMainVerb++
	}
	if indexMainVerb >= len(tokens) {
		return nil
	}

	// Java always uses posWord+2 for verbPostag/lemma (first V of group), not indexMainVerb.
	vr := readingWithTagRegex(tokens[posWord+2], `V.*`)
	if vr == nil || vr.GetPOSTag() == nil {
		return nil
	}
	verbPostag := *vr.GetPOSTag()
	if len(verbPostag) < 8 {
		return nil
	}
	lemma := ""
	if vr.GetLemma() != nil {
		lemma = *vr.GetLemma()
	}
	if lemma == "passar" {
		lemma = "descuidar"
	}

	synthForms := f.Synthesize(languagetool.NewAnalyzedToken("", nil, &lemma),
		verbPostag[:4]+pronomGenderNumber+verbPostag[6:8])
	if len(synthForms) == 0 {
		return nil
	}
	newVerb := synthForms[0]
	for i := posWord + 3; i < indexMainVerb+1; i++ {
		if i < 0 || i >= len(tokens) {
			break
		}
		tok := tokens[i].GetToken()
		tok = strings.ReplaceAll(tok, "passar", "descuidar")
		tok = strings.ReplaceAll(tok, "passat", "descuidat")
		tok = strings.ReplaceAll(tok, "passant", "descuidant")
		// Java: getWhitespaceBefore() string
		newVerb = newVerb + tokens[i].GetWhitespaceBefore() + tok
	}

	verbVowel := pApostropheNeededOblidar.MatchString(newVerb)
	wordAfter := ""
	if indexMainVerb+1 < len(tokens) {
		wordAfterReading := readingWithTagRegex(tokens[indexMainVerb+1], `D.*|V.N.*|P[DI].*|NC.*`)
		if wordAfterReading != nil {
			wordAfter = wordAfterReading.GetToken()
		}
		// exceptions: com, de, d', que
		low := strings.ToLower(tokens[indexMainVerb+1].GetToken())
		switch low {
		case "com", "de", "d'", "que":
			wordAfter = tokens[indexMainVerb+1].GetToken()
		}
	}

	// Java condition kept bug-for-bug: en-forms only when wordAfter is empty.
	var transform map[string]string
	if wordAfter == "" && !strings.EqualFold(wordAfter, "de") && !strings.EqualFold(wordAfter, "d'") &&
		!strings.EqualFold(wordAfter, "que") {
		if verbVowel {
			transform = addReflexiveEnVowel
		} else {
			transform = addReflexiveEnConsonant
		}
	} else {
		if verbVowel {
			transform = addReflexiveVowel
		} else {
			transform = addReflexiveConsonant
		}
	}

	prefix, ok := transform[pronomGenderNumber]
	if !ok {
		return nil
	}
	var suggBld strings.Builder
	suggBld.WriteString(prefix)
	suggBld.WriteString(newVerb)

	charactersAfterCorrection := 0
	if strings.EqualFold(wordAfter, "el") || strings.EqualFold(wordAfter, "els") {
		suggBld.WriteString(" d")
		suggBld.WriteString(strings.ToLower(wordAfter))
		charactersAfterCorrection = len(wordAfter) + 1
	} else if wordAfter != "" && !strings.EqualFold(wordAfter, "de") &&
		!strings.EqualFold(wordAfter, "d'") && !strings.EqualFold(wordAfter, "que") {
		wordAfterApostrophe := pApostropheNeededOblidar.MatchString(wordAfter)
		if wordAfterApostrophe {
			suggBld.WriteString(" d'")
			charactersAfterCorrection = 1
		} else {
			suggBld.WriteString(" de")
			charactersAfterCorrection = 0
		}
	}

	replacement := tools.PreserveCase(suggBld.String(), tokens[posWord].GetToken())
	var replacements []string
	replacements = append(replacements, replacement)
	for _, s := range match.GetSuggestedReplacements() {
		if charactersAfterCorrection == 1 {
			s = s + " "
		}
		replacements = append(replacements, AdaptSuggestion(s, tokens[posWord].GetToken()))
	}
	if len(replacements) == 0 {
		return nil
	}

	msg := strings.ReplaceAll(match.GetMessage(), "passar", "descuidar")
	toPos := tokens[indexMainVerb].GetEndPos() + charactersAfterCorrection
	out := rules.NewRuleMatch(match.GetRule(), match.Sentence,
		tokens[posWord].GetStartPos(), toPos, msg)
	out.ShortMessage = match.GetShortMessage()
	out.SetSuggestedReplacements(replacements)
	return out
}
