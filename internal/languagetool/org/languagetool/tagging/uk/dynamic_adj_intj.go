package uk

import (
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

// ukVowels for elongated collapse (Java ([аеєиіїоуюя])\1{2,} — RE2 has no backrefs).
const ukVowels = "аеєиіїоуюяАЕЄИІЇОУЮЯ"

// collapseElongatedVowels replaces any vowel repeated 3+ times with a single copy.
func collapseElongatedVowels(token string) (adjusted string, changed bool) {
	if token == "" {
		return token, false
	}
	var b strings.Builder
	rs := []rune(token)
	for i := 0; i < len(rs); {
		r := rs[i]
		b.WriteRune(r)
		if strings.ContainsRune(ukVowels, r) {
			j := i + 1
			for j < len(rs) && rs[j] == r {
				j++
			}
			if j-i >= 3 {
				changed = true
				// already wrote one; skip the rest of the run
				i = j
				continue
			}
		}
		i++
	}
	return b.String(), changed
}

// DynamicAdjReadings ports CompoundTagger for X-подібний / X-вмісний compounds.
// Java: right side from wordTagger (вмісн* via "боро"+right); no invent case endings.
// Fail-closed without tagWord hits.
func DynamicAdjReadings(token string, tagWord func(string) []tagging.TaggedWord) []struct{ Lemma, POS string } {
	if tagWord == nil || token == "" || !strings.Contains(token, "-") {
		return nil
	}
	dash := strings.LastIndex(token, "-")
	if dash <= 0 || dash == len(token)-1 {
		return nil
	}
	leftWord := token[:dash]
	rightWord := token[dash+1:]
	rightLow := strings.ToLower(rightWord)

	var tws []tagging.TaggedWord
	lemmaRight := ""
	switch {
	case strings.HasPrefix(rightLow, "вмісн"):
		// Java: Fe-вмісний → tag "боро"+right, lemma "вмісний"
		adjusted := "боро" + rightWord
		tws = tagWord(adjusted)
		if len(tws) == 0 {
			adjLow := "боро" + rightLow
			if adjLow != adjusted {
				tws = tagWord(adjLow)
			}
		}
		lemmaRight = "вмісний"
	case strings.HasPrefix(rightLow, "подібн"):
		// Java: tag right as-is (подібному ∈ dict) then generateTokensWithRighInflected
		tws = tagWord(rightWord)
		if len(tws) == 0 && rightWord != rightLow {
			tws = tagWord(rightLow)
		}
		// lemma comes from dict; fallback to lemma base
		lemmaRight = "подібний"
	default:
		return nil
	}
	if len(tws) == 0 {
		return nil
	}

	var out []struct{ Lemma, POS string }
	seen := map[string]struct{}{}
	for _, tw := range tws {
		pos := tw.PosTag
		if pos == "" || !strings.HasPrefix(pos, "adj") {
			continue
		}
		if strings.Contains(pos, "v_kly") {
			continue
		}
		// Java dropTag :comp.
		if i := strings.Index(pos, ":comp"); i >= 0 {
			end := i + len(":comp")
			for end < len(pos) && pos[end] != ':' {
				end++
			}
			pos = pos[:i] + pos[end:]
		}
		lem := lemmaRight
		if strings.HasPrefix(rightLow, "подібн") && tw.Lemma != "" {
			lem = tw.Lemma
		}
		// lemma = left + "-" + rightLemma (Java generateTokensWithRighInflected — keep left surface)
		fullLemma := leftWord + "-" + lem
		key := fullLemma + "|" + pos
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, struct{ Lemma, POS string }{Lemma: fullLemma, POS: pos})
	}
	return out
}

// ElongatedAltReadings ports UkrainianTagger elongated-vowel collapse:
// when surface has a vowel repeated 3+ times, re-tag the collapsed form and mark :alt.
// Fail closed without dictionary hits (Java getAdjustedAnalyzedTokens; no invent intj list).
// Skips "ііі" (often Latin number III) like Java.
func ElongatedAltReadings(token string, tagWord func(string) []tagging.TaggedWord) []*languagetool.AnalyzedToken {
	if tagWord == nil || token == "" {
		return nil
	}
	if strings.EqualFold(token, "ііі") {
		return nil
	}
	adjusted, changed := collapseElongatedVowels(token)
	if !changed || adjusted == "" {
		return nil
	}
	// Java posTagRegex: (?!noun.*:prop|.*abbr).*
	tws := tagWord(adjusted)
	if len(tws) == 0 {
		low := strings.ToLower(adjusted)
		if low != adjusted {
			tws = tagWord(low)
		}
	}
	if len(tws) == 0 {
		return nil
	}
	var out []*languagetool.AnalyzedToken
	for _, tw := range tws {
		pos := tw.PosTag
		if pos == "" {
			continue
		}
		// skip proper nouns and abbr (Java negative lookahead)
		if strings.Contains(pos, "noun") && strings.Contains(pos, "prop") {
			continue
		}
		if strings.Contains(pos, "abbr") {
			continue
		}
		if !strings.Contains(pos, ":alt") {
			pos = pos + ":alt"
		}
		lemma := tw.Lemma
		if lemma == "" {
			lemma = adjusted
		}
		p, l := pos, lemma
		out = append(out, languagetool.NewAnalyzedToken(token, &p, &l))
	}
	return out
}

func lowerFirst(s string) string {
	rs := []rune(s)
	if len(rs) == 0 {
		return s
	}
	rs[0] = unicode.ToLower(rs[0])
	return string(rs)
}
