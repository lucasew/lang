package language

import (
	"regexp"
	"sort"
	"strings"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

func init() {
	languagetool.FilterCatalanRuleMatchesHook = FilterCatalanRuleMatches
	languagetool.FilterCatalanRuleMatchesAfterOverlappingHook = FilterCatalanRuleMatchesAfterOverlapping
}

// POSSESSIUS_v / POSSESSIUS_V ports Catalan.adjustCatalanMatch patterns.
var (
	caPossessiusV = regexp.MustCompile(`(?i)\b([mtsMTS]e)v(a|es)\b`)
	caPossessiusU = regexp.MustCompile(`\b([MTS]E)V(A|ES)\b`)
)

func enabledHas(enabled map[string]struct{}, id string) bool {
	if enabled == nil {
		return false
	}
	_, ok := enabled[id]
	return ok
}

// FilterCatalanRuleMatchesAfterOverlapping ports Catalan.filterRuleMatchesAfterOverlapping:
// map trimMatchEnds then sort by FromPos.
func FilterCatalanRuleMatchesAfterOverlapping(matches []languagetool.LocalMatch) []languagetool.LocalMatch {
	if len(matches) == 0 {
		return nil
	}
	out := make([]languagetool.LocalMatch, len(matches))
	for i := range matches {
		out[i] = matches[i].TrimMatchEnds()
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].FromPos != out[j].FromPos {
			return out[i].FromPos < out[j].FromPos
		}
		return out[i].ToPos < out[j].ToPos
	})
	return out
}

// FilterCatalanRuleMatches ports Catalan.filterRuleMatches.
//
// Incomplete vs Java (explicit, not invent):
//   - FALTA_ELEMENT_ENTRE_VERBS[n] FullId gates need RuleID to carry [n] suffix when known.
// hasTypographicApostrophe: LocalMatch.HasTypographicApostropheInSentence (Check stamps
// from analyzed tokens / tagger SetTypographicApostrophe).
func FilterCatalanRuleMatches(matches []languagetool.LocalMatch) []languagetool.LocalMatch {
	if len(matches) == 0 {
		return nil
	}
	var results []languagetool.LocalMatch
	ignoreMatchInPos := -1
	var previous *languagetool.LocalMatch
	for i := range matches {
		ruleMatch := matches[i]
		// remove IGNORE_PROPER_NOUNS and MORFOLOGIK_RULE_CA_ES if same position
		if ruleMatch.RuleID == "IGNORE_PROPER_NOUNS" {
			if previous != nil && previous.RuleID == "MORFOLOGIK_RULE_CA_ES" &&
				ruleMatch.FromPos == previous.FromPos {
				if len(results) > 0 {
					results = results[:len(results)-1]
				}
				ignoreMatchInPos = -1
				continue
			}
			ignoreMatchInPos = ruleMatch.FromPos
			continue
		}
		if ruleMatch.RuleID == "MORFOLOGIK_RULE_CA_ES" && ruleMatch.FromPos == ignoreMatchInPos {
			ignoreMatchInPos = -1
			continue
		}
		// FALTA_ELEMENT_ENTRE_VERBS[3]/[4] skip when next match close and not [5]
		fullID := ruleMatch.RuleID
		if fullID == "FALTA_ELEMENT_ENTRE_VERBS[3]" || fullID == "FALTA_ELEMENT_ENTRE_VERBS[4]" {
			if i+1 < len(matches) {
				next := matches[i+1]
				if next.FromPosSentence > -1 &&
					next.RuleID != "FALTA_ELEMENT_ENTRE_VERBS[5]" &&
					next.FromPosSentence-ruleMatch.ToPosSentence < 20 {
					continue
				}
			}
		}
		if i > 0 && fullID == "FALTA_ELEMENT_ENTRE_VERBS[5]" &&
			matches[i-1].RuleID == "FALTA_ELEMENT_ENTRE_VERBS" {
			continue
		}
		adj := adjustCatalanMatchLocal(ruleMatch, ruleMatch.EnabledRules)
		results = append(results, adj)
		// Java: previousRuleMatch = ruleMatch (pre-adjust id/pos).
		prevCopy := ruleMatch
		previous = &prevCopy
	}
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].FromPos != results[j].FromPos {
			return results[i].FromPos < results[j].FromPos
		}
		return results[i].ToPos < results[j].ToPos
	})
	return results
}

// adjustCatalanMatchLocal ports Catalan.adjustCatalanMatch for LocalMatch.
// Empty suggestion + spaces around span → extend ToPos by 1 (avoid double spaces).
// enabledRules: LocalMatch.EnabledRules (Java Set); nil/empty → DIACRITICS strip on,
// EXIGEIX_*/APOSTROF_* branches off (Java empty set).
func adjustCatalanMatchLocal(m languagetool.LocalMatch, enabledRules map[string]struct{}) languagetool.LocalMatch {
	if len(m.Suggestions) == 1 && m.Suggestions[0] == "" {
		sent := m.SentenceText
		fromSent, toSent := m.FromPosSentence, m.ToPosSentence
		// Java: sentenceText.length()/substring are UTF-16 code units.
		// right char exists iff substring(toSent, toSent+1) is non-empty (and space).
		if sent != "" && fromSent >= 0 {
			// Java: substring(fromSent-1, fromSent).equals(" ")
			//       && substring(toSent, toSent+1).equals(" ")
			leftOK := fromSent == 0 || caUTF16Substring(sent, fromSent-1, fromSent) == " "
			rightOK := caUTF16Substring(sent, toSent, toSent+1) == " "
			if leftOK && rightOK {
				m.ToPos = m.ToPos + 1
				m.ToPosSentence = toSent + 1
				m.Suggestions = []string{""}
				return m
			}
		}
	}
	errorStr := m.OriginalSurface()
	// Snapshot original suggestions for EXIGEIX accent skip (Java contains checks).
	origSugs := append([]string(nil), m.Suggestions...)
	if len(m.Suggestions) > 0 {
		newSugs := make([]string, 0, len(m.Suggestions))
		seen := make(map[string]struct{}, len(m.Suggestions))
		for _, sug := range m.Suggestions {
			newRepl := sug
			if errorStr != "" && len(errorStr) > 2 && strings.HasSuffix(errorStr, "'") &&
				!strings.HasSuffix(newRepl, "'") && !strings.HasSuffix(newRepl, "’") {
				newRepl = newRepl + " "
			}
			// EXIGEIX_ACCENTUACIO_GENERAL: skip é form if è alternate exists (Java continue).
			if !strings.EqualFold(newRepl, "després") && enabledHas(enabledRules, "EXIGEIX_ACCENTUACIO_GENERAL") {
				if strings.Contains(newRepl, "é") && containsString(origSugs, strings.ReplaceAll(newRepl, "é", "è")) {
					continue
				}
				if strings.Contains(newRepl, "É") && containsString(origSugs, strings.ReplaceAll(newRepl, "É", "È")) {
					continue
				}
			} else if enabledHas(enabledRules, "EXIGEIX_ACCENTUACIO_VALENCIANA") {
				if strings.Contains(newRepl, "è") && containsString(origSugs, strings.ReplaceAll(newRepl, "è", "é")) {
					continue
				}
				if strings.Contains(newRepl, "È") && containsString(origSugs, strings.ReplaceAll(newRepl, "È", "É")) {
					continue
				}
			}
			// Java: (APOSTROF_TIPOGRAFIC || hasTypographicApostrophe) && !APOSTROF_RECTE
			if (enabledHas(enabledRules, "APOSTROF_TIPOGRAFIC") || m.HasTypographicApostropheInSentence) &&
				len(newRepl) > 1 && !enabledHas(enabledRules, "APOSTROF_RECTE") {
				newRepl = strings.ReplaceAll(newRepl, "'", "’")
			}
			// EXIGEIX_POSSESSIUS_U
			if enabledHas(enabledRules, "EXIGEIX_POSSESSIUS_U") && len(newRepl) > 3 {
				newRepl = caPossessiusV.ReplaceAllString(newRepl, "${1}u${2}")
				newRepl = caPossessiusU.ReplaceAllString(newRepl, "${1}U${2}")
				newRepl = strings.ReplaceAll(newRepl, "feina", "faena")
				newRepl = strings.ReplaceAll(newRepl, "feiner", "faener")
				newRepl = strings.ReplaceAll(newRepl, "feinera", "faenera")
			}
			// IEC orthography: strip traditional diacritics unless traditional rules enabled.
			if !enabledHas(enabledRules, "DIACRITICS_TRADITIONAL_RULES") &&
				CatalanSuggestionNeedsOldDiacriticStrip(newRepl) {
				newRepl = CatalanRemoveOldDiacritics(newRepl)
			}
			if _, ok := seen[newRepl]; !ok {
				seen[newRepl] = struct{}{}
				newSugs = append(newSugs, newRepl)
			}
		}
		m.Suggestions = newSugs
	}
	return m
}

func containsString(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}

// caUTF16Substring ports Java String.substring(from, to) with UTF-16 indices
// (Catalan.adjustCatalanMatch space checks). Local helper avoids rules import cycle.
func caUTF16Substring(s string, from, to int) string {
	u := utf16.Encode([]rune(s))
	if from < 0 {
		from = 0
	}
	if to > len(u) {
		to = len(u)
	}
	if from >= to {
		return ""
	}
	return string(utf16.Decode(u[from:to]))
}
