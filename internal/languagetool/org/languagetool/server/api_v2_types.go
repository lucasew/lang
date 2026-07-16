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
	Software SoftwareInfo `json:"software"`
	Language LanguageInfo `json:"language"`
	Matches  []MatchInfo  `json:"matches"`
}

// SoftwareInfo identifies the server.
type SoftwareInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// LanguageInfo describes detected/used language.
type LanguageInfo struct {
	Name string `json:"name"`
	Code string `json:"code"`
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
	return SoftwareInfo{Name: "LanguageTool-Go", Version: version}
}
