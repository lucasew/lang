package server

import (
	"net"
	"net/http"
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// GetHTTPRequestIp ports ServerTools.getHTTPRequestIp-like extraction.
func GetHTTPRequestIP(r *http.Request, trustXForwardedFor bool) string {
	if r == nil {
		return ""
	}
	if trustXForwardedFor {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			parts := strings.Split(xff, ",")
			// Typical Java XFF first hop trim (String.trim).
			return tools.JavaStringTrim(parts[0])
		}
		if xri := r.Header.Get("X-Real-IP"); xri != "" {
			return tools.JavaStringTrim(xri)
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// CleanUserQuery soft-sanitizes user query text for logs (truncate).
func CleanUserQuery(q string, max int) string {
	if max <= 0 {
		max = 200
	}
	q = strings.ReplaceAll(q, "\n", " ")
	q = tools.JavaStringTrim(q)
	if len(q) > max {
		return q[:max] + "…"
	}
	return q
}

// sentContentRE matches <sentcontent>…</sentcontent> including newlines (Java DOTALL).
var sentContentRE = regexp.MustCompile(`(?s)<sentcontent>.*?</sentcontent>`)

// CleanUserTextFromMessage ports ServerTools.cleanUserTextFromMessage.
// When logging map has inputLogging exactly "no" (case-sensitive), strips
// <sentcontent>…</sentcontent> payloads (Java params.getOrDefault(...).equals("no")).
func CleanUserTextFromMessage(message string, logging map[string]string) string {
	if logging != nil && logging["inputLogging"] == "no" {
		return sentContentRE.ReplaceAllString(message, "<< content removed >>")
	}
	return message
}

// GetMode ports ServerTools.getMode — case-sensitive API values only.
// Missing mode key → ALL; "batch" (undocumented words API) → ALL.
// Present but unknown (including empty string) → BadRequestException.
func GetMode(params map[string]string) (CheckMode, error) {
	if params == nil {
		return CheckModeAll, nil
	}
	modeParam, ok := params["mode"]
	if !ok {
		return CheckModeAll, nil
	}
	switch modeParam {
	case "textLevelOnly":
		return CheckModeTextLevelOnly, nil
	case "allButTextLevelOnly":
		return CheckModeAllButTextLevelOnly, nil
	case "all":
		return CheckModeAll, nil
	case "batch":
		// undocumented API for /words/add, /words/delete
		return CheckModeAll, nil
	default:
		return "", NewBadRequestError(
			"Mode must be one of 'textLevelOnly', 'allButTextLevelOnly', or 'all' but was: '" + modeParam + "'")
	}
}

// GetModeForLog ports ServerTools.getModeForLog.
func GetModeForLog(mode CheckMode) string {
	switch mode {
	case CheckModeTextLevelOnly:
		return "tlo"
	case CheckModeAllButTextLevelOnly:
		return "!tlo"
	case CheckModeAll:
		return "all"
	default:
		return "?"
	}
}

// GetLevel ports ServerTools.getLevel — lowercase API values only (case-sensitive).
// Missing level → DEFAULT.
func GetLevel(params map[string]string) (CheckLevel, error) {
	if params == nil {
		return CheckLevelDefault, nil
	}
	param, ok := params["level"]
	if !ok {
		return CheckLevelDefault, nil
	}
	switch param {
	case "default":
		return CheckLevelDefault, nil
	case "picky":
		return CheckLevelPicky, nil
	case "academic":
		return CheckLevelAcademic, nil
	case "clarity":
		return CheckLevelClarity, nil
	case "professional":
		return CheckLevelProfessional, nil
	case "creative":
		return CheckLevelCreative, nil
	case "customer":
		return CheckLevelCustomer, nil
	case "jobapp":
		return CheckLevelJobApp, nil
	case "objective":
		return CheckLevelObjective, nil
	case "elegant":
		return CheckLevelElegant, nil
	default:
		// Java: Valid values: stream(Level.values()).map(k -> k.toString().toLowerCase())
		parts := make([]string, 0, len(allCheckLevels))
		for _, lv := range allCheckLevels {
			parts = append(parts, strings.ToLower(string(lv)))
		}
		return "", NewBadRequestError(
			"Unknown value '" + param + "' for parameter 'level'. Valid values: " + strings.Join(parts, ", "))
	}
}

// ParseToneTags ports TextChecker toneTags handling from check request params.
//
// Java:
//   - missing toneTags → {ALL_WITHOUT_GOAL_SPECIFIC}
//   - toneTags= (single empty after split) → {ALL_WITHOUT_GOAL_SPECIFIC}
//   - valueOf names; unknown skipped; NO_TONE_RULE / ALL_TONE_RULES ignored when length>1
func ParseToneTags(params map[string]string) []string {
	if params == nil {
		return []string{string(languagetool.ToneAllWithoutGoalSpecific)}
	}
	raw, ok := params["toneTags"]
	if !ok {
		return []string{string(languagetool.ToneAllWithoutGoalSpecific)}
	}
	// Java: params.get("toneTags") != null ? split(",") : null
	names := strings.Split(raw, ",")
	if len(names) == 1 && names[0] == "" {
		// &toneTags=
		return []string{string(languagetool.ToneAllWithoutGoalSpecific)}
	}
	out := make([]string, 0, len(names))
	for _, name := range names {
		if len(names) > 1 && (name == "NO_TONE_RULE" || name == "ALL_TONE_RULES") {
			// Java log.warn and continue
			continue
		}
		if _, ok := languagetool.ParseToneTag(name); ok {
			out = append(out, name)
		}
		// unsupported toneTags: Java log.warn, ignore
	}
	return out
}
