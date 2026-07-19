package rules

import "github.com/lucasew/lang/internal/languagetool/org/languagetool"

// RuleMatchType ports org.languagetool.rules.RuleMatch.Type (underline category).
type RuleMatchType string

const (
	// RuleMatchTypeUnknownWord is spelling errors (typically red).
	RuleMatchTypeUnknownWord RuleMatchType = "UnknownWord"
	// RuleMatchTypeHint is style errors (typically light blue).
	RuleMatchTypeHint RuleMatchType = "Hint"
	// RuleMatchTypeOther is grammar/other (typically yellow/orange). Default.
	RuleMatchTypeOther RuleMatchType = "Other"
)

// RuleMatch ports org.languagetool.rules.RuleMatch (fields needed by unit tests).
type RuleMatch struct {
	Rule     any
	Sentence *languagetool.AnalyzedSentence
	FromPos  int
	ToPos    int
	// FromPosSentence/ToPosSentence port Java sentencePosition (-1 = unset).
	FromPosSentence int
	ToPosSentence   int
	Message         string
	ShortMessage    string
	// Type ports RuleMatch.type (default Other; spelling uses UnknownWord).
	Type                  RuleMatchType
	SuggestedReplacements []string
	// SuggestedReplacementObjects ports RuleMatch.suggestedReplacements objects
	// (confidence, type, …). When set, string list mirrors GetReplacement().
	SuggestedReplacementObjects []*SuggestedReplacement
	// OriginalErrorStr ports RuleMatch.originalErrorStr (inmarker / templates).
	OriginalErrorStr string
	// URL optional match-level link (overrides rule URL when set).
	URL string
	// SpecificRuleId ports RuleMatch.specificRuleId (setSpecificRuleId / getSpecificRuleId).
	// Empty means getSpecificRuleId falls back to getRule().getId().
	SpecificRuleId string
	// Soft metadata carried from LocalMatch (false friends / soft grammar).
	IssueType    string
	CategoryID   string
	CategoryName string
}

func NewRuleMatch(rule any, sentence *languagetool.AnalyzedSentence, fromPos, toPos int, message string) *RuleMatch {
	return &RuleMatch{
		Rule:            rule,
		Sentence:        sentence,
		FromPos:         fromPos,
		ToPos:           toPos,
		FromPosSentence: -1,
		ToPosSentence:   -1,
		Message:         message,
		Type:            RuleMatchTypeOther,
	}
}

// GetType ports RuleMatch.getType.
func (m *RuleMatch) GetType() RuleMatchType {
	if m == nil || m.Type == "" {
		return RuleMatchTypeOther
	}
	return m.Type
}

// SetType ports RuleMatch.setType.
func (m *RuleMatch) SetType(t RuleMatchType) {
	if m != nil {
		m.Type = t
	}
}

// SetShortMessage ports RuleMatch short description (constructor 6-arg form).
func (m *RuleMatch) SetShortMessage(s string) {
	if m != nil {
		m.ShortMessage = s
	}
}

func (m *RuleMatch) GetFromPos() int { return m.FromPos }
func (m *RuleMatch) GetToPos() int   { return m.ToPos }
func (m *RuleMatch) SetSuggestedReplacement(s string) {
	m.SuggestedReplacements = []string{s}
}
func (m *RuleMatch) GetSuggestedReplacements() []string { return m.SuggestedReplacements }

func (m *RuleMatch) GetRule() any       { return m.Rule }
func (m *RuleMatch) GetMessage() string { return m.Message }

// SetSpecificRuleId ports RuleMatch.setSpecificRuleId.
func (m *RuleMatch) SetSpecificRuleId(id string) {
	if m != nil {
		m.SpecificRuleId = id
	}
}

// GetSpecificRuleId ports RuleMatch.getSpecificRuleId (empty → rule.GetID when available).
func (m *RuleMatch) GetSpecificRuleId() string {
	if m == nil {
		return ""
	}
	if m.SpecificRuleId != "" {
		return m.SpecificRuleId
	}
	if m.Rule == nil {
		return ""
	}
	if g, ok := m.Rule.(interface{ GetID() string }); ok {
		return g.GetID()
	}
	return ""
}
func (m *RuleMatch) GetShortMessage() string {
	if m == nil {
		return ""
	}
	return m.ShortMessage
}

func (m *RuleMatch) SetOffsetPosition(from, to int) {
	m.FromPos = from
	m.ToPos = to
}

// GetFromPosSentence / GetToPosSentence / SetSentencePosition port Java RuleMatch sentence positions.
func (m *RuleMatch) GetFromPosSentence() int {
	if m == nil {
		return -1
	}
	return m.FromPosSentence
}
func (m *RuleMatch) GetToPosSentence() int {
	if m == nil {
		return -1
	}
	return m.ToPosSentence
}
func (m *RuleMatch) SetSentencePosition(from, to int) {
	if m == nil {
		return
	}
	m.FromPosSentence = from
	m.ToPosSentence = to
}

func (m *RuleMatch) SetSuggestedReplacements(reps []string) {
	m.SuggestedReplacements = append([]string(nil), reps...)
	// keep objects in sync when only strings are provided
	m.SuggestedReplacementObjects = ConvertSuggestions(reps)
}

// SetSuggestedReplacementObjects ports RuleMatch.setSuggestedReplacementObjects.
func (m *RuleMatch) SetSuggestedReplacementObjects(objs []*SuggestedReplacement) {
	if m == nil {
		return
	}
	m.SuggestedReplacementObjects = append([]*SuggestedReplacement(nil), objs...)
	m.SuggestedReplacements = m.SuggestedReplacements[:0]
	for _, o := range objs {
		if o == nil {
			continue
		}
		m.SuggestedReplacements = append(m.SuggestedReplacements, o.GetReplacement())
	}
}

// GetSuggestedReplacementObjects ports RuleMatch.getSuggestedReplacementObjects.
func (m *RuleMatch) GetSuggestedReplacementObjects() []*SuggestedReplacement {
	if m == nil {
		return nil
	}
	return m.SuggestedReplacementObjects
}

// SetOriginalErrorStr ports RuleMatch.setOriginalErrorStr from sentence span.
// Prefers FromPosSentence/ToPosSentence when set (Java); falls back to FromPos/ToPos
// when sentence positions are unset (common for hand-built matches in this port).
// Positions are Java UTF-16 code units (same as AnalyzedTokenReadings start/end).
func (m *RuleMatch) SetOriginalErrorStr() {
	if m == nil {
		return
	}
	if m.OriginalErrorStr != "" {
		return
	}
	if m.Sentence == nil {
		return
	}
	text := m.Sentence.GetText()
	if text == "" {
		return
	}
	from, to := m.FromPosSentence, m.ToPosSentence
	if from < 0 || to < 0 {
		from, to = m.FromPos, m.ToPos
	}
	if from < 0 || to < 0 || from >= to {
		return
	}
	// Java String.substring uses UTF-16 indices — not Go byte offsets.
	if to > utf16Len(text) {
		return
	}
	m.OriginalErrorStr = utf16Substring(text, from, to)
}

// GetOriginalErrorStr ports RuleMatch.getOriginalErrorStr.
func (m *RuleMatch) GetOriginalErrorStr() string {
	if m == nil {
		return ""
	}
	return m.OriginalErrorStr
}

func (m *RuleMatch) SetURL(u string) {
	if m != nil {
		m.URL = u
	}
}

func (m *RuleMatch) GetURL() string {
	if m == nil {
		return ""
	}
	return m.URL
}

// IsUnderlinedErrorSingleToken ports RuleMatch.isUnderlinedErrorSingleToken:
// true when the underlined span covers exactly one non-whitespace token.
func (m *RuleMatch) IsUnderlinedErrorSingleToken() bool {
	if m == nil || m.Sentence == nil {
		return false
	}
	tokens := m.Sentence.GetTokensWithoutWhitespace()
	fromIdx, toIdx := -1, -1
	for i, tok := range tokens {
		if tok == nil || tok.IsSentenceStart() {
			continue
		}
		// token overlaps [FromPos, ToPos)
		if tok.GetEndPos() > m.FromPos && tok.GetStartPos() < m.ToPos {
			if fromIdx < 0 {
				fromIdx = i
			}
			toIdx = i
		}
	}
	if fromIdx < 0 {
		return false
	}
	return fromIdx == toIdx
}
