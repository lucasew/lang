package server

import (
	"os"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/commandline"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/markup"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/corepack"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/en"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// newConfiguredLT builds a language tool with core packs and pipeline filters applied.
func (p *Pipeline) newConfiguredLT() *languagetool.JLanguageTool {
	if p == nil {
		return languagetool.NewJLanguageTool("en")
	}
	lang := p.settings.LangCode
	if lang == "" {
		lang = "en"
	}
	lt := languagetool.NewJLanguageTool(lang)
	corepack.Register(lt, lang)
	if dir := softGrammarDirFromEnv(); dir != "" {
		_, _ = patterns.RegisterSoftGrammarDir(lt, dir, lang)
	}
	// soft false friends when mother tongue is set
	if mt := strings.TrimSpace(p.settings.MotherTongueCode); mt != "" {
		if path := softFalseFriendsPath(); path != "" {
			_, _ = patterns.RegisterFalseFriendsFile(lt, path, lang, mt)
		}
	}
	// EN speller: prefer CFSA2 en_US.dict; optional map demo fallback
	base := lang
	if i := strings.IndexByte(lang, '-'); i > 0 {
		base = lang[:i]
	}
	if strings.EqualFold(base, "en") {
		demoSpell := os.Getenv("LANG_DEMO_SPELLER") == "1"
		nearest := en.DemoEnglishKnownWords()
		sugs := en.CommonDemoSpellerSuggestions
		if typoPath := softEnglishTyposPath(); typoPath != "" {
			if extra, err := en.LoadSoftTyposFile(typoPath); err == nil && len(extra) > 0 {
				sugs = en.MergeSpellerSuggestions(sugs, extra)
			}
		}
		spellOK := false
		if dictPath := softEnglishUSDictPath(); dictPath != "" {
			spellOK = en.RegisterBinaryEnglishSpeller(lt, dictPath, nearest, sugs)
		}
		if !spellOK && demoSpell {
			en.RegisterDemoEnglishSpeller(lt, nearest, sugs)
		}
		taggerOK := false
		if posPath := softEnglishPOSDictPath(); posPath != "" {
			taggerOK = en.RegisterBinaryEnglishTagger(lt, posPath)
		}
		if !taggerOK && demoSpell {
			en.RegisterDemoEnglishTagger(lt)
		}
		en.RegisterSoftEnglishDisambiguator(lt, softEnglishMultiwordsPath(), softEnglishDisambigXMLPath(), softEnglishIgnoreSpellingPath())
	} else {
		// Java createDefaultTagger + createDefaultDisambiguator for non-EN
		// (same soft wiring as commandline.configureCoreLT).
		if posPath := commandline.DiscoverLanguagePOSDict(nil, base); posPath != "" {
			_ = languagetool.RegisterBinaryPOSTagger(lt, posPath)
		}
		softPaths := commandline.SoftHybridPaths{
			Multiwords:      commandline.DiscoverLanguageMultiwords(nil, base),
			SoftDisambigXML: commandline.DiscoverLanguageSoftDisambiguationXML(nil, base),
		}
		if strings.EqualFold(base, "de") {
			// Multitoken lists are large; only attach when paths resolve (cached).
			softPaths.DEMultitokenIgnore = commandline.DiscoverGermanMultitokenIgnore(nil)
			softPaths.DEMultitokenSuggest = commandline.DiscoverGermanMultitokenSuggest(nil)
		}
		_ = commandline.RegisterSoftHybridDisambiguator(lt, base, softPaths)
	}

	// soft: Query.LanguageCode may carry check mode (TEXTLEVEL_ONLY / ALL_BUT_TEXTLEVEL_ONLY)
	switch strings.ToUpper(p.settings.Query.LanguageCode) {
	case "TEXTLEVEL_ONLY", "TEXTLEVELONLY":
		lt.SetMode(languagetool.ModeTextLevelOnly)
	case "ALL_BUT_TEXTLEVEL_ONLY", "ALLBUTTEXTLEVELONLY":
		lt.SetMode(languagetool.ModeAllButTextLevel)
	}

	// apply pipeline disabled rules
	for id := range p.disabledRules {
		lt.DisableRule(id)
	}
	// query disabled
	for _, id := range p.settings.Query.DisabledRules {
		lt.DisableRule(id)
	}
	// soft: expand SOFT_OPTIONAL / SOFT_OPT_ALL → all SOFT_OPT_* rules
	enabledExpanded := languagetool.ExpandSoftEnableRuleIDs(lt.GetAllRegisteredRuleIDs(), p.settings.Query.EnabledRules)
	// soft: re-enable optional packs when listed (not only under enabled-only)
	for _, id := range enabledExpanded {
		// skip aliases already expanded out of the slice
		if id != "" {
			lt.EnableRule(id)
		}
	}
	// query enabled-only: disable every registered rule not listed
	if p.settings.Query.UseEnabledOnly {
		enabled := map[string]struct{}{}
		for _, id := range enabledExpanded {
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
	}
	return lt
}

// softGrammarDirFromEnv resolves LANG_GRAMMAR_DIR, LANG_DATA_DIR/grammar, or walk-up testdata/grammar.
func softGrammarDirFromEnv() string {
	if dir := os.Getenv("LANG_GRAMMAR_DIR"); dir != "" {
		return dir
	}
	if dir := os.Getenv("LANG_DATA_DIR"); dir != "" {
		return dir + "/grammar"
	}
	return walkUpFind("testdata/grammar")
}

// softEnglishUSDictPath resolves LANG_EN_US_DICT or walk-up third_party en_US.dict.
func softEnglishUSDictPath() string {
	if p := os.Getenv("LANG_EN_US_DICT"); p != "" {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	if dir := os.Getenv("LANG_DATA_DIR"); dir != "" {
		for _, rel := range []string{
			dir + "/en/hunspell/en_US.dict",
			dir + "/en_US.dict",
		} {
			if _, err := os.Stat(rel); err == nil {
				return rel
			}
		}
	}
	if p := walkUpFind("third_party/english-pos-dict/org/languagetool/resource/en/hunspell/en_US.dict"); p != "" {
		return p
	}
	return walkUpFind("inspiration/languagetool/languagetool-language-modules/en/src/main/resources/org/languagetool/resource/en/hunspell/en_US.dict")
}

// softEnglishDisambigXMLPath resolves LANG_DISAMBIGUATION_FILE or walk-up en-soft.xml.
func softEnglishDisambigXMLPath() string {
	if p := os.Getenv("LANG_DISAMBIGUATION_FILE"); p != "" {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return walkUpFind("testdata/disambiguation/en-soft.xml")
}

// softEnglishIgnoreSpellingPath resolves LANG_IGNORE_SPELLING_FILE or walk-up word list.
func softEnglishIgnoreSpellingPath() string {
	if p := os.Getenv("LANG_IGNORE_SPELLING_FILE"); p != "" {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return walkUpFind("testdata/disambiguation/en-ignore-spelling.txt")
}

// softEnglishTyposPath resolves LANG_EN_TYPOS_FILE or walk-up en-typos.tsv.
func softEnglishTyposPath() string {
	if p := os.Getenv("LANG_EN_TYPOS_FILE"); p != "" {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return walkUpFind("testdata/spelling/en-typos.tsv")
}

// softEnglishMultiwordsPath resolves LANG_EN_MULTIWORDS or walk-up multiwords.txt.
func softEnglishMultiwordsPath() string {
	if p := os.Getenv("LANG_EN_MULTIWORDS"); p != "" {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	if p := walkUpFind("inspiration/languagetool/languagetool-language-modules/en/src/main/resources/org/languagetool/resource/en/multiwords.txt"); p != "" {
		return p
	}
	return walkUpFind("third_party/english-pos-dict/org/languagetool/resource/en/multiwords.txt")
}

// softEnglishPOSDictPath resolves LANG_ENGLISH_DICT or walk-up third_party english.dict.
func softEnglishPOSDictPath() string {
	if p := os.Getenv("LANG_ENGLISH_DICT"); p != "" {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	if dir := os.Getenv("LANG_DATA_DIR"); dir != "" {
		for _, rel := range []string{
			dir + "/en/english.dict",
			dir + "/english.dict",
		} {
			if _, err := os.Stat(rel); err == nil {
				return rel
			}
		}
	}
	if p := walkUpFind("third_party/english-pos-dict/org/languagetool/resource/en/english.dict"); p != "" {
		return p
	}
	return walkUpFind("inspiration/languagetool/languagetool-language-modules/en/src/main/resources/org/languagetool/resource/en/english.dict")
}

// softFalseFriendsPath resolves LANG_FALSEFRIENDS_FILE or a well-known testdata path.
func softFalseFriendsPath() string {
	if p := os.Getenv("LANG_FALSEFRIENDS_FILE"); p != "" {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	if dir := os.Getenv("LANG_DATA_DIR"); dir != "" {
		p := dir + "/false-friends-soft.xml"
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return walkUpFind("testdata/false-friends-soft.xml")
}

func walkUpFind(rel string) string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for i := 0; i < 12; i++ {
		cand := dir + "/" + rel
		if _, err := os.Stat(cand); err == nil {
			return cand
		}
		parent := dir
		// trim last segment
		for j := len(dir) - 1; j >= 0; j-- {
			if dir[j] == '/' {
				parent = dir[:j]
				if parent == "" {
					parent = "/"
				}
				break
			}
		}
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

func (p *Pipeline) cleanMatches(matches []languagetool.LocalMatch) []languagetool.LocalMatch {
	if p == nil || !p.cleanOverlaps {
		return matches
	}
	for i := range matches {
		id := matches[i].RuleID
		if id == "EN_A_VS_AN" || strings.Contains(id, "WORD_REPEAT") ||
			strings.HasPrefix(id, "EN_") && strings.Contains(id, "_OF") {
			matches[i].Priority = 5
		} else if matches[i].Priority == 0 {
			matches[i].Priority = 1
		}
	}
	return languagetool.CleanOverlappingLocalMatches(matches)
}

// Check runs a language-aware core rule pack on text (full XML grammar deferred).
// Honors pipeline disabled-rule IDs and optional overlap cleaning.
// Uses multi-threaded Check for multi-sentence texts (pool size = GOMAXPROCS soft).
func (p *Pipeline) Check(text string) []languagetool.LocalMatch {
	if p == nil {
		return nil
	}
	lt := p.newConfiguredLT()
	// Heuristic multi-sentence detection avoids a double full Analyze before Check.
	var matches []languagetool.LocalMatch
	if multiSentenceHeuristic(text) {
		mtl := languagetool.NewMultiThreadedJLanguageTool(lt.GetLanguageCode(), 0)
		mtl.JLanguageTool = lt
		matches = mtl.Check(text)
	} else {
		matches = lt.Check(text)
	}
	return p.cleanMatches(matches)
}

// multiSentenceHeuristic reports likely multi-sentence input (terminators + space/capital).
func multiSentenceHeuristic(text string) bool {
	n := 0
	for i := 0; i < len(text); i++ {
		c := text[i]
		if c == '.' || c == '!' || c == '?' {
			// count only if not last char and something follows
			if i+1 < len(text) {
				n++
				if n >= 2 {
					return true
				}
			}
		}
	}
	return false
}

// CheckAnnotated runs Check on annotated plain text and projects offsets onto the original markup.
func (p *Pipeline) CheckAnnotated(at *markup.AnnotatedText) []languagetool.LocalMatch {
	if p == nil || at == nil {
		return nil
	}
	lt := p.newConfiguredLT()
	matches := lt.CheckAnnotated(at)
	matches = languagetool.ProjectMatchesToOriginal(at, matches)
	return p.cleanMatches(matches)
}

// DisableRuleID records a rule to skip (before SetupFinished).
func (p *Pipeline) DisableRuleID(id string) error {
	if err := p.preventModification(); err != nil {
		return err
	}
	if p.disabledRules == nil {
		p.disabledRules = map[string]struct{}{}
	}
	p.disabledRules[id] = struct{}{}
	return nil
}
