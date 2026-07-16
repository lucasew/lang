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
type UnitConversionRule struct {
	Messages map[string]string
}

func NewUnitConversionRule(messages map[string]string) *UnitConversionRule {
	return &UnitConversionRule{Messages: messages}
}

func (r *UnitConversionRule) GetID() string { return "METRIC_UNITS_EN_GENERAL" }

var enNumberRE = regexp.MustCompile(`^-?\d{1,3}(,\d{3})*(\.\d+)?$|^-?\d+(\.\d+)?$`)

func (r *UnitConversionRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	tokens := sentence.GetTokensWithoutWhitespace()
	var matches []*rules.RuleMatch
	for i := 1; i < len(tokens)-1; i++ {
		numTok := tokens[i]
		numStr := numTok.GetToken()
		if !enNumberRE.MatchString(numStr) {
			continue
		}
		// skip years like 1989's handled by next token check
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

		// degrees Fahrenheit: "100 degrees Fahrenheit" or "100 °F"
		if unit == "degrees" || unit == "degree" {
			if i+2 < len(tokens) && strings.EqualFold(tokens[i+2].GetToken(), "Fahrenheit") {
				metricVal = (val - 32) * 5 / 9
				metricUnit = "°C"
				spanEnd = tokens[i+2]
				full = numStr + " degrees Fahrenheit"
			} else {
				continue
			}
		} else if unit == "°f" || unit == "℉" || (unit == "f" && i > 0 && strings.Contains(tokens[i].GetToken(), "")) {
			// "100 °F" may tokenize as 100, °F or 100, °, F
			if unit == "°f" || unit == "℉" {
				metricVal = (val - 32) * 5 / 9
				metricUnit = "°C"
			} else if unit == "°" && i+2 < len(tokens) && strings.EqualFold(tokens[i+2].GetToken(), "F") {
				metricVal = (val - 32) * 5 / 9
				metricUnit = "°C"
				spanEnd = tokens[i+2]
				full = numStr + " °F"
			} else {
				// fall through to other units
				goto other
			}
		} else {
			goto other
		}
		goto emit
	other:
		switch unit {
		case "feet", "foot", "ft":
			metricVal = val * 0.3048
			metricUnit = "m"
		case "miles", "mile":
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
		default:
			continue
		}
	emit:
		if hasNearbyMetricEN(tokens, i+2) {
			continue
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
		case "m", "km", "kg", "t", "g", "m²", "°c", "celsius":
			return true
		}
		// "1.82 m"
		if j+1 < limit {
			u := strings.ToLower(tokens[j+1].GetToken())
			if u == "m" || u == "km" || u == "kg" || u == "t" {
				return true
			}
		}
	}
	return false
}
