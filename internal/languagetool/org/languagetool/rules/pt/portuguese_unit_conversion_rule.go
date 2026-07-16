package pt

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// PortugueseUnitConversionRule ports org.languagetool.rules.pt.PortugueseUnitConversionRule.
type PortugueseUnitConversionRule struct {
	*rules.AbstractUnitConversionRule
}

func NewPortugueseUnitConversionRule(messages map[string]string) *PortugueseUnitConversionRule {
	base := rules.NewAbstractUnitConversionRule(messages)
	base.ID = "UNITS_PT"
	// Portuguese unit names (subset of Java registrations).
	base.AddUnit(`(?:qui|ki)lo(?:grama)?s?`, rules.UnitKilogram, "quilogramas", 1, true)
	base.AddUnit(`gramas?`, rules.UnitGram, "gramas", 1, true)
	base.AddUnit(`toneladas?`, rules.UnitKilogram, "toneladas", 1e3, true)
	base.AddUnit(`libras?`, rules.UnitPound, "libras", 1, false)
	base.AddUnit(`milhas?`, rules.UnitMile, "milhas", 1, false)
	base.AddUnit(`jardas?`, rules.UnitYard, "jardas", 1, false)
	base.AddUnit(`pés?`, rules.UnitFeet, "pés", 1, false)
	base.AddUnit(`polegadas?`, rules.UnitInch, "polegadas", 1, false)
	base.AddUnit(`metros?`, rules.UnitMetre, "metros", 1, true)
	base.AddUnit(`(?:qu|k)il[oô]metros?`, rules.UnitKilometre, "quilômetros", 1, true)
	base.AddUnit(`cent[ií]metros?`, rules.UnitCentimetre, "centímetros", 1, true)
	base.AddUnit(`mil[ií]metros?`, rules.UnitMillimetre, "milímetros", 1, true)
	base.AddUnit(`(?:qu|k)il[oô]metros? por hora`, rules.UnitKmh, "quilômetros por hora", 1, true)
	base.AddUnit(`milhas? por hora`, rules.UnitMph, "milhas por hora", 1, false)
	return &PortugueseUnitConversionRule{AbstractUnitConversionRule: base}
}

func (r *PortugueseUnitConversionRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	ms, err := r.AbstractUnitConversionRule.Match(sentence)
	if err != nil {
		return nil
	}
	return ms
}
