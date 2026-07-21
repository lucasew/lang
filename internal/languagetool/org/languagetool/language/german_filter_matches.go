package language

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// deJavaTrailingWS ports Java replaceAll("\\s+$", "") without UNICODE_CHARACTER_CLASS.
var deJavaTrailingWS = regexp.MustCompile(`[ \t\n\v\f\r]+$`)

// FilterGermanRuleMatches ports German.filterRuleMatches (AI_DE_GGEC adjacent merge / overlap skip).
//
// Incomplete vs Java (explicit, not invent):
//   - AI_DE_GGEC_MISSING_PUNCTUATION_PERIOD drop needs SentenceText (Check fills it);
//     without sentence text the match is kept (fail-closed).
func FilterGermanRuleMatches(matches []languagetool.LocalMatch) []languagetool.LocalMatch {
	if len(matches) == 0 {
		return nil
	}
	var result []languagetool.LocalMatch
	var previous *languagetool.LocalMatch

	for i := range matches {
		current := matches[i]
		cur := current // stable copy for pointer

		// Java: ignore adding punctuation at the sentence end (single-suggestion GGEC period).
		if germanDropTrailingPeriodSuggestion(cur) {
			continue
		}

		// Java filterRuleMatches gates on getRule().getId().startsWith("AI_DE_GGEC").
		// mergeMatches keeps match1's Rule object and only setSpecificRuleId("AI_DE_MERGED_MATCH"),
		// so a already-merged previous still qualifies for further chain merges. LocalMatch has a
		// single RuleID (API surface = specific id), so treat AI_DE_MERGED_MATCH as still GGEC-mergeable.
		if previous != nil &&
			isGermanGGECMergeID(previous.RuleID) &&
			isGermanGGECMergeID(cur.RuleID) {

			if previous.ToPos > cur.FromPos {
				// Skip overlapping matches
				continue
			}
			// Adjacent and same picky status (Java Tag.picky equality via LocalMatch.IsPicky)
			adjacent := previous.ToPos == cur.FromPos || previous.ToPos+1 == cur.FromPos
			samePicky := previous.IsPicky == cur.IsPicky
			if adjacent && samePicky {
				sameITS := previous.IssueType == cur.IssueType
				if sameITS {
					merged := mergeGermanGGECMatches(*previous, cur)
					previous = &merged
					continue
				}
				// Different ITS but neither is Style → merge as Grammar
				if !isStyleIssue(previous.IssueType) && !isStyleIssue(cur.IssueType) {
					merged := mergeGermanGGECMatches(*previous, cur)
					merged.IssueType = "grammar"
					previous = &merged
					continue
				}
			}
			// No merge: flush previous, advance
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

// germanDropTrailingPeriodSuggestion ports German.filterRuleMatches skip when
// AI_DE_GGEC_MISSING_PUNCTUATION_PERIOD only appends "." to the already-complete sentence.
// Java: sentence.getText().replaceAll("\\s+$", "").endsWith(suggestion without trailing ".").
func germanDropTrailingPeriodSuggestion(m languagetool.LocalMatch) bool {
	if m.RuleID != "AI_DE_GGEC_MISSING_PUNCTUATION_PERIOD" {
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
		// No sentence surface → keep (fail-closed; do not invent drop).
		return false
	}
	// Java replaceAll("\\s+$", "") — Pattern \s without UNICODE_CHARACTER_CLASS (not NBSP).
	trimmed := deJavaTrailingWS.ReplaceAllString(sent, "")
	prefix := sug[:len(sug)-1]
	return strings.HasSuffix(trimmed, prefix)
}

func isStyleIssue(it string) bool {
	return strings.EqualFold(tools.JavaStringTrim(it), "style")
}

// isGermanGGECMergeID is true for live AI_DE_GGEC* rules and for AI_DE_MERGED_MATCH
// (Java still has getRule().getId() starting with AI_DE_GGEC after merge).
func isGermanGGECMergeID(id string) bool {
	return strings.HasPrefix(id, "AI_DE_GGEC") || id == "AI_DE_MERGED_MATCH"
}

func mergeGermanGGECMatches(match1, match2 languagetool.LocalMatch) languagetool.LocalMatch {
	// Java: separator " " when positions are one apart
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
	// Java mergeMatches: originalErrorStr + separator + originalErrorStr.
	// Resolve surfaces like RuleMatch.getOriginalErrorStr / sentence substring.
	s1, s2 := match1.OriginalSurface(), match2.OriginalSurface()
	var newErr string
	if s1 != "" || s2 != "" {
		newErr = s1 + sep + s2
	}
	merged := languagetool.LocalMatch{
		FromPos:          match1.FromPos,
		ToPos:            match2.ToPos,
		Message:          "Hier scheint es einen Fehler zu geben.",
		ShortMessage:     "Potenzieller Fehler",
		RuleID:           "AI_DE_MERGED_MATCH",
		IssueType:        match1.IssueType,
		CategoryID:       match1.CategoryID,
		CategoryName:     match1.CategoryName,
		Priority:         match1.Priority,
		Description:      match1.Description,
		OriginalErrorStr: newErr,
		SentenceText:     firstNonEmpty(match1.SentenceText, match2.SentenceText),
		// Sentence-relative span of the merged range (for further surface filters).
		FromPosSentence: mergeSentenceFrom(match1),
		ToPosSentence:   mergeSentenceTo(match2),
		// Same picky status required to merge; keep that flag on the result.
		IsPicky: match1.IsPicky,
	}
	if newRepl != "" {
		merged.Suggestions = []string{newRepl}
	}
	// Java mergeMatches: if ITS differ → Grammar; both Style → Style
	if match1.IssueType != match2.IssueType {
		if isStyleIssue(match1.IssueType) && isStyleIssue(match2.IssueType) {
			merged.IssueType = match1.IssueType
		} else {
			merged.IssueType = "grammar"
		}
	}
	// RuleMeta fill when category left empty (same path as ToLocalMatches;
	// Java keeps match1 rule category — fallback only when LocalMatch metadata empty).
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

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// mergeSentenceFrom / mergeSentenceTo prefer FromPosSentence/ToPosSentence when set.
func mergeSentenceFrom(m languagetool.LocalMatch) int {
	if m.FromPosSentence > -1 && m.ToPosSentence > m.FromPosSentence {
		return m.FromPosSentence
	}
	return m.FromPos
}

func mergeSentenceTo(m languagetool.LocalMatch) int {
	if m.FromPosSentence > -1 && m.ToPosSentence > m.FromPosSentence {
		return m.ToPosSentence
	}
	return m.ToPos
}
