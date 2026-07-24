package fr

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// InterrogativeVerbFilter ports org.languagetool.rules.fr.InterrogativeVerbFilter.
// SpellingSuggestions + MatchesDesiredPostag are required for speller path;
// SynthesizeParticiple optional for je-form extras (trompè-je / trompé-je).
type InterrogativeVerbFilter struct {
	// SpellingSuggestions returns candidates for a (possibly wrong) verb form.
	SpellingSuggestions func(wrongVerb string) []string
	// MatchesDesiredPostag reports whether a candidate matches desired POS regex.
	MatchesDesiredPostag func(candidate, desiredPostag string) bool
	// SynthesizeParticiple(lemmaReading, postagRE) → participle surfaces (Java FrenchSynthesizer).
	SynthesizeParticiple func(token, lemma, postag string) []string
}

func NewInterrogativeVerbFilter() *InterrogativeVerbFilter {
	return &InterrogativeVerbFilter{}
}

// DesiredPostagFromPronounPOS ports the Java MatchesPosTagRegex chain on the pronoun ATR.
func DesiredPostagFromPronounPOS(pronoun *languagetool.AnalyzedTokenReadings) string {
	if pronoun == nil {
		return ""
	}
	// Order matches Java InterrogativeVerbFilter.acceptRuleMatch.
	switch {
	case pronoun.MatchesPosTagRegex(`R pers obj 2 p`):
		return "V.* (imp) [23] [sp]|V .*(ind|cond).* 2 p"
	case pronoun.MatchesPosTagRegex(`R pers obj 1 p`):
		return "V.* (imp) .*|V .*(ind|cond).* 1 p"
	case pronoun.MatchesPosTagRegex(`R pers obj.*`):
		return "V.* (imp) .*"
	case pronoun.MatchesPosTagRegex(`.* 1 s`):
		return "V .*(ind|cond).* 1 s"
	case pronoun.MatchesPosTagRegex(`.* 2 s`):
		return "V .*(ind|cond).* 2 s"
	case pronoun.MatchesPosTagRegex(`.* 3( [mfe])? s`):
		return "V .*(ind|cond).* 3 s"
	case pronoun.MatchesPosTagRegex(`.* 1 p`):
		return "V .*(ind|cond).* 1 p"
	case pronoun.MatchesPosTagRegex(`.* 2 p`):
		return "V .*(ind|cond).* 2 p"
	case pronoun.MatchesPosTagRegex(`.* 3( [mf])? p`):
		return "V .*(ind|cond).* 3 p"
	default:
		return ""
	}
}

// DesiredPostagForPronoun maps surface pronouns (legacy helper / tests).
// Prefer DesiredPostagFromPronounPOS when POS readings are available.
func (f *InterrogativeVerbFilter) DesiredPostagForPronoun(pronoun string) string {
	switch strings.ToLower(pronoun) {
	case "tu":
		return "V.* (imp) [23] [sp]|V .*(ind|cond).* 2 p"
	case "vous":
		return "V.* (imp) .*|V .*(ind|cond).* 1 p"
	case "nous":
		return "V.* (imp) .*"
	case "je", "j":
		return "V .*(ind|cond).* 1 s"
	case "il", "elle", "on":
		return "V .*(ind|cond).* 3 s"
	case "ils", "elles":
		return "V .*(ind|cond).* 3 p"
	case "toi":
		return "V .*(ind|cond).* 2 s"
	default:
		return ""
	}
}

// MakeWrong invents a wrong word for the speller when the original is correct (Java makeWrong).
func MakeWrong(s string) string {
	repls := []struct{ old, new string }{
		{"a", "ä"}, {"e", "ë"}, {"i", "í"}, {"o", "ö"}, {"u", "ü"},
		{"é", "ë"}, {"à", "ä"}, {"è", "ë"}, {"ù", "ü"},
		{"â", "ä"}, {"ê", "ë"}, {"î", "ï"}, {"ô", "ö"}, {"û", "ü"},
	}
	for _, r := range repls {
		if strings.Contains(s, r.old) {
			return strings.Replace(s, r.old, r.new, 1)
		}
	}
	return s + "-"
}

// AcceptRuleMatch ports InterrogativeVerbFilter.acceptRuleMatch.
// Args: PronounFrom, VerbFrom (1-based pattern token indexes).
// Without SpellingSuggestions returns match with no new suggestions (Java empty list keeps match).
func (f *InterrogativeVerbFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	patternTokens []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if f == nil || match == nil {
		return nil
	}
	pronounFrom, ok1 := arguments["PronounFrom"]
	verbFrom, ok2 := arguments["VerbFrom"]
	if !ok1 || !ok2 {
		panic("Missing key 'PronounFrom' or 'VerbFrom'")
	}
	posPronoun, err1 := strconv.Atoi(pronounFrom)
	posVerb, err2 := strconv.Atoi(verbFrom)
	if err1 != nil || err2 != nil {
		return nil
	}
	if posPronoun < 1 || posPronoun > len(patternTokens) {
		panic(fmt.Sprintf("ConfusionCheckFilter: Index out of bounds, PronounFrom: %d", posPronoun))
	}
	if posVerb < 1 || posVerb > len(patternTokens) {
		panic(fmt.Sprintf("ConfusionCheckFilter: Index out of bounds, VerbFrom: %d", posVerb))
	}
	atrPronoun := patternTokens[posPronoun-1]
	atrVerb := patternTokens[posVerb-1]
	if atrPronoun == nil || atrVerb == nil {
		return match
	}

	desiredPostag := DesiredPostagFromPronounPOS(atrPronoun)
	var replacements []string
	var extraSuggestions []string

	// Java: extra participles for ".* 1 s" (je) via FrenchSynthesizer.
	if atrPronoun.MatchesPosTagRegex(`.* 1 s`) && f.SynthesizeParticiple != nil {
		for _, r := range atrVerb.GetReadings() {
			if r == nil || r.GetPOSTag() == nil {
				continue
			}
			// Java: readingWithTagRegex("V .*")
			if !strings.HasPrefix(*r.GetPOSTag(), "V") {
				continue
			}
			lemma := ""
			if r.GetLemma() != nil {
				lemma = *r.GetLemma()
			}
			parts := f.SynthesizeParticiple(r.GetToken(), lemma, "V ppa [me] sp?")
			if len(parts) > 0 && strings.HasSuffix(parts[0], "é") {
				p0 := parts[0]
				extraSuggestions = append(extraSuggestions, p0)
				extraSuggestions = append(extraSuggestions, p0[:len(p0)-len("é")]+"è")
			}
			break
		}
	}

	pronounTok := atrPronoun.GetToken()
	sep := "-"
	if strings.HasPrefix(pronounTok, "-") {
		sep = ""
	}

	if len(extraSuggestions) > 0 {
		for _, es := range extraSuggestions {
			complete := es + sep + pronounTok
			if !containsStr(replacements, complete) && !strings.HasSuffix(complete, "e-je") {
				replacements = append(replacements, complete)
			}
		}
	} else if desiredPostag != "" && f.SpellingSuggestions != nil {
		query := atrVerb.GetToken()
		if atrVerb.IsTagged() {
			query = MakeWrong(query)
		}
		for _, sug := range f.SpellingSuggestions(query) {
			if f.MatchesDesiredPostag != nil && !f.MatchesDesiredPostag(sug, desiredPostag) {
				continue
			}
			complete := sug + sep + pronounTok
			if strings.EqualFold(complete, "peux-je") {
				complete = tools.PreserveCase("puis-je", complete)
			}
			if strings.HasSuffix(complete, "e-je") {
				base := complete[:len(complete)-4]
				for _, end := range []string{"é-je", "è-je"} {
					c2 := base + end
					if !containsStr(replacements, c2) {
						replacements = append(replacements, c2)
					}
				}
			} else if !containsStr(replacements, complete) {
				replacements = append(replacements, complete)
			}
		}
	}

	out := rules.NewRuleMatch(match.GetRule(), match.Sentence, match.GetFromPos(), match.GetToPos(), match.GetMessage())
	out.ShortMessage = match.ShortMessage
	if len(replacements) > 0 {
		out.SetSuggestedReplacements(replacements)
	}
	return out
}

// FilterByDesiredPOS keeps candidates whose POS matches desiredPostag regex (tests).
func (f *InterrogativeVerbFilter) FilterByDesiredPOS(candidates []string, desiredPostag string, matchesPOS func(form, postagRE string) bool) []string {
	if matchesPOS == nil || desiredPostag == "" {
		return candidates
	}
	var out []string
	for _, c := range candidates {
		if matchesPOS(c, desiredPostag) {
			out = append(out, c)
		}
	}
	return out
}

func containsStr(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}
