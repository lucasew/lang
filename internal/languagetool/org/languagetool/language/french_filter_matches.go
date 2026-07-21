package language

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// frJavaTrailingWS ports Java replaceAll("\\s+$", "") without UNICODE_CHARACTER_CLASS.
var frJavaTrailingWS = regexp.MustCompile(`[ \t\n\v\f\r]+$`)

func init() {
	// Wire without rules/fr importing language (cycle via french_rules_smoke_test).
	languagetool.FilterFrenchRuleMatchesHook = FilterFrenchRuleMatches
}

// FilterFrenchRuleMatches ports French.filterRuleMatches (AI_FR_GGEC adjacent merge /
// overlap skip, trailing-period drop, adjustFrenchRuleMatch).
//
// enabledRules: LocalMatch.EnabledRules stamped by Check (Java Set enabledRules).
// Period drop needs SentenceText (Check fills it); without it, keep (fail-closed).
func FilterFrenchRuleMatches(matches []languagetool.LocalMatch) []languagetool.LocalMatch {
	if len(matches) == 0 {
		return nil
	}
	var result []languagetool.LocalMatch
	var previous *languagetool.LocalMatch

	for i := range matches {
		cur := adjustFrenchRuleMatch(matches[i], matches[i].EnabledRules)

		// Java: ignore adding punctuation at the sentence end (single-suggestion GGEC period).
		if frenchDropTrailingPeriodSuggestion(cur) {
			continue
		}

		// Java gates on getRule().getId().startsWith("AI_FR_GGEC"); merge keeps match1 rule
		// and only setSpecificRuleId(AI_FR_MERGED_MATCH*). LocalMatch RuleID is the API id,
		// so AI_FR_MERGED_MATCH* stays merge-eligible for chain merges.
		if previous != nil &&
			isFrenchGGECMergeID(previous.RuleID) &&
			isFrenchGGECMergeID(cur.RuleID) {

			if previous.ToPos > cur.FromPos {
				continue // Skip overlapping matches
			}
			adjacent := previous.ToPos == cur.FromPos || previous.ToPos+1 == cur.FromPos
			samePicky := previous.IsPicky == cur.IsPicky
			if adjacent && samePicky {
				sameITS := previous.IssueType == cur.IssueType
				if sameITS {
					merged := mergeFrenchGGECMatches(*previous, cur)
					previous = &merged
					continue
				}
				if !isStyleIssue(previous.IssueType) && !isStyleIssue(cur.IssueType) {
					merged := mergeFrenchGGECMatches(*previous, cur)
					merged.IssueType = "grammar"
					previous = &merged
					continue
				}
			}
			result = append(result, *previous)
			previous = &cur
			continue
		}

		if previous != nil {
			result = append(result, *previous)
		}
		previous = &cur
	}
	if previous != nil {
		result = append(result, *previous)
	}
	return result
}

// adjustFrenchRuleMatch ports French.adjustFrenchRuleMatch.
func adjustFrenchRuleMatch(m languagetool.LocalMatch, enabledRules map[string]struct{}) languagetool.LocalMatch {
	errorStr := m.OriginalSurface()
	if len(m.Suggestions) == 1 && strings.HasPrefix(m.RuleID, "AI_FR_GGEC") {
		sug := m.Suggestions[0]
		// Java: suggestion.equalsIgnoreCase(errorStr) — casing-only (or identical) path
		if errorStr != "" && strings.EqualFold(sug, errorStr) {
			m.Message = "Un usage différent des majuscules et des minuscules est recommandé."
			m.ShortMessage = "Majuscules et minuscules"
			m.IssueType = "typographical"
			m.CategoryID = "CASING"
			m.CategoryName = "Majuscules"
			m.RuleID = strings.ReplaceAll(m.RuleID, "ORTHOGRAPHY", "CASING")
		}
	}
	if len(m.Suggestions) == 1 && strings.HasPrefix(m.RuleID, "AI_FR_GGEC") &&
		strings.Contains(m.RuleID, "MISSING_PRONOUN_LAPOSTROPHE") {
		if errorStr == "on" && m.Suggestions[0] == "l'on" &&
			strings.Contains(strings.ToLower(m.SentenceText), "si on") {
			m.RuleID = "AI_FR_GGEC_SI_LON"
			m.IsPicky = true
		}
	}
	if strings.HasPrefix(m.RuleID, "AI_FR_GGEC") &&
		strings.Contains(m.RuleID, "REPLACEMENT_PUNCTUATION_QUOTE") {
		m.RuleID = "AI_FR_GGEC_QUOTES"
		m.IsPicky = true
		m.IssueType = "typographical"
	}
	// Java: if APOS_TYP enabled, use typographic apostrophe in suggestions.
	if enabledRules != nil {
		if _, ok := enabledRules["APOS_TYP"]; ok {
			newSugs := make([]string, len(m.Suggestions))
			for i, s := range m.Suggestions {
				if len(s) > 1 {
					s = strings.ReplaceAll(s, "'", "’")
				}
				newSugs[i] = s
			}
			m.Suggestions = newSugs
		}
	}
	return m
}

// frenchDropTrailingPeriodSuggestion ports French filter skip for
// AI_FR_GGEC_MISSING_PUNCTUATION_PERIOD when suggestion only appends ".".
func frenchDropTrailingPeriodSuggestion(m languagetool.LocalMatch) bool {
	if m.RuleID != "AI_FR_GGEC_MISSING_PUNCTUATION_PERIOD" {
		return false
	}
	if len(m.Suggestions) != 1 {
		return false
	}
	sug := m.Suggestions[0]
	if !strings.HasSuffix(sug, ".") {
		return false
	}
	sent := m.SentenceText
	if sent == "" {
		return false // fail-closed
	}
	// Java replaceAll("\\s+$", "") — ASCII \s only (not NBSP).
	trimmed := frJavaTrailingWS.ReplaceAllString(sent, "")
	prefix := sug[:len(sug)-1]
	return strings.HasSuffix(trimmed, prefix)
}

// isFrenchGGECMergeID is true for live AI_FR_GGEC* and AI_FR_MERGED_MATCH*
// (Java still has getRule().getId() starting with AI_FR_GGEC after merge).
func isFrenchGGECMergeID(id string) bool {
	return strings.HasPrefix(id, "AI_FR_GGEC") || strings.HasPrefix(id, "AI_FR_MERGED_MATCH")
}

func mergeFrenchGGECMatches(match1, match2 languagetool.LocalMatch) languagetool.LocalMatch {
	sep := ""
	if match1.ToPos+1 == match2.FromPos {
		sep = " "
	}
	var newRepl string
	if len(match1.Suggestions) > 0 && len(match2.Suggestions) > 0 {
		newRepl = match1.Suggestions[0] + sep + match2.Suggestions[0]
	} else if len(match1.Suggestions) > 0 {
		newRepl = match1.Suggestions[0]
	} else if len(match2.Suggestions) > 0 {
		newRepl = match2.Suggestions[0]
	}
	s1, s2 := match1.OriginalSurface(), match2.OriginalSurface()
	var newErr string
	if s1 != "" || s2 != "" {
		newErr = s1 + sep + s2
	}
	// Java: AI_FR_MERGED_MATCH + optional _STYLE + optional _PICKY
	newID := "AI_FR_MERGED_MATCH"
	if isStyleIssue(match1.IssueType) && isStyleIssue(match2.IssueType) {
		newID += "_STYLE"
	}
	if match1.IsPicky && match2.IsPicky {
		newID += "_PICKY"
	}
	merged := languagetool.LocalMatch{
		FromPos:          match1.FromPos,
		ToPos:            match2.ToPos,
		Message:          "Il pourrait y avoir un problème ici.",
		ShortMessage:     "Erreur potentielle",
		RuleID:           newID,
		IssueType:        match1.IssueType,
		CategoryID:       match1.CategoryID,
		CategoryName:     match1.CategoryName,
		Priority:         match1.Priority,
		Description:      match1.Description,
		OriginalErrorStr: newErr,
		SentenceText:     firstNonEmpty(match1.SentenceText, match2.SentenceText),
		FromPosSentence:  mergeSentenceFrom(match1),
		ToPosSentence:    mergeSentenceTo(match2),
		IsPicky:          match1.IsPicky,
	}
	if newRepl != "" {
		merged.Suggestions = []string{newRepl}
	}
	// Java: if ITS differ → Grammar; both Style → Style
	if match1.IssueType != match2.IssueType {
		if isStyleIssue(match1.IssueType) && isStyleIssue(match2.IssueType) {
			merged.IssueType = match1.IssueType
		} else {
			merged.IssueType = "grammar"
		}
	} else if isStyleIssue(match1.IssueType) {
		merged.IssueType = match1.IssueType
	}
	// RuleMeta fill when category left empty (ToLocalMatches soft path).
	if merged.CategoryID == "" {
		catID, catName, issue, _ := languagetool.RuleMeta(merged.RuleID)
		if issue != "" && issue != "uncategorized" {
			merged.CategoryID = catID
			if merged.CategoryName == "" {
				merged.CategoryName = catName
			}
		}
	}
	if merged.Description == "" {
		if d := languagetool.RuleDescription(merged.RuleID); d != "" && d != merged.RuleID {
			merged.Description = d
		}
	}
	return merged
}
