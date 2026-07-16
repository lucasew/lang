package fr

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/multitoken"

// FrenchMultitokenSpeller ports org.languagetool.rules.fr.FrenchMultitokenSpeller.
type FrenchMultitokenSpeller struct {
	*multitoken.MultitokenSpeller
}

func NewFrenchMultitokenSpeller() *FrenchMultitokenSpeller {
	return &FrenchMultitokenSpeller{MultitokenSpeller: multitoken.NewMultitokenSpeller()}
}
