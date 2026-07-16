package rules

// SpecificIdRule ports org.languagetool.rules.SpecificIdRule — a no-op rule
// that only carries metadata (id, description, category, issue type, tags).
type SpecificIdRule struct {
	ID          string
	Description string
	Premium     bool
	Category    *Category
	IssueType   ITSIssueType
	Tags        []Tag
}

func NewSpecificIdRule(id, desc string, isPremium bool, category *Category, issueType ITSIssueType, tags []Tag) *SpecificIdRule {
	if id == "" || desc == "" {
		panic("id and desc required")
	}
	return &SpecificIdRule{
		ID:          id,
		Description: desc,
		Premium:     isPremium,
		Category:    category,
		IssueType:   issueType,
		Tags:        append([]Tag(nil), tags...),
	}
}

func (r *SpecificIdRule) GetID() string          { return r.ID }
func (r *SpecificIdRule) GetDescription() string { return r.Description }
func (r *SpecificIdRule) GetTags() []Tag         { return r.Tags }
func (r *SpecificIdRule) HasTag(tag Tag) bool {
	for _, t := range r.Tags {
		if t == tag {
			return true
		}
	}
	return false
}
func (r *SpecificIdRule) IsPremium() bool { return r.Premium }

// Match always returns no hits.
func (r *SpecificIdRule) Match(_ any) []*RuleMatch { return nil }
