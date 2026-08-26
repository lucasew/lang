package server

// UserLimits ports org.languagetool.server.UserLimits (no DB; config-driven).
type UserLimits struct {
	MaxTextLength       int
	MaxCheckTimeMillis  int64
	HasPremium          bool
	PremiumUID          *int64
	DictionaryCacheSize *int64
	RequestsPerDay      *int64
	LimitEnforcement    LimitEnforcementMode
	SkipLimits          bool
	JWT                 JwtContent
}

// DefaultUserLimits builds anonymous or premium-always limits from config.
func DefaultUserLimits(cfg *HTTPServerConfig) *UserLimits {
	if cfg == nil {
		cfg = NewHTTPServerConfig()
	}
	if cfg.PremiumAlways {
		uid := int64(1)
		return &UserLimits{
			MaxTextLength:      cfg.MaxTextLengthPremium,
			MaxCheckTimeMillis: cfg.MaxCheckTimeMillisAnonymous,
			HasPremium:         true,
			PremiumUID:         &uid,
			LimitEnforcement:   LimitEnforcementDisabled,
			JWT:                JwtNone,
		}
	}
	return &UserLimits{
		MaxTextLength:      cfg.MaxTextLengthAnonymous,
		MaxCheckTimeMillis: cfg.MaxCheckTimeMillisAnonymous,
		HasPremium:         false,
		LimitEnforcement:   LimitEnforcementDisabled,
		JWT:                JwtNone,
	}
}

// NewUserLimits constructs limits for a known account tier.
func NewUserLimits(maxText int, maxCheckMs int64, premiumUID *int64, hasPremium bool) *UserLimits {
	return &UserLimits{
		MaxTextLength:      maxText,
		MaxCheckTimeMillis: maxCheckMs,
		PremiumUID:         premiumUID,
		HasPremium:         hasPremium,
		LimitEnforcement:   LimitEnforcementDisabled,
		JWT:                JwtNone,
	}
}

// GetLimitsWithJwtToken ports UserLimits.getLimitsWithJwtToken.
// Without a real JWT library/secret, empty or invalid tokens yield anonymous defaults
// with JwtNone (Java open-source path without premium config).
func GetLimitsWithJwtToken(cfg *HTTPServerConfig, token, username, apiKey string) *UserLimits {
	_ = token
	_ = username
	_ = apiKey
	lim := DefaultUserLimits(cfg)
	if lim == nil {
		return nil
	}
	// invalid/empty JWT → JwtNone; do not grant premium from unverified token
	lim.JWT = JwtNone
	return lim
}

// GetUserLimits ports ServerTools.getUserLimits(params, config, authHeader).
// Auth header / tokenV2 without a configured secret stay on anonymous defaults.
func GetUserLimits(params map[string]string, cfg *HTTPServerConfig, authHeader string) *UserLimits {
	_ = authHeader
	if params != nil {
		if token := params["token"]; token != "" {
			return GetLimitsWithJwtToken(cfg, token, params["username"], params["apiKey"])
		}
		if token := params["tokenV2"]; token != "" {
			return GetLimitsWithJwtToken(cfg, token, params["username"], params["apiKey"])
		}
	}
	return DefaultUserLimits(cfg)
}

// GetAuthHeader extracts a Bearer token from an Authorization header value.
func GetAuthHeader(authorization string) string {
	const prefix = "Bearer "
	const prefixColon = "Bearer: "
	if len(authorization) > len(prefixColon) && (authorization[:len(prefixColon)] == prefixColon || authorization[:7] == "Bearer:") {
		// "Bearer: token" form used in JwtTest
		i := 0
		for i < len(authorization) && authorization[i] != ' ' && authorization[i] != ':' {
			i++
		}
		for i < len(authorization) && (authorization[i] == ' ' || authorization[i] == ':') {
			i++
		}
		return authorization[i:]
	}
	if len(authorization) > len(prefix) && authorization[:len(prefix)] == prefix {
		return authorization[len(prefix):]
	}
	return authorization
}
