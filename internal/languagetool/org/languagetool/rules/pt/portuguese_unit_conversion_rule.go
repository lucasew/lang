package pt

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// PortugueseUnitConversionRule ports org.languagetool.rules.pt.PortugueseUnitConversionRule.
type PortugueseUnitConversionRule struct {
	*rules.AbstractUnitConversionRule
}

// Java: OUNCE = POUND.divide(12) (AbstractUnitConversionRule).
const ptOunceOnPound = 1.0 / 12.0

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
			panic(fmt.Sprintf("Unknown message type.%v", m))
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
			panic(fmt.Sprintf("Unknown message type.%v", m))
		}
	}

	// Java PortugueseUnitConversionRule constructor addUnit registrations (order preserved).
	// Mass
	base.AddUnit(`(qui|ki)lo(grama)?`, rules.UnitKilogram, "quilogramas", 1, true)
	base.AddUnit(`grama`, rules.UnitKilogram, "gramas", 1e-3, true)
	base.AddUnit(`toneladas?`, rules.UnitKilogram, "toneladas", 1e3, true)
	base.AddUnit(`libras?`, rules.UnitPound, "libras", 1, false)
	// Java: addUnit("onças?", OUNCE, "onças", 1, false) with OUNCE = POUND.divide(12)
	base.AddUnit(`onças?`, rules.UnitPound, "onças", ptOunceOnPound, false)

	// Length (imperial) — Java order
	base.AddUnit(`milhas?`, rules.UnitMile, "milhas", 1, false)
	base.AddUnit(`jardas?`, rules.UnitYard, "jardas", 1, false)
	base.AddUnit(`pés?`, rules.UnitFeet, "pés", 1, false)
	base.AddUnit(`polegadas?`, rules.UnitInch, "polegadas", 1, false)

	// Speed — Java: km/h before milhas por hora
	base.AddUnit(`(qu|k)ilômetros? por hora`, rules.UnitKmh, "quilômetros por hora", 1, true)
	base.AddUnit(`milhas? por hora`, rules.UnitMph, "milhas por hora", 1, false)

	// Metric length (Java uses accented unit spellings as literals)
	base.AddUnit(`metros?`, rules.UnitMetre, "metros", 1, true)
	base.AddUnit(`(qu|k)ilômetros?`, rules.UnitMetre, "quilômetros", 1e3, true)
	base.AddUnit(`decímetros?`, rules.UnitMetre, "decímetros", 1e-1, false) // metric, but should not be suggested
	base.AddUnit(`centímetros?`, rules.UnitMetre, "centímetros", 1e-2, true)
	base.AddUnit(`milímetros?`, rules.UnitMetre, "milímetros", 1e-3, true)
	base.AddUnit(`micrômetros?`, rules.UnitMetre, "micrômetros", 1e-6, true)
	base.AddUnit(`nanômetros?`, rules.UnitMetre, "nanômetros", 1e-9, true)
	base.AddUnit(`picômetros?`, rules.UnitMetre, "picômetros", 1e-12, true)
	base.AddUnit(`fentômetros?`, rules.UnitMetre, "fentômetros", 1e-15, true)

	// Area
	base.AddUnit(`metros? quadrados?`, rules.UnitSquareMetre, "metros quadrados", 1, true)
	base.AddUnit(`hectar(es)?`, rules.UnitSquareMetre, "hectares", 1e4, true)
	base.AddUnit(`ares?`, rules.UnitSquareMetre, "ares", 1e2, true)
	base.AddUnit(`(k|qui)ilômetros? quadrados?`, rules.UnitSquareMetre, "quilômetros quadrados", 1e6, true)
	base.AddUnit(`decímetros? quadrados?`, rules.UnitSquareMetre, "decímetros quadrados", 1e-2, false) // Metric, but not commonly used
	base.AddUnit(`centímetros? quadrados?`, rules.UnitSquareMetre, "centímetros quadrados", 1e-4, true)
	base.AddUnit(`milímetros? quadrados?`, rules.UnitSquareMetre, "milímetros quadrados", 1e-6, true)
	base.AddUnit(`micrômetros? quadrados?`, rules.UnitSquareMetre, "micrômetros quadrados", 1e-12, true)
	base.AddUnit(`nanômetros? quadrados?`, rules.UnitSquareMetre, "nanômetros quadrados", 1e-18, true)
	// Java abstract defaults already register "(?:sq|square) (?:ft|feet|foot)" for sq ft tests.

	// Volume (cubic metres)
	base.AddUnit(`metros? cúbicos?`, rules.UnitCubicMetre, "metros cúbicos", 1, true)
	base.AddUnit(`(k|qu)ilômetros? cúbicos?`, rules.UnitCubicMetre, "quilômetros cúbicos", 1e9, true)
	base.AddUnit(`decímetros? cúbicos?`, rules.UnitCubicMetre, "decímetros cúbicos", 1e-3, false) // Metric, but not commonly used
	base.AddUnit(`centímetros? cúbicos?`, rules.UnitCubicMetre, "centímetros cúbicos", 1e-6, true)
	base.AddUnit(`milímetros? cúbicos?`, rules.UnitCubicMetre, "milímetros cúbicos", 1e-9, true)
	base.AddUnit(`micrômetros? cúbicos?`, rules.UnitCubicMetre, "micrômetros cúbicos", 1e-18, true)
	base.AddUnit(`nanômetros? cúbicos?`, rules.UnitCubicMetre, "nanômetros cúbicos", 1e-27, true)

	base.AddUnit(`litros?`, rules.UnitLitre, "litros", 1, true)
	base.AddUnit(`mililitros?`, rules.UnitLitre, "mililitros", 1e-3, true)

	// Temperature — Java: "(?:Graus)? Fahrenheit" / "(?:Graus)? (Celsi[ou]s|[cC]entígrados?)"
	// (literal space after optional Graus)
	base.AddUnit(`(?:Graus)? Fahrenheit`, rules.UnitFahrenheit, "graus Fahrenheit", 1, false)
	base.AddUnit(`(?:Graus)? (Celsi[ou]s|[cC]entígrados?)`, rules.UnitCelsius, "graus Celsius", 1, true)

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
	s = tools.JavaStringTrim(s)
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
// Grouping separators omitted — Java enables them, but the PT test matrix only
// asserts decimal forms like "1,83 metros" / "4,54 toneladas" / "160,93 quilômetros".
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
