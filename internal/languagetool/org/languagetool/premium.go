package languagetool

// Premium describes premium-only rule gating (ports org.languagetool.Premium).
type Premium interface {
	IsPremiumRule(ruleID string) bool
}

// PremiumOff is the open-source / non-premium implementation.
type PremiumOff struct{}

func (PremiumOff) IsPremiumRule(ruleID string) bool { return false }

// tempNotPremiumRules ports Premium.tempNotPremiumRules (Java: private static final
// List from Arrays.asList() — currently empty; no public mutator).
var tempNotPremiumRules = []string{}

// IsTempNotPremium ports Premium.isTempNotPremium (contains rule id in the fixed list).
func IsTempNotPremium(ruleID string) bool {
	for _, id := range tempNotPremiumRules {
		if id == ruleID {
			return true
		}
	}
	return false
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
// Open-source Go port defaults to false (Java: Class.forName("PremiumOn") missing → false).
func IsPremiumVersion() bool {
	_, ok := DefaultPremium.(PremiumOff)
	return !ok && DefaultPremium != nil
}

// PremiumBuildInfo is the process-wide PREMIUM LtBuildInfo snapshot (Java LtBuildInfo.PREMIUM).
// Empty until git-premium.properties are loaded via LoadLtBuildInfo.
var PremiumBuildInfo = LoadLtBuildInfo("PREMIUM", nil)

// Deprecated Premium instance methods (Java Premium.getBuildDate/getShortGitId/getVersion)
// delegate to LtBuildInfo.PREMIUM. Nullability preserved via *string.

func (PremiumOff) GetBuildDate() *string  { return PremiumBuildInfo.GetBuildDate() }
func (PremiumOff) GetShortGitId() *string { return PremiumBuildInfo.GetShortGitId() }
func (PremiumOff) GetVersion() *string    { return PremiumBuildInfo.GetVersion() }
