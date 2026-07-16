package de

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// SubjectVerbAgreementRule is a surface stand-in for SubjectVerbAgreementRule.
// Flags clear plural-subject + singular "ist" / "war" mismatches without full chunking.
type SubjectVerbAgreementRule struct {
	Messages map[string]string
}

func NewSubjectVerbAgreementRule(messages map[string]string) *SubjectVerbAgreementRule {
	return &SubjectVerbAgreementRule{Messages: messages}
}

func (r *SubjectVerbAgreementRule) GetID() string { return "DE_SUBJECT_VERB_AGREEMENT" }

func looksPluralNounDE(w string) bool {
	if !tools.StartsWithUppercase(w) || utf8.RuneCountInString(w) < 4 {
		return false
	}
	lc := strings.ToLower(w)
	// common plural endings; exclude some frequent singulars
	if strings.HasSuffix(lc, "en") || strings.HasSuffix(lc, "ern") || strings.HasSuffix(lc, "eln") {
		return true
	}
	if strings.HasSuffix(lc, "s") && !strings.HasSuffix(lc, "us") && !strings.HasSuffix(lc, "is") {
		return true
	}
	// Kenntnisse, Maße, ...
	if strings.HasSuffix(lc, "sse") || strings.HasSuffix(lc, "nse") {
		return true
	}
	return false
}

func (r *SubjectVerbAgreementRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	tokens := sentence.GetTokensWithoutWhitespace()
	var matches []*rules.RuleMatch
	for i := 2; i < len(tokens); i++ {
		v := strings.ToLower(tokens[i].GetToken())
		if v != "ist" && v != "war" {
			continue
		}
		// look back for plural subject pattern
		if !pluralSubjectBefore(tokens, i) {
			continue
		}
		msg := "Möglicherweise fehlt die Subjekt-Verb-Kongruenz (Plural-Subjekt mit Singular-Verb)."
		rm := rules.NewRuleMatch(r, sentence, tokens[i].GetStartPos(), tokens[i].GetEndPos(), msg)
		rm.ShortMessage = "Subjekt-Verb-Kongruenz"
		if v == "ist" {
			rm.SetSuggestedReplacement("sind")
		} else {
			rm.SetSuggestedReplacement("waren")
		}
		matches = append(matches, rm)
	}
	return matches
}

func pluralSubjectBefore(tokens []*languagetool.AnalyzedTokenReadings, verbIdx int) bool {
	// scan back until clause-ish boundary
	sawPluralNoun := false
	sawUnd := false
	sawDie := false
	sawNumeral := false
	for j := verbIdx - 1; j >= 1; j-- {
		w := tokens[j].GetToken()
		lc := strings.ToLower(w)
		if w == "," || w == ";" || w == ":" || w == "." {
			break
		}
		if lc == "und" || lc == "sowie" {
			sawUnd = true
			continue
		}
		if lc == "die" || lc == "diese" || lc == "jene" || lc == "beide" || lc == "allen" {
			sawDie = true
			continue
		}
		if lc == "viele" || lc == "mehrere" || lc == "einige" || lc == "drei" || lc == "zwei" ||
			lc == "vier" || lc == "fünf" || isDigitToken(w) {
			sawNumeral = true
			continue
		}
		if looksPluralNounDE(w) {
			sawPluralNoun = true
		}
		// singular noun after Die often ends with e (Katze) — not pluralish
	}
	if sawUnd && sawPluralNoun {
		// "Hund und die Katze ist" — und between nouns
		return true
	}
	if sawDie && sawPluralNoun {
		return true
	}
	if sawNumeral && sawPluralNoun {
		return true
	}
	// "Drei Katzen ist"
	if sawNumeral {
		// noun right before verb?
		if verbIdx >= 2 && looksPluralNounDE(tokens[verbIdx-1].GetToken()) {
			return true
		}
	}
	return false
}

func isDigitToken(w string) bool {
	if w == "" {
		return false
	}
	for _, r := range w {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
