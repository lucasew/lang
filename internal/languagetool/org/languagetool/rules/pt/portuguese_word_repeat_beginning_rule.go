package pt

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// PortugueseWordRepeatBeginningRule ports org.languagetool.rules.pt.PortugueseWordRepeatBeginningRule.
// Extends WordRepeatBeginningRule with a fixed Portuguese ADVERBS surface set for isAdverb,
// overrides category to STYLE (Java super.setCategory(Categories.STYLE)), and ITS Style.
type PortugueseWordRepeatBeginningRule struct {
	*rules.WordRepeatBeginningRule
}

// portugueseWordRepeatBeginningAdverbs ports Java private static final Set<String> ADVERBS
// (exact surfaces, case-sensitive HashSet.contains on token.getToken()). Order matches Java Arrays.asList.
var portugueseWordRepeatBeginningAdverbs = func() map[string]bool {
	m := make(map[string]bool, 129)
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
		m[w] = true
	}
	return m
}()

// NewPortugueseWordRepeatBeginningRule ports PortugueseWordRepeatBeginningRule(messages, language).
// Java Language is unused by this type and the core match loop — Go takes messages only.
func NewPortugueseWordRepeatBeginningRule(messages map[string]string) *PortugueseWordRepeatBeginningRule {
	base := rules.NewWordRepeatBeginningRule(messages)
	// Java @Override getId
	base.IDOverride = "PORTUGUESE_WORD_REPEAT_BEGINNING_RULE"
	// Java: super.setCategory(Categories.STYLE.getCategory(messages));
	// (overrides base WordRepeatBeginningRule REPETITIONS_STYLE)
	base.Category = rules.CatStyle.GetCategory(messages)
	// Java: setLocQualityIssueType(ITSIssueType.Style);
	base.IssueType = rules.ITSStyle
	// Java addExamplePair(Example.wrong(...), Example.fixed(...)) — exact strings
	base.AddExamplePair(
		rules.Wrong("Além disso, a rua é quase completamente residêncial. <marker>Além</marker> disso, foi chamada em nome de um poeta."),
		rules.Fixed("Além disso, a rua é quase completamente residêncial. <marker>Foi</marker> chamada em nome de um poeta."),
	)
	r := &PortugueseWordRepeatBeginningRule{WordRepeatBeginningRule: base}
	// Java @Override isAdverb
	base.IsAdverbFn = r.isAdverb
	return r
}

// isAdverb ports PortugueseWordRepeatBeginningRule.isAdverb:
// return ADVERBS.contains(token.getToken());
func (r *PortugueseWordRepeatBeginningRule) isAdverb(token *languagetool.AnalyzedTokenReadings) bool {
	if token == nil {
		return false
	}
	return portugueseWordRepeatBeginningAdverbs[token.GetToken()]
}
