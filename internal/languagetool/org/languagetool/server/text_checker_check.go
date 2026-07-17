package server

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/markup"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/en"
)

// CheckOptions carries optional check-query knobs beyond language/disabled rules.
type CheckOptions struct {
	Disabled       []string
	Enabled        []string
	UseEnabledOnly bool
	Mode           CheckMode
	Level          CheckLevel
	// IgnoreWords soft user-dictionary surfaces (suppresses spelling matches).
	IgnoreWords []string
	// Category filters (SoftRuleMeta / LocalMatch category IDs).
	DisabledCategories []string
	EnabledCategories  []string
}

// Check runs core rules for language on text and returns RemoteRuleMatch results.
// Uses Pipeline so disabled/enabled-only rules from query params can be applied.
func (t *TextChecker) Check(text, lang string, disabled []string) []RemoteRuleMatch {
	return t.CheckWithOptions(text, lang, CheckOptions{Disabled: disabled})
}

// pipelineSettingsFor builds pool key settings for a check.
func pipelineSettingsFor(lang string, opts CheckOptions) PipelineSettings {
	if lang == "" {
		lang = "en"
	}
	settings := NewPipelineSettings(lang, "check")
	settings.Query.DisabledRules = append([]string(nil), opts.Disabled...)
	settings.Query.EnabledRules = append([]string(nil), opts.Enabled...)
	settings.Query.UseEnabledOnly = opts.UseEnabledOnly
	settings.Query.UseQuerySettings = len(opts.Disabled) > 0 || len(opts.Enabled) > 0 || opts.UseEnabledOnly
	// include filters in pool key (Key() does not hash full rule lists)
	var keyParts []string
	if opts.Level != "" {
		keyParts = append(keyParts, "level:"+string(opts.Level))
	}
	if opts.Mode != "" {
		// soft: Query.LanguageCode carries check mode for Pipeline.Check
		settings.Query.LanguageCode = string(opts.Mode)
		keyParts = append(keyParts, "mode:"+string(opts.Mode))
	}
	if opts.UseEnabledOnly {
		keyParts = append(keyParts, "eo:"+strings.Join(opts.Enabled, ","))
	} else if len(opts.Enabled) > 0 {
		keyParts = append(keyParts, "en:"+strings.Join(opts.Enabled, ","))
	}
	if len(opts.Disabled) > 0 {
		keyParts = append(keyParts, "dis:"+strings.Join(opts.Disabled, ","))
	}
	if len(opts.DisabledCategories) > 0 {
		keyParts = append(keyParts, "dcat:"+strings.Join(opts.DisabledCategories, ","))
	}
	if len(opts.EnabledCategories) > 0 {
		keyParts = append(keyParts, "ecat:"+strings.Join(opts.EnabledCategories, ","))
	}
	settings.GlobalConfigKey = strings.Join(keyParts, "|")
	return settings
}

// preparePipeline builds a pipeline for check options (from pool when available).
// Caller must call releasePipeline when done if pool was used.
func (t *TextChecker) preparePipeline(lang string, opts CheckOptions) (pl *Pipeline, settings PipelineSettings, fromPool bool) {
	settings = pipelineSettingsFor(lang, opts)
	if t != nil && t.Pool != nil {
		borrowed, err := t.Pool.Borrow(settings)
		if err == nil && borrowed != nil {
			// Query disabled/enabled filters are reapplied inside Pipeline.Check
			if !borrowed.IsFrozen() {
				_ = borrowed.SetCleanOverlappingMatches(true)
				borrowed.SetupFinished()
			}
			return borrowed, settings, true
		}
	}
	p := NewPipeline(settings)
	for _, id := range opts.Disabled {
		id = strings.TrimSpace(id)
		if id != "" {
			_ = p.DisableRuleID(id)
		}
	}
	_ = p.SetCleanOverlappingMatches(true)
	p.SetupFinished()
	return p, settings, false
}

func (t *TextChecker) releasePipeline(settings PipelineSettings, pl *Pipeline, fromPool bool) {
	if fromPool && t != nil && t.Pool != nil {
		t.Pool.Return(settings, pl)
	}
}

// CheckWithOptions is Check with enabled-only, mode, and level support.
func (t *TextChecker) CheckWithOptions(text, lang string, opts CheckOptions) []RemoteRuleMatch {
	p, settings, fromPool := t.preparePipeline(lang, opts)
	defer t.releasePipeline(settings, p, fromPool)
	locals := p.Check(text)
	locals = applyLevelPickyBoost(lang, opts.Level, locals, text)
	locals = filterLocalsByIgnoreWords(text, locals, opts.IgnoreWords)
	locals = filterLocalsByCategories(locals, opts)
	ctxSize := DefaultContextSize
	if t != nil && t.ContextSize > 0 {
		ctxSize = t.ContextSize
	}
	return LocalMatchesToRemote(text, locals, ctxSize)
}

// CheckAnnotatedWithOptions checks annotated markup text; match offsets are in original markup space.
func (t *TextChecker) CheckAnnotatedWithOptions(at *markup.AnnotatedText, lang string, opts CheckOptions) []RemoteRuleMatch {
	if at == nil {
		return nil
	}
	p, settings, fromPool := t.preparePipeline(lang, opts)
	defer t.releasePipeline(settings, p, fromPool)
	locals := p.CheckAnnotated(at)
	plain := at.GetPlainText()
	locals = applyLevelPickyBoost(lang, opts.Level, locals, plain)
	locals = filterLocalsByIgnoreWords(plain, locals, opts.IgnoreWords)
	locals = filterLocalsByCategories(locals, opts)
	// Context uses original markup string so projected offsets align.
	orig := at.GetTextWithMarkup()
	ctxSize := DefaultContextSize
	if t != nil && t.ContextSize > 0 {
		ctxSize = t.ContextSize
	}
	return LocalMatchesToRemote(orig, locals, ctxSize)
}

func filterLocalsByIgnoreWords(text string, ms []languagetool.LocalMatch, ignore []string) []languagetool.LocalMatch {
	return languagetool.FilterMatchesByIgnoreWords(text, ms, ignore)
}

// filterLocalsByCategories applies disabledCategories / enabledCategories (with enabledOnly).
func filterLocalsByCategories(ms []languagetool.LocalMatch, opts CheckOptions) []languagetool.LocalMatch {
	if len(ms) == 0 {
		return ms
	}
	dis := stringSetFold(opts.DisabledCategories)
	en := stringSetFold(opts.EnabledCategories)
	if len(dis) == 0 && !(opts.UseEnabledOnly && len(en) > 0) {
		return ms
	}
	out := make([]languagetool.LocalMatch, 0, len(ms))
	for _, m := range ms {
		catID := m.CategoryID
		if catID == "" {
			catID, _, _, _ = SoftRuleMeta(m.RuleID)
		}
		catKey := strings.ToUpper(catID)
		if _, drop := dis[catKey]; drop {
			continue
		}
		if opts.UseEnabledOnly && len(en) > 0 {
			if _, ok := en[catKey]; !ok {
				continue
			}
		}
		out = append(out, m)
	}
	return out
}

func stringSetFold(items []string) map[string]struct{} {
	if len(items) == 0 {
		return nil
	}
	m := make(map[string]struct{}, len(items))
	for _, s := range items {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		m[strings.ToUpper(s)] = struct{}{}
	}
	return m
}

// applyLevelPickyBoost runs extra EN picky patterns when level is PICKY (soft).
func applyLevelPickyBoost(lang string, level CheckLevel, existing []languagetool.LocalMatch, text string) []languagetool.LocalMatch {
	if !strings.EqualFold(string(level), string(CheckLevelPicky)) {
		return existing
	}
	base := lang
	if i := strings.IndexByte(lang, '-'); i > 0 {
		base = lang[:i]
	}
	if !strings.EqualFold(base, "en") {
		return existing
	}
	lt := languagetool.NewJLanguageTool(lang)
	en.RegisterCoreEnglishLanguageRules(lt)
	en.RegisterPickyEnglishRules(lt)
	// only keep picky-only rule ids from this pass
	picky := map[string]struct{}{
		"EN_A_LOT": {}, "EN_IRREGARDLESS": {},
	}
	for _, m := range lt.Check(text) {
		if _, ok := picky[m.RuleID]; ok {
			existing = append(existing, m)
		}
	}
	return existing
}

// CheckAndBuildJSON is a V2 convenience: Check + BuildResponse.
func (v *V2TextChecker) CheckAndBuildJSON(text, langCode, langName string, disabled []string) (string, error) {
	return v.CheckAndBuildJSONWithOptions(text, langCode, langName, CheckOptions{Disabled: disabled})
}

// CheckAndBuildJSONWithOptions builds JSON with full check options.
func (v *V2TextChecker) CheckAndBuildJSONWithOptions(text, langCode, langName string, opts CheckOptions) (string, error) {
	if v == nil {
		return "", nil
	}
	if langName == "" {
		langName = langCode
	}
	matches := v.CheckWithOptions(text, langCode, opts)
	return v.BuildResponse(text, langCode, langName, matches)
}

// LocalMatchesToRemote maps cycle-free LocalMatch to API RemoteRuleMatch with context.
func LocalMatchesToRemote(text string, matches []languagetool.LocalMatch, contextSize int) []RemoteRuleMatch {
	if len(matches) == 0 {
		return nil
	}
	if contextSize <= 0 {
		contextSize = DefaultContextSize
	}
	out := make([]RemoteRuleMatch, 0, len(matches))
	for _, m := range matches {
		from, to := m.FromPos, m.ToPos
		if from < 0 {
			from = 0
		}
		if to < from {
			to = from
		}
		if to > len(text) {
			to = len(text)
		}
		start := from - contextSize
		if start < 0 {
			start = 0
		}
		end := to + contextSize
		if end > len(text) {
			end = len(text)
		}
		ctx := text[start:end]
		if ctx == "" {
			ctx = " " // NewRemoteRuleMatch requires non-empty context
		}
		msg := m.Message
		if msg == "" {
			msg = m.RuleID
		}
		if msg == "" {
			msg = "match"
		}
		ruleID := m.RuleID
		if ruleID == "" {
			ruleID = "UNKNOWN_RULE"
		}
		rm := NewRemoteRuleMatch(ruleID, msg, ctx, from-start, from, to-from)
		rm.ShortMessage = m.ShortMessage
		catID, catName, issue, short := SoftRuleMeta(ruleID)
		// Prefer metadata carried on LocalMatch (soft grammar XML categories).
		if m.CategoryID != "" {
			catID = m.CategoryID
		}
		if m.CategoryName != "" {
			catName = m.CategoryName
		}
		if m.IssueType != "" {
			issue = m.IssueType
		}
		rm.CategoryID = catID
		rm.Category = catName
		rm.LocQualityIssueType = issue
		if m.Description != "" {
			rm.Description = m.Description
		} else {
			rm.Description = SoftRuleDescription(ruleID)
		}
		if rm.ShortMessage == "" {
			rm.ShortMessage = short
		}
		rm.Replacements = append([]string(nil), m.Suggestions...)
		out = append(out, *rm)
	}
	return out
}
