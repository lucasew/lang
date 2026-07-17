package server

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/de"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/en"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/es"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/fr"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/nl"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/pl"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/uk"
)

// registerPipelineCore installs language-specific core packs when available.
func registerPipelineCore(lt *languagetool.JLanguageTool, lang string) {
	base := lang
	if i := strings.IndexByte(lang, '-'); i > 0 {
		base = lang[:i]
	}
	switch strings.ToLower(base) {
	case "en":
		en.RegisterCoreEnglishLanguageRules(lt)
	case "de":
		de.RegisterCoreGermanRules(lt)
	case "fr":
		fr.RegisterCoreFrenchRules(lt)
	case "es":
		es.RegisterCoreSpanishRules(lt)
	case "nl":
		nl.RegisterCoreDutchRules(lt)
	case "pl":
		pl.RegisterCorePolishRules(lt)
	case "uk":
		uk.RegisterCoreUkrainianRules(lt)
	default:
		rules.RegisterCoreRules(lt, lang)
	}
}

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
	registerPipelineCore(lt, lang)

	// apply pipeline disabled rules
	for id := range p.disabledRules {
		lt.DisableRule(id)
	}
	// query disabled
	for _, id := range p.settings.Query.DisabledRules {
		lt.DisableRule(id)
	}

	matches := lt.Check(text)
	if p.cleanOverlaps {
		// assign soft priorities by rule family so layout doesn't stomp grammar injects
		for i := range matches {
			switch matches[i].RuleID {
			case "EN_A_VS_AN", "WORD_REPEAT_RULE", "GERMAN_WORD_REPEAT_RULE",
				"FR_WORD_REPEAT_RULE", "SPANISH_WORD_REPEAT_RULE", "NL_WORD_REPEAT_RULE",
				"PL_WORD_REPEAT", "UKRAINIAN_WORD_REPEAT_RULE":
				matches[i].Priority = 5
			default:
				if matches[i].Priority == 0 {
					matches[i].Priority = 1
				}
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
