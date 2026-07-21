package patterns

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// RegisterGrammarFile loads a grammar/rules XML file onto lt.
// Returns the number of pattern rules registered.
func RegisterGrammarFile(lt *languagetool.JLanguageTool, path, languageCode string) (int, error) {
	if lt == nil || path == "" {
		return 0, nil
	}
	data, err := ReadExpandedGrammarFile(path)
	if err != nil {
		return 0, err
	}
	return RegisterGrammarXML(lt, string(data), path, languageCode)
}

// RegisterGrammarXML registers pattern rules from a simplified rules XML string.
func RegisterGrammarXML(lt *languagetool.JLanguageTool, xmlStr, filename, languageCode string) (int, error) {
	if lt == nil || strings.TrimSpace(xmlStr) == "" {
		return 0, nil
	}
	if languageCode == "" {
		languageCode = "en"
	}
	loader := NewPatternRuleLoader()
	loader.SetRelaxedMode(true)
	abstracts, err := loader.GetRulesFromString(xmlStr, filename, languageCode)
	if err != nil {
		return 0, err
	}
	// Track default-off categories once (Java Category.isDefaultOff on each rule's category).
	for _, ar := range abstracts {
		if ar != nil && ar.CategoryDefaultOff && ar.CategoryID != "" {
			lt.MarkCategoryDefaultOff(ar.CategoryID)
		}
	}

	type builtRule struct {
		pr   *PatternRule
		ar   *AbstractPatternRule
		meta grammarRuleMeta
	}
	var built []builtRule
	for _, ar := range abstracts {
		if ar == nil || len(ar.PatternTokens) == 0 {
			continue
		}
		pr := NewPatternRule(ar.ID, ar.LanguageCode, ar.PatternTokens, ar.Description, ar.Message, ar.ShortMessage)
		pr.AntiPatterns = append([]*PatternRule(nil), ar.AntiPatterns...)
		pr.Filter = ar.Filter
		pr.FilterArgs = ar.FilterArgs
		pr.UnifierConfig = ar.UnifierConfig
		pr.SuggestionMatches = append([]*Match(nil), ar.SuggestionMatches...)
		pr.SuggestionMatchesOutMsg = append([]*Match(nil), ar.SuggestionMatchesOutMsg...)
		pr.SuggestionsOutMsg = ar.SuggestionsOutMsg
		pr.StartPositionCorrection = ar.StartPositionCorrection
		pr.EndPositionCorrection = ar.EndPositionCorrection
		pr.InterpretPreDisambig = ar.InterpretPreDisambig
		pr.ToneTags = append([]languagetool.ToneTag(nil), ar.ToneTags...)
		pr.GoalSpecific = ar.GoalSpecific
		pr.DefaultOff = ar.DefaultOff
		pr.DefaultTempOff = ar.DefaultTempOff
		pr.SubID = ar.SubID
		pr.SourceFile = ar.SourceFile
		pr.LineNumber = ar.LineNumber
		pr.IssueType = ar.IssueType
		pr.URL = ar.URL
		pr.Priority = ar.Priority
		pr.Premium = ar.Premium
		pr.MinPrevMatches = ar.MinPrevMatches
		pr.DistanceTokens = ar.DistanceTokens
		if len(ar.Tags) > 0 {
			pr.SetTags(ar.Tags)
		}
		// Java PatternRule keeps <suggestion>…</suggestion> in message/suggestionsOutMsg.
		// FormatMatches + removeSuppressMisspelled + RuleMatch ctor extract depend on tags
		// remaining (do not strip into SuggestionTemplates — that soft-invent path broke
		// suppress_misspelled and multi-synthesis).
		if pr.GetID() == "" {
			continue
		}
		built = append(built, builtRule{
			pr: pr,
			ar: ar,
			meta: grammarRuleMeta{
				CatID:     ar.CategoryID,
				CatName:   ar.CategoryName,
				IssueType: ar.IssueType,
				URL:       ar.URL,
				Priority:  ar.Priority,
				Desc:      ar.Description,
			},
		})
	}

	// Java transformPatternRules: RepeatedPatternRuleTransformer then
	// ConsistencyPatternRuleTransformer (remaining stay sentence-level).
	repeatedByID := map[string][]builtRule{}
	var repeatedOrder []string
	var afterRepeated []builtRule
	for _, b := range built {
		if b.pr.MinPrevMatches > 0 {
			id := b.pr.GetID()
			if _, ok := repeatedByID[id]; !ok {
				repeatedOrder = append(repeatedOrder, id)
			}
			repeatedByID[id] = append(repeatedByID[id], b)
			continue
		}
		afterRepeated = append(afterRepeated, b)
	}

	consistPrefix := tools.ConsistencyRulePrefix
	consistByMain := map[string][]builtRule{}
	var consistOrder []string
	var sentenceLevel []builtRule
	for _, b := range afterRepeated {
		id := b.pr.GetID()
		if strings.HasPrefix(id, consistPrefix) {
			main := GetMainRuleId(id)
			if _, ok := consistByMain[main]; !ok {
				consistOrder = append(consistOrder, main)
			}
			consistByMain[main] = append(consistByMain[main], b)
			continue
		}
		sentenceLevel = append(sentenceLevel, b)
	}

	n := 0
	for _, id := range repeatedOrder {
		group := repeatedByID[id]
		prs := make([]*PatternRule, 0, len(group))
		for _, b := range group {
			prs = append(prs, b.pr)
		}
		rep := &RepeatedPatternRule{
			LanguageCode:             languageCode,
			PatternRules:             prs,
			DefaultMaxDistanceTokens: 60,
		}
		meta := group[0].meta
		// Use first rule for default-off / temp-off tracking (shared id).
		ar0 := group[0].ar
		lt.AddTextLevelRuleChecker(id, func(sents []*languagetool.AnalyzedSentence) []languagetool.LocalMatch {
			out := rep.MatchSentences(sents)
			return enrichLocalMatches(out, "", meta)
		})
		if ar0 != nil {
			if ar0.DefaultTempOff {
				lt.MarkDefaultTempOff(id)
			} else if ar0.DefaultOff {
				lt.MarkDefaultOff(id)
			}
		}
		n++
	}

	for _, main := range consistOrder {
		group := consistByMain[main]
		prs := make([]*PatternRule, 0, len(group))
		ars := make([]*AbstractPatternRule, 0, len(group))
		for _, b := range group {
			prs = append(prs, b.pr)
			ars = append(ars, b.ar)
		}
		consist := &ConsistencyPatternRule{
			MainID:        main,
			LanguageCode:  languageCode,
			PatternRules:  prs,
			AbstractRules: ars,
		}
		meta := group[0].meta
		ar0 := group[0].ar
		lt.AddTextLevelRuleChecker(main, func(sents []*languagetool.AnalyzedSentence) []languagetool.LocalMatch {
			out := consist.MatchSentences(sents)
			return enrichLocalMatches(out, "", meta)
		})
		if ar0 != nil {
			if ar0.DefaultTempOff {
				lt.MarkDefaultTempOff(main)
			} else if ar0.DefaultOff {
				lt.MarkDefaultOff(main)
			}
		}
		n++
	}

	for _, b := range sentenceLevel {
		pr := b.pr
		id := pr.GetID()
		rule := pr
		meta := b.meta
		lt.AddRuleChecker(id, func(s *languagetool.AnalyzedSentence) []languagetool.LocalMatch {
			ms, err := rule.Match(s)
			if err != nil || len(ms) == 0 {
				return nil
			}
			out := rules.ToLocalMatches(ms)
			text := ""
			if s != nil {
				text = s.GetText()
			}
			return enrichLocalMatches(out, text, meta)
		})
		if b.ar != nil {
			if b.ar.DefaultTempOff {
				lt.MarkDefaultTempOff(id)
			} else if b.ar.DefaultOff {
				lt.MarkDefaultOff(id)
			}
		}
		n++
	}
	// Java activateDefaultPatternRules: after loading pattern rules, apply
	// getDefaultEnabledRulesForVariant / getDefaultDisabledRulesForVariant
	// (setDefaultOn / setDefaultOff on matching rule IDs).
	lt.ApplyVariantDefaultRules()
	return n, nil
}

// grammarRuleMeta is match enrichment shared by sentence- and text-level pattern rules.
type grammarRuleMeta struct {
	CatID, CatName, IssueType, URL, Desc string
	Priority                             int
}

func enrichLocalMatches(out []languagetool.LocalMatch, text string, meta grammarRuleMeta) []languagetool.LocalMatch {
	for i := range out {
		if out[i].Description == "" {
			out[i].Description = meta.Desc
		}
		if out[i].CategoryID == "" {
			out[i].CategoryID = meta.CatID
		}
		if out[i].CategoryName == "" {
			out[i].CategoryName = meta.CatName
		}
		if out[i].IssueType == "" {
			if meta.IssueType != "" {
				out[i].IssueType = meta.IssueType
			} else if meta.CatID != "" {
				switch strings.ToUpper(meta.CatID) {
				case "TYPOS":
					out[i].IssueType = "misspelling"
				case "STYLE":
					out[i].IssueType = "style"
				case "TYPOGRAPHY":
					out[i].IssueType = "typographical"
				case "CASING":
					out[i].IssueType = "typographical"
				default:
					out[i].IssueType = "grammar"
				}
			}
		}
		if out[i].URL == "" && meta.URL != "" {
			out[i].URL = meta.URL
		}
		if out[i].Priority == 0 && meta.Priority != 0 {
			out[i].Priority = meta.Priority
		}
		if text != "" {
			from, to := out[i].FromPos, out[i].ToPos
			// LocalMatch positions are UTF-16 units (Java RuleMatch); not byte indices.
			if from >= 0 && from < to && len(out[i].Suggestions) > 0 {
				matched := rules.UTF16Substring(text, from, to)
				if matched != "" {
					for j, sug := range out[i].Suggestions {
						out[i].Suggestions[j] = languagetool.PreserveCase(matched, sug)
					}
				}
			}
		}
	}
	return out
}


