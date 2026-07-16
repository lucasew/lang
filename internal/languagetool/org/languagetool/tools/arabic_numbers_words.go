package tools

import (
	"math/big"
	"strings"
)

// NumberToArabicWords ports ArabicNumbersWords.numberToArabicWords (nominative, masculine).
func NumberToArabicWords(n string) string {
	return NumberToArabicWordsGender(n, false)
}

// NumberToArabicWordsGender converts a decimal integer string to Arabic words.
func NumberToArabicWordsGender(n string, feminine bool) string {
	bi, ok := new(big.Int).SetString(strings.TrimSpace(n), 10)
	if !ok {
		return ""
	}
	return NumberToArabicWordsBig(bi, feminine)
}

func NumberToArabicWordsBig(number *big.Int, feminine bool) string {
	if number == nil {
		return ""
	}
	if number.Sign() < 0 {
		return "سالب " + NumberToArabicWordsBig(new(big.Int).Neg(number), feminine)
	}
	if number.Sign() == 0 {
		return "صفر"
	}
	if number.Cmp(big.NewInt(1)) == 0 {
		if feminine {
			return ArabicFeminineOnes[1]
		}
		return "واحد"
	}
	if number.Cmp(big.NewInt(2)) == 0 {
		if feminine {
			return ArabicFeminineOnes[2]
		}
		return ArabicOnes[2]
	}

	// process groups of 1000
	temp := new(big.Int).Set(number)
	thousand := big.NewInt(1000)
	var parts []string
	group := 0
	for temp.Sign() > 0 {
		mod := new(big.Int).Mod(temp, thousand)
		temp.Div(temp, thousand)
		g := int(mod.Int64())
		if g != 0 {
			desc := processArabicGroup(g, group, feminine)
			if group > 0 {
				groupName := arabicGroupName(g, group)
				switch {
				case desc == "" && groupName != "":
					// e.g. 1000 → "ألف"
					desc = groupName
				case desc != "" && groupName != "":
					desc = desc + " " + groupName
				case desc == "" && g == 1 && group < len(ArabicGroup):
					desc = ArabicGroup[group]
				}
			}
			if desc != "" {
				parts = append([]string{desc}, parts...)
			}
		}
		group++
		if group > 7 {
			break
		}
	}
	return strings.TrimSpace(strings.Join(parts, " و"))
}

func arabicGroupName(n, group int) string {
	if group <= 0 || group >= len(ArabicGroup) {
		return ""
	}
	if n == 2 {
		if group < len(ArabicTwos) {
			// "ألفان" already includes the dual; processArabicGroup returns empty for 2? handle in process
			return ""
		}
	}
	if n >= 3 && n <= 10 {
		if group < len(ArabicPluralGroups) {
			return ArabicPluralGroups[group]
		}
	}
	if group < len(ArabicGroup) {
		return ArabicGroup[group]
	}
	return ""
}

func processArabicGroup(n, group int, feminine bool) string {
	if n == 0 {
		return ""
	}
	// dual group unit alone (e.g. 2000 → ألفان)
	if n == 2 && group > 0 && group < len(ArabicTwos) {
		return ArabicTwos[group]
	}
	if n == 1 && group > 0 {
		return "" // group name alone: "ألف"
	}
	if n == 2 && group == 0 {
		if feminine {
			return ArabicFeminineOnes[2]
		}
		return ArabicOnes[2]
	}

	hundreds := n / 100
	rest := n % 100
	var b strings.Builder
	if hundreds > 0 && hundreds < len(ArabicHundreds) {
		b.WriteString(ArabicHundreds[hundreds])
	}
	if rest > 0 {
		if b.Len() > 0 {
			b.WriteString(" و")
		}
		// tens+ones; feminine applies to ones when no higher place in group for 1-19 style
		useFem := feminine && group == 0
		if rest < 20 {
			ones := ArabicOnes
			if useFem {
				ones = ArabicFeminineOnes
			}
			if rest < len(ones) {
				b.WriteString(ones[rest])
			}
		} else {
			onesDigit := rest % 10
			tensDigit := rest / 10
			if onesDigit > 0 {
				ones := ArabicOnes
				if useFem {
					ones = ArabicFeminineOnes
				}
				if onesDigit < len(ones) {
					b.WriteString(ones[onesDigit])
					b.WriteString(" و")
				}
			}
			if tensDigit >= 2 && tensDigit-2 < len(ArabicTens) {
				b.WriteString(ArabicTens[tensDigit-2])
			}
		}
	}
	return b.String()
}
