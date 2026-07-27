package ga

// Utils is the Java-name twin for package-level Irish mutation helpers
// (org.languagetool.tagging.ga.Utils).
type Utils struct{}

func (Utils) FixSuffix(in string) *Retaggable          { return FixSuffix(in) }
func (Utils) MorphWord(in string) []*Retaggable         { return MorphWord(in) }
func (Utils) Demutate(in string) *Retaggable            { return Demutate(in) }
func (Utils) UnEclipse(in string) string               { return UnEclipse(in) }
func (Utils) UnLenite(in string) string                { return UnLenite(in) }
func (Utils) UnLeniteDefiniteS(in string) string       { return UnLeniteDefiniteS(in) }
func (Utils) UnEclipseF(in string) string              { return UnEclipseF(in) }
func (Utils) UnEclipseChar(in string, first, second rune) string {
	return UnEclipseChar(in, first, second)
}
func (Utils) IsUpperVowel(c rune) bool     { return IsUpperVowel(c) }
func (Utils) IsLowerVowel(c rune) bool     { return IsLowerVowel(c) }
func (Utils) IsVowel(c rune) bool          { return IsVowel(c) }
func (Utils) IsUpperLenitable(c rune) bool { return IsUpperLenitable(c) }
func (Utils) IsLowerLenitable(c rune) bool { return IsLowerLenitable(c) }
func (Utils) IsSLenitable(c rune) bool     { return IsSLenitable(c) }
func (Utils) Lenite(in string) string      { return Lenite(in) }
func (Utils) Eclipse(in string) string     { return Eclipse(in) }
func (Utils) ToLowerCaseIrish(s string) string {
	return ToLowerCaseIrish(s)
}
func (Utils) StartsWithUppercase(s string) bool { return StartsWithUppercase(s) }
func (Utils) IsAllUppercase(s string) bool      { return IsAllUppercase(s) }
func (Utils) IsPonc(c rune) bool                { return IsPonc(c) }
func (Utils) ContainsPonc(s string) bool        { return ContainsPonc(s) }
func (Utils) UnPonc(s string) string            { return UnPonc(s) }
func (Utils) IsAllMathsChars(s string) bool     { return IsAllMathsChars(s) }
func (Utils) IsAllHalfWidthChars(s string) bool { return IsAllHalfWidthChars(s) }
func (Utils) HalfwidthLatinToLatin(s string) string {
	return HalfwidthLatinToLatin(s)
}
func (Utils) SimplifyMathematical(s string) string {
	return SimplifyMathematical(s)
}
func (Utils) SimplifyMathematicalFlags(s string, normaliseGreek, normaliseDigits bool) string {
	return SimplifyMathematicalFlags(s, normaliseGreek, normaliseDigits)
}
func (Utils) GreekToLatin(s string) string { return GreekToLatin(s) }
func (Utils) HasMixedGreekAndLatin(s string) bool {
	return HasMixedGreekAndLatin(s)
}
func (Utils) HasMixedGreekAndCyrillic(s string) bool {
	return HasMixedGreekAndCyrillic(s)
}
