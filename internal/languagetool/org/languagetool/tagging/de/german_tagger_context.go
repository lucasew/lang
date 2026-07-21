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

// innenPattern1: Java Pattern "in(nen)-[A-ZÖÄÜ][a-zöäüß-]+" — (nen) is required, not optional.
// matches() anchors whole string → e.g. innen-Zielgruppe
var reInnenPattern1 = regexp.MustCompile(`^innen-[A-ZÖÄÜ][a-zöäüß-]+$`)

// anythingDash: .*-  Java replaceFirst strips through last dash (greedy .*)
var reAnythingDash = regexp.MustCompile(`.*-`)

// innenPattern2: Java "innen[a-zöäüß-]+" e.g. innenzielgruppe
var reInnenPattern2 = regexp.MustCompile(`^innen[a-zöäüß-]+$`)

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
// charPos is Java's running `pos` (UTF-16 start offset of this token) — NOT token index.
func (t *GermanTagger) getImperativeForm(word string, sentenceTokens []string, charPos int) []*languagetool.AnalyzedToken {
	if t == nil || word == "" {
		return nil
	}
	// Java: int idx = sentenceTokens.indexOf(word); then walk back for previous non-ws
	idx := indexOfToken(sentenceTokens, word)
	prevWord := ""
	for j := idx - 1; j >= 0; j-- {
		// Java: !StringUtils.isWhitespace(previousWord)
		if !isJavaWhitespaceToken(sentenceTokens[j]) {
			prevWord = sentenceTokens[j]
			break
		}
	}
	// Java: pos == 0 && sentenceTokens.size() > 1  (char offset, not token index)
	atStart := charPos == 0 && len(sentenceTokens) > 1
	prevOK := equalsAnyIgnoreCase(prevWord, "ich", "er", "es", "sie", "bitte", "aber", "nun", "jetzt", "„")
	if !atStart && !prevOK {
		return nil
	}
	w := word
	if charPos == 0 || prevWord == "„" {
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

// isJavaWhitespaceToken ports StringUtils.isWhitespace (all chars whitespace, or empty).
func isJavaWhitespaceToken(s string) bool {
	if s == "" {
		return true
	}
	for _, r := range s {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return true
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

// genderGapTaggerTokens ports gender star / :innen token merging
// (jede*r, Werkstudent:innen-Zielgruppe, Werkstudent:innenzielgruppe).
// Returns non-nil (possibly empty) when a gender-gap branch matched — Java
// sets taggerTokens = new ArrayList<>(...) so later tag(word) is skipped.
// Returns nil when no gender-gap pattern applies (taggerTokens stays null).
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
		// "jede*r", "sein*e" — always allocate (Java new ArrayList)
		out := make([]tagging.TaggedWord, 0, 4)
		out = append(out, t.TagWordExact(word)...)
		out = append(out, t.TagWordExact(word+after)...)
		return out
	}
	if reInnenPattern1.MatchString(after) {
		// e.g. Werkstudent:innen-Zielgruppe → tags of 'Zielgruppe'
		// Java: anythingDash.matcher(after).replaceFirst("")
		lastPart := reAnythingDash.ReplaceAllString(after, "")
		// Copy into new slice so empty result is non-nil when tag misses.
		return append(make([]tagging.TaggedWord, 0, 1), t.TagWordExact(lastPart)...)
	}
	if reInnenPattern2.MatchString(after) {
		// e.g. Werkstudent:innenzielgruppe → uppercaseFirst(substring after last "innen")
		idx := strings.LastIndex(after, "innen")
		if idx < 0 {
			// pattern requires "innen"; keep Java-safe empty non-nil
			return make([]tagging.TaggedWord, 0)
		}
		// Java: substring(idx + "innen".length()) — "innen" is ASCII (5 UTF-16 units)
		rest := after[idx+len("innen"):]
		lastPart := tools.UppercaseFirstChar(rest)
		return append(make([]tagging.TaggedWord, 0, 1), t.TagWordExact(lastPart)...)
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
