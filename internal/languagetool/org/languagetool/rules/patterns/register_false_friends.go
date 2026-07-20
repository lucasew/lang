package patterns

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterFalseFriendsFile loads official false-friend pattern rules for textLang/motherLang
// (Java FalseFriendRuleLoader path — not soft invent packs).
// Returns the number of rules registered.
func RegisterFalseFriendsFile(lt *languagetool.JLanguageTool, path, textLang, motherLang string) (int, error) {
	if lt == nil || path == "" || motherLang == "" {
		return 0, nil
	}
	data, err := ReadExpandedGrammarFile(path)
	if err != nil {
		return 0, err
	}
	return RegisterFalseFriendsXML(lt, string(data), textLang, motherLang)
}

// RegisterFalseFriendsXML registers false-friend rules from an XML string.
func RegisterFalseFriendsXML(lt *languagetool.JLanguageTool, xmlStr, textLang, motherLang string) (int, error) {
	if lt == nil || strings.TrimSpace(xmlStr) == "" || motherLang == "" {
		return 0, nil
	}
	if textLang == "" {
		textLang = "en"
	}
	loader := NewFalseFriendRuleLoader("", "")
	ffRules, err := loader.GetRulesFromString(xmlStr, textLang, motherLang)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, fr := range ffRules {
		if fr == nil || fr.PatternRule == nil {
			continue
		}
		pr := fr.PatternRule
		id := pr.GetID()
		if id == "" {
			id = "FALSE_FRIEND"
		}
		suggs := append([]string(nil), loader.SuggestionMap[id]...)
		rule := pr
		suggestions := suggs
		lt.AddRuleChecker(id, func(s *languagetool.AnalyzedSentence) []languagetool.LocalMatch {
			ms, err := rule.Match(s)
			if err != nil || len(ms) == 0 {
				return nil
			}
			if len(suggestions) > 0 {
				for _, m := range ms {
					if m != nil && len(m.GetSuggestedReplacements()) == 0 {
						m.SetSuggestedReplacements(append([]string(nil), suggestions...))
					}
				}
			}
			out := rules.ToLocalMatches(ms)
			for i := range out {
				if out[i].CategoryID == "" {
					out[i].CategoryID = "FALSEFRIENDS"
					out[i].CategoryName = "False Friends"
					out[i].IssueType = "misspelling"
				}
			}
			return out
		})
		n++
	}
	return n, nil
}
