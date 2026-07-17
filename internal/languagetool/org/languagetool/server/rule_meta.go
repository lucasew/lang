package server

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
	case strings.HasPrefix(id, "EN_") && (strings.Contains(id, "_OF") || strings.Contains(id, "A_LOT") || strings.Contains(id, "IRREGARDLESS")):
		return "GRAMMAR", "Grammar", "grammar", "Possible grammar error"
	case strings.Contains(id, "SOFT_") || strings.Contains(id, "FALSE_FRIEND") || strings.Contains(id, "ABILITY"):
		return "FALSEFRIENDS", "False Friends", "misspelling", "False friend"
	default:
		if ruleID == "" {
			return "MISC", "Miscellaneous", "uncategorized", ""
		}
		return "MISC", "Miscellaneous", "uncategorized", ""
	}
}
