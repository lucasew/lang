package uk

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

// Java CompoundTagger SKY / SKYI patterns for по-… adverbs.
var (
	rePoSky  = regexp.MustCompile(`(?i).*[сзц]ьки$`)  // по-сибірськи → lookup +й
	rePoSkyi = regexp.MustCompile(`(?i).*[сзц]ький$`) // already -ський
)

// DynamicPoAdvReadings ports CompoundTagger left "по" + poAdvMatch.
// Requires right-side adj tags from wordTagger (fail-closed without dict).
// Yields a single adv reading with lemma = full surface (Java).
func DynamicPoAdvReadings(token string, tagWord func(string) []tagging.TaggedWord) []struct{ Lemma, POS string } {
	if tagWord == nil || token == "" || !strings.Contains(token, "-") {
		return nil
	}
	if strings.Count(token, "-") != 1 {
		return nil
	}
	dash := strings.LastIndex(token, "-")
	leftWord := token[:dash]
	rightWord := token[dash+1:]
	if !strings.EqualFold(leftWord, "по") || rightWord == "" {
		return nil
	}

	// Java: required adj tag fragment on the right analysis.
	var needPrefix string
	lookup := rightWord
	lowR := strings.ToLower(rightWord)
	switch {
	case strings.HasSuffix(lowR, "ому"):
		// poAdvMatch(..., ADJ_TAG_FOR_PO_ADV_MIS)
		needPrefix = "adj:m:v_mis"
	case rePoSkyi.MatchString(rightWord):
		// poAdvMatch(..., ADJ_TAG_FOR_PO_ADV_NAZ)
		needPrefix = "adj:m:v_naz"
	case rePoSky.MatchString(rightWord):
		// early adjust: rightWord += "й" then SKYI path
		lookup = rightWord + "й"
		needPrefix = "adj:m:v_naz"
	default:
		return nil
	}

	tws := tagWord(lookup)
	if len(tws) == 0 {
		low := strings.ToLower(lookup)
		if low != lookup {
			tws = tagWord(low)
		}
	}
	if len(tws) == 0 {
		return nil
	}
	for _, tw := range tws {
		if tw.PosTag == "" {
			continue
		}
		// Java: posTag.startsWith(adjTag)
		if strings.HasPrefix(tw.PosTag, needPrefix) {
			return []struct{ Lemma, POS string }{{Lemma: token, POS: "adv"}}
		}
	}
	return nil
}
