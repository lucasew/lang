package de

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// UnitConversionRule is a simplified surface stand-in for DE UnitConversionRule.
// Converts a few imperial units to metric without the full unit library.
type UnitConversionRule struct {
	Messages map[string]string
}

func NewUnitConversionRule(messages map[string]string) *UnitConversionRule {
	return &UnitConversionRule{Messages: messages}
}

func (r *UnitConversionRule) GetID() string { return "UNITS_DE" }

// Match scans tokens for number + unit pairs.
func (r *UnitConversionRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	tokens := sentence.GetTokensWithoutWhitespace()
	var matches []*rules.RuleMatch
	for i := 1; i < len(tokens)-1; i++ {
		numTok := tokens[i]
		unitTok := tokens[i+1]
		numStr := numTok.GetToken()
		if !looksLikeNumberDE(numStr) {
			continue
		}
		unit := unitTok.GetToken()
		unitLC := strings.ToLower(unit)
		unitLC = strings.ReplaceAll(unitLC, "ß", "ss")
		val, err := parseDEFloat(numStr)
		if err != nil {
			continue
		}
		var metricVal float64
		var metricUnit string
		spanEnd := unitTok
		full := numStr + " " + unit
		switch {
		case unitLC == "fuss" || unitLC == "ft":
			metricVal = val * 0.3048
			metricUnit = "Meter"
		case unitLC == "meile" || unitLC == "meilen":
			metricVal = val * 1.609344
			metricUnit = "Kilometer"
		case unitLC == "pfund":
			if i+2 < len(tokens) && strings.EqualFold(tokens[i+2].GetToken(), "Sterling") {
				continue
			}
			kg := val * 0.45359237
			if kg >= 1000 {
				metricVal = kg / 1000
				metricUnit = "Tonnen"
			} else {
				metricVal = kg
				metricUnit = "Kilogramm"
			}
		case unitLC == "sq" && i+2 < len(tokens) && strings.EqualFold(tokens[i+2].GetToken(), "ft"):
			metricVal = val * 0.09290304
			metricUnit = "Quadratmeter"
			spanEnd = tokens[i+2]
			full = numStr + " sq ft"
		default:
			continue
		}
		if hasNearbyMetric(tokens, i+2) {
			continue
		}
		sug := formatDE(metricVal) + " " + metricUnit
		msg := "Entspricht ca. " + sug + "."
		rm := rules.NewRuleMatch(r, sentence, numTok.GetStartPos(), spanEnd.GetEndPos(), msg)
		rm.ShortMessage = "Einheit"
		rm.SetSuggestedReplacement(full + " (" + sug + ")")
		matches = append(matches, rm)
	}
	return matches
}

var deNumberRE = regexp.MustCompile(`^\d{1,3}(\.\d{3})*(,\d+)?$|^\d+(,\d+)?$`)

func looksLikeNumberDE(s string) bool {
	return deNumberRE.MatchString(s)
}

func hasNearbyMetric(tokens []*languagetool.AnalyzedTokenReadings, from int) bool {
	limit := from + 8
	if limit > len(tokens) {
		limit = len(tokens)
	}
	for j := from; j < limit; j++ {
		t := strings.ToLower(tokens[j].GetToken())
		t = strings.ReplaceAll(t, "ß", "ss")
		switch t {
		case "m", "meter", "km", "kilometer", "tonne", "tonnen", "quadratmeter", "kg", "kilogramm":
			return true
		}
	}
	return false
}

func parseDEFloat(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if strings.Contains(s, ",") && strings.Contains(s, ".") {
		s = strings.ReplaceAll(s, ".", "")
		s = strings.ReplaceAll(s, ",", ".")
	} else if strings.Contains(s, ",") {
		s = strings.ReplaceAll(s, ",", ".")
	}
	return strconv.ParseFloat(s, 64)
}

func formatDE(f float64) string {
	s := fmt.Sprintf("%.2f", f)
	s = strings.ReplaceAll(s, ".", ",")
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ",")
	return s
}
