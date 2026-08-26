package cache

import (
	"encoding/json"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools/grpc"
)

// CachedRule is a protobuf-free stand-in for ProtoResultCache.CachedRule.
type CachedRule struct {
	ID                          string   `json:"id"`
	SubID                       string   `json:"subId,omitempty"`
	Description                 string   `json:"description,omitempty"`
	EstimateContextForSureMatch int      `json:"estimateContextForSureMatch,omitempty"`
	SourceFile                  string   `json:"sourceFile,omitempty"`
	IssueType                   string   `json:"issueType,omitempty"`
	TempOff                     bool     `json:"tempOff,omitempty"`
	IsPremium                   bool     `json:"isPremium,omitempty"`
	CategoryID                  string   `json:"categoryId,omitempty"`
	CategoryName                string   `json:"categoryName,omitempty"`
	Tags                        []string `json:"tags,omitempty"`
}

// CachedMatchPosition ports MatchPosition in the cache proto.
type CachedMatchPosition struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

// CachedResultMatch is a protobuf-free CachedResultMatch.
type CachedResultMatch struct {
	Rule                  CachedRule          `json:"rule"`
	Message               string              `json:"message"`
	ShortMessage          string              `json:"shortMessage,omitempty"`
	OffsetPosition        CachedMatchPosition `json:"offsetPosition"`
	SuggestedReplacements []string            `json:"suggestedReplacements,omitempty"`
	URL                   string              `json:"url,omitempty"`
	Type                  string              `json:"type,omitempty"`
	AutoCorrect           bool                `json:"autoCorrect,omitempty"`
	SpecificRuleID        string              `json:"specificRuleId,omitempty"`
}

// SerializeResultMatch ports CacheUtils.serializeResultMatch (JSON/struct form).
func SerializeResultMatch(m *rules.RuleMatch) CachedResultMatch {
	if m == nil {
		return CachedResultMatch{}
	}
	id, sub, desc := "UNKNOWN", "", ""
	if r, ok := m.Rule.(interface{ GetID() string }); ok {
		id = r.GetID()
	}
	if r, ok := m.Rule.(interface{ GetSubID() string }); ok {
		sub = r.GetSubID()
	}
	if r, ok := m.Rule.(interface{ GetDescription() string }); ok {
		desc = r.GetDescription()
	}
	return CachedResultMatch{
		Rule: CachedRule{
			ID:          id,
			SubID:       sub,
			Description: desc,
		},
		Message:               m.Message,
		ShortMessage:          m.ShortMessage,
		OffsetPosition:        CachedMatchPosition{Start: m.FromPos, End: m.ToPos},
		SuggestedReplacements: append([]string(nil), m.SuggestedReplacements...),
		URL:                   grpc.CoalesceURL("", ""),
		SpecificRuleID:        id,
	}
}

// DeserializeResultMatch rebuilds a RuleMatch from a cached DTO.
func DeserializeResultMatch(c CachedResultMatch, sentence *languagetool.AnalyzedSentence) *rules.RuleMatch {
	rule := rules.NewFakeRule(c.Rule.ID)
	m := rules.NewRuleMatch(rule, sentence, c.OffsetPosition.Start, c.OffsetPosition.End, c.Message)
	m.ShortMessage = c.ShortMessage
	if len(c.SuggestedReplacements) > 0 {
		m.SetSuggestedReplacements(c.SuggestedReplacements)
	}
	return m
}

// MarshalResultMatchJSON encodes a match for cache storage.
func MarshalResultMatchJSON(m *rules.RuleMatch) ([]byte, error) {
	return json.Marshal(SerializeResultMatch(m))
}

// UnmarshalResultMatchJSON decodes a cached match.
func UnmarshalResultMatchJSON(data []byte, sentence *languagetool.AnalyzedSentence) (*rules.RuleMatch, error) {
	var c CachedResultMatch
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return DeserializeResultMatch(c, sentence), nil
}
