package languagetool

// Premium describes premium-only rule gating (ports org.languagetool.Premium).
type Premium interface {
	IsPremiumRule(ruleID string) bool
}

// PremiumOff is the open-source / non-premium implementation.
type PremiumOff struct{}

func (PremiumOff) IsPremiumRule(ruleID string) bool { return false }

// IsTempNotPremium always false until a temp-not-premium list is configured.
func IsTempNotPremium(ruleID string) bool {
	for _, id := range tempNotPremiumRules {
		if id == ruleID {
			return true
		}
	}
	return false
}

var tempNotPremiumRules []string

// SetTempNotPremiumRules configures the temporary non-premium allowlist (tests).
func SetTempNotPremiumRules(ids []string) {
	tempNotPremiumRules = append([]string(nil), ids...)
}

// PremiumStatusCheckText is the well-known premium probe text from Java.
const (
	PremiumStatusCheckText  = "languagetool testrule 8634756"
	PremiumStatusCheckText2 = "The languagetool testrule 8634756."
)

// IsPremiumStatusCheck ports Premium.isPremiumStatusCheck for plain text.
func IsPremiumStatusCheck(originalText string) bool {
	return originalText == PremiumStatusCheckText || originalText == PremiumStatusCheckText2
}

// DefaultPremium is the process-wide premium gate (defaults to PremiumOff).
var DefaultPremium Premium = PremiumOff{}

// IsPremiumVersion reports whether a premium implementation is active.
// Open-source Go port defaults to false.
func IsPremiumVersion() bool {
	_, ok := DefaultPremium.(PremiumOff)
	return !ok && DefaultPremium != nil
}

// Experimental marks APIs that may change (annotation port).
const Experimental = true
