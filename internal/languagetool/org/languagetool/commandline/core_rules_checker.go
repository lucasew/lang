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
	enabledIDs := languagetool.ExpandSoftEnableRuleIDs(lt.GetAllRegisteredRuleIDs(), opts.GetEnabledRules())
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
		// soft: ensure expanded optional IDs are active under enabled-only
		for id := range enabled {
			lt.EnableRule(id)
		}
		return
	}
	// soft: --enable only re-enables previously disabled (Java also restricts category defaults)
	for _, id := range enabledIDs {
		lt.EnableRule(id)
	}
}

// RegisterSoftPickyGrammar loads {base}-picky-soft.xml from grammar dir when present.
// Used only for --level picky (or server picky boost). Prefer base language code
// (en, de, fr) so en-US still gets en-picky-soft.xml.
func RegisterSoftPickyGrammar(lt *languagetool.JLanguageTool, grammarDir, languageCode string) int {
	if lt == nil || grammarDir == "" {
		return 0
	}
	base := languageCode
	if i := strings.IndexByte(languageCode, '-'); i > 0 {
		base = languageCode[:i]
	}
	total := 0
	for _, name := range []string{
		base + "-picky-soft.xml",
		languageCode + "-picky-soft.xml",
	} {
		p := filepath.Join(grammarDir, name)
		if st, err := os.Stat(p); err != nil || !st.Mode().IsRegular() {
			continue
		}
		n, err := patterns.RegisterGrammarFile(lt, p, languageCode)
		if err != nil {
			continue
		}
		total += n
	}
	return total
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
			// soft picky level for English (core inject + soft XML pack below)
			en.RegisterPickyEnglishRules(lt)
		}
		// optional soft grammar directory (e.g. testdata/grammar) with walk-up discovery
		if dir := DiscoverGrammarDir(opts); dir != "" {
			_, _ = patterns.RegisterSoftGrammarDir(lt, dir, lang)
			if picky {
				// {base}-picky-soft.xml: pedantic style rules only when --level picky
				RegisterSoftPickyGrammar(lt, dir, lang)
			}
		}
		base := lang
		if i := strings.IndexByte(lang, '-'); i > 0 {
			base = lang[:i]
		}
		if strings.EqualFold(base, "en") {
			// Prefer CFSA2 en_US.dict when present; else optional map demo speller.
			demoSpell := os.Getenv("LANG_DEMO_SPELLER") == "1"
			// Always provide a small nearest set for edit-distance soft suggestions.
			nearest := en.DemoEnglishKnownWords()
			sugs := en.CommonDemoSpellerSuggestions
			if typoPath := DiscoverEnglishTyposFile(opts); typoPath != "" {
				if extra, err := en.LoadSoftTyposFile(typoPath); err == nil && len(extra) > 0 {
					sugs = en.MergeSpellerSuggestions(sugs, extra)
				}
			}
			spellRegistered := false
			if dictPath := DiscoverEnglishUSDict(opts); dictPath != "" {
				spellRegistered = en.RegisterBinaryEnglishSpeller(lt, dictPath, nearest, sugs)
			}
			if !spellRegistered && demoSpell {
				en.RegisterDemoEnglishSpeller(lt, nearest, sugs)
			}
			// Prefer CFSA2 english.dict POS tagger; else demo closed-class map under LANG_DEMO_SPELLER.
			taggerOK := false
			if posPath := DiscoverEnglishPOSDict(opts); posPath != "" {
				taggerOK = en.RegisterBinaryEnglishTagger(lt, posPath)
			}
			if !taggerOK && demoSpell {
				en.RegisterDemoEnglishTagger(lt)
			}
			// Soft multiword chunker + soft XML + ignore-spelling word list.
			en.RegisterSoftEnglishDisambiguator(lt,
				DiscoverEnglishMultiwords(opts),
				DiscoverEnglishSoftDisambiguationXML(opts),
				DiscoverEnglishIgnoreSpellingList(opts),
			)
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
		// Regional soft spelling packs (loaded only for matching lang codes).
		regional := []string{
			"en-US-soft.xml", "en-GB-soft.xml",
			"pt-BR-soft.xml", "pt-PT-soft.xml",
			"es-MX-soft.xml", "es-ES-soft.xml",
			"de-CH-soft.xml", "de-AT-soft.xml",
			"fr-CA-soft.xml",
		}
		regionalN := 0
		for _, name := range regional {
			p := filepath.Join(gdir, name)
			if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
				regionalN++
				_, _ = fmt.Fprintf(w, "soft spelling pack: %s\n", p)
			}
		}
		if regionalN > 0 {
			_, _ = fmt.Fprintf(w, "regional soft packs: %d\n", regionalN)
		}
		_, _ = fmt.Fprintf(w, "soft category filters: --disablecategories / --enablecategories\n")
		pickyN := 0
		for _, name := range []string{
			"en-picky-soft.xml", "de-picky-soft.xml", "fr-picky-soft.xml",
			"es-picky-soft.xml", "pt-picky-soft.xml", "it-picky-soft.xml",
			"nl-picky-soft.xml", "sv-picky-soft.xml", "pl-picky-soft.xml",
			"da-picky-soft.xml", "ru-picky-soft.xml",
			"uk-picky-soft.xml", "ca-picky-soft.xml",
			"el-picky-soft.xml", "ro-picky-soft.xml", "gl-picky-soft.xml",
			"sk-picky-soft.xml", "sl-picky-soft.xml",
			"be-picky-soft.xml", "sr-picky-soft.xml", "lt-picky-soft.xml",
			"is-picky-soft.xml", "ga-picky-soft.xml", "eo-picky-soft.xml",
			"fa-picky-soft.xml", "ar-picky-soft.xml", "zh-picky-soft.xml",
			"ja-picky-soft.xml", "br-picky-soft.xml", "ast-picky-soft.xml",
			"km-picky-soft.xml", "ta-picky-soft.xml", "tl-picky-soft.xml",
			"crh-picky-soft.xml", "ml-picky-soft.xml",
		} {
			pickySoft := filepath.Join(gdir, name)
			if st, err := os.Stat(pickySoft); err == nil && st.Mode().IsRegular() {
				pickyN++
				_, _ = fmt.Fprintf(w, "picky soft pack: %s (load with --level picky)\n", pickySoft)
			}
		}
		if pickyN > 0 {
			_, _ = fmt.Fprintf(w, "picky soft packs: %d\n", pickyN)
		}
		optN := 0
		for _, name := range []string{
			"en-optional-soft.xml", "de-optional-soft.xml", "fr-optional-soft.xml",
			"es-optional-soft.xml", "pt-optional-soft.xml", "it-optional-soft.xml",
			"nl-optional-soft.xml", "pl-optional-soft.xml", "sv-optional-soft.xml",
			"da-optional-soft.xml", "ru-optional-soft.xml", "uk-optional-soft.xml",
			"ca-optional-soft.xml", "el-optional-soft.xml", "ro-optional-soft.xml",
			"gl-optional-soft.xml", "sk-optional-soft.xml", "sl-optional-soft.xml",
			"be-optional-soft.xml", "sr-optional-soft.xml", "lt-optional-soft.xml",
			"is-optional-soft.xml", "ga-optional-soft.xml", "eo-optional-soft.xml",
			"fa-optional-soft.xml", "ar-optional-soft.xml", "zh-optional-soft.xml",
			"ja-optional-soft.xml", "br-optional-soft.xml", "ast-optional-soft.xml",
			"km-optional-soft.xml", "ta-optional-soft.xml", "tl-optional-soft.xml",
		} {
			optPath := filepath.Join(gdir, name)
			if st, err := os.Stat(optPath); err == nil && st.Mode().IsRegular() {
				optN++
				_, _ = fmt.Fprintf(w, "optional soft pack: %s (default off; enable with -e SOFT_OPTIONAL)\n", optPath)
			}
		}
		if optN > 0 {
			_, _ = fmt.Fprintf(w, "optional soft packs: %d\n", optN)
			_, _ = fmt.Fprintf(w, "soft optional enable: -e SOFT_OPTIONAL (or SOFT_OPT_ALL)\n")
		}
	}
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
	if ign := DiscoverEnglishIgnoreSpellingList(opts); ign != "" {
		_, _ = fmt.Fprintf(w, "ignore-spelling list: %s\n", ign)
	} else {
		_, _ = fmt.Fprintf(w, "ignore-spelling list: (unset)\n")
	}
	if dx := DiscoverEnglishSoftDisambiguationXML(opts); dx != "" {
		_, _ = fmt.Fprintf(w, "soft disambiguation XML: %s\n", dx)
	} else {
		_, _ = fmt.Fprintf(w, "soft disambiguation XML: (unset)\n")
	}
	if ty := DiscoverEnglishTyposFile(opts); ty != "" {
		_, _ = fmt.Fprintf(w, "en-typos.tsv: %s\n", ty)
	} else {
		_, _ = fmt.Fprintf(w, "en-typos.tsv: (unset)\n")
	}
	if mw := DiscoverEnglishMultiwords(opts); mw != "" {
		_, _ = fmt.Fprintf(w, "en multiwords: %s\n", mw)
	} else {
		_, _ = fmt.Fprintf(w, "en multiwords: (embedded soft defaults)\n")
	}
	// smoke check
	lt, err := configureCoreLT("en", opts)
	if err != nil {
		return err
	}
	ids := lt.GetAllRegisteredRuleIDs()
	softEN := 0
	for _, id := range ids {
		if strings.Contains(id, "_SOFT_") {
			softEN++
		}
	}
	_, _ = fmt.Fprintf(w, "en registered rules: %d\n", len(ids))
	_, _ = fmt.Fprintf(w, "en soft rules: %d\n", softEN)
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
	// soft grammar smoke: fused contraction
	msSoft := lt.Check("I wouldve gone.")
	softHit := false
	for _, m := range msSoft {
		if m.RuleID == "EN_SOFT_WOULDVE" {
			softHit = true
			break
		}
	}
	if softHit {
		_, _ = fmt.Fprintf(w, "en soft smoke: EN_SOFT_WOULDVE ok\n")
	} else {
		_, _ = fmt.Fprintf(w, "en soft smoke: EN_SOFT_WOULDVE missing\n")
	}
	// regional soft spelling smoke (only if packs load via grammar dir)
	if gdir != "" {
		if ltUS, err := configureCoreLT("en-US", opts); err == nil {
			hit := false
			for _, m := range ltUS.Check("Pick a colour.") {
				if m.RuleID == "EN_SOFT_COLOUR_US" {
					hit = true
					break
				}
			}
			if hit {
				_, _ = fmt.Fprintf(w, "en-US soft smoke: EN_SOFT_COLOUR_US ok\n")
			} else {
				_, _ = fmt.Fprintf(w, "en-US soft smoke: EN_SOFT_COLOUR_US missing\n")
			}
		}
		if ltGB, err := configureCoreLT("en-GB", opts); err == nil {
			hit := false
			for _, m := range ltGB.Check("Pick a color.") {
				if m.RuleID == "EN_SOFT_COLOR_GB" {
					hit = true
					break
				}
			}
			if hit {
				_, _ = fmt.Fprintf(w, "en-GB soft smoke: EN_SOFT_COLOR_GB ok\n")
			} else {
				_, _ = fmt.Fprintf(w, "en-GB soft smoke: EN_SOFT_COLOR_GB missing\n")
			}
		}
		// multi-lang soft pack smoke (walk-up / data-dir soft grammar)
		for _, sm := range []struct {
			lang, text, rule, label string
		}{
			{"de", "Ich denke das es so ist.", "DE_SOFT_DAS_DASS", "de soft smoke"},
			{"fr", "Je vais a la maison.", "FR_SOFT_A_LA", "fr soft smoke"},
			{"es", "Voy a el parque.", "ES_SOFT_A_EL", "es soft smoke"},
			{"pt", "Vou a o mercado.", "PT_SOFT_A_O", "pt soft smoke"},
			{"pt-BR", "Peguei o autocarro cedo.", "PT_SOFT_AUTOCARRO_BR", "pt-BR soft smoke"},
			{"pt-PT", "Peguei o ônibus cedo.", "PT_SOFT_ONIBUS_PT", "pt-PT soft smoke"},
			{"es-MX", "Uso el ordenador hoy.", "ES_SOFT_ORDENADOR_MX", "es-MX soft smoke"},
			{"es-ES", "Uso la computadora hoy.", "ES_SOFT_COMPUTADORA_ES", "es-ES soft smoke"},
			{"de-CH", "Die Straße ist nass.", "DE_SOFT_STRASSE_CH", "de-CH soft smoke"},
			{"de-AT", "Im Januar schneit es.", "DE_SOFT_JANUAR_AT", "de-AT soft smoke"},
			{"fr-CA", "Bon week-end à tous.", "FR_SOFT_WEEKEND_CA", "fr-CA soft smoke"},
		} {
			ltL, err := configureCoreLT(sm.lang, opts)
			if err != nil {
				continue
			}
			hit := false
			for _, m := range ltL.Check(sm.text) {
				if m.RuleID == sm.rule {
					hit = true
					break
				}
			}
			if hit {
				_, _ = fmt.Fprintf(w, "%s: %s ok\n", sm.label, sm.rule)
			} else {
				_, _ = fmt.Fprintf(w, "%s: %s missing\n", sm.label, sm.rule)
			}
		}
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
