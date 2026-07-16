package ca

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// AdjustPronounsFilter ports action-based pronoun rewrites from
// org.languagetool.rules.ca.AdjustPronounsFilter (without VerbSynthesizer).
// Callers supply verb/pronoun surface context.
type AdjustPronounsFilter struct{}

func NewAdjustPronounsFilter() *AdjustPronounsFilter {
	return &AdjustPronounsFilter{}
}

// PronounVerbContext is the surface input for AdjustPronouns actions.
type PronounVerbContext struct {
	PronounsStr            string // weak pronouns as surface string
	VerbStr                string // main verb form
	VerbStr2               string // optional alternate lemma form (changeOnlyLemma)
	FirstVerbPersonaNumber string // e.g. "1S"
	PronounsAfter          bool   // clitics after verb
	WholeOriginal          string // original span for de-dup
	CasingModel            string // model token for PreserveCase
}

// Suggest applies comma-separated actions and returns unique replacements.
func (f *AdjustPronounsFilter) Suggest(ctx PronounVerbContext, actionsCSV string) []string {
	var out []string
	seen := map[string]struct{}{}
	actions := strings.Split(actionsCSV, ",")
	for _, action := range actions {
		action = strings.TrimSpace(action)
		var replacement string
		switch action {
		case "removePronounEn":
			pr := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(ctx.PronounsStr, "en", ""), "n'", ""), "'n", ""))
			replacement = TransformDavant(pr, ctx.VerbStr) + ctx.VerbStr
		case "addPronounEn":
			newPronoun := DoAddPronounEn(ctx.PronounsStr, ctx.VerbStr, ctx.PronounsAfter)
			if newPronoun != "" {
				if !ctx.PronounsAfter {
					replacement = newPronoun + ctx.VerbStr
				} else {
					replacement = ctx.VerbStr + newPronoun
				}
			}
		case "removePronounReflexive":
			replacement = DoRemovePronounReflexive(ctx.PronounsStr, ctx.VerbStr, false)
		case "replaceEmEn":
			replacement = DoReplaceEmEn(ctx.PronounsStr, ctx.VerbStr, false)
		case "replaceHiEn":
			// Java: doAddPronounEn(transform(pronouns without hi), verb, false) + verb
			norm := Transform(strings.TrimSpace(strings.ReplaceAll(ctx.PronounsStr, "hi", "")), PronounNormalized)
			replacement = DoAddPronounEn(norm, ctx.VerbStr, false) + ctx.VerbStr
		case "addPronounReflexive":
			replacement = DoAddPronounReflexive(ctx.PronounsStr, ctx.VerbStr, ctx.FirstVerbPersonaNumber, ctx.PronounsAfter)
		case "addPronounReflexiveHi":
			replacement = DoAddPronounReflexive(ctx.PronounsStr, "hi "+ctx.VerbStr, ctx.FirstVerbPersonaNumber, false)
		case "addPronounReflexiveImperative":
			replacement = DoAddPronounReflexiveImperative(ctx.PronounsStr, ctx.VerbStr, ctx.FirstVerbPersonaNumber)
		case "changeOnlyLemma":
			pronounsStr := ctx.PronounsStr
			if containsAction(actions, "replaceHiEn") {
				pronounsStr = strings.TrimSpace(strings.ReplaceAll(pronounsStr, "hi", ""))
			}
			verb2 := ctx.VerbStr2
			if verb2 == "" {
				verb2 = ctx.VerbStr
			}
			pronounsStr = TransformDavant(pronounsStr, verb2)
			replacement = pronounsStr + verb2
		}
		if replacement == "" {
			continue
		}
		if strings.EqualFold(replacement, ctx.WholeOriginal) {
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

func containsAction(actions []string, want string) bool {
	for _, a := range actions {
		if strings.TrimSpace(a) == want {
			return true
		}
	}
	return false
}
