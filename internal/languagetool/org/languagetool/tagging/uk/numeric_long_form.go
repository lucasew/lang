package uk

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

// RICCHA ports UkrainianTagger.RICCHA — …річч* long numeric compounds (стодвадцятиріччя).
// Groups: 1 optional hundreds, 2 optional tens-ish, 3 optional units-ish, 4 річч…
var ricchaRE = regexp.MustCompile(
	`(?i)^(сто|[а-яіїєґ']+?(?:сот))?([а-яіїє']+?(?:ти|ка|то))?([а-яіїє']+?(?:ти|ри|ох|ми|во))?(річч[а-яі]{1,3})$`,
)

// OTYI ports UkrainianTagger.OTYI — …мільйон*/тисяч*/річн* long forms.
var otyiRE = regexp.MustCompile(
	`(?i)^(сто|[а-яіїєґ']+?(?:сот))?([а-яіїє']+?(?:ти|ка|то))?([а-яіїє']+?(?:ти|ри|ох|ми|во|но))?((?:мільйон|тисяч|річн)[а-яії]+)$`,
)

// badOrSubstRE ports isAllNum exclude pattern .*(bad|subst).*
var badOrSubstRE = regexp.MustCompile(`(?s).*(?:bad|subst).*`)

// NumericLongFormReadings ports UkrainianTagger.additionalTags RICCHA + OTYI arms.
// Dict-gated via wordTagger; empty when groups lack num POS or end-word untagged.
func NumericLongFormReadings(word string, tagWord func(string) []tagging.TaggedWord) []*languagetool.AnalyzedToken {
	if tagWord == nil || tagging.UTF16Len(word) < 10 {
		return nil
	}
	if rs := ricchaReadings(word, tagWord); len(rs) > 0 {
		return rs
	}
	return otyiReadings(word, tagWord)
}

func ricchaReadings(word string, tagWord func(string) []tagging.TaggedWord) []*languagetool.AnalyzedToken {
	m := ricchaRE.FindStringSubmatch(word)
	if m == nil {
		return nil
	}
	// groups: full, 1, 2, 3, 4
	endWord := m[4]
	rightWdList := tagWord("сто" + endWord)
	if len(rightWdList) == 0 {
		return nil
	}
	if !isAllNumGroups(tagWord, m[1], m[2], m[3]) {
		return nil
	}
	prefix := m[1] + m[2] + m[3]
	var out []*languagetool.AnalyzedToken
	for _, tw := range rightWdList {
		pos := tw.PosTag
		if pos == "" || strings.Contains(pos, "v_kly") || strings.Contains(pos, ":p:") {
			continue
		}
		lemma := tw.Lemma
		// Java: concatGroups(1,3) + lemma.substring(3)  — strip "сто" from lemma
		if tagging.UTF16Len(lemma) >= 3 {
			// substring(3) is byte index for BMP-safe "сто" (3 runes = 6 bytes for UTF-8)
			// "сто" is 3 runes, 6 bytes in UTF-8
			if strings.HasPrefix(strings.ToLower(lemma), "сто") {
				// strip first 3 runes
				rs := []rune(lemma)
				if len(rs) >= 3 {
					lemma = prefix + string(rs[3:])
				} else {
					lemma = prefix + lemma
				}
			} else {
				// Java always substring(3) after dict hit on "сто"+end
				rs := []rune(lemma)
				if len(rs) >= 3 {
					lemma = prefix + string(rs[3:])
				} else {
					lemma = prefix + lemma
				}
			}
		} else {
			lemma = prefix + lemma
		}
		p, l := pos, lemma
		out = append(out, languagetool.NewAnalyzedToken(word, &p, &l))
	}
	return out
}

func otyiReadings(word string, tagWord func(string) []tagging.TaggedWord) []*languagetool.AnalyzedToken {
	m := otyiRE.FindStringSubmatch(word)
	if m == nil {
		return nil
	}
	endWord := m[4]
	rightWdList := tagWord(endWord)
	if len(rightWdList) == 0 {
		return nil
	}
	if !isAllNumGroups(tagWord, m[1], m[2], m[3]) {
		return nil
	}
	prefix := m[1] + m[2] + m[3]
	var out []*languagetool.AnalyzedToken
	for _, tw := range rightWdList {
		pos := tw.PosTag
		if pos == "" || !strings.HasPrefix(pos, "adj") || strings.Contains(pos, "v_kly") {
			continue
		}
		lemma := prefix + tw.Lemma
		p, l := pos, lemma
		out = append(out, languagetool.NewAnalyzedToken(word, &p, &l))
	}
	return out
}

// isAllNumGroups ports isAllNum for optional groups 1..3.
func isAllNumGroups(tagWord func(string) []tagging.TaggedWord, groups ...string) bool {
	for _, g := range groups {
		if g == "" {
			continue
		}
		w := tagWord(strings.ToLower(g))
		if !HasPosTagPart2(w, "num") {
			return false
		}
		if HasPosTag2(w, badOrSubstRE) {
			return false
		}
	}
	return true
}
