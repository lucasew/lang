package rules

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// LongParagraphRule ports org.languagetool.rules.LongParagraphRule.
// Java: STYLE, Style, setDefaultOff(), setOfficeDefaultOn(), Tag.picky; minToCheckParagraph=0.
// Note: Java 5-arg defaultActive is unused — always setDefaultOff().
type LongParagraphRule struct {
	Messages map[string]string
	MaxWords int
	// SingleLineBreaksMarksPara ports Language.getSentenceTokenizer().singleLineBreaksMarksPara()
	// as used by Tools.isParagraphEnd (Demo/SRX default false → need \n\n).
	SingleLineBreaksMarksPara bool
	Category                  *Category
	IssueType                 ITSIssueType
	DefaultOff                bool
	// OfficeDefaultOn ports Rule.setOfficeDefaultOn (Java always On for LO/OO).
	OfficeDefaultOn bool
	// Tags ports Rule.tags (Java picky).
	Tags []Tag
}

// Java LongParagraphRule constants.
const (
	LongParagraphRuleID          = "TOO_LONG_PARAGRAPH"
	longParagraphDefaultMaxWords = 220
	longParagraphLimitMaxWords   = 300
)

// leadingNewlinesRE ports Java sentence.getText().replaceFirst("^\n+", "").
var leadingNewlinesRE = regexp.MustCompile(`^\n+`)

// NewLongParagraphRule ports LongParagraphRule(messages, lang, null, defaultWords).
// defaultWords <= 0 keeps Java DEFAULT_MAX_WORDS (220).
func NewLongParagraphRule(messages map[string]string, defaultWords int) *LongParagraphRule {
	return NewLongParagraphRuleWithUserConfig(messages, nil, defaultWords)
}

// NewLongParagraphRuleWithUserConfig ports
// LongParagraphRule(ResourceBundle, Language, UserConfig, int defaultWords[, boolean]).
// Language is represented by SingleLineBreaksMarksPara (set by callers when needed).
func NewLongParagraphRuleWithUserConfig(messages map[string]string, userConfig *languagetool.UserConfig, defaultWords int) *LongParagraphRule {
	maxWords := longParagraphDefaultMaxWords
	if defaultWords > 0 {
		maxWords = defaultWords
	}
	if userConfig != nil {
		if cf := userConfig.GetConfigValueByID(LongParagraphRuleID); len(cf) > 0 && cf[0] != nil {
			if n, ok := configValueAsInt(cf[0]); ok {
				maxWords = n
			}
		}
	}
	return &LongParagraphRule{
		Messages:        messages,
		MaxWords:        maxWords,
		Category:        CatStyle.GetCategory(messages),
		IssueType:       ITSStyle,
		DefaultOff:      true, // Java always setDefaultOff() (defaultActive param unused)
		OfficeDefaultOn: true,
		Tags:            []Tag{TagPicky},
	}
}

// configValueAsInt ports Java (int) cf[0] for UserConfig values.
func configValueAsInt(v any) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int32:
		return int(n), true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	default:
		return 0, false
	}
}

func (r *LongParagraphRule) GetID() string { return LongParagraphRuleID }

// GetDescription ports MessageFormat(long_paragraph_rule_desc, maxWords).
func (r *LongParagraphRule) GetDescription() string {
	if r != nil && r.Messages != nil {
		if s := r.Messages["long_paragraph_rule_desc"]; s != "" {
			return messageFormat0(s, r.maxWords())
		}
	}
	return fmt.Sprintf("Paragraph is too long (more than %d words)", r.maxWords())
}

func (r *LongParagraphRule) maxWords() int {
	if r == nil || r.MaxWords <= 0 {
		return longParagraphDefaultMaxWords
	}
	return r.MaxWords
}

func (r *LongParagraphRule) GetCategory() *Category {
	if r == nil {
		return nil
	}
	return r.Category
}

func (r *LongParagraphRule) GetLocQualityIssueType() ITSIssueType {
	if r == nil || r.IssueType == "" {
		return ITSStyle
	}
	return r.IssueType
}

func (r *LongParagraphRule) IsDefaultOff() bool { return r != nil && r.DefaultOff }

// IsOfficeDefaultOn ports Rule.isOfficeDefaultOn.
func (r *LongParagraphRule) IsOfficeDefaultOn() bool { return r != nil && r.OfficeDefaultOn }

// GetTags ports Rule.getTags (Java picky).
func (r *LongParagraphRule) GetTags() []Tag {
	if r == nil || len(r.Tags) == 0 {
		return nil
	}
	return append([]Tag(nil), r.Tags...)
}

// HasTag ports Rule.hasTag.
func (r *LongParagraphRule) HasTag(tag Tag) bool {
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

// GetRuleOptions ports LongParagraphRule.getRuleOptions.
func (r *LongParagraphRule) GetRuleOptions() []*RuleOption {
	mw := r.maxWords()
	cfgText := ""
	if r != nil && r.Messages != nil {
		if s := r.Messages["guiLongParagraphsText"]; s != "" {
			cfgText = messageFormat0(s, longParagraphLimitMaxWords)
		}
	}
	if cfgText == "" {
		cfgText = messageFormat0("Maximum paragraphs length (5 - {0} words):", longParagraphLimitMaxWords)
	}
	return []*RuleOption{
		NewRuleOption(mw, cfgText, 5, longParagraphLimitMaxWords),
	}
}

// MinToCheckParagraph ports LongParagraphRule.minToCheckParagraph (Java returns 0).
func (r *LongParagraphRule) MinToCheckParagraph() int { return 0 }

// GetMessage ports LongParagraphRule.getMessage.
func (r *LongParagraphRule) GetMessage() string {
	msg := ""
	if r != nil && r.Messages != nil {
		msg = r.Messages["long_paragraph_rule_msg"]
	}
	if msg == "" {
		msg = "This paragraph is too long ({0} words)"
	}
	return messageFormat0(msg, r.maxWords())
}

func (r *LongParagraphRule) isParagraphEnd(sentences []*languagetool.AnalyzedSentence, nTest int) bool {
	return languagetool.IsParagraphEnd(sentences, nTest, r.SingleLineBreaksMarksPara)
}

// MatchList ports match(List<AnalyzedSentence>).
func (r *LongParagraphRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*RuleMatch {
	var ruleMatches []*RuleMatch
	pos := 0
	startPos, endPos := 0, 0
	wordCount := 0
	paraHasLinebreaks := false
	maxWords := r.maxWords()
	for n := 0; n < len(sentences); n++ {
		sentence := sentences[n]
		paragraphEnd := r.isParagraphEnd(sentences, n)
		// Java: !paragraphEnd && sentence.getText().replaceFirst("^\n+", "").contains("\n")
		trimmed := leadingNewlinesRE.ReplaceAllString(sentence.GetText(), "")
		if !paragraphEnd && strings.Contains(trimmed, "\n") {
			// e.g. text with manually added line breaks (e.g. issues on github with "- [ ]" syntax)
			paraHasLinebreaks = true
		}
		tokens := sentence.GetTokensWithoutWhitespace()
		for _, token := range tokens {
			// Java: !token.isWhitespace() && !token.isSentenceStart() && !token.isNonWord()
			if !token.IsWhitespace() && !token.IsSentenceStart() && !token.IsNonWord() {
				wordCount++
				if wordCount == maxWords {
					endPos = token.GetEndPos() + pos
				} else if wordCount == maxWords-1 {
					startPos = token.GetStartPos() + pos
				}
			}
		}
		if paragraphEnd {
			// + 5: don't show match almost at end of paragraph
			if wordCount > maxWords+5 && !paraHasLinebreaks {
				rm := NewRuleMatch(r, sentence, startPos, endPos, r.GetMessage())
				ruleMatches = append(ruleMatches, rm)
			}
			wordCount = 0
			paraHasLinebreaks = false
		}
		pos += sentence.GetCorrectedTextLength()
	}
	// Java dead path in practice (last sentence is always paragraphEnd) — keep bug-for-bug.
	if wordCount > maxWords {
		var last *languagetool.AnalyzedSentence
		if len(sentences) > 0 {
			last = sentences[len(sentences)-1]
		}
		// Java: new RuleMatch(this, startPos, endPos, getMessage()) — no sentence arg
		rm := NewRuleMatch(r, last, startPos, endPos, r.GetMessage())
		ruleMatches = append(ruleMatches, rm)
	}
	return ruleMatches
}
