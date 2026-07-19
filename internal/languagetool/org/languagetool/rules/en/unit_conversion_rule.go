package en

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// UnitConversionRule ports org.languagetool.rules.en.UnitConversionRule
// (extends AbstractUnitConversionRule — not a surface stand-in Match).
// Java UnitConversionRule.setTags(Tag.picky) applies to general + Imperial + US.
type UnitConversionRule struct {
	*rules.AbstractUnitConversionRule
	// Tags ports Rule.tags (Java Tag.picky).
	Tags []rules.Tag
}

// Imperial / US volume factors from Java AbstractUnitConversionRule (litre multiples).
const (
	// Java: OUNCE = POUND.divide(12)
	enOunceOnPound = 1.0 / 12.0
	// Imperial volume (Java IMP_*)
	impPintL     = 0.5682612532
	impQuartL    = impPintL * 2
	impGallonL   = impQuartL * 4
	impFlOunceL  = impPintL / 20
	// US volume (Java US_*)
	usQuartL    = 0.946352946
	usGallonL   = usQuartL * 4
	usPintL     = usQuartL / 2
	usCupL      = usQuartL / 4
	usFlOunceL  = usQuartL / 32
)

func newENUnitConversion(messages map[string]string, id string) *UnitConversionRule {
	base := rules.NewAbstractUnitConversionRule(messages)
	base.ID = id
	return &UnitConversionRule{
		AbstractUnitConversionRule: base,
		// Java: setTags(Arrays.asList(Tag.picky))
		Tags: []rules.Tag{rules.TagPicky},
	}
}

func NewUnitConversionRule(messages map[string]string) *UnitConversionRule {
	r := newENUnitConversion(messages, "METRIC_UNITS_EN_GENERAL")
	registerENGeneralUnits(r.AbstractUnitConversionRule)
	return r
}

func NewUnitConversionRuleImperial(messages map[string]string) *UnitConversionRule {
	r := newENUnitConversion(messages, "METRIC_UNITS_EN_IMPERIAL")
	registerENGeneralUnits(r.AbstractUnitConversionRule)
	registerENImperialUnits(r.AbstractUnitConversionRule)
	return r
}

func NewUnitConversionRuleUS(messages map[string]string) *UnitConversionRule {
	r := newENUnitConversion(messages, "METRIC_UNITS_EN_US")
	registerENGeneralUnits(r.AbstractUnitConversionRule)
	registerENUSUnits(r.AbstractUnitConversionRule)
	return r
}

// registerENGeneralUnits ports UnitConversionRule constructor addUnit calls only
// (Java UnitConversionRule; abstract defaults already supply kg/mi/°F/…).
func registerENGeneralUnits(base *rules.AbstractUnitConversionRule) {
	if base == nil {
		return
	}
	// Java: setTags(Tag.picky) handled by rule registry
	base.AddUnit(`miles per hour`, rules.UnitMph, "miles per hour", 1, false)

	base.AddUnit(`kilograms?`, rules.UnitKilogram, "kilogram", 1, true)
	base.AddUnit(`grams?`, rules.UnitKilogram, "gram", 1e-3, true)
	base.AddUnit(`tons?`, rules.UnitKilogram, "ton", 1e3, true)

	base.AddUnit(`pounds?`, rules.UnitPound, "pounds", 1, false)
	// Java OUNCE = POUND.divide(12)
	base.AddUnit(`ounces?`, rules.UnitPound, "ounces", enOunceOnPound, false)

	base.AddUnit(`feet`, rules.UnitFeet, "feet", 1, false)
	base.AddUnit(`miles?`, rules.UnitMile, "miles", 1, false)
	base.AddUnit(`yards?`, rules.UnitYard, "yards", 1, false)
	base.AddUnit(`inch(?:es)?`, rules.UnitInch, "inches", 1, false)

	base.AddUnit(`(?:degrees?)?\s*Fahrenheit`, rules.UnitFahrenheit, "degree Fahrenheit", 1, false)
	base.AddUnit(`(?:degrees?)?\s*Celsius`, rules.UnitCelsius, "degree Celsius", 1, true)
}

// registerENImperialUnits ports UnitConversionRuleImperial addUnit list.
func registerENImperialUnits(base *rules.AbstractUnitConversionRule) {
	if base == nil {
		return
	}
	base.AddUnit(`(?:kilometre|kilometer)s?\s+per\s+hour`, rules.UnitKmh, "kilometres per hour", 1, true)

	base.AddUnit(`kilomet(?:re|er)s?`, rules.UnitMetre, "kilometres", 1e3, true)
	base.AddUnit(`met(?:re|er)s?`, rules.UnitMetre, "metres", 1, true)
	base.AddUnit(`decimet(?:re|er)s?`, rules.UnitMetre, "decimetres", 1e-1, false)
	base.AddUnit(`centimet(?:re|er)s?`, rules.UnitMetre, "centimetres", 1e-2, true)
	// Java symbol for mm is "micrometres" (upstream typo); keep millimetres spelling for factor 1e-3
	base.AddUnit(`millimet(?:re|er)s?`, rules.UnitMetre, "millimetres", 1e-3, true)
	base.AddUnit(`micromet(?:re|er)s?`, rules.UnitMetre, "micrometres", 1e-6, true)
	base.AddUnit(`nanomet(?:re|er)s?`, rules.UnitMetre, "nanometres", 1e-9, true)

	base.AddUnit(`square\s+met(?:re|er)s?`, rules.UnitSquareMetre, "square metres", 1, true)
	base.AddUnit(`square\s+kilomet(?:re|er)s?`, rules.UnitSquareMetre, "square kilometres", 1e6, true)
	base.AddUnit(`square\s+decimet(?:re|er)s?`, rules.UnitSquareMetre, "square decimetres", 1e-2, false)
	base.AddUnit(`square\s+centimet(?:re|er)s?`, rules.UnitSquareMetre, "square centimetres", 1e-4, true)
	base.AddUnit(`square\s+millimet(?:re|er)s?`, rules.UnitSquareMetre, "square millimetres", 1e-6, true)
	base.AddUnit(`square\s+micromet(?:re|er)s?`, rules.UnitSquareMetre, "square micrometres", 1e-12, true)
	base.AddUnit(`square\s+nanomet(?:re|er)s?`, rules.UnitSquareMetre, "square nanometres", 1e-18, true)

	base.AddUnit(`cubic\s+met(?:re|er)s?`, rules.UnitCubicMetre, "cubic metres", 1, true)
	base.AddUnit(`cubic\s+kilomet(?:re|er)s?`, rules.UnitCubicMetre, "cubic kilometres", 1e9, true)
	base.AddUnit(`cubic\s+decimet(?:re|er)s?`, rules.UnitCubicMetre, "cubic decimetres", 1e-3, false)
	base.AddUnit(`cubic\s+centimet(?:re|er)s?`, rules.UnitCubicMetre, "cubic centimetres", 1e-6, true)
	base.AddUnit(`cubic\s+millimet(?:re|er)s?`, rules.UnitCubicMetre, "cubic millimetres", 1e-9, true)
	base.AddUnit(`cubic\s+micromet(?:re|er)s?`, rules.UnitCubicMetre, "cubic micrometres", 1e-18, true)
	base.AddUnit(`cubic\s+nanomet(?:re|er)s?`, rules.UnitCubicMetre, "cubic nanometres", 1e-27, true)

	base.AddUnit(`lit(?:re|er)s?`, rules.UnitLitre, "litres", 1, true)
	base.AddUnit(`millilit(?:re|er)s?`, rules.UnitLitre, "millilitres", 1e-3, true)

	// Imperial volume abbreviations + long forms (Java IMP_*)
	base.AddUnit(`qt\.`, rules.UnitLitre, "qt.", impQuartL, false)
	base.AddUnit(`gal`, rules.UnitLitre, "gal", impGallonL, false)
	base.AddUnit(`pt`, rules.UnitLitre, "pt", impPintL, false)
	base.AddUnit(`(?:fl\.?\s*oz\.?|oz\.\s*fl\.)`, rules.UnitLitre, "fl oz", impFlOunceL, false)

	base.AddUnit(`quarts?`, rules.UnitLitre, "quarts", impQuartL, false)
	base.AddUnit(`gallons?`, rules.UnitLitre, "gallons", impGallonL, false)
	base.AddUnit(`pints?`, rules.UnitLitre, "pints", impPintL, false)
	base.AddUnit(`(?:fluid\s+)?ounces?`, rules.UnitLitre, "fluid ounces", impFlOunceL, false)
}

// registerENUSUnits ports UnitConversionRuleUS addUnit list.
func registerENUSUnits(base *rules.AbstractUnitConversionRule) {
	if base == nil {
		return
	}
	base.AddUnit(`(?:kilometre|kilometer)s?\s+per\s+hour`, rules.UnitKmh, "kilometers per hour", 1, true)

	base.AddUnit(`kilomet(?:re|er)s?`, rules.UnitMetre, "kilometers", 1e3, true)
	base.AddUnit(`met(?:re|er)s?`, rules.UnitMetre, "meters", 1, true)
	base.AddUnit(`decimet(?:re|er)s?`, rules.UnitMetre, "decimeters", 1e-1, false)
	base.AddUnit(`centimet(?:re|er)s?`, rules.UnitMetre, "centimeters", 1e-2, true)
	base.AddUnit(`millimet(?:re|er)s?`, rules.UnitMetre, "millimeters", 1e-3, true)
	base.AddUnit(`micromet(?:re|er)s?`, rules.UnitMetre, "micrometers", 1e-6, true)
	base.AddUnit(`nanomet(?:re|er)s?`, rules.UnitMetre, "nanometers", 1e-9, true)

	base.AddUnit(`square\s+met(?:re|er)s?`, rules.UnitSquareMetre, "square meters", 1, true)
	base.AddUnit(`square\s+kilomet(?:re|er)s?`, rules.UnitSquareMetre, "square kilometers", 1e6, true)
	base.AddUnit(`square\s+decimet(?:re|er)s?`, rules.UnitSquareMetre, "square decimeters", 1e-2, false)
	base.AddUnit(`square\s+centimet(?:re|er)s?`, rules.UnitSquareMetre, "square centimeters", 1e-4, true)
	base.AddUnit(`square\s+millimet(?:re|er)s?`, rules.UnitSquareMetre, "square millimeters", 1e-6, true)
	base.AddUnit(`square\s+micromet(?:re|er)s?`, rules.UnitSquareMetre, "square micrometers", 1e-12, true)
	base.AddUnit(`square\s+nanomet(?:re|er)s?`, rules.UnitSquareMetre, "square nanometers", 1e-18, true)

	base.AddUnit(`cubic\s+met(?:re|er)s?`, rules.UnitCubicMetre, "cubic meters", 1, true)
	base.AddUnit(`cubic\s+kilomet(?:re|er)s?`, rules.UnitCubicMetre, "cubic kilometers", 1e9, true)
	base.AddUnit(`cubic\s+decimet(?:re|er)s?`, rules.UnitCubicMetre, "cubic decimeters", 1e-3, false)
	base.AddUnit(`cubic\s+centimet(?:re|er)s?`, rules.UnitCubicMetre, "cubic centimeters", 1e-6, true)
	base.AddUnit(`cubic\s+millimet(?:re|er)s?`, rules.UnitCubicMetre, "cubic millimeters", 1e-9, true)
	base.AddUnit(`cubic\s+micromet(?:re|er)s?`, rules.UnitCubicMetre, "cubic micrometers", 1e-18, true)
	base.AddUnit(`cubic\s+nanomet(?:re|er)s?`, rules.UnitCubicMetre, "cubic nanometers", 1e-27, true)

	base.AddUnit(`lit(?:re|er)s?`, rules.UnitLitre, "liters", 1, true)
	base.AddUnit(`millilit(?:re|er)s?`, rules.UnitLitre, "milliliters", 1e-3, true)

	// US volume abbreviations + long forms (Java US_*)
	base.AddUnit(`qt\.`, rules.UnitLitre, "qt.", usQuartL, false)
	base.AddUnit(`gal`, rules.UnitLitre, "gal", usGallonL, false)
	base.AddUnit(`pt`, rules.UnitLitre, "pt", usPintL, false)
	base.AddUnit(`cup`, rules.UnitLitre, "cups", usCupL, false)
	base.AddUnit(`(?:fl\.?\s*oz\.?|oz\.\s*fl\.)`, rules.UnitLitre, "fl oz", usFlOunceL, false)

	base.AddUnit(`quarts?`, rules.UnitLitre, "quarts", usQuartL, false)
	base.AddUnit(`gallons?`, rules.UnitLitre, "gallons", usGallonL, false)
	base.AddUnit(`pints?`, rules.UnitLitre, "pints", usPintL, false)
	base.AddUnit(`cups?`, rules.UnitLitre, "cups", usCupL, false)
	base.AddUnit(`(?:fluid\s+)?ounces?`, rules.UnitLitre, "fluid ounces", usFlOunceL, false)
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

// GetDescription ports UnitConversionRule.getDescription.
func (r *UnitConversionRule) GetDescription() string {
	return "Suggests or checks conversion of units to their metric equivalents."
}

// GetTags ports Rule.getTags (Java Tag.picky).
func (r *UnitConversionRule) GetTags() []rules.Tag {
	if r == nil {
		return nil
	}
	return r.Tags
}

// HasTag ports Rule.hasTag.
func (r *UnitConversionRule) HasTag(tag rules.Tag) bool {
	if r == nil {
		return false
	}
	for _, t := range r.Tags {
		if t == tag {
			return true
		}
	}
	return false
}
