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
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// newConfiguredLT builds a language tool aligned with commandline.configureCoreLT
// (official resources only — no soft invent packs or soft false-friends paths).
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
	base := lang
	if i := strings.IndexByte(lang, '-'); i > 0 {
		base = lang[:i]
	}
	mt := tools.JavaStringTrim(p.settings.MotherTongueCode)

	// EN: official dicts/filters/multitoken/hybrid; demo only under LANG_DEMO_SPELLER.
	if strings.EqualFold(base, "en") {
		demoSpell := os.Getenv("LANG_DEMO_SPELLER") == "1"
		nearest := en.DemoEnglishKnownWords()
		spellOK := false
		ruleID, _ := en.EnglishVariantSpellerMeta(lang)
		if dictPath := commandline.DiscoverEnglishVariantDict(nil, lang); dictPath != "" {
			_ = en.WireEnglishFilterSpeller(dictPath)
			// Dict SuggestEdits only — no invent typo map.
			// Skip if core pack already registered this speller ID.
			for _, id := range lt.GetAllRegisteredRuleIDs() {
				if id == ruleID {
					spellOK = true
					break
				}
			}
			if !spellOK {
				spellOK = en.RegisterBinaryEnglishSpellerID(lt, dictPath, ruleID, nearest, nil)
			}
		}
		if !spellOK && demoSpell {
			en.RegisterDemoEnglishSpeller(lt, nearest, en.CommonDemoSpellerSuggestions)
		}
		taggerOK := false
		if posPath := commandline.DiscoverEnglishPOSDict(nil); posPath != "" {
			taggerOK = en.RegisterBinaryEnglishTagger(lt, posPath)
			_ = en.WireEnglishFilterTagger(posPath)
		}
		if !taggerOK && demoSpell {
			en.RegisterDemoEnglishTagger(lt)
		}
		// Java English.createDefaultSynthesizer for pattern match suggestions.
		if synthPath := commandline.DiscoverEnglishSynthDict(nil); synthPath != "" {
			if synth := commandline.OpenLanguageSynthesizer("en", synthPath); synth != nil {
				patterns.RegisterLanguageSynthesizer("en", synth)
				patterns.RegisterLanguageSynthesizer(lang, synth)
			}
		}
		en.RegisterEnglishChunker(lt)
		// Multitoken speller for MultitokenSpellerFilter (official multiwords + spelling_global).
		if mw, sg := commandline.DiscoverEnglishMultiwords(nil), commandline.DiscoverSpellingGlobal(nil); mw != "" || sg != "" {
			if sp, err := en.LoadEnglishMultitokenSpeller(mw, sg); err == nil && sp != nil {
				var isMiss func(string) bool
				if en.FilterDictAvailable() {
					isMiss = en.FilterDictIsMisspelled
				}
				patterns.SetDefaultMultitokenSpeller(sp.MultitokenSpeller, isMiss)
			}
		}
		_ = commandline.RegisterHybridDisambiguator(lt, base, nil)
	} else {
		if posPath := commandline.DiscoverLanguagePOSDict(nil, base); posPath != "" {
			_ = languagetool.RegisterBinaryPOSTagger(lt, posPath)
		}
		// Java createDefaultSynthesizer (DE/PL language-specific; others Base).
		if synthPath := commandline.DiscoverLanguageSynthDict(nil, base); synthPath != "" {
			if synth := commandline.OpenLanguageSynthesizer(base, synthPath); synth != nil {
				patterns.RegisterLanguageSynthesizer(base, synth)
				patterns.RegisterLanguageSynthesizer(lang, synth)
			}
		}
		_ = commandline.RegisterHybridDisambiguator(lt, base, nil)
	}

	// Official grammar/style/variant files (Java getRuleFileNames); default on.
	if languagetool.UseUpstreamGrammar() {
		for _, rpath := range commandline.DiscoverLanguagePatternRuleFiles(nil, lang) {
			_, _ = patterns.RegisterGrammarFile(lt, rpath, lang)
		}
	}
	// Java English L2 grammar when mother tongue is de/fr.
	if strings.EqualFold(base, "en") && mt != "" {
		if l2 := commandline.DiscoverEnglishL2GrammarXML(nil, mt); l2 != "" {
			_, _ = patterns.RegisterGrammarFile(lt, l2, lang)
		}
	}
	// Official false friends (no soft invent file).
	if mt != "" {
		if path := commandline.DiscoverFalseFriendsFile(nil); path != "" {
			_, _ = patterns.RegisterFalseFriendsFile(lt, path, lang, mt)
		}
	}

	// Query.LanguageCode may carry check mode (TEXTLEVEL_ONLY / ALL_BUT_TEXTLEVEL_ONLY)
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
	// Java: enable listed rule IDs only (no invent alias expansion).
	enabledExpanded := p.settings.Query.EnabledRules
	for _, id := range enabledExpanded {
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

// Check runs configured core (+ optional official grammar) rules on text.
// Honors pipeline disabled-rule IDs and optional overlap cleaning.
//
// Java server TextChecker uses new JLanguageTool(lang) (not MultiThreaded) —
// no invent multi-sentence period-count heuristic for threading.
// Overlap cleaning is Java CleanOverlappingFilter via lt.Check (PriorityForId).
func (p *Pipeline) Check(text string) []languagetool.LocalMatch {
	cr, _ := p.CheckWithResults(text)
	return languagetool.LocalMatchesFromCheckResults(cr)
}

// CheckWithResults ports Pipeline/JLanguageTool check2 surface for plain text:
// matches + ignored ranges from RuleMatch.getNewLanguageMatches (not invent
// foreign-script heuristics).
func (p *Pipeline) CheckWithResults(text string) (*languagetool.CheckResults, error) {
	if p == nil {
		return languagetool.NewCheckResults(nil, nil), nil
	}
	lt := p.newConfiguredLT()
	// Java JLanguageTool.setLevel: DEFAULT filters Tag.picky (false friends, …).
	if strings.EqualFold(string(p.settings.Level), string(CheckLevelPicky)) {
		lt.Level = languagetool.LevelPicky
	}
	// Pipeline cleanOverlaps=false → JLanguageTool.setCleanOverlappingMatches(false).
	if !p.cleanOverlaps {
		lt.DisableCleanOverlapping()
	}
	if p.maxErrRate > 0 {
		lt.MaxErrorsPerWordRate = p.maxErrRate
	}
	return lt.CheckWithResults(text)
}

// CheckAnnotated runs Check on annotated plain text and projects offsets onto the original markup.
func (p *Pipeline) CheckAnnotated(at *markup.AnnotatedText) []languagetool.LocalMatch {
	if p == nil || at == nil {
		return nil
	}
	lt := p.newConfiguredLT()
	if strings.EqualFold(string(p.settings.Level), string(CheckLevelPicky)) {
		lt.Level = languagetool.LevelPicky
	}
	if p != nil && !p.cleanOverlaps {
		lt.DisableCleanOverlapping()
	}
	matches := lt.CheckAnnotated(at)
	return languagetool.ProjectMatchesToOriginal(at, matches)
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
