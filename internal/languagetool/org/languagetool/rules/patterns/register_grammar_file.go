package patterns

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
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
	n := 0
	// Track default-off categories once (Java Category.isDefaultOff on each rule's category).
	for _, ar := range abstracts {
		if ar != nil && ar.CategoryDefaultOff && ar.CategoryID != "" {
			lt.MarkCategoryDefaultOff(ar.CategoryID)
		}
	}
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
		pr.InterpretPreDisambig = ar.InterpretPreDisambig
		pr.ToneTags = append([]languagetool.ToneTag(nil), ar.ToneTags...)
		pr.GoalSpecific = ar.GoalSpecific
		pr.DefaultOff = ar.DefaultOff
		pr.DefaultTempOff = ar.DefaultTempOff
		pr.SubID = ar.SubID
		pr.SourceFile = ar.SourceFile
		pr.IssueType = ar.IssueType
		pr.URL = ar.URL
		pr.Priority = ar.Priority
		pr.Premium = ar.Premium
		pr.MinPrevMatches = ar.MinPrevMatches
		pr.DistanceTokens = ar.DistanceTokens
		if len(ar.Tags) > 0 {
			pr.SetTags(ar.Tags)
		}
		// strip XML suggestion tags from message for display; templates expanded in matcher.
		msg, suggs := extractSuggestions(pr.Message)
		pr.Message = msg
		pr.SuggestionTemplates = append([]string(nil), suggs...)
		id := pr.GetID()
		if id == "" {
			continue
		}
		rule := pr
		catID, catName := ar.CategoryID, ar.CategoryName
		issueType := ar.IssueType
		ruleURL := ar.URL
		rulePrio := ar.Priority
		desc := ar.Description
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
			for i := range out {
				if out[i].Description == "" {
					out[i].Description = desc
				}
				if out[i].CategoryID == "" {
					out[i].CategoryID = catID
				}
				if out[i].CategoryName == "" {
					out[i].CategoryName = catName
				}
				if out[i].IssueType == "" {
					// Java: rule/group/category type; then soft id-based fallback.
					if issueType != "" {
						out[i].IssueType = issueType
					} else if catID != "" {
						switch strings.ToUpper(catID) {
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
				if out[i].URL == "" && ruleURL != "" {
					out[i].URL = ruleURL
				}
				// Java Rule.getPriority before Language.getRulePriority overlay.
				if out[i].Priority == 0 && rulePrio != 0 {
					out[i].Priority = rulePrio
				}
				// Case adjustment when matcher left suggestions (formatMatches already ran).
				if text != "" {
					from, to := out[i].FromPos, out[i].ToPos
					if from >= 0 && to <= len(text) && from < to && len(out[i].Suggestions) > 0 {
						matched := text[from:to]
						for j, sug := range out[i].Suggestions {
							out[i].Suggestions[j] = languagetool.PreserveCase(matched, sug)
						}
					}
				}
			}
			return out
		})
		// XML default="off" / temp_off → registered but disabled (Java Rule.defaultOff).
		// temp_off also tracked for enableTempOff / JSON rule.tempOff (Java isDefaultTempOff).
		// Category default="off" does NOT set rule.defaultOff (Java Category.isDefaultOff only).
		if ar.DefaultTempOff {
			lt.MarkDefaultTempOff(id)
		} else if ar.DefaultOff {
			lt.MarkDefaultOff(id)
		}
		n++
	}
	// Java activateDefaultPatternRules: after loading pattern rules, apply
	// getDefaultEnabledRulesForVariant / getDefaultDisabledRulesForVariant
	// (setDefaultOn / setDefaultOff on matching rule IDs).
	lt.ApplyVariantDefaultRules()
	return n, nil
}

// extractSuggestions pulls <suggestion>…</suggestion> from rule messages (Java markup).
// Does not invent suggestions from quoted prose.
func extractSuggestions(msg string) (clean string, suggs []string) {
	clean = msg
	for {
		start := strings.Index(clean, "<suggestion>")
		if start < 0 {
			break
		}
		end := strings.Index(clean[start:], "</suggestion>")
		if end < 0 {
			break
		}
		end += start
		inner := clean[start+len("<suggestion>") : end]
		suggs = append(suggs, inner)
		clean = clean[:start] + inner + clean[end+len("</suggestion>"):]
	}
	return strings.TrimSpace(clean), suggs
}

// matchSpanTokens returns non-whitespace token surfaces whose span overlaps [from,to).
// Incomplete vs Java formatMatches (pattern-element indices / optional tokens / SENT_START):
// only real token surfaces in the match range — no invent of empty SENT_START slots.
func matchSpanTokens(s *languagetool.AnalyzedSentence, from, to int) []string {
	if s == nil || from < 0 || to <= from {
		return nil
	}
	var out []string
	for _, tok := range s.GetTokensWithoutWhitespace() {
		if tok == nil || tok.IsSentenceEnd() || tok.IsSentenceStart() {
			continue
		}
		ts, te := tok.GetStartPos(), tok.GetEndPos()
		if te <= from || ts >= to {
			continue
		}
		if t := tok.GetToken(); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// expandPatternBackrefs replaces \1, \2, … (multi-digit) with span tokens (1-based).
// Ports the digit-run scan in Java PatternRuleMatcher.formatMatches.
// Unknown backrefs stay literal (do not invent empty replacements).
func expandPatternBackrefs(sug string, spanToks []string) string {
	if sug == "" || !strings.Contains(sug, `\`) {
		return sug
	}
	var b strings.Builder
	b.Grow(len(sug))
	for i := 0; i < len(sug); i++ {
		if sug[i] != '\\' || i+1 >= len(sug) || sug[i+1] < '1' || sug[i+1] > '9' {
			b.WriteByte(sug[i])
			continue
		}
		// Java: while Character.isDigit — multi-digit backrefs
		j := i + 1
		for j < len(sug) && sug[j] >= '0' && sug[j] <= '9' {
			j++
		}
		n := 0
		for k := i + 1; k < j; k++ {
			n = n*10 + int(sug[k]-'0')
		}
		if n >= 1 && n <= len(spanToks) {
			b.WriteString(spanToks[n-1])
		} else {
			// Unknown backref: leave literal (do not invent empty).
			b.WriteString(sug[i:j])
		}
		i = j - 1
	}
	return b.String()
}
