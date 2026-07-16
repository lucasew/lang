package rules

import (
	"fmt"
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

var (
	quotedSentEnd = regexp.MustCompile(`[?!.]["“”„»«]`)
	sentEndRE     = regexp.MustCompile(`^[?!.]$`)
	openingQuotes = []string{"\"", "“", "„", "«", "(", "[", "{", "—"}
	closingQuotes = []string{"\"", "”", "“", "»", ")", "]", "}", "—"}
)

// LongSentenceRule ports org.languagetool.rules.LongSentenceRule.
type LongSentenceRule struct {
	Messages map[string]string
	MaxWords int
}

func NewLongSentenceRule(messages map[string]string, maxWords int) *LongSentenceRule {
	return &LongSentenceRule{Messages: messages, MaxWords: maxWords}
}

func (r *LongSentenceRule) GetID() string { return "TOO_LONG_SENTENCE" }

func (r *LongSentenceRule) GetMessage() string {
	msg := r.Messages["long_sentence_rule_msg2"]
	if msg == "" {
		msg = "This sentence is too long (%d words)"
	}
	// MessageFormat with one number
	if containsPercentD(msg) {
		return fmt.Sprintf(msg, r.MaxWords)
	}
	return fmt.Sprintf(msg, r.MaxWords)
}

func containsPercentD(s string) bool {
	return regexp.MustCompile(`\{\d+\}|%d`).MatchString(s)
}

func isWordCount(tokenText string) bool {
	if len(tokenText) == 0 {
		return false
	}
	// first char as string for isNotWordCharacter
	first := string([]rune(tokenText)[0])
	return !tools.IsNotWordCharacter(first)
}

// MatchList ports match(List<AnalyzedSentence>).
func (r *LongSentenceRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*RuleMatch {
	var ruleMatches []*RuleMatch
	pos := 0
	msg := r.GetMessage()
	for _, sentence := range sentences {
		tokens := sentence.GetTokens()
		if len(tokens) < r.MaxWords {
			pos += sentence.GetCorrectedTextLength()
			continue
		}
		if quotedSentEnd.MatchString(sentence.GetText()) {
			pos += sentence.GetCorrectedTextLength()
			continue
		}
		i := 0
		var fromPos, toPos []int
		var fromPosToken, toPosToken *languagetool.AnalyzedTokenReadings
		indexOfQuote := -1
		for i < len(tokens) {
			numWords := 0
			fromPosToken = nil
			toPosToken = nil
			for i < len(tokens) && tokens[i].GetToken() != ":" && tokens[i].GetToken() != ";" &&
				tokens[i].GetToken() != "\n" && tokens[i].GetToken() != "\r\n" && tokens[i].GetToken() != "\n\r" {
				token := tokens[i].GetToken()
				if indexOfQuote == -1 {
					for oi, oq := range openingQuotes {
						if token == oq {
							indexOfQuote = oi
							break
						}
					}
				} else if indexOfQuote > -1 {
					if indexOfQuote < len(closingQuotes) && token == closingQuotes[indexOfQuote] {
						indexOfQuote = -1
					}
				}
				if isWordCount(token) && indexOfQuote == -1 {
					if fromPosToken == nil {
						fromPosToken = tokens[i]
					}
					if numWords == r.MaxWords {
						if toPosToken == nil {
							for j := len(tokens) - 1; j >= 0; j-- {
								if isWordCount(tokens[j].GetToken()) {
									if j+1 < len(tokens) && sentEndRE.MatchString(tokens[j+1].GetToken()) {
										toPosToken = tokens[j+1]
									} else {
										toPosToken = tokens[j]
									}
									break
								}
							}
						}
						if fromPosToken != nil && toPosToken != nil {
							fromPos = append(fromPos, fromPosToken.GetStartPos())
							toPos = append(toPos, toPosToken.GetEndPos())
						} else {
							fromPos = append(fromPos, tokens[0].GetStartPos())
							toPos = append(toPos, tokens[len(tokens)-1].GetEndPos())
						}
						break
					}
					numWords++
				}
				i++
			}
			i++
		}
		for j := 0; j < len(fromPos); j++ {
			rm := NewRuleMatch(r, sentence, pos+fromPos[j], pos+toPos[j], msg)
			ruleMatches = append(ruleMatches, rm)
		}
		pos += sentence.GetCorrectedTextLength()
	}
	return ruleMatches
}
