package server

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/corepack"
)

// Check runs a language-aware core rule pack on text (full XML grammar deferred).
// Honors pipeline disabled-rule IDs and optional overlap cleaning.
func (p *Pipeline) Check(text string) []languagetool.LocalMatch {
	if p == nil {
		return nil
	}
	lang := p.settings.LangCode
	if lang == "" {
		lang = "en"
	}
	lt := languagetool.NewJLanguageTool(lang)
	corepack.Register(lt, lang)

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

	matches := lt.Check(text)
	if p.cleanOverlaps {
		// assign soft priorities by rule family so layout doesn't stomp grammar injects
		for i := range matches {
			id := matches[i].RuleID
			if id == "EN_A_VS_AN" || strings.Contains(id, "WORD_REPEAT") ||
				strings.HasPrefix(id, "EN_") && strings.Contains(id, "_OF") {
				matches[i].Priority = 5
			} else if matches[i].Priority == 0 {
				matches[i].Priority = 1
			}
		}
		matches = languagetool.CleanOverlappingLocalMatches(matches)
	}
	return matches
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
