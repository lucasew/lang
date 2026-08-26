package tools

// StringTools is the Java-name twin for package string helpers.
type StringTools struct{}

func (StringTools) IsEmpty(s string) bool           { return IsEmptyStr(s) }
func (StringTools) IsAllUppercase(s string) bool    { return IsAllUppercase(s) }
func (StringTools) PreserveCase(in, model string) string {
	return PreserveCase(in, model)
}

// StringInterner is the Java-name twin for Intern.
type StringInterner struct{}

func (StringInterner) Intern(s string) string { return Intern(s) }

// Tools is the Java-name twin for misc Tools helpers (i18n, etc.).
type Tools struct{}

func (Tools) I18n(pattern string, args ...any) string { return I18n(pattern, args...) }

// LoggingTools is the Java-name twin for logging markers.
type LoggingTools struct{}

func (LoggingTools) MarkerInit() string       { return LogMarkerInit }
func (LoggingTools) MarkerCheck() string      { return LogMarkerCheck }
func (LoggingTools) MarkerRequest() string    { return LogMarkerRequest }
func (LoggingTools) MarkerBadRequest() string { return LogMarkerBadRequest }

// JnaTools is the Java-name twin for JNA workarounds.
type JnaTools struct{}

func (JnaTools) SetBugWorkaroundProperty() { SetJnaBugWorkaroundProperty() }
