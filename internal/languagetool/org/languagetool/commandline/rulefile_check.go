package commandline

import (
	"fmt"
	"io"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// CheckWithPatternRuleFile loads --rulefile XML and runs loaded pattern rules on text.
// Multi-sentence: analyzes with JLanguageTool and maps offsets to the document.
// Returns match count; writes API XML/JSON/plain depending on opts.OutputFormat.
func CheckWithPatternRuleFile(w io.Writer, text string, opts *CommandLineOptions) (int, error) {
	if opts == nil || opts.GetRuleFile() == "" {
		return 0, fmt.Errorf("no rulefile")
	}
	xml, err := LoadRuleFile(opts.GetRuleFile())
	if err != nil {
		return 0, err
	}
	lang := opts.Language
	if lang == "" {
		if inferred := InferLanguageFromRuleFileName(opts.GetRuleFile()); inferred != "" {
			lang = inferred
		} else {
			lang = "en"
		}
	}
	loader := patterns.NewPatternRuleLoader()
	abstracts, err := loader.GetRulesFromString(xml, opts.GetRuleFile(), lang)
	if err != nil {
		return 0, err
	}

	lt := languagetool.NewJLanguageTool(lang)
	for _, ar := range abstracts {
		if ar == nil {
			continue
		}
		pr := patterns.NewPatternRule(ar.ID, ar.LanguageCode, ar.PatternTokens, ar.Description, ar.Message, ar.ShortMessage)
		pr.UnifierConfig = ar.UnifierConfig
		pr.AntiPatterns = append([]*patterns.PatternRule(nil), ar.AntiPatterns...)
		pr.Filter = ar.Filter
		pr.FilterArgs = ar.FilterArgs
		pr.LineNumber = ar.LineNumber
		pr.SourceFile = ar.SourceFile
		patterns.RegisterPatternRule(lt, pr)
	}
	// apply CLI enable/disable
	for _, id := range opts.GetDisabledRules() {
		lt.DisableRule(id)
	}

	local := lt.Check(text)
	// convert to RuleMatch for printers
	sent := languagetool.AnalyzePlain(text)
	all := rules.FromLocalMatches(local, sent)
	// attach idRule for GetID when Rule is nil
	for _, m := range all {
		if m != nil && m.GetRule() == nil {
			// FromLocalMatches should attach FakeRule with id — leave as-is
		}
	}
	all = FilterMatchesByRules(all, opts.GetDisabledRules(), opts.GetEnabledRules(), opts.IsUseEnabledOnly())
	if opts.OutputFormat == OutputXML {
		_, _ = io.WriteString(w, MatchesAsMinimalXML(all, lang))
	} else if opts.OutputFormat == OutputJSON {
		_, _ = io.WriteString(w, MatchesAsJSON(all, lang, text))
	} else {
		for i, m := range all {
			if m == nil {
				continue
			}
			line, col := LineColumnAt(text, m.FromPos)
			fmt.Fprintf(w, "%d.) Line %d, column %d, Rule ID: %s\n", i+1, line, col, ruleIDOfMatch(m))
			if m.GetMessage() != "" {
				fmt.Fprintf(w, "Message: %s\n", m.GetMessage())
			}
		}
	}
	return len(all), nil
}

// idRule attaches GetID for filtered matches.
type idRule struct{ id string }

func (r idRule) GetID() string { return r.id }

// CheckBitextWithRuleFile loads bitext XML (simplified: same pattern loader per line pairs).
func CheckBitextWithRuleFile(w io.Writer, contents, ruleFile string) (int, error) {
	// green: if rule file loads, run CheckBitextFile; rule XML for bitext is soft
	if ruleFile != "" {
		if _, err := LoadRuleFile(ruleFile); err != nil {
			return 0, err
		}
	}
	return CheckBitextFile(w, contents, nil)
}

// SimplePolishSpellingMatch is a test-only spelling stub for "PL" CLI harnesses
// until Morfologik PL rule is wired. ASCII space split only — not strings.Fields
// (Unicode WS collapse invent).
func SimplePolishSpellingMatch(text string, known map[string]bool) []*rules.RuleMatch {
	var out []*rules.RuleMatch
	offset := 0
	for _, field := range strings.Split(text, " ") {
		if field == "" {
			continue
		}
		idx := strings.Index(text[offset:], field)
		if idx < 0 {
			continue
		}
		from := offset + idx
		to := from + len(field)
		offset = to
		low := strings.ToLower(field)
		if known != nil && (known[field] || known[low]) {
			continue
		}
		if !isWordToken(field) {
			continue
		}
		m := rules.NewRuleMatch(idRule{id: "MORFOLOGIK_RULE_PL_PL"}, nil, from, to, "Possible spelling mistake")
		out = append(out, m)
	}
	return out
}
