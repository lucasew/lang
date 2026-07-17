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
	case id == "EMPTY_LINE":
		// Java EmptyLineRule: Categories.STYLE + ITSIssueType.Style
		return "STYLE", "Style", "style", "Empty line"
	case strings.Contains(id, "WORD_REPEAT"):
		return "MISC", "Miscellaneous", "duplication", "Word repetition"
	case id == "EN_A_VS_AN" || strings.Contains(id, "A_VS_AN"):
		return "GRAMMAR", "Grammar", "grammar", "Wrong article"
	case id == "PHRASE_REPLACE" || strings.Contains(id, "PHRASE_REPLACE") || strings.HasSuffix(id, "_OF"):
		// EN_COULD_OF family + phrase injects
		return "GRAMMAR", "Grammar", "grammar", "Possible grammar error"
	case strings.Contains(id, "UNPAIRED") || strings.Contains(id, "BRACKET"):
		return "TYPOGRAPHY", "Typography", "typographical", "Unpaired symbol"
	case strings.Contains(id, "UPPERCASE") || strings.Contains(id, "SENTENCE_START"):
		return "CASING", "Capitalization", "typographical", "Capitalization"
	case strings.Contains(id, "LONG_SENTENCE") || strings.Contains(id, "TOO_LONG"):
		return "STYLE", "Style", "style", "Long sentence"
	case strings.Contains(id, "LONG_PARAGRAPH"):
		return "STYLE", "Style", "style", "Long paragraph"
	case strings.Contains(id, "FALSE_FRIEND") || isSoftFalseFriendGroupID(id):
		return "FALSEFRIENDS", "False Friends", "misspelling", "False friend"
	// Soft grammar XML slices (e.g. EN_SOFT_*) — refine issue type from ID shape.
	case strings.Contains(id, "_SOFT_"):
		if strings.Contains(id, "DOUBLE_BANG") || strings.Contains(id, "DOUBLE_Q") ||
			strings.Contains(id, "TYPOGRAPHY") || strings.Contains(id, "SPACE_BEFORE") {
			return "TYPOGRAPHY", "Typography", "typographical", SoftRuleDescription(ruleID)
		}
		// Regional soft packs: …_US / …_GB / …_BR / …_MX / …
		if softRegionalTypoID(id) {
			return "TYPOS", "Possible Typo", "misspelling", SoftRuleDescription(ruleID)
		}
		if strings.Contains(id, "CASING") || strings.Contains(id, "LOWERCASE_I") ||
			strings.Contains(id, "UPPERCASE") || strings.HasSuffix(id, "_LOWER_I") ||
			strings.Contains(id, "SENTENCE_START") || strings.Contains(id, "CASE_SENSITIVE") {
			return "CASING", "Capitalization", "typographical", SoftRuleDescription(ruleID)
		}
		if softStyleID(id) {
			return "STYLE", "Style", "style", SoftRuleDescription(ruleID)
		}
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
	case id == "EMPTY_LINE":
		return "Empty line"
	case strings.Contains(id, "UNPAIRED") || strings.Contains(id, "BRACKET"):
		return "Unpaired brackets"
	case strings.Contains(id, "UPPERCASE") || strings.Contains(id, "SENTENCE_START"):
		return "Capitalization"
	case strings.Contains(id, "LONG_SENTENCE") || strings.Contains(id, "TOO_LONG"):
		return "Long sentence"
	case strings.Contains(id, "LONG_PARAGRAPH"):
		return "Long paragraph"
	case strings.Contains(id, "FALSE_FRIEND") || isSoftFalseFriendGroupID(id):
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

// SoftRuleURL returns a community rule page URL for a rule ID (soft documentation link).
// When lang is empty, SoftRuleLangHint is used (e.g. DE_SOFT_* → de), then "en".
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
	// strip variant
	if i := strings.IndexByte(lang, '-'); i > 0 {
		lang = lang[:i]
	}
	return "https://community.languagetool.org/rule/show/" + ruleID + "?lang=" + lang
}

// isSoftFalseFriendGroupID matches soft false-friends-soft.xml rulegroup ids.
func isSoftFalseFriendGroupID(id string) bool {
	switch id {
	case "ABILITY", "GIFT", "ACTUAL", "LIBRARY", "EVENTUAL", "BECOME",
		"EMBARRASSED", "PARENTS", "SYMPATHIC", "FABRIC", "ARGUMENT",
		"SENSIBLE", "CONSTIPATED", "PRESERVATIVE", "EVENTUALLY",
		"ROMAN", "CARPET", "ASSIST", "DECEPTION", "BRAVE", "MIST",
		"CHEMIST", "PREFIX", "COLLEGE", "LOCATION", "LECTURE", "FIGURE",
		"EXIT", "CONSTIPATION":
		return true
	default:
		return false
	}
}

// softRegionalTypoID is true for regional soft pack rules (…_US, …_GB, …_BR, …).
func softRegionalTypoID(id string) bool {
	for _, suf := range []string{"_US", "_GB", "_BR", "_PT", "_MX", "_ES", "_CH", "_AT", "_CA"} {
		if strings.HasSuffix(id, suf) {
			return true
		}
	}
	return false
}

// softStyleID classifies soft redundancy / informal-style rule IDs.
func softStyleID(id string) bool {
	if strings.Contains(id, "_STYLE") || strings.Contains(id, "KIND_OF") ||
		strings.Contains(id, "LITERALLY") || strings.Contains(id, "VERY_UNIQUE") ||
		strings.Contains(id, "IN_ORDER_TO") || strings.Contains(id, "DUE_TO_THE_FACT") ||
		strings.Contains(id, "POINT_IN_TIME") || strings.Contains(id, "IN_THE_EVENT") ||
		strings.Contains(id, "END_RESULT") || strings.Contains(id, "PAST_HISTORY") ||
		strings.Contains(id, "FREE_GIFT") || strings.Contains(id, "COMPLETELY_ELIMINATE") ||
		strings.Contains(id, "DIFFERENT_THAN") || strings.Contains(id, "EACH_AND_EVERY") ||
		strings.Contains(id, "FIRST_AND_FOREMOST") || strings.Contains(id, "BASIC_FUNDAMENTALS") ||
		strings.Contains(id, "GOES_WITHOUT_SAYING") || strings.Contains(id, "THESE_ONES") ||
		strings.Contains(id, "REASON_IS_BECAUSE") || strings.Contains(id, "WHETHER_OR_NOT") ||
		strings.Contains(id, "ACTUAL_FACT") || strings.Contains(id, "TRUE_FACT") ||
		strings.Contains(id, "ADVANCE_PLANNING") || strings.Contains(id, "CLOSE_PROXIMITY") ||
		strings.Contains(id, "FUTURE_PLANS") || strings.Contains(id, "UNEXPECTED_SURPRISE") ||
		strings.Contains(id, "REVERT_BACK") || strings.Contains(id, "REPEAT_AGAIN") ||
		strings.Contains(id, "FINAL_OUTCOME") || strings.Contains(id, "GENERAL_CONSENSUS") ||
		strings.Contains(id, "PERSONAL_OPINION") || strings.Contains(id, "COMPLETE_STOP") ||
		strings.Contains(id, "ABSOLUTELY_ESSENTIAL") || strings.Contains(id, "EXACTLY_THE_SAME") ||
		strings.Contains(id, "CURRENTLY_IN_PROGRESS") || strings.Contains(id, "ADDED_BONUS") ||
		strings.Contains(id, "BRIEF_MOMENT") || strings.Contains(id, "JOIN_TOGETHER") ||
		strings.Contains(id, "PLAN_AHEAD") || strings.Contains(id, "STILL_REMAINS") ||
		strings.Contains(id, "CIRCLE_AROUND") || strings.Contains(id, "RETURN_BACK") ||
		strings.Contains(id, "GOTTA") || strings.Contains(id, "WANNA") ||
		strings.Contains(id, "GONNA") || strings.Contains(id, "PROLLY") ||
		strings.Contains(id, "DEFFO") || strings.Contains(id, "BASICALLY") ||
		strings.Contains(id, "ACTUALLY_ACTUALLY") || strings.Contains(id, "HONESTLY_HONESTLY") ||
		strings.Contains(id, "REALLY_REALLY") || strings.Contains(id, "VERY_VERY") ||
		strings.Contains(id, "JUST_JUST") || strings.Contains(id, "THE_THE") ||
		strings.Contains(id, "AND_AND") || strings.Contains(id, "OF_OF") ||
		strings.Contains(id, "TO_TO") || strings.Contains(id, "IN_IN") ||
		strings.Contains(id, "ON_ON") || strings.Contains(id, "FOR_FOR") ||
		strings.Contains(id, "WITH_WITH") || strings.Contains(id, "A_A") ||
		strings.Contains(id, "SO_SO") || strings.Contains(id, "IRREGARDLESS") ||
		strings.Contains(id, "SUPPOSABLY") || strings.Contains(id, "ANYWAYS") ||
		strings.Contains(id, "BEGS_THE_QUESTION") {
		return true
	}
	return false
}

// SoftRuleLangHint infers a language code from a rule ID prefix (soft fallback).
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
