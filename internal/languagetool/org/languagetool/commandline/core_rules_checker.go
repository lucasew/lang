package commandline

import (
	"io"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/de"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/en"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// CoreRulesChecker implements TextChecker using RegisterCore* packs.
type CoreRulesChecker struct {
	Lang string
	lt   *languagetool.JLanguageTool
}

// NewCoreRulesChecker builds a checker for lang (e.g. "en", "en-US", "de-DE").
func NewCoreRulesChecker(lang string) *CoreRulesChecker {
	if lang == "" {
		lang = "en"
	}
	lt := languagetool.NewJLanguageTool(lang)
	registerCoreForLang(lt, lang)
	return &CoreRulesChecker{Lang: lang, lt: lt}
}

func registerCoreForLang(lt *languagetool.JLanguageTool, lang string) {
	base := lang
	if i := strings.IndexByte(lang, '-'); i > 0 {
		base = lang[:i]
	}
	switch base {
	case "en":
		en.RegisterCoreEnglishLanguageRules(lt)
	case "de":
		de.RegisterCoreGermanRules(lt)
	default:
		rules.RegisterCoreRules(lt, lang)
	}
}

// Check runs the core rule pack and returns rules.RuleMatch for CLI printing.
func (c *CoreRulesChecker) Check(text string) ([]*rules.RuleMatch, error) {
	if c == nil || c.lt == nil {
		return nil, nil
	}
	ms := c.lt.Check(text)
	// Attach full-text analysis for print context (soft single blob)
	sent := languagetool.AnalyzePlain(text)
	return rules.FromLocalMatches(ms, sent), nil
}

// LT exposes the underlying JLanguageTool for disable/enable in tests.
func (c *CoreRulesChecker) LT() *languagetool.JLanguageTool {
	if c == nil {
		return nil
	}
	return c.lt
}

// CoreCheckHook is a RunHooks.Check that uses CoreRulesChecker + CheckTextOpts.
func CoreCheckHook(w io.Writer, text string, opts *CommandLineOptions) (int, error) {
	lang := "en"
	if opts != nil && opts.Language != "" {
		lang = opts.Language
	}
	checker := NewCoreRulesChecker(lang)
	cto := CheckTextOptions{
		JSON:        opts != nil && opts.OutputFormat == OutputJSON,
		Verbose:     opts != nil && opts.Verbose,
		ListUnknown: opts != nil && opts.IsListUnknown(),
	}
	if cto.JSON {
		cto.JSONSerializer = func(matches []*rules.RuleMatch, contents string, contextSize int) string {
			return ruleMatchesToJSON(matches, contents, contextSize, lang)
		}
	}
	// unknown words if requested
	if cto.ListUnknown {
		checker.lt.SetListUnknownWords(true)
		// soft: without dict, leave empty
	}
	return CheckTextOpts(w, text, checker, cto)
}

// CoreApplySuggestionsHook writes text with first suggestions applied.
func CoreApplySuggestionsHook(w io.Writer, text string, opts *CommandLineOptions) (int, error) {
	lang := "en"
	if opts != nil && opts.Language != "" {
		lang = opts.Language
	}
	checker := NewCoreRulesChecker(lang)
	ms := checker.lt.Check(text)
	// apply all with suggestions in document order
	var withSug []languagetool.LocalMatch
	for _, m := range ms {
		if len(m.Suggestions) > 0 {
			withSug = append(withSug, m)
		}
	}
	fixed := languagetool.CorrectTextFromLocalMatches(text, withSug)
	_, _ = io.WriteString(w, fixed)
	if !strings.HasSuffix(fixed, "\n") {
		_, _ = io.WriteString(w, "\n")
	}
	return len(withSug), nil
}

func ruleMatchesToJSON(matches []*rules.RuleMatch, contents string, contextSize int, lang string) string {
	s := tools.NewRuleMatchesAsJsonSerializer()
	s.LanguageCode = lang
	s.LanguageName = lang
	var mj []tools.MatchForJSON
	for _, m := range matches {
		if m == nil {
			continue
		}
		item := tools.MatchForJSON{
			Message:               m.GetMessage(),
			FromPos:               m.GetFromPos(),
			ToPos:                 m.GetToPos(),
			SuggestedReplacements: m.GetSuggestedReplacements(),
		}
		if g, ok := m.Rule.(interface{ GetID() string }); ok {
			item.RuleID = g.GetID()
		}
		mj = append(mj, item)
	}
	out, err := s.RuleMatchesToJSON(mj, contents, contextSize)
	if err != nil {
		return matchesToMinimalJSON(matches)
	}
	return out
}
