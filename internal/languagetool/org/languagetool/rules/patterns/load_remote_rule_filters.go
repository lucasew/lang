package patterns

import (
	"os"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// LoadRemoteRuleFiltersFile ports RemoteRuleFilters.load: parse remote-rule-filters.xml
// pattern rules and register them on GlobalRemoteRuleFilters for langCode.
// Rule ids are treated as regexes over remote match rule ids (Java compilePatterns).
// Missing/unreadable path returns (0, nil) — fail-closed, no invent.
// Returns the number of filter rules registered.
func LoadRemoteRuleFiltersFile(path, langCode string) (int, error) {
	if path == "" || langCode == "" {
		return 0, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	return LoadRemoteRuleFiltersXML(string(data), path, langCode)
}

// LoadRemoteRuleFiltersXML registers filters from a remote-rule-filters XML string.
func LoadRemoteRuleFiltersXML(xmlStr, filename, langCode string) (int, error) {
	if strings.TrimSpace(xmlStr) == "" || langCode == "" {
		return 0, nil
	}
	if filename == "" {
		filename = "remote-rule-filters.xml"
	}
	// Java uses lang.getShortCode() for the registry key path; register under short code.
	short := langCode
	if i := strings.IndexByte(short, '-'); i > 0 {
		// keep de-DE-x-simple-language → de (Java special-case uses de-DE → short de)
		if short == "de-DE-x-simple-language" {
			short = "de"
		} else {
			short = short[:i]
		}
	}

	loader := NewPatternRuleLoader()
	loader.SetRelaxedMode(true)
	abstracts, err := loader.GetRulesFromString(xmlStr, filename, short)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, ar := range abstracts {
		if ar == nil || ar.ID == "" || len(ar.PatternTokens) == 0 {
			continue
		}
		idRe, err := rules.CompileRemoteRuleIDPattern(ar.ID)
		if err != nil {
			// Invalid id-as-regex: skip (do not invent a literal-only filter).
			continue
		}
		pr := NewPatternRule(ar.ID, ar.LanguageCode, ar.PatternTokens, ar.Description, ar.Message, ar.ShortMessage)
		pr.AntiPatterns = append([]*PatternRule(nil), ar.AntiPatterns...)
		pr.Filter = ar.Filter
		pr.FilterArgs = ar.FilterArgs
		pr.UnifierConfig = ar.UnifierConfig
		pr.InterpretPreDisambig = ar.InterpretPreDisambig
		rule := pr
		rules.GlobalRemoteRuleFilters.Register(short, &rules.FilterRule{
			IDPattern: idRe,
			MatchPositions: func(sentence *languagetool.AnalyzedSentence) []rules.MatchPosition {
				if sentence == nil || rule == nil {
					return nil
				}
				ms, err := rule.Match(sentence)
				if err != nil || len(ms) == 0 {
					return nil
				}
				out := make([]rules.MatchPosition, 0, len(ms))
				for _, m := range ms {
					if m == nil {
						continue
					}
					out = append(out, rules.MatchPosition{Start: m.FromPos, End: m.ToPos})
				}
				return out
			},
		})
		n++
	}
	return n, nil
}
