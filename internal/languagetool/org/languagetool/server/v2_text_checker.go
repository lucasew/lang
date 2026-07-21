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
	return v.BuildResponseExFull(text, langCode, langName, matches, autoDetected, "", nil, 0)
}

// BuildResponseExWithIncomplete is BuildResponseEx with Java incompleteResultsReason.
func (v *V2TextChecker) BuildResponseExWithIncomplete(text, langCode, langName string, matches []RemoteRuleMatch, autoDetected bool, incompleteReason string) (string, error) {
	return v.BuildResponseExFull(text, langCode, langName, matches, autoDetected, incompleteReason, nil, 0)
}

// BuildResponseExFull builds /v2/check JSON.
// incompleteReason ports incompleteResultsReason (Java writeWarningsSection):
// when non-empty → warnings.incompleteResults=true + incompleteResultsReason.
// checkMs is wall-clock check duration for metrics (milliseconds).
func (v *V2TextChecker) BuildResponseExFull(text, langCode, langName string, matches []RemoteRuleMatch, autoDetected bool, incompleteReason string, ignore []IgnoreRangeInfo, checkMs int64) (string, error) {
	if langName == "" || langName == langCode {
		if n := LanguageNameForCode(langCode); n != "" {
			langName = n
		}
	}
	lang := LanguageInfo{Name: langName, Code: langCode, LongCode: langCode}
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
	// Java always writes warnings object when not compactMode; include when incomplete.
	if incompleteReason != "" {
		resp.Warnings = &WarningsInfo{
			IncompleteResults:       true,
			IncompleteResultsReason: incompleteReason,
		}
	} else {
		resp.Warnings = &WarningsInfo{IncompleteResults: false}
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
	if len(ignore) > 0 {
		resp.IgnoreRanges = append([]IgnoreRangeInfo(nil), ignore...)
	} else {
		resp.IgnoreRanges = []IgnoreRangeInfo{}
	}
	b, err := json.Marshal(resp)
	if err != nil {
		return "", err
	}
	if v != nil && v.Metrics != nil {
		v.Metrics.LogCheck(langCode, checkMs, len(text), len(matches), string(CheckModeAll))
	} else {
		Metrics().LogCheck(langCode, checkMs, len(text), len(matches), string(CheckModeAll))
	}
	return string(b), nil
}

// CheckParams adds V2-specific validation on top of TextChecker
// (ports V2TextChecker.checkParams renamed-parameter guards).
func (v *V2TextChecker) CheckParams(parameters map[string]string) error {
	if err := v.TextChecker.CheckParams(parameters); err != nil {
		return err
	}
	if parameters == nil {
		return nil
	}
	// Java V2TextChecker.checkParams
	if parameters["enabled"] != "" {
		return NewBadRequestError("You specified 'enabled' but the parameter is now called 'enabledRules' in v2 of the API")
	}
	if parameters["disabled"] != "" {
		return NewBadRequestError("You specified 'disabled' but the parameter is now called 'disabledRules' in v2 of the API")
	}
	if parameters["preferredvariants"] != "" {
		return NewBadRequestError("You specified 'preferredvariants' but the parameter is now called 'preferredVariants' (uppercase 'V') in v2 of the API")
	}
	if parameters["autodetect"] != "" {
		return NewBadRequestError("You specified 'autodetect' but automatic language detection is now activated with 'language=auto' in v2 of the API")
	}
	return nil
}

// GetPreferredVariants ports V2TextChecker.getPreferredVariants.
func (v *V2TextChecker) GetPreferredVariants(parameters map[string]string) ([]string, error) {
	return ParsePreferredVariants(parameters)
}
