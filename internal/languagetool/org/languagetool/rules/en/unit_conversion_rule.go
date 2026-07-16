package en

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// UnitConversionRule is a simplified surface stand-in for EN UnitConversionRule.
// Variant selects imperial (UK) vs US pint sizes and metre/meter spelling in messages.
type UnitConversionRule struct {
	Messages map[string]string
	// Variant: "" general, "imperial", "us"
	Variant string
}

func NewUnitConversionRule(messages map[string]string) *UnitConversionRule {
	return &UnitConversionRule{Messages: messages}
}

func NewUnitConversionRuleImperial(messages map[string]string) *UnitConversionRule {
	return &UnitConversionRule{Messages: messages, Variant: "imperial"}
}

func NewUnitConversionRuleUS(messages map[string]string) *UnitConversionRule {
	return &UnitConversionRule{Messages: messages, Variant: "us"}
}

func (r *UnitConversionRule) GetID() string {
	switch r.Variant {
	case "imperial":
		return "METRIC_UNITS_EN_IMPERIAL"
	case "us":
		return "METRIC_UNITS_EN_US"
	default:
		return "METRIC_UNITS_EN_GENERAL"
	}
}

var enNumberRE = regexp.MustCompile(`^-?\d{1,3}(,\d{3})*(\.\d+)?$|^-?\d+(\.\d+)?$`)

func (r *UnitConversionRule) metreWord() string {
	if r.Variant == "us" {
		return "meters"
	}
	if r.Variant == "imperial" {
		return "metres"
	}
	return "m"
}

func (r *UnitConversionRule) pintLiters() float64 {
	if r.Variant == "us" {
		return 0.473176473 // US liquid pint
	}
	// UK imperial pint (also used for imperial variant)
	return 0.56826125
}

func (r *UnitConversionRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	tokens := sentence.GetTokensWithoutWhitespace()
	var matches []*rules.RuleMatch
	for i := 1; i < len(tokens)-1; i++ {
		numTok := tokens[i]
		numStr := numTok.GetToken()
		if !enNumberRE.MatchString(numStr) {
			continue
		}
		// skip ranges like 3-5
		if i+1 < len(tokens) && tokens[i+1].GetToken() == "-" {
			continue
		}
		// skip fractions 1/4
		if i+1 < len(tokens) && tokens[i+1].GetToken() == "/" {
			continue
		}
		unitTok := tokens[i+1]
		unit := strings.ToLower(unitTok.GetToken())
		val, err := parseENFloat(numStr)
		if err != nil {
			continue
		}
		var metricVal float64
		var metricUnit string
		spanEnd := unitTok
		full := numStr + " " + unitTok.GetToken()

		if unit == "degrees" || unit == "degree" {
			if i+2 < len(tokens) && strings.EqualFold(tokens[i+2].GetToken(), "Fahrenheit") {
				metricVal = (val - 32) * 5 / 9
				metricUnit = "°C"
				spanEnd = tokens[i+2]
				full = numStr + " degrees Fahrenheit"
			} else {
				continue
			}
		} else if unit == "°" && i+2 < len(tokens) && strings.EqualFold(tokens[i+2].GetToken(), "F") {
			metricVal = (val - 32) * 5 / 9
			metricUnit = "°C"
			spanEnd = tokens[i+2]
			full = numStr + " °F"
		} else if unit == "°f" || unit == "℉" {
			metricVal = (val - 32) * 5 / 9
			metricUnit = "°C"
		} else {
			switch unit {
			case "feet", "foot", "ft":
				metricVal = val * 0.3048
				metricUnit = r.metreWord()
				if metricUnit == "m" {
					// keep short for general
				}
			case "miles", "mile", "mi":
				metricVal = val * 1.609344
				metricUnit = "km"
			case "yards", "yard":
				metricVal = val * 0.9144
				metricUnit = "m"
			case "inches", "inch":
				metricVal = val * 0.0254
				metricUnit = "m"
			case "pounds", "pound", "lbs", "lb":
				kg := val * 0.45359237
				if kg >= 1000 {
					metricVal = kg / 1000
					metricUnit = "t"
				} else {
					metricVal = kg
					metricUnit = "kg"
				}
			case "pints", "pint":
				metricVal = val * r.pintLiters()
				metricUnit = "l"
			case "ounces", "ounce", "oz":
				metricVal = val * 28.349523125
				metricUnit = "g"
			case "sq":
				if i+2 < len(tokens) && strings.EqualFold(tokens[i+2].GetToken(), "ft") {
					metricVal = val * 0.09290304
					metricUnit = "m²"
					spanEnd = tokens[i+2]
					full = numStr + " sq ft"
				} else {
					continue
				}
			case "kilometers", "kilometres", "kilometer", "kilometre", "km":
				// metric → imperial suggestion (US/imperial variants)
				if r.Variant == "" {
					continue
				}
				metricVal = val / 1.609344
				metricUnit = "miles"
				// if nearby already has miles with approx, skip
				if hasNearbyImperialOK(tokens, i+2, metricVal) {
					continue
				}
				// wrong conversion still flags (16 km (10 miles) is slightly off → 1 match in Java)
				sug := formatEN(metricVal) + " " + metricUnit
				msg := "Corresponds to about " + sug + "."
				rm := rules.NewRuleMatch(r, sentence, numTok.GetStartPos(), spanEnd.GetEndPos(), msg)
				rm.ShortMessage = "unit conversion"
				rm.SetSuggestedReplacement(full + " (" + sug + ")")
				matches = append(matches, rm)
				continue
			default:
				continue
			}
		}
		if hasNearbyMetricEN(tokens, i+2) {
			// still flag if feet with wrong m value? Java flags "6 feet (2.02 m)"
			// only skip if value is roughly correct
			if !nearbyMetricRoughlyCorrect(tokens, i+2, metricVal) {
				// continue to emit
			} else {
				continue
			}
		}
		sug := formatEN(metricVal) + " " + metricUnit
		msg := "Corresponds to about " + sug + "."
		rm := rules.NewRuleMatch(r, sentence, numTok.GetStartPos(), spanEnd.GetEndPos(), msg)
		rm.ShortMessage = "unit conversion"
		rm.SetSuggestedReplacement(full + " (" + sug + ")")
		matches = append(matches, rm)
	}
	return matches
}

func parseENFloat(s string) (float64, error) {
	s = strings.ReplaceAll(s, ",", "")
	return strconv.ParseFloat(s, 64)
}

func formatEN(f float64) string {
	s := fmt.Sprintf("%.2f", f)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s
}

func hasNearbyMetricEN(tokens []*languagetool.AnalyzedTokenReadings, from int) bool {
	limit := from + 8
	if limit > len(tokens) {
		limit = len(tokens)
	}
	for j := from; j < limit; j++ {
		t := strings.ToLower(tokens[j].GetToken())
		switch t {
		case "m", "km", "kg", "t", "g", "m²", "°c", "celsius", "l", "metres", "meters", "litres", "liters":
			return true
		}
	}
	return false
}

// hasNearbyImperialOK reports whether a roughly-correct imperial conversion
// is already present (or explicitly approximate with ca./about).
func hasNearbyImperialOK(tokens []*languagetool.AnalyzedTokenReadings, from int, expectedMiles float64) bool {
	limit := from + 10
	if limit > len(tokens) {
		limit = len(tokens)
	}
	for j := from; j < limit; j++ {
		t := strings.ToLower(tokens[j].GetToken())
		switch t {
		case "miles", "mile", "mi", "feet", "foot", "ft", "yards", "inches":
			// approximate markers
			for k := from; k < j; k++ {
				prev := strings.ToLower(tokens[k].GetToken())
				if prev == "ca" || prev == "ca." || prev == "approx" || prev == "about" || prev == "≈" {
					return true
				}
			}
			// find number before unit
			if j > 0 && enNumberRE.MatchString(tokens[j-1].GetToken()) {
				v, err := parseENFloat(tokens[j-1].GetToken())
				if err == nil && expectedMiles != 0 && absFloat((v-expectedMiles)/expectedMiles) < 0.02 {
					return true
				}
			}
			// present but wrong/unparsed — not OK
			return false
		}
	}
	return false
}

func nearbyMetricRoughlyCorrect(tokens []*languagetool.AnalyzedTokenReadings, from int, expected float64) bool {
	limit := from + 10
	if limit > len(tokens) {
		limit = len(tokens)
	}
	for j := from; j < limit; j++ {
		if !enNumberRE.MatchString(tokens[j].GetToken()) {
			continue
		}
		// reassemble "1" "." "82" → 1.82
		num := tokens[j].GetToken()
		if j+2 < limit && tokens[j+1].GetToken() == "." && enNumberRE.MatchString(tokens[j+2].GetToken()) {
			num = num + "." + tokens[j+2].GetToken()
		}
		v, err := parseENFloat(num)
		if err != nil {
			continue
		}
		// within 3% (tokenizer/rounding slack)
		if expected != 0 && absFloat((v-expected)/expected) < 0.03 {
			return true
		}
	}
	return false
}

func absFloat(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
