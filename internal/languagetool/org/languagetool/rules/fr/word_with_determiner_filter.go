package fr

import "regexp"

// WordWithDeterminerFilter ports surface constants and POS pattern helpers from
// org.languagetool.rules.fr.WordWithDeterminerFilter (synthesis deferred).
type WordWithDeterminerFilter struct{}

func NewWordWithDeterminerFilter() *WordWithDeterminerFilter {
	return &WordWithDeterminerFilter{}
}

var (
	// DetPOS matches determiner/adjective/participle POS used as determiner.
	DetPOS = regexp.MustCompile(`^(P.)?D .*|^J .*|^V.* ppa .*`)
	// WordPOS matches noun/adj/participle targets.
	WordPOS = regexp.MustCompile(`^[ZNJ] .*|^V.* ppa .*`)
)

// GenderNumberPatterns are MS/FS/MP/FP POS fragments (Java genderNumber array).
var GenderNumberPatterns = []string{
	`([me]) (s|sp)`,
	`([fe]) (s|sp)`,
	`([me]) (p|sp)`,
	`([fe]) (p|sp)`,
}

// ExceptionsDeterminer are irregular plural det forms that skip some rewrites.
var ExceptionsDeterminer = map[string]struct{}{
	"bels": {}, "fols": {}, "mols": {}, "nouvels": {},
}

// ElisionRulesToCheck are French rule IDs related to elision for post-filtering.
var ElisionRulesToCheck = []string{"CET_CE", "CE_CET", "MA_VOYELLE", "MON_NFS", "VIEUX"}

// CategoryToCheck is the category id used when validating elision.
const CategoryToCheck = "CAT_ELISION"

// IsExceptionDeterminer reports irregular determiner plurals.
func (f *WordWithDeterminerFilter) IsExceptionDeterminer(token string) bool {
	_, ok := ExceptionsDeterminer[token]
	return ok
}

// MatchesDetPOS / MatchesWordPOS check POS patterns.
func (f *WordWithDeterminerFilter) MatchesDetPOS(pos string) bool  { return DetPOS.MatchString(pos) }
func (f *WordWithDeterminerFilter) MatchesWordPOS(pos string) bool { return WordPOS.MatchString(pos) }

// NounAdjPrefix returns the synthesizer prefix for noun-only / adj-only / both.
func (f *WordWithDeterminerFilter) NounAdjPrefix(isNoun, isAdjective bool) string {
	if isNoun && !isAdjective {
		return "[NZ] "
	}
	if !isNoun && isAdjective {
		return "J "
	}
	return "[ZNJ] "
}
