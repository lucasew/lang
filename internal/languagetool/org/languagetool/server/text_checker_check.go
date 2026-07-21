package server

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/markup"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
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
	// AltLanguages ports QueryParams.altLanguages (e.g. "de-DE", "ru-RU").
	// Java: Pipeline(lang, altLanguages, …) → JLanguageTool.altLanguages.
	AltLanguages []string
	// AllowIncompleteResults ports QueryParams.allowIncompleteResults —
	// when true, ErrorRateTooHighException / check timeout return partial matches + incomplete reason.
	AllowIncompleteResults bool
	// MaxErrorsPerWordRate ports JLanguageTool.maxErrorsPerWordRate (0 = disabled).
	// Used for tests / server config; Java often sets from server properties.
	MaxErrorsPerWordRate float64
	// MaxCheckTimeMillis ports UserLimits.maxCheckTimeMillis / future.get(timeout).
	// <0 means unlimited (Java). When >0 and exceeded with AllowIncompleteResults,
	// incompleteReason = "Results are incomplete: text checking took longer than allowed maximum of X.XX seconds".
	MaxCheckTimeMillis int64
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
	settings.Query.AltLanguages = append([]string(nil), opts.AltLanguages...)
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
	// altLanguages also in Key() via Query.AltLanguages
	if len(opts.AltLanguages) > 0 {
		keyParts = append(keyParts, "alt:"+strings.Join(opts.AltLanguages, ","))
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
		// Java getCommaSeparatedStrings: split(",") with no per-id trim.
		if id != "" {
			_ = p.DisableRuleID(id)
		}
	}
	_ = p.SetCleanOverlappingMatches(true)
	// maxErrorsPerWordRate before freeze (Java JLanguageTool field used in TextCheckCallable).
	if opts.MaxErrorsPerWordRate > 0 {
		_ = p.SetMaxErrorsPerWordRate(opts.MaxErrorsPerWordRate)
	}
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
	ms, _, _ := t.CheckWithOptionsAndIgnore(text, lang, opts)
	return ms
}

// CheckWithOptionsAndIgnore ports check2 → matches + ignoreRanges + incomplete reason.
// When AllowIncompleteResults and ErrorRateTooHighException: returns matches so far and
// incompleteReason = "Results are incomplete: " + exception message (Java TextChecker).
// When MaxCheckTimeMillis > 0 and check exceeds it with AllowIncompleteResults:
// incompleteReason = "Results are incomplete: text checking took longer than allowed maximum of X.XX seconds"
// (Java TimeoutException path). Not invent ForeignScriptIgnoreRanges / size-threshold soft warnings.
func (t *TextChecker) CheckWithOptionsAndIgnore(text, lang string, opts CheckOptions) (matches []RemoteRuleMatch, ignore []IgnoreRangeInfo, incompleteReason string) {
	type out struct {
		matches []RemoteRuleMatch
		ignore  []IgnoreRangeInfo
		reason  string
	}
	run := func() out {
		p, settings, fromPool := t.preparePipeline(lang, opts)
		defer t.releasePipeline(settings, p, fromPool)
		cr, err := p.CheckWithResults(text)
		var reason string
		if err != nil {
			var rateErr *languagetool.ErrorRateTooHighException
			if errors.As(err, &rateErr) {
				if opts.AllowIncompleteResults {
					// Java: localReason = "Results are incomplete: " + rootCause.getMessage()
					reason = "Results are incomplete: " + rateErr.Error()
				} else {
					return out{}
				}
			}
		}
		locals := languagetool.LocalMatchesFromCheckResults(cr)
		// Tag.picky rules are gated by Pipeline → lt.Level (Java setLevel), not invent re-check.
		locals = applyRuleValues(lang, text, locals, opts.RuleValues)
		locals = filterLocalsByIgnoreWords(text, locals, opts.IgnoreWords)
		locals = filterLocalsByCategories(locals, opts)
		ctxSize := DefaultContextSize
		if t != nil && t.ContextSize > 0 {
			ctxSize = t.ContextSize
		}
		var ign []IgnoreRangeInfo
		if cr != nil {
			ign = RangesToIgnoreRangeInfo(cr.GetIgnoredRanges())
		}
		return out{
			matches: LocalMatchesToRemote(text, locals, ctxSize, lang),
			ignore:  ign,
			reason:  reason,
		}
	}

	maxMs := opts.MaxCheckTimeMillis
	if maxMs < 0 {
		// unlimited — Java future.get() without timeout
		o := run()
		return o.matches, o.ignore, o.reason
	}
	if maxMs == 0 {
		// treat 0 as unlimited unless config says otherwise
		o := run()
		return o.matches, o.ignore, o.reason
	}

	// Java: future.get(limits.getMaxCheckTimeMillis(), TimeUnit.MILLISECONDS)
	ch := make(chan out, 1)
	go func() { ch <- run() }()
	select {
	case o := <-ch:
		return o.matches, o.ignore, o.reason
	case <-time.After(time.Duration(maxMs) * time.Millisecond):
		// Without cooperative cancel, matches-so-far is empty (no invent partial invent).
		// Message format matches Java Locale.ENGLISH "%.2f" seconds.
		if opts.AllowIncompleteResults {
			return nil, nil, formatTimeoutIncompleteReason(maxMs)
		}
		// Java: throw RuntimeException — callers without allowIncomplete get empty (no body invent).
		return nil, nil, ""
	}
}

// formatTimeoutIncompleteReason ports Java TextChecker TimeoutException incomplete message:
// "Results are incomplete: text checking took longer than allowed maximum of " +
// String.format(Locale.ENGLISH, "%.2f", maxCheckTimeMillis / 1000.0) + " seconds"
func formatTimeoutIncompleteReason(maxCheckTimeMillis int64) string {
	sec := float64(maxCheckTimeMillis) / 1000.0
	return fmt.Sprintf(
		"Results are incomplete: text checking took longer than allowed maximum of %.2f seconds",
		sec,
	)
}

// RangesToIgnoreRangeInfo maps languagetool.Range (UTF-16 spans from Java Range)
// to API IgnoreRangeInfo for /v2/check JSON.
func RangesToIgnoreRangeInfo(ranges []languagetool.Range) []IgnoreRangeInfo {
	if len(ranges) == 0 {
		return nil
	}
	out := make([]IgnoreRangeInfo, 0, len(ranges))
	for _, r := range ranges {
		out = append(out, IgnoreRangeInfo{From: r.FromPos, To: r.ToPos, Lang: r.Lang})
	}
	return out
}

// CheckAnnotatedWithOptions checks annotated markup text; match offsets are in original markup space.
func (t *TextChecker) CheckAnnotatedWithOptions(at *markup.AnnotatedText, lang string, opts CheckOptions) []RemoteRuleMatch {
	ms, _, _ := t.CheckAnnotatedWithOptionsAndIgnore(at, lang, opts)
	return ms
}

// CheckAnnotatedWithOptionsAndIgnore is the annotated twin of CheckWithOptionsAndIgnore
// (Java check2 on AnnotatedText → matches + ignoreRanges + incomplete reason).
func (t *TextChecker) CheckAnnotatedWithOptionsAndIgnore(at *markup.AnnotatedText, lang string, opts CheckOptions) (matches []RemoteRuleMatch, ignore []IgnoreRangeInfo, incompleteReason string) {
	if at == nil {
		return nil, nil, ""
	}
	p, settings, fromPool := t.preparePipeline(lang, opts)
	defer t.releasePipeline(settings, p, fromPool)
	cr, err := p.CheckAnnotatedWithResults(at)
	if err != nil {
		var rateErr *languagetool.ErrorRateTooHighException
		if errors.As(err, &rateErr) && opts.AllowIncompleteResults {
			incompleteReason = "Results are incomplete: " + rateErr.Error()
		} else if errors.As(err, &rateErr) {
			return nil, nil, ""
		}
	}
	locals := languagetool.LocalMatchesFromCheckResults(cr)
	plain := at.GetPlainText()
	// Tag.picky rules are gated by Pipeline → lt.Level (Java setLevel), not invent re-check.
	locals = applyRuleValues(lang, plain, locals, opts.RuleValues)
	locals = filterLocalsByIgnoreWords(plain, locals, opts.IgnoreWords)
	locals = filterLocalsByCategories(locals, opts)
	// Context uses original markup string so projected offsets align.
	orig := at.GetTextWithMarkup()
	ctxSize := DefaultContextSize
	if t != nil && t.ContextSize > 0 {
		ctxSize = t.ContextSize
	}
	if cr != nil {
		ignore = RangesToIgnoreRangeInfo(cr.GetIgnoredRanges())
	}
	return LocalMatchesToRemote(orig, locals, ctxSize, lang), ignore, incompleteReason
}

func filterLocalsByIgnoreWords(text string, ms []languagetool.LocalMatch, ignore []string) []languagetool.LocalMatch {
	return languagetool.FilterMatchesByIgnoreWords(text, ms, ignore)
}

// filterLocalsByCategories applies disabledCategories / enabledCategories (with enabledOnly).
func filterLocalsByCategories(ms []languagetool.LocalMatch, opts CheckOptions) []languagetool.LocalMatch {
	return languagetool.FilterMatchesByCategories(ms, opts.DisabledCategories, opts.EnabledCategories, opts.UseEnabledOnly)
}

// ParseRuleValues ports TextChecker.getRuleValues for string maps used by applyRuleValues.
// Java: parameterString.split(","); pair.split(":"); put(ruleAndValue[0], …) — no trim.
// Accepts either a single comma-joined blob or a pre-split []string of pairs.
func ParseRuleValues(items []string) map[string]string {
	if len(items) == 0 {
		return nil
	}
	out := map[string]string{}
	for _, item := range items {
		if item == "" {
			continue
		}
		// also allow comma-joined blob (one element holding many pairs)
		for _, part := range strings.Split(item, ",") {
			if part == "" {
				continue
			}
			// Java pair.split(":") — first colon only needed for id/value (limit not set → all colons)
			i := strings.IndexByte(part, ':')
			if i < 0 || i == len(part)-1 {
				// Java would throw ArrayIndexOutOfBounds on missing [1]; fail-closed skip
				continue
			}
			id := part[:i]
			val := part[i+1:]
			// Java uses ruleAndValue[0] as key without ToUpper; applyRuleValues may fold later.
			// Keep historical ToUpper for map lookup consistency with existing Go consumers.
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
	// Default LongSentenceRule is Tag.picky; ruleValues re-run is an explicit
	// user threshold so use Level.PICKY (Java UserConfig max-words still picky).
	lt := languagetool.NewJLanguageTool(lang)
	lt.Level = languagetool.LevelPicky
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
		// Prefer metadata carried on LocalMatch (Java Rule category / grammar XML).
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


