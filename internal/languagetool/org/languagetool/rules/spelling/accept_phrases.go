package spelling

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// AcceptPhrases ports SpellingCheckRule.acceptPhrases.
// Java splits on space (not WordTokenizer) and builds IGNORE_SPELLING antipatterns;
// lowercase-leading phrases also get a sentence-start uppercase variant.
func (r *SpellingCheckRule) AcceptPhrases(phrases []string) {
	if r == nil {
		return
	}
	for _, phrase := range phrases {
		phrase = strings.TrimSpace(phrase)
		if phrase == "" {
			continue
		}
		// Java: phrase.split(" ") — drop empty mid segments from double spaces.
		raw := strings.Split(phrase, " ")
		parts := make([]string, 0, len(raw))
		for _, p := range raw {
			if p != "" {
				parts = append(parts, p)
			}
		}
		if len(parts) == 0 {
			continue
		}
		// Match-time multi-word ignore (MarkMultiWordIgnoreSpelling).
		r.MultiWordIgnore = append(r.MultiWordIgnore, append([]string(nil), parts...))
		// Java acceptPhrases → makeAntiPatterns → IMMUNIZE (not IGNORE_SPELLING).
		r.appendImmunizeAntiPattern(parts)

		// Java: if first token is not title-cased, also accept sentence-start form.
		if !isTitleCaseToken(parts[0]) {
			ucParts := append([]string(nil), parts...)
			ucParts[0] = tools.UppercaseFirstChar(parts[0])
			// Sentence-start pattern: SENT_START + uppercased first + rest (IMMUNIZE).
			r.appendSentenceStartImmunizeAntiPattern(ucParts)
			// Match path: capitalized phrase after sentence start.
			r.MultiWordIgnore = append(r.MultiWordIgnore, append([]string(nil), ucParts...))
		}
	}
}

// GetAntiPatterns ports SpellingCheckRule.getAntiPatterns (IGNORE_SPELLING rules).
func (r *SpellingCheckRule) GetAntiPatterns() []*disambigrules.DisambiguationPatternRule {
	if r == nil || len(r.AntiPatterns) == 0 {
		return nil
	}
	return append([]*disambigrules.DisambiguationPatternRule(nil), r.AntiPatterns...)
}

// appendIgnoreSpellingAntiPattern builds Java multi-token addIgnoreWords antipattern
// (DisambiguatorAction.IGNORE_SPELLING).
func (r *SpellingCheckRule) appendIgnoreSpellingAntiPattern(tokens []string) {
	if r == nil || len(tokens) == 0 {
		return
	}
	pts := make([]*patterns.PatternToken, 0, len(tokens))
	for _, t := range tokens {
		// Java: new PatternToken(token, true, false, false) — case-sensitive surface.
		pts = append(pts, patterns.CsToken(t))
	}
	ap := disambigrules.NewDisambiguationPatternRule(
		"INTERNAL_ANTIPATTERN", "(no description)", r.LanguageCode,
		pts, "", nil, disambigrules.ActionIgnoreSpelling,
	)
	r.AntiPatterns = append(r.AntiPatterns, ap)
}

// appendImmunizeAntiPattern ports Rule.makeAntiPatterns (IMMUNIZE action).
// Used by acceptPhrases.
func (r *SpellingCheckRule) appendImmunizeAntiPattern(tokens []string) {
	if r == nil || len(tokens) == 0 {
		return
	}
	pts := make([]*patterns.PatternToken, 0, len(tokens))
	for _, t := range tokens {
		pts = append(pts, patterns.CsToken(t))
	}
	ap := disambigrules.NewDisambiguationPatternRule(
		"INTERNAL_ANTIPATTERN", "(no description)", r.LanguageCode,
		pts, "", nil, disambigrules.ActionImmunize,
	)
	r.AntiPatterns = append(r.AntiPatterns, ap)
}

// appendSentenceStartImmunizeAntiPattern ports getTokensForSentenceStart via makeAntiPatterns:
// posRegex SENT_START + csToken(uppercased first) + csToken(rest…) with IMMUNIZE.
func (r *SpellingCheckRule) appendSentenceStartImmunizeAntiPattern(tokens []string) {
	if r == nil || len(tokens) == 0 {
		return
	}
	pts := make([]*patterns.PatternToken, 0, len(tokens)+1)
	pts = append(pts, patterns.PosRegex(languagetool.SentenceStartTagName))
	for _, t := range tokens {
		pts = append(pts, patterns.CsToken(t))
	}
	ap := disambigrules.NewDisambiguationPatternRule(
		"INTERNAL_ANTIPATTERN", "(no description)", r.LanguageCode,
		pts, "", nil, disambigrules.ActionImmunize,
	)
	r.AntiPatterns = append(r.AntiPatterns, ap)
}

func isTitleCaseToken(s string) bool {
	if s == "" {
		return true
	}
	// Java: part.equals(StringTools.uppercaseFirstChar(part))
	return s == tools.UppercaseFirstChar(s)
}
