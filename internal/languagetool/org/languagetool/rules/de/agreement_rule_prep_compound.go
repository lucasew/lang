package de

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// ReplacementType ports AgreementRule.ReplacementType for contracted prepositions.
type ReplacementType int

const (
	ReplNone ReplacementType = iota
	ReplIns                  // ins, ans, aufs, … → treated as ART "das"
	ReplZur                  // zur → treated as ART "der"
)

// insPrepositions ports the ins/ans/… list in replacePrepositionsByArticle.
var insPrepositions = map[string]struct{}{
	"ins": {}, "ans": {}, "aufs": {}, "vors": {}, "durchs": {},
	"hinters": {}, "unters": {}, "übers": {}, "fürs": {}, "ums": {},
}

// replacePrepositionsByArticle ports AgreementRule.replacePrepositionsByArticle.
// Mutates tokens in place (like Java) and returns index → replacement type.
func replacePrepositionsByArticle(tokens []*languagetool.AnalyzedTokenReadings) map[int]ReplacementType {
	out := map[int]ReplacementType{}
	if tokens == nil {
		return out
	}
	for i, t := range tokens {
		if t == nil {
			continue
		}
		tok := t.GetToken()
		if _, ok := insPrepositions[tok]; ok {
			// Java INS_REPLACEMENT: AnalyzedToken("das", "ART:DEF:AKK:SIN:NEU", "das")
			pos := "ART:DEF:AKK:SIN:NEU"
			lem := "das"
			start := t.GetStartPos()
			rep := languagetool.NewAnalyzedTokenReadingsAt(
				languagetool.NewAnalyzedToken("das", &pos, &lem), start,
			)
			// Preserve immunization / length of original surface for span bookkeeping
			if t.IsImmunized() {
				rep.Immunize(0)
			}
			tokens[i] = rep
			out[i] = ReplIns
		} else if tok == "zur" {
			pos := "ART:DEF:DAT:SIN:FEM"
			lem := "der"
			start := t.GetStartPos()
			rep := languagetool.NewAnalyzedTokenReadingsAt(
				languagetool.NewAnalyzedToken("der", &pos, &lem), start,
			)
			if t.IsImmunized() {
				rep.Immunize(0)
			}
			tokens[i] = rep
			out[i] = ReplZur
		}
	}
	return out
}

const compoundErrorMsg = "Wenn es sich um ein zusammengesetztes Nomen handelt, wird es zusammengeschrieben."

// getCompoundErrorDetNoun ports getCompoundError(token1, token2, tokenPos, sentence)
// for DET + NOUN + Capitalized. Java only emits matches when lt.check(phrase).isEmpty()
// (DE_AGREEMENT + GERMAN_SPELLER); CompoundPhraseValid is that gate — nil → fail-closed.
func getCompoundErrorDetNoun(det, noun *languagetool.AnalyzedTokenReadings, tokenPos int,
	origTokens []*languagetool.AnalyzedTokenReadings, sentence *languagetool.AnalyzedSentence, rule any) *rules.RuleMatch {
	if tokenPos < 0 || origTokens == nil || tokenPos+2 >= len(origTokens) {
		return nil
	}
	next := origTokens[tokenPos+2]
	if next == nil || !tools.StartsWithUppercase(next.GetToken()) {
		return nil
	}
	if noun != nil && noun.GetStartPos() == next.GetStartPos() {
		return nil
	}
	// Java: all readings must allow SUB (skip if none start with SUB)
	if !nextIsPossibleNoun(next) {
		return nil
	}
	origDet := origTokens[tokenPos].GetToken()
	closed := noun.GetToken() + tools.LowercaseFirstChar(next.GetToken())
	hyphen := noun.GetToken() + "-" + next.GetToken()
	testPhrase := origDet + " " + closed
	hyphenPhrase := origDet + " " + hyphen
	return newCompoundRuleMatch(rule, sentence, det, next, testPhrase, hyphenPhrase)
}

// getCompoundErrorDetAdjNoun ports 3-token compound error (DET ADJ NOUN Capitalized).
func getCompoundErrorDetAdjNoun(det, adj, noun *languagetool.AnalyzedTokenReadings, tokenPos int,
	origTokens []*languagetool.AnalyzedTokenReadings, sentence *languagetool.AnalyzedSentence, rule any) *rules.RuleMatch {
	if tokenPos < 0 || origTokens == nil || tokenPos+3 >= len(origTokens) {
		return nil
	}
	next := origTokens[tokenPos+3]
	if next == nil || !tools.StartsWithUppercase(next.GetToken()) {
		return nil
	}
	if noun != nil && noun.GetStartPos() == next.GetStartPos() {
		return nil
	}
	if !nextIsPossibleNoun(next) {
		return nil
	}
	origDet := origTokens[tokenPos].GetToken()
	closed := noun.GetToken() + tools.LowercaseFirstChar(next.GetToken())
	hyphen := noun.GetToken() + "-" + next.GetToken()
	testPhrase := origDet + " " + adj.GetToken() + " " + closed
	hyphenPhrase := origDet + " " + adj.GetToken() + " " + hyphen
	return newCompoundRuleMatch(rule, sentence, det, next, testPhrase, hyphenPhrase)
}

// getCompoundErrorDetAdjAdjNoun ports 4-token path (optional skippedStr for modifiers).
func getCompoundErrorDetAdjAdjNoun(det, adj1, adj2, noun *languagetool.AnalyzedTokenReadings, tokenPos int,
	skippedStr string, origTokens []*languagetool.AnalyzedTokenReadings, sentence *languagetool.AnalyzedSentence, rule any) *rules.RuleMatch {
	extra := 0
	if skippedStr != "" {
		extra = 1
	}
	idx := tokenPos + 4 + extra
	if tokenPos < 0 || origTokens == nil || idx >= len(origTokens) {
		return nil
	}
	next := origTokens[idx]
	if next == nil || noun == nil {
		return nil
	}
	if !tools.StartsWithUppercase(noun.GetToken()) || !tools.StartsWithUppercase(next.GetToken()) {
		return nil
	}
	if noun.GetStartPos() == next.GetStartPos() {
		return nil
	}
	if !nextIsPossibleNoun(next) {
		return nil
	}
	origDet := origTokens[tokenPos].GetToken()
	mid := adj1.GetToken() + " " + adj2.GetToken()
	if skippedStr != "" {
		mid = skippedStr + " " + mid
	}
	closed := noun.GetToken() + tools.LowercaseFirstChar(next.GetToken())
	hyphen := noun.GetToken() + "-" + next.GetToken()
	testPhrase := origDet + " " + mid + " " + closed
	hyphenPhrase := origDet + " " + mid + " " + hyphen
	return newCompoundRuleMatch(rule, sentence, det, next, testPhrase, hyphenPhrase)
}

func nextIsPossibleNoun(next *languagetool.AnalyzedTokenReadings) bool {
	if next == nil {
		return false
	}
	// Java: if all readings have POS and none start with SUB → null
	// If untagged (soft), still allow (Java requires isTagged for adding replacements)
	if !next.IsTagged() {
		return false
	}
	hasNonSub := false
	hasSub := false
	for _, rd := range next.GetReadings() {
		if rd == nil || rd.GetPOSTag() == nil {
			continue
		}
		p := *rd.GetPOSTag()
		if strings.HasPrefix(p, "SUB") {
			hasSub = true
		} else {
			hasNonSub = true
		}
	}
	// Java: allMatch(!startsWith SUB) → return null. So if every tagged reading is non-SUB, skip.
	if hasNonSub && !hasSub {
		return false
	}
	return hasSub || next.IsTagged()
}

func newCompoundRuleMatch(rule any, sentence *languagetool.AnalyzedSentence,
	from, to *languagetool.AnalyzedTokenReadings, testPhrase, hyphenPhrase string) *rules.RuleMatch {
	if from == nil || to == nil {
		return nil
	}
	// Java getRuleMatch: token2 must be tagged for each replacement branch.
	if !to.IsTagged() {
		return nil
	}
	var valid func(string) bool
	if ar, ok := rule.(*AgreementRule); ok && ar != nil {
		valid = ar.CompoundPhraseValid
	}
	// Java always runs lt.check before adding suggestions. Without CompoundPhraseValid
	// we cannot invent that the closed/hyphen form is grammatical or spelled correctly.
	if valid == nil {
		return nil
	}
	var reps []string
	for _, p := range []string{testPhrase, hyphenPhrase} {
		if p == "" {
			continue
		}
		if !valid(p) {
			continue
		}
		reps = append(reps, p)
	}
	if len(reps) == 0 {
		return nil
	}
	// Java: RuleMatch without shortMessage; setUrl to leo Nomen grammar anchor.
	rm := rules.NewRuleMatch(rule, sentence, from.GetStartPos(), to.GetEndPos(), compoundErrorMsg)
	rm.SetSuggestedReplacements(reps)
	rm.SetURL("https://dict.leo.org/grammatik/deutsch/Rechtschreibung/Regeln/Getrennt-zusammen/Nomen.html#grammarAnchor-Nomen-49575")
	return rm
}
