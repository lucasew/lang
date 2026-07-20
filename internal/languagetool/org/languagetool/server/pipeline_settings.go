package server

import "fmt"

// QueryParams is a lightweight stand-in for TextChecker.QueryParams.
type QueryParams struct {
	EnabledRules     []string
	DisabledRules    []string
	EnabledCategories  []string
	DisabledCategories []string
	UseEnabledOnly   bool
	EnableTempOffRules bool
	Premium          bool
	UseQuerySettings bool
	EnableHiddenRules bool
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
	Level CheckLevel
}

func NewPipelineSettings(langCode string, userKey string) PipelineSettings {
	return PipelineSettings{
		LangCode:      langCode,
		UserConfigKey: userKey,
	}
}

func NewPipelineSettingsFull(lang, mother string, q QueryParams, globalKey, userKey string) PipelineSettings {
	return PipelineSettings{
		LangCode:         lang,
		MotherTongueCode: mother,
		Query:            q,
		GlobalConfigKey:  globalKey,
		UserConfigKey:    userKey,
	}
}

// Key returns a stable map key for pooling.
func (s PipelineSettings) Key() string {
	return fmt.Sprintf("%s|%s|%s|%s|en=%v|prem=%v",
		s.LangCode, s.MotherTongueCode, s.UserConfigKey, s.GlobalConfigKey,
		s.Query.UseEnabledOnly, s.Query.Premium)
}

func (s PipelineSettings) Equal(o PipelineSettings) bool {
	return s.Key() == o.Key()
}

func (s PipelineSettings) String() string {
	return "PipelineSettings{" + s.Key() + "}"
}
