package commandline

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
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
	// Java JLanguageTool: -e enables listed rule IDs; no SOFT_* invent expansion.
	enabledIDs := opts.GetEnabledRules()
	if opts.IsUseEnabledOnly() {
		enabled := map[string]struct{}{}
		for _, id := range enabledIDs {
			if id != "" {
				enabled[id] = struct{}{}
			}
		}
		for _, id := range lt.GetAllRegisteredRuleIDs() {
			if _, ok := enabled[id]; !ok {
				lt.DisableRule(id)
			}
		}
		for id := range enabled {
			lt.EnableRule(id)
		}
		return
	}
	for _, id := range enabledIDs {
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
		picky := strings.EqualFold(opts.Level, "PICKY")
		baseLang := lang
		if i := strings.IndexByte(lang, '-'); i > 0 {
			baseLang = lang[:i]
		}
		if picky && strings.EqualFold(baseLang, "en") {
			// Java English picky-level rules (not soft invent packs).
			en.RegisterPickyEnglishRules(lt)
		}
		// Soft grammar packs (testdata/*-soft.xml) are not loaded — faithful port only.
		// Official grammar.xml when LANG_USE_UPSTREAM_GRAMMAR=1. PatternRuleLoader
		// attaches registered RuleFilters; unknown filter classes skip the rule
		// (fail-closed). <antipattern> is loaded and applied in PatternRule.Match.
		base := lang
		if i := strings.IndexByte(lang, '-'); i > 0 {
			base = lang[:i]
		}
		if strings.EqualFold(base, "en") {
			// Prefer CFSA2 en_US.dict when present; demo only under LANG_DEMO_SPELLER.
			demoSpell := os.Getenv("LANG_DEMO_SPELLER") == "1"
			nearest := en.DemoEnglishKnownWords()
			spellRegistered := false
			if dictPath := DiscoverEnglishUSDict(opts); dictPath != "" {
				// Grammar filters (NumberInWord / FindSuggestions / SuppressMisspelled)
				// share the same dict Java MorfologikAmericanSpellerRule uses.
				_ = en.WireEnglishFilterSpeller(dictPath)
				// Binary speller: dict SuggestEdits only (no invent typo map).
				spellRegistered = en.RegisterBinaryEnglishSpeller(lt, dictPath, nearest, nil)
			}
			// Multitoken after filter dict so isMisspelled can use it.
			wireEnglishMultitokenSpeller(opts)
			if !spellRegistered && demoSpell {
				// Explicit demo-only path (not default engine).
				en.RegisterDemoEnglishSpeller(lt, nearest, en.CommonDemoSpellerSuggestions)
			}
			// Prefer CFSA2 english.dict POS tagger; else demo under LANG_DEMO_SPELLER.
			taggerOK := false
			if posPath := DiscoverEnglishPOSDict(opts); posPath != "" {
				taggerOK = en.RegisterBinaryEnglishTagger(lt, posPath)
				// FindSuggestionsFilter desiredPostag (Java EnglishTagger.INSTANCE).
				_ = en.WireEnglishFilterTagger(posPath)
			}
			if !taggerOK && demoSpell {
				en.RegisterDemoEnglishTagger(lt)
			}
			// Java English.createDefaultChunker().
			en.RegisterEnglishChunker(lt)
			// Java English.createDefaultDisambiguator(): EnglishHybridDisambiguator
			// (spelling_global → multiwords → XmlRuleDisambiguator with global).
			_ = RegisterHybridDisambiguator(lt, base, opts)
		} else {
			if posPath := DiscoverLanguagePOSDict(opts, base); posPath != "" {
				// Official Morfologik POS dicts (FSA5/CFSA2) like Java createDefaultTagger().
				if base == "ar" {
					_ = RegisterArabicPOSTagger(lt, posPath)
				} else {
					_ = languagetool.RegisterBinaryPOSTagger(lt, posPath)
				}
			}
			// Java createDefaultDisambiguator(): FR/ES/PT hybrids when resources exist.
			_ = RegisterHybridDisambiguator(lt, base, opts)
		}
		// Grammar after multitoken speller so MultitokenSpellerFilter can use the dict.
		if os.Getenv("LANG_USE_UPSTREAM_GRAMMAR") == "1" {
			if gpath := DiscoverLanguageGrammarXML(opts, base); gpath != "" {
				_, _ = patterns.RegisterGrammarFile(lt, gpath, lang)
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
		if words := opts.GetIgnoreWords(); len(words) > 0 {
			lt.AddIgnoreWords(words...)
		}
		ApplyCLIRuleFilters(lt, opts)
	}
	return lt, nil
}

// wireEnglishMultitokenSpeller ports English.getMultitokenSpeller resource load
// (multiwords.txt + spelling_global.txt) into MultitokenSpellerFilter.
func wireEnglishMultitokenSpeller(opts *CommandLineOptions) {
	mw := DiscoverEnglishMultiwords(opts)
	sg := DiscoverSpellingGlobal(opts)
	if mw == "" && sg == "" {
		return
	}
	sp, err := en.LoadEnglishMultitokenSpeller(mw, sg)
	if err != nil || sp == nil || sp.MultitokenSpeller == nil {
		return
	}
	// Java MultitokenSpellerFilter.isMisspelled uses language.getDefaultSpellingRule().
	// When en_US.dict is wired, FilterDictIsMisspelled matches that; else nil-like (false).
	var isMiss func(string) bool
	if en.FilterDictAvailable() {
		isMiss = en.FilterDictIsMisspelled
	}
	patterns.SetDefaultMultitokenSpeller(sp.MultitokenSpeller, isMiss)
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
	// Prefer higher-priority / non-overlapping spans so soft+speller on the same
	// token do not undo each other (e.g. EN_SOFT_ALOT "a lot" vs MORFOLOGIK "lot").
	ms = languagetool.CleanOverlappingLocalMatches(ms)
	// apply non-overlapping suggestions in document order
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

// CoreTagHook prints a word/lemma/POS dump for --taggeronly (uses configureCoreLT tagger).
func CoreTagHook(w io.Writer, text string, opts *CommandLineOptions) error {
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
		return err
	}
	sents := lt.Analyze(text)
	for _, s := range sents {
		if s == nil {
			continue
		}
		var parts []string
		for _, t := range s.GetTokensWithoutWhitespace() {
			if t == nil || t.IsSentenceStart() || t.IsSentenceEnd() {
				continue
			}
			parts = append(parts, FormatTaggedToken(t))
		}
		_, _ = fmt.Fprintln(w, FormatTagLine(s.GetText(), parts))
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
// Soft rules (ID contains _SOFT_) append a fifth column "soft" for easy filtering.
// Soft rules are listed first (sorted), then core (sorted). Footer counts soft by issue type.
func CoreListRules(w io.Writer, lang string) error {
	return CoreListRulesOpts(w, &CommandLineOptions{Language: lang})
}

// CoreListRulesOpts is like CoreListRules but honors opts (e.g. Level=PICKY for picky soft packs).
func CoreListRulesOpts(w io.Writer, opts *CommandLineOptions) error {
	if w == nil {
		return nil
	}
	if opts == nil {
		opts = &CommandLineOptions{Language: "en"}
	}
	lang := opts.Language
	if lang == "" {
		lang = "en"
	}
	// ensure Language is set for configureCoreLT soft discovery
	if opts.Language == "" {
		opts.Language = lang
	}
	lt, err := configureCoreLT(lang, opts)
	if err != nil {
		return err
	}
	ids := lt.GetAllRegisteredRuleIDs()
	active := map[string]struct{}{}
	for _, id := range lt.GetAllActiveRuleIDs() {
		active[id] = struct{}{}
	}
	var softIDs, coreIDs []string
	for _, id := range ids {
		if strings.Contains(id, "_SOFT_") {
			softIDs = append(softIDs, id)
		} else {
			coreIDs = append(coreIDs, id)
		}
	}
	sort.Strings(softIDs)
	sort.Strings(coreIDs)
	ordered := append(softIDs, coreIDs...)
	softN := len(softIDs)
	softByIssue := map[string]int{}
	pickySoftN := 0
	optSoftN := 0
	softOffN := 0
	for _, id := range ordered {
		cat, _, issue, _ := languagetool.SoftRuleMeta(id)
		if cat == "" {
			cat = "MISC"
		}
		if issue == "" {
			issue = "uncategorized"
		}
		url := languagetool.SoftRuleURL(id, lang)
		kind := "core"
		if strings.Contains(id, "_SOFT_") {
			kind = "soft"
			softByIssue[issue]++
			if strings.Contains(id, "SOFT_PICKY") {
				pickySoftN++
			}
			if strings.Contains(id, "SOFT_OPT_") {
				optSoftN++
			}
		}
		// soft: sixth column on|off (default="off" soft rules start off)
		state := "on"
		if _, ok := active[id]; !ok {
			state = "off"
			if kind == "soft" {
				softOffN++
			}
		}
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", id, cat, issue, url, kind, state); err != nil {
			return err
		}
	}
	// stable soft breakdown for tooling / doctor-adjacent use
	parts := []string{
		fmt.Sprintf("total=%d", len(ids)),
		fmt.Sprintf("soft=%d", softN),
	}
	for _, k := range []string{"grammar", "style", "misspelling", "typographical", "whitespace", "uncategorized"} {
		if n := softByIssue[k]; n > 0 {
			parts = append(parts, fmt.Sprintf("soft_%s=%d", k, n))
		}
	}
	if pickySoftN > 0 {
		parts = append(parts, fmt.Sprintf("soft_picky=%d", pickySoftN))
	}
	if optSoftN > 0 {
		parts = append(parts, fmt.Sprintf("soft_opt=%d", optSoftN))
	}
	if softOffN > 0 {
		parts = append(parts, fmt.Sprintf("soft_off=%d", softOffN))
	}
	if strings.EqualFold(opts.Level, "PICKY") {
		parts = append(parts, "level=picky")
	}
	_, err = fmt.Fprintf(w, "# %s\n", strings.Join(parts, " "))
	return err
}

// DefaultCoreHooks returns RunHooks wired to the pure-Go core packs.
func DefaultCoreHooks() RunHooks {
	return RunHooks{
		Check:         CoreCheckHook,
		Tag:           CoreTagHook,
		ListLanguages: CoreListLanguages,
	}
}

// CoreDoctor writes environment diagnostics (SPEC §2.3 doctor).
func CoreDoctor(w io.Writer, opts *CommandLineOptions) error {
	if w == nil {
		return nil
	}
	_, _ = fmt.Fprintf(w, "lang doctor\n")
	_, _ = fmt.Fprintf(w, "version: %s\n", VersionString)
	_, _ = fmt.Fprintf(w, "go: %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	_, _ = fmt.Fprintf(w, "corepack languages: %d\n", len(corepack.Supported))
	// Soft grammar packs are not part of the faithful engine; report real resources only.
	ff := DiscoverFalseFriendsFile(opts)
	if ff == "" {
		ff = "(unset)"
	}
	_, _ = fmt.Fprintf(w, "false-friends: %s\n", ff)
	if dict := DiscoverEnglishUSDict(opts); dict != "" {
		_, _ = fmt.Fprintf(w, "en_US.dict: %s\n", dict)
	} else {
		_, _ = fmt.Fprintf(w, "en_US.dict: (unset)\n")
	}
	if pos := DiscoverEnglishPOSDict(opts); pos != "" {
		_, _ = fmt.Fprintf(w, "english.dict: %s\n", pos)
	} else {
		_, _ = fmt.Fprintf(w, "english.dict: (unset)\n")
	}
	if mw := DiscoverEnglishMultiwords(opts); mw != "" {
		_, _ = fmt.Fprintf(w, "en multiwords: %s\n", mw)
	} else {
		_, _ = fmt.Fprintf(w, "en multiwords: (unset)\n")
	}
	// smoke check (core pack only — no soft invent packs)
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
		if m.IssueType != "" {
			issue = m.IssueType
		}
		if m.CategoryID != "" {
			catID = m.CategoryID
		}
		if m.CategoryName != "" {
			catName = m.CategoryName
		}
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
