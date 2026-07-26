package uk

import (
	"regexp"
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
	// Java tagEitherCase(rightWord) — as-is, then lower if empty (not BaseTagger case-merge)
	tws := tagEitherCase(right, tagWord)
	if len(tws) == 0 {
		return nil
	}
	// addTag: Upper right → :alt (пів-України); lower → :bad (пів-години)
	addTag := ":bad"
	if unicode.IsUpper([]rune(right)[0]) {
		addTag = ":alt"
	}
	// Java addPluralNvTokens:
	//   posTag = rightPosTag.replace("v_rod", vid).replaceFirst(":[mfn]:v_", ":p:v_") + ":nv" + addTag
	// Always appends :nv+addTag (no Contains checks → stable :nv:alt order).
	var out []*languagetool.AnalyzedToken
	seen := map[string]struct{}{}
	for _, tw := range tws {
		pos := tw.PosTag
		// Java NOUN_SING_V_ROD_REGEX: noun.*?:[mfn]:v_rod.*
		if pos == "" || !nounSingVRodRE.MatchString(pos) {
			continue
		}
		for _, vid := range nvCases {
			np := strings.Replace(pos, "v_rod", vid, 1)
			// replaceFirst ":[mfn]:v_" → ":p:v_"
			np = nounSingGenderVRE.ReplaceAllString(np, ":p:v_")
			np = np + ":nv" + addTag
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

// nounSingVRodRE ports CompoundTagger.NOUN_SING_V_ROD_REGEX.
var nounSingVRodRE = regexp.MustCompile(`^noun.*?:[mfn]:v_rod.*$`)

// nounSingGenderVRE ports replaceFirst(":[mfn]:v_", ":p:v_").
var nounSingGenderVRE = regexp.MustCompile(`:[mfn]:v_`)

// generateTokensForNv delegates to GenerateTokensForNv (PosTagHelper twin).
func generateTokensForNv(word, genders, extraTags string) []*languagetool.AnalyzedToken {
	return GenerateTokensForNv(word, genders, extraTags)
}
