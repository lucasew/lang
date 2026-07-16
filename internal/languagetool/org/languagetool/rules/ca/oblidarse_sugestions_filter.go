package ca

import "regexp"

// OblidarseSugestionsFilter reflexive prefix tables.
var addReflexiveVowel = map[string]string{
	"1S": "m'",
	"2S": "t'",
	"3S": "s'",
	"1P": "ens ",
	"2P": "us ",
	"3P": "s'",
}
var addReflexiveConsonant = map[string]string{
	"1S": "em ",
	"2S": "et ",
	"3S": "es ",
	"1P": "ens ",
	"2P": "us ",
	"3P": "es ",
}
var addReflexiveEnVowel = map[string]string{
	"1S": "me n'",
	"2S": "te n'",
	"3S": "se n'",
	"1P": "ens n'",
	"2P": "us n'",
	"3P": "se n'",
}
var addReflexiveEnConsonant = map[string]string{
	"1S": "me'n ",
	"2S": "te'n ",
	"3S": "se'n ",
	"1P": "ens en ",
	"2P": "us en ",
	"3P": "se'n ",
}

var pApostropheNeededOblidar = regexp.MustCompile(`(?i)^h?[aeiouàèéíòóú].*`)

// OblidarseSugestionsFilter ports reflexive prefix selection for OBLIDARSE suggestions.
type OblidarseSugestionsFilter struct{}

func NewOblidarseSugestionsFilter() *OblidarseSugestionsFilter {
	return &OblidarseSugestionsFilter{}
}

// ReflexivePrefix returns the weak-pronoun prefix for personNumber (e.g. "1S")
// given whether the next word needs an apostrophe and whether "en" is included.
func (f *OblidarseSugestionsFilter) ReflexivePrefix(personNumber string, nextNeedsApos, withEn bool) string {
	if withEn {
		if nextNeedsApos {
			return addReflexiveEnVowel[personNumber]
		}
		return addReflexiveEnConsonant[personNumber]
	}
	if nextNeedsApos {
		return addReflexiveVowel[personNumber]
	}
	return addReflexiveConsonant[personNumber]
}

// NeedsApostrophe reports vowel-initial following words.
func (f *OblidarseSugestionsFilter) NeedsApostrophe(nextWord string) bool {
	return pApostropheNeededOblidar.MatchString(nextWord)
}
