package rules

import (
	"fmt"
	"math"
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
	UnitKilogram = UnitDef{ID: "kg", Kind: UnitMass, Factor: 1, Symbol: "kg", Metric: true}
	UnitGram     = UnitDef{ID: "g", Kind: UnitMass, Factor: 1e-3, Symbol: "g", Metric: true}
	UnitPound    = UnitDef{ID: "lb", Kind: UnitMass, Factor: 0.45359237, Symbol: "lb", Metric: false}
	UnitMetre    = UnitDef{ID: "m", Kind: UnitLength, Factor: 1, Symbol: "m", Metric: true}
	UnitKilometre = UnitDef{ID: "km", Kind: UnitLength, Factor: 1e3, Symbol: "km", Metric: true}
	UnitCentimetre = UnitDef{ID: "cm", Kind: UnitLength, Factor: 1e-2, Symbol: "cm", Metric: true}
	UnitMillimetre = UnitDef{ID: "mm", Kind: UnitLength, Factor: 1e-3, Symbol: "mm", Metric: true}
	UnitMile     = UnitDef{ID: "mi", Kind: UnitLength, Factor: 1609.344, Symbol: "mi", Metric: false}
	UnitYard     = UnitDef{ID: "yd", Kind: UnitLength, Factor: 0.9144, Symbol: "yd", Metric: false}
	UnitFeet     = UnitDef{ID: "ft", Kind: UnitLength, Factor: 0.3048, Symbol: "ft", Metric: false}
	UnitInch     = UnitDef{ID: "inch", Kind: UnitLength, Factor: 0.0254, Symbol: "inch", Metric: false}
	UnitLitre    = UnitDef{ID: "l", Kind: UnitVolume, Factor: 1e-3, Symbol: "l", Metric: true}
	UnitMillilitre = UnitDef{ID: "ml", Kind: UnitVolume, Factor: 1e-6, Symbol: "ml", Metric: true}
	UnitCelsius  = UnitDef{ID: "celsius", Kind: UnitTemperature, Factor: 1, Offset: 0, Symbol: "°C", Metric: true}
	UnitFahrenheit = UnitDef{ID: "fahrenheit", Kind: UnitTemperature, Factor: 5.0 / 9.0, Offset: -32 * 5.0 / 9.0, Symbol: "°F", Metric: false}
	UnitKmh      = UnitDef{ID: "kmh", Kind: UnitSpeed, Factor: 1000.0 / 3600.0, Symbol: "km/h", Metric: true}
	UnitMph      = UnitDef{ID: "mph", Kind: UnitSpeed, Factor: 1609.344 / 3600.0, Symbol: "mph", Metric: false}
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
	unitNumberRegex = `(-?\d{1,32}(?:[.,]\d{1,32})?)`
	unitWSLimit     = 5
	unitMaxSuggestions = 5
	unitDelta       = 1e-2
)

// AbstractUnitConversionRule ports org.languagetool.rules.AbstractUnitConversionRule
// without javax.measure — uses fixed SI conversion factors.
type AbstractUnitConversionRule struct {
	ID           string
	Messages     map[string]string
	unitPatterns []unitPattern
	metricUnits  []UnitDef
	antiPatterns []*regexp.Regexp
}

type unitPattern struct {
	re   *regexp.Regexp
	unit UnitDef
}

func NewAbstractUnitConversionRule(messages map[string]string) *AbstractUnitConversionRule {
	r := &AbstractUnitConversionRule{
		ID:       "UNIT_CONVERSION",
		Messages: messages,
		antiPatterns: []*regexp.Regexp{
			regexp.MustCompile(`\s?\d+'\d{3}\s?`),
			regexp.MustCompile(`\d+[-‐–]\d+`),
			regexp.MustCompile(`\d+/\d+`),
			regexp.MustCompile(`\d+:\d+`),
			regexp.MustCompile(`\d+⁄\d+`),
		},
	}
	// default unit registrations (subset of Java defaults)
	r.AddUnit(`kg`, UnitKilogram, "kg", 1, true)
	r.AddUnit(`g`, UnitGram, "g", 1, true)
	r.AddUnit(`lb`, UnitPound, "lb", 1, false)
	r.AddUnit(`mi`, UnitMile, "mi", 1, false)
	r.AddUnit(`yd`, UnitYard, "yd", 1, false)
	// RE2 has no lookahead; use simple unit tokens (Java uses negative lookahead).
	r.AddUnit(`ft`, UnitFeet, "ft", 1, false)
	r.AddUnit(`inch`, UnitInch, "inch", 1, false)
	r.AddUnit(`(?:km/h|kmh)`, UnitKmh, "km/h", 1, true)
	r.AddUnit(`mph`, UnitMph, "mph", 1, false)
	r.AddUnit(`km`, UnitKilometre, "km", 1, true)
	r.AddUnit(`m`, UnitMetre, "m", 1, true)
	r.AddUnit(`cm`, UnitCentimetre, "cm", 1, true)
	r.AddUnit(`mm`, UnitMillimetre, "mm", 1, true)
	r.AddUnit(`l`, UnitLitre, "l", 1, true)
	r.AddUnit(`ml`, UnitMillilitre, "ml", 1, true)
	r.AddUnit(`°F`, UnitFahrenheit, "°F", 1, false)
	r.AddUnit(`°C`, UnitCelsius, "°C", 1, true)
	return r
}

// AddUnit registers a unit pattern (Java addUnit).
// pattern is the unit body; number + whitespace are prepended.
func (r *AbstractUnitConversionRule) AddUnit(pattern string, base UnitDef, symbol string, factor float64, metric bool) {
	u := base
	u.Factor = base.Factor * factor
	u.Symbol = symbol
	u.Metric = metric
	ws := fmt.Sprintf(`[ \x{00A0}]{0,%d}`, unitWSLimit)
	re := regexp.MustCompile(`(?i)` + unitNumberRegex + ws + pattern + `\b`)
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

func (r *AbstractUnitConversionRule) GetMessage(m UnitConversionMessage) string {
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

func naturalness(v float64) float64 {
	av := math.Abs(v)
	if av == 0 {
		return math.Inf(1)
	}
	// prefer values near 1..100
	if av >= 1 && av <= 100 {
		return 0
	}
	return math.Abs(math.Log10(av) - 1)
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
	for _, ap := range r.antiPatterns {
		if ap.MatchString(text) {
			// soft anti-pattern: still try, but Java skips whole spans — skip simple whole-match cases
			_ = ap
		}
	}
	var matches []*RuleMatch
	seen := map[string]struct{}{}
	for _, up := range r.unitPatterns {
		if up.unit.Metric {
			continue
		}
		locs := up.re.FindAllStringSubmatchIndex(text, -1)
		for _, loc := range locs {
			if len(loc) < 4 {
				continue
			}
			full := text[loc[0]:loc[1]]
			numStr := text[loc[2]:loc[3]]
			numStr = strings.ReplaceAll(numStr, ",", ".")
			val, err := strconv.ParseFloat(numStr, 64)
			if err != nil {
				continue
			}
			key := fmt.Sprintf("%d:%d", loc[0], loc[1])
			if _, ok := seen[key]; ok {
				continue
			}
			// anti-pattern span check
			skip := false
			for _, ap := range r.antiPatterns {
				if ap.MatchString(full) {
					skip = true
					break
				}
			}
			if skip {
				continue
			}
			equivs := r.GetMetricEquivalent(val, up.unit)
			if len(equivs) == 0 {
				continue
			}
			seen[key] = struct{}{}
			var suggs []string
			for i, eq := range equivs {
				if i >= unitMaxSuggestions {
					break
				}
				conv := formatUnitNumber(eq.Value) + " " + eq.Unit.Symbol
				suggs = append(suggs, r.FormatSuggestion(strings.TrimSpace(full), conv))
			}
			m := NewRuleMatch(r, sentence, loc[0], loc[1], r.GetMessage(UnitMsgSuggestion))
			m.SetSuggestedReplacements(suggs)
			matches = append(matches, m)
		}
	}
	return matches, nil
}
