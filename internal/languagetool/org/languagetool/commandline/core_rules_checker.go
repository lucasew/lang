package commandline

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/corepack"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/en"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// CoreRulesChecker implements TextChecker using RegisterCore* packs.
type CoreRulesChecker struct {
	Lang               string
	lt                 *languagetool.JLanguageTool
	CleanOverlaps      bool
	DisabledCategories []string
	EnabledCategories  []string
	UseEnabledOnly     bool
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
	ms = languagetool.FilterMatchesByCategories(ms, c.DisabledCategories, c.EnabledCategories, c.UseEnabledOnly)
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
	_, err := patterns.RegisterFalseFriendsFile(lt, falseFriendsFile, textLang, motherLang)
	return err
}

// configureCoreLT builds a language tool with core pack + optional rulefile/false friends + CLI filters.
func configureCoreLT(lang string, opts *CommandLineOptions) (*languagetool.JLanguageTool, error) {
	checker := NewCoreRulesChecker(lang)
	lt := checker.lt
	if opts != nil {
		// soft picky level for English
		if strings.EqualFold(opts.Level, "PICKY") {
			base := lang
			if i := strings.IndexByte(lang, '-'); i > 0 {
				base = lang[:i]
			}
			if strings.EqualFold(base, "en") {
				en.RegisterPickyEnglishRules(lt)
			}
		}
		// optional soft grammar directory (e.g. testdata/grammar) with walk-up discovery
		if dir := DiscoverGrammarDir(opts); dir != "" {
			_, _ = patterns.RegisterSoftGrammarDir(lt, dir, lang)
		}
		if os.Getenv("LANG_DEMO_SPELLER") == "1" {
			base := lang
			if i := strings.IndexByte(lang, '-'); i > 0 {
				base = lang[:i]
			}
			if strings.EqualFold(base, "en") {
				en.RegisterDemoEnglishSpeller(lt, en.DemoEnglishKnownWords(), map[string][]string{
					"teh": {"the"}, "recieve": {"receive"},
				})
				en.RegisterDemoEnglishTagger(lt)
			}
		}
		if opts.GetRuleFile() != "" {
			if err := RegisterRuleFilePatterns(lt, opts.GetRuleFile(), lang); err != nil {
				return nil, err
			}
		}
		ffFile := opts.FalseFriendsFile
		if ffFile == "" && opts.MotherTongue != "" {
			ffFile = DiscoverFalseFriendsFile(opts)
		}
		if ffFile != "" && opts.MotherTongue != "" {
			if err := RegisterFalseFriends(lt, ffFile, lang, opts.MotherTongue); err != nil {
				return nil, err
			}
		}
		ApplyCLIRuleFilters(lt, opts)
	}
	return lt, nil
}

// resolveGrammarDir prefers --data-dir/grammar, then LANG_GRAMMAR_DIR, then LANG_DATA_DIR/grammar.
func resolveGrammarDir(opts *CommandLineOptions) string {
	if opts != nil && opts.GetDataDir() != "" {
		return filepath.Join(opts.GetDataDir(), "grammar")
	}
	if dir := os.Getenv("LANG_GRAMMAR_DIR"); dir != "" {
		return dir
	}
	if dir := os.Getenv("LANG_DATA_DIR"); dir != "" {
		return filepath.Join(dir, "grammar")
	}
	return ""
}

// resolveFalseFriendsFile prefers LANG_FALSEFRIENDS_FILE, then --data-dir/false-friends-soft.xml.
func resolveFalseFriendsFile(opts *CommandLineOptions) string {
	if p := os.Getenv("LANG_FALSEFRIENDS_FILE"); p != "" {
		return p
	}
	if opts != nil && opts.GetDataDir() != "" {
		return filepath.Join(opts.GetDataDir(), "false-friends-soft.xml")
	}
	if dir := os.Getenv("LANG_DATA_DIR"); dir != "" {
		return filepath.Join(dir, "false-friends-soft.xml")
	}
	return ""
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

	checker := &CoreRulesChecker{
		Lang:          lang,
		lt:            lt,
		CleanOverlaps: opts != nil && opts.CleanOverlapping,
	}
	if opts != nil {
		checker.DisabledCategories = append([]string(nil), opts.DisabledCategories...)
		checker.EnabledCategories = append([]string(nil), opts.EnabledCategories...)
		checker.UseEnabledOnly = opts.IsUseEnabledOnly()
	}
	cto := CheckTextOptions{
		JSON:        opts != nil && (opts.OutputFormat == OutputJSON || opts.OutputFormat == OutputSARIF),
		Lint:        opts != nil && opts.OutputFormat == OutputLint,
		Verbose:     opts != nil && opts.Verbose,
		ListUnknown: opts != nil && opts.IsListUnknown(),
	}
	if opts != nil {
		cto.Filename = opts.Filename
	}
	if cto.JSON {
		if opts != nil && opts.OutputFormat == OutputSARIF {
			fn := opts.Filename
			cto.JSONSerializer = func(matches []*rules.RuleMatch, contents string, contextSize int) string {
				return MatchesAsSARIF(matches, contents, fn, lang)
			}
		} else {
			cto.JSONSerializer = func(matches []*rules.RuleMatch, contents string, contextSize int) string {
				return ruleMatchesToJSON(matches, contents, contextSize, lang)
			}
		}
	}
	if cto.ListUnknown {
		lt.SetListUnknownWords(true)
	}
	// Soft ruleValues (e.g. TOO_LONG_SENTENCE:10) applied after the core pack check.
	var checkerRun TextChecker = checker
	if opts != nil && len(opts.GetRuleValues()) > 0 {
		checkerRun = &ruleValuesChecker{
			inner:  checker,
			lang:   lang,
			values: opts.GetRuleValues(),
		}
	}
	n, err := CheckTextOpts(w, text, checkerRun, cto)
	if err != nil {
		return n, err
	}
	if opts != nil && (opts.OutputFormat == OutputLint || opts.OutputFormat == OutputSARIF) {
		ms, _ := checkerRun.Check(text)
		return countFailOnMatches(ms, opts.GetFailOn()), nil
	}
	return n, nil
}

// ruleValuesChecker wraps a TextChecker and applies soft ruleValues post-processing.
type ruleValuesChecker struct {
	inner  TextChecker
	lang   string
	values []string
}

func (c *ruleValuesChecker) Check(text string) ([]*rules.RuleMatch, error) {
	if c == nil || c.inner == nil {
		return nil, nil
	}
	ms, err := c.inner.Check(text)
	if err != nil {
		return ms, err
	}
	return applyCLIRuleValues(c.lang, text, ms, c.values), nil
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

// CoreListRules writes registered rule IDs for lang (one per line), with soft category.
func CoreListRules(w io.Writer, lang string) error {
	if w == nil {
		return nil
	}
	if lang == "" {
		lang = "en"
	}
	lt, err := configureCoreLT(lang, &CommandLineOptions{Language: lang})
	if err != nil {
		return err
	}
	ids := lt.GetAllRegisteredRuleIDs()
	for _, id := range ids {
		cat, _, issue, _ := languagetool.SoftRuleMeta(id)
		if cat == "" {
			cat = "MISC"
		}
		if issue == "" {
			issue = "uncategorized"
		}
		url := languagetool.SoftRuleURL(id, lang)
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", id, cat, issue, url); err != nil {
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

// CoreDoctor writes soft environment diagnostics (SPEC §2.3 doctor).
func CoreDoctor(w io.Writer, opts *CommandLineOptions) error {
	if w == nil {
		return nil
	}
	_, _ = fmt.Fprintf(w, "lang doctor\n")
	_, _ = fmt.Fprintf(w, "version: %s\n", VersionString)
	_, _ = fmt.Fprintf(w, "go: %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	_, _ = fmt.Fprintf(w, "corepack languages: %d\n", len(corepack.Supported))
	gdir := DiscoverGrammarDir(opts)
	if gdir == "" {
		_, _ = fmt.Fprintf(w, "grammar dir: (unset)\n")
	} else {
		softN := countSoftGrammarFiles(gdir)
		_, _ = fmt.Fprintf(w, "grammar dir: %s\n", gdir)
		_, _ = fmt.Fprintf(w, "soft grammar files: %d\n", softN)
	}
	ff := DiscoverFalseFriendsFile(opts)
	if ff == "" {
		ff = "(unset)"
	}
	_, _ = fmt.Fprintf(w, "false-friends: %s\n", ff)
	// smoke check
	lt, err := configureCoreLT("en", opts)
	if err != nil {
		return err
	}
	ids := lt.GetAllRegisteredRuleIDs()
	_, _ = fmt.Fprintf(w, "en registered rules: %d\n", len(ids))
	ms := lt.Check("This is an test.")
	_, _ = fmt.Fprintf(w, "en smoke matches: %d\n", len(ms))
	found := false
	for _, m := range ms {
		if m.RuleID == "EN_A_VS_AN" {
			found = true
			break
		}
	}
	if found {
		_, _ = fmt.Fprintf(w, "en smoke: EN_A_VS_AN ok\n")
	} else {
		_, _ = fmt.Fprintf(w, "en smoke: EN_A_VS_AN missing\n")
	}
	_, _ = fmt.Fprintf(w, "status: ok\n")
	return nil
}

// countSoftGrammarFiles counts *-soft.xml (and grammar-soft.xml) under dir.
func countSoftGrammarFiles(dir string) int {
	if dir == "" || dir == "(unset)" {
		return 0
	}
	ents, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	n := 0
	for _, e := range ents {
		if e.IsDir() {
			// {lang}/grammar-soft.xml layout
			p := filepath.Join(dir, e.Name(), "grammar-soft.xml")
			if st, err := os.Stat(p); err == nil && !st.IsDir() {
				n++
			}
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, "-soft.xml") || name == "grammar-soft.xml" {
			n++
		}
	}
	return n
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
		id := ""
		if g, ok := m.Rule.(interface{ GetID() string }); ok {
			id = g.GetID()
		}
		catID, catName, issue, short := languagetool.SoftRuleMeta(id)
		sm := m.GetShortMessage()
		if sm == "" {
			sm = short
		}
		item := tools.MatchForJSON{
			Message:               m.GetMessage(),
			ShortMessage:          sm,
			FromPos:               m.GetFromPos(),
			ToPos:                 m.GetToPos(),
			SuggestedReplacements: m.GetSuggestedReplacements(),
			RuleID:                id,
			RuleDescription:       languagetool.SoftRuleDescription(id),
			IssueType:             issue,
			CategoryID:            catID,
			CategoryName:          catName,
			Severity:              languagetool.SeverityFromIssueType(issue),
			RuleURL:               languagetool.SoftRuleURL(id, lang),
		}
		mj = append(mj, item)
	}
	out, err := s.RuleMatchesToJSON(mj, contents, contextSize)
	if err != nil {
		return matchesToMinimalJSON(matches)
	}
	return out
}
