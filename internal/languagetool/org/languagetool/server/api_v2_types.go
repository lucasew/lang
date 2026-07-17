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
type CheckResponse struct {
	Software          SoftwareInfo  `json:"software"`
	Language          LanguageInfo  `json:"language"`
	Matches           []MatchInfo   `json:"matches"`
	// DetectedLanguage is set when language=auto (soft; same as Language when fixed).
	DetectedLanguage  *LanguageInfo `json:"detectedLanguage,omitempty"`
	// Warnings is a soft list of non-fatal notices.
	Warnings          []string      `json:"warnings,omitempty"`
}

// SoftwareInfo identifies the server.
type SoftwareInfo struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	BuildDate  string `json:"buildDate,omitempty"`
	APIVersion int    `json:"apiVersion,omitempty"`
}

// LanguageInfo describes detected/used language.
type LanguageInfo struct {
	Name                 string `json:"name"`
	Code                 string `json:"code"`
	DetectedLanguage     string `json:"detectedLanguage,omitempty"`
	Confidence           float64 `json:"confidence,omitempty"`
}

// MatchInfo is one public API match.
type MatchInfo struct {
	Message             string              `json:"message"`
	ShortMessage        string              `json:"shortMessage,omitempty"`
	Offset              int                 `json:"offset"`
	Length              int                 `json:"length"`
	Replacements        []ReplacementInfo   `json:"replacements,omitempty"`
	Context             ContextInfo         `json:"context"`
	Rule                RuleInfo            `json:"rule"`
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
}

// NewSoftwareInfo returns default open-source identity.
func NewSoftwareInfo(version string) SoftwareInfo {
	if version == "" {
		version = "dev"
	}
	return SoftwareInfo{Name: "LanguageTool-Go", Version: version, APIVersion: 1}
}

// LanguageNameForCode maps short/variant codes to display names via DefaultCoreLanguages.
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
		if equalFoldASCII(li.Code, code) || equalFoldASCII(li.Code, base) {
			return li.Name
		}
	}
	// soft: title-case base
	if len(base) >= 2 {
		return base
	}
	return low
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
