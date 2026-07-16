package ca

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// FindSuggestionsEsFilter ports org.languagetool.rules.ca.FindSuggestionsEsFilter
// suggestion rewriting for "es" + misspelling → "és X" / "es V".
type FindSuggestionsEsFilter struct {
	*rules.FindSuggestionsFilter
}

func NewFindSuggestionsEsFilter() *FindSuggestionsEsFilter {
	return &FindSuggestionsEsFilter{FindSuggestionsFilter: rules.NewFindSuggestionsFilter()}
}

var (
	pApostropheNeededES = regexp.MustCompile(`(?i)^h?[aeiouàèéíòóú].*`)
	pPostagNominal      = regexp.MustCompile(`^NP..[^0].*$|^NC.[SN].*$|^A...[SN].$|^V\.P..S..$|^V\.[NG].*$|^RG$|^PX..S...$`)
	pPostagVerb3person  = regexp.MustCompile(`^V...3.*$`)
)

// RewriteEsSuggestions builds "és "+nominal / "es "+verb3 suggestions from tagged candidates.
// candidates are (form, pos) pairs from tagger.
func (f *FindSuggestionsEsFilter) RewriteEsSuggestions(candidates []struct{ Form, POS string }, max int) []string {
	if max <= 0 {
		max = 20
	}
	var out []string
	seen := map[string]struct{}{}
	for _, c := range candidates {
		if len(out) >= max {
			break
		}
		if pPostagNominal.MatchString(c.POS) {
			s := "és " + c.Form
			if _, ok := seen[s]; !ok {
				seen[s] = struct{}{}
				out = append(out, s)
			}
		}
		if pPostagVerb3person.MatchString(c.POS) {
			s := "es " + c.Form
			if pApostropheNeededES.MatchString(c.Form) {
				// keep space form; Java may use s' in some paths — surface keep "es "
				_ = strings.TrimSpace
			}
			if _, ok := seen[s]; !ok {
				seen[s] = struct{}{}
				out = append(out, s)
			}
		}
	}
	return out
}
