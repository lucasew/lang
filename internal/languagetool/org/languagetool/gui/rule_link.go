package gui

import "fmt"

const (
	deactivateURL = "http://languagetool.org/deactivate/"
	reactivateURL = "http://languagetool.org/reactivate/"
)

// RuleLink ports org.languagetool.gui.RuleLink.
type RuleLink struct {
	URLPrefix string
	ID        string
}

func BuildDeactivationLink(ruleID string) RuleLink {
	return RuleLink{URLPrefix: deactivateURL, ID: ruleID}
}

func BuildReactivationLink(ruleID string) RuleLink {
	return RuleLink{URLPrefix: reactivateURL, ID: ruleID}
}

func GetRuleLinkFromString(ruleLink string) (RuleLink, error) {
	if len(ruleLink) >= len(deactivateURL) && ruleLink[:len(deactivateURL)] == deactivateURL {
		return RuleLink{URLPrefix: deactivateURL, ID: ruleLink[len(deactivateURL):]}, nil
	}
	if len(ruleLink) >= len(reactivateURL) && ruleLink[:len(reactivateURL)] == reactivateURL {
		return RuleLink{URLPrefix: reactivateURL, ID: ruleLink[len(reactivateURL):]}, nil
	}
	return RuleLink{}, fmt.Errorf("unknown link prefix: %s", ruleLink)
}

func (r RuleLink) GetID() string { return r.ID }

func (r RuleLink) String() string { return r.URLPrefix + r.ID }
