package language

// SerbianBasicRelevantRuleIDs ports Serbian.getBasicRules shared IDs.
func SerbianBasicRelevantRuleIDs() []string {
	return []string{
		"COMMA_PARENTHESIS_WHITESPACE",
		"DOUBLE_PUNCTUATION",
		"UNPAIRED_BRACKETS",
		"UPPERCASE_SENTENCE_START",
		"WHITESPACE_RULE",
		"SENTENCE_WHITESPACE",
		"WORD_REPEAT_RULE",
	}
}

// SerbianEkavianRelevantRuleIDs ports Serbian.getRelevantRules (Ekavian default).
func SerbianEkavianRelevantRuleIDs() []string {
	out := append([]string(nil), SerbianBasicRelevantRuleIDs()...)
	out = append(out,
		"MORFOLOGIK_RULE_SR_EKAVIAN",
		"SR_EKAVIAN_SIMPLE_GRAMMAR_REPLACE_RULE",
		"SR_EKAVIAN_SIMPLE_STYLE_REPLACE_RULE",
	)
	return out
}

// SerbianJekavianRelevantRuleIDs ports JekavianSerbian.getRelevantRules.
func SerbianJekavianRelevantRuleIDs() []string {
	out := append([]string(nil), SerbianBasicRelevantRuleIDs()...)
	out = append(out,
		"MORFOLOGIK_RULE_SR_JEKAVIAN",
		"SR_JEKAVIAN_SIMPLE_GRAMMAR_REPLACE_RULE",
		"SR_JEKAVIAN_SIMPLE_STYLE_REPLACE_RULE",
	)
	return out
}

// GetRelevantRuleIDs ports Serbian / JekavianSerbian.getRelevantRules.
// Base Serbian and SerbianSerbian use Ekavian rules; Jekavian* use Jekavian rules.
func (s Serbian) GetRelevantRuleIDs() []string {
	if s.Jekavian {
		return SerbianJekavianRelevantRuleIDs()
	}
	return SerbianEkavianRelevantRuleIDs()
}
