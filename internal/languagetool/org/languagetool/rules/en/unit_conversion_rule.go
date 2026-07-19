package en

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// UnitConversionRule ports org.languagetool.rules.en.UnitConversionRule
// (extends AbstractUnitConversionRule — not a surface stand-in Match).
type UnitConversionRule struct {
	*rules.AbstractUnitConversionRule
}

func NewUnitConversionRule(messages map[string]string) *UnitConversionRule {
	base := rules.NewAbstractUnitConversionRule(messages)
	base.ID = "METRIC_UNITS_EN_GENERAL"
	registerENGeneralUnits(base)
	return &UnitConversionRule{AbstractUnitConversionRule: base}
}

func NewUnitConversionRuleImperial(messages map[string]string) *UnitConversionRule {
	base := rules.NewAbstractUnitConversionRule(messages)
	base.ID = "METRIC_UNITS_EN_IMPERIAL"
	registerENGeneralUnits(base)
	registerENImperialUnits(base)
	return &UnitConversionRule{AbstractUnitConversionRule: base}
}

func NewUnitConversionRuleUS(messages map[string]string) *UnitConversionRule {
	base := rules.NewAbstractUnitConversionRule(messages)
	base.ID = "METRIC_UNITS_EN_US"
	registerENGeneralUnits(base)
	registerENUSUnits(base)
	return &UnitConversionRule{AbstractUnitConversionRule: base}
}

// registerENGeneralUnits ports UnitConversionRule constructor addUnit calls.
func registerENGeneralUnits(base *rules.AbstractUnitConversionRule) {
	if base == nil {
		return
	}
	// Java: setTags(Tag.picky) handled by rule registry; units:
	base.AddUnit(`miles per hour`, rules.UnitMph, "miles per hour", 1, false)

	base.AddUnit(`kilograms?`, rules.UnitKilogram, "kilogram", 1, true)
	base.AddUnit(`grams?`, rules.UnitGram, "gram", 1, true)
	base.AddUnit(`tons?`, rules.UnitTonne, "ton", 1, true)

	base.AddUnit(`pounds?`, rules.UnitPound, "pounds", 1, false)
	// Java ounces? → OUNCE mass (1/16 lb)
	base.AddUnit(`ounces?`, rules.UnitPound, "ounces", 1.0/16.0, false)

	base.AddUnit(`feet`, rules.UnitFeet, "feet", 1, false)
	base.AddUnit(`miles?`, rules.UnitMile, "miles", 1, false)
	base.AddUnit(`yards?`, rules.UnitYard, "yards", 1, false)
	base.AddUnit(`inch(?:es)?`, rules.UnitInch, "inches", 1, false)

	// Fahrenheit / Celsius word forms (abstract already has °F / °C)
	base.AddUnit(`(?:degrees?)?\s*Fahrenheit`, rules.UnitFahrenheit, "degree Fahrenheit", 1, false)
	base.AddUnit(`(?:degrees?)?\s*Celsius`, rules.UnitCelsius, "degree Celsius", 1, true)

	// Java AbstractUnitConversionRule defaults also include sq ft (area)
	base.AddUnit(`(?:sq|square)\s+(?:ft|feet|foot)`, rules.UnitSqFt, "sq ft", 1, false)
	base.AddUnit(`sf`, rules.UnitSqFt, "sf", 1, false)
	base.AddUnit(`ft(?:\^2|2|²)`, rules.UnitSqFt, "ft²", 1, false)
	base.AddUnit(`m(?:\^2|2|²)`, rules.UnitSquareMetre, "m²", 1, true)
}

// registerENImperialUnits ports UnitConversionRuleImperial extra units (UK spelling + imperial volume).
func registerENImperialUnits(base *rules.AbstractUnitConversionRule) {
	if base == nil {
		return
	}
	base.AddUnit(`(?:kilometre|kilometer)s?\s+per\s+hour`, rules.UnitKmh, "kilometres per hour", 1, true)

	base.AddUnit(`kilomet(?:re|er)s?`, rules.UnitKilometre, "kilometres", 1, true)
	base.AddUnit(`met(?:re|er)s?`, rules.UnitMetre, "metres", 1, true)
	base.AddUnit(`centimet(?:re|er)s?`, rules.UnitCentimetre, "centimetres", 1, true)
	base.AddUnit(`millimet(?:re|er)s?`, rules.UnitMillimetre, "millimetres", 1, true)

	base.AddUnit(`square\s+met(?:re|er)s?`, rules.UnitSquareMetre, "square metres", 1, true)
	base.AddUnit(`lit(?:re|er)s?`, rules.UnitLitre, "litres", 1, true)
	base.AddUnit(`millilit(?:re|er)s?`, rules.UnitMillilitre, "millilitres", 1, true)

	// Imperial pint ≈ 0.56826125 L (UnitLitre base factor 1e-3 m³)
	base.AddUnit(`pints?`, rules.UnitLitre, "pints", 0.56826125, false)
	base.AddUnit(`pt`, rules.UnitLitre, "pt", 0.56826125, false)
	// Imperial fluid ounce ≈ 0.0284130625 L
	base.AddUnit(`(?:fluid\s+)?ounces?`, rules.UnitLitre, "fluid ounces", 0.0284130625, false)
	base.AddUnit(`gallons?`, rules.UnitLitre, "gallons", 4.54609, false)
	base.AddUnit(`quarts?`, rules.UnitLitre, "quarts", 1.1365225, false)
}

// registerENUSUnits ports UnitConversionRuleUS extra units (US spelling + US volume).
func registerENUSUnits(base *rules.AbstractUnitConversionRule) {
	if base == nil {
		return
	}
	base.AddUnit(`(?:kilometre|kilometer)s?\s+per\s+hour`, rules.UnitKmh, "kilometers per hour", 1, true)

	base.AddUnit(`kilomet(?:re|er)s?`, rules.UnitKilometre, "kilometers", 1, true)
	base.AddUnit(`met(?:re|er)s?`, rules.UnitMetre, "meters", 1, true)
	base.AddUnit(`centimet(?:re|er)s?`, rules.UnitCentimetre, "centimeters", 1, true)
	base.AddUnit(`millimet(?:re|er)s?`, rules.UnitMillimetre, "millimeters", 1, true)

	base.AddUnit(`square\s+met(?:re|er)s?`, rules.UnitSquareMetre, "square meters", 1, true)
	base.AddUnit(`lit(?:re|er)s?`, rules.UnitLitre, "liters", 1, true)
	base.AddUnit(`millilit(?:re|er)s?`, rules.UnitMillilitre, "milliliters", 1, true)

	// US liquid pint ≈ 0.473176473 L
	base.AddUnit(`pints?`, rules.UnitLitre, "pints", 0.473176473, false)
	base.AddUnit(`pt`, rules.UnitLitre, "pt", 0.473176473, false)
	// US fluid ounce ≈ 0.0295735295625 L
	base.AddUnit(`(?:fluid\s+)?ounces?`, rules.UnitLitre, "fluid ounces", 0.0295735295625, false)
	base.AddUnit(`gallons?`, rules.UnitLitre, "gallons", 3.785411784, false)
	base.AddUnit(`quarts?`, rules.UnitLitre, "quarts", 0.946352946, false)
	base.AddUnit(`cups?`, rules.UnitLitre, "cups", 0.2365882365, false)
}

func (r *UnitConversionRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || r.AbstractUnitConversionRule == nil {
		return nil
	}
	ms, err := r.AbstractUnitConversionRule.Match(sentence)
	if err != nil {
		return nil
	}
	return ms
}

// GetID ports Rule.getId (variant IDs set on AbstractUnitConversionRule.ID).
func (r *UnitConversionRule) GetID() string {
	if r == nil || r.AbstractUnitConversionRule == nil {
		return "METRIC_UNITS_EN_GENERAL"
	}
	return r.AbstractUnitConversionRule.GetID()
}
