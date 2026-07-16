package pt

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/multitoken"

// PortugueseMultitokenSpeller ports org.languagetool.rules.pt.PortugueseMultitokenSpeller.
type PortugueseMultitokenSpeller struct {
	*multitoken.MultitokenSpeller
}

func NewPortugueseMultitokenSpeller() *PortugueseMultitokenSpeller {
	return &PortugueseMultitokenSpeller{MultitokenSpeller: multitoken.NewMultitokenSpeller()}
}
