package ca

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"

// CatalanNumberInWordFilter ports AbstractNumberInWordFilter for Catalan.
type CatalanNumberInWordFilter struct {
	*rules.NumberInWordFilter
}

func NewCatalanNumberInWordFilter() *CatalanNumberInWordFilter {
	return &CatalanNumberInWordFilter{NumberInWordFilter: rules.NewNumberInWordFilter()}
}
