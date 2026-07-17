package commandline

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/corepack"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// CoreRulesChecker implements TextChecker using RegisterCore* packs.
type CoreRulesChecker struct {
	Lang          string
	lt            *languagetool.JLanguageTool
	CleanOverlaps bool
}

// NewCoreRulesChecker builds a checker for lang (e.g. "en", "en-US", "de-DE").
func NewCoreRulesChecker(lang string) *CoreRulesChecker {
	if lang == "" {
		lang = "en"
	}
	lt := languagetool.NewJLanguageTool(lang)
	corepack.Register(lt, lang)
	return &CoreRulesChecker{Lang: lang, lt: lt}
}

// Check runs the core rule pack and returns rules.RuleMatch for CLI printing.
func (c *CoreRulesChecker) Check(text string) ([]*rules.RuleMatch, error) {
	if c == nil || c.lt == nil {
		return nil, nil
	}
	ms := c.lt.Check(text)
	if c.CleanOverlaps {
		for i := range ms {
			id := ms[i].RuleID
			if id == "EN_A_VS_AN" || strings.Contains(id, "WORD_REPEAT") {
				ms[i].Priority = 5
			} else if ms[i].Priority == 0 {
				ms[i].Priority = 1
			}
		}
		ms = languagetool.CleanOverlappingLocalMatches(ms)
	}
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

// ApplyCLIRuleFilters applies -d/-e/--enabledonly from CommandLineOptions.
func ApplyCLIRuleFilters(lt *languagetool.JLanguageTool, opts *CommandLineOptions) {
	if lt == nil || opts == nil {
		return
	}
	for _, id := range opts.GetDisabledRules() {
		lt.DisableRule(id)
	}
	if opts.IsUseEnabledOnly() {
		enabled := map[string]struct{}{}
		for _, id := range opts.GetEnabledRules() {
			if id != "" {
				enabled[id] = struct{}{}
			}
		}
		for _, id := range lt.GetAllRegisteredRuleIDs() {
			if _, ok := enabled[id]; !ok {
				lt.DisableRule(id)
			}
		}
		return
	}
	// soft: --enable only re-enables previously disabled (Java also restricts category defaults)
	for _, id := range opts.GetEnabledRules() {
		lt.EnableRule(id)
	}
}

// RegisterRuleFilePatterns loads --rulefile XML pattern rules onto lt.
func RegisterRuleFilePatterns(lt *languagetool.JLanguageTool, ruleFile, lang string) error {
	if lt == nil || ruleFile == "" {
		return nil
	}
	xml, err := LoadRuleFile(ruleFile)
	if err != nil {
		return err
	}
	if lang == "" {
		lang = "en"
	}
	loader := patterns.NewPatternRuleLoader()
	abstracts, err := loader.GetRulesFromString(xml, ruleFile, lang)
	if err != nil {
		return err
	}
	for _, ar := range abstracts {
		if ar == nil {
			continue
		}
		pr := patterns.NewPatternRule(ar.ID, ar.LanguageCode, ar.PatternTokens, ar.Description, ar.Message, ar.ShortMessage)
		patterns.RegisterPatternRule(lt, pr)
	}
	return nil
}


// RegisterFalseFriends loads --falsefriends XML pattern rules for mother-tongue pairs.
func RegisterFalseFriends(lt *languagetool.JLanguageTool, falseFriendsFile, textLang, motherLang string) error {
	if lt == nil || falseFriendsFile == "" || motherLang == "" {
		return nil
	}
	data, err := os.ReadFile(falseFriendsFile)
	if err != nil {
		return err
	}
	if textLang == "" {
		textLang = "en"
	}
	loader := patterns.NewFalseFriendRuleLoader("", "")
	ffRules, err := loader.GetRulesFromString(string(data), textLang, motherLang)
	if err != nil {
		return err
	}
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
		// capture for closure
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
						m.SetSuggestedReplacements(suggestions)
					}
				}
			}
			return rules.ToLocalMatches(ms)
		})
	}
	return nil
}

// configureCoreLT builds a language tool with core pack + optional rulefile/false friends + CLI filters.
func configureCoreLT(lang string, opts *CommandLineOptions) (*languagetool.JLanguageTool, error) {
	checker := NewCoreRulesChecker(lang)
	lt := checker.lt
	if opts != nil {
		if opts.GetRuleFile() != "" {
			if err := RegisterRuleFilePatterns(lt, opts.GetRuleFile(), lang); err != nil {
				return nil, err
			}
		}
		if opts.FalseFriendsFile != "" && opts.MotherTongue != "" {
			if err := RegisterFalseFriends(lt, opts.FalseFriendsFile, lang, opts.MotherTongue); err != nil {
				return nil, err
			}
		}
		ApplyCLIRuleFilters(lt, opts)
	}
	return lt, nil
}

// CoreCheckHook is a RunHooks.Check that uses CoreRulesChecker + CheckTextOpts.
// Honors XML filter, auto-detect language, disable/enable, --rulefile, --falsefriends, line mode, and --apply.
func CoreCheckHook(w io.Writer, text string, opts *CommandLineOptions) (int, error) {
	if opts != nil && opts.XMLFiltering {
		text = MaybeFilterXML(text, true)
	}
	lang := "en"
	if opts != nil {
		if opts.Language != "" {
			lang = opts.Language
		}
		if opts.IsAutoDetect() {
			lang = ResolveLanguage(text, opts, DetectLanguageHeuristic)
		}
	}
	if opts != nil && opts.IsApplySuggestions() {
		return CoreApplySuggestionsHook(w, text, opts)
	}

	lt, err := configureCoreLT(lang, opts)
	if err != nil {
		return 0, err
	}

	if opts != nil && opts.LineByLine {
		return CheckLineByLine(w, text, func(seg string) ([]*rules.RuleMatch, error) {
			ms := lt.Check(seg)
			sent := languagetool.AnalyzePlain(seg)
			return rules.FromLocalMatches(ms, sent), nil
		})
	}

	checker := &CoreRulesChecker{Lang: lang, lt: lt, CleanOverlaps: opts != nil && opts.CleanOverlapping}
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
	if cto.ListUnknown {
		lt.SetListUnknownWords(true)
	}
	return CheckTextOpts(w, text, checker, cto)
}

// CoreApplySuggestionsHook writes text with first suggestions applied.
func CoreApplySuggestionsHook(w io.Writer, text string, opts *CommandLineOptions) (int, error) {
	if opts != nil && opts.XMLFiltering {
		text = MaybeFilterXML(text, true)
	}
	lang := "en"
	if opts != nil {
		if opts.Language != "" {
			lang = opts.Language
		}
		if opts.IsAutoDetect() {
			lang = ResolveLanguage(text, opts, DetectLanguageHeuristic)
		}
	}
	lt, err := configureCoreLT(lang, opts)
	if err != nil {
		return 0, err
	}
	ms := lt.Check(text)
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

// CoreTagHook prints a soft word-token tag dump for --taggeronly.
func CoreTagHook(w io.Writer, text string, opts *CommandLineOptions) error {
	if opts != nil && opts.XMLFiltering {
		text = MaybeFilterXML(text, true)
	}
	lang := "en"
	if opts != nil && opts.Language != "" {
		lang = opts.Language
	}
	if opts != nil && opts.IsAutoDetect() {
		lang = ResolveLanguage(text, opts, DetectLanguageHeuristic)
	}
	lt := languagetool.NewJLanguageTool(lang)
	sents := lt.Analyze(text)
	for _, s := range sents {
		if s == nil {
			continue
		}
		var toks []string
		for _, t := range s.GetTokensWithoutWhitespace() {
			if t == nil || t.IsSentenceStart() || t.IsSentenceEnd() {
				continue
			}
			toks = append(toks, t.GetToken())
		}
		_, _ = fmt.Fprintln(w, FormatTagLine(s.GetText(), toks))
	}
	return nil
}

// CoreListLanguages writes corepack-supported language codes (one per line).
func CoreListLanguages(w io.Writer) error {
	for _, s := range corepack.Supported {
		// prefer variant-style codes when common
		line := s.Code
		switch s.Code {
		case "en":
			line = "en-US"
		case "de":
			line = "de-DE"
		case "pt":
			line = "pt-BR"
		case "uk":
			line = "uk-UA"
		}
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}
	return nil
}

// DefaultCoreHooks returns RunHooks wired to the pure-Go core packs.
func DefaultCoreHooks() RunHooks {
	return RunHooks{
		Check:         CoreCheckHook,
		Tag:           CoreTagHook,
		ListLanguages: CoreListLanguages,
	}
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
