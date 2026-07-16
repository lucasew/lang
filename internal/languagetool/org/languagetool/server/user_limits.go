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
