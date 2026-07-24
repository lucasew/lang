package de

import "regexp"

// Multi-replacement regex ADDITIONAL_SUGGESTIONS (put + Arrays.asList of fixed strings).

type additionalSugMultiEntry struct {
	word  *regexp.Regexp
	repls []string
}

var additionalSugMulti []additionalSugMultiEntry

func init() {
	registerAdditionalSugMulti(`[aA]wa`, []string{`AWA`, `ach was`, `aber`})
	registerAdditionalSugMulti(`allmöglichen?`, []string{`alle möglichen`, `alle mögliche`})
	registerAdditionalSugMulti(`vorr?auss?etzlich`, []string{`voraussichtlich`, `vorausgesetzt`})
	registerAdditionalSugMulti(`Büff?(ee|é)`, []string{`Buffet`, `Büfett`})
	registerAdditionalSugMulti(`[wW]elan`, []string{`WLAN`, `W-LAN`})
}

func registerAdditionalSugMulti(wordPat string, repls []string) {
	w, err := regexp.Compile("^(?:" + wordPat + ")$")
	if err != nil {
		return
	}
	additionalSugMulti = append(additionalSugMulti, additionalSugMultiEntry{word: w, repls: repls})
}
