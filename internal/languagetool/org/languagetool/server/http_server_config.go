package server

import (
	"math"
	"strconv"
	"strings"
)

// HTTPServerConfig ports the core surface of org.languagetool.server.HTTPServerConfig.
// Full property-file loading is simplified; fields match Java defaults for open-source use.
type HTTPServerConfig struct {
	Verbose      bool
	PublicAccess bool
	Port         int
	MinPort      int
	MaxPort      int

	AllowOriginURL string
	LogIP          bool
	ServerURL      string

	MaxTextLengthAnonymous int
	MaxTextLengthLoggedIn  int
	MaxTextLengthPremium   int
	MaxTextHardLength      int

	MaxCheckTimeMillisAnonymous int64
	MaxCheckTimeMillisLoggedIn  int64
	MaxCheckTimeMillisPremium   int64

	MaxCheckThreads       int
	MaxTextCheckerThreads int
	TextCheckerQueueSize  int

	PipelineCaching     bool
	PipelinePrewarming  bool
	MaxPipelinePoolSize int
	PipelineExpireTime  int // seconds

	RequestLimit               int
	RequestLimitInBytes        int
	TimeoutRequestLimit        int
	RequestLimitPeriodInSeconds int
	IPFingerprintFactor        int
	TrustXForwardedForHeader   bool
	MaxWorkQueueSize           int

	CacheSize      int
	CacheTTLSeconds int64

	MaxErrorsPerWordRate  float64
	SuggestionsEnabled    bool
	MaxSpellingSuggestions int

	PremiumAlways bool
	PremiumOnly   bool

	AnonymousAccessAllowed bool
	DisabledRuleIDs        []string

	PrometheusMonitoring bool
	PrometheusPort       int

	MotherTongue       string
	PreferredLanguages []string
	LocalAPIMode       bool

	// LanguageModelDir is the optional ngram / LM directory from --languageModel.
	LanguageModelDir string

	// BlockedReferrers is a list of HTTP Referer / Origin substrings that are rejected.
	BlockedReferrers []string
}

const (
	DefaultHost           = "localhost"
	DefaultPort           = 8081
	LanguageModelOption   = "--languageModel"
)

func NewHTTPServerConfig() *HTTPServerConfig {
	return &HTTPServerConfig{
		Port:                        DefaultPort,
		LogIP:                       true,
		MaxTextLengthAnonymous:      math.MaxInt32,
		MaxTextLengthLoggedIn:       math.MaxInt32,
		MaxTextLengthPremium:        math.MaxInt32,
		MaxTextHardLength:           math.MaxInt32,
		MaxCheckTimeMillisAnonymous: -1,
		MaxCheckTimeMillisLoggedIn:  -1,
		MaxCheckTimeMillisPremium:   -1,
		MaxCheckThreads:             10,
		TextCheckerQueueSize:        8,
		IPFingerprintFactor:         1,
		SuggestionsEnabled:          true,
		AnonymousAccessAllowed:      true,
		CacheTTLSeconds:             300,
		PrometheusPort:              9301,
		MotherTongue:                "en-US",
	}
}

func NewHTTPServerConfigPort(port int) *HTTPServerConfig {
	c := NewHTTPServerConfig()
	if port > 0 {
		c.Port = port
	}
	return c
}

func NewHTTPServerConfigPortVerbose(port int, verbose bool) *HTTPServerConfig {
	c := NewHTTPServerConfigPort(port)
	c.Verbose = verbose
	return c
}

// ApplyArgs applies a subset of CLI flags (--port, --public, --verbose).
func (c *HTTPServerConfig) ApplyArgs(args []string) error {
	if c == nil {
		return NewIllegalConfigurationError("nil config")
	}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--port":
			if i+1 >= len(args) {
				return NewIllegalConfigurationError("missing value for --port")
			}
			p, err := strconv.Atoi(args[i+1])
			if err != nil {
				return NewIllegalConfigurationError("invalid --port: " + args[i+1])
			}
			c.Port = p
			i++
		case "--public":
			c.PublicAccess = true
		case "--verbose":
			c.Verbose = true
		case "--premiumAlways":
			c.PremiumAlways = true
		case LanguageModelOption:
			if i+1 >= len(args) {
				return NewIllegalConfigurationError("missing value for --languageModel")
			}
			c.LanguageModelDir = args[i+1]
			i++
		}
	}
	return nil
}

// NewHTTPServerConfigFromArgs builds a config and applies CLI flags (Java HTTPServerConfig(String[])).
func NewHTTPServerConfigFromArgs(args []string) (*HTTPServerConfig, error) {
	c := NewHTTPServerConfig()
	if err := c.ApplyArgs(args); err != nil {
		return nil, err
	}
	return c, nil
}

// ApplyProperties applies key=value pairs (simplified property map).
func (c *HTTPServerConfig) ApplyProperties(props map[string]string) {
	if c == nil || props == nil {
		return
	}
	if v, ok := props["port"]; ok {
		if p, err := strconv.Atoi(v); err == nil {
			c.Port = p
		}
	}
	if v, ok := props["maxTextLength"]; ok {
		if n, err := strconv.Atoi(v); err == nil {
			c.MaxTextLengthAnonymous = n
			c.MaxTextLengthLoggedIn = n
			c.MaxTextLengthPremium = n
		}
	}
	if v, ok := props["maxTextLengthAnonymous"]; ok {
		if n, err := strconv.Atoi(v); err == nil {
			c.MaxTextLengthAnonymous = n
		}
	}
	if v, ok := props["maxTextLengthLoggedIn"]; ok {
		if n, err := strconv.Atoi(v); err == nil {
			c.MaxTextLengthLoggedIn = n
		}
	}
	if v, ok := props["maxTextLengthPremium"]; ok {
		if n, err := strconv.Atoi(v); err == nil {
			c.MaxTextLengthPremium = n
		}
	}
	if v, ok := props["maxCheckTimeMillis"]; ok {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			c.MaxCheckTimeMillisAnonymous = n
			c.MaxCheckTimeMillisLoggedIn = n
			c.MaxCheckTimeMillisPremium = n
		}
	}
	if v, ok := props["requestLimit"]; ok {
		if n, err := strconv.Atoi(v); err == nil {
			c.RequestLimit = n
		}
	}
	if v, ok := props["requestLimitPeriodInSeconds"]; ok {
		if n, err := strconv.Atoi(v); err == nil {
			c.RequestLimitPeriodInSeconds = n
		}
	}
	if v, ok := props["pipelineCaching"]; ok {
		c.PipelineCaching = strings.EqualFold(strings.TrimSpace(v), "true")
	}
	if v, ok := props["maxPipelinePoolSize"]; ok {
		if n, err := strconv.Atoi(v); err == nil {
			c.MaxPipelinePoolSize = n
		}
	}
	if v, ok := props["pipelineExpireTimeInSeconds"]; ok {
		if n, err := strconv.Atoi(v); err == nil {
			c.PipelineExpireTime = n
		}
	}
	if v, ok := props["trustXForwardForHeader"]; ok {
		c.TrustXForwardedForHeader = strings.EqualFold(strings.TrimSpace(v), "true")
	}
	if v, ok := props["premiumAlways"]; ok {
		c.PremiumAlways = strings.EqualFold(strings.TrimSpace(v), "true")
	}
	if v, ok := props["disabledRuleIds"]; ok && v != "" {
		parts := strings.Split(v, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				out = append(out, p)
			}
		}
		c.DisabledRuleIDs = out
	}
}

func (c *HTTPServerConfig) IsPipelineCachingEnabled() bool {
	return c != nil && c.PipelineCaching
}

func (c *HTTPServerConfig) GetMaxPipelinePoolSize() int {
	if c == nil || c.MaxPipelinePoolSize <= 0 {
		return 10
	}
	return c.MaxPipelinePoolSize
}

// SetBlockedReferrers replaces the blocked referrer list (HTTPSServerTest parity).
func (c *HTTPServerConfig) SetBlockedReferrers(refs []string) {
	if c == nil {
		return
	}
	c.BlockedReferrers = append([]string(nil), refs...)
}

// GetBlockedReferrers returns a copy of the blocked referrer list.
func (c *HTTPServerConfig) GetBlockedReferrers() []string {
	if c == nil {
		return nil
	}
	return append([]string(nil), c.BlockedReferrers...)
}

// IsBlockedReferrer reports whether referer/origin matches any blocked entry
// (substring match, case-sensitive like Java contains checks).
func (c *HTTPServerConfig) IsBlockedReferrer(referer string) bool {
	if c == nil || referer == "" {
		return false
	}
	for _, b := range c.BlockedReferrers {
		b = strings.TrimSpace(b)
		if b == "" {
			continue
		}
		if strings.Contains(referer, b) {
			return true
		}
	}
	return false
}
