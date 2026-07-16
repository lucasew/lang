package rules

// AbstractCheckCaseRule ports org.languagetool.rules.AbstractCheckCaseRule —
// an AbstractSimpleReplaceRule2 that checks case (ITS typographical / CASING).
// Construct via NewAbstractCheckCaseRule after loading replacement data.
func NewAbstractCheckCaseRule(id, description string) *AbstractSimpleReplaceRule2 {
	r := &AbstractSimpleReplaceRule2{
		ID:           id,
		Description:  description,
		CheckingCase: true,
	}
	return r
}
