package languagetool

import "strings"

// SoftRuleMeta assigns category and issue type for well-known rule ID families.
func SoftRuleMeta(ruleID string) (categoryID, categoryName, issueType, short string) {
	id := strings.ToUpper(ruleID)
	switch {
	case strings.Contains(id, "MORFOLOGIK") || strings.Contains(id, "HUNSPELL") || strings.Contains(id, "SPELL"):
		return "TYPOS", "Possible Typo", "misspelling", "Spelling mistake"
	case strings.Contains(id, "WHITESPACE") || strings.Contains(id, "PUNCT") || id == "COMMA_WHITESPACE" ||
		id == "DOUBLE_PUNCTUATION" || id == "SENTENCE_WHITESPACE":
		return "TYPOGRAPHY", "Typography", "whitespace", "Typography"
	case strings.Contains(id, "WORD_REPEAT"):
		return "MISC", "Miscellaneous", "duplication", "Word repetition"
	case id == "EN_A_VS_AN" || strings.Contains(id, "A_VS_AN"):
		return "GRAMMAR", "Grammar", "grammar", "Wrong article"
	case strings.Contains(id, "UNPAIRED") || strings.Contains(id, "BRACKET"):
		return "TYPOGRAPHY", "Typography", "typographical", "Unpaired symbol"
	case strings.Contains(id, "UPPERCASE") || strings.Contains(id, "SENTENCE_START"):
		return "CASING", "Capitalization", "typographical", "Capitalization"
	case strings.Contains(id, "LONG_SENTENCE") || strings.Contains(id, "TOO_LONG"):
		return "STYLE", "Style", "style", "Long sentence"
	case strings.Contains(id, "LONG_PARAGRAPH"):
		return "STYLE", "Style", "style", "Long paragraph"
	case strings.Contains(id, "FALSE_FRIEND"):
		return "FALSEFRIENDS", "False Friends", "misspelling", "False friend"
	// Soft grammar XML slices (e.g. EN_SOFT_*) are grammar, not false friends.
	case strings.Contains(id, "_SOFT_"):
		return "GRAMMAR", "Grammar", "grammar", SoftRuleDescription(ruleID)
	case strings.HasPrefix(id, "EN_") && (strings.Contains(id, "_OF") || strings.Contains(id, "A_LOT") || strings.Contains(id, "IRREGARDLESS")):
		return "GRAMMAR", "Grammar", "grammar", "Possible grammar error"
	default:
		if ruleID == "" {
			return "MISC", "Miscellaneous", "uncategorized", ""
		}
		return "MISC", "Miscellaneous", "uncategorized", ""
	}
}

// SoftRuleDescription returns a stable rule-level description (not the match message).
func SoftRuleDescription(ruleID string) string {
	id := strings.ToUpper(ruleID)
	switch {
	case id == "EN_A_VS_AN" || strings.Contains(id, "A_VS_AN"):
		return "Use of 'a' versus 'an'"
	case strings.Contains(id, "WORD_REPEAT"):
		return "Word repetition"
	case strings.Contains(id, "MORFOLOGIK") || strings.Contains(id, "HUNSPELL") || strings.Contains(id, "SPELL"):
		return "Possible spelling mistake"
	case strings.Contains(id, "WHITESPACE") || id == "COMMA_WHITESPACE" || id == "SENTENCE_WHITESPACE":
		return "Whitespace"
	case strings.Contains(id, "UNPAIRED") || strings.Contains(id, "BRACKET"):
		return "Unpaired brackets"
	case strings.Contains(id, "UPPERCASE") || strings.Contains(id, "SENTENCE_START"):
		return "Capitalization"
	case strings.Contains(id, "LONG_SENTENCE") || strings.Contains(id, "TOO_LONG"):
		return "Long sentence"
	case strings.Contains(id, "LONG_PARAGRAPH"):
		return "Long paragraph"
	case strings.Contains(id, "FALSE_FRIEND"):
		return "False friend"
	case strings.Contains(id, "_SOFT_"):
		if i := strings.Index(id, "_SOFT_"); i >= 0 && i+6 < len(id) {
			return strings.ReplaceAll(id[i+6:], "_", " ")
		}
		return ruleID
	case ruleID == "":
		return ""
	default:
		return ruleID
	}
}

// SeverityFromIssueType maps ITS issue types to SARIF 2.1 result levels (SPEC §2.2).
func SeverityFromIssueType(issueType string) string {
	switch strings.ToLower(strings.TrimSpace(issueType)) {
	case "misspelling", "grammar":
		return "error"
	case "style", "register", "locale-violation", "locale-specific-content":
		return "note"
	case "":
		return "warning"
	default:
		// whitespace, typographical, duplication, uncategorized, …
		return "warning"
	}
}
