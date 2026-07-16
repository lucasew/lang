package nl

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"

// DutchNumberInWordFilter ports AbstractNumberInWordFilter for Dutch.
type DutchNumberInWordFilter struct {
	*rules.NumberInWordFilter
}

func NewDutchNumberInWordFilter() *DutchNumberInWordFilter {
	return &DutchNumberInWordFilter{NumberInWordFilter: rules.NewNumberInWordFilter()}
}
