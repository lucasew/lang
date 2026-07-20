package fr

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// Static lists from MorfologikFrenchSpellerRule (Java).

var frenchTokenAtStart = func() map[string]struct{} {
	words := []string{
		"non", "en", "a", "le", "la", "les", "pour", "de", "du", "des", "un", "une",
		"mon", "ma", "mes", "ton", "ta", "tes", "son", "sa", "ses", "leur", "leurs",
		"ce", "cet",
	}
	m := make(map[string]struct{}, len(words))
	for _, w := range words {
		m[w] = struct{}{}
	}
	return m
}()

var frenchPrefixWithWhitespace = func() map[string]struct{} {
	words := []string{
		"agro", "anti", "archi", "auto", "aéro", "cardio", "co", "cyber", "demi", "ex",
		"extra", "géo", "hospitalo", "hydro", "hyper", "hypo", "infra", "inter", "macro",
		"mega", "meta", "mi", "micro", "mini", "mono", "multi", "musculo", "méga", "méta",
		"néo", "omni", "pan", "para", "pluri", "poly", "post", "prim", "pro", "proto",
		"pré", "pseudo", "psycho", "péri", "re", "retro", "ré", "semi", "simili", "socio",
		"super", "supra", "sus", "trans", "tri", "télé", "ultra", "uni", "vice", "éco",
	}
	m := make(map[string]struct{}, len(words))
	for _, w := range words {
		m[w] = struct{}{}
	}
	return m
}()

var frenchExceptionsEgrave = map[string]struct{}{
	"burkinabè": {}, "koinè": {}, "épistémè": {},
}

var hyphenOrQuote = regexp.MustCompile(`['\-]`)

// orderFrenchSuggestions ports MorfologikFrenchSpellerRule.orderSuggestions (string form).
func orderFrenchSuggestions(suggestions []string, word string) []string {
	wordWithoutDiacritics := tools.RemoveDiacritics(word)
	var out []string
	n := len(suggestions)
	for i, sug := range suggestions {
		low := strings.ToLower(sug)
		parts := strings.Split(low, " ")
		if len(parts) == 0 || parts[0] == "" {
			continue
		}
		// remove wrong split prefixes
		if len(parts) == 2 {
			if _, bad := frenchPrefixWithWhitespace[parts[0]]; bad {
				continue
			}
		}
		// single-letter first token unless a/à/y
		if utf8.RuneCountInString(parts[0]) == 1 {
			if parts[0] != "a" && parts[0] != "à" && parts[0] != "y" {
				continue
			}
		}
		// remove ...è unless exception
		if strings.HasSuffix(low, "è") {
			if _, ok := frenchExceptionsEgrave[low]; !ok {
				continue
			}
		}

		posNew := 0
		for posNew < len(out) &&
			strings.EqualFold(tools.RemoveDiacritics(out[posNew]), wordWithoutDiacritics) {
			posNew++
		}

		// TOKEN_AT_START split → near front
		if len(parts) == 2 {
			if _, ok := frenchTokenAtStart[parts[0]]; ok && utf8.RuneCountInString(parts[1]) > 1 {
				out = insertAt(out, posNew, sug)
				continue
			}
		}

		// diacritic-equivalent → near front
		if strings.EqualFold(tools.RemoveDiacritics(sug), wordWithoutDiacritics) {
			out = insertAt(out, posNew, sug)
			continue
		}

		// apostrophe/hyphen cleaned equals word → second position (when i>1)
		clean := hyphenOrQuote.ReplaceAllString(sug, "")
		if i > 1 && n > 2 && strings.EqualFold(clean, word) {
			if posNew == 0 {
				posNew = 1
			}
			if posNew > len(out) {
				posNew = len(out)
			}
			out = insertAt(out, posNew, sug)
			continue
		}
		out = append(out, sug)
	}
	return out
}

func insertAt(slice []string, i int, v string) []string {
	if i < 0 {
		i = 0
	}
	if i >= len(slice) {
		return append(slice, v)
	}
	out := make([]string, 0, len(slice)+1)
	out = append(out, slice[:i]...)
	out = append(out, v)
	out = append(out, slice[i:]...)
	return out
}

// additionalTopFrenchSuggestions ports getAdditionalTopSuggestionsString (static arms).
func additionalTopFrenchSuggestions(word string) []string {
	// Java: word.equals("voulai") exact
	if word == "voulai" {
		return []string{"voulais", "voulait"}
	}
	// units: equalsIgnoreCase
	switch strings.ToLower(word) {
	case "mm2":
		return []string{"mm²"}
	case "cm2":
		return []string{"cm²"}
	case "dm2":
		return []string{"dm²"}
	case "m2":
		return []string{"m²"}
	case "km2":
		return []string{"km²"}
	case "mm3":
		return []string{"mm³"}
	case "cm3":
		return []string{"cm³"}
	case "dm3":
		return []string{"dm³"}
	case "m3":
		return []string{"m³"}
	case "km3":
		return []string{"km³"}
	}
	return nil
}

// splitCamelCase delegates to tools.SplitCamelCase (StringTools).
func splitCamelCase(word string) []string {
	if word == "" {
		return nil
	}
	return tools.SplitCamelCase(word)
}
