package server

import (
	"strconv"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/markup"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/en"
)

// CheckOptions carries optional check-query knobs beyond language/disabled rules.
type CheckOptions struct {
	Disabled       []string
	Enabled        []string
	UseEnabledOnly bool
	Mode           CheckMode
	Level          CheckLevel
	// MotherTongue enables false-friend rules when official false-friends.xml is available.
	MotherTongue string
	// IgnoreWords user-dictionary surfaces (suppresses spelling matches).
	IgnoreWords []string
	// Category filters (rule meta / LocalMatch category IDs).
	DisabledCategories []string
	EnabledCategories  []string
	// RuleValues configurable rule options (e.g. "TOO_LONG_SENTENCE:10").
	RuleValues []string
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
	if opts.MotherTongue != "" {
		settings.MotherTongueCode = opts.MotherTongue
	}
	if opts.Level != "" {
		settings.Level = opts.Level
	}
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
		// Query.LanguageCode carries check mode for Pipeline.Check (TEXTLEVEL_ONLY, …).
		settings.Query.LanguageCode = string(opts.Mode)
		keyParts = append(keyParts, "mode:"+string(opts.Mode))
	}
	if opts.MotherTongue != "" {
		keyParts = append(keyParts, "mt:"+opts.MotherTongue)
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
	if len(opts.RuleValues) > 0 {
		keyParts = append(keyParts, "rv:"+strings.Join(opts.RuleValues, ","))
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
	locals = applyRuleValues(lang, text, locals, opts.RuleValues)
	locals = filterLocalsByIgnoreWords(text, locals, opts.IgnoreWords)
	locals = filterLocalsByCategories(locals, opts)
	ctxSize := DefaultContextSize
	if t != nil && t.ContextSize > 0 {
		ctxSize = t.ContextSize
	}
	return LocalMatchesToRemote(text, locals, ctxSize, lang)
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
	locals = applyRuleValues(lang, plain, locals, opts.RuleValues)
	locals = filterLocalsByIgnoreWords(plain, locals, opts.IgnoreWords)
	locals = filterLocalsByCategories(locals, opts)
	// Context uses original markup string so projected offsets align.
	orig := at.GetTextWithMarkup()
	ctxSize := DefaultContextSize
	if t != nil && t.ContextSize > 0 {
		ctxSize = t.ContextSize
	}
	return LocalMatchesToRemote(orig, locals, ctxSize, lang)
}

func filterLocalsByIgnoreWords(text string, ms []languagetool.LocalMatch, ignore []string) []languagetool.LocalMatch {
	return languagetool.FilterMatchesByIgnoreWords(text, ms, ignore)
}

// filterLocalsByCategories applies disabledCategories / enabledCategories (with enabledOnly).
func filterLocalsByCategories(ms []languagetool.LocalMatch, opts CheckOptions) []languagetool.LocalMatch {
	return languagetool.FilterMatchesByCategories(ms, opts.DisabledCategories, opts.EnabledCategories, opts.UseEnabledOnly)
}

// ParseRuleValues parses "RULE_ID:value,OTHER:2" into a map (soft; last wins).
func ParseRuleValues(items []string) map[string]string {
	if len(items) == 0 {
		return nil
	}
	out := map[string]string{}
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		// also allow comma-joined blob
		for _, part := range strings.Split(item, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			i := strings.IndexByte(part, ':')
			if i <= 0 || i == len(part)-1 {
				continue
			}
			id := strings.TrimSpace(part[:i])
			val := strings.TrimSpace(part[i+1:])
			if id != "" && val != "" {
				out[strings.ToUpper(id)] = val
			}
		}
	}
	return out
}

// applyRuleValues re-runs long-sentence detection with a custom max when configured.
func applyRuleValues(lang, text string, existing []languagetool.LocalMatch, raw []string) []languagetool.LocalMatch {
	vals := ParseRuleValues(raw)
	if len(vals) == 0 {
		return existing
	}
	maxStr, ok := vals["TOO_LONG_SENTENCE"]
	if !ok {
		maxStr, ok = vals["LONG_SENTENCE_RULE"]
	}
	if !ok {
		return existing
	}
	maxWords, err := strconv.Atoi(maxStr)
	if err != nil || maxWords <= 0 {
		return existing
	}
	// drop existing long-sentence matches
	out := make([]languagetool.LocalMatch, 0, len(existing))
	for _, m := range existing {
		id := strings.ToUpper(m.RuleID)
		if strings.Contains(id, "LONG_SENTENCE") || strings.Contains(id, "TOO_LONG_SENTENCE") {
			continue
		}
		out = append(out, m)
	}
	// soft re-check with custom threshold
	ls := rules.NewLongSentenceRule(map[string]string{
		"long_sentence_rule_msg2": "This sentence is too long (%d words)",
	}, maxWords)
	lt := languagetool.NewJLanguageTool(lang)
	lt.AddTextLevelRuleChecker(ls.GetID(), rules.AsTextLevelChecker(ls.MatchList))
	// Disable all other rules so only long-sentence fires
	for _, id := range lt.GetAllRegisteredRuleIDs() {
		if id != ls.GetID() {
			lt.DisableRule(id)
		}
	}
	for _, m := range lt.Check(text) {
		if m.RuleID == ls.GetID() || strings.Contains(strings.ToUpper(m.RuleID), "LONG_SENTENCE") || strings.Contains(strings.ToUpper(m.RuleID), "TOO_LONG") {
			m.CategoryID = "STYLE"
			m.CategoryName = "Style"
			m.IssueType = "style"
			if m.ShortMessage == "" {
				m.ShortMessage = "Long sentence"
			}
			out = append(out, m)
		}
	}
	return out
}

// applyLevelPickyBoost runs official Tag.picky rules when level is PICKY.
// Java: English.getRelevantRules Tag.picky (e.g. SimpleReplaceProfanityRule).
// Unit conversion is locale default (RegisterEnglishVariantExtraRules), not picky.
// Soft invent packs / picky-soft.xml are not loaded (faithful-port policy).
func applyLevelPickyBoost(lang string, level CheckLevel, existing []languagetool.LocalMatch, text string) []languagetool.LocalMatch {
	if !strings.EqualFold(string(level), string(CheckLevelPicky)) {
		return existing
	}
	base := lang
	if i := strings.IndexByte(lang, '-'); i > 0 {
		base = lang[:i]
	}
	if !strings.EqualFold(base, "en") {
		// Non-EN picky grammar.xml packs deferred until full official load is wired.
		return existing
	}
	lt := languagetool.NewJLanguageTool(lang)
	en.RegisterPickyEnglishRules(lt)
	// Official picky rule IDs from RegisterPickyEnglishRules (not invent sequences).
	pickyIDs := map[string]struct{}{}
	for _, id := range lt.GetAllRegisteredRuleIDs() {
		pickyIDs[id] = struct{}{}
	}
	seen := map[string]struct{}{}
	for _, m := range existing {
		seen[m.RuleID] = struct{}{}
	}
	for _, m := range lt.Check(text) {
		if _, ok := pickyIDs[m.RuleID]; !ok {
			continue
		}
		if _, dup := seen[m.RuleID]; dup {
			continue
		}
		existing = append(existing, m)
		seen[m.RuleID] = struct{}{}
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
// langCode drives RuleURL (community rule page language).
func LocalMatchesToRemote(text string, matches []languagetool.LocalMatch, contextSize int, langCode string) []RemoteRuleMatch {
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
		catID, catName, issue, short := RuleMeta(ruleID)
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
		// Java Rule.estimateContextForSureMatch — carried on LocalMatch (text-level -1).
		// Do not invent from ITS issue type (style/register → -1 was soft invent).
		rm.EstimatedContextForSureMatch = m.EstimateContextForSureMatch
		if m.Description != "" {
			rm.Description = m.Description
		} else {
			rm.Description = RuleDescription(ruleID)
		}
		if rm.ShortMessage == "" {
			rm.ShortMessage = short
		}
		if rm.URL == "" {
			rm.URL = languagetool.RuleURL(ruleID, langCode)
		}
		rm.Replacements = append([]string(nil), m.Suggestions...)
		out = append(out, *rm)
	}
	return out
}


