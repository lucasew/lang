package rules

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// LongSentenceRuleID ports LongSentenceRule.RULE_ID.
const LongSentenceRuleID = "TOO_LONG_SENTENCE"

var (
	// quotedSentEnd ports QUOTED_SENT_END = Pattern.compile("[?!.][\"“”„»«]", DOTALL).
	// DOTALL is a no-op for this character class (no '.' wildcard).
	quotedSentEnd = regexp.MustCompile(`[?!.]["“”„»«]`)
	// sentEndRE ports SENT_END = Pattern.compile("[?!.]") used with Matcher.matches()
	// (entire token must be a single sentence-end mark).
	sentEndRE     = regexp.MustCompile(`^[?!.]$`)
	openingQuotes = []string{"\"", "“", "„", "«", "(", "[", "{", "—"}
	closingQuotes = []string{"\"", "”", "“", "»", ")", "]", "}", "—"}
)

// LongSentenceRule ports org.languagetool.rules.LongSentenceRule.
// Java: Categories.STYLE, ITSIssueType.Style, Tag.picky; minToCheckParagraph=0.
type LongSentenceRule struct {
	Messages map[string]string
	MaxWords int
	// RuleID overrides GetID when set (e.g. TOO_LONG_SENTENCE_DE).
	RuleID string
	// Description overrides GetDescription when set (language modules).
	Description string
	// ShortMsg optional short message for language wrappers.
	ShortMsg string
	// Category ports Rule.category (Java STYLE).
	Category *Category
	// IssueType ports getLocQualityIssueType (Java Style).
	IssueType ITSIssueType
	// Tags ports Rule.tags (Java picky).
	Tags []Tag
	// URL ports Rule.url (Java setUrl; EN subclass sets splitting-long-sentences).
	URL string
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []IncorrectExample
	correctExamples   []CorrectExample
}

// NewLongSentenceRule ports LongSentenceRule(messages, null, maxWords).
func NewLongSentenceRule(messages map[string]string, maxWords int) *LongSentenceRule {
	return NewLongSentenceRuleWithUserConfig(messages, nil, maxWords)
}

// NewLongSentenceRuleWithUserConfig ports
// LongSentenceRule(ResourceBundle messages, UserConfig userConfig, int maxWords).
// UserConfig.getConfigValueByID(getId()) overrides maxWords when present.
func NewLongSentenceRuleWithUserConfig(messages map[string]string, userConfig *languagetool.UserConfig, maxWords int) *LongSentenceRule {
	tmpMaxWords := maxWords
	if userConfig != nil {
		if cf := userConfig.GetConfigValueByID(LongSentenceRuleID); len(cf) > 0 && cf[0] != nil {
			if n, ok := configValueAsInt(cf[0]); ok {
				tmpMaxWords = n
			}
		}
	}
	return &LongSentenceRule{
		Messages:  messages,
		MaxWords:  tmpMaxWords,
		Category:  CatStyle.GetCategory(messages),
		IssueType: ITSStyle,
		Tags:      []Tag{TagPicky},
	}
}

func (r *LongSentenceRule) GetID() string {
	if r != nil && r.RuleID != "" {
		return r.RuleID
	}
	return LongSentenceRuleID
}

func (r *LongSentenceRule) GetCategory() *Category {
	if r == nil {
		return nil
	}
	return r.Category
}

func (r *LongSentenceRule) GetLocQualityIssueType() ITSIssueType {
	if r == nil || r.IssueType == "" {
		return ITSStyle
	}
	return r.IssueType
}

func (r *LongSentenceRule) GetTags() []Tag {
	if r == nil {
		return nil
	}
	return r.Tags
}

func (r *LongSentenceRule) HasTag(tag Tag) bool {
	if r == nil {
		return false
	}
	for _, t := range r.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// GetURL ports Rule.getUrl.
func (r *LongSentenceRule) GetURL() string {
	if r == nil {
		return ""
	}
	return r.URL
}

// SetURL ports Rule.setUrl.
func (r *LongSentenceRule) SetURL(u string) {
	if r != nil {
		r.URL = u
	}
}

// AddExamplePair ports Rule.addExamplePair.
func (r *LongSentenceRule) AddExamplePair(incorrect IncorrectExample, correct CorrectExample) {
	if r == nil {
		return
	}
	appendExamplePair(&r.incorrectExamples, &r.correctExamples, incorrect, correct)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *LongSentenceRule) GetIncorrectExamples() []IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *LongSentenceRule) GetCorrectExamples() []CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

// EstimateContextForSureMatch ports TextLevelRule (Java always -1).
func (r *LongSentenceRule) EstimateContextForSureMatch() int { return -1 }

// MinToCheckParagraph ports LongSentenceRule.minToCheckParagraph (Java returns 0).
func (r *LongSentenceRule) MinToCheckParagraph() int { return 0 }

// GetDescription ports LongSentenceRule.getDescription.
// Java: MessageFormat.format(messages.getString("long_sentence_rule_desc"), maxWords).
func (r *LongSentenceRule) GetDescription() string {
	if r != nil && r.Description != "" {
		return r.Description
	}
	if r != nil && r.Messages != nil {
		if tmpl := r.Messages["long_sentence_rule_desc"]; tmpl != "" {
			return messageFormat0(tmpl, r.maxWords())
		}
	}
	return fmt.Sprintf("Finds long sentences (more than %d words)", r.maxWords())
}

// maxWords returns the configured threshold (Java private final maxWords field).
func (r *LongSentenceRule) maxWords() int {
	if r == nil {
		return 0
	}
	return r.MaxWords
}

// GetMessage ports LongSentenceRule.getMessage.
// Java: MessageFormat.format(messages.getString("long_sentence_rule_msg2"), maxWords).
func (r *LongSentenceRule) GetMessage() string {
	msg := ""
	if r != nil && r.Messages != nil {
		msg = r.Messages["long_sentence_rule_msg2"]
	}
	if msg == "" {
		msg = "This sentence is too long ({0} words)"
	}
	return messageFormat0(msg, r.maxWords())
}

// GetRuleOptions ports LongSentenceRule.getRuleOptions.
// Java: new RuleOption(maxWords, messages.getString("guiLongSentencesText"), 5, 100).
func (r *LongSentenceRule) GetRuleOptions() []*RuleOption {
	cfgText := ""
	if r != nil && r.Messages != nil {
		cfgText = r.Messages["guiLongSentencesText"]
	}
	if cfgText == "" {
		cfgText = "Maximum sentences length (5 - 100 words):"
	}
	return []*RuleOption{
		NewRuleOption(r.maxWords(), cfgText, 5, 100),
	}
}

// messageFormat0 ports MessageFormat with a single {0} integer argument.
func messageFormat0(tmpl string, n int) string {
	s := strings.ReplaceAll(tmpl, "{0}", "%d")
	if strings.Contains(s, "%d") {
		return fmt.Sprintf(s, n)
	}
	return fmt.Sprintf("%s (%d)", s, n)
}

func isWordCount(tokenText string) bool {
	// Java: if (tokenText.length() > 0) return !StringTools.isNotWordCharacter(tokenText.substring(0,1));
	// length/substring are UTF-16 code units (BMP letter = 1 unit; emoji high surrogate is not \p{L}).
	if utf16Len(tokenText) == 0 {
		return false
	}
	first := utf16Substring(tokenText, 0, 1)
	return !tools.IsNotWordCharacter(first)
}

// MatchList ports match(List<AnalyzedSentence>).
func (r *LongSentenceRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*RuleMatch {
	var ruleMatches []*RuleMatch
	pos := 0
	maxWords := r.maxWords()
	// Java builds msg once per sentence after short-circuits; message is independent of sentence.
	msg := r.GetMessage()
	for _, sentence := range sentences {
		tokens := sentence.GetTokens()
		// Java: if (tokens.length < maxWords) { pos += …; continue; } — short-circuit on token count.
		if len(tokens) < maxWords {
			pos += sentence.GetCorrectedTextLength()
			continue
		}
		if quotedSentEnd.MatchString(sentence.GetText()) {
			pos += sentence.GetCorrectedTextLength()
			continue
		}
		// Java: fromPosToken / toPosToken / indexOfQuote are declared outside the outer
		// segment loop — not reset at ":" / ";" / newline boundaries. A match after a
		// colon still spans from the first counted word of the sentence (bug-for-bug).
		i := 0
		var fromPos, toPos []int
		var fromPosToken, toPosToken *languagetool.AnalyzedTokenReadings
		indexOfQuote := -1
		for i < len(tokens) {
			numWords := 0
			for i < len(tokens) && tokens[i].GetToken() != ":" && tokens[i].GetToken() != ";" &&
				tokens[i].GetToken() != "\n" && tokens[i].GetToken() != "\r\n" && tokens[i].GetToken() != "\n\r" {
				token := tokens[i].GetToken()
				// Java: OPENING_QUOTES.contains then indexOf; CLOSING_QUOTES.indexOf(token) == indexOfQuote.
				if indexOfQuote == -1 {
					for oi, oq := range openingQuotes {
						if token == oq {
							indexOfQuote = oi
							break
						}
					}
				} else if indexOfQuote > -1 {
					// Equivalent to CLOSING_QUOTES.indexOf(token) == indexOfQuote (exact slot match).
					if indexOfQuote < len(closingQuotes) && token == closingQuotes[indexOfQuote] {
						indexOfQuote = -1
					}
				}
				if isWordCount(token) && indexOfQuote == -1 {
					if fromPosToken == nil {
						fromPosToken = tokens[i]
					}
					if numWords == maxWords {
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
							// keep old logic if we could not find word tokens
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
