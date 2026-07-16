package rules

// ConfusionString ports org.languagetool.rules.ConfusionString.
type ConfusionString struct {
	str         string
	description *string
}

func NewConfusionString(str string, description *string) *ConfusionString {
	// Java requires non-null str; empty string is allowed.
	return &ConfusionString{str: str, description: description}
}

func (c *ConfusionString) GetString() string { return c.str }

func (c *ConfusionString) GetDescription() *string { return c.description }

func (c *ConfusionString) String() string { return c.str }

func (c *ConfusionString) Equal(o *ConfusionString) bool {
	if c == o {
		return true
	}
	if c == nil || o == nil {
		return false
	}
	if c.str != o.str {
		return false
	}
	if c.description == nil && o.description == nil {
		return true
	}
	if c.description == nil || o.description == nil {
		return false
	}
	return *c.description == *o.description
}
