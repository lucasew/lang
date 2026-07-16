package server

// LimitEnforcementMode ports org.languagetool.server.LimitEnforcementMode.
type LimitEnforcementMode int

const (
	LimitEnforcementDisabled LimitEnforcementMode = 1
	LimitEnforcementPerDay   LimitEnforcementMode = 2
)

// ParseLimitEnforcementMode maps an integer id; unknown/null → Disabled.
func ParseLimitEnforcementMode(value *int) LimitEnforcementMode {
	if value == nil || *value <= 0 {
		return LimitEnforcementDisabled
	}
	switch LimitEnforcementMode(*value) {
	case LimitEnforcementDisabled, LimitEnforcementPerDay:
		return LimitEnforcementMode(*value)
	default:
		return LimitEnforcementDisabled
	}
}

func (m LimitEnforcementMode) ID() int { return int(m) }
