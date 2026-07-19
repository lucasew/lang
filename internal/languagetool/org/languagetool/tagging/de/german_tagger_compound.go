package de

import (
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// matchesUppercaseAdjective ports GermanTagger.matchesUppercaseAdjective:
// lowercaseFirst has ADJ reading.
func (t *GermanTagger) matchesUppercaseAdjective(unknownUppercaseToken string) bool {
	if t == nil || unknownUppercaseToken == "" {
		return false
	}
	temp := t.TagWordExact(tools.LowercaseFirstChar(unknownUppercaseToken))
	return len(temp) > 0 && strings.HasPrefix(temp[0].PosTag, "ADJ")
}

// sanitizeWord ports GermanTagger.sanitizeWord: for dash compounds, return the
// last noun/adj part when compound-tokenizable; else original word.
func (t *GermanTagger) sanitizeWord(word string) string {
	if t == nil || word == "" || strings.HasSuffix(word, "-") {
		return word
	}
	result := word
	parts := strings.Split(word, "-")
	lastPart := word
	if len(parts) > 1 && strings.TrimSpace(parts[len(parts)-1]) != "" {
		lastPart = parts[len(parts)-1]
	}
	// compound tokenize last segment when SplitCompound available
	compounded := []string{lastPart}
	if t.SplitCompound != nil {
		if cp := t.SplitCompound(lastPart); len(cp) > 0 {
			compounded = cp
		}
	}
	if len(compounded) > 1 && tools.StartsWithUppercase(word) {
		lastPart = tools.UppercaseFirstChar(compounded[len(compounded)-1])
	} else {
		lastPart = compounded[len(compounded)-1]
	}
	tagged := t.TagWordExact(lastPart)
	if len(tagged) > 0 {
		pos := tagged[0].PosTag
		if strings.HasPrefix(pos, "SUB") || strings.HasPrefix(pos, "ADJ") || t.matchesUppercaseAdjective(lastPart) {
			result = lastPart
		}
	}
	return result
}

// addStem ports GermanTagger.addStem: prefix stem onto each lemma (lowercase
// lemma for SUB when stem does not end with '-').
func addStem(analyzed []tagging.TaggedWord, stem string) []tagging.TaggedWord {
	if len(analyzed) == 0 {
		return nil
	}
	out := make([]tagging.TaggedWord, 0, len(analyzed))
	for _, tw := range analyzed {
		lemma := tw.Lemma
		if len(stem) > 0 && stem[len(stem)-1] != '-' && strings.HasPrefix(tw.PosTag, "SUB") {
			lemma = strings.ToLower(lemma)
		}
		out = append(out, tagging.NewTaggedWord(stem+lemma, tw.PosTag))
	}
	return out
}

// tagUnknownDashAndPrefix ports the dash-linked + separable-prefix unknown-word
// branch after elative (sanitizeWord, addStem, prefixesVerbs NEB/EIZ).
func (t *GermanTagger) tagUnknownDashAndPrefix(word string, sentenceTokens []string, idxPos int) []*languagetool.AnalyzedToken {
	if t == nil || word == "" || strings.Contains(word, " ") {
		return nil
	}
	for _, r0 := range word {
		if unicode.IsDigit(r0) {
			return nil
		}
		break
	}
	var readings []*languagetool.AnalyzedToken
	wordOrig := word
	sanitized := t.sanitizeWord(word)
	wordStem := ""
	if len(sanitized) < len(wordOrig) && strings.HasSuffix(wordOrig, sanitized) {
		wordStem = wordOrig[:len(wordOrig)-len(sanitized)]
	} else if sanitized != wordOrig && strings.Contains(wordOrig, "-") {
		// stem is everything before last dash part used as sanitized
		if i := strings.LastIndex(wordOrig, sanitized); i > 0 {
			wordStem = wordOrig[:i]
		}
	}
	// compound tokenize sanitized head
	head := sanitized
	if t.SplitCompound != nil {
		if cp := t.SplitCompound(head); len(cp) > 1 {
			head = tools.UppercaseFirstChar(cp[len(cp)-1])
		}
	}
	linked := addStem(t.TagWordExact(head), wordStem)
	// dash + uppercase adj: retry lowercase
	if wordOrig != "" && strings.Contains(wordOrig, "-") && len(linked) == 0 && t.matchesUppercaseAdjective(head) {
		lowHead := tools.LowercaseFirstChar(head)
		linked = t.TagWordExact(lowHead)
		// no addStem on this Java branch for empty linked retry — just re-tag lower
		if len(linked) > 0 && wordStem != "" {
			linked = addStem(linked, wordStem)
		}
	}
	if len(linked) > 0 {
		for _, tw := range linked {
			if strings.HasPrefix(tw.PosTag, "VER:IMP") {
				continue // compound path skips IMP
			}
			readings = append(readings, toToken(wordOrig, tw))
		}
		if len(readings) > 0 {
			return readings
		}
	}

	// Separable / general verb prefixes: einlädst → VER:…:NEB
	low := strings.ToLower(wordOrig)
	if !startsWithAnyPrefix(low, prefixesVerbsLongestList) || containsNotAVerb(low) {
		return readings
	}
	if !isTitleOrLower(wordOrig) {
		return readings
	}
	lastPart, firstPart := stripLongestPrefix(low, prefixesVerbsLongestList)
	if len(lastPart) <= 2 {
		return readings
	}
	// recover firstPart casing from original surface
	if len(wordOrig) >= len(firstPart) {
		// firstPart is lower; map length
		firstPart = wordOrig[:len(firstPart)]
	}
	// zu + infinitive → EIZ
	if strings.HasPrefix(lastPart, "zu") {
		infinitiv := strings.TrimPrefix(lastPart, "zu")
		for _, inf := range t.TagWordExact(infinitiv) {
			if strings.HasPrefix(inf.PosTag, "VER:INF") {
				pstg := strings.Replace(inf.PosTag, "INF", "EIZ", 1)
				lemma := strings.ToLower(firstPart) + inf.Lemma
				readings = append(readings, toToken(wordOrig, tagging.NewTaggedWord(lemma, pstg)))
			}
		}
	}
	for _, taggedWord := range t.TagWordExact(lastPart) {
		pos := taggedWord.PosTag
		if !strings.HasPrefix(pos, "VER") || strings.HasPrefix(pos, "VER:PA") ||
			strings.HasPrefix(pos, "VER:AUX") || strings.HasPrefix(pos, "VER:MOD") {
			continue
		}
		if strings.EqualFold(firstPart, "un") {
			continue
		}
		lemmaBase := taggedWord.Lemma
		if lemmaBase == "" {
			lemmaBase = lastPart
		}
		fullLemma := strings.ToLower(firstPart) + lemmaBase
		if strings.HasPrefix(pos, "VER:INF") {
			// Title case infinitive → also nominalized SUB:…:INF
			if isTitleCaseWord(wordOrig) {
				readings = append(readings,
					toToken(wordOrig, tagging.NewTaggedWord(wordOrig, "SUB:NOM:SIN:NEU:INF")),
					toToken(wordOrig, tagging.NewTaggedWord(wordOrig, "SUB:DAT:SIN:NEU:INF")),
					toToken(wordOrig, tagging.NewTaggedWord(wordOrig, "SUB:AKK:SIN:NEU:INF")),
				)
				if idxPos == 0 {
					readings = append(readings, toToken(wordOrig, tagging.NewTaggedWord(fullLemma, pos)))
				}
			} else {
				readings = append(readings, toToken(wordOrig, tagging.NewTaggedWord(fullLemma, pos)))
			}
			continue
		}
		// finite / other VER: separable → append :NEB when conjugations 1/2/3 or general
		fpLow := strings.ToLower(firstPart)
		if startsWithAnyPrefix(fpLow, prefixesSeparableVerbsLongestList) || isExactSeparablePrefix(fpLow) {
			if strings.HasPrefix(pos, "VER:1") || strings.HasPrefix(pos, "VER:2") || strings.HasPrefix(pos, "VER:3") {
				if idxPos == 0 || wordOrig == strings.ToLower(wordOrig) || isTitleCaseWord(wordOrig) {
					nebPos := pos
					if !strings.HasSuffix(nebPos, ":NEB") {
						nebPos = nebPos + ":NEB"
					}
					readings = append(readings, toToken(wordOrig, tagging.NewTaggedWord(fullLemma, nebPos)))
					// durch/um: also non-separable PA2 readings (Java dual separable/non-sep)
					if (fpLow == "durch" || fpLow == "um") && strings.HasPrefix(pos, "VER:3:SIN:PRÄ") {
						if strings.HasPrefix(pos, "VER:3:SIN:PRÄ") {
							// Java: VER:3:SIN:PRÄ → VER:PA2:SFT (inner if always true when outer matches)
							readings = append(readings, toToken(wordOrig, tagging.NewTaggedWord(firstPart+lemmaBase, "VER:PA2:SFT")))
						} else {
							readings = append(readings, toToken(wordOrig, tagging.NewTaggedWord(firstPart+lemmaBase, "VER:PA2:NON")))
						}
						readings = append(readings, toToken(wordOrig, tagging.NewTaggedWord(wordOrig, "PA2:PRD:GRU:VER")))
					}
					continue
				}
			}
			if !strings.HasPrefix(pos, "VER:IMP") {
				readings = append(readings, toToken(wordOrig, tagging.NewTaggedWord(fullLemma, pos)))
			}
		} else {
			// non-separable prefix verb forms
			readings = append(readings, toToken(wordOrig, tagging.NewTaggedWord(fullLemma, pos)))
		}
	}
	// PA2 derivation for non-separable prefixes (erstickt, erstickter, …)
	readings = t.addPartizip2FromLastPart(wordOrig, firstPart, lastPart, idxPos, readings)
	return readings
}

func containsNotAVerb(wordLower string) bool {
	for n := range notAVerb {
		if strings.Contains(wordLower, n) {
			return true
		}
	}
	return false
}

func stripLongestPrefix(wordLower string, prefixes []string) (last, first string) {
	for _, p := range prefixes {
		if strings.HasPrefix(wordLower, p) && len(wordLower) > len(p) {
			return wordLower[len(p):], p
		}
	}
	return wordLower, ""
}

func isTitleCaseWord(word string) bool {
	rs := []rune(word)
	if len(rs) == 0 || !unicode.IsUpper(rs[0]) {
		return false
	}
	for _, r := range rs[1:] {
		if unicode.IsLetter(r) && !unicode.IsLower(r) {
			return false
		}
	}
	return true
}

// isDomainLikeSequence ports GermanTagger domain skip: word . com|net|org|…
func isDomainLikeSequence(sentenceTokens []string, idxPos int) bool {
	if idxPos+2 >= len(sentenceTokens) {
		return false
	}
	if sentenceTokens[idxPos+1] != "." {
		return false
	}
	_, ok := domainTLDs[strings.ToLower(sentenceTokens[idxPos+2])]
	return ok
}
