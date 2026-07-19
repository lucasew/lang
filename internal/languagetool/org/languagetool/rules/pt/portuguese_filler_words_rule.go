package pt

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// PortugueseFillerWordsRule ports org.languagetool.rules.pt.PortugueseFillerWordsRule
// Java default minPercent=8 (AbstractFillerWordsRule).
type PortugueseFillerWordsRule struct {
	*rules.AbstractFillerWordsRule
}

func NewPortugueseFillerWordsRule(messages map[string]string) *PortugueseFillerWordsRule {
	fillers := map[string]struct{}{}
	for _, w := range []string{
		"abundante", "acrescentou", "acrescidamente", "adição", "agora", "ainda", "além", "algo",
		"algum", "alguma", "algumas", "alguns", "aparecer", "aparentemente", "apenas", "apesar",
		"aproximadamente", "assim", "atrás", "atualmente", "automaticamente", "bem", "bonito",
		"certamente", "certo", "claramente", "claro", "completam", "completamente", "completo",
		"comumente", "consequentemente", "consistentemente", "continuamente", "contra", "contraste",
		"contudo", "cuidado", "curto", "dependendo", "depois", "desigual", "determinado", "deve",
		"dever", "difícil", "direito", "dúvida", "embora", "enquanto", "entanto", "ergo", "especial",
		"estranhamente", "eventualmente", "evidentemente", "expressar", "extremamente", "fácil",
		"famoso", "feio", "felizmente", "francamente", "frequência", "frequentemente", "geralmente",
		"graças", "impressionante", "impronunciável", "incomum", "indizível", "infelizmente",
		"irrelevante", "irrelevantes", "já", "justo", "lento", "longo", "lugares", "maior", "mais",
		"mas", "melhor", "mesmo", "muita", "muitas", "muito", "muitos", "múltipla", "nada", "não",
		"natural", "naturalmente", "natureza", "nehumas", "nenhum", "nenhuma", "nenhuns",
		"nomeadamente", "normalmente", "novo", "número", "nunca", "óbvio", "ocasionalmente", "outra",
		"outros", "para", "parente", "particularmente", "pessoa", "pode", "poderia", "pois", "porém",
		"porque", "portanto", "possível", "possivelmente", "pouca", "poucas", "pouco", "poucos",
		"prático", "precisas", "principalmente", "provável", "provavelmente", "quaisquer", "qualquer",
		"quase", "rápido", "raramente", "razoavelmente", "realmente", "recentemente", "relativamente",
		"repente", "sempre", "senão", "sentida", "sentidas", "sentido", "sentidos", "siga",
		"significativo", "sim", "simples", "simplesmente", "sobre", "sozinho", "suave", "suavemente",
		"substancialmente", "suficientemente", "tipo", "tornar", "tornaram", "tornou", "total",
		"totalmente", "toda", "todas", "todo", "todos", "tudo", "ultrajante", "velho", "verdade",
		"vez", "vezes", "volta",
	} {
		fillers[w] = struct{}{}
	}
	base := &rules.AbstractFillerWordsRule{
		AbstractStatisticStyleRule: &rules.AbstractStatisticStyleRule{},
		Messages:                   messages,
		ID:                         "FILLER_WORDS_PT",
		Description:                "Filler words",
		ShortMsg:                   "Filler word",
		Message:                    "Esta palavra pode ser um enchimento estilístico.",
		FillerWords:                fillers,
		IsException: func(tokens []*languagetool.AnalyzedTokenReadings, num int) bool {
			if num >= 1 && tokens[num].GetToken() == "mas" && tokens[num-1].GetToken() == "," {
				return true
			}
			return false
		},
	}
	rules.InitFillerWordsMeta(base, messages, false)
	return &PortugueseFillerWordsRule{AbstractFillerWordsRule: base}
}

func (r *PortugueseFillerWordsRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractFillerWordsRule.Match(sentence)
}
