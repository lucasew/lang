package rules

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// CaseSensitivity ports AbstractSimpleReplaceRule2.CaseSensitivy.
type CaseSensitivity int

const (
	// CaseSensitive matches exact case (CS).
	CaseSensitive CaseSensitivity = iota
	// CaseInsensitive lowercases keys (CI) — default for locale replace rules.
	CaseInsensitive
	// CaseSensitiveExceptAtSentenceStart ports CSExceptAtSentenceStart.
	CaseSensitiveExceptAtSentenceStart
)

const maxTokensInMultiword = 20

// SuggestionWithMessage ports org.languagetool.rules.SuggestionWithMessage.
type SuggestionWithMessage struct {
	Suggestion string
	Message    string
}

// AbstractSimpleReplaceRule2 ports org.languagetool.rules.AbstractSimpleReplaceRule2.
type AbstractSimpleReplaceRule2 struct {
	Messages             map[string]string
	ID                   string
	Description          string
	ShortMsg             string
	MessageTemplate      string // $match, $suggestions
	SuggestionsSeparator string
	CaseSens             CaseSensitivity
	SubRuleSpecificIDs   bool
	CheckingCase         bool
	LanguageCode         string
	// MatchShortAllUpperInCheckCase ports setIgnoreShortUppercaseWords(false):
	// when true, short ALLCAPS tokens (len≤4) still match in CheckingCase mode (Dutch).
	MatchShortAllUpperInCheckCase bool

	IsException          func(matchedText string) bool
	IsTokenException     func(atr *languagetool.AnalyzedTokenReadings) bool
	IsRuleMatchException func(m *RuleMatch) bool

	mStartSpace   map[string]int
	mStartNoSpace map[string]int
	mFullSpace    map[string]SuggestionWithMessage
	mFullNoSpace  map[string]SuggestionWithMessage
}

// LoadSimpleReplaceRule2Data fills maps from a replace.txt reader.
func (r *AbstractSimpleReplaceRule2) LoadSimpleReplaceRule2Data(reader io.Reader, path string) error {
	if r.mStartSpace == nil {
		r.mStartSpace = make(map[string]int)
		r.mStartNoSpace = make(map[string]int)
		r.mFullSpace = make(map[string]SuggestionWithMessage)
		r.mFullNoSpace = make(map[string]SuggestionWithMessage)
	}
	sc := bufio.NewScanner(reader)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || line[0] == '#' {
			continue
		}
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		if line == "" {
			continue
		}
		if r.CheckingCase {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) >= 1 {
				line = strings.ToLower(strings.TrimSpace(parts[0])) + "=" + strings.TrimSpace(parts[0])
				if len(parts) == 2 {
					line = line + "\t" + strings.TrimSpace(parts[1])
				}
			}
		}
		parts := strings.SplitN(line, "\t", 2)
		confPair := parts[0]
		msg := ""
		if len(parts) == 2 {
			msg = parts[1]
		}
		confPairParts := strings.SplitN(confPair, "=", 2)
		if len(confPairParts) < 2 {
			return fmt.Errorf("format error in %s: missing suggestion: %s", path, line)
		}
		suggestion := confPairParts[1]
		for _, wrongForm := range strings.Split(confPairParts[0], "|") {
			searchKey := wrongForm
			if r.CaseSens == CaseInsensitive {
				searchKey = strings.ToLower(wrongForm)
			}
			if !r.CheckingCase && searchKey == suggestion {
				return fmt.Errorf("format error in %s: same word both sides: %s", path, line)
			}
			swm := SuggestionWithMessage{Suggestion: suggestion, Message: msg}
			if !strings.Contains(wrongForm, " ") {
				runes := []rune(searchKey)
				if len(runes) == 0 {
					continue
				}
				firstChar := string(runes[0])
				if cur, ok := r.mStartNoSpace[firstChar]; !ok || cur < len(searchKey) {
					r.mStartNoSpace[firstChar] = len(searchKey)
				}
				r.mFullNoSpace[searchKey] = swm
			} else {
				toks := strings.Split(searchKey, " ")
				firstToken := toks[0]
				if cur, ok := r.mStartSpace[firstToken]; !ok || cur < len(toks) {
					r.mStartSpace[firstToken] = len(toks)
				}
				r.mFullSpace[searchKey] = swm
			}
		}
	}
	return sc.Err()
}

func (r *AbstractSimpleReplaceRule2) GetID() string {
	if r.ID != "" {
		return r.ID
	}
	return "SIMPLE_REPLACE_2"
}

// Match ports AbstractSimpleReplaceRule2.match.
func (r *AbstractSimpleReplaceRule2) Match(sentence *languagetool.AnalyzedSentence) []*RuleMatch {
	tokens := sentence.GetTokensWithoutWhitespace()
	var ruleMatches []*RuleMatch
	sentStart := 1
	for sentStart < len(tokens) && r.isPunctuationStart(tokens[sentStart].GetToken()) {
		sentStart++
	}
	checkCaseCoveredUpto := 0
	for startIndex := sentStart; startIndex < len(tokens); startIndex++ {
		if r.IsTokenException != nil && r.IsTokenException(tokens[startIndex]) {
			continue
		}
		tok := tokens[startIndex].GetToken()
		if len(tok) < 1 {
			continue
		}
		k := startIndex + 1
		for k < len(tokens) && !tokens[k].IsWhitespaceBefore() {
			tok = tok + tokens[k].GetToken()
			k++
		}
		lookupTok := tok
		if r.CaseSens == CaseInsensitive {
			lookupTok = strings.ToLower(tok)
		}
		if maxTokenLen, ok := r.mStartSpace[lookupTok]; ok {
			var keyBuilder strings.Builder
			endIndex := startIndex
			for endIndex < len(tokens) && endIndex-startIndex < maxTokensInMultiword {
				if endIndex > startIndex && tokens[endIndex].IsWhitespaceBefore() {
					keyBuilder.WriteByte(' ')
				}
				keyBuilder.WriteString(tokens[endIndex].GetToken())
				originalStr := keyBuilder.String()
				numberOfSpaces := strings.Count(originalStr, " ")
				if numberOfSpaces+1 > maxTokenLen {
					break
				}
				if numberOfSpaces > 0 {
					keyStr := originalStr
					if r.CaseSens == CaseInsensitive {
						keyStr = strings.ToLower(keyStr)
					}
					if swm, found := r.mFullSpace[keyStr]; found {
						r.createMatch(&ruleMatches, &swm, startIndex, endIndex, originalStr, tokens, sentence, sentStart, &checkCaseCoveredUpto)
					} else if sentStart == startIndex && r.CaseSens == CaseSensitiveExceptAtSentenceStart {
						lc := tools.LowercaseFirstChar(keyStr)
						if lc != keyStr {
							if swm2, ok2 := r.mFullSpace[lc]; ok2 {
								r.createMatch(&ruleMatches, &swm2, startIndex, endIndex, originalStr, tokens, sentence, sentStart, &checkCaseCoveredUpto)
							}
						}
					}
				}
				endIndex++
			}
		}
		if len(lookupTok) > 0 {
			first := string([]rune(lookupTok)[0])
			if _, ok := r.mStartNoSpace[first]; ok {
				endIndex := startIndex
				var keyBuilder strings.Builder
				for endIndex < len(tokens) && endIndex-startIndex < maxTokensInMultiword {
					if endIndex > startIndex && tokens[endIndex].IsWhitespaceBefore() {
						break
					}
					keyBuilder.WriteString(tokens[endIndex].GetToken())
					originalStr := keyBuilder.String()
					keyStr := originalStr
					if r.CaseSens == CaseInsensitive {
						keyStr = strings.ToLower(keyStr)
					}
					if swm, found := r.mFullNoSpace[keyStr]; found {
						r.createMatch(&ruleMatches, &swm, startIndex, endIndex, originalStr, tokens, sentence, sentStart, &checkCaseCoveredUpto)
					} else if sentStart == startIndex && r.CaseSens == CaseSensitiveExceptAtSentenceStart {
						lc := tools.LowercaseFirstChar(keyStr)
						if lc != keyStr {
							if swm2, ok2 := r.mFullNoSpace[lc]; ok2 {
								r.createMatch(&ruleMatches, &swm2, startIndex, endIndex, originalStr, tokens, sentence, sentStart, &checkCaseCoveredUpto)
							}
						}
					}
					endIndex++
				}
			}
		}
	}
	return ruleMatches
}

func (r *AbstractSimpleReplaceRule2) createMatch(
	ruleMatches *[]*RuleMatch,
	swm *SuggestionWithMessage,
	startIndex, endIndex int,
	originalStr string,
	tokens []*languagetool.AnalyzedTokenReadings,
	sentence *languagetool.AnalyzedSentence,
	sentStart int,
	checkCaseCoveredUpto *int,
) {
	if swm == nil {
		return
	}
	if r.IsException != nil && r.IsException(originalStr) {
		return
	}
	replacements := strings.Split(swm.Suggestion, "|")
	fromPos := tokens[startIndex].GetStartPos()
	toPos := tokens[endIndex].GetEndPos()
	if len(*ruleMatches) > 0 {
		last := (*ruleMatches)[len(*ruleMatches)-1]
		if last.GetFromPos() <= fromPos && last.GetToPos() >= toPos {
			return
		}
	}
	firstWordInSuggIsCamelCase := false
	for _, k := range replacements {
		parts := strings.SplitN(k, " ", 2)
		if tools.IsCamelCase(parts[0]) {
			firstWordInSuggIsCamelCase = true
			break
		}
	}
	isAllUppercase := tools.IsAllUppercase(originalStr)
	firstOrig := strings.SplitN(originalStr, " ", 2)[0]
	isCapitalized := tools.IsCapitalizedWord(firstOrig)

	if r.CheckingCase {
		if endIndex <= *checkCaseCoveredUpto {
			return
		}
		replacementCheckCase := replacements[0]
		if (sentStart == startIndex && originalStr == tools.UppercaseFirstChar(replacementCheckCase)) ||
			originalStr == replacementCheckCase {
			if len(*ruleMatches) > 0 {
				last := (*ruleMatches)[len(*ruleMatches)-1]
				if last.GetToPos() > fromPos {
					*ruleMatches = (*ruleMatches)[:len(*ruleMatches)-1]
				}
			}
			*checkCaseCoveredUpto = endIndex
			return
		}
		// Allow all-upper case, except for CamelCase suggestions and (by default) short words.
		if !firstWordInSuggIsCamelCase && originalStr == strings.ToUpper(originalStr) {
			const maxLengthShortWords = 4
			if !r.MatchShortAllUpperInCheckCase || len([]rune(originalStr)) > maxLengthShortWords {
				*checkCaseCoveredUpto = endIndex
				return
			}
		}
	}

	var finalReplacements []string
	for _, repl := range replacements {
		finalRepl := repl
		if !firstWordInSuggIsCamelCase && (sentStart == startIndex || (isCapitalized && !r.CheckingCase)) {
			finalRepl = tools.UppercaseFirstChar(repl)
		}
		if !r.CheckingCase && isAllUppercase {
			finalRepl = strings.ToUpper(repl)
		}
		if repl != originalStr && finalRepl != originalStr && !containsStringSlice(finalReplacements, finalRepl) {
			finalReplacements = append(finalReplacements, finalRepl)
		}
		if finalRepl == originalStr {
			finalReplacements = nil
			break
		}
	}
	if len(finalReplacements) == 0 {
		return
	}

	msg := swm.Message
	if msg == "" {
		var msgSuggestions strings.Builder
		for k, rep := range replacements {
			if k > 0 {
				if k == len(replacements)-1 {
					sep := r.SuggestionsSeparator
					if sep == "" {
						sep = ", "
					}
					msgSuggestions.WriteString(sep)
				} else {
					msgSuggestions.WriteString(", ")
				}
			}
			msgSuggestions.WriteString("<suggestion>")
			msgSuggestions.WriteString(rep)
			msgSuggestions.WriteString("</suggestion>")
		}
		tmpl := r.MessageTemplate
		if tmpl == "" {
			tmpl = "'$match' is incorrect."
		}
		msg = strings.Replace(tmpl, "$match", originalStr, 1)
		msg = strings.Replace(msg, "$suggestions", msgSuggestions.String(), 1)
	}

	ruleMatch := NewRuleMatch(r, sentence, fromPos, toPos, msg)
	ruleMatch.ShortMessage = r.ShortMsg
	ruleMatch.SetSuggestedReplacements(finalReplacements)
	if r.IsRuleMatchException != nil && r.IsRuleMatchException(ruleMatch) {
		return
	}
	if len(*ruleMatches) > 0 {
		last := (*ruleMatches)[len(*ruleMatches)-1]
		if last.GetFromPos() >= fromPos && last.GetToPos() <= toPos {
			*ruleMatches = (*ruleMatches)[:len(*ruleMatches)-1]
		}
	}
	*ruleMatches = append(*ruleMatches, ruleMatch)
}

func (r *AbstractSimpleReplaceRule2) isPunctuationStart(word string) bool {
	if hasDigitRune(word) {
		return true
	}
	return tools.IsPunctuationMark(word) || tools.IsNotWordCharacter(word)
}

func hasDigitRune(s string) bool {
	for _, r := range s {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func containsStringSlice(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}
