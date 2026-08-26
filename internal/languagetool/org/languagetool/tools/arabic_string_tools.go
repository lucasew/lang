package tools

import "regexp"

// TashkeelChars ports ArabicStringTools.TASHKEEL_CHARS.
const TashkeelChars = "\u064B" + // Fathatan
	"\u064C" + // Dammatan
	"\u064D" + // Kasratan
	"\u064E" + // Fatha
	"\u064F" + // Damma
	"\u0650" + // Kasra
	"\u0651" + // Shadda
	"\u0652" + // Sukun
	"\u0653" + // Maddah Above
	"\u0654" + // Hamza Above
	"\u0655" + // Hamza Below
	"\u0656" + // Subscript Alef
	"\u0640" // Tatweel

var tashkeelPattern = regexp.MustCompile("[" + TashkeelChars + "]")

// RemoveTashkeel ports ArabicStringTools.removeTashkeel.
func RemoveTashkeel(str string) string {
	return tashkeelPattern.ReplaceAllString(str, "")
}
