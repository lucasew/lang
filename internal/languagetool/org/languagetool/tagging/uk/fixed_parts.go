package uk

import (
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

// cityAvenu ports LemmaHelper.CITY_AVENU (Java CompoundTagger street/avenue right parts).
// All generateTokensForNv with gender "f" + :prop.
var cityAvenu = map[string]struct{}{
	"сіті": {}, "ситі": {}, "стріт": {}, "стрит": {},
	"рівер": {}, "ривер": {}, "авеню": {},
	"штрасе": {}, "штрассе": {}, "сьоркл": {}, "сквер": {}, "плац": {},
}

// nvCases ports PosTagHelper.VIDMINKY_MAP without v_kly (Java generateTokensForNv).
var nvCases = []string{"v_naz", "v_rod", "v_dav", "v_zna", "v_oru", "v_mis"}

// FixedPartReadings ports CompoundTagger city-avenue and пів- compounds.
// Street suffixes use official CITY_AVENU list. пів- needs dictionary readings
// on the right (Java tagEitherCase + addPluralNvTokens) — fail closed without tags.
func FixedPartReadings(token string, tagWord func(string) []tagging.TaggedWord) []*languagetool.AnalyzedToken {
	t := strings.ReplaceAll(token, "–", "-")
	t = strings.ReplaceAll(t, "—", "-")
	rs := []rune(t)
	if len(rs) < 5 {
		return nil
	}

	// Пенсильванія-авеню / Уолл-стрит (Java CITY_AVENU + generateTokensForNv f :prop)
	if i := strings.LastIndex(t, "-"); i > 0 {
		left, right := t[:i], t[i+1:]
		if left != "" && unicode.IsUpper([]rune(left)[0]) {
			lowR := strings.ToLower(right)
			if _, ok := cityAvenu[lowR]; ok {
				addPos := ""
				if lowR == "штрассе" {
					addPos = ":alt"
				}
				return generateTokensForNv(token, "f", ":prop"+addPos)
			}
		}
	}

	// пів-… (Java left "пів" / startsWith "пів-")
	if !strings.HasPrefix(strings.ToLower(t), "пів-") {
		return nil
	}
	idx := strings.Index(t, "-")
	if idx <= 0 || idx+1 >= len(t) {
		return nil
	}
	right := t[idx+1:]
	if right == "" || tagWord == nil {
		return nil
	}
	// Java tagEitherCase(rightWord)
	tws := tagWord(right)
	if len(tws) == 0 {
		low := strings.ToLower(right)
		if low != right {
			tws = tagWord(low)
		}
	}
	if len(tws) == 0 {
		return nil
	}
	// addTag: Upper right → :alt (пів-України); lower → :bad (пів-години)
	addTag := ":bad"
	if unicode.IsUpper([]rune(right)[0]) {
		addTag = ":alt"
	}
	// Java addPluralNvTokens: only noun sing v_rod readings
	var out []*languagetool.AnalyzedToken
	seen := map[string]struct{}{}
	for _, tw := range tws {
		pos := tw.PosTag
		if pos == "" || !strings.HasPrefix(pos, "noun") {
			continue
		}
		if !strings.Contains(pos, "v_rod") {
			continue
		}
		// must be singular :m: / :f: / :n: before v_rod
		if strings.Contains(pos, ":p:") {
			continue
		}
		for _, vid := range nvCases {
			// replace v_rod with vid and force plural :p:
			np := pos
			np = strings.Replace(np, "v_rod", vid, 1)
			// :m:v_ / :f:v_ / :n:v_ → :p:v_
			for _, g := range []string{":m:v_", ":f:v_", ":n:v_"} {
				if strings.Contains(np, g) {
					np = strings.Replace(np, g, ":p:v_", 1)
					break
				}
			}
			if !strings.Contains(np, ":nv") {
				np = np + ":nv"
			}
			if addTag != "" && !strings.Contains(np, addTag) {
				np = np + addTag
			}
			if _, ok := seen[np]; ok {
				continue
			}
			seen[np] = struct{}{}
			p, l := np, token
			out = append(out, languagetool.NewAnalyzedToken(token, &p, &l))
		}
	}
	return out
}

// generateTokensForNv ports PosTagHelper.generateTokensForNv (lemma = surface).
func generateTokensForNv(word, genders, extraTags string) []*languagetool.AnalyzedToken {
	var out []*languagetool.AnalyzedToken
	for _, gen := range genders {
		for _, vidm := range nvCases {
			pos := "noun:inanim:" + string(gen) + ":" + vidm + ":nv"
			if extraTags != "" {
				pos += extraTags
			}
			p, l := pos, word
			out = append(out, languagetool.NewAnalyzedToken(word, &p, &l))
		}
	}
	return out
}
