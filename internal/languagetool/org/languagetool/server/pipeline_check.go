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
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	ensynth "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/en"
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
	mt := strings.TrimSpace(p.settings.MotherTongueCode)

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
			if synth := ensynth.OpenEnglishSynthesizerFromDictPath(synthPath); synth != nil {
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
		if synthPath := commandline.DiscoverLanguageSynthDict(nil, base); synthPath != "" {
			if synth := synthesis.OpenBaseSynthesizerFromDictPath(base, synthPath); synth != nil {
				patterns.RegisterLanguageSynthesizer(base, synth)
				patterns.RegisterLanguageSynthesizer(lang, synth)
			}
		}
		_ = commandline.RegisterHybridDisambiguator(lt, base, nil)
	}

	// Official grammar/style/variant files (Java getRuleFileNames) when enabled.
	if os.Getenv("LANG_USE_UPSTREAM_GRAMMAR") == "1" {
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

func (p *Pipeline) cleanMatches(matches []languagetool.LocalMatch) []languagetool.LocalMatch {
	if p == nil || !p.cleanOverlaps {
		return matches
	}
	// Prefer higher priority for known core grammar IDs over speller on the same span
	// (Java CleanOverlappingFilter / priority tables — incomplete subset, not invent IDs).
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

// Check runs configured core (+ optional official grammar) rules on text.
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
