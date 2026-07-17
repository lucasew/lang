package patterns

import (
	"os"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterGrammarFile loads a simplified grammar/rules XML file onto lt.
// Complex constructs (unify, phrases, exceptions) are skipped by the soft loader.
// Returns the number of pattern rules registered.
func RegisterGrammarFile(lt *languagetool.JLanguageTool, path, languageCode string) (int, error) {
	if lt == nil || path == "" {
		return 0, nil
	}
	data, err := os.ReadFile(path)
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
					// soft default: grammar categories map to grammar ITS type
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
				// Soft pattern rules beat map/CFSA2 speller on the same span for --apply.
				if out[i].Priority == 0 {
					out[i].Priority = 3
				}
				// Preserve sentence-case / ALL-CAPS from the matched surface on suggestions.
				if text != "" && len(out[i].Suggestions) > 0 {
					from, to := out[i].FromPos, out[i].ToPos
					if from >= 0 && to <= len(text) && from < to {
						matched := text[from:to]
						for j, sug := range out[i].Suggestions {
							out[i].Suggestions[j] = languagetool.SoftPreserveCase(matched, sug)
						}
					}
				}
			}
			return out
		})
		n++
	}
	return n, nil
}

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
	// soft: also pull "quoted" segments from Did you mean "..."?
	if len(suggs) == 0 {
		for _, q := range []string{`"`, `'`} {
			i := strings.Index(clean, q)
			for i >= 0 {
				j := strings.Index(clean[i+1:], q)
				if j < 0 {
					break
				}
				j += i + 1
				inner := clean[i+1 : j]
				if len(inner) > 0 && len(inner) < 80 {
					suggs = append(suggs, inner)
				}
				i = strings.Index(clean[j+1:], q)
				if i >= 0 {
					i += j + 1
				}
			}
			if len(suggs) > 0 {
				break
			}
		}
	}
	return strings.TrimSpace(clean), suggs
}

// RegisterSoftGrammarDir loads {dir}/{lang}-soft.xml or {dir}/{lang}/grammar-soft.xml if present.
// Paths are de-duplicated so en and en-US do not register the same en-soft.xml twice.
func RegisterSoftGrammarDir(lt *languagetool.JLanguageTool, dir, languageCode string) (int, error) {
	if lt == nil || dir == "" {
		return 0, nil
	}
	base := languageCode
	if i := strings.IndexByte(languageCode, '-'); i > 0 {
		base = languageCode[:i]
	}
	raw := []string{
		dir + "/" + base + "-soft.xml",
		dir + "/" + languageCode + "-soft.xml",
		dir + "/" + base + "/grammar-soft.xml",
	}
	seen := map[string]struct{}{}
	var candidates []string
	for _, c := range raw {
		if _, ok := seen[c]; ok {
			continue
		}
		seen[c] = struct{}{}
		candidates = append(candidates, c)
	}
	total := 0
	for _, c := range candidates {
		n, err := RegisterGrammarFile(lt, c, languageCode)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return total, err
		}
		total += n
	}
	return total, nil
}
