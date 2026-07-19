package de

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// UnitConversionRule ports org.languagetool.rules.de.UnitConversionRule
// (extends AbstractUnitConversionRule with German unit names and Locale.GERMANY formatting).
// Without javax.measure: fixed SI factors (same approach as rules.AbstractUnitConversionRule).
type UnitConversionRule struct {
	*rules.AbstractUnitConversionRule
}

func NewUnitConversionRule(messages map[string]string) *UnitConversionRule {
	base := rules.NewAbstractUnitConversionRule(messages)
	// Java getId
	base.ID = "EINHEITEN_METRISCH"
	base.FormatNumber = formatDEUnitNumber
	base.ParseNumber = parseDEUnitNumber
	base.MessageFor = func(m rules.UnitConversionMessage) string {
		switch m {
		case rules.UnitMsgCheck:
			return "Diese Umrechnung scheint falsch zu sein. Wollen Sie sie automatisch korrigieren lassen?"
		case rules.UnitMsgSuggestion:
			return "Wollen Sie eine Umwandlung ins metrische System automatisch hinzufügen?"
		case rules.UnitMsgCheckUnknownUnit:
			return "Die in dieser Umrechnung verwendete Einheit wurde nicht erkannt."
		case rules.UnitMsgUnitMismatch:
			return "Diese Einheiten sind nicht kompatibel."
		default:
			return "Einheit"
		}
	}
	// Pfund Sterling anti-pattern (not a mass)
	base.AntiPatternsAppend(`\d+[.,\d]*\s*Pfund\s+Sterling`)

	// Java UnitConversionRule unit registrations (German names + SI factors)
	// Mass
	base.AddUnit(`Kilo(?:gramm)?`, rules.UnitKilogram, "Kilogramm", 1, true)
	base.AddUnit(`Gramm`, rules.UnitKilogram, "Gramm", 1e-3, true)
	base.AddUnit(`Tonnen?`, rules.UnitTonne, "Tonnen", 1, true)
	base.AddUnit(`Pfund`, rules.UnitPound, "Pfund", 1, false)

	// Length
	base.AddUnit(`Meilen?`, rules.UnitMile, "Meile", 1, false)
	base.AddUnit(`Yard`, rules.UnitYard, "Yard", 1, false)
	base.AddUnit(`Fu(?:ß|ss)`, rules.UnitFeet, "Fuß", 1, false)
	base.AddUnit(`Zoll`, rules.UnitInch, "Zoll", 1, false)

	// Speed
	base.AddUnit(`(?:Kilometer pro Stunde|Stundenkilometer)`, rules.UnitKmh, "Kilometer pro Stunde", 1, true)
	base.AddUnit(`Meilen pro Stunde`, rules.UnitMph, "Meilen pro Stunde", 1, false)

	// Metric length names
	base.AddUnit(`Meter`, rules.UnitMetre, "Meter", 1, true)
	base.AddUnit(`Kilometer`, rules.UnitKilometre, "Kilometer", 1, true)
	base.AddUnit(`Zentimeter`, rules.UnitCentimetre, "Zentimeter", 1, true)
	base.AddUnit(`Millimeter`, rules.UnitMillimetre, "Millimeter", 1, true)
	base.AddUnit(`Mikrometer`, rules.UnitMetre, "Mikrometer", 1e-6, true)
	base.AddUnit(`Nanometer`, rules.UnitMetre, "Nanometer", 1e-9, true)
	base.AddUnit(`Pikometer`, rules.UnitMetre, "Pikometer", 1e-12, true)
	base.AddUnit(`Femtometer`, rules.UnitMetre, "Femtometer", 1e-15, true)

	// Area
	base.AddUnit(`Quadratmeter`, rules.UnitSquareMetre, "Quadratmeter", 1, true)
	base.AddUnit(`Hektar`, rules.UnitHectare, "Hektar", 1, true)
	base.AddUnit(`Ar`, rules.UnitSquareMetre, "Ar", 1e2, true)
	base.AddUnit(`Quadratkilometer`, rules.UnitSquareMetre, "Quadratkilometer", 1e6, true)
	base.AddUnit(`Quadratzentimeter`, rules.UnitSquareMetre, "Quadratzentimeter", 1e-4, true)
	base.AddUnit(`Quadratmillimeter`, rules.UnitSquareMetre, "Quadratmillimeter", 1e-6, true)
	base.AddUnit(`sq\s*ft`, rules.UnitSqFt, "sq ft", 1, false)

	// Volume
	base.AddUnit(`Liter`, rules.UnitLitre, "Liter", 1, true)
	base.AddUnit(`Milliliter`, rules.UnitLitre, "Milliliter", 1e-3, true)

	// Temperature
	base.AddUnit(`(?:Grad\s*)?Fahrenheit`, rules.UnitFahrenheit, "Grad Fahrenheit", 1, false)
	base.AddUnit(`(?:Grad\s*)?Celsius`, rules.UnitCelsius, "Grad Celsius", 1, true)

	// English symbols still common
	base.AddUnit(`ft`, rules.UnitFeet, "ft", 1, false)
	base.AddUnit(`mi`, rules.UnitMile, "mi", 1, false)

	return &UnitConversionRule{AbstractUnitConversionRule: base}
}

func (r *UnitConversionRule) GetID() string {
	if r != nil && r.AbstractUnitConversionRule != nil && r.ID != "" {
		return r.ID
	}
	return "EINHEITEN_METRISCH"
}

func (r *UnitConversionRule) GetDescription() string {
	return "Schlägt vor oder überprüft Angaben des metrischen Äquivalentes bei bestimmten Maßeinheiten."
}

// Match adapts abstract Match to []*RuleMatch (no error).
func (r *UnitConversionRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || r.AbstractUnitConversionRule == nil {
		return nil
	}
	ms, _ := r.AbstractUnitConversionRule.Match(sentence)
	return ms
}

// parseDEUnitNumber: German thousands '.' and decimal ','.
func parseDEUnitNumber(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if strings.Contains(s, ",") && strings.Contains(s, ".") {
		s = strings.ReplaceAll(s, ".", "")
		s = strings.ReplaceAll(s, ",", ".")
	} else if strings.Contains(s, ",") {
		s = strings.ReplaceAll(s, ",", ".")
	}
	// pure thousand separators like 1.800 (no comma) → 1800 if pattern is groups of 3
	if strings.Count(s, ".") == 1 {
		parts := strings.Split(s, ".")
		if len(parts) == 2 && len(parts[1]) == 3 && !strings.Contains(parts[1], "e") {
			// ambiguous: 1.82 could be 1.82 or 1820 — Java NumberFormat GERMANY:
			// 1.800 = 1800, 1,82 = 1.82. Three digits after dot → thousands.
			s = parts[0] + parts[1]
		}
	}
	return strconv.ParseFloat(s, 64)
}

// formatDEUnitNumber: Locale.GERMANY, max 2 fraction digits, HALF_UP.
func formatDEUnitNumber(v float64) string {
	// HALF_UP to 2 decimals
	sign := ""
	if v < 0 {
		sign = "-"
		v = -v
	}
	scaled := math.Round(v*100) / 100
	if math.Abs(scaled-math.Round(scaled)) < 1e-9 {
		return sign + strconv.FormatInt(int64(math.Round(scaled)), 10)
	}
	s := fmt.Sprintf("%.2f", scaled)
	s = strings.ReplaceAll(s, ".", ",")
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ",")
	return sign + s
}
