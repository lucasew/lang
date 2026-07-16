package en

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/multitoken"

// EnglishMultitokenSpeller ports org.languagetool.rules.en.EnglishMultitokenSpeller
// as a typed MultitokenSpeller handle; callers LoadWords from embedded multiword resources.
type EnglishMultitokenSpeller struct {
	*multitoken.MultitokenSpeller
}

func NewEnglishMultitokenSpeller() *EnglishMultitokenSpeller {
	return &EnglishMultitokenSpeller{MultitokenSpeller: multitoken.NewMultitokenSpeller()}
}
