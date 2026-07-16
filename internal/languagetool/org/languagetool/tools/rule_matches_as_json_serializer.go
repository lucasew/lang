package tools

import (
	"encoding/json"
	"strings"
)

const (
	jsonAPIVersion = 1
	jsonStatus     = ""
	premiumHint    = "You might be missing errors only the Premium version can find. Contact us at support<at>languagetoolplus.com."
)

// MatchForJSON is the minimal match surface used by RuleMatchesAsJsonSerializer
// (avoids importing the rules package and creating an import cycle).
type MatchForJSON struct {
	Message               string
	ShortMessage          string
	FromPos               int
	ToPos                 int
	SuggestedReplacements []string
	RuleID                string
	RuleDescription       string
}

// RuleMatchesAsJsonSerializer ports org.languagetool.tools.RuleMatchesAsJsonSerializer
// as a compact JSON encoder for match lists.
type RuleMatchesAsJsonSerializer struct {
	CompactMode         int
	LanguageCode        string
	LanguageName        string
	DetectedCode        string
	DetectedName        string
	DetectionConfidence float64
	DetectionSource     string
	Premium             bool
}

func NewRuleMatchesAsJsonSerializer() *RuleMatchesAsJsonSerializer {
	return &RuleMatchesAsJsonSerializer{
		LanguageCode: "en",
		LanguageName: "English",
	}
}

// MatchJSON is the per-match JSON shape (subset of LT API).
type MatchJSON struct {
	Message      string            `json:"message"`
	ShortMessage string            `json:"shortMessage,omitempty"`
	Offset       int               `json:"offset"`
	Length       int               `json:"length"`
	Replacements []ReplacementJSON `json:"replacements"`
	Rule         RuleJSON          `json:"rule"`
	Context      *ContextJSON      `json:"context,omitempty"`
}

type ReplacementJSON struct {
	Value string `json:"value"`
}

type RuleJSON struct {
	ID          string `json:"id"`
	Description string `json:"description,omitempty"`
}

type ContextJSON struct {
	Text   string `json:"text"`
	Offset int    `json:"offset"`
	Length int    `json:"length"`
}

// ResponseJSON is the top-level object.
type ResponseJSON struct {
	Software map[string]any `json:"software,omitempty"`
	Warnings map[string]any `json:"warnings,omitempty"`
	Language map[string]any `json:"language,omitempty"`
	Matches  []MatchJSON    `json:"matches"`
}

// RuleMatchesToJSON serializes matches for plain text.
func (s *RuleMatchesAsJsonSerializer) RuleMatchesToJSON(matches []MatchForJSON, text string, contextSize int) (string, error) {
	return s.RuleMatchesToJSONWithReason(matches, text, contextSize, "", true)
}

// RuleMatchesToJSONWithReason adds incompleteResults warning when reason is set.
func (s *RuleMatchesAsJsonSerializer) RuleMatchesToJSONWithReason(matches []MatchForJSON, text string, contextSize int, incompleteReason string, showPremiumHint bool) (string, error) {
	resp := ResponseJSON{Matches: make([]MatchJSON, 0, len(matches))}
	if s.CompactMode != 1 {
		soft := map[string]any{
			"name":       "LanguageTool",
			"apiVersion": jsonAPIVersion,
			"premium":    s.Premium,
			"status":     jsonStatus,
		}
		if showPremiumHint {
			soft["premiumHint"] = premiumHint
		}
		resp.Software = soft
	}
	if s.CompactMode != 1 || incompleteReason != "" {
		w := map[string]any{"incompleteResults": incompleteReason != ""}
		if incompleteReason != "" {
			w["incompleteResultsReason"] = incompleteReason
		}
		resp.Warnings = w
	}
	lang := map[string]any{
		"name": s.LanguageName,
		"code": s.LanguageCode,
	}
	detCode, detName := s.DetectedCode, s.DetectedName
	if detCode == "" {
		detCode = s.LanguageCode
	}
	if detName == "" {
		detName = s.LanguageName
	}
	lang["detectedLanguage"] = map[string]any{
		"name":       detName,
		"code":       detCode,
		"confidence": s.DetectionConfidence,
		"source":     s.DetectionSource,
	}
	resp.Language = lang

	for _, m := range matches {
		mj := MatchJSON{
			Message:      cleanSuggestion(m.Message),
			ShortMessage: cleanSuggestion(m.ShortMessage),
			Offset:       m.FromPos,
			Length:       m.ToPos - m.FromPos,
			Rule:         RuleJSON{ID: m.RuleID, Description: m.RuleDescription},
		}
		for _, r := range m.SuggestedReplacements {
			mj.Replacements = append(mj.Replacements, ReplacementJSON{Value: r})
		}
		if mj.Replacements == nil {
			mj.Replacements = []ReplacementJSON{}
		}
		if contextSize > 0 && text != "" {
			mj.Context = buildContext(text, m.FromPos, m.ToPos, contextSize)
		}
		resp.Matches = append(resp.Matches, mj)
	}
	b, err := json.Marshal(resp)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func cleanSuggestion(s string) string {
	s = strings.ReplaceAll(s, "<suggestion>", "")
	s = strings.ReplaceAll(s, "</suggestion>", "")
	return s
}

func buildContext(text string, from, to, size int) *ContextJSON {
	if from < 0 {
		from = 0
	}
	if to > len(text) {
		to = len(text)
	}
	start := from - size
	if start < 0 {
		start = 0
	}
	end := to + size
	if end > len(text) {
		end = len(text)
	}
	return &ContextJSON{
		Text:   text[start:end],
		Offset: from - start,
		Length: to - from,
	}
}
