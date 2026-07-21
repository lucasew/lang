package server

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// CheckMode ports JLanguageTool.Mode for check queries.
type CheckMode string

const (
	CheckModeAll                 CheckMode = "ALL"
	CheckModeTextLevelOnly       CheckMode = "TEXTLEVEL_ONLY"
	CheckModeAllButTextLevelOnly CheckMode = "ALL_BUT_TEXTLEVEL_ONLY"
)

// CheckLevel ports JLanguageTool.Level.
type CheckLevel string

const (
	CheckLevelDefault CheckLevel = "DEFAULT"
	CheckLevelPicky   CheckLevel = "PICKY"
)

// CheckQueryParams ports TextChecker.QueryParams (full check request options).
type CheckQueryParams struct {
	AltLanguages           []string
	EnabledRules           []string
	DisabledRules          []string
	EnabledCategories      []string
	DisabledCategories     []string
	UseEnabledOnly         bool
	UseQuerySettings       bool
	AllowIncompleteResults bool
	EnableHiddenRules      bool
	Premium                bool
	EnableTempOffRules     bool
	RegressionTestMode     bool
	Mode                   CheckMode
	Level                  CheckLevel
	ToneTags               []string
	Callback               string
	InputLogging           bool
}

var callbackPattern = regexp.MustCompile(`^[a-zA-Z]+$`)

func NewCheckQueryParams() CheckQueryParams {
	return CheckQueryParams{
		Mode:         CheckModeAll,
		Level:        CheckLevelDefault,
		InputLogging: true,
	}
}

func (p CheckQueryParams) Validate() error {
	if p.Callback != "" && !callbackPattern.MatchString(p.Callback) {
		return NewBadRequestError("'callback' value must match [a-zA-Z]+: '" + p.Callback + "'")
	}
	return nil
}

// ToPipelineQuery maps check params into the lighter pool key QueryParams.
func (p CheckQueryParams) ToPipelineQuery() QueryParams {
	return QueryParams{
		EnabledRules:       p.EnabledRules,
		DisabledRules:      p.DisabledRules,
		EnabledCategories:  p.EnabledCategories,
		DisabledCategories: p.DisabledCategories,
		UseEnabledOnly:     p.UseEnabledOnly,
		EnableTempOffRules: p.EnableTempOffRules,
		Premium:            p.Premium,
		UseQuerySettings:   p.UseQuerySettings,
		EnableHiddenRules:  p.EnableHiddenRules,
	}
}

// TextChecker ports org.languagetool.server.TextChecker core validation surface.
type TextChecker struct {
	Config         *HTTPServerConfig
	InternalServer bool
	ReqCounter     *RequestCounter
	Metrics        *ServerMetricsCollector
	ContextSize    int
	// Pool optional pipeline cache; when nil, pipelines are created per check.
	Pool *PipelinePool
}

const DefaultContextSize = 40

func NewTextChecker(cfg *HTTPServerConfig, internal bool, reqCounter *RequestCounter) *TextChecker {
	if cfg == nil {
		cfg = NewHTTPServerConfig()
	}
	if reqCounter == nil {
		reqCounter = NewRequestCounter()
	}
	tc := &TextChecker{
		Config:         cfg,
		InternalServer: internal,
		ReqCounter:     reqCounter,
		Metrics:        Metrics(),
		ContextSize:    DefaultContextSize,
	}
	if cfg != nil && cfg.IsPipelineCachingEnabled() {
		tc.Pool = NewPipelinePool(cfg)
	}
	return tc
}

// CheckParams validates required/common query parameters for a check.
func (t *TextChecker) CheckParams(parameters map[string]string) error {
	if parameters == nil {
		return NewBadRequestError("missing parameters")
	}
	if parameters["language"] == "" {
		return NewBadRequestError("'language' parameter missing")
	}
	if parameters["text"] != "" && parameters["data"] != "" {
		return NewBadRequestError("Set only 'text' or 'data' parameter, not both")
	}
	if parameters["text"] == "" && parameters["data"] == "" {
		return NewBadRequestError("Missing 'text' or 'data' parameter")
	}
	return nil
}

// ValidateTextLength enforces configured max text length for the user tier.
func (t *TextChecker) ValidateTextLength(text string, limits *UserLimits) error {
	if t == nil || t.Config == nil {
		return nil
	}
	maxLen := t.Config.MaxTextLengthAnonymous
	if limits != nil && limits.MaxTextLength > 0 {
		maxLen = limits.MaxTextLength
	}
	if maxLen < int(^uint(0)>>1) && len(text) > maxLen {
		if t.Metrics != nil {
			t.Metrics.LogRequestError(RequestErrorMaxTextSize)
		}
		return NewTextTooLongError("Text exceeds maximum length of " + itoa(maxLen))
	}
	if t.Config.MaxTextHardLength < int(^uint(0)>>1) && len(text) > t.Config.MaxTextHardLength {
		if t.Metrics != nil {
			t.Metrics.LogRequestError(RequestErrorMaxTextSize)
		}
		return NewTextTooLongError("Text exceeds hard maximum length")
	}
	return nil
}

// multiVariantLangBases are short codes that require a country/variant when used as altLanguages.
// Ports TextChecker altLanguage hasVariant() && !isVariant() check for common LT languages.
var multiVariantLangBases = map[string]struct{}{
	"en": {}, "de": {}, "pt": {}, "es": {}, "fr": {}, "nl": {}, "ca": {}, "pl": {},
	"it": {}, "ru": {}, "uk": {}, "gl": {}, "el": {}, "da": {}, "sv": {}, "sk": {},
}

// splitCommaWhitespace ports TextChecker.COMMA_WHITESPACE_PATTERN = ",\\s*".
// Used for altLanguages only (Java); other params use plain split(",").
func splitCommaWhitespace(s string) []string {
	if s == "" {
		return nil
	}
	// Java Pattern.split: ",\\s*" — comma plus optional ASCII whitespace after.
	parts := regexp.MustCompile(`,\s*`).Split(s, -1)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// ParseAltLanguages ports TextChecker altLanguages split (COMMA_WHITESPACE_PATTERN).
func ParseAltLanguages(altCSV string) []string {
	return splitCommaWhitespace(altCSV)
}

// ValidateAltLanguages ports TextChecker altLanguages parsing/validation.
// Unknown codes and bare multi-variant bases (e.g. "en") return BadRequestError.
func ValidateAltLanguages(altCSV string) error {
	codes := ParseAltLanguages(altCSV)
	for _, code := range codes {
		low := strings.ToLower(code)
		if low == "xy" || low == "zz-xx" {
			return NewBadRequestError("Unknown altLanguage '" + code + "'")
		}
		// structural: nonsense codes without known shape
		if err := validateLangCodeShape(code); err != nil {
			return NewBadRequestError(err.Error())
		}
		if !strings.Contains(code, "-") {
			if _, ok := multiVariantLangBases[low]; ok {
				return NewBadRequestError(
					"You specified altLanguage '" + code + "', but for this language you need to specify a variant, e.g. 'en-GB' instead of just 'en'")
			}
		}
		// reject inventively invalid variants like xx-YY when short base is garbage
		base := low
		if i := strings.IndexByte(low, '-'); i >= 0 {
			base = low[:i]
		}
		if len(base) != 2 && len(base) != 3 {
			return NewBadRequestError("Unknown altLanguage '" + code + "'")
		}
		if base == "xy" {
			return NewBadRequestError("Unknown altLanguage '" + code + "'")
		}
	}
	return nil
}

// ValidatePreferredVariants ports detectLanguageOfString preferredVariants checks.
// Each entry must contain a dash (e.g. en-GB); unknown variants fail when isKnown returns false.
// isKnown may be nil (format-only checks).
func ValidatePreferredVariants(variants []string, isKnown func(code string) bool) error {
	for _, preferredVariant := range variants {
		if preferredVariant == "" {
			continue
		}
		if !strings.Contains(preferredVariant, "-") {
			return NewBadRequestError(
				"Invalid format for 'preferredVariants', expected a dash as in 'en-GB': '" + preferredVariant + "'")
		}
		if isKnown != nil && !isKnown(preferredVariant) {
			return NewBadRequestError(
				"Invalid 'preferredVariants', no such language/variant found: '" + preferredVariant + "'")
		}
	}
	return nil
}

func validateLangCodeShape(code string) error {
	// Java language code checks use isEmpty / shape, not Unicode TrimSpace.
	if tools.JavaStringTrimIsEmpty(code) {
		return fmt.Errorf("empty language code")
	}
	// Allow plain short codes and BCP47-like variants (en-US, de-DE, pt-BR).
	parts := strings.Split(code, "-")
	if len(parts) > 3 {
		return fmt.Errorf("'%s' isn't a valid language code", code)
	}
	for _, p := range parts {
		if p == "" {
			return fmt.Errorf("'%s' isn't a valid language code", code)
		}
	}
	return nil
}

// ParseCheckQueryParams builds CheckQueryParams from HTTP query map.
func ParseCheckQueryParams(parameters map[string]string) (CheckQueryParams, error) {
	p := NewCheckQueryParams()
	if parameters == nil {
		return p, nil
	}
	p.EnabledRules = commaSeparated(parameters["enabledRules"])
	p.DisabledRules = commaSeparated(parameters["disabledRules"])
	p.EnabledCategories = commaSeparated(parameters["enabledCategories"])
	p.DisabledCategories = commaSeparated(parameters["disabledCategories"])
	p.UseEnabledOnly = strings.EqualFold(parameters["enabledOnly"], "true")
	p.AllowIncompleteResults = strings.EqualFold(parameters["allowIncompleteResults"], "true")
	p.EnableHiddenRules = strings.EqualFold(parameters["enableHiddenRules"], "true")
	p.EnableTempOffRules = strings.EqualFold(parameters["enableTempOffRules"], "true")
	p.RegressionTestMode = p.EnableTempOffRules
	p.Callback = parameters["callback"]
	if v := parameters["mode"]; v != "" {
		p.Mode = CheckMode(strings.ToUpper(v))
	}
	if v := parameters["level"]; v != "" {
		p.Level = CheckLevel(strings.ToUpper(v))
	}
	p.UseQuerySettings = len(p.EnabledRules) > 0 || len(p.DisabledRules) > 0 ||
		len(p.EnabledCategories) > 0 || len(p.DisabledCategories) > 0 || p.UseEnabledOnly
	if err := p.Validate(); err != nil {
		return p, err
	}
	return p, nil
}

// commaSeparated ports TextChecker.getCommaSeparatedStrings:
// Arrays.asList(disabledParam.split(",")) — no per-item trim; empty slots kept out only when "".
func commaSeparated(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		// Java keeps " STYLE" with leading space; do not TrimSpace.
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
