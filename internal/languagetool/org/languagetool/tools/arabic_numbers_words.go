package tools

import (
	"math/big"
	"strconv"
	"strings"
)

// ArabicNumbersWords ports org.languagetool.tools.ArabicNumbersWords (static helpers).

// NumberToArabicWords ports numberToArabicWords(String) — masculine, no inflection.
func NumberToArabicWords(n string) string {
	return NumberToArabicWordsGender(n, false)
}

// NumberToArabicWordsGender ports numberToArabicWords(String, boolean).
func NumberToArabicWordsGender(n string, isFeminine bool) string {
	bi, ok := new(big.Int).SetString(n, 10)
	if !ok {
		return ""
	}
	return NumberToArabicWordsBig(bi, isFeminine)
}

// NumberToArabicWordsInflected ports numberToArabicWords(String, boolean, boolean, String).
func NumberToArabicWordsInflected(number string, isFeminine, isAttached bool, inflectionCase string) string {
	bi, ok := new(big.Int).SetString(number, 10)
	if !ok {
		return ""
	}
	return NumberToArabicWordsBigInflected(bi, isFeminine, isAttached, inflectionCase)
}

// NumberToArabicWordsBig ports numberToArabicWords(BigInteger, boolean).
func NumberToArabicWordsBig(number *big.Int, isFeminine bool) string {
	return convertToArabic(number, isFeminine, false, "")
}

// NumberToArabicWordsBigInflected ports numberToArabicWords(BigInteger, boolean, boolean, String).
func NumberToArabicWordsBigInflected(number *big.Int, isFeminine, isAttached bool, inflectionCase string) string {
	return convertToArabic(number, isFeminine, isAttached, inflectionCase)
}

// convertToArabic ports private convertToArabic (full arity). Result is String.trim()'d.
func convertToArabic(number *big.Int, isFeminine, isAttached bool, inflectionCase string) string {
	if number == nil {
		return ""
	}
	if number.Sign() == 0 {
		return "صفر"
	}
	if number.Cmp(big.NewInt(1)) == 0 {
		// Java early-return ignores feminine flag.
		return "واحد"
	}
	if number.Cmp(big.NewInt(2)) == 0 {
		// Java: getDigitInflectedStatus(2, 0, false, false, inflectionCase)
		return getDigitInflectedStatus(2, 0, false, false, inflectionCase)
	}

	tempNumber := new(big.Int).Set(number)
	thousand := big.NewInt(1000)
	result := ""
	group := 0
	one := big.NewInt(1)

	// Java: while (tempNumber.compareTo(BigDecimal.ONE) >= 0)
	for tempNumber.Cmp(one) >= 0 {
		mod := new(big.Int).Mod(new(big.Int).Set(tempNumber), thousand)
		tempNumber.Div(tempNumber, thousand)
		// tempValue = tempNumber.setScale(0, FLOOR).intValue()
		tempValue := bigIntToIntClamped(tempNumber)
		ntp := int(mod.Int64())
		groupDescription := processArabicGroup(ntp, group, tempValue, isFeminine, isAttached, inflectionCase)

		if groupDescription != "" {
			if group > 0 {
				if result != "" {
					// Java: result.insert(0, "و" + "");
					result = "و" + result
				}
				if ntp != 2 {
					if ntp%100 != 1 {
						if ntp >= 3 && ntp <= 10 {
							// plural group name
							if group < len(ArabicPluralGroups) {
								result = ArabicPluralGroups[group] + " " + result
							}
						} else {
							if result != "" {
								// appending case
								if group < len(ArabicAppendedGroup) {
									result = ArabicAppendedGroup[group] + " " + result
								}
							} else {
								// normal case
								if group < len(ArabicGroup) {
									result = ArabicGroup[group] + " " + result
								}
							}
						}
					}
				}
			}
			// Java: result.insert(0, groupDescription + " ");
			result = groupDescription + " " + result
		}
		group++
		if group > 20 {
			break
		}
	}
	if result != "" {
		return JavaStringTrim(result)
	}
	return ""
}

func bigIntToIntClamped(n *big.Int) int {
	if n == nil || n.Sign() == 0 {
		return 0
	}
	if n.IsInt64() {
		v := n.Int64()
		const maxInt = int(^uint(0) >> 1)
		if v > int64(maxInt) {
			return maxInt
		}
		if v < int64(-maxInt-1) {
			return -maxInt - 1
		}
		return int(v)
	}
	// processArabicGroup only cares about remainingNumber == 0
	return 1
}

// processArabicGroup ports processArabicGroup with full inflection args.
func processArabicGroup(groupNumber, groupLevel, remainingNumber int, isFeminine, isAttached bool, inflectionCase string) string {
	tens := groupNumber % 100
	hundreds := groupNumber / 100
	result := ""

	if hundreds > 0 {
		if tens == 0 && hundreds == 2 { // حالة المضاف
			if groupLevel == 0 {
				result = getDigitHundredJarStatus(hundreds, inflectionCase)
			} else {
				result = getDigitTwosJarStatus(0, inflectionCase, true)
			}
		} else {
			result = getDigitHundredJarStatus(hundreds, inflectionCase)
		}
	}

	if tens > 0 {
		if tens < 20 {
			if tens == 2 && hundreds == 0 && groupLevel > 0 {
				result = getDigitTwosJarStatus(groupLevel, inflectionCase, false)
			} else {
				if result != "" {
					result += " و"
				}
				if tens == 1 && groupLevel > 0 {
					if groupLevel < len(ArabicGroup) {
						result += ArabicGroup[groupLevel]
					}
				} else if (tens == 1 || tens == 2) &&
					(groupLevel == 0 || groupLevel == -1) &&
					hundreds == 0 && remainingNumber == 0 {
					result += "" // special case for 1 and 2 alone
				} else {
					result += getDigitInflectedStatus(tens, groupLevel, isFeminine, isAttached, inflectionCase)
				}
			}
		} else {
			ones := tens % 10
			tensIdx := (tens / 10) - 2 // 20's offset

			if ones > 0 {
				if result != "" {
					result += " و"
				}
				result += getDigitInflectedStatus(ones, groupLevel, isFeminine, isAttached, inflectionCase)
			}
			if result != "" {
				result += " و"
			}
			result += getDigitTensJarStatus(tensIdx, inflectionCase)
		}
	}
	return result
}

func getDigitInflectedStatus(digit, groupLevel int, isFeminine, isAttached bool, inflectionCase string) string {
	_ = isAttached // Java parameter unused in method body
	if inflectionCase == "jar" {
		if groupLevel == -1 || groupLevel == 0 {
			if !isFeminine {
				return listAt(ArabicJarOnes, digit)
			}
			return listAt(ArabicJarFeminineOnes, digit)
		}
	} else {
		if groupLevel == -1 || groupLevel == 0 {
			if !isFeminine {
				return listAt(ArabicOnes, digit)
			}
			return listAt(ArabicFeminineOnes, digit)
		}
	}
	return listAt(ArabicOnes, digit)
}

func getDigitTensJarStatus(digit int, inflectionCase string) string {
	// Java: equals("jar") || equals("jar") — redundant twin; single check is identical.
	if inflectionCase == "jar" {
		return listAt(ArabicJarTens, digit)
	}
	return listAt(ArabicTens, digit)
}

func getDigitHundredJarStatus(digit int, inflectionCase string) string {
	// Java: equals("jar") || equals("jar") — redundant twin; single check is identical.
	if inflectionCase == "jar" {
		return listAt(ArabicJarHundreds, digit)
	}
	return listAt(ArabicHundreds, digit)
}

func getDigitTwosJarStatus(digit int, inflectionCase string, isAppended bool) string {
	if !isAppended {
		// Java: equals("jar") || equals("jar")
		if inflectionCase == "jar" {
			return listAt(ArabicJarTwos, digit)
		}
		return listAt(ArabicTwos, digit)
	}
	if inflectionCase == "jar" {
		return listAt(ArabicJarAppendedTwos, digit)
	}
	return listAt(ArabicAppendedTwos, digit)
}

func listAt(list []string, i int) string {
	if i < 0 || i >= len(list) {
		return ""
	}
	return list[i]
}

// TextToNumber ports textToNumber(String).
func TextToNumber(text string) int {
	text = RemoveTashkeel(text)
	words := strings.Split(text, " ")
	return TextToNumberWords(words)
}

// TextToNumberWords ports textToNumber(List<String>).
func TextToNumberWords(words []string) int {
	total := 0
	partial := 0
	for _, word := range words {
		if word != "" && word != "واحد" &&
			(strings.HasPrefix(word, "و") ||
				strings.HasPrefix(word, "ف") ||
				strings.HasPrefix(word, "ب") ||
				strings.HasPrefix(word, "ك") ||
				strings.HasPrefix(word, "ل")) {
			r := []rune(word)
			if len(r) > 0 {
				word = string(r[1:])
			}
		}
		if word != "واحد" && strings.HasPrefix(word, "و") {
			r := []rune(word)
			if len(r) > 0 {
				word = string(r[1:])
			}
		}
		if IsNumericWord(word) {
			actualnumber := GetNumericWordValue(word)
			if actualnumber%1000 == 0 {
				if partial == 0 {
					partial = 1
				}
				total += partial * actualnumber
				partial = 0
			} else {
				partial += GetNumericWordValue(word)
			}
		}
	}
	total += partial
	return total
}

// GetUnitForm ports getUnitForm.
func GetUnitForm(n int, unitLemma, inflection string) string {
	phraseMap := NumberToWordsWithUnitsMap(n, unitLemma, inflection)
	return phraseMap["unit"]
}

// NumberToWordsWithUnits ports numberToWordsWithUnits.
func NumberToWordsWithUnits(n int, unit, inflection string) string {
	phraseMap := NumberToWordsWithUnitsMap(n, unit, inflection)
	return phraseMap["all"]
}

// NumberToWordsWithUnitsMap ports numberToWordsWithUnitsMap.
func NumberToWordsWithUnitsMap(n int, unit, inflection string) map[string]string {
	feminin := IsArabicUnitFeminin(unit)
	unitInflection := ""
	unitNumber := ""
	numberPhrase := NumberToArabicWordsInflected(strconv.Itoa(n), feminin, true, inflection)
	var phrase strings.Builder
	newUnit := ""

	switch {
	case n == 0:
		newUnit = GetArabicUnitForm(unit, "plural", "nasb")
		unitInflection = "nasb"
		unitNumber = "plural"
		phrase.WriteString("لا")
		phrase.WriteString(" ")
		phrase.WriteString(newUnit)
	case n == 1:
		newUnit = GetArabicUnitForm(unit, "one", inflection)
		unitInflection = inflection
		unitNumber = "one"
		phrase.WriteString(newUnit)
		phrase.WriteString(" ")
		phrase.WriteString(numberPhrase)
	case n == 2:
		newUnit = GetArabicUnitForm(unit, "two", inflection)
		unitInflection = inflection
		unitNumber = "two"
		phrase.WriteString(newUnit)
	case n%100 == 1:
		numberPhraseHundred := NumberToArabicWordsInflected(strconv.Itoa(n-1), feminin, true, inflection)
		newUnitHundred := GetArabicUnitForm(unit, "one", "jar")
		newUnitOne := GetArabicUnitForm(unit, "one", inflection)
		unitInflection = inflection
		unitNumber = "one"
		newUnit = newUnitOne
		phrase.WriteString(numberPhraseHundred)
		phrase.WriteString(" ")
		phrase.WriteString(newUnitHundred)
		phrase.WriteString(" ")
		phrase.WriteString("و")
		phrase.WriteString(newUnitOne)
	case n%100 == 2:
		numberPhraseHundred := NumberToArabicWordsInflected(strconv.Itoa(n-2), feminin, true, inflection)
		newUnitHundred := GetArabicUnitForm(unit, "one", "jar")
		newUnitTwo := GetArabicUnitForm(unit, "two", inflection)
		unitInflection = inflection
		unitNumber = "two"
		newUnit = newUnitTwo
		phrase.WriteString(numberPhraseHundred)
		phrase.WriteString(" ")
		phrase.WriteString(newUnitHundred)
		phrase.WriteString(" ")
		phrase.WriteString("و")
		phrase.WriteString(newUnitTwo)
	case n%100 >= 3 && n%100 <= 10:
		newUnit = GetArabicUnitForm(unit, "plural", "jar")
		unitInflection = "jar"
		unitNumber = "plural"
		phrase.WriteString(numberPhrase)
		phrase.WriteString(" ")
		phrase.WriteString(newUnit)
	case n%100 >= 11:
		newUnit = GetArabicUnitForm(unit, "one", "nasb")
		unitInflection = "nasb"
		unitNumber = "one"
		phrase.WriteString(numberPhrase)
		phrase.WriteString(" ")
		phrase.WriteString(newUnit)
	case n%100 == 0:
		newUnit = GetArabicUnitForm(unit, "one", "jar")
		unitInflection = "jar"
		unitNumber = "one"
		phrase.WriteString(numberPhrase)
		phrase.WriteString(" ")
		phrase.WriteString(newUnit)
	default:
		phrase.WriteString(numberPhrase)
		phrase.WriteString(" **")
		phrase.WriteString(unit + "**")
	}

	return map[string]string{
		"all":            phrase.String(),
		"phrase":         numberPhrase,
		"unit":           newUnit,
		"number":         strconv.Itoa(n),
		"unitLemma":      unit,
		"unitNumber":     unitNumber,
		"unitInflection": unitInflection,
		"inflection":     inflection,
		"feminin":        strconv.FormatBool(feminin),
	}
}

// CheckNumericPhrase ports checkNumericPhrase.
func CheckNumericPhrase(phraseInput string, feminin, attached bool, inflection string) bool {
	phrase := RemoveTashkeel(phraseInput)
	x := TextToNumber(phrase)
	autoPhrase := NumberToArabicWordsInflected(strconv.Itoa(x), feminin, attached, inflection)
	autoPhrase = RemoveTashkeel(autoPhrase)
	return autoPhrase == phrase
}

// GetSuggestionsNumericPhrase ports getSuggestionsNumericPhrase.
func GetSuggestionsNumericPhrase(phraseInput string, feminin, attached bool, inflection string) []string {
	var suggestions []string
	phrase := RemoveTashkeel(phraseInput)
	x := TextToNumber(phrase)
	autoPhrase := NumberToArabicWordsInflected(strconv.Itoa(x), feminin, attached, inflection)
	autoPhrase = RemoveTashkeel(autoPhrase)
	if autoPhrase != phrase {
		if inflection != "" {
			suggestions = append(suggestions, autoPhrase)
		} else {
			// Raf3
			inflection = "raf3"
			autoPhrase = NumberToArabicWordsInflected(strconv.Itoa(x), feminin, attached, inflection)
			suggestions = append(suggestions, autoPhrase)
			// nasb+jar
			inflection = "jar"
			autoPhrase = NumberToArabicWordsInflected(strconv.Itoa(x), feminin, attached, inflection)
			suggestions = append(suggestions, autoPhrase)
		}
	}
	return suggestions
}

// GetSuggestionsNumericPhraseWithUnits ports getSuggestionsNumericPhraseWithUnits.
func GetSuggestionsNumericPhraseWithUnits(phraseInput, unit string, feminin, attached bool, inflection string) []map[string]string {
	var suggestions []map[string]string
	phrase := RemoveTashkeel(phraseInput)
	x := TextToNumber(phrase)
	autoPhraseMap := NumberToWordsWithUnitsMap(x, unit, inflection)
	autoPhrase := RemoveTashkeel(autoPhraseMap["phrase"])
	if autoPhrase != phrase {
		if inflection != "" {
			suggestions = append(suggestions, autoPhraseMap)
		} else {
			inflection = "raf3"
			autoPhraseMap = NumberToWordsWithUnitsMap(x, unit, inflection)
			suggestions = append(suggestions, autoPhraseMap)
			inflection = "jar"
			autoPhraseMap = NumberToWordsWithUnitsMap(x, unit, inflection)
			suggestions = append(suggestions, autoPhraseMap)
		}
	}
	return suggestions
}

// IsNumericWord ports isNumericWord.
func IsNumericWord(word string) bool {
	_, ok := ArabicNumberWords[word]
	return ok
}

// GetNumericWordValue ports getNumericWordValue.
func GetNumericWordValue(word string) int {
	return ArabicNumberWords[word]
}
