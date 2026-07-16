package nl

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/multitoken"

// DutchMultitokenSpeller ports org.languagetool.rules.nl.DutchMultitokenSpeller.
type DutchMultitokenSpeller struct {
	*multitoken.MultitokenSpeller
}

func NewDutchMultitokenSpeller() *DutchMultitokenSpeller {
	return &DutchMultitokenSpeller{MultitokenSpeller: multitoken.NewMultitokenSpeller()}
}
