package server

import (
	"os"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
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
	if dir := os.Getenv("LANG_GRAMMAR_DIR"); dir != "" {
		_, _ = patterns.RegisterSoftGrammarDir(lt, dir, lang)
	}
	// optional demo EN speller/tagger injects for local smoke servers
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
	// query enabled-only: disable every registered rule not listed
	if p.settings.Query.UseEnabledOnly {
		enabled := map[string]struct{}{}
		for _, id := range p.settings.Query.EnabledRules {
			if id != "" {
				enabled[id] = struct{}{}
			}
		}
		for _, id := range lt.GetAllRegisteredRuleIDs() {
			if _, ok := enabled[id]; !ok {
				lt.DisableRule(id)
			}
		}
	}
	return lt
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
