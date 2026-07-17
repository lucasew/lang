package pt

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

// Verb prefixes that always require a hyphen (Java PREFIXES_FOR_VERBS subset).
var rePrefixedVerb = regexp.MustCompile(`(?i)^(soto-|anti-|super-|pseudo-|ultra-)(.+)$`)

var reVerbPOS = regexp.MustCompile(`(?i)^V`)

// PrefixedVerbReadings tags soto-trepei style forms when the bare verb is in the dict.
func PrefixedVerbReadings(word string, tagWord func(string) []tagging.TaggedWord) []*languagetool.AnalyzedToken {
	m := rePrefixedVerb.FindStringSubmatch(word)
	if m == nil {
		return nil
	}
	prefix, verb := strings.ToLower(m[1]), strings.ToLower(m[2])
	tws := tagWord(verb)
	var out []*languagetool.AnalyzedToken
	for _, tw := range tws {
		if tw.PosTag == "" || !reVerbPOS.MatchString(tw.PosTag) {
			continue
		}
		lemma := prefix + tw.Lemma
		if tw.Lemma == "" {
			lemma = prefix + verb
		}
		// only if combined lemma is unknown (Java: lookup empty)
		if len(tagWord(lemma)) > 0 {
			continue
		}
		p, l := tw.PosTag, lemma
		out = append(out, languagetool.NewAnalyzedToken(word, &p, &l))
	}
	return out
}
