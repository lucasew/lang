package languagetool

import "strings"

// SoftRuleMeta is a fallback when LocalMatch has no CategoryID/IssueType.
// It maps well-known Java rule ID families to Categories / ITS types used in LT
// (e.g. Morfologik → TYPOS/misspelling). It must not invent metadata for soft packs.
// Prefer setting CategoryID/IssueType on LocalMatch from the rule itself.
func SoftRuleMeta(ruleID string) (categoryID, categoryName, issueType, short string) {
	id := strings.ToUpper(ruleID)
	switch {
	case strings.Contains(id, "MORFOLOGIK") || strings.Contains(id, "HUNSPELL") ||
		(strings.Contains(id, "SPELL") && !strings.Contains(id, "IGNORE_SPELLING")):
		// Java SpellingCheckRule / MorfologikSpellerRule: Categories.TYPOS, ITS misspelling
		return "TYPOS", "Possible Typo", "misspelling", "Spelling mistake"
	case strings.Contains(id, "WHITESPACE") || strings.Contains(id, "PUNCT") ||
		id == "COMMA_WHITESPACE" || id == "DOUBLE_PUNCTUATION" || id == "SENTENCE_WHITESPACE":
		return "TYPOGRAPHY", "Typography", "whitespace", "Typography"
	case id == "EMPTY_LINE":
		// Java EmptyLineRule: Categories.STYLE + ITSIssueType.Style
		return "STYLE", "Style", "style", "Empty line"
	case strings.Contains(id, "WORD_REPEAT"):
		return "MISC", "Miscellaneous", "duplication", "Word repetition"
	case id == "EN_A_VS_AN" || strings.Contains(id, "A_VS_AN"):
		return "GRAMMAR", "Grammar", "grammar", "Wrong article"
	case strings.Contains(id, "UNPAIRED") || strings.Contains(id, "BRACKET"):
		return "TYPOGRAPHY", "Typography", "typographical", "Unpaired symbol"
	case strings.Contains(id, "UPPERCASE_SENTENCE_START") || id == "UPPERCASE_SENTENCE_START":
		return "CASING", "Capitalization", "typographical", "Capitalization"
	case strings.Contains(id, "LONG_SENTENCE") || strings.Contains(id, "TOO_LONG"):
		return "STYLE", "Style", "style", "Long sentence"
	case strings.Contains(id, "LONG_PARAGRAPH"):
		return "STYLE", "Style", "style", "Long paragraph"
	case strings.Contains(id, "FALSE_FRIEND"):
		// Java FalseFriendRule: Categories.FALSE_FRIENDS
		return "FALSEFRIENDS", "False Friends", "misspelling", "False friend"
	default:
		if ruleID == "" {
			return "MISC", "Miscellaneous", "uncategorized", ""
		}
		// Unknown rule: uncategorized — do not invent grammar/style from ID shape.
		return "MISC", "Miscellaneous", "uncategorized", ""
	}
}

// SoftRuleDescription returns a short description for known Java rule families.
// Prefer rule.GetDescription() when available; this is CLI/API fallback only.
func SoftRuleDescription(ruleID string) string {
	id := strings.ToUpper(ruleID)
	switch {
	case id == "EN_A_VS_AN" || strings.Contains(id, "A_VS_AN"):
		return "Use of 'a' versus 'an'"
	case strings.Contains(id, "WORD_REPEAT"):
		return "Word repetition"
	case strings.Contains(id, "MORFOLOGIK") || strings.Contains(id, "HUNSPELL") ||
		(strings.Contains(id, "SPELL") && !strings.Contains(id, "IGNORE_SPELLING")):
		return "Possible spelling mistake"
	case strings.Contains(id, "WHITESPACE") || id == "COMMA_WHITESPACE" || id == "SENTENCE_WHITESPACE":
		return "Whitespace"
	case id == "EMPTY_LINE":
		return "Empty line"
	case strings.Contains(id, "UNPAIRED") || strings.Contains(id, "BRACKET"):
		return "Unpaired brackets"
	case strings.Contains(id, "UPPERCASE_SENTENCE_START"):
		return "Capitalization"
	case strings.Contains(id, "LONG_SENTENCE") || strings.Contains(id, "TOO_LONG"):
		return "Long sentence"
	case strings.Contains(id, "LONG_PARAGRAPH"):
		return "Long paragraph"
	case strings.Contains(id, "FALSE_FRIEND"):
		return "False friend"
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

// SoftRuleURL returns the LanguageTool community rule page for a rule ID.
// Product help link (not soft invent). lang defaults via SoftRuleLangHint then "en".
func SoftRuleURL(ruleID, lang string) string {
	if ruleID == "" {
		return ""
	}
	if lang == "" {
		lang = SoftRuleLangHint(ruleID)
	}
	if lang == "" {
		lang = "en"
	}
	if i := strings.IndexByte(lang, '-'); i > 0 {
		lang = lang[:i]
	}
	return "https://community.languagetool.org/rule/show/" + ruleID + "?lang=" + lang
}

// SoftRuleLangHint infers a language code from a rule ID prefix (e.g. DE_… → de).
// Only known LT language codes; empty if unknown.
func SoftRuleLangHint(ruleID string) string {
	up := strings.ToUpper(strings.TrimSpace(ruleID))
	i := strings.IndexByte(up, '_')
	if i < 2 || i > 3 {
		return ""
	}
	p := strings.ToLower(up[:i])
	switch p {
	case "en", "de", "fr", "es", "pt", "it", "nl", "pl", "ru", "uk", "sv", "da",
		"ca", "gl", "sk", "ro", "el", "ar", "fa", "ga", "br", "eo", "sl", "sr",
		"be", "is", "ja", "km", "lt", "ml", "ta", "tl", "zh", "ast", "crh":
		return p
	default:
		return ""
	}
}
