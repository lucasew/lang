package patterns

import (
	"strconv"
	"strings"
)

// FalseFriendRuleHandler ports the SAX-side surface of
// org.languagetool.rules.patterns.FalseFriendRuleHandler without embedding a full SAX stack.
// Pairing logic lives in FalseFriendRuleLoader; this type accumulates mother-tongue suggestions
// and formats description hints the way the Java handler does.
type FalseFriendRuleHandler struct {
	TextLanguage   string
	MotherTongue   string
	FalseFriendHint string
	// SuggestionMap is rule ID → mother-tongue translation strings for the reverse side.
	SuggestionMap map[string][]string
	InTestMode    bool
}

func NewFalseFriendRuleHandler(textLang, motherLang, falseFriendHint string) *FalseFriendRuleHandler {
	return &FalseFriendRuleHandler{
		TextLanguage:    textLang,
		MotherTongue:    motherLang,
		FalseFriendHint: falseFriendHint,
		SuggestionMap:   map[string][]string{},
	}
}

// FormatTranslations joins translations as "a", "b" (Java stream quoting).
func FormatTranslations(translations []string) string {
	parts := make([]string, 0, len(translations))
	for _, t := range translations {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		parts = append(parts, `"`+t+`"`)
	}
	return strings.Join(parts, ", ")
}

// FormatHint applies a simple {0}-style format using the false-friend hint template.
// Java uses MessageFormat; here we replace {0},{1},{2},{3} positionally.
func (h *FalseFriendRuleHandler) FormatHint(tokensAsString, textLangName, translations, motherName string) string {
	tpl := ""
	if h != nil {
		tpl = h.FalseFriendHint
	}
	if tpl == "" {
		// MessagesBundle_en false_friend_hint (not invent 2-arg templates)
		tpl = messagesFalseFriendHint
	}
	repl := []string{tokensAsString, textLangName, translations, motherName}
	out := tpl
	for i, v := range repl {
		// support both '{0}' and {0}
		out = strings.ReplaceAll(out, "'{"+itoa(i)+"}'", v)
		out = strings.ReplaceAll(out, "{"+itoa(i)+"}", v)
	}
	return out
}

// AddSuggestions records reverse-side suggestions for a rule id.
func (h *FalseFriendRuleHandler) AddSuggestions(ruleID string, translations []string) {
	if h == nil || ruleID == "" || len(translations) == 0 {
		return
	}
	if h.SuggestionMap == nil {
		h.SuggestionMap = map[string][]string{}
	}
	cur := h.SuggestionMap[ruleID]
	seen := map[string]struct{}{}
	for _, s := range cur {
		seen[s] = struct{}{}
	}
	for _, t := range translations {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		cur = append(cur, t)
	}
	h.SuggestionMap[ruleID] = cur
}

// ShouldEmitRule mirrors Java endElement(RULE) language/mother-tongue pairing checks.
func (h *FalseFriendRuleHandler) ShouldEmitRule(patternLang, translationLang string, hasTranslations bool) bool {
	if h == nil || !hasTranslations {
		return false
	}
	if baseLang(patternLang) != baseLang(h.TextLanguage) {
		return false
	}
	if baseLang(translationLang) != baseLang(h.MotherTongue) {
		return false
	}
	if baseLang(patternLang) == baseLang(h.MotherTongue) {
		return false
	}
	return true
}

func itoa(i int) string {
	return strconv.Itoa(i)
}
