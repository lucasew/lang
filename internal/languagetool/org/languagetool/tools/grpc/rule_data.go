package grpc

// RuleData ports org.languagetool.tools.grpc.RuleData metadata (no rules package import).
type RuleData struct {
	ID                          string
	SubID                       string
	Description                 string
	EstimateContextForSureMatch int
	SourceFile                  string
	IssueType                   string
	TempOff                     bool
	Premium                     bool
	CategoryID                  string
	CategoryName                string
	Tags                        []string
}

func NewRuleData(id, subID, description string) *RuleData {
	return &RuleData{ID: id, SubID: subID, Description: description}
}

func (r *RuleData) GetID() string {
	if r == nil {
		return ""
	}
	return r.ID
}

func (r *RuleData) GetSubID() string {
	if r == nil {
		return ""
	}
	return r.SubID
}

func (r *RuleData) GetDescription() string {
	if r == nil {
		return ""
	}
	return r.Description
}

func (r *RuleData) GetSourceFile() string {
	if r == nil {
		return ""
	}
	return r.SourceFile
}
