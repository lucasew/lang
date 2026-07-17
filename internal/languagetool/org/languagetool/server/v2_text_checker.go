package server

import (
	"encoding/json"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// V2TextChecker ports org.languagetool.server.V2TextChecker (JSON API).
type V2TextChecker struct {
	*TextChecker
}

func NewV2TextChecker(cfg *HTTPServerConfig, internal bool, reqCounter *RequestCounter) *V2TextChecker {
	return &V2TextChecker{TextChecker: NewTextChecker(cfg, internal, reqCounter)}
}

// GetEnabledRuleIDs ports V2TextChecker.getEnabledRuleIds.
func (v *V2TextChecker) GetEnabledRuleIDs(parameters map[string]string) []string {
	return commaSeparated(parameters["enabledRules"])
}

// GetDisabledRuleIDs ports V2TextChecker.getDisabledRuleIds.
func (v *V2TextChecker) GetDisabledRuleIDs(parameters map[string]string) []string {
	return commaSeparated(parameters["disabledRules"])
}

// GetLanguageAutoDetect ports getLanguageAutoDetect.
func (v *V2TextChecker) GetLanguageAutoDetect(parameters map[string]string) bool {
	return strings.EqualFold(parameters["language"], "auto")
}

// BuildResponse builds a minimal JSON /v2/check response from matches.
func (v *V2TextChecker) BuildResponse(text, langCode, langName string, matches []RemoteRuleMatch) (string, error) {
	return v.BuildResponseEx(text, langCode, langName, matches, false)
}

// BuildResponseEx builds a check response; when autoDetected is true, sets detectedLanguage.
func (v *V2TextChecker) BuildResponseEx(text, langCode, langName string, matches []RemoteRuleMatch, autoDetected bool) (string, error) {
	return v.BuildResponseExWarnings(text, langCode, langName, matches, autoDetected, nil)
}

// BuildResponseExWarnings is BuildResponseEx with optional non-fatal warnings.
func (v *V2TextChecker) BuildResponseExWarnings(text, langCode, langName string, matches []RemoteRuleMatch, autoDetected bool, warnings []string) (string, error) {
	if langName == "" || langName == langCode {
		if n := LanguageNameForCode(langCode); n != "" {
			langName = n
		}
	}
	lang := LanguageInfo{Name: langName, Code: langCode}
	if autoDetected {
		lang.Confidence = 0.5 // soft heuristic confidence
	}
	resp := CheckResponse{
		Software: NewSoftwareInfo("dev"),
		Language: lang,
	}
	if autoDetected {
		dl := lang
		resp.DetectedLanguage = &dl
	}
	if len(warnings) > 0 {
		resp.Warnings = append([]string(nil), warnings...)
	}
	for i := range matches {
		resp.Matches = append(resp.Matches, matches[i].ToMatchInfo())
	}
	if resp.Matches == nil {
		resp.Matches = []MatchInfo{}
	}
	// soft sentence ranges for clients that want sentence boundaries
	for _, sr := range languagetool.PlainSentenceRanges(text, langCode) {
		if sr.ToPos < sr.FromPos {
			continue
		}
		resp.SentenceRanges = append(resp.SentenceRanges, SentenceRangeInfo{
			Offset: sr.FromPos,
			Length: sr.ToPos - sr.FromPos,
		})
	}
	// Multi-language ignore ranges are empty until foreign-span detection is wired;
	// emit an empty array so clients expecting the field can rely on a stable shape.
	if resp.IgnoreRanges == nil {
		resp.IgnoreRanges = []IgnoreRangeInfo{}
	}
	b, err := json.Marshal(resp)
	if err != nil {
		return "", err
	}
	if v != nil && v.Metrics != nil {
		v.Metrics.LogCheck(langCode, 0, len(text), len(matches), string(CheckModeAll))
	}
	return string(b), nil
}

// CheckParams adds V2-specific validation on top of TextChecker.
func (v *V2TextChecker) CheckParams(parameters map[string]string) error {
	if err := v.TextChecker.CheckParams(parameters); err != nil {
		return err
	}
	return nil
}
