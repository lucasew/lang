package language

// KhmerRelevantRuleIDs ports Khmer.getRelevantRules rule IDs (class getId only).
func KhmerRelevantRuleIDs() []string {
	return []string{
		"HUNSPELL_RULE", // KhmerHunspellRule extends HunspellRule
		"KM_SIMPLE_REPLACE",
		"KM_WORD_REPEAT_RULE",
		"KM_UNPAIRED_BRACKETS",
		"KM_SPACE_BEFORE_CONJUNCTION",
	}
}
