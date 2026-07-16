package ca

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// DonarseliBeFilter ports surface constants and pronoun rewrites for
// "donar-se'n bé/malament" style suggestions.
type DonarseliBeFilter struct{}

func NewDonarseliBeFilter() *DonarseliBeFilter {
	return &DonarseliBeFilter{}
}

// AdverbiFinal are terminal adverbs accepted after the verb cluster.
var AdverbiFinal = map[string]struct{}{
	"bé": {}, "malament": {}, "mal": {}, "millor": {}, "pitjor": {}, "fatal": {},
}

// PronomsPersonals are strong personal pronouns for "a mi/tu/…" spans.
var PronomsPersonals = map[string]struct{}{
	"mi": {}, "tu": {}, "ell": {}, "ella": {},
	"nosaltres": {}, "vosaltres": {}, "ells": {}, "elles": {},
}

// ExceptionsQue words that block a preceding "que".
var ExceptionsQue = map[string]struct{}{
	"ja": {}, "ara": {}, "per": {}, "de": {}, "a": {}, "en": {},
}

// DespresDarrerAdverbiPOS matches tokens allowed after the final adverb.
var DespresDarrerAdverbiPOS = regexp.MustCompile(`^V\.N.*$|^D.*$|^PD.*$`)

// NormalizeAdverbi maps mal/fatal → malament.
func NormalizeAdverbi(token string) string {
	switch strings.ToLower(token) {
	case "mal", "fatal":
		return "malament"
	default:
		return token
	}
}

// IsAdverbiFinal reports terminal adverbs.
func IsAdverbiFinal(token string) bool {
	_, ok := AdverbiFinal[strings.ToLower(token)]
	return ok
}

// IsPronomPersonal reports strong personal pronouns.
func IsPronomPersonal(token string) bool {
	_, ok := PronomsPersonals[strings.ToLower(token)]
	return ok
}

// BuildDonarSuggestion attaches "en" to a weak pronoun cluster before/after the verb.
// pronom is the relevant weak pronoun token; verb is the main verb form.
func (f *DonarseliBeFilter) BuildDonarSuggestion(pronom, verb string, pronounsBefore bool, casingModel string) string {
	norm := Transform(pronom, PronounNormalized) + " en"
	var s string
	if pronounsBefore {
		s = TransformDavant(norm, verb) + verb
	} else {
		s = verb + TransformDarrere(norm, verb)
	}
	if casingModel != "" {
		s = tools.PreserveCase(s, casingModel)
	}
	return s
}

// IsExceptionQue reports whether a token before "que" blocks the rewrite.
func IsExceptionQue(token string) bool {
	_, ok := ExceptionsQue[strings.ToLower(token)]
	return ok
}
