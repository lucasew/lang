package pt

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// PortugueseWordRepeatBeginningRule ports org.languagetool.rules.pt.PortugueseWordRepeatBeginningRule.
type PortugueseWordRepeatBeginningRule struct {
	*rules.WordRepeatBeginningRule
}

var portugueseAdverbs = map[string]bool{}

func init() {
	for _, w := range []string{
		"Abaixo", "Acaso", "Acima", "Acolá",
		"Ademais", "Adentro", "Adiante", "Adicionalmente",
		"Afinal", "Afora", "Agora", "Aí", "Ainda",
		"Além", "Algures", "Ali", "Aliás",
		"Amanhã", "Amiúde", "Antigamente", "Aonde",
		"Apenas", "Apesar", "Aquém", "Aqui", "Assaz",
		"Assim", "Até", "Atrás", "Bastante", "Bem",
		"Bondosamente", "Breve", "Cá", "Casualmente",
		"Cedo", "Certamente", "Certo", "Constantemente",
		"Cuidadosamente", "Dantes", "Debaixo", "Debalde",
		"Decerto", "Defronte", "Demais", "Demasiado",
		"Dentro", "Depois", "Depressa", "Detrás", "Devagar",
		"Doravante", "E", "Efetivamente", "Embaixo",
		"Embora", "Enfim", "Então", "Entrementes",
		"Exclusivamente", "Externamente", "Fora",
		"Frequentemente", "Generosamente", "Hoje",
		"Imediatamente", "Inclusivamente", "Inda", "Já",
		"Jamais", "Lá", "Logo", "Longe", "Mais", "Mal",
		"Mas", "Melhor", "Menos", "Mesmo", "Muito", "Não",
		"Nem", "Nenhures", "Nunca", "Onde", "Ontem", "Ora",
		"Ou", "Outra", "Outro", "Outrora", "Outrossim",
		"Perto", "Pior", "Porventura", "Possivelmente",
		"Pouco", "Primeiramente", "Primeiro",
		"Principalmente", "Provavelmente", "Provisoriamente",
		"Quanto", "Quão", "Quase", "Quiçá", "Realmente",
		"Salvo", "Seguidamente", "Sempre", "Senão", "Será",
		"Sim", "Simplesmente", "Só", "Sobremaneira",
		"Sobremodo", "Sobretudo", "Somente", "Sucessivamente",
		"Talvez", "Também", "Tampouco", "Tanto", "Tão",
		"Tarde", "Ultimamente", "Unicamente",
	} {
		portugueseAdverbs[w] = true
	}
}

func NewPortugueseWordRepeatBeginningRule(messages map[string]string) *PortugueseWordRepeatBeginningRule {
	base := rules.NewWordRepeatBeginningRule(messages)
	base.IDOverride = "PORTUGUESE_WORD_REPEAT_BEGINNING_RULE"
	// Java: Além → Foi
	base.AddExamplePair(
		rules.Wrong("Além disso, a rua é quase completamente residêncial. <marker>Além</marker> disso, foi chamada em nome de um poeta."),
		rules.Fixed("Além disso, a rua é quase completamente residêncial. <marker>Foi</marker> chamada em nome de um poeta."),
	)
	r := &PortugueseWordRepeatBeginningRule{WordRepeatBeginningRule: base}
	base.IsAdverbFn = r.isAdverb
	return r
}

func (r *PortugueseWordRepeatBeginningRule) isAdverb(token *languagetool.AnalyzedTokenReadings) bool {
	return portugueseAdverbs[token.GetToken()]
}
