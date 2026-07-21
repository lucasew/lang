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
	rulesca "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ca"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/de"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/en"
	ruleses "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/es"
	rulesfr "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/fr"
	rulesga "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ga"
	rulesnl "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/nl"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	rulespt "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/pt"
	rulesru "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ru"
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
	// Java Tools.selectRules(…, enableTempOff): activate default='temp_off' rules first.
	if opts.IsEnableTempOff() {
		lt.EnableTempOffRules()
	}
	// Category enable/disable (Java disableCategory / enableRuleCategory).
	for _, id := range opts.GetDisabledCategories() {
		lt.DisableCategory(id)
	}
	for _, id := range opts.GetEnabledCategories() {
		lt.EnableCategory(id)
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
		pr.UnifierConfig = ar.UnifierConfig
		pr.AntiPatterns = append([]*patterns.PatternRule(nil), ar.AntiPatterns...)
		pr.Filter = ar.Filter
		pr.FilterArgs = ar.FilterArgs
		pr.LineNumber = ar.LineNumber
		pr.SourceFile = ar.SourceFile
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
		// Java JLanguageTool.setLevel: DEFAULT filters Tag.picky (false friends,
		// PROFANITY, long sentence, …). Tag.picky rules live in core packs —
		// no invent RegisterPickyEnglishRules / picky-soft packs.
		if strings.EqualFold(opts.Level, "PICKY") {
			lt.Level = languagetool.LevelPicky
		}
		// Soft grammar packs (testdata/*-soft.xml) are not loaded — faithful port only.
		// Official grammar.xml by default (Java getRuleFileNames); opt out with
		// LANG_USE_UPSTREAM_GRAMMAR=0. PatternRuleLoader attaches registered
		// RuleFilters; unknown filter classes skip the rule (fail-closed).
		// <antipattern> is loaded and applied in PatternRule.Match.
		base := lang
		if i := strings.IndexByte(lang, '-'); i > 0 {
			base = lang[:i]
		}
		if strings.EqualFold(base, "en") {
			// Prefer CFSA2 locale hunspell/*.dict when present; demo only under LANG_DEMO_SPELLER.
			// RegisterCoreEnglishLanguageRules already installs the binary speller when dict is found;
			// only fill in here if core pack did not (e.g. dict discovered via --data-dir only).
			demoSpell := os.Getenv("LANG_DEMO_SPELLER") == "1"
			nearest := en.DemoEnglishKnownWords()
			spellRegistered := false
			ruleID, _ := en.EnglishVariantSpellerMeta(lang)
			for _, id := range lt.GetAllRegisteredRuleIDs() {
				if id == ruleID {
					spellRegistered = true
					break
				}
			}
			if dictPath := DiscoverEnglishVariantDict(opts, lang); dictPath != "" {
				// Grammar filters (NumberInWord / FindSuggestions / SuppressMisspelled)
				// share the same dict Java Morfologik*SpellerRule uses.
				_ = en.WireEnglishFilterSpeller(dictPath)
				// Binary speller: dict SuggestEdits only (no invent typo map).
				// Skip second registration when RegisterCore already wired this ID.
				if !spellRegistered {
					spellRegistered = en.RegisterBinaryEnglishSpellerID(lt, dictPath, ruleID, nearest, nil)
				}
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
			// Java English.createDefaultSynthesizer() → EnglishSynthesizer.INSTANCE
			// for pattern <match postag="…"/> suggestions (RegisterLanguageSynthesizer).
			if synthPath := DiscoverEnglishSynthDict(opts); synthPath != "" {
				if synth := OpenLanguageSynthesizer("en", synthPath); synth != nil {
					patterns.RegisterLanguageSynthesizer("en", synth)
					patterns.RegisterLanguageSynthesizer(lang, synth)
				}
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
				// Java *PartialPosTagFilter uses Languages.get(lang).getTagger().
				if lt.TagWord != nil {
					switch strings.ToLower(base) {
					case "ru":
						rulesru.WireRussianFilterTaggerFromTagWord(lt.TagWord)
					case "fr":
						rulesfr.WireFrenchFilterTaggerFromTagWord(lt.TagWord)
					case "es":
						// FindSuggestionsFilter.Tag (Java SpanishTagger)
						ruleses.WireSpanishFilterTaggerFromTagWord(lt.TagWord)
					case "ca":
						// FindSuggestionsFilter.Tag (Java CatalanTagger)
						rulesca.WireCatalanFilterTaggerFromTagWord(lt.TagWord)
					case "ga":
						rulesga.WireIrishFilterTaggerFromTagWord(lt.TagWord)
					case "pt":
						rulespt.WirePortugueseFilterTaggerFromTagWord(lt.TagWord)
					}
				}
			}
			// Java createDefaultSynthesizer when *_synth.dict is present
			// (EN/DE/PL language-specific types; other langs BaseSynthesizer).
			if synthPath := DiscoverLanguageSynthDict(opts, base); synthPath != "" {
				if synth := OpenLanguageSynthesizer(base, synthPath); synth != nil {
					patterns.RegisterLanguageSynthesizer(base, synth)
					patterns.RegisterLanguageSynthesizer(lang, synth)
				}
			}
			// Java createDefaultDisambiguator(): FR/ES/PT hybrids when resources exist.
			_ = RegisterHybridDisambiguator(lt, base, opts)
			// German multitoken speller for MultitokenSpellerFilter (Java GermanMultitokenSpeller.INSTANCE).
			if strings.EqualFold(base, "de") {
				wireGermanMultitokenSpeller(opts)
			}
			// Dutch multitoken speller (Java DutchMultitokenSpeller.INSTANCE).
			if strings.EqualFold(base, "nl") {
				wireDutchMultitokenSpeller(opts)
			}
			// Portuguese multitoken speller (Java PortugueseMultitokenSpeller.INSTANCE;
			// MultitokenSpellerFilter shortCode gate includes "pt").
			if strings.EqualFold(base, "pt") {
				wirePortugueseMultitokenSpeller(opts)
			}
			// French/Spanish/Catalan MultitokenSpeller.INSTANCE (areTokensAcceptedBySpeller=false).
			if strings.EqualFold(base, "fr") {
				wireFrenchMultitokenSpeller(opts)
			}
			if strings.EqualFold(base, "es") {
				wireSpanishMultitokenSpeller(opts)
			}
			if strings.EqualFold(base, "ca") {
				wireCatalanMultitokenSpeller(opts)
			}
		}
		// Pattern rule files after multitoken speller so MultitokenSpellerFilter can use the dict.
		// Java Language.getRuleFileNames(): grammar, style, custom, then variant files.
		if languagetool.UseUpstreamGrammar() {
			for _, rpath := range DiscoverLanguagePatternRuleFiles(opts, lang) {
				_, _ = patterns.RegisterGrammarFile(lt, rpath, lang)
			}
		}
		// Java English.getRelevantRules: L2 grammar when mother tongue is de/fr (always, not gated).
		if strings.EqualFold(base, "en") && opts.MotherTongue != "" {
			if l2 := DiscoverEnglishL2GrammarXML(opts, opts.MotherTongue); l2 != "" {
				_, _ = patterns.RegisterGrammarFile(lt, l2, lang)
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

// wireGermanMultitokenSpeller ports GermanMultitokenSpeller resource load
// (multitoken-suggest.txt, spelling_global.txt, hunspell/spelling.txt).
func wireGermanMultitokenSpeller(opts *CommandLineOptions) {
	var paths []string
	if p := DiscoverGermanMultitokenSuggest(opts); p != "" {
		paths = append(paths, p)
	}
	if p := DiscoverSpellingGlobal(opts); p != "" {
		paths = append(paths, p)
	}
	// hunspell/spelling.txt beside multitoken-suggest when present
	if len(paths) > 0 {
		dir := filepath.Dir(paths[0])
		// multitoken-suggest is under de/; spelling.txt under de/hunspell/
		sp := filepath.Join(dir, "hunspell", "spelling.txt")
		if st, err := os.Stat(sp); err == nil && st.Mode().IsRegular() {
			paths = append(paths, sp)
		}
	}
	if len(paths) == 0 {
		sp := de.DiscoverAndLoadGermanMultitokenSpeller()
		if sp != nil && sp.MultitokenSpeller != nil {
			var isMiss func(string) bool
			if de.FilterDictAvailable() {
				isMiss = de.FilterDictIsMisspelled
			}
			patterns.SetDefaultMultitokenSpeller(sp.MultitokenSpeller, isMiss)
		}
		return
	}
	sp, err := de.LoadGermanMultitokenSpeller(paths...)
	if err != nil || sp == nil || sp.MultitokenSpeller == nil {
		return
	}
	var isMiss func(string) bool
	if de.FilterDictAvailable() {
		isMiss = de.FilterDictIsMisspelled
	}
	patterns.SetDefaultMultitokenSpeller(sp.MultitokenSpeller, isMiss)
}

// wireDutchMultitokenSpeller ports Dutch.getMultitokenSpeller resource load
// (/nl/multiwords.txt + /spelling_global.txt) into MultitokenSpellerFilter.
func wireDutchMultitokenSpeller(opts *CommandLineOptions) {
	mw := DiscoverLanguageMultiwords(opts, "nl")
	sg := DiscoverSpellingGlobal(opts)
	if mw == "" && sg == "" {
		// still try package discover (inspiration / testdata walk)
		sp := rulesnl.DiscoverAndLoadDutchMultitokenSpeller()
		if sp != nil && sp.MultitokenSpeller != nil {
			var isMiss func(string) bool
			if rulesnl.FilterDictAvailable() {
				isMiss = rulesnl.FilterDictIsMisspelled
			}
			patterns.SetDefaultMultitokenSpeller(sp.MultitokenSpeller, isMiss)
		}
		return
	}
	sp, err := rulesnl.LoadDutchMultitokenSpeller(mw, sg)
	if err != nil || sp == nil || sp.MultitokenSpeller == nil {
		return
	}
	var isMiss func(string) bool
	if rulesnl.FilterDictAvailable() {
		isMiss = rulesnl.FilterDictIsMisspelled
	}
	patterns.SetDefaultMultitokenSpeller(sp.MultitokenSpeller, isMiss)
}

// wirePortugueseMultitokenSpeller ports Portuguese.getMultitokenSpeller resource load
// (/pt/multiwords.txt + /spelling_global.txt + /pt/hyphenated_words.txt).
// Java MultitokenSpellerFilter shortCode gate includes "pt".
func wirePortugueseMultitokenSpeller(opts *CommandLineOptions) {
	// Wire default spelling dict for MultitokenSpellerFilter.isMisspelled when present.
	_ = rulespt.TryWirePortugueseFilterSpeller()
	// Discover covers multiwords + spelling_global + hyphenated (Java Arrays.asList order).
	// CLI data-dir paths are also found by spelling.Discover* walk.
	_ = opts
	sp := rulespt.DiscoverAndLoadPortugueseMultitokenSpeller()
	if sp == nil || sp.MultitokenSpeller == nil {
		return
	}
	var isMiss func(string) bool
	if rulespt.FilterDictAvailable() {
		isMiss = rulespt.FilterDictIsMisspelled
	}
	patterns.SetDefaultMultitokenSpellerWithOptions(sp.MultitokenSpeller, isMiss, true)
}

// wireFrenchMultitokenSpeller ports French.getMultitokenSpeller
// (/fr/multiwords.txt + /spelling_global.txt + /fr/hyphenated_words.txt).
// Java MultitokenSpellerFilter leaves areTokensAcceptedBySpeller=false for "fr",
// but MultitokenSpeller still uses getDefaultSpellingRule for discardRunOnWords.
func wireFrenchMultitokenSpeller(opts *CommandLineOptions) {
	_ = opts
	sp := rulesfr.DiscoverAndLoadFrenchMultitokenSpeller()
	if sp == nil || sp.MultitokenSpeller == nil {
		return
	}
	// checkSpelling=false: filter gate off; isMiss still feeds discardRunOnWords.
	var isMiss func(string) bool
	if rulesfr.FilterDictAvailable() {
		isMiss = rulesfr.FilterDictIsMisspelled
	}
	patterns.SetDefaultMultitokenSpellerWithOptions(sp.MultitokenSpeller, isMiss, false)
}

// wireSpanishMultitokenSpeller ports Spanish.getMultitokenSpeller
// (/es/multiwords.txt + /spelling_global.txt + /es/hyphenated_words.txt).
// Java MultitokenSpellerFilter leaves areTokensAcceptedBySpeller=false for "es",
// but MultitokenSpeller still uses getDefaultSpellingRule for discardRunOnWords.
func wireSpanishMultitokenSpeller(opts *CommandLineOptions) {
	_ = opts
	sp := ruleses.DiscoverAndLoadSpanishMultitokenSpeller()
	if sp == nil || sp.MultitokenSpeller == nil {
		return
	}
	var isMiss func(string) bool
	if ruleses.FilterDictAvailable() {
		isMiss = ruleses.FilterDictIsMisspelled
	}
	patterns.SetDefaultMultitokenSpellerWithOptions(sp.MultitokenSpeller, isMiss, false)
}

// wireCatalanMultitokenSpeller ports Catalan.getMultitokenSpeller
// (/ca/multiwords.txt + /spelling_global.txt + /ca/hyphenated_words.txt + Morfologik extra).
// Java MultitokenSpellerFilter leaves areTokensAcceptedBySpeller=false for "ca",
// but MultitokenSpeller still uses getDefaultSpellingRule for discardRunOnWords.
func wireCatalanMultitokenSpeller(opts *CommandLineOptions) {
	_ = opts
	sp := rulesca.DiscoverAndLoadCatalanMultitokenSpeller()
	if sp == nil || sp.MultitokenSpeller == nil {
		return
	}
	var isMiss func(string) bool
	if rulesca.FilterDictAvailable() {
		isMiss = rulesca.FilterDictIsMisspelled
	}
	patterns.SetDefaultMultitokenSpellerWithOptions(sp.MultitokenSpeller, isMiss, false)
}

// resolveFalseFriendsFile prefers LANG_FALSEFRIENDS_FILE, then data-dir official names.
// Java: /org/languagetool/rules/false-friends.xml — not soft invent files.
func resolveFalseFriendsFile(opts *CommandLineOptions) string {
	if p := os.Getenv("LANG_FALSEFRIENDS_FILE"); p != "" {
		return p
	}
	try := func(dir string) string {
		if dir == "" {
			return ""
		}
		for _, name := range []string{
			"false-friends-nodtd.xml",
			"false-friends.xml",
		} {
			p := filepath.Join(dir, name)
			if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
				return p
			}
		}
		return ""
	}
	if opts != nil && opts.GetDataDir() != "" {
		if p := try(opts.GetDataDir()); p != "" {
			return p
		}
	}
	if dir := os.Getenv("LANG_DATA_DIR"); dir != "" {
		if p := try(dir); p != "" {
			return p
		}
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
		// Java CommandLineTools.checkText: lt.getLanguage().getSentenceTokenizer().tokenize(contents).size()
		SentenceTokenize: func(s string) []string { return lt.SentenceTokenize(s) },
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
	// Prefer higher-priority / non-overlapping spans so grammar+speller on the same
	// token do not undo each other (e.g. EN_A_VS_AN vs MORFOLOGIK_RULE_EN_US).
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
			// Skip pure SENT_START only — last content word carries SENT_END in LT.
			if t == nil || t.IsSentenceStart() {
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

// CoreListRules writes registered rule IDs for lang (tab columns: id cat issue url kind state).
func CoreListRules(w io.Writer, lang string) error {
	return CoreListRulesOpts(w, &CommandLineOptions{Language: lang})
}

// CoreListRulesOpts is like CoreListRules but honors opts (e.g. Level=PICKY).
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
	sort.Strings(ids)
	offN := 0
	for _, id := range ids {
		cat, _, issue, _ := languagetool.RuleMeta(id)
		if cat == "" {
			cat = "MISC"
		}
		if issue == "" {
			issue = "uncategorized"
		}
		url := languagetool.RuleURL(id, lang)
		// kind is always "core" once invent soft packs are gone; keep column for tooling.
		kind := "core"
		state := "on"
		if _, ok := active[id]; !ok {
			state = "off"
			offN++
		}
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", id, cat, issue, url, kind, state); err != nil {
			return err
		}
	}
	parts := []string{
		fmt.Sprintf("total=%d", len(ids)),
		fmt.Sprintf("off=%d", offN),
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
		catID, catName, issue, short := languagetool.RuleMeta(id)
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
		ruleURL := languagetool.RuleURL(id, lang)
		if u := matchURL(m); u != "" {
			ruleURL = u
		}
		item := tools.MatchForJSON{
			Message:               m.GetMessage(),
			ShortMessage:          sm,
			FromPos:               m.GetFromPos(),
			ToPos:                 m.GetToPos(),
			SuggestedReplacements: m.GetSuggestedReplacements(),
			RuleID:                id,
			RuleDescription:       languagetool.RuleDescription(id),
			IssueType:             issue,
			CategoryID:            catID,
			CategoryName:          catName,
			Severity:              languagetool.SeverityFromIssueType(issue),
			RuleURL:               ruleURL,
			Tags:                  ruleTagsOf(m),
			TempOff:               ruleTempOffOf(m),
			IsPremium:             ruleIsPremiumOf(m),
			SubID:                 ruleSubIDOf(m),
			SourceFile:            ruleSourceFileOf(m),
		}
		mj = append(mj, item)
	}
	out, err := s.RuleMatchesToJSON(mj, contents, contextSize)
	if err != nil {
		return matchesToMinimalJSON(matches)
	}
	return out
}
