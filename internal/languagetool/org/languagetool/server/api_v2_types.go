package server

// CheckParams is the surface of ApiV2 /v2/check request parameters.
type CheckParams struct {
	Text           string
	Language       string // code or "auto"
	MotherTongue   string
	EnabledRules   []string
	DisabledRules  []string
	EnabledOnly    bool
	PreferredVariants []string
	Level          string // default, picky
	Mode           string // textLevelOnly, allButTextLevelOnly, all
}

// CheckResponse is a minimal /v2/check JSON shape.
// Java RuleMatchesAsJsonSerializer: software, warnings, language (with nested
// detectedLanguage), matches, ignoreRanges, sentenceRanges — no top-level detectedLanguage.
type CheckResponse struct {
	Software SoftwareInfo `json:"software"`
	Language LanguageInfo `json:"language"`
	Matches  []MatchInfo  `json:"matches"`
	// SentenceRanges lists plain-text sentence spans (offset/length).
	SentenceRanges []SentenceRangeInfo `json:"sentenceRanges,omitempty"`
	// IgnoreRanges is multi-language foreign-span output from CheckResults.
	IgnoreRanges []IgnoreRangeInfo `json:"ignoreRanges"`
	// Warnings ports RuleMatchesAsJsonSerializer.writeWarningsSection:
	// { "incompleteResults": bool, "incompleteResultsReason"?: string }.
	Warnings *WarningsInfo `json:"warnings,omitempty"`
}

// WarningsInfo ports the JSON "warnings" object (not a string list invent).
type WarningsInfo struct {
	IncompleteResults       bool   `json:"incompleteResults"`
	IncompleteResultsReason string `json:"incompleteResultsReason,omitempty"`
}

// SentenceRangeInfo is one sentence span in the public API.
type SentenceRangeInfo struct {
	Offset int `json:"offset"`
	Length int `json:"length"`
}

// IgnoreRangeInfo ports RemoteIgnoreRange for /v2/check (multi-language).
type IgnoreRangeInfo struct {
	From int    `json:"from"`
	To   int    `json:"to"`
	Lang string `json:"language,omitempty"`
}

// SoftwareInfo identifies the server.
type SoftwareInfo struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	BuildDate  string `json:"buildDate,omitempty"`
	APIVersion int    `json:"apiVersion,omitempty"`
}

// LanguageInfo describes used language (and optional nested detection).
// Ports RuleMatchesAsJsonSerializer.writeLanguageSection outer object:
// name, code, [spellCheckOnly], detectedLanguage{name,code,confidence,[spellCheckOnly],source}.
type LanguageInfo struct {
	Name string `json:"name"`
	Code string `json:"code"`
	// LongCode is used by /languages list (not writeLanguageSection).
	LongCode string `json:"longCode,omitempty"`
	// SpellCheckOnly ports Language.isSpellcheckOnlyLanguage when true.
	SpellCheckOnly bool `json:"spellCheckOnly,omitempty"`
	// DetectedLanguage nested object (Java always writes for /v2/check).
	DetectedLanguage *DetectedLanguageInfo `json:"detectedLanguage,omitempty"`
}

// DetectedLanguageInfo ports the nested language.detectedLanguage JSON object.
type DetectedLanguageInfo struct {
	Name               string  `json:"name"`
	Code               string  `json:"code"`
	Confidence         float64 `json:"confidence"`
	SpellCheckOnly     bool    `json:"spellCheckOnly,omitempty"`
	// Source is always serialized (null when unknown) like Java writeStringField("source", …).
	Source *string `json:"source"`
}

// MatchInfo is one public API match.
type MatchInfo struct {
	Message             string              `json:"message"`
	ShortMessage        string              `json:"shortMessage,omitempty"`
	Offset              int                 `json:"offset"`
	Length              int                 `json:"length"`
	Replacements        []ReplacementInfo   `json:"replacements,omitempty"`
	Context             ContextInfo         `json:"context"`
	// ContextForSureMatch ports Rule.estimateContextForSureMatch (0 omitted; text-level -1).
	ContextForSureMatch int                 `json:"contextForSureMatch,omitempty"`
	// Type soft-ports the ITS type wrapper object from the Java JSON API.
	Type                *MatchTypeInfo      `json:"type,omitempty"`
	Rule                RuleInfo            `json:"rule"`
}

// MatchTypeInfo is the soft ITS type object on a match.
type MatchTypeInfo struct {
	TypeName string `json:"typeName"`
}

type ReplacementInfo struct {
	Value string `json:"value"`
}

type ContextInfo struct {
	Text   string `json:"text"`
	Offset int    `json:"offset"`
	Length int    `json:"length"`
}

type RuleInfo struct {
	ID          string `json:"id"`
	Description string `json:"description,omitempty"`
	IssueType   string `json:"issueType,omitempty"`
	Category    struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"category"`
	// Urls soft documentation links (community rule pages).
	Urls []struct {
		Value string `json:"value"`
	} `json:"urls,omitempty"`
}

// NewSoftwareInfo returns default open-source identity.
func NewSoftwareInfo(version string) SoftwareInfo {
	if version == "" {
		version = "dev"
	}
	return SoftwareInfo{
		Name:       "LanguageTool-Go",
		Version:    version,
		BuildDate:  "dev",
		APIVersion: 1,
	}
}

// LanguageNameForCode maps short/variant codes to the simple display name
// (e.g. en / en-US → "English") via corepack-supported languages.
func LanguageNameForCode(code string) string {
	if code == "" {
		return ""
	}
	low := code
	base := code
	if i := indexDash(code); i > 0 {
		base = code[:i]
	}
	for _, li := range DefaultCoreLanguages() {
		if equalFoldASCII(li.Code, base) {
			// Strip variant parenthetical "English (US)" → "English" (display name only).
			name := li.Name
			if j := indexByte(name, '('); j > 0 {
				name = trimSpace(name[:j])
			}
			if name != "" {
				return name
			}
			return li.Name
		}
	}
	// Unknown code: return base as-is (fail closed — no invent display name).
	if base != "" {
		return base
	}
	return low
}

func indexByte(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

func trimSpace(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}
	return s
}

func indexDash(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == '-' {
			return i
		}
	}
	return -1
}

func equalFoldASCII(a, b string) bool {
	if len(a) != len(b) {
		// also compare bases
	}
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}
