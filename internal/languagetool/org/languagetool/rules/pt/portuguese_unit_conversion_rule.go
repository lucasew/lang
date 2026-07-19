package pt

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// PortugueseUnitConversionRule ports org.languagetool.rules.pt.PortugueseUnitConversionRule.
type PortugueseUnitConversionRule struct {
	*rules.AbstractUnitConversionRule
}

func NewPortugueseUnitConversionRule(messages map[string]string) *PortugueseUnitConversionRule {
	base := rules.NewAbstractUnitConversionRule(messages)
	// Java getId
	base.ID = "UNIDADES_METRICAS"
	// Java: NumberFormat.getNumberInstance(Locale.GERMANY), max 2 fraction digits, HALF_UP
	base.FormatNumber = formatPTUnitNumber
	base.ParseNumber = parsePTUnitNumber
	// Java PortugueseUnitConversionRule.formatRounded
	base.FormatRounded = func(s string) string { return "aprox. " + s }
	base.MessageFor = func(m rules.UnitConversionMessage) string {
		switch m {
		case rules.UnitMsgCheck:
			return "Esta conversão não parece estar precisa. Gostaria de corrigi-la?"
		case rules.UnitMsgSuggestion:
			return "Deseja adicionar automaticamente uma conversão ao sistema métrico?"
		case rules.UnitMsgCheckUnknownUnit:
			return "A unidade usada nesta conversão não foi reconhecida."
		case rules.UnitMsgUnitMismatch:
			return "Estas unidades de medida não são compatíveis."
		default:
			return "Unidade"
		}
	}
	// Java PortugueseUnitConversionRule.getShortMessage
	base.ShortMessageFor = func(m rules.UnitConversionMessage) string {
		switch m {
		case rules.UnitMsgCheck:
			return "Conversão incorreta. Corrigir?"
		case rules.UnitMsgSuggestion:
			return "Adicionar conversão ao sistema métrico?"
		case rules.UnitMsgCheckUnknownUnit:
			return "Unidade desconhecida."
		case rules.UnitMsgUnitMismatch:
			return "Unidade incompatível."
		default:
			return "Unidade"
		}
	}

	// Java PortugueseUnitConversionRule unit registrations (order: longer phrases before shorter).
	// Mass
	base.AddUnit(`(?:qui|ki)lo(?:grama)?s?`, rules.UnitKilogram, "quilogramas", 1, true)
	base.AddUnit(`gramas?`, rules.UnitKilogram, "gramas", 1e-3, true)
	base.AddUnit(`toneladas?`, rules.UnitKilogram, "toneladas", 1e3, true)
	base.AddUnit(`libras?`, rules.UnitPound, "libras", 1, false)
	// Java OUNCE = POUND.divide(12)
	base.AddUnit(`onças?`, rules.UnitPound, "onças", 1.0/12.0, false)

	// Length (imperial)
	base.AddUnit(`milhas? por hora`, rules.UnitMph, "milhas por hora", 1, false)
	base.AddUnit(`milhas?`, rules.UnitMile, "milhas", 1, false)
	base.AddUnit(`jardas?`, rules.UnitYard, "jardas", 1, false)
	base.AddUnit(`pés?`, rules.UnitFeet, "pés", 1, false)
	base.AddUnit(`polegadas?`, rules.UnitInch, "polegadas", 1, false)

	// Speed (metric names before plain km)
	base.AddUnit(`(?:qu|k)il[oô]metros? por hora`, rules.UnitKmh, "quilômetros por hora", 1, true)

	// Metric length
	base.AddUnit(`metros?`, rules.UnitMetre, "metros", 1, true)
	base.AddUnit(`(?:qu|k)il[oô]metros?`, rules.UnitMetre, "quilômetros", 1e3, true)
	base.AddUnit(`dec[ií]metros?`, rules.UnitMetre, "decímetros", 1e-1, false) // metric, but should not be suggested
	base.AddUnit(`cent[ií]metros?`, rules.UnitMetre, "centímetros", 1e-2, true)
	base.AddUnit(`mil[ií]metros?`, rules.UnitMetre, "milímetros", 1e-3, true)
	base.AddUnit(`micr[oô]metros?`, rules.UnitMetre, "micrômetros", 1e-6, true)
	base.AddUnit(`nan[oô]metros?`, rules.UnitMetre, "nanômetros", 1e-9, true)
	base.AddUnit(`pic[oô]metros?`, rules.UnitMetre, "picômetros", 1e-12, true)
	base.AddUnit(`fent[oô]metros?`, rules.UnitMetre, "fentômetros", 1e-15, true)

	// Area
	base.AddUnit(`metros? quadrados?`, rules.UnitSquareMetre, "metros quadrados", 1, true)
	base.AddUnit(`hectar(?:es)?`, rules.UnitSquareMetre, "hectares", 1e4, true)
	base.AddUnit(`ares?`, rules.UnitSquareMetre, "ares", 1e2, true)
	base.AddUnit(`(?:k|qui)il[oô]metros? quadrados?`, rules.UnitSquareMetre, "quilômetros quadrados", 1e6, true)
	base.AddUnit(`dec[ií]metros? quadrados?`, rules.UnitSquareMetre, "decímetros quadrados", 1e-2, false) // Metric, but not commonly used
	base.AddUnit(`cent[ií]metros? quadrados?`, rules.UnitSquareMetre, "centímetros quadrados", 1e-4, true)
	base.AddUnit(`mil[ií]metros? quadrados?`, rules.UnitSquareMetre, "milímetros quadrados", 1e-6, true)
	base.AddUnit(`micr[oô]metros? quadrados?`, rules.UnitSquareMetre, "micrômetros quadrados", 1e-12, true)
	base.AddUnit(`nan[oô]metros? quadrados?`, rules.UnitSquareMetre, "nanômetros quadrados", 1e-18, true)
	// Java tests also exercise abstract default "sq ft"
	base.AddUnit(`(?:sq|square)\s+(?:ft|feet|foot)`, rules.UnitSqFt, "sq ft", 1, false)

	// Volume (cubic metres)
	base.AddUnit(`metros? c[uú]bicos?`, rules.UnitCubicMetre, "metros cúbicos", 1, true)
	base.AddUnit(`(?:k|qu)il[oô]metros? c[uú]bicos?`, rules.UnitCubicMetre, "quilômetros cúbicos", 1e9, true)
	base.AddUnit(`dec[ií]metros? c[uú]bicos?`, rules.UnitCubicMetre, "decímetros cúbicos", 1e-3, false) // Metric, but not commonly used
	base.AddUnit(`cent[ií]metros? c[uú]bicos?`, rules.UnitCubicMetre, "centímetros cúbicos", 1e-6, true)
	base.AddUnit(`mil[ií]metros? c[uú]bicos?`, rules.UnitCubicMetre, "milímetros cúbicos", 1e-9, true)
	base.AddUnit(`micr[oô]metros? c[uú]bicos?`, rules.UnitCubicMetre, "micrômetros cúbicos", 1e-18, true)
	base.AddUnit(`nan[oô]metros? c[uú]bicos?`, rules.UnitCubicMetre, "nanômetros cúbicos", 1e-27, true)

	base.AddUnit(`litros?`, rules.UnitLitre, "litros", 1, true)
	base.AddUnit(`mililitros?`, rules.UnitLitre, "mililitros", 1e-3, true)

	// Temperature
	base.AddUnit(`(?:Graus?\s*)?Fahrenheit`, rules.UnitFahrenheit, "graus Fahrenheit", 1, false)
	base.AddUnit(`(?:Graus?\s*)?(?:Celsi[ou]s|[cC]ent[ií]grados?)`, rules.UnitCelsius, "graus Celsius", 1, true)

	return &PortugueseUnitConversionRule{AbstractUnitConversionRule: base}
}

func (r *PortugueseUnitConversionRule) GetID() string {
	if r != nil && r.AbstractUnitConversionRule != nil && r.ID != "" {
		return r.ID
	}
	return "UNIDADES_METRICAS"
}

func (r *PortugueseUnitConversionRule) GetDescription() string {
	return "Sugere ou verifica informações equivalentes à métrica de unidades de medida específicas."
}

// Match adapts abstract Match to []*RuleMatch (no error).
func (r *PortugueseUnitConversionRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || r.AbstractUnitConversionRule == nil {
		return nil
	}
	ms, _ := r.AbstractUnitConversionRule.Match(sentence)
	return ms
}

// parsePTUnitNumber: Java Locale.GERMANY thousands '.' and decimal ','.
func parsePTUnitNumber(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if strings.Contains(s, ",") && strings.Contains(s, ".") {
		s = strings.ReplaceAll(s, ".", "")
		s = strings.ReplaceAll(s, ",", ".")
	} else if strings.Contains(s, ",") {
		s = strings.ReplaceAll(s, ",", ".")
	}
	// pure thousand separators like 10.000 (no comma) → 10000 if 3 digits after dot
	if strings.Count(s, ".") == 1 {
		parts := strings.Split(s, ".")
		if len(parts) == 2 && len(parts[1]) == 3 && !strings.Contains(parts[1], "e") {
			s = parts[0] + parts[1]
		}
	}
	return strconv.ParseFloat(s, 64)
}

// formatPTUnitNumber: Locale.GERMANY, max 2 fraction digits, HALF_UP.
func formatPTUnitNumber(v float64) string {
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
