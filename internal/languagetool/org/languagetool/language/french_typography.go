package language

import (
	"regexp"
	"strconv"
	"strings"
)

// French quote characters (French.getOpening/Closing*).
const (
	frOpenDouble  = "«"
	frCloseDouble = "»"
	frOpenSingle  = "‘"
	frCloseSingle = "’"
)

var (
	insideSuggestion = regexp.MustCompile(`(?s)<suggestion>(.+?)</suggestion>`)
	aposLetter       = regexp.MustCompile(`(?i)([\p{L}\d-])'([\p{L}«])`)
	nbspace1         = regexp.MustCompile(`\b([a-zA-Z]\.) ([a-zA-Z]\.)`)
	nbspace2         = regexp.MustCompile(`\b([a-zA-Z]\.) `)
	quotedChar       = regexp.MustCompile(` '(.)'`)
	// base Language patterns — use real unicode runes (Go RE2 has no \u escapes)
	typ1 = regexp.MustCompile("([\u202f\u00a0 «\"(])'")
	typ2 = regexp.MustCompile("'([\u202f\u00a0 !?,.;:\" )])")
	typ3 = regexp.MustCompile("‘s\\b([^’])")
	typ4 = regexp.MustCompile(`([ (])"`)
	typ5 = regexp.MustCompile("\"([\u202f\u00a0 !?,.;:)])")

	beforeApos  = `([cjnmtsldCJNMTSLD]|qu|jusqu|lorsqu|puisqu|quoiqu|Qu|Jusqu|Lorsqu|Puisqu|Quoiqu|QU|JUSQU|LORSQU|PUISQU|QUOIQU)`
	beforeApos1 = regexp.MustCompile(`(\b` + beforeApos + `)'`)
	beforeApos2 = regexp.MustCompile(`(\b` + beforeApos + `)’"`)
	beforeApos3 = regexp.MustCompile(`(\b` + beforeApos + `)’'`)

	frPatNBSPSemi  = regexp.MustCompile("\u00a0;")
	frPatNBSPBang  = regexp.MustCompile("\u00a0!")
	frPatNBSPQ     = regexp.MustCompile("\u00a0\\?")
	frPatSemi      = regexp.MustCompile(`;`)
	frPatBang      = regexp.MustCompile(`!`)
	frPatQ         = regexp.MustCompile(`\?`)
	frPatColon     = regexp.MustCompile(`:`)
	frPatCloseGuil = regexp.MustCompile(`»`)
	frPatOpenGuil  = regexp.MustCompile(`«`)
	frDupNBSP      = regexp.MustCompile("\u00a0\u00a0")
	frDupThin      = regexp.MustCompile("\u202f\u202f")
	frDupSpace     = regexp.MustCompile(`  `)
	frNBSPSpace    = regexp.MustCompile("\u00a0 ")
	frSpaceNBSP    = regexp.MustCompile(" \u00a0")
	frSpaceThin    = regexp.MustCompile(" \u202f")
	frThinSpace    = regexp.MustCompile("\u202f ")
)

// GetOpeningDoubleQuote ports French.getOpeningDoubleQuote ("«").
func (v FrenchVariant) GetOpeningDoubleQuote() string { return frOpenDouble }

// GetClosingDoubleQuote ports French.getClosingDoubleQuote ("»").
func (v FrenchVariant) GetClosingDoubleQuote() string { return frCloseDouble }

// GetOpeningSingleQuote ports French.getOpeningSingleQuote ("‘").
func (v FrenchVariant) GetOpeningSingleQuote() string { return frOpenSingle }

// GetClosingSingleQuote ports French.getClosingSingleQuote ("’").
func (v FrenchVariant) GetClosingSingleQuote() string { return frCloseSingle }

// IsAdvancedTypographyEnabled ports French.isAdvancedTypographyEnabled (true).
func (v FrenchVariant) IsAdvancedTypographyEnabled() bool { return true }

// HasMinMatchesRules ports French.hasMinMatchesRules (true).
func (v FrenchVariant) HasMinMatchesRules() bool { return true }

// ToAdvancedTypography ports French.toAdvancedTypography (base Language + FR rules).
func (v FrenchVariant) ToAdvancedTypography(input string) string {
	return frenchAdvancedTypography(input)
}

// FrenchAdvancedTypography is the package-level entry used by twins.
func FrenchAdvancedTypography(input string) string {
	return frenchAdvancedTypography(input)
}

func frenchAdvancedTypography(input string) string {
	output := input

	var preserved []string
	output = insideSuggestion.ReplaceAllStringFunc(output, func(m string) string {
		sub := insideSuggestion.FindStringSubmatch(m)
		if len(sub) < 2 {
			return m
		}
		idx := len(preserved)
		preserved = append(preserved, sub[1])
		return "\\" + strconv.Itoa(idx)
	})

	output = strings.ReplaceAll(output, "...", "…")
	output = nbspace1.ReplaceAllString(output, "$1\u00a0$2")
	output = nbspace2.ReplaceAllString(output, "$1\u00a0")
	output = aposLetter.ReplaceAllString(output, "$1’$2")

	if strings.HasPrefix(output, "'") {
		output = frOpenSingle + output[1:]
	}
	if strings.HasSuffix(output, "'") {
		output = output[:len(output)-1] + frCloseSingle
	}
	output = quotedChar.ReplaceAllString(output, " "+frOpenSingle+"$1"+frCloseSingle)
	output = typ1.ReplaceAllString(output, "$1"+frOpenSingle)
	output = typ2.ReplaceAllString(output, frCloseSingle+"$1")
	output = typ3.ReplaceAllString(output, "’s$1")

	if strings.HasPrefix(output, `"`) {
		output = frOpenDouble + output[1:]
	}
	if strings.HasSuffix(output, `"`) {
		output = output[:len(output)-1] + frCloseDouble
	}
	output = typ4.ReplaceAllString(output, "$1"+frOpenDouble)
	output = typ5.ReplaceAllString(output, frCloseDouble+"$1")

	for i, p := range preserved {
		output = strings.Replace(output, "\\"+strconv.Itoa(i), frOpenDouble+p+frCloseDouble, 1)
	}
	output = strings.ReplaceAll(output, "<suggestion>", frOpenDouble)
	output = strings.ReplaceAll(output, "</suggestion>", frCloseDouble)

	output = beforeApos1.ReplaceAllString(output, "$1’")
	output = beforeApos2.ReplaceAllString(output, "$1’"+frOpenDouble)
	output = beforeApos3.ReplaceAllString(output, "$1’"+frOpenSingle)

	output = frPatNBSPSemi.ReplaceAllString(output, "\u202f;")
	output = frPatNBSPBang.ReplaceAllString(output, "\u202f!")
	output = frPatNBSPQ.ReplaceAllString(output, "\u202f?")
	output = frPatSemi.ReplaceAllString(output, "\u202f;")
	output = frPatBang.ReplaceAllString(output, "\u202f!")
	output = frPatQ.ReplaceAllString(output, "\u202f?")

	output = frPatColon.ReplaceAllString(output, "\u00a0:")
	output = frPatCloseGuil.ReplaceAllString(output, "\u00a0»")
	output = frPatOpenGuil.ReplaceAllString(output, "«\u00a0")

	output = frDupNBSP.ReplaceAllString(output, "\u00a0")
	output = frDupThin.ReplaceAllString(output, "\u202f")
	output = frDupSpace.ReplaceAllString(output, " ")
	output = frNBSPSpace.ReplaceAllString(output, "\u00a0")
	output = frSpaceNBSP.ReplaceAllString(output, "\u00a0")
	output = frSpaceThin.ReplaceAllString(output, "\u202f")
	output = frThinSpace.ReplaceAllString(output, "\u202f")

	return output
}
