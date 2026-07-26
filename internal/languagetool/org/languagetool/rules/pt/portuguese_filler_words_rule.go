package pt

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// PortugueseFillerWordsRule ports org.languagetool.rules.pt.PortugueseFillerWordsRule
// (extends AbstractFillerWordsRule; Java default minPercent=8, default off).
type PortugueseFillerWordsRule struct {
	*rules.AbstractFillerWordsRule
}

// portugueseFillerWords is the exact static HashSet from Java PortugueseFillerWordsRule
// (172 tokens; case-sensitive lookup via isFillerWord).
var portugueseFillerWords = func() map[string]struct{} {
	words := []string{
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
	}
	m := make(map[string]struct{}, len(words))
	for _, w := range words {
		m[w] = struct{}{}
	}
	return m
}()

// NewPortugueseFillerWordsRule ports PortugueseFillerWordsRule(messages, lang, null).
func NewPortugueseFillerWordsRule(messages map[string]string) *PortugueseFillerWordsRule {
	return NewPortugueseFillerWordsRuleWithUserConfig(messages, nil)
}

// NewPortugueseFillerWordsRuleWithUserConfig ports
// PortugueseFillerWordsRule(ResourceBundle messages, Language lang, UserConfig userConfig).
// Language is unused in Java beyond AbstractStatisticStyleRule storage; Go omits it.
func NewPortugueseFillerWordsRuleWithUserConfig(messages map[string]string, userConfig *languagetool.UserConfig) *PortugueseFillerWordsRule {
	base := &rules.AbstractFillerWordsRule{
		AbstractStatisticStyleRule: &rules.AbstractStatisticStyleRule{},
		Messages:                   messages,
		// Only set ID; Description/Message come from messages keys via InitFillerWordsMeta
		// (filler_words_rule_desc / filler_words_rule_msg) — do not invent PT copy.
		ID:          "FILLER_WORDS_PT",
		FillerWords: portugueseFillerWords,
		IsException: portugueseFillerIsException,
	}
	rules.InitFillerWordsMeta(base, messages, false)

	// Java AbstractStatisticStyleRule: minPercent / excludeDirectSpeech from userConfig.
	if userConfig != nil {
		if cf := userConfig.GetConfigValueByID("FILLER_WORDS_PT"); len(cf) > 0 && cf[0] != nil {
			if n, ok := portugueseFillerConfigInt(cf[0]); ok {
				base.SetMinPercent(n)
			}
		}
		if cf := userConfig.GetConfigValueByID("FILLER_WORDS_PT"); len(cf) > 1 && cf[1] != nil {
			if b, ok := cf[1].(bool); ok {
				base.ExcludeDirectSpeech = b
			}
		}
	}

	r := &PortugueseFillerWordsRule{AbstractFillerWordsRule: base}
	// PT Java isFillerWord is case-sensitive HashSet.contains; abstract Go isFiller
	// lowercases. Override ConditionFulfilled so only exact keys match.
	base.ConditionFulfilled = r.conditionFulfilled
	return r
}

// NewPortugueseFillerWordsRuleWithMinPercent builds with explicit percent (0 = show all).
func NewPortugueseFillerWordsRuleWithMinPercent(messages map[string]string, minPercent int) *PortugueseFillerWordsRule {
	r := NewPortugueseFillerWordsRule(messages)
	r.SetMinPercent(minPercent)
	return r
}

// isFillerWord ports PortugueseFillerWordsRule.isFillerWord — exact, case-sensitive.
func (r *PortugueseFillerWordsRule) isFillerWord(token string) bool {
	if r == nil || r.FillerWords == nil {
		return false
	}
	_, ok := r.FillerWords[token]
	return ok
}

// conditionFulfilled ports AbstractFillerWordsRule.conditionFulfilled with PT isFillerWord.
func (r *PortugueseFillerWordsRule) conditionFulfilled(tokens []*languagetool.AnalyzedTokenReadings, n int) int {
	if r == nil || n < 0 || n >= len(tokens) || tokens[n] == nil {
		return -1
	}
	if r.isFillerWord(tokens[n].GetToken()) && !portugueseFillerIsException(tokens, n) {
		return n
	}
	return -1
}

// portugueseFillerIsException ports PortugueseFillerWordsRule.isException:
// "mas" after "," is not a filler hit.
func portugueseFillerIsException(tokens []*languagetool.AnalyzedTokenReadings, num int) bool {
	if num >= 1 && tokens[num] != nil && tokens[num-1] != nil &&
		tokens[num].GetToken() == "mas" && tokens[num-1].GetToken() == "," {
		return true
	}
	return false
}

func portugueseFillerConfigInt(v any) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int32:
		return int(n), true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	default:
		return 0, false
	}
}

func (r *PortugueseFillerWordsRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || r.AbstractFillerWordsRule == nil {
		return nil
	}
	return r.AbstractFillerWordsRule.Match(sentence)
}

func (r *PortugueseFillerWordsRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || r.AbstractFillerWordsRule == nil {
		return nil
	}
	return r.AbstractFillerWordsRule.MatchList(sentences)
}
