package language

// Italian language twin — see more_variants.go for the Italian var.
func NewItalian() struct {
	ShortCode, Name, SpellerRuleID string
	Countries                      []string
} {
	return Italian
}

func ItalianShortCode() string { return Italian.ShortCode }
func ItalianName() string      { return Italian.Name }
