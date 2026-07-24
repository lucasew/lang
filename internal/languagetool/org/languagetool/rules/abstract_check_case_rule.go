package rules

// AbstractCheckCaseRule ports org.languagetool.rules.AbstractCheckCaseRule —
// an AbstractSimpleReplaceRule2 that checks case (ITS typographical / CASING).
// Construct via NewAbstractCheckCaseRule after loading replacement data.
//
// Java ctor: super(messages, language); setLocQualityIssueType(Typographical);
// setCategory(Categories.CASING.getCategory(messages)); isCheckingCase() → true.
func NewAbstractCheckCaseRule(messages map[string]string, id, description string) *AbstractSimpleReplaceRule2 {
	r := &AbstractSimpleReplaceRule2{
		Messages:     messages,
		ID:           id,
		Description:  description,
		CheckingCase: true,
		IssueType:    ITSTypographical,
		Category:     CatCasing.GetCategory(messages),
	}
	return r
}
