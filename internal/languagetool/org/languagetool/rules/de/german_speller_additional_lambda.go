package de

import (
	"regexp"
	"strings"
)

// Replace-style ADDITIONAL_SUGGESTIONS lambdas (hand-ported from Java static block).

var additionalSugLambdas = []struct {
	word *regexp.Regexp
	fn   func(string) []string
}{
	{regexp.MustCompile(`^(?:[aA]lsallerersten?s)$`), func(w string) []string {
		re := regexp.MustCompile(`lsallerersten?s`)
		return []string{replaceFirst(re, w, "ls allererstes"), replaceFirst(re, w, "ls Allererstes")}
	}},
	{regexp.MustCompile(`^(?:[pP]roblemhaft(?:e[nmrs]?)?)$`), func(w string) []string {
		return []string{strings.Replace(w, "haft", "behaftet", 1), strings.Replace(w, "haft", "atisch", 1)}
	}},
	{regexp.MustCompile(`^(?:rosane[mnrs]?)$`), func(w string) []string {
		return []string{"rosa", replaceFirst(regexp.MustCompile(`^rosan`), w, "rosafarben")}
	}},
	{regexp.MustCompile(`^(?:kreativlos(?:e[nmrs]?)?)$`), func(w string) []string {
		return []string{
			strings.Replace(w, "kreativ", "fantasie", 1),
			strings.Replace(w, "kreativ", "einfalls", 1),
			strings.Replace(w, "kreativlos", "unkreativ", 1),
			strings.Replace(w, "kreativlos", "uninspiriert", 1),
		}
	}},
	{regexp.MustCompile(`^(?:[mM]illion(?:en)?mal)$`), func(w string) []string {
		s := strings.Replace(w, "mal", " Mal", 1)
		return []string{uppercaseFirstChar(s)}
	}},
	{regexp.MustCompile(`^(?:[zZ]auberlich(?:e[mnrs]?)?)$`), func(w string) []string {
		return []string{strings.Replace(w, "lich", "isch", 1), strings.Replace(w, "lich", "haft", 1)}
	}},
	{regexp.MustCompile(`^(?:unverantwortungs?los(?:e[nmrs]?)?)$`), func(w string) []string {
		return []string{
			replaceFirst(regexp.MustCompile(`unverantwortungs?`), w, "verantwortungs"),
			replaceFirst(regexp.MustCompile(`ungs?los`), w, "lich"),
		}
	}},
	{regexp.MustCompile(`^(?:[wW]ohlfühlseins?)$`), func(w string) []string {
		re := regexp.MustCompile(`[wW]ohlfühlsein`)
		return []string{"Wellness", replaceFirst(re, w, "Wohlbefinden"), replaceFirst(re, w, "Wohlfühlen")}
	}},
	{regexp.MustCompile(`^(?:beidige[mnrs]?)$`), func(w string) []string {
		return []string{strings.Replace(w, "ig", "", 1), strings.Replace(w, "beid", "beiderseit", 1), "beeidigen"}
	}},
	{regexp.MustCompile(`^(?:palletten?)$`), func(w string) []string {
		return []string{strings.Replace(w, "pall", "Pal", 1), strings.Replace(w, "pa", "Pai", 1)}
	}},
	{regexp.MustCompile(`^(?:schafen?)$`), func(w string) []string {
		return []string{
			strings.Replace(w, "sch", "schl", 1),
			strings.Replace(w, "af", "arf", 1),
			strings.Replace(w, "af", "aff", 1),
		}
	}},
	{regexp.MustCompile(`^(?:Panelen?)$`), func(w string) []string {
		return []string{strings.Replace(w, "Panel", "Paneel", 1), "Panels"}
	}},
}

func lookupAdditionalSuggestionsLambda(word string) []string {
	// Java String.length() (UTF-16) oversized-token gate
	if word == "" || utf16LenDE(word) > additionalSugMaxMatchLen {
		return nil
	}
	for _, e := range additionalSugLambdas {
		if e.word.MatchString(word) {
			return e.fn(word)
		}
	}
	return nil
}
