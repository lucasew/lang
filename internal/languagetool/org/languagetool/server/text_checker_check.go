package server

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/en"
)

// CheckOptions carries optional check-query knobs beyond language/disabled rules.
type CheckOptions struct {
	Disabled       []string
	Enabled        []string
	UseEnabledOnly bool
	Mode           CheckMode
	Level          CheckLevel
}

// Check runs core rules for language on text and returns RemoteRuleMatch results.
// Uses Pipeline so disabled/enabled-only rules from query params can be applied.
func (t *TextChecker) Check(text, lang string, disabled []string) []RemoteRuleMatch {
	return t.CheckWithOptions(text, lang, CheckOptions{Disabled: disabled})
}

// CheckWithOptions is Check with enabled-only, mode, and level support.
func (t *TextChecker) CheckWithOptions(text, lang string, opts CheckOptions) []RemoteRuleMatch {
	if lang == "" {
		lang = "en"
	}
	settings := NewPipelineSettings(lang, "check")
	settings.Query.DisabledRules = append([]string(nil), opts.Disabled...)
	settings.Query.EnabledRules = append([]string(nil), opts.Enabled...)
	settings.Query.UseEnabledOnly = opts.UseEnabledOnly
	p := NewPipeline(settings)
	for _, id := range opts.Disabled {
		id = strings.TrimSpace(id)
		if id != "" {
			_ = p.DisableRuleID(id)
		}
	}
	_ = p.SetCleanOverlappingMatches(true)

	// soft: apply level before freeze via post-register picky inject on pipeline Check
	// Pipeline.Check builds a fresh JLT; pass level through settings.Query language field abuse avoided —
	// store on PipelineSettings via MotherTongueCode unused field is wrong.
	// Instead: set level on pipeline settings GlobalConfigKey soft tag.
	if opts.Level != "" {
		settings.GlobalConfigKey = "level:" + string(opts.Level)
	}
	if opts.Mode != "" {
		// soft: Query.LanguageCode carries check mode for Pipeline.Check
		settings.Query.LanguageCode = string(opts.Mode)
	}
	p.settings = settings

	p.SetupFinished()
	locals := p.Check(text)

	// post-filter mode soft: already handled if pipeline honors mode; else filter here
	locals = applyLevelPickyBoost(lang, opts.Level, locals, text)

	ctxSize := DefaultContextSize
	if t != nil && t.ContextSize > 0 {
		ctxSize = t.ContextSize
	}
	return LocalMatchesToRemote(text, locals, ctxSize)
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
		rm.Replacements = append([]string(nil), m.Suggestions...)
		out = append(out, *rm)
	}
	return out
}
