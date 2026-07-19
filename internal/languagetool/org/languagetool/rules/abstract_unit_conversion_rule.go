package rules

import (
	"fmt"
	"math"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// UnitKind identifies dimension for compatibility checks.
type UnitKind int

const (
	UnitMass UnitKind = iota
	UnitLength
	UnitArea
	UnitVolume
	UnitTemperature
	UnitSpeed
)

// UnitDef is a measurement unit with SI conversion (value_si = value * Factor + Offset).
// Temperature uses Offset (e.g. Fahrenheit: (F-32)*5/9 → Celsius as SI base).
type UnitDef struct {
	ID     string
	Kind   UnitKind
	Factor float64 // multiply source value to reach SI base
	Offset float64 // added after multiply for temperature
	Symbol string
	Metric bool
}

// SI conversion bases: mass kg, length m, area m², volume m³, temp °C, speed m/s.
var (
	UnitKilogram   = UnitDef{ID: "kg", Kind: UnitMass, Factor: 1, Symbol: "kg", Metric: true}
	UnitGram       = UnitDef{ID: "g", Kind: UnitMass, Factor: 1e-3, Symbol: "g", Metric: true}
	UnitPound      = UnitDef{ID: "lb", Kind: UnitMass, Factor: 0.45359237, Symbol: "lb", Metric: false}
	UnitMetre      = UnitDef{ID: "m", Kind: UnitLength, Factor: 1, Symbol: "m", Metric: true}
	UnitKilometre  = UnitDef{ID: "km", Kind: UnitLength, Factor: 1e3, Symbol: "km", Metric: true}
	UnitCentimetre = UnitDef{ID: "cm", Kind: UnitLength, Factor: 1e-2, Symbol: "cm", Metric: true}
	UnitMillimetre = UnitDef{ID: "mm", Kind: UnitLength, Factor: 1e-3, Symbol: "mm", Metric: true}
	UnitMile       = UnitDef{ID: "mi", Kind: UnitLength, Factor: 1609.344, Symbol: "mi", Metric: false}
	UnitYard       = UnitDef{ID: "yd", Kind: UnitLength, Factor: 0.9144, Symbol: "yd", Metric: false}
	UnitFeet       = UnitDef{ID: "ft", Kind: UnitLength, Factor: 0.3048, Symbol: "ft", Metric: false}
	UnitInch       = UnitDef{ID: "inch", Kind: UnitLength, Factor: 0.0254, Symbol: "inch", Metric: false}
	UnitLitre      = UnitDef{ID: "l", Kind: UnitVolume, Factor: 1e-3, Symbol: "l", Metric: true}
	UnitMillilitre = UnitDef{ID: "ml", Kind: UnitVolume, Factor: 1e-6, Symbol: "ml", Metric: true}
	UnitCelsius    = UnitDef{ID: "celsius", Kind: UnitTemperature, Factor: 1, Offset: 0, Symbol: "°C", Metric: true}
	UnitFahrenheit = UnitDef{ID: "fahrenheit", Kind: UnitTemperature, Factor: 5.0 / 9.0, Offset: -32 * 5.0 / 9.0, Symbol: "°F", Metric: false}
	UnitKmh        = UnitDef{ID: "kmh", Kind: UnitSpeed, Factor: 1000.0 / 3600.0, Symbol: "km/h", Metric: true}
	UnitMph        = UnitDef{ID: "mph", Kind: UnitSpeed, Factor: 1609.344 / 3600.0, Symbol: "mph", Metric: false}
	// Area / tonne (DE UnitConversionRule / AbstractUnitConversionRule defaults)
	UnitSquareMetre = UnitDef{ID: "m2", Kind: UnitArea, Factor: 1, Symbol: "m²", Metric: true}
	UnitHectare     = UnitDef{ID: "ha", Kind: UnitArea, Factor: 1e4, Symbol: "ha", Metric: true}
	UnitSqFt        = UnitDef{ID: "sqft", Kind: UnitArea, Factor: 0.09290304, Symbol: "sq ft", Metric: false}
	UnitSqIn        = UnitDef{ID: "sqin", Kind: UnitArea, Factor: 0.0254 * 0.0254, Symbol: "sq in", Metric: false}
	UnitSqYd        = UnitDef{ID: "sqyd", Kind: UnitArea, Factor: 0.9144 * 0.9144, Symbol: "sq yd", Metric: false}
	UnitTonne       = UnitDef{ID: "t", Kind: UnitMass, Factor: 1e3, Symbol: "t", Metric: true}
	// Cubic (Java CUBIC_METRE / ft³ / in³ / yd³)
	UnitCubicMetre = UnitDef{ID: "m3", Kind: UnitVolume, Factor: 1, Symbol: "m³", Metric: true}
	UnitCubicFeet  = UnitDef{ID: "ft3", Kind: UnitVolume, Factor: 0.3048 * 0.3048 * 0.3048, Symbol: "ft³", Metric: false}
	UnitCubicInch  = UnitDef{ID: "in3", Kind: UnitVolume, Factor: 0.0254 * 0.0254 * 0.0254, Symbol: "inch³", Metric: false}
	UnitCubicYard  = UnitDef{ID: "yd3", Kind: UnitVolume, Factor: 0.9144 * 0.9144 * 0.9144, Symbol: "yard³", Metric: false}
)

// UnitConversionMessage ports AbstractUnitConversionRule.Message.
type UnitConversionMessage int

const (
	UnitMsgSuggestion UnitConversionMessage = iota
	UnitMsgCheck
	UnitMsgCheckUnknownUnit
	UnitMsgUnitMismatch
)

const (
	// Java AbstractUnitConversionRule.NUMBER_REGEX / NUMBER_REGEX_WITH_BOUNDARY body
	// (allows thousands separators: 10.000,75 or 10,000.75).
	unitNumberRegex    = `(-?\d{1,32}[\d,.]{0,32})`
	unitWSLimit        = 5
	unitMaxSuggestions = 5
	unitDelta          = 1e-2
	// Java AbstractUnitConversionRule.ROUNDING_DELTA
	unitRoundingDelta = 0.05
)

// convertedParenRE ports AbstractUnitConversionRule.convertedPatterns:
// whitespace + (optional "ca. " / "aprox. ") + number + unit body in parentheses.
var convertedParenRE = regexp.MustCompile(`(?i)^\s*\((?:(?:ca\.|aprox\.)\s*)?` + unitNumberRegex + `\s*([^)]+?)\s*\)`)

// feetInchParenBodyRE ports Java convertedMatcher.group().trim().matches(
// "\\(\\d+ (feet|ft) \\d+ inch\\)") on the unit-body half after the number group.
// e.g. "(2 ft 6 inch)" → unitBody "ft 6 inch" — skip CHECK (would misread as 2 ft).
var feetInchParenBodyRE = regexp.MustCompile(`(?i)^(feet|ft)\s+\d+\s+inch$`)

// numberRangePartRE ports AbstractUnitConversionRule.numberRangePart:
// a number at the end of the text before a match that captured a leading "-".
// Used for ranges like "1-5 miles" (GitHub languagetool#2170).
var numberRangePartRE = regexp.MustCompile(unitNumberRegex + `$`)

// AbstractUnitConversionRule ports org.languagetool.rules.AbstractUnitConversionRule
// without javax.measure — uses fixed SI conversion factors.
// AbstractUnitConversionRule ports org.languagetool.rules.AbstractUnitConversionRule.
// Java: STYLE, Style.
type AbstractUnitConversionRule struct {
	ID           string
	Messages     map[string]string
	// Category ports Rule.category (Java STYLE).
	Category *Category
	// IssueType ports getLocQualityIssueType (Java Style).
	IssueType ITSIssueType
	unitPatterns []unitPattern
	// specialPatterns ports Java specialPatterns (e.g. 5'6" → feet + inches).
	specialPatterns []specialUnitPattern
	metricUnits     []UnitDef
	antiPatterns    []*regexp.Regexp
	// FormatNumber formats converted values (default English-style). DE uses comma decimals.
	FormatNumber func(v float64) string
	// FormatRounded ports formatRounded — prefix for near-integer suggestions (default "ca. ").
	// PT uses "aprox. ".
	FormatRounded func(s string) string
	// MessageFor optional override of GetMessage (language-specific).
	MessageFor func(m UnitConversionMessage) string
	// ShortMessageFor optional override of GetShortMessage (language-specific).
	ShortMessageFor func(m UnitConversionMessage) string
	// ParseNumber optional locale number parse (default: comma→dot only).
	ParseNumber func(s string) (float64, error)
	// incorrectExamples / correctExamples port Rule lists (Java addExamplePair).
	incorrectExamples []IncorrectExample
	correctExamples   []CorrectExample
}

// AddExamplePair ports Rule.addExamplePair for unit conversion subclasses.
func (r *AbstractUnitConversionRule) AddExamplePair(incorrect IncorrectExample, correct CorrectExample) {
	if r == nil {
		return
	}
	appendExamplePair(&r.incorrectExamples, &r.correctExamples, incorrect, correct)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *AbstractUnitConversionRule) GetIncorrectExamples() []IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *AbstractUnitConversionRule) GetCorrectExamples() []CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

type unitPattern struct {
	re   *regexp.Regexp
	unit UnitDef
}

// specialUnitPattern is a full-match unit with custom value parser (groups 1..n).
type specialUnitPattern struct {
	re    *regexp.Regexp
	unit  UnitDef
	parse func(submatches []string) (float64, bool) // full match + groups
}

func NewAbstractUnitConversionRule(messages map[string]string) *AbstractUnitConversionRule {
	r := &AbstractUnitConversionRule{
		ID:        "UNIT_CONVERSION",
		Messages:  messages,
		Category:  CatStyle.GetCategory(messages),
		IssueType: ITSStyle,
		antiPatterns: []*regexp.Regexp{
			regexp.MustCompile(`\s?\d+'\d{3}\s?`),
			regexp.MustCompile(`\d+[-‐–]\d+`),
			regexp.MustCompile(`\d+/\d+`),
			regexp.MustCompile(`\d+:\d+`),
			regexp.MustCompile(`\d+⁄\d+`),
		},
	}
	// default unit registrations (Java AbstractUnitConversionRule constructor; commented-out Java units omitted)
	r.AddUnit(`kg`, UnitKilogram, "kg", 1, true)
	r.AddUnit(`g`, UnitGram, "g", 1, true)
	r.AddUnit(`t`, UnitTonne, "t", 1, true)

	r.AddUnit(`lb`, UnitPound, "lb", 1, false)
	// Java: //addUnit("oz", OUNCE, ...) commented out

	r.AddUnit(`mi`, UnitMile, "mi", 1, false)
	r.AddUnit(`yd`, UnitYard, "yd", 1, false)
	// RE2 has no lookahead; use simple unit tokens (Java uses negative lookahead).
	r.AddUnit(`(?:ft|′|')`, UnitFeet, "ft", 1, false)
	r.AddUnit(`(?:inch|″)`, UnitInch, "inch", 1, false)

	r.AddUnit(`(?:km/h|kmh)`, UnitKmh, "km/h", 1, true)
	r.AddUnit(`mph`, UnitMph, "mph", 1, false)

	r.AddUnit(`km`, UnitMetre, "km", 1e3, true)
	r.AddUnit(`m`, UnitMetre, "m", 1, true)
	// Java: //addUnit("dm", ...) commented out
	r.AddUnit(`cm`, UnitMetre, "cm", 1e-2, true)
	r.AddUnit(`mm`, UnitMetre, "mm", 1e-3, true)
	r.AddUnit(`µm`, UnitMetre, "µm", 1e-6, true)
	r.AddUnit(`nm`, UnitMetre, "nm", 1e-9, true)

	r.AddUnit(`m(?:\^2|2|²)`, UnitSquareMetre, "m²", 1, true)
	r.AddUnit(`ha`, UnitSquareMetre, "ha", 1e4, true)
	r.AddUnit(`a`, UnitSquareMetre, "a", 1e2, true)
	r.AddUnit(`km(?:\^2|2|²)`, UnitSquareMetre, "km²", 1e6, true)
	r.AddUnit(`cm(?:\^2|2|²)`, UnitSquareMetre, "cm²", 1e-4, true)
	r.AddUnit(`mm(?:\^2|2|²)`, UnitSquareMetre, "mm²", 1e-6, true)
	r.AddUnit(`µm(?:\^2|2|²)`, UnitSquareMetre, "µm²", 1e-12, true)
	r.AddUnit(`nm(?:\^2|2|²)`, UnitSquareMetre, "nm²", 1e-18, true)

	r.AddUnit(`(?:sq|square) (?:in(?:ch)?|inches)`, UnitSqIn, "sq in", 1, false)
	r.AddUnit(`(?:inches|in|inch) (?:\^2|2|²)`, UnitSqIn, "in²", 1, false)

	r.AddUnit(`(?:sq|square) (?:ft|feet|foot)`, UnitSqFt, "sq ft", 1, false)
	r.AddUnit(`sf`, UnitSqFt, "sf", 1, false)
	r.AddUnit(`ft(?:\^2|2|²)`, UnitSqFt, "ft²", 1, false)

	r.AddUnit(`(?:sq|square) (?:yds?|yards?)`, UnitSqYd, "sq yd", 1, false)
	r.AddUnit(`(?:yards?|yds?)(?:\^2|2|²)`, UnitSqYd, "yd²", 1, false)

	r.AddUnit(`m(?:\^3|3|³)`, UnitCubicMetre, "m³", 1, true)
	r.AddUnit(`km(?:\^3|3|³)`, UnitCubicMetre, "km³", 1e9, true)
	r.AddUnit(`cm(?:\^3|3|³)`, UnitCubicMetre, "cm³", 1e-6, true)
	r.AddUnit(`mm(?:\^3|3|³)`, UnitCubicMetre, "mm³", 1e-9, true)
	r.AddUnit(`µm(?:\^3|3|³)`, UnitCubicMetre, "µm³", 1e-18, true)
	r.AddUnit(`nm(?:\^3|3|³)`, UnitCubicMetre, "nm³", 1e-27, true)

	r.AddUnit(`(?:cubic|cu) (?:feet|ft|foot)`, UnitCubicFeet, "cubic feet", 1, false)
	r.AddUnit(`(?:feet|ft|foot)(?:\^3|3|³)`, UnitCubicFeet, "ft³", 1, false)

	r.AddUnit(`(?:cubic|cu) (?:inch|in|inches)`, UnitCubicInch, "cubic inch", 1, false)
	r.AddUnit(`(?:inch|in)(?:\^3|3|³)`, UnitCubicInch, "inch³", 1, false)

	r.AddUnit(`(?:cubic|cu) (?:yards?|yds?)`, UnitCubicYard, "cubic yard", 1, false)
	r.AddUnit(`(?:yard|yd)(?:\^3|3|³)`, UnitCubicYard, "yard³", 1, false)

	r.AddUnit(`l`, UnitLitre, "l", 1, true)
	r.AddUnit(`ml`, UnitLitre, "ml", 1e-3, true)

	r.AddUnit(`°F`, UnitFahrenheit, "°F", 1, false)
	r.AddUnit(`°C`, UnitCelsius, "°C", 1, true)

	// Java specialPatterns: 5'6" / 5ft 6in → feet + inches as FEET value
	// RE2 has no lookbehind; approximate with leading char / start of string.
	parseFeetInch := func(subs []string) (float64, bool) {
		// subs[0]=full, [1]=feet, [2]=inch (optional empty)
		if len(subs) < 2 {
			return 0, false
		}
		feet, err := r.parseNumber(subs[1])
		if err != nil {
			return 0, false
		}
		inch := 0.0
		if len(subs) > 2 && subs[2] != "" {
			if v, err2 := r.parseNumber(subs[2]); err2 == nil {
				inch = v
			}
		}
		return feet + inch/12.0, true
	}
	// with leading whitespace: " 5'6" or " 5ft 6"
	r.specialPatterns = append(r.specialPatterns, specialUnitPattern{
		re:    regexp.MustCompile(`(?:^|[^º°\d])\s(\d+)(?:ft|′|')\s*(\d+)\s*(?:in|"|″)?`),
		unit:  UnitFeet,
		parse: parseFeetInch,
	})
	// no leading space after non-digit non-space: "ist5'6"
	r.specialPatterns = append(r.specialPatterns, specialUnitPattern{
		re:    regexp.MustCompile(`(?:^|[^º°\d\s])(\d+)(?:ft|′|')\s*(\d+)\s*(?:in|"|″)?`),
		unit:  UnitFeet,
		parse: parseFeetInch,
	})
	return r
}

// AntiPatternsAppend adds a full-span anti-pattern (e.g. "Pfund Sterling").
func (r *AbstractUnitConversionRule) AntiPatternsAppend(pattern string) {
	if r == nil || pattern == "" {
		return
	}
	r.antiPatterns = append(r.antiPatterns, regexp.MustCompile(`(?i)`+pattern))
}

// AddUnit registers a unit pattern (Java addUnit).
// pattern is the unit body; number + whitespace are prepended.
func (r *AbstractUnitConversionRule) AddUnit(pattern string, base UnitDef, symbol string, factor float64, metric bool) {
	u := base
	u.Factor = base.Factor * factor
	u.Symbol = symbol
	u.Metric = metric
	ws := fmt.Sprintf(`[ \x{00A0}]{0,%d}`, unitWSLimit)
	// Use \p{L} boundaries — ASCII \b fails after non-ASCII letters (e.g. German ß in Fuß).
	re := regexp.MustCompile(`(?i)` + unitNumberRegex + ws + `(?:` + pattern + `)(?:[^\p{L}]|$)`)
	r.unitPatterns = append(r.unitPatterns, unitPattern{re: re, unit: u})
	if metric {
		for _, m := range r.metricUnits {
			if m.ID == u.ID && m.Symbol == u.Symbol {
				return
			}
		}
		r.metricUnits = append(r.metricUnits, u)
	}
}

func (r *AbstractUnitConversionRule) GetID() string { return r.ID }

func (r *AbstractUnitConversionRule) GetCategory() *Category {
	if r == nil {
		return nil
	}
	return r.Category
}

func (r *AbstractUnitConversionRule) GetLocQualityIssueType() ITSIssueType {
	if r == nil || r.IssueType == "" {
		return ITSStyle
	}
	return r.IssueType
}

func (r *AbstractUnitConversionRule) GetMessage(m UnitConversionMessage) string {
	if r != nil && r.MessageFor != nil {
		return r.MessageFor(m)
	}
	switch m {
	case UnitMsgCheck:
		return "This unit conversion doesn't seem right. Do you want to correct it automatically?"
	case UnitMsgSuggestion:
		return "Writing for an international audience? Consider adding the metric equivalent."
	case UnitMsgCheckUnknownUnit:
		return "This unit conversion doesn't seem right, unable to recognize the used unit."
	case UnitMsgUnitMismatch:
		return "These units don't seem to be compatible."
	default:
		return "Unit conversion"
	}
}

// GetShortMessage ports AbstractUnitConversionRule.getShortMessage.
func (r *AbstractUnitConversionRule) GetShortMessage(m UnitConversionMessage) string {
	if r != nil && r.ShortMessageFor != nil {
		return r.ShortMessageFor(m)
	}
	switch m {
	case UnitMsgCheck:
		return "Incorrect unit conversion. Correct it?"
	case UnitMsgSuggestion:
		return "Add metric equivalent?"
	case UnitMsgCheckUnknownUnit:
		return "Unknown unit used in conversion."
	case UnitMsgUnitMismatch:
		return "Units incompatible."
	default:
		return "Unit conversion"
	}
}

// newUnitMatch builds a RuleMatch with long + short messages (Java RuleMatch constructor).
func (r *AbstractUnitConversionRule) newUnitMatch(
	sentence *languagetool.AnalyzedSentence, from, to int, msg UnitConversionMessage,
) *RuleMatch {
	m := NewRuleMatch(r, sentence, from, to, r.GetMessage(msg))
	m.SetShortMessage(r.GetShortMessage(msg))
	return m
}

// newUnitMatchWithURL builds a match and sets Wolfram explanation URL for original span text.
func (r *AbstractUnitConversionRule) newUnitMatchWithURL(
	sentence *languagetool.AnalyzedSentence, from, to int, msg UnitConversionMessage, original string,
) *RuleMatch {
	m := r.newUnitMatch(sentence, from, to, msg)
	if u := buildURLForExplanation(original); u != "" {
		m.SetURL(u)
	}
	return m
}

func (r *AbstractUnitConversionRule) formatNumber(v float64) string {
	if r != nil && r.FormatNumber != nil {
		return r.FormatNumber(v)
	}
	return formatUnitNumber(v)
}

// formatRounded ports AbstractUnitConversionRule.formatRounded (default "ca. ").
func (r *AbstractUnitConversionRule) formatRounded(s string) string {
	if r != nil && r.FormatRounded != nil {
		return r.FormatRounded(s)
	}
	return "ca. " + s
}

// formatConversionSuggestion ports Java getFormattedConversions for one unit value:
// optional near-integer rounded form + exact formatNumber form.
func (r *AbstractUnitConversionRule) formatConversionSuggestion(value float64, symbol string) []string {
	var out []string
	rounded := math.Round(value)
	if rounded != 0 && math.Abs(value) > 0 {
		if math.Abs(value-rounded)/math.Abs(value) < unitRoundingDelta {
			out = append(out, r.formatRounded(r.formatNumber(rounded)+" "+symbol))
		}
	}
	num := r.formatNumber(value)
	if num != "0" {
		out = append(out, num+" "+symbol)
	}
	return out
}

func (r *AbstractUnitConversionRule) parseNumber(s string) (float64, error) {
	if r != nil && r.ParseNumber != nil {
		return r.ParseNumber(s)
	}
	s = strings.ReplaceAll(s, ",", ".")
	return strconv.ParseFloat(s, 64)
}

// detectNumberRange ports AbstractUnitConversionRule.detectNumberRange.
// True when numStr is a negative capture that is actually the end of a range (e.g. "1-5 miles").
func detectNumberRange(text string, matchStart int, numStr string) bool {
	if !strings.HasPrefix(numStr, "-") || matchStart <= 0 || matchStart > len(text) {
		return false
	}
	return numberRangePartRE.MatchString(text[:matchStart])
}

// adjustRangeNumber strips a spurious leading hyphen from range ends (Java tryConversion).
func adjustRangeNumber(text string, matchStart int, numStr string) string {
	if detectNumberRange(text, matchStart, numStr) {
		return numStr[1:]
	}
	return numStr
}

// ToSI converts a value in unit to SI base.
func ToSI(value float64, u UnitDef) float64 {
	return value*u.Factor + u.Offset
}

// FromSI converts SI base value into unit.
func FromSI(si float64, u UnitDef) float64 {
	if u.Factor == 0 {
		return si
	}
	return (si - u.Offset) / u.Factor
}

// Convert converts value from src to dst when kinds match.
func Convert(value float64, src, dst UnitDef) (float64, bool) {
	if src.Kind != dst.Kind {
		return 0, false
	}
	return FromSI(ToSI(value, src), dst), true
}

// GetMetricEquivalent returns metric conversions sorted by "naturalness" (closeness to 1–100).
func (r *AbstractUnitConversionRule) GetMetricEquivalent(value float64, unit UnitDef) []struct {
	Unit  UnitDef
	Value float64
} {
	if unit.Metric {
		// already metric — no conversion needed for suggestion path
		return nil
	}
	var out []struct {
		Unit  UnitDef
		Value float64
	}
	for _, m := range r.metricUnits {
		if m.Kind != unit.Kind {
			continue
		}
		if m.ID == unit.ID {
			continue
		}
		v, ok := Convert(value, unit, m)
		if !ok {
			continue
		}
		out = append(out, struct {
			Unit  UnitDef
			Value float64
		}{Unit: m, Value: v})
	}
	sort.SliceStable(out, func(i, j int) bool {
		return naturalness(out[i].Value) < naturalness(out[j].Value)
	})
	return out
}

// naturalness ports AbstractUnitConversionRule.sortByNaturalness score
// (smaller score → better). Java:
//
//	abs < 1 → 1/(abs²*2); abs < 100 → abs-50; else abs²
func naturalness(v float64) float64 {
	av := math.Abs(v)
	if av < 1.0 {
		if av == 0 {
			return math.Inf(1)
		}
		return 1.0 / (av * av * 2)
	}
	if av < 100 {
		return av - 50
	}
	return av * av
}

// buildURLForExplanation ports AbstractUnitConversionRule.buildURLForExplanation
// (WolframAlpha "convert … to metric").
func buildURLForExplanation(original string) string {
	if original == "" {
		return ""
	}
	q := url.QueryEscape("convert " + original + " to metric")
	return "http://www.wolframalpha.com/input/?i=" + q
}

// FormatSuggestion builds "original (converted)" style text.
func (r *AbstractUnitConversionRule) FormatSuggestion(original, converted string) string {
	return original + " (" + converted + ")"
}

func formatUnitNumber(v float64) string {
	if math.Abs(v-math.Round(v)) < unitDelta {
		return strconv.FormatInt(int64(math.Round(v)), 10)
	}
	return strconv.FormatFloat(v, 'f', 2, 64)
}

// Match finds non-metric measurements and suggests metric equivalents.
func (r *AbstractUnitConversionRule) Match(sentence *languagetool.AnalyzedSentence) ([]*RuleMatch, error) {
	if r == nil || sentence == nil {
		return nil, nil
	}
	text := sentence.GetText()
	if text == "" {
		return nil, nil
	}
	var matches []*RuleMatch
	// claimed spans: specialPatterns claim first so plain "5ft" does not double-hit "5ft 6in"
	var claimed [][2]int
	overlaps := func(a, b int) bool {
		for _, c := range claimed {
			// exclusive end; overlap if a < c1 && b > c0
			if a < c[1] && b > c[0] {
				return true
			}
		}
		return false
	}
	claim := func(a, b int) { claimed = append(claimed, [2]int{a, b}) }

	// specialPatterns first (Java: longer composite measures like 5'6")
	for _, sp := range r.specialPatterns {
		if sp.re == nil || sp.parse == nil {
			continue
		}
		all := sp.re.FindAllStringSubmatchIndex(text, -1)
		for _, loc := range all {
			if len(loc) < 2 {
				continue
			}
			// Build submatch strings (full + groups)
			var subs []string
			for g := 0; g*2+1 < len(loc); g++ {
				a, b := loc[g*2], loc[g*2+1]
				if a < 0 || b < 0 {
					subs = append(subs, "")
					continue
				}
				subs = append(subs, text[a:b])
			}
			// Trim leading boundary char from full match when pattern includes it
			fullStart, fullEnd := loc[0], loc[1]
			// Prefer numeric start for highlight (skip leading non-digit from boundary)
			from := fullStart
			for from < fullEnd {
				c := text[from]
				if c >= '0' && c <= '9' {
					break
				}
				from++
			}
			if from >= fullEnd {
				from = fullStart
			}
			if overlaps(from, fullEnd) {
				continue
			}
			val, ok := sp.parse(subs)
			if !ok {
				continue
			}
			// Parenthetical conversion after special measure (CHECK path).
			if cm := convertedParenRE.FindStringSubmatchIndex(text[fullEnd:]); cm != nil {
				cFrom := fullEnd + cm[0]
				cTo := fullEnd + cm[1]
				numInParen := text[fullEnd+cm[2] : fullEnd+cm[3]]
				unitBody := strings.TrimSpace(text[fullEnd+cm[4] : fullEnd+cm[5]])
				if given, errG := r.parseNumber(numInParen); errG == nil {
					highlight := text[from:fullEnd]
					if check := r.checkParentheticalConversion(sentence, from, fullEnd, cFrom, cTo, val, sp.unit, given, unitBody, highlight); check != nil {
						claim(from, cTo)
						matches = append(matches, check)
						continue
					}
					claim(from, cTo)
					continue
				}
			}
			if hasNearbyMetricInText(text, fullEnd) {
				continue
			}
			equivs := r.GetMetricEquivalent(val, sp.unit)
			if len(equivs) == 0 {
				continue
			}
			claim(from, fullEnd)
			highlight := text[from:fullEnd]
			var suggs []string
			for _, eq := range equivs {
				for _, conv := range r.formatConversionSuggestion(eq.Value, eq.Unit.Symbol) {
					if len(suggs) >= unitMaxSuggestions {
						break
					}
					suggs = append(suggs, r.FormatSuggestion(strings.TrimSpace(highlight), conv))
				}
				if len(suggs) >= unitMaxSuggestions {
					break
				}
			}
			m := r.newUnitMatchWithURL(sentence, from, fullEnd, UnitMsgSuggestion, strings.TrimSpace(highlight))
			m.SetSuggestedReplacements(suggs)
			matches = append(matches, m)
		}
	}

	// Java match: first metric units (set ignore ranges for "10 km (5 miles)"), then non-metric.
	// metricPass: CHECK parenthetical only (already metric → no SUGGESTION).
	// nonMetricPass: SUGGESTION + CHECK.
	for _, metricPass := range []bool{true, false} {
		for _, up := range r.unitPatterns {
			if up.unit.Metric != metricPass {
				continue
			}
			locs := up.re.FindAllStringSubmatchIndex(text, -1)
			for _, loc := range locs {
				if len(loc) < 4 {
					continue
				}
				full := text[loc[0]:loc[1]]
				numStr := text[loc[2]:loc[3]]
				// Java: range "1-5 miles" may capture "-5"; strip hyphen and convert end only.
				numStr = adjustRangeNumber(text, loc[0], numStr)
				val, err := r.parseNumber(numStr)
				if err != nil {
					continue
				}
				if overlaps(loc[0], loc[1]) {
					continue
				}
				// Java removeAntiPatternMatches: drop when anti-pattern covers match edges.
				if unitHitByAntiPattern(text, loc[0], loc[1], r.antiPatterns) {
					continue
				}
				// Currency: "Pfund Sterling" is not mass.
				if after := trailingContext(text, loc[1]); strings.Contains(strings.ToLower(after), "sterling") {
					continue
				}
				// Existing conversion in parentheses — Java CHECK path.
				if cm := convertedParenRE.FindStringSubmatchIndex(text[loc[1]:]); cm != nil {
					cFrom := loc[1] + cm[0]
					cTo := loc[1] + cm[1]
					numInParen := text[loc[1]+cm[2] : loc[1]+cm[3]]
					unitBody := strings.TrimSpace(text[loc[1]+cm[4] : loc[1]+cm[5]])
					given, errG := r.parseNumber(numInParen)
					if errG == nil {
						if check := r.checkParentheticalConversion(sentence, loc[0], loc[1], cFrom, cTo, val, up.unit, given, unitBody, full); check != nil {
							claim(loc[0], cTo)
							matches = append(matches, check)
							continue
						}
						// conversion present and accepted → claim span (blocks secondary unit)
						claim(loc[0], cTo)
						continue
					}
				}
				if metricPass {
					// already metric, no parenthetical → nothing to suggest
					continue
				}
				// Skip when metric unit already appears nearby.
				if hasNearbyMetricInText(text, loc[1]) {
					continue
				}
				equivs := r.GetMetricEquivalent(val, up.unit)
				if len(equivs) == 0 {
					continue
				}
				claim(loc[0], loc[1])
				var suggs []string
				for _, eq := range equivs {
					for _, conv := range r.formatConversionSuggestion(eq.Value, eq.Unit.Symbol) {
						if len(suggs) >= unitMaxSuggestions {
							break
						}
						suggs = append(suggs, r.FormatSuggestion(strings.TrimSpace(full), conv))
					}
					if len(suggs) >= unitMaxSuggestions {
						break
					}
				}
				m := r.newUnitMatchWithURL(sentence, loc[0], loc[1], UnitMsgSuggestion, strings.TrimSpace(full))
				m.SetSuggestedReplacements(suggs)
				matches = append(matches, m)
			}
		}
	}
	// Java: deduplicate matches with equal start; longer match wins (miles per hour > miles).
	return dedupeUnitMatchesByStart(matches), nil
}

// dedupeUnitMatchesByStart keeps the longest match for each FromPos (Java matchesByStart).
func dedupeUnitMatchesByStart(matches []*RuleMatch) []*RuleMatch {
	if len(matches) <= 1 {
		return matches
	}
	byStart := map[int]*RuleMatch{}
	order := []int{}
	for _, m := range matches {
		if m == nil {
			continue
		}
		if prev, ok := byStart[m.FromPos]; ok {
			if m.ToPos > prev.ToPos {
				byStart[m.FromPos] = m
			}
			continue
		}
		byStart[m.FromPos] = m
		order = append(order, m.FromPos)
	}
	out := make([]*RuleMatch, 0, len(order))
	for _, from := range order {
		out = append(out, byStart[from])
	}
	return out
}

// checkParentheticalConversion ports the Java CHECK branch when a conversion already follows.
// Returns a RuleMatch if the given conversion is wrong; nil if OK or not verifiable.
func (r *AbstractUnitConversionRule) checkParentheticalConversion(
	sentence *languagetool.AnalyzedSentence,
	srcFrom, srcTo, convFrom, convTo int,
	srcVal float64, srcUnit UnitDef,
	given float64, unitBody, originalFull string,
) *RuleMatch {
	// Java: if convertedMatcher.group().trim().matches("\\(\\d+ (feet|ft) \\d+ inch\\)") return;
	// e.g. "(2 ft 6 inch)" would be interpreted as just "2 ft", giving a wrong suggestion.
	if feetInchParenBodyRE.MatchString(strings.TrimSpace(unitBody)) {
		return nil
	}
	// Java: match converted unit against unitPatterns (not only metricUnits).
	// Include non-metric paren units (e.g. "10 km (6.21 mi)").
	var convertedUnit *UnitDef
	bodyLow := strings.ToLower(strings.TrimSpace(unitBody))
	// Prefer longer symbol matches: scan all registered unit patterns.
	for i := range r.unitPatterns {
		u := r.unitPatterns[i].unit
		sym := strings.ToLower(u.Symbol)
		if bodyLow == sym || strings.HasPrefix(bodyLow, sym+" ") || strings.HasPrefix(bodyLow, sym+")") {
			// keep longest symbol when multiple match
			if convertedUnit == nil || len(sym) > len(convertedUnit.Symbol) {
				uu := u
				convertedUnit = &uu
			}
			continue
		}
		// common DE/EN/PT long names used in parenthetical CHECK path
		matched := false
		switch {
		case bodyLow == "m" || bodyLow == "meter" || bodyLow == "metre" || bodyLow == "metern" ||
			strings.HasPrefix(bodyLow, "metro"):
			matched = u.Kind == UnitLength && math.Abs(u.Factor-1) < 1e-12
		case bodyLow == "km" || strings.HasPrefix(bodyLow, "kilometer") || strings.HasPrefix(bodyLow, "kilometre") ||
			strings.HasPrefix(bodyLow, "quilômetro") || strings.HasPrefix(bodyLow, "quilometro"):
			matched = u.Kind == UnitLength && math.Abs(u.Factor-1e3) < 1e-9
		case bodyLow == "cm" || strings.HasPrefix(bodyLow, "zentimeter") || strings.HasPrefix(bodyLow, "centimeter") ||
			strings.HasPrefix(bodyLow, "centímetro") || strings.HasPrefix(bodyLow, "centimetro"):
			matched = u.Kind == UnitLength && math.Abs(u.Factor-1e-2) < 1e-12
		case bodyLow == "mm" || strings.HasPrefix(bodyLow, "millimeter") || strings.HasPrefix(bodyLow, "milímetro") ||
			strings.HasPrefix(bodyLow, "milimetro"):
			matched = u.Kind == UnitLength && math.Abs(u.Factor-1e-3) < 1e-12
		case bodyLow == "kg" || strings.HasPrefix(bodyLow, "kilogramm") || strings.HasPrefix(bodyLow, "kilogram") ||
			strings.HasPrefix(bodyLow, "quilogram"):
			matched = u.Kind == UnitMass && math.Abs(u.Factor-1) < 1e-12
		case strings.HasPrefix(bodyLow, "tonelada") || bodyLow == "t" || strings.HasPrefix(bodyLow, "tonne"):
			matched = u.Kind == UnitMass && math.Abs(u.Factor-1e3) < 1e-9
		case bodyLow == "mi" || strings.HasPrefix(bodyLow, "mile"):
			matched = u.Kind == UnitLength && math.Abs(u.Factor-UnitMile.Factor) < 1e-6
		case bodyLow == "ft" || bodyLow == "feet" || bodyLow == "foot" || bodyLow == "fuß" || bodyLow == "fuss":
			matched = u.Kind == UnitLength && math.Abs(u.Factor-UnitFeet.Factor) < 1e-9
		case bodyLow == "lb" || strings.HasPrefix(bodyLow, "pound") || bodyLow == "pfund" || strings.HasPrefix(bodyLow, "libra"):
			matched = u.Kind == UnitMass && math.Abs(u.Factor-UnitPound.Factor) < 1e-9
		}
		if matched {
			uu := u
			convertedUnit = &uu
			// don't break — prefer exact symbol match above; long-name first hit is ok
			if bodyLow == sym {
				break
			}
		}
	}
	if convertedUnit == nil {
		// unknown unit in paren — Java CHECK_UNKNOWN_UNIT; report on paren span
		m := r.newUnitMatch(sentence, convFrom, convTo, UnitMsgCheckUnknownUnit)
		if u := buildURLForExplanation(originalFull); u != "" {
			m.SetURL(u)
		}
		return m
	}
	// same unit as source → leave alone (Java)
	if convertedUnit.ID == srcUnit.ID && convertedUnit.Symbol == srcUnit.Symbol {
		return nil
	}
	expected, ok := Convert(srcVal, srcUnit, *convertedUnit)
	if !ok {
		m := r.newUnitMatch(sentence, convFrom, convTo, UnitMsgUnitMismatch)
		if u := buildURLForExplanation(originalFull); u != "" {
			m.SetURL(u)
		}
		return m
	}
	if math.Abs(expected-given) <= unitDelta*math.Max(1, math.Abs(expected)) {
		// accurate enough
		return nil
	}
	// also accept if given matches any formatted suggestion string number
	equivs := r.GetMetricEquivalent(srcVal, srcUnit)
	for _, eq := range equivs {
		if eq.Unit.ID == convertedUnit.ID && math.Abs(eq.Value-given) <= unitDelta*math.Max(1, math.Abs(eq.Value)) {
			return nil
		}
	}
	m := r.newUnitMatch(sentence, convFrom, convTo, UnitMsgCheck)
	// suggest corrected number + unit
	m.SetSuggestedReplacement(r.formatNumber(expected) + " " + convertedUnit.Symbol)
	if u := buildURLForExplanation(originalFull); u != "" {
		m.SetURL(u)
	}
	return m
}

func trailingContext(text string, pos int) string {
	if pos < 0 || pos >= len(text) {
		return ""
	}
	end := pos + 24
	if end > len(text) {
		end = len(text)
	}
	return text[pos:end]
}

// unitHitByAntiPattern ports Java removeAntiPatternMatches for one candidate span.
// Anti-patterns run on the full sentence text; a hit drops the unit match when the
// anti-pattern range covers either match edge (Java entry removal conditions).
func unitHitByAntiPattern(text string, from, to int, anti []*regexp.Regexp) bool {
	for _, ap := range anti {
		if ap == nil {
			continue
		}
		all := ap.FindAllStringIndex(text, -1)
		for _, loc := range all {
			if len(loc) < 2 {
				continue
			}
			a, b := loc[0], loc[1]
			// Java: matcher.start() <= from && matcher.end() >= from ||
			//       matcher.start() <= to && matcher.end() >= to
			if (a <= from && b >= from) || (a <= to && b >= to) {
				return true
			}
		}
	}
	return false
}

// hasNearbyMetricInText reports metric unit tokens shortly after pos (parenthetical equivalents).
func hasNearbyMetricInText(text string, pos int) bool {
	if pos < 0 || pos >= len(text) {
		return false
	}
	end := pos + 40
	if end > len(text) {
		end = len(text)
	}
	window := strings.ToLower(text[pos:end])
	// strip non-breaking spaces
	window = strings.ReplaceAll(window, "\u00a0", " ")
	metrics := []string{
		" m)", " m ", " m.", " meter", " metre", " km", "kilometer", "kilometre",
		" kg", "kilogramm", " tonne", " tonnen", " m²", " m2", "quadratmeter",
		" cm", " mm", " °c", " celsius",
	}
	for _, m := range metrics {
		if strings.Contains(window, m) {
			return true
		}
	}
	// bare (1,82 m) style
	if strings.Contains(window, "(") && (strings.Contains(window, " m)") || strings.Contains(window, "m)")) {
		return true
	}
	return false
}
