package language

func NewPolish() struct {
	ShortCode, Name, SpellerRuleID string
	Countries                      []string
} {
	return Polish
}

func PolishShortCode() string { return Polish.ShortCode }
