package language

func NewRussian() struct {
	ShortCode, Name, SpellerRuleID string
	Countries                      []string
} {
	return Russian
}

func RussianShortCode() string { return Russian.ShortCode }
