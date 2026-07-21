package de

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// mitarbeitendenPattern: [A-ZÖÄÜ][a-zöäüß]{2,25}mitarbeitenden?
var reMitarbeitenden = regexp.MustCompile(`^[A-ZÖÄÜ][a-zöäüß]{2,25}mitarbeitenden?$`)

// DDD_ER_PATTERN: \d{4}+er  (Java \d{4}+er)
var reDDDEr = regexp.MustCompile(`^\d{4,}er$`)

// genderGapChars: [*:_/]
var reGenderGapChar = regexp.MustCompile(`^[\*:_/]$`)

// afterAsterisk: in(nen)?|r|e
var reAfterAsterisk = regexp.MustCompile(`^(?:in(?:nen)?|r|e)$`)

// allAdjGruTags ports GermanTagger.allAdjGruTags (NOM/AKK/GEN/DAT × PLU/SIN × MAS/FEM/NEU × DEF/IND/SOL).
var allAdjGruTags []string

func init() {
	for _, case_ := range []string{"NOM", "AKK", "GEN", "DAT"} {
		for _, num := range []string{"PLU", "SIN"} {
			for _, gen := range []string{"MAS", "FEM", "NEU"} {
				for _, art := range []string{"DEF", "IND", "SOL"} {
					allAdjGruTags = append(allAdjGruTags, "ADJ:"+case_+":"+num+":"+gen+":GRU:"+art)
				}
			}
		}
	}
}

// getImperativeForm ports GermanTagger.getImperativeForm:
// short form "Geh" by tagging "gehe" for VER:IMP:SIN when at sentence start
// or after ich/er/es/sie/bitte/aber/nun/jetzt/„.
func (t *GermanTagger) getImperativeForm(word string, sentenceTokens []string, pos int) []*languagetool.AnalyzedToken {
	if t == nil || word == "" {
		return nil
	}
	// previous non-whitespace token
	idx := -1
	for i, tok := range sentenceTokens {
		if tok == word {
			// use first occurrence matching index if multiple — prefer pos if tokens are unique path
			idx = i
			// if multiple same word, match by walking with pos... use first for simplicity
			break
		}
	}
	// Prefer index == pos when sentenceTokens[pos]==word
	if pos >= 0 && pos < len(sentenceTokens) && sentenceTokens[pos] == word {
		idx = pos
	}
	prevWord := ""
	for j := idx - 1; j >= 0; j-- {
		if strings.TrimSpace(sentenceTokens[j]) != "" {
			prevWord = sentenceTokens[j]
			break
		}
	}
	atStart := pos == 0 && len(sentenceTokens) > 1
	prevOK := equalsAnyIgnoreCase(prevWord, "ich", "er", "es", "sie", "bitte", "aber", "nun", "jetzt", "„")
	if !atStart && !prevOK {
		return nil
	}
	w := word
	if pos == 0 || prevWord == "„" {
		w = strings.ToLower(word)
	}
	// tag w+"e" for VER:IMP:SIN
	for _, tagged := range t.TagWordExact(w + "e") {
		if !strings.HasPrefix(tagged.PosTag, "VER:IMP:SIN") {
			continue
		}
		// Java: do not overwrite manually removed tags
		// if (removalTagger == null || !removalTagger.tag(w).contains(tagged))
		if t.RemovalTagger != nil {
			removed := false
			for _, r := range t.RemovalTagger.Tag(w) {
				if r.Equal(tagged) {
					removed = true
					break
				}
			}
			if removed {
				break // Java: break after blocked match
			}
		}
		return []*languagetool.AnalyzedToken{toToken(word, tagged)}
	}
	return nil
}

// getSubstantivatedForms ports GermanTagger.getSubstantivatedForms (…er → SUB:…:ADJ / 2019er ADJ).
func (t *GermanTagger) getSubstantivatedForms(word string, sentenceTokens []string) []*languagetool.AnalyzedToken {
	if t == nil || !strings.HasSuffix(word, "er") {
		return nil
	}
	if reDDDEr.MatchString(word) {
		// e.g. "2019er"
		var list []*languagetool.AnalyzedToken
		for _, tag := range allAdjGruTags {
			list = append(list, toToken(word, tagging.NewTaggedWord(word, tag)))
		}
		return list
	}
	// do not tag if lowercase is ADV (e.g. Früher)
	for _, tw := range t.TagWordExact(strings.ToLower(word)) {
		if strings.HasPrefix(tw.PosTag, "ADV") {
			return nil
		}
	}
	// followed by uppercase or "als"? then not substantivated
	idx := indexOfToken(sentenceTokens, word)
	for j := idx + 1; j < len(sentenceTokens); j++ {
		next := sentenceTokens[j]
		if strings.TrimSpace(next) == "" {
			continue
		}
		// Java: nextWord.length() > 0 && (Character.isUpperCase(nextWord.charAt(0)) || "als")
		if tagging.UTF16Len(next) > 0 {
			if unicode.IsUpper(javaFirstUTF16RuneDE(next)) || next == "als" {
				return nil
			}
		}
		break
	}
	// Java: word.substring(0, word.length()-1) — drop last UTF-16 unit ("…er" → "…e")
	female := javaDropLastUTF16(word)
	isSub := false
	for _, tw := range t.TagWordExact(female) {
		if tw.PosTag == "SUB:NOM:SIN:FEM:ADJ" {
			isSub = true
			break
		}
	}
	if !isSub {
		return nil
	}
	return []*languagetool.AnalyzedToken{
		toToken(word, tagging.NewTaggedWord(word, "SUB:NOM:SIN:MAS:ADJ")),
		toToken(word, tagging.NewTaggedWord(word, "SUB:GEN:PLU:MAS:ADJ")),
	}
}

// tagMitarbeitenden ports mitarbeitendenPattern branch.
func (t *GermanTagger) tagMitarbeitenden(word string) []*languagetool.AnalyzedToken {
	if t == nil || !reMitarbeitenden.MatchString(word) {
		return nil
	}
	idx := strings.Index(word, "mitarbeitende")
	if idx < 0 {
		return nil
	}
	firstPart := word[:idx]
	lastPart := word[idx:]
	var readings []*languagetool.AnalyzedToken
	for _, tw := range t.TagWordExact(tools.UppercaseFirstChar(lastPart)) {
		lemma := firstPart + "mitarbeitende"
		readings = append(readings, toToken(word, tagging.NewTaggedWord(lemma, tw.PosTag)))
	}
	return readings
}

// genderGapTaggerTokens ports gender star token merging (jede*r etc.).
func (t *GermanTagger) genderGapTaggerTokens(sentenceTokens []string, idxPos int, word string) []tagging.TaggedWord {
	if t == nil || idxPos+2 >= len(sentenceTokens) {
		return nil
	}
	mid := sentenceTokens[idxPos+1]
	after := sentenceTokens[idxPos+2]
	if !reGenderGapChar.MatchString(mid) {
		return nil
	}
	if reAfterAsterisk.MatchString(after) {
		// "jede*r", "sein*e"
		var out []tagging.TaggedWord
		out = append(out, t.TagWordExact(word)...)
		out = append(out, t.TagWordExact(word+after)...)
		return out
	}
	return nil
}

func equalsAnyIgnoreCase(s string, opts ...string) bool {
	for _, o := range opts {
		if strings.EqualFold(s, o) {
			return true
		}
	}
	return false
}

func indexOfToken(tokens []string, word string) int {
	for i, t := range tokens {
		if t == word {
			return i
		}
	}
	return -1
}

func javaFirstUTF16RuneDE(s string) rune {
	u := utf16.Encode([]rune(s))
	if len(u) == 0 {
		return 0
	}
	return rune(u[0])
}

func javaDropLastUTF16(s string) string {
	u := utf16.Encode([]rune(s))
	if len(u) == 0 {
		return s
	}
	return string(utf16.Decode(u[:len(u)-1]))
}
