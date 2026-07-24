package server

import (
	"fmt"
	"strings"
)

// QueryParams ports TextChecker.QueryParams fields that participate in
// equals/hashCode for PipelineSettings pooling (Java QueryParams).
// Note: toneTags are NOT in Java equals/hashCode — omitted here on purpose.
// AltLanguages ports QueryParams.altLanguages (COMMA_WHITESPACE_PATTERN CSV).
type QueryParams struct {
	EnabledRules           []string
	DisabledRules          []string
	EnabledCategories      []string
	DisabledCategories     []string
	// AltLanguages ports TextChecker.QueryParams.altLanguages (codes like "de-DE").
	// Passed into Pipeline → JLanguageTool like Java Pipeline(lang, altLanguages, …).
	AltLanguages           []string
	UseEnabledOnly         bool
	EnableTempOffRules     bool
	RegressionTestMode     bool // Java: same as enableTempOffRules
	Premium                bool
	UseQuerySettings       bool
	AllowIncompleteResults bool
	EnableHiddenRules      bool
	Mode                   CheckMode
	Level                  CheckLevel
	Callback               string
	InputLogging           bool
	// LanguageCode is a Go-only carrier for check mode when Mode is empty (legacy).
	LanguageCode     string
	MotherTongueCode string
}

// PipelineSettings ports org.languagetool.server.PipelineSettings as a pool key.
type PipelineSettings struct {
	LangCode         string
	MotherTongueCode string
	Query            QueryParams
	// UserConfigKey is a stable hash stand-in (e.g. username or "anon").
	UserConfigKey string
	// GlobalConfigKey is a stable stand-in for GlobalConfig identity.
	GlobalConfigKey string
	// Level is the check level (DEFAULT / PICKY). Empty means DEFAULT.
	// Java JLanguageTool.Level filters Tag.picky rules (false friends, long sentence, …).
	// Also mirrored on Query.Level for Java QueryParams equality.
	Level CheckLevel
}

func NewPipelineSettings(langCode string, userKey string) PipelineSettings {
	return PipelineSettings{
		LangCode:      langCode,
		UserConfigKey: userKey,
		Query: QueryParams{
			Mode:         CheckModeAll,
			Level:        CheckLevelDefault,
			InputLogging: true,
		},
	}
}

func NewPipelineSettingsFull(lang, mother string, q QueryParams, globalKey, userKey string) PipelineSettings {
	return PipelineSettings{
		LangCode:         lang,
		MotherTongueCode: mother,
		Query:            q,
		GlobalConfigKey:  globalKey,
		UserConfigKey:    userKey,
		Level:            q.Level,
	}
}

// joinCSV joins slices without invent trim (Java list equality is order-sensitive).
func joinCSV(parts []string) string {
	return strings.Join(parts, ",")
}

// Key returns a stable map key for pooling.
// Mirrors Java PipelineSettings/QueryParams equals fields (not toneTags).
func (s PipelineSettings) Key() string {
	q := s.Query
	mode := string(q.Mode)
	if mode == "" {
		mode = q.LanguageCode // legacy mode carrier
	}
	level := string(q.Level)
	if level == "" {
		level = string(s.Level)
	}
	return fmt.Sprintf(
		"%s|%s|%s|%s|alt=%s|er=%s|dr=%s|ec=%s|dc=%s|eo=%v|uqs=%v|air=%v|ehr=%v|prem=%v|etor=%v|rtm=%v|mode=%s|level=%s|cb=%s|il=%v",
		s.LangCode, s.MotherTongueCode, s.UserConfigKey, s.GlobalConfigKey,
		joinCSV(q.AltLanguages),
		joinCSV(q.EnabledRules), joinCSV(q.DisabledRules),
		joinCSV(q.EnabledCategories), joinCSV(q.DisabledCategories),
		q.UseEnabledOnly, q.UseQuerySettings, q.AllowIncompleteResults, q.EnableHiddenRules,
		q.Premium, q.EnableTempOffRules, q.RegressionTestMode,
		mode, level, q.Callback, q.InputLogging,
	)
}

func (s PipelineSettings) Equal(o PipelineSettings) bool {
	return s.Key() == o.Key()
}

func (s PipelineSettings) String() string {
	return "PipelineSettings{" + s.Key() + "}"
}
