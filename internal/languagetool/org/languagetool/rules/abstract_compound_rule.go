package rules

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

var (
	compoundDigitRE      = regexp.MustCompile(`\d+`)
	compoundWhitespaceRE = regexp.MustCompile(`\s+`)
	compoundDashesRE     = regexp.MustCompile(`--+`)
)

// AbstractCompoundRule ports org.languagetool.rules.AbstractCompoundRule.
// Java ctor: MISC, Misspelling.
type AbstractCompoundRule struct {
	Messages                   map[string]string
	ID                         string
	Description                string
	WithHyphenMessage          string
	WithoutHyphenMessage       string
	WithOrWithoutHyphenMessage string
	ShortDesc                  string
	// Category ports Rule.category (Java MISC).
	Category *Category
	// IssueType ports getLocQualityIssueType (Java Misspelling).
	IssueType ITSIssueType
	// URL ports Rule.url (Java setUrl).
	URL string
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []IncorrectExample
	correctExamples   []CorrectExample
	// SentenceStartsWithUpperCase uncapitalizes the first word when matching after SENT_START.
	SentenceStartsWithUpperCase bool
	SubRuleSpecificIDs          bool
	Data                        *CompoundRuleData
	// IsMisspelled optional; default treats all as correctly spelled (Java returns false).
	IsMisspelled func(word string) bool
}

// InitCompoundRuleMeta applies Java AbstractCompoundRule constructor metadata.
func InitCompoundRuleMeta(r *AbstractCompoundRule, messages map[string]string) {
	if r == nil {
		return
	}
	r.Messages = messages
	if r.Category == nil {
		r.Category = CatMisc.GetCategory(messages)
	}
	if r.IssueType == "" {
		r.IssueType = ITSMisspelling
	}
}

func (r *AbstractCompoundRule) GetID() string {
	if r.ID != "" {
		return r.ID
	}
	return "ABSTRACT_COMPOUND_RULE"
}

func (r *AbstractCompoundRule) GetDescription() string {
	return r.Description
}

func (r *AbstractCompoundRule) GetCategory() *Category {
	if r == nil {
		return nil
	}
	return r.Category
}

func (r *AbstractCompoundRule) GetLocQualityIssueType() ITSIssueType {
	if r == nil || r.IssueType == "" {
		return ITSMisspelling
	}
	return r.IssueType
}

// GetURL ports Rule.getUrl.
func (r *AbstractCompoundRule) GetURL() string {
	if r == nil {
		return ""
	}
	return r.URL
}

// SetURL ports Rule.setUrl.
func (r *AbstractCompoundRule) SetURL(u string) {
	if r != nil {
		r.URL = u
	}
}

// AddExamplePair ports Rule.addExamplePair.
func (r *AbstractCompoundRule) AddExamplePair(incorrect IncorrectExample, correct CorrectExample) {
	if r == nil {
		return
	}
	appendExamplePair(&r.incorrectExamples, &r.correctExamples, incorrect, correct)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *AbstractCompoundRule) GetIncorrectExamples() []IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *AbstractCompoundRule) GetCorrectExamples() []CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

func (r *AbstractCompoundRule) GetCompoundRuleData() *CompoundRuleData {
	return r.Data
}

func (r *AbstractCompoundRule) UseSubRuleSpecificIDs() {
	r.SubRuleSpecificIDs = true
}

// Match ports AbstractCompoundRule.match.
func (r *AbstractCompoundRule) Match(sentence *languagetool.AnalyzedSentence) []*RuleMatch {
	if r.Data == nil {
		return nil
	}
	tokens := sentence.GetTokensWithoutWhitespace()
	var ruleMatches []*RuleMatch
	var prevRuleMatch *RuleMatch
	prevTokens := make([]*languagetool.AnalyzedTokenReadings, 0, MaxCompoundTerms)
	hasDigitPatterns := r.Data.HasDigitPatterns

	// Extend token list with dummies so end-of-sentence compounds flush.
	limit := len(tokens) + MaxCompoundTerms
	for i := 0; i < limit; i++ {
		var token *languagetool.AnalyzedTokenReadings
		if i >= len(tokens) {
			start := 0
			if len(prevTokens) > 0 {
				start = prevTokens[0].GetStartPos()
			}
			empty := languagetool.NewAnalyzedToken("", nil, nil)
			token = languagetool.NewAnalyzedTokenReadingsAt(empty, start)
		} else {
			token = tokens[i]
		}
		if i == 0 {
			prevTokens = addToCompoundQueue(token, prevTokens)
			continue
		} else if token.IsImmunized() {
			continue
		}

		firstMatchToken := prevTokens[0]
		stringsToCheck, origStringsToCheck, stringToToken := r.getStringToTokenMap(prevTokens)

		for k := len(stringsToCheck) - 1; k >= 0; k-- {
			stringToCheck := stringsToCheck[k]
			origStringToCheck := origStringsToCheck[k]
			digitsRegexp := ""
			containsDigits := false
			if hasDigitPatterns {
				for _, part := range strings.Split(stringToCheck, " ") {
					if isAllDigits(part) {
						containsDigits = true
						break
					}
				}
			}
			matched := r.Data.ContainsIncorrect(stringToCheck)
			if !matched && containsDigits {
				digitsRegexp = compoundDigitRE.ReplaceAllString(stringToCheck, `\d+`)
				matched = r.Data.ContainsIncorrect(digitsRegexp)
			}
			if !matched {
				continue
			}

			atr := stringToToken[stringToCheck]
			if atr == nil {
				continue
			}
			var msg string
			var replacement []string

			if r.Data.ContainsDash(stringToCheck) && !strings.Contains(origStringToCheck, " ") {
				// already joined with hyphens
				break
			}
			if r.Data.ContainsDash(stringToCheck) || (containsDigits && r.Data.ContainsIncorrect(digitsRegexp)) {
				replacement = append(replacement, strings.ReplaceAll(origStringToCheck, " ", "-"))
				msg = r.WithHyphenMessage
			}
			if isNotAllUppercase(origStringToCheck) && r.Data.ContainsJoined(stringToCheck) {
				uncap := r.Data.JoinedLowerCaseAnyMatch(stringToCheck)
				replacement = append(replacement, r.MergeCompound(origStringToCheck, uncap))
				msg = r.WithoutHyphenMessage
			}
			parts := strings.Split(stringToCheck, " ")
			if len(parts) > 0 && len(parts[0]) == 1 {
				replacement = []string{strings.ReplaceAll(origStringToCheck, " ", "-")}
				msg = r.WithHyphenMessage
			} else if len(replacement) == 0 || len(replacement) == 2 {
				msg = r.WithOrWithoutHyphenMessage
			}

			from := firstMatchToken.GetStartPos()
			to := atr.GetEndPos()
			// original covered text for filter
			origCovered := sentence.GetText()
			// Use UTF-16-safe substring via token text rebuild when GetText spans full sentence.
			// Java: sentence.getText().substring(firstMatchToken.getStartPos(), atr.getEndPos())
			// Our positions are UTF-16; GetText is UTF-8. Prefer joining tokens when available.
			covered := utf16Slice(origCovered, from, to)
			replacement = r.filterReplacements(replacement, covered)
			if len(replacement) == 0 {
				break
			}

			ruleMatch := NewRuleMatch(r, sentence, from, to, msg)
			ruleMatch.ShortMessage = r.ShortDesc
			ruleMatch.SetSuggestedReplacements(replacement)

			if prevRuleMatch != nil && prevRuleMatch.GetFromPos() == ruleMatch.GetFromPos() {
				prevRuleMatch = ruleMatch
				break
			}
			prevRuleMatch = ruleMatch
			ruleMatches = append(ruleMatches, ruleMatch)
			break
		}
		prevTokens = addToCompoundQueue(token, prevTokens)
	}
	return ruleMatches
}

func (r *AbstractCompoundRule) filterReplacements(replacements []string, original string) []string {
	var out []string
	for _, rep := range replacements {
		newRep := compoundDashesRE.ReplaceAllString(rep, "-")
		if newRep != original && r.isCorrectSpell(newRep) {
			out = append(out, newRep)
		}
	}
	return out
}

func (r *AbstractCompoundRule) isCorrectSpell(word string) bool {
	if r.IsMisspelled == nil {
		return true // Java isMisspelled defaults to false → correct
	}
	return !r.IsMisspelled(word)
}

func (r *AbstractCompoundRule) getStringToTokenMap(prevTokens []*languagetool.AnalyzedTokenReadings) (
	stringsToCheck, origStringsToCheck []string,
	stringToToken map[string]*languagetool.AnalyzedTokenReadings,
) {
	stringToToken = make(map[string]*languagetool.AnalyzedTokenReadings, len(prevTokens)*2)
	var sb strings.Builder
	isFirstSentStart := false
	for j, atr := range prevTokens {
		if atr.IsWhitespaceBefore() {
			sb.WriteByte(' ')
		}
		sb.WriteString(atr.GetToken())
		if j == 0 {
			isFirstSentStart = atr.IsSentenceStart()
		}
		if j >= 1 || (j == 0 && !isFirstSentStart) {
			stringToCheck := normalizeCompound(sb.String())
			if r.SentenceStartsWithUpperCase && isFirstSentStart {
				stringToCheck = uncapitalize(stringToCheck)
			}
			stringsToCheck = append(stringsToCheck, stringToCheck)
			origStringsToCheck = append(origStringsToCheck, strings.TrimSpace(sb.String()))
			if _, ok := stringToToken[stringToCheck]; !ok {
				stringToToken[stringToCheck] = atr
			}
		}
	}
	return stringsToCheck, origStringsToCheck, stringToToken
}

func normalizeCompound(inStr string) string {
	str := strings.TrimSpace(inStr)
	str = strings.ReplaceAll(str, " - ", " ")
	str = strings.ReplaceAll(str, "-", " ")
	str = compoundWhitespaceRE.ReplaceAllString(str, " ")
	return str
}

func isNotAllUppercase(str string) bool {
	parts := strings.Split(str, " ")
	for _, part := range parts {
		if part == "-" {
			continue
		}
		if tools.IsAllUppercase(part) {
			return false
		}
	}
	return true
}

// MergeCompound ports AbstractCompoundRule.mergeCompound.
func (r *AbstractCompoundRule) MergeCompound(str string, uncapitalizeMidWords bool) string {
	stringParts := strings.Split(strings.ReplaceAll(str, "-", " "), " ")
	var sb strings.Builder
	for k, part := range stringParts {
		if k == 0 {
			sb.WriteString(part)
		} else if uncapitalizeMidWords {
			sb.WriteString(uncapitalize(part))
		} else {
			sb.WriteString(part)
		}
	}
	return sb.String()
}

func addToCompoundQueue(token *languagetool.AnalyzedTokenReadings, prev []*languagetool.AnalyzedTokenReadings) []*languagetool.AnalyzedTokenReadings {
	if len(prev) == MaxCompoundTerms {
		prev = prev[1:]
	}
	return append(prev, token)
}

func uncapitalize(s string) string {
	return tools.LowercaseFirstChar(s)
}

func isAllDigits(s string) bool {
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

// utf16Slice returns the UTF-8 substring corresponding to UTF-16 code unit range [from,to).
func utf16Slice(text string, from, to int) string {
	if from < 0 {
		from = 0
	}
	if to < from {
		return ""
	}
	u16 := 0
	startByte := -1
	endByte := len(text)
	for i, r := range text {
		w := 1
		if r >= 0x10000 {
			w = 2
		}
		if startByte < 0 && u16 >= from {
			startByte = i
		}
		if u16 >= to {
			endByte = i
			break
		}
		u16 += w
	}
	if startByte < 0 {
		return ""
	}
	// if to past end, endByte stays len
	if u16 < to {
		endByte = len(text)
	}
	// handle to exactly at end after last rune
	if startByte > endByte {
		return ""
	}
	// re-scan for precise end when last rune ends at to
	u16 = 0
	startByte = -1
	endByte = len(text)
	for i, r := range text {
		w := 1
		if r >= 0x10000 {
			w = 2
		}
		if startByte < 0 && u16 == from {
			startByte = i
		}
		if u16 == to {
			endByte = i
			break
		}
		u16 += w
	}
	if startByte < 0 {
		// from beyond text
		return ""
	}
	return text[startByte:endByte]
}
