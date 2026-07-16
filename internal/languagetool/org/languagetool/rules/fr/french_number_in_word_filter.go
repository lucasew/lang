package fr

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"

// FrenchNumberInWordFilter ports AbstractNumberInWordFilter for French.
type FrenchNumberInWordFilter struct {
	*rules.NumberInWordFilter
}

func NewFrenchNumberInWordFilter() *FrenchNumberInWordFilter {
	return &FrenchNumberInWordFilter{NumberInWordFilter: rules.NewNumberInWordFilter()}
}
