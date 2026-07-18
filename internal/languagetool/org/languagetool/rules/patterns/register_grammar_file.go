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
	for _, ar := range abstracts {
		if ar == nil || len(ar.PatternTokens) == 0 {
			continue
		}
		pr := NewPatternRule(ar.ID, ar.LanguageCode, ar.PatternTokens, ar.Description, ar.Message, ar.ShortMessage)
		pr.AntiPatterns = append([]*PatternRule(nil), ar.AntiPatterns...)
		pr.Filter = ar.Filter
		pr.FilterArgs = ar.FilterArgs
		// strip XML suggestion tags from message for display; keep as suggestions if present
		msg, suggs := extractSuggestions(pr.Message)
		pr.Message = msg
		id := pr.GetID()
		if id == "" {
			continue
		}
		rule := pr
		suggestions := suggs
		catID, catName := ar.CategoryID, ar.CategoryName
		desc := ar.Description
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
				if out[i].IssueType == "" && catID != "" {
					// Java category → ITS type (subset used by XML categories).
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
				// Expand \N backrefs (Java formatMatches subset) using tokens in the match span.
				if text != "" {
					from, to := out[i].FromPos, out[i].ToPos
					spanToks := matchSpanTokens(s, from, to)
					if out[i].Message != "" {
						out[i].Message = expandPatternBackrefs(out[i].Message, spanToks)
					}
					for j, sug := range out[i].Suggestions {
						out[i].Suggestions[j] = expandPatternBackrefs(sug, spanToks)
					}
					// Java RuleMatch startsWithUppercase / isAllUppercase adjustment.
					if from >= 0 && to <= len(text) && from < to && len(out[i].Suggestions) > 0 {
						matched := text[from:to]
						for j, sug := range out[i].Suggestions {
							out[i].Suggestions[j] = languagetool.SoftPreserveCase(matched, sug)
						}
					}
				}
			}
			return out
		})
		// XML default="off" / temp_off → registered but disabled (Java Rule.defaultOff).
		if ar.DefaultOff {
			lt.DisableRule(id)
			lt.MarkDefaultOff(id)
		}
		n++
	}
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

// expandPatternBackrefs replaces \1..\9 with span tokens (1-based). Unknown stays as-is.
// Subset of Java PatternRuleMatcher.formatMatches backref handling.
func expandPatternBackrefs(sug string, spanToks []string) string {
	if sug == "" || !strings.Contains(sug, `\`) {
		return sug
	}
	var b strings.Builder
	b.Grow(len(sug))
	for i := 0; i < len(sug); i++ {
		if sug[i] == '\\' && i+1 < len(sug) && sug[i+1] >= '1' && sug[i+1] <= '9' {
			n := int(sug[i+1] - '0')
			if n >= 1 && n <= len(spanToks) {
				b.WriteString(spanToks[n-1])
			} else {
				// Unknown backref: leave literal (incomplete vs Java; do not invent empty).
				b.WriteByte('\\')
				b.WriteByte(sug[i+1])
			}
			i++
			continue
		}
		b.WriteByte(sug[i])
	}
	return b.String()
}
