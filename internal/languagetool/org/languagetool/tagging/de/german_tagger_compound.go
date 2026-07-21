package de

import (
	"strings"
	"unicode"
	"unicode/utf16"

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
		if tagging.UTF16Len(stem) > 0 && javaLastUTF16RuneDE(stem) != '-' && strings.HasPrefix(tw.PosTag, "SUB") {
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
	// Java: !Character.isDigit(word.charAt(0))
	if tagging.UTF16Len(word) > 0 && unicode.IsDigit(javaFirstUTF16RuneDE(word)) {
		return nil
	}
	var readings []*languagetool.AnalyzedToken
	wordOrig := word
	sanitized := t.sanitizeWord(word)
	// Java: wordStem = wordOrig.substring(0, wordOrig.length() - word.length()) after sanitize
	wordStem := ""
	if tagging.UTF16Len(sanitized) <= tagging.UTF16Len(wordOrig) {
		wordStem = javaUTF16Prefix(wordOrig, tagging.UTF16Len(wordOrig)-tagging.UTF16Len(sanitized))
	}
	// Tokenize sanitized head (Java compoundedWord)
	compoundedWord := []string{sanitized}
	if t.SplitCompound != nil {
		if cp := t.SplitCompound(sanitized); len(cp) > 0 {
			compoundedWord = cp
		}
	}
	head := compoundedWord[len(compoundedWord)-1]
	if len(compoundedWord) > 1 {
		// Java always uppercases last when multi-part (unlike sanitizeWord which requires StartsWithUppercase)
		head = tools.UppercaseFirstChar(head)
	}
	linked := addStem(t.TagWordExact(head), wordStem)
	// dash + uppercase adj: retry lowercase — Java does NOT addStem on this branch
	if strings.Contains(wordOrig, "-") && len(linked) == 0 && t.matchesUppercaseAdjective(head) {
		lowHead := tools.LowercaseFirstChar(head)
		linked = t.TagWordExact(lowHead)
	}
	// Java: if (!linked.isEmpty()) { upper → simple tokens; lower → compound rebuild } else { prefix path }
	// When linked non-empty, never fall through to prefix path (even if all IMP filtered).
	if len(linked) > 0 {
		if tools.StartsWithUppercase(wordOrig) {
			// getAnalyzedTokens(linked, word) — no IMP filter
			for _, tw := range linked {
				readings = append(readings, toToken(wordOrig, tw))
			}
			return readings
		}
		// getAnalyzedTokens(linked, word, compoundedWord) — skip VER:IMP; rebuild lemma
		for _, tw := range linked {
			if tw.PosTag != "" && strings.HasPrefix(tw.PosTag, "VER:IMP") {
				continue
			}
			stem := ""
			if len(compoundedWord) > 1 {
				for i, p := range compoundedWord[:len(compoundedWord)-1] {
					if i == 0 {
						stem += p
					} else {
						stem += tools.LowercaseFirstChar(p)
					}
				}
			}
			lem := tw.Lemma
			if lem == "" {
				lem = head
			}
			lem = tools.LowercaseFirstChar(lem)
			readings = append(readings, toToken(wordOrig, tagging.NewTaggedWord(stem+lem, tw.PosTag)))
		}
		return readings
	}

	// Separable / general verb prefixes: einlädst → VER:…:NEB
	// Java firstPart = removeEnd(word, lastPart) with lastPart from lower removePattern
	low := strings.ToLower(wordOrig)
	if !startsWithAnyPrefix(low, prefixesVerbsLongestList) || containsNotAVerb(low) {
		return readings
	}
	if !isTitleOrLower(wordOrig) {
		return readings
	}
	lastPart, _ := stripLongestPrefix(low, prefixesVerbsLongestList)
	// Java: lastPart.length() > 2
	if tagging.UTF16Len(lastPart) <= 2 {
		return readings
	}
	// Java: firstPart = StringUtils.removeEnd(word, lastPart) — case-sensitive suffix strip
	firstPart := javaRemoveEnd(wordOrig, lastPart)
	// zu + infinitive → EIZ; lemma = firstPart + inf.lemma (NOT firstPart.toLowerCase)
	if strings.HasPrefix(lastPart, "zu") {
		infinitiv := strings.TrimPrefix(lastPart, "zu")
		for _, inf := range t.TagWordExact(infinitiv) {
			if strings.HasPrefix(inf.PosTag, "VER:INF") {
				// Java: RegExUtils.replaceFirst(pos, "INF", "EIZ")
				pstg := strings.Replace(inf.PosTag, "INF", "EIZ", 1)
				lemma := firstPart + inf.Lemma
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
		// Java: !firstPart.equals("un") — case-sensitive
		if firstPart == "un" {
			continue
		}
		lemmaBase := taggedWord.Lemma
		if lemmaBase == "" {
			lemmaBase = lastPart
		}
		fullLemma := strings.ToLower(firstPart) + lemmaBase
		if strings.HasPrefix(pos, "VER:INF") {
			// Java: word.equals(substring(0,1).toUpperCase()+substring(1)) — first upper, rest unchanged
			if isFirstCharUpperRestUnchanged(wordOrig) {
				readings = append(readings,
					toToken(wordOrig, tagging.NewTaggedWord(wordOrig, "SUB:NOM:SIN:NEU:INF")),
					toToken(wordOrig, tagging.NewTaggedWord(wordOrig, "SUB:DAT:SIN:NEU:INF")),
					toToken(wordOrig, tagging.NewTaggedWord(wordOrig, "SUB:AKK:SIN:NEU:INF")),
				)
				if indexOfToken(sentenceTokens, wordOrig) == 0 {
					readings = append(readings, toToken(wordOrig, tagging.NewTaggedWord(fullLemma, pos)))
				}
			} else {
				readings = append(readings, toToken(wordOrig, tagging.NewTaggedWord(fullLemma, pos)))
			}
			continue
		}
		fpLow := strings.ToLower(firstPart)
		// Java: word.equals(toLowerCase()) || sentenceTokens.indexOf(word) == 0
		atStartOrAllLower := wordOrig == strings.ToLower(wordOrig) || indexOfToken(sentenceTokens, wordOrig) == 0
		if isExactSeparablePrefix(fpLow) {
			if strings.HasPrefix(pos, "VER:IMP") {
				flekt := posTagLast3(pos)
				if atStartOrAllLower {
					if flekt == "SFT" || !wordMatchesIInfix(wordOrig) {
						// separable: lemma firstPart (not lower) + lemma
						readings = append(readings, toToken(wordOrig, tagging.NewTaggedWord(firstPart+lemmaBase, "VER:1:SIN:PRÄ:"+flekt+":NEB")))
					}
				}
				continue
			}
			if (strings.HasPrefix(pos, "VER:1") || strings.HasPrefix(pos, "VER:2") || strings.HasPrefix(pos, "VER:3")) && atStartOrAllLower {
				nebPos := pos
				if !strings.HasSuffix(nebPos, ":NEB") {
					nebPos = nebPos + ":NEB"
				}
				readings = append(readings, toToken(wordOrig, tagging.NewTaggedWord(fullLemma, nebPos)))
				// durch/um: Java firstPart.equals("durch"|"um") case-sensitive
				if (firstPart == "durch" || firstPart == "um") && strings.HasPrefix(pos, "VER:3:SIN:PRÄ") {
					readings = append(readings, toToken(wordOrig, tagging.NewTaggedWord(firstPart+lemmaBase, "VER:PA2:SFT")))
					readings = append(readings, toToken(wordOrig, tagging.NewTaggedWord(wordOrig, "PA2:PRD:GRU:VER")))
				}
				continue
			}
			if !strings.HasPrefix(pos, "VER:IMP") {
				readings = append(readings, toToken(wordOrig, tagging.NewTaggedWord(fullLemma, pos)))
			}
		} else if isExactNonSeparablePrefix(fpLow) && atStartOrAllLower {
			if strings.HasPrefix(pos, "VER:IMP") {
				flekt := posTagLast3(pos)
				readings = append(readings, toToken(wordOrig, tagging.NewTaggedWord(fullLemma, pos)))
				if strings.HasPrefix(pos, "VER:IMP:SIN") {
					readings = append(readings, toToken(wordOrig, tagging.NewTaggedWord(firstPart+lemmaBase, "VER:1:SIN:PRÄ:"+flekt)))
				}
			} else {
				readings = append(readings, toToken(wordOrig, tagging.NewTaggedWord(fullLemma, pos)))
			}
		}
	}
	// PA / VER:PA on lastPart tags (same loop conditions in Java as else-if of VER branch)
	if wordOrig == strings.ToLower(wordOrig) || indexOfToken(sentenceTokens, wordOrig) == 0 {
		for _, taggedWord := range t.TagWordExact(lastPart) {
			pos := taggedWord.PosTag
			if !(strings.HasPrefix(pos, "PA") || strings.HasPrefix(pos, "VER:PA")) {
				continue
			}
			if firstPart == "un" && strings.HasPrefix(pos, "VER:PA") {
				continue
			}
			lemmaBase := taggedWord.Lemma
			if lemmaBase == "" {
				lemmaBase = lastPart
			}
			if firstPart != "" {
				readings = append(readings, toToken(wordOrig, tagging.NewTaggedWord(strings.ToLower(firstPart)+lemmaBase, pos)))
			}
		}
	}
	// PA2 derivation for non-separable prefixes (erstickt, erstickter, …)
	readings = t.addPartizip2FromLastPart(wordOrig, firstPart, lastPart, idxPos, readings)
	return readings
}

// javaRemoveEnd ports StringUtils.removeEnd(str, remove): if str ends with remove, strip it.
func javaRemoveEnd(str, remove string) string {
	if remove == "" || !strings.HasSuffix(str, remove) {
		return str
	}
	return str[:len(str)-len(remove)]
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

// isTitleCaseWord ports word.equals(substring(0,1).toUpperCase()+substring(1).toLowerCase()) UTF-16.
func isTitleCaseWord(word string) bool {
	if word == "" {
		return false
	}
	return word == utf16FirstUpperRestLower(word)
}

// isFirstCharUpperRestUnchanged ports
// word.equals(substring(0,1).toUpperCase()+substring(1)) — UTF-16 (VER:INF nominalization gate).
func isFirstCharUpperRestUnchanged(word string) bool {
	if word == "" {
		return false
	}
	return word == utf16FirstUpperRest(word)
}

func utf16FirstUpperRest(word string) string {
	u := utf16.Encode([]rune(word))
	if len(u) == 0 {
		return word
	}
	first := string(utf16.Decode(u[:1]))
	rest := string(utf16.Decode(u[1:]))
	return strings.ToUpper(first) + rest
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

func javaLastUTF16RuneDE(s string) rune {
	u := utf16.Encode([]rune(s))
	if len(u) == 0 {
		return 0
	}
	return rune(u[len(u)-1])
}

func javaUTF16Prefix(s string, n int) string {
	u := utf16.Encode([]rune(s))
	if n <= 0 {
		return ""
	}
	if n > len(u) {
		n = len(u)
	}
	return string(utf16.Decode(u[:n]))
}

// javaUTF16Suffix ports Java substring(n): drop first n UTF-16 units.
func javaUTF16Suffix(s string, n int) string {
	u := utf16.Encode([]rune(s))
	if n <= 0 {
		return s
	}
	if n >= len(u) {
		return ""
	}
	return string(utf16.Decode(u[n:]))
}
