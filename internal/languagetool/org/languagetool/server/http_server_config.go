package server

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
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

	// FasttextModel / FasttextBinary ports HTTPServerConfig.fasttextModel/fasttextBinary
	// for LanguageIdentifierService.getDefaultLanguageIdentifier.
	FasttextModel  string
	FasttextBinary string
	// NgramLangIdentData ports ngramLangIdentData (ZIP path for lang-id ngrams).
	NgramLangIdentData string

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

// getOptionalProperty ports HTTPServerConfig.getOptionalProperty:
// props.get(name); null → defaultValue (empty string key present is not null).
func getOptionalProperty(props map[string]string, name, defaultValue string) string {
	if props == nil {
		return defaultValue
	}
	if v, ok := props[name]; ok {
		return v
	}
	return defaultValue
}

func parseBoolJava(v string) bool {
	// Boolean.parseBoolean / Boolean.valueOf after String.trim
	return strings.EqualFold(tools.JavaStringTrim(v), "true")
}

func parseIntJava(v string, def int) int {
	n, err := strconv.Atoi(tools.JavaStringTrim(v))
	if err != nil {
		return def
	}
	return n
}

func parseInt64Java(v string, def int64) int64 {
	n, err := strconv.ParseInt(tools.JavaStringTrim(v), 10, 64)
	if err != nil {
		return def
	}
	return n
}

func parseFloatJava(v string, def float64) float64 {
	n, err := strconv.ParseFloat(tools.JavaStringTrim(v), 64)
	if err != nil {
		return def
	}
	return n
}

// ApplyProperties applies present keys from a property map (partial update).
// Order matches Java parseConfigFile for keys that share defaults (maxTextLength → tier overrides).
func (c *HTTPServerConfig) ApplyProperties(props map[string]string) {
	if c == nil || props == nil {
		return
	}
	if _, ok := props["port"]; ok {
		c.Port = parseIntJava(props["port"], c.Port)
	}
	if _, ok := props["minPort"]; ok {
		c.MinPort = parseIntJava(props["minPort"], 0)
	}
	if _, ok := props["maxPort"]; ok {
		c.MaxPort = parseIntJava(props["maxPort"], 0)
	}
	if _, ok := props["serverURL"]; ok {
		c.ServerURL = props["serverURL"]
	}
	if _, ok := props["maxTextHardLength"]; ok {
		c.MaxTextHardLength = parseIntJava(props["maxTextHardLength"], math.MaxInt32)
	}
	// Java: maxTextLength sets all three, then tier keys override
	if _, ok := props["maxTextLength"]; ok {
		n := parseIntJava(props["maxTextLength"], math.MaxInt32)
		c.MaxTextLengthAnonymous = n
		c.MaxTextLengthLoggedIn = n
		c.MaxTextLengthPremium = n
	}
	if _, ok := props["maxTextLengthAnonymous"]; ok {
		c.MaxTextLengthAnonymous = parseIntJava(props["maxTextLengthAnonymous"], c.MaxTextLengthAnonymous)
	}
	if _, ok := props["maxTextLengthLoggedIn"]; ok {
		c.MaxTextLengthLoggedIn = parseIntJava(props["maxTextLengthLoggedIn"], c.MaxTextLengthLoggedIn)
	}
	if _, ok := props["maxTextLengthPremium"]; ok {
		c.MaxTextLengthPremium = parseIntJava(props["maxTextLengthPremium"], c.MaxTextLengthPremium)
	}
	if _, ok := props["maxCheckTimeMillis"]; ok {
		n := parseInt64Java(props["maxCheckTimeMillis"], -1)
		c.MaxCheckTimeMillisAnonymous = n
		c.MaxCheckTimeMillisLoggedIn = n
		c.MaxCheckTimeMillisPremium = n
	}
	if _, ok := props["maxCheckTimeMillisAnonymous"]; ok {
		c.MaxCheckTimeMillisAnonymous = parseInt64Java(props["maxCheckTimeMillisAnonymous"], c.MaxCheckTimeMillisAnonymous)
	}
	if _, ok := props["maxCheckTimeMillisLoggedIn"]; ok {
		c.MaxCheckTimeMillisLoggedIn = parseInt64Java(props["maxCheckTimeMillisLoggedIn"], c.MaxCheckTimeMillisLoggedIn)
	}
	if _, ok := props["maxCheckTimeMillisPremium"]; ok {
		c.MaxCheckTimeMillisPremium = parseInt64Java(props["maxCheckTimeMillisPremium"], c.MaxCheckTimeMillisPremium)
	}
	if _, ok := props["requestLimit"]; ok {
		c.RequestLimit = parseIntJava(props["requestLimit"], 0)
	}
	if _, ok := props["requestLimitInBytes"]; ok {
		c.RequestLimitInBytes = parseIntJava(props["requestLimitInBytes"], 0)
	}
	if _, ok := props["timeoutRequestLimit"]; ok {
		c.TimeoutRequestLimit = parseIntJava(props["timeoutRequestLimit"], 0)
	}
	if _, ok := props["requestLimitPeriodInSeconds"]; ok {
		c.RequestLimitPeriodInSeconds = parseIntJava(props["requestLimitPeriodInSeconds"], 0)
	}
	if _, ok := props["ipFingerprintFactor"]; ok {
		c.IPFingerprintFactor = parseIntJava(props["ipFingerprintFactor"], 1)
	}
	if _, ok := props["maxWorkQueueSize"]; ok {
		n := parseIntJava(props["maxWorkQueueSize"], 0)
		if n < 0 {
			panic("maxWorkQueueSize must be >= 0: " + strconv.Itoa(n))
		}
		c.MaxWorkQueueSize = n
	}
	if _, ok := props["pipelineCaching"]; ok {
		c.PipelineCaching = parseBoolJava(props["pipelineCaching"])
	}
	if _, ok := props["pipelinePrewarming"]; ok {
		c.PipelinePrewarming = parseBoolJava(props["pipelinePrewarming"])
	}
	if _, ok := props["maxPipelinePoolSize"]; ok {
		c.MaxPipelinePoolSize = parseIntJava(props["maxPipelinePoolSize"], 5)
	}
	if _, ok := props["pipelineExpireTimeInSeconds"]; ok {
		c.PipelineExpireTime = parseIntJava(props["pipelineExpireTimeInSeconds"], 10)
	}
	if _, ok := props["trustXForwardForHeader"]; ok {
		c.TrustXForwardedForHeader = parseBoolJava(props["trustXForwardForHeader"])
	}
	if _, ok := props["maxCheckThreads"]; ok {
		n := parseIntJava(props["maxCheckThreads"], 10)
		if n < 1 {
			panic("Invalid value for maxCheckThreads, must be >= 1: " + strconv.Itoa(n))
		}
		c.MaxCheckThreads = n
	}
	if _, ok := props["maxTextCheckerThreads"]; ok {
		n := parseIntJava(props["maxTextCheckerThreads"], 0)
		if n < 0 {
			panic("Invalid value for maxTextCheckerThreads, must be >= 1: " + strconv.Itoa(n))
		}
		c.MaxTextCheckerThreads = n
	}
	if _, ok := props["textCheckerQueueSize"]; ok {
		n := parseIntJava(props["textCheckerQueueSize"], 8)
		if n < 0 {
			panic("Invalid value for textCheckerQueueSize, must be >= 1: " + strconv.Itoa(n))
		}
		c.TextCheckerQueueSize = n
	}
	if _, ok := props["cacheSize"]; ok {
		c.CacheSize = parseIntJava(props["cacheSize"], 0)
	}
	if _, ok := props["cacheTTLSeconds"]; ok {
		c.CacheTTLSeconds = parseInt64Java(props["cacheTTLSeconds"], 300)
	}
	if _, ok := props["maxErrorsPerWordRate"]; ok {
		c.MaxErrorsPerWordRate = parseFloatJava(props["maxErrorsPerWordRate"], 0)
	}
	if _, ok := props["suggestionsEnabled"]; ok {
		c.SuggestionsEnabled = parseBoolJava(props["suggestionsEnabled"])
	}
	if _, ok := props["maxSpellingSuggestions"]; ok {
		c.MaxSpellingSuggestions = parseIntJava(props["maxSpellingSuggestions"], 0)
	}
	if _, ok := props["blockedReferrers"]; ok {
		// Java: Arrays.asList(split(",\\s*")) even when empty → [""] from ""
		c.BlockedReferrers = splitCommaOptionalASCIIWS(props["blockedReferrers"])
	}
	if _, ok := props["premiumAlways"]; ok {
		c.PremiumAlways = parseBoolJava(props["premiumAlways"])
	}
	if _, ok := props["premiumOnly"]; ok {
		c.PremiumOnly = parseBoolJava(props["premiumOnly"])
	}
	if _, ok := props["anonymousAccessAllowed"]; ok {
		c.AnonymousAccessAllowed = parseBoolJava(props["anonymousAccessAllowed"])
	}
	if _, ok := props["prometheusMonitoring"]; ok {
		c.PrometheusMonitoring = parseBoolJava(props["prometheusMonitoring"])
	}
	if _, ok := props["prometheusPort"]; ok {
		c.PrometheusPort = parseIntJava(props["prometheusPort"], 9301)
	}
	if _, ok := props["disabledRuleIds"]; ok {
		// Java always splits; empty → [""] via Pattern; we use splitCommaOptionalASCIIWS
		c.DisabledRuleIDs = splitCommaOptionalASCIIWS(props["disabledRuleIds"])
	}
	if _, ok := props["localApiMode"]; ok {
		c.LocalAPIMode = parseBoolJava(props["localApiMode"])
	}
	if _, ok := props["motherTongue"]; ok {
		c.MotherTongue = props["motherTongue"]
	}
	if _, ok := props["preferredLanguages"]; ok {
		// Java: replace(" ", "") then split(",")
		raw := strings.ReplaceAll(props["preferredLanguages"], " ", "")
		if raw != "" {
			c.PreferredLanguages = strings.Split(raw, ",")
		} else {
			c.PreferredLanguages = nil
		}
	}
	// Java: if (fasttextBinary != null && fasttextModel != null) setFasttextPaths(...)
	ftModel := props["fasttextModel"]
	ftBin := props["fasttextBinary"]
	if ftModel != "" && ftBin != "" {
		if err := c.SetFasttextPaths(ftModel, ftBin); err != nil {
			panic(err.Error())
		}
	}
	// Java: ngramLangIdentData must exist and not be a directory → IllegalArgumentException
	if v, ok := props["ngramLangIdentData"]; ok && v != "" {
		if err := c.SetNgramLangIdentData(v); err != nil {
			panic(err.Error())
		}
	}
	if v, ok := props["languageModel"]; ok && v != "" {
		c.LanguageModelDir = v
	}
	if mode, ok := props["mode"]; ok {
		// Java: AfterTheDeadline mode rejected
		if strings.EqualFold(tools.JavaStringTrim(mode), "AfterTheDeadline") {
			panic("The AfterTheDeadline mode is not supported anymore in LanguageTool 3.8 or later")
		}
	}
}

// ParseJavaProperties ports java.util.Properties.load for simple UTF-8 property files
// (key=value / key:value / key value, # and ! comments).
// Separator rules match LineReader/loadConvert key scanning in OpenJDK Properties.
func ParseJavaProperties(data string) map[string]string {
	out := map[string]string{}
	data = strings.ReplaceAll(data, "\r\n", "\n")
	data = strings.ReplaceAll(data, "\r", "\n")
	lines := strings.Split(data, "\n")
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		// line continuation with trailing \
		for strings.HasSuffix(line, `\`) && !strings.HasSuffix(line, `\\`) && i+1 < len(lines) {
			line = line[:len(line)-1] + lines[i+1]
			i++
		}
		// Java does not trim whole line first; skip leading whitespace then check comment
		s := line
		for len(s) > 0 && (s[0] == ' ' || s[0] == '\t' || s[0] == '\f') {
			s = s[1:]
		}
		if s == "" || s[0] == '#' || s[0] == '!' {
			continue
		}
		// key: until first unescaped =, :, or whitespace
		keyEnd := -1
		hasSep := false
		for j := 0; j < len(s); j++ {
			if s[j] == '\\' {
				j++
				continue
			}
			if s[j] == '=' || s[j] == ':' || s[j] == ' ' || s[j] == '\t' || s[j] == '\f' {
				keyEnd = j
				hasSep = true
				break
			}
		}
		var key, val string
		if !hasSep {
			key = s
			val = ""
		} else {
			key = s[:keyEnd]
			rest := s[keyEnd:]
			// skip whitespace in separator region
			for len(rest) > 0 && (rest[0] == ' ' || rest[0] == '\t' || rest[0] == '\f') {
				rest = rest[1:]
			}
			// if next is = or :, consume it and following whitespace
			if len(rest) > 0 && (rest[0] == '=' || rest[0] == ':') {
				rest = rest[1:]
				for len(rest) > 0 && (rest[0] == ' ' || rest[0] == '\t' || rest[0] == '\f') {
					rest = rest[1:]
				}
			}
			val = rest
		}
		if key == "" {
			continue
		}
		// Java Properties.loadConvert on key and value
		out[javaPropertiesLoadConvert(key)] = javaPropertiesLoadConvert(val)
	}
	return out
}

// javaPropertiesLoadConvert ports Properties.loadConvert:
// \t \n \r \f \\ and \uXXXX (4 hex digits, case-insensitive).
func javaPropertiesLoadConvert(s string) string {
	if !strings.Contains(s, `\`) {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c != '\\' {
			b.WriteByte(c)
			continue
		}
		i++
		if i >= len(s) {
			break
		}
		switch s[i] {
		case 't':
			b.WriteByte('\t')
		case 'n':
			b.WriteByte('\n')
		case 'r':
			b.WriteByte('\r')
		case 'f':
			b.WriteByte('\f')
		case '\\':
			b.WriteByte('\\')
		case 'u':
			// \uXXXX — need 4 hex digits
			if i+4 < len(s) {
				hex := s[i+1 : i+5]
				if v, err := strconv.ParseUint(hex, 16, 16); err == nil {
					b.WriteRune(rune(v))
					i += 4
					continue
				}
			}
			// malformed: keep as-is
			b.WriteByte('u')
		default:
			// Java: other escapes just yield the char (e.g. \= → =)
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

// LoadPropertiesFile reads a Java .properties file into a map.
func LoadPropertiesFile(path string) (map[string]string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseJavaProperties(string(b)), nil
}

// LoadFromPropertyFile ports parseConfigFile(File, loadLangModel=true) for open-source fields.
// Missing keys use Java getOptionalProperty defaults (not Go NewHTTPServerConfig zeros).
func (c *HTTPServerConfig) LoadFromPropertyFile(path string) error {
	if c == nil {
		return NewIllegalConfigurationError("nil config")
	}
	props, err := LoadPropertiesFile(path)
	if err != nil {
		return fmt.Errorf("Could not load properties from '%s': %w", path, err)
	}
	c.applyPropertyFileDefaults(props)
	return nil
}

// applyPropertyFileDefaults applies Java parseConfigFile defaults for every key.
func (c *HTTPServerConfig) applyPropertyFileDefaults(props map[string]string) {
	// Seed with Java file-load defaults then ApplyProperties for present keys
	// so missing keys still get Java defaults (not leftover NewHTTPServerConfig).
	c.MaxTextHardLength = parseIntJava(getOptionalProperty(props, "maxTextHardLength", strconv.Itoa(math.MaxInt32)), math.MaxInt32)
	mt := parseIntJava(getOptionalProperty(props, "maxTextLength", strconv.Itoa(math.MaxInt32)), math.MaxInt32)
	c.MaxTextLengthAnonymous = parseIntJava(getOptionalProperty(props, "maxTextLengthAnonymous", strconv.Itoa(mt)), mt)
	c.MaxTextLengthLoggedIn = parseIntJava(getOptionalProperty(props, "maxTextLengthLoggedIn", strconv.Itoa(mt)), mt)
	c.MaxTextLengthPremium = parseIntJava(getOptionalProperty(props, "maxTextLengthPremium", strconv.Itoa(mt)), mt)

	mct := parseInt64Java(getOptionalProperty(props, "maxCheckTimeMillis", "-1"), -1)
	c.MaxCheckTimeMillisAnonymous = parseInt64Java(getOptionalProperty(props, "maxCheckTimeMillisAnonymous", strconv.FormatInt(mct, 10)), mct)
	c.MaxCheckTimeMillisLoggedIn = parseInt64Java(getOptionalProperty(props, "maxCheckTimeMillisLoggedIn", strconv.FormatInt(mct, 10)), mct)
	c.MaxCheckTimeMillisPremium = parseInt64Java(getOptionalProperty(props, "maxCheckTimeMillisPremium", strconv.FormatInt(mct, 10)), mct)

	c.RequestLimit = parseIntJava(getOptionalProperty(props, "requestLimit", "0"), 0)
	c.RequestLimitInBytes = parseIntJava(getOptionalProperty(props, "requestLimitInBytes", "0"), 0)
	c.TimeoutRequestLimit = parseIntJava(getOptionalProperty(props, "timeoutRequestLimit", "0"), 0)
	c.RequestLimitPeriodInSeconds = parseIntJava(getOptionalProperty(props, "requestLimitPeriodInSeconds", "0"), 0)
	c.IPFingerprintFactor = parseIntJava(getOptionalProperty(props, "ipFingerprintFactor", "1"), 1)
	c.MaxWorkQueueSize = parseIntJava(getOptionalProperty(props, "maxWorkQueueSize", "0"), 0)
	if c.MaxWorkQueueSize < 0 {
		panic("maxWorkQueueSize must be >= 0: " + strconv.Itoa(c.MaxWorkQueueSize))
	}
	c.PipelineCaching = parseBoolJava(getOptionalProperty(props, "pipelineCaching", "false"))
	c.PipelinePrewarming = parseBoolJava(getOptionalProperty(props, "pipelinePrewarming", "false"))
	c.MaxPipelinePoolSize = parseIntJava(getOptionalProperty(props, "maxPipelinePoolSize", "5"), 5)
	c.PipelineExpireTime = parseIntJava(getOptionalProperty(props, "pipelineExpireTimeInSeconds", "10"), 10)
	c.TrustXForwardedForHeader = parseBoolJava(getOptionalProperty(props, "trustXForwardForHeader", "false"))
	c.MinPort = parseIntJava(getOptionalProperty(props, "minPort", "0"), 0)
	c.MaxPort = parseIntJava(getOptionalProperty(props, "maxPort", "0"), 0)
	c.ServerURL = getOptionalProperty(props, "serverURL", "")
	c.MaxCheckThreads = parseIntJava(getOptionalProperty(props, "maxCheckThreads", "10"), 10)
	if c.MaxCheckThreads < 1 {
		panic("Invalid value for maxCheckThreads, must be >= 1: " + strconv.Itoa(c.MaxCheckThreads))
	}
	c.MaxTextCheckerThreads = parseIntJava(getOptionalProperty(props, "maxTextCheckerThreads", "0"), 0)
	c.TextCheckerQueueSize = parseIntJava(getOptionalProperty(props, "textCheckerQueueSize", "8"), 8)
	c.CacheSize = parseIntJava(getOptionalProperty(props, "cacheSize", "0"), 0)
	c.CacheTTLSeconds = parseInt64Java(getOptionalProperty(props, "cacheTTLSeconds", "300"), 300)
	c.MaxErrorsPerWordRate = parseFloatJava(getOptionalProperty(props, "maxErrorsPerWordRate", "0"), 0)
	c.SuggestionsEnabled = parseBoolJava(getOptionalProperty(props, "suggestionsEnabled", "true"))
	c.MaxSpellingSuggestions = parseIntJava(getOptionalProperty(props, "maxSpellingSuggestions", "0"), 0)
	c.BlockedReferrers = splitCommaOptionalASCIIWS(getOptionalProperty(props, "blockedReferrers", ""))
	c.PremiumAlways = parseBoolJava(getOptionalProperty(props, "premiumAlways", "false"))
	c.PremiumOnly = parseBoolJava(getOptionalProperty(props, "premiumOnly", "false"))
	c.AnonymousAccessAllowed = parseBoolJava(getOptionalProperty(props, "anonymousAccessAllowed", "true"))
	c.PrometheusMonitoring = parseBoolJava(getOptionalProperty(props, "prometheusMonitoring", "false"))
	c.PrometheusPort = parseIntJava(getOptionalProperty(props, "prometheusPort", "9301"), 9301)
	c.DisabledRuleIDs = splitCommaOptionalASCIIWS(getOptionalProperty(props, "disabledRuleIds", ""))
	c.LocalAPIMode = parseBoolJava(getOptionalProperty(props, "localApiMode", "false"))
	c.MotherTongue = getOptionalProperty(props, "motherTongue", "en-US")
	pref := strings.ReplaceAll(getOptionalProperty(props, "preferredLanguages", ""), " ", "")
	if pref != "" {
		c.PreferredLanguages = strings.Split(pref, ",")
	} else {
		c.PreferredLanguages = nil
	}
	if mode := getOptionalProperty(props, "mode", "LanguageTool"); strings.EqualFold(mode, "AfterTheDeadline") {
		panic("The AfterTheDeadline mode is not supported anymore in LanguageTool 3.8 or later")
	}
	// paths that validate
	ftModel := getOptionalProperty(props, "fasttextModel", "")
	ftBin := getOptionalProperty(props, "fasttextBinary", "")
	if ftModel != "" && ftBin != "" {
		if err := c.SetFasttextPaths(ftModel, ftBin); err != nil {
			panic(err.Error())
		}
	}
	if ngram := getOptionalProperty(props, "ngramLangIdentData", ""); ngram != "" {
		if err := c.SetNgramLangIdentData(ngram); err != nil {
			panic(err.Error())
		}
	}
	if lm := getOptionalProperty(props, "languageModel", ""); lm != "" {
		c.LanguageModelDir = lm
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

// SetFasttextPaths ports HTTPServerConfig.setFasttextPaths(model, binary).
// Validates files exist; binary must be executable (Java canExecute).
func (c *HTTPServerConfig) SetFasttextPaths(fasttextModelPath, fasttextBinaryPath string) error {
	if c == nil {
		return NewIllegalConfigurationError("nil config")
	}
	if err := validateFasttextModelPath(fasttextModelPath); err != nil {
		return err
	}
	if err := validateFasttextBinaryPath(fasttextBinaryPath); err != nil {
		return err
	}
	c.FasttextModel = fasttextModelPath
	c.FasttextBinary = fasttextBinaryPath
	return nil
}

// SetNgramLangIdentData ports ngramLangIdentData property load:
// must exist and must be a file (ZIP), not a directory.
func (c *HTTPServerConfig) SetNgramLangIdentData(path string) error {
	if c == nil {
		return NewIllegalConfigurationError("nil config")
	}
	if path == "" {
		c.NgramLangIdentData = ""
		return nil
	}
	st, err := os.Stat(path)
	if err != nil || st.IsDir() {
		return NewIllegalConfigurationError(
			"ngramLangIdentData does not exist or is a directory (needs to be a ZIP file): " + path)
	}
	c.NgramLangIdentData = path
	return nil
}

func validateFasttextModelPath(path string) error {
	st, err := os.Stat(path)
	if err != nil || st.IsDir() {
		return fmt.Errorf("Fasttext model path not valid (file doesn't exist or is a directory): %s", path)
	}
	return nil
}

func validateFasttextBinaryPath(path string) error {
	st, err := os.Stat(path)
	if err != nil || st.IsDir() {
		return fmt.Errorf("Fasttext binary path not valid (file doesn't exist, is a directory or not executable): %s", path)
	}
	// Java File.canExecute(); Unix: any execute bit.
	if st.Mode()&0o111 == 0 {
		return fmt.Errorf("Fasttext binary path not valid (file doesn't exist, is a directory or not executable): %s", path)
	}
	return nil
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
		// Loaded via split(",\\s*") already; do not Unicode-trim entries.
		if b == "" {
			continue
		}
		if strings.Contains(referer, b) {
			return true
		}
	}
	return false
}

// splitCommaOptionalASCIIWS ports Java String.split(",\\s*") without UNICODE_CHARACTER_CLASS
// (HTTPServerConfig disabledRuleIds / blockedReferrers / whitelist users).
func splitCommaOptionalASCIIWS(s string) []string {
	if s == "" {
		return nil
	}
	// Manual scan: split on ',' then skip following ASCII whitespace.
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] != ',' {
			continue
		}
		out = append(out, s[start:i])
		j := i + 1
		for j < len(s) {
			c := s[j]
			if c != ' ' && c != '\t' && c != '\n' && c != '\v' && c != '\f' && c != '\r' {
				break
			}
			j++
		}
		start = j
		i = j - 1
	}
	out = append(out, s[start:])
	// Java limit 0 drops trailing empties
	for len(out) > 0 && out[len(out)-1] == "" {
		out = out[:len(out)-1]
	}
	return out
}
