package ca

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// AdjustVerbSuggestionsFilter ports action-based verb+pronoun rewrites
// (surface twin of AdjustVerbSuggestionsFilter using PronomsFeblesHelper).
type AdjustVerbSuggestionsFilter struct{}

func NewAdjustVerbSuggestionsFilter() *AdjustVerbSuggestionsFilter {
	return &AdjustVerbSuggestionsFilter{}
}

// VerbSuggestionContext is the surface verb/pronoun input.
type VerbSuggestionContext struct {
	PronounsStr            string
	VerbStr                string
	FirstVerbPersonaNumber string
	PronounsAfter          bool
	WholeOriginal          string
	CasingModel            string
}

// Suggest applies the first action (Java uses actions[0] primarily) and optional extras.
func (f *AdjustVerbSuggestionsFilter) Suggest(ctx VerbSuggestionContext, actionsCSV string) []string {
	actions := strings.Split(actionsCSV, ",")
	if len(actions) == 0 {
		return nil
	}
	// default like Java optional default
	if actionsCSV == "" {
		actions = []string{"removePronounReflexive"}
	}
	var out []string
	seen := map[string]struct{}{}
	for _, action := range actions {
		action = strings.TrimSpace(action)
		var replacement string
		switch action {
		case "addPronounEn":
			np := DoAddPronounEn(ctx.PronounsStr, ctx.VerbStr, ctx.PronounsAfter)
			if np != "" {
				if ctx.PronounsAfter {
					replacement = ctx.VerbStr + np
				} else {
					replacement = np + ctx.VerbStr
				}
			}
		case "removePronounReflexive":
			replacement = DoRemovePronounReflexive(ctx.PronounsStr, ctx.VerbStr, ctx.PronounsAfter)
		case "addPronounReflexiveEn":
			replacement = DoAddPronounReflexiveEn(ctx.PronounsStr, ctx.VerbStr, ctx.FirstVerbPersonaNumber, ctx.PronounsAfter)
		case "replaceEmEn":
			replacement = DoReplaceEmEn(ctx.PronounsStr, ctx.VerbStr, ctx.PronounsAfter)
		case "addPronounReflexive":
			replacement = DoAddPronounReflexive(ctx.PronounsStr, ctx.VerbStr, ctx.FirstVerbPersonaNumber, ctx.PronounsAfter)
		case "addPronounReflexiveHi":
			replacement = DoAddPronounReflexive(ctx.PronounsStr, "hi "+ctx.VerbStr, ctx.FirstVerbPersonaNumber, false)
		case "addPronounReflexiveImperative":
			replacement = DoAddPronounReflexiveImperative(ctx.PronounsStr, ctx.VerbStr, ctx.FirstVerbPersonaNumber)
		case "addPronounHi":
			// simple surface: prefix "hi " if not present
			if !strings.Contains(ctx.PronounsStr, "hi") {
				replacement = TransformDavant("hi", ctx.VerbStr) + ctx.VerbStr
			}
		case "None", "none", "":
			continue
		default:
			// addPronounDative / Les / Ho etc. need synthesizer — skip
			continue
		}
		if replacement == "" || strings.EqualFold(replacement, ctx.WholeOriginal) {
			continue
		}
		if ctx.CasingModel != "" {
			replacement = tools.PreserveCase(replacement, ctx.CasingModel)
		}
		if _, ok := seen[replacement]; ok {
			continue
		}
		seen[replacement] = struct{}{}
		out = append(out, replacement)
	}
	return out
}
