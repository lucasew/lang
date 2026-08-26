package de

import (
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// AgreementRule2 ports org.languagetool.rules.de.AgreementRule2:
// SENT_START (minus quotes) ADJ + SUB agreement.
// Java: hasPosTagStartingWith ADJ/SUB only — no surface invent.
type AgreementRule2 struct {
	Messages map[string]string
	// Category ports setCategory(GRAMMAR).
	Category *rules.Category
	// Synth optional for ADJ:NOM suggestions (Java GermanSynthesizer.INSTANCE).
	Synth synthesis.Synthesizer
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []rules.IncorrectExample
	correctExamples   []rules.CorrectExample
}

func NewAgreementRule2(messages map[string]string) *AgreementRule2 {
	r := &AgreementRule2{
		Messages: messages,
		Category: rules.CatGrammar.GetCategory(messages),
	}
	// Java: Kleiner Haus → Kleines Haus
	r.AddExamplePair(
		rules.Wrong("<marker>Kleiner Haus</marker> am Waldrand"),
		rules.Fixed("<marker>Kleines Haus</marker> am Waldrand"),
	)
	return r
}

// AddExamplePair ports Rule.addExamplePair.
func (r *AgreementRule2) AddExamplePair(incorrect rules.IncorrectExample, correct rules.CorrectExample) {
	if r == nil {
		return
	}
	var br rules.BaseRule
	br.AddExamplePair(incorrect, correct)
	r.incorrectExamples = append(r.incorrectExamples, br.GetIncorrectExamples()...)
	r.correctExamples = append(r.correctExamples, br.GetCorrectExamples()...)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *AgreementRule2) GetIncorrectExamples() []rules.IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]rules.IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *AgreementRule2) GetCorrectExamples() []rules.CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]rules.CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

// WithSynth sets optional synthesizer for suggestions.
func (r *AgreementRule2) WithSynth(s synthesis.Synthesizer) *AgreementRule2 {
	if r != nil {
		r.Synth = s
	}
	return r
}

func (r *AgreementRule2) GetID() string { return "DE_AGREEMENT2" }

// GetDescription ports AgreementRule2.getDescription.
func (r *AgreementRule2) GetDescription() string {
	return "Kongruenz von Adjektiv und Nomen (unvollständig!), z.B. 'kleiner (kleines) Haus'"
}

func (r *AgreementRule2) GetCategory() *rules.Category {
	if r == nil {
		return nil
	}
	return r.Category
}

// EstimateContextForSureMatch ports estimateContextForSureMatch (max ANTI_PATTERNS length).
func (r *AgreementRule2) EstimateContextForSureMatch() int {
	max := 0
	for _, ap := range AgreementRule2AntiPatterns {
		if n := len(ap); n > max {
			max = n
		}
	}
	return max
}

// GetAntiPatterns ports AgreementRule2.getAntiPatterns (cached ANTI_PATTERNS IMMUNIZE rules).
func (r *AgreementRule2) GetAntiPatterns() []*disambigrules.DisambiguationPatternRule {
	return agreementRule2AntiPatterns()
}

const (
	agreement2Msg = "Möglicherweise fehlende grammatikalische Übereinstimmung zwischen Adjektiv und " +
		"Nomen bezüglich Kasus, Numerus oder Genus. Beispiel: 'kleiner Haus' statt 'kleines Haus'"
	agreement2Short = "Möglicherweise keine Übereinstimmung bezüglich Kasus, Numerus oder Genus"
)

// agreementRule2AdjGru ports AgreementRule2.ADJ_GRU (used in anti-patterns).
const agreementRule2AdjGru = "Allgemein|Ausgiebig|Stilvoll|Link|Direkt|Gegenseitig|Offensichtlich|Weitgehend|Frei|Prinzipiell|Regelrecht|Kostenlos|Gleichzeitig|Ganzjährig|Überraschend|Entsprechend|Ordentlich|Gelangweilt"

// AgreementRule2AntiPatterns ports AgreementRule2.ANTI_PATTERNS (from Java).
var AgreementRule2AntiPatterns = [][]*patterns.PatternToken{
	{
		patterns.CsRegex("Willkommen|Link|Aktuell|Diverse|Flächendeckend|Entsprechende|Angeblich|Gelegentlich|Antizyklisch|Unbedingt|Zusätzlich|Natürlich|Äußerlich|Erfolgreich|" +
			"Spät|Länger|Vorrangig|Rechtzeitig|Typisch|Allwöchentlich|Wöchentlich|Inhaltlich|Tagtäglich|Täglich|Komplett|" +
			"Genau|Gerade|Bewusst|Vereinzelt|Gänzlich|Ständig|Okay|Meist|Generell|Ausreichend|Genügend|Reichlich|" +
			"Regelmäßig(e|es)?|Unregelmäßig|Hauptsächlich"),
		patterns.PosRegex("SUB:.*"),
	},
	{patterns.CsRegex(agreementRule2AdjGru), patterns.PosRegex("SUB:.*"), patterns.PosRegex("VER:.*")},
	{patterns.CsRegex(agreementRule2AdjGru), patterns.PosRegex("SUB:.*"), patterns.PosRegex("PRP.*")},
	{patterns.CsRegex(agreementRule2AdjGru), patterns.PosRegex("SUB:.*"), patterns.Token(",")},
	{patterns.CsRegex("Gut|Schlecht|Existenziell|Ganz|Gering|Viel|Wenig"), patterns.PosRegex("SUB:.*ADJ")},
	{patterns.Regex("Nachhaltig|Direkt"), patterns.PosRegex("SUB:NOM:.*"), patterns.PosRegex("VER:INF:(SFT|NON)")},
	{patterns.Regex(`\d0er`), patterns.Regex("Jahren?")},
	{patterns.Token("Liebe"), patterns.Token("Mai")},
	{patterns.Token("Ganz"), patterns.Token("Ohr")},
	{patterns.Token("Klar"), patterns.Token("Schiff")},
	{patterns.Token("Echt"), patterns.TokenRegex("Scheiße|Mist")},
	{patterns.Token("Dickes"), patterns.Token("Danke")},
	{patterns.Token("Personal"), patterns.Token("Shopper")},
	{patterns.Token("Schwäbisch"), patterns.Token("Hall")},
	{patterns.Token("Herzlich"), patterns.Token("Willkommen")},
	{patterns.Token("Gut"), patterns.TokenRegex("Ding|Holz")},
	{patterns.Token("Urban"), patterns.Token("Mining")},
	{patterns.Token("Responsive"), patterns.Token("Design")},
	{patterns.Token("Dual"), patterns.Token("Studierende")},
	{patterns.Token("Deutsche"), patterns.CsRegex("Grammophon|Wohnen")},
	{patterns.PosRegex("ADJ.*"), patterns.TokenRegex(".+beamte")},
	{patterns.NewPatternTokenBuilder().Token("Anderen").SetSkip(5).Build(), patterns.PosRegex("VER:INF:(SFT|NON)")},
	{patterns.Regex("echt|absolut|voll|total"), patterns.Regex("Wahnsinn|Klasse")},
	{patterns.Pos("SENT_START"), patterns.Pos("ADJ:PRD:GRU"), patterns.PosRegex("SUB:NOM:SIN:NEU:INF")},
	{patterns.TokenRegex("voll|voller"), patterns.PosRegex("SUB:NOM:SIN:.*")},
	{patterns.Token("einzig"), patterns.PosRegex("SUB:NOM:.*")},
	{patterns.TokenRegex("Intelligent|Urban"), patterns.Token("Design")},
	{patterns.Token("Alternativ"), patterns.Token("Berufserfahrung")},
	{patterns.Token("Maritim"), patterns.Token("Hotel")},
	{patterns.CsToken("Russisch"), patterns.CsToken("Brot")},
	{patterns.Token("ruhig"), patterns.CsToken("Blut")},
	{patterns.Token("Blind"), patterns.Regex("Dates?")},
	{patterns.Token("Fair"), patterns.Token("Trade")},
	{patterns.Token("Frei"), patterns.Token("Haus")},
	{patterns.Token("Global"), patterns.Token("Player")},
	{patterns.Token("psychisch"), patterns.Regex("Kranken?")},
	{patterns.Token("sportlich"), patterns.Regex("Aktiven?")},
	{patterns.Token("politisch"), patterns.Regex("Interessierten?")},
	{patterns.Token("voraussichtlich"), patterns.Regex("Ende|Anfang")},
	{patterns.Regex("gesetzlich|privat|freiwillig"), patterns.Regex("(Kranken)?Versicherten?")},
	{patterns.Token("typisch"), patterns.PosRegex("SUB:.*"), patterns.Regex("[!?.]")},
	{patterns.Token("lecker"), patterns.Token("Essen")},
	{patterns.Token("erneut"), patterns.PosRegex("SUB:.*")},
	{patterns.Token("Gesetzlich"), patterns.Regex("Krankenversicherten?")},
	{patterns.Token("weitgehend"), patterns.Token("Einigkeit")},
	{patterns.Token("Ernst")},
	{patterns.Token("Anders")},
	{patterns.Token("wirklich")},
	{patterns.Token("gemeinsam")},
	{patterns.Token("wenig")},
	{patterns.Token("weniger")},
	{patterns.Token("unaufgefordert")},
	{patterns.Token("richtig")},
	{patterns.Token("weiß")},
	{patterns.Token("speziell")},
	{patterns.Token("proaktiv")},
	{patterns.Token("halb")},
	{patterns.Token("hinter")},
	{patterns.Token("vermutlich")},
	{patterns.Token("eventuell")},
	{patterns.Token("ausschließlich")},
	{patterns.Token("ausschliesslich")},
	{patterns.Token("bloß")},
	{patterns.Token("einfach")},
	{patterns.Token("egal")},
	{patterns.Token("endlich")},
	{patterns.Token("unbemerkt")},
	{patterns.Token("Typisch"), patterns.TokenRegex("Mann|Frau")},
	{patterns.Token("Ausreichend"), patterns.TokenRegex("Bewegung")},
	{patterns.Token("Genau"), patterns.Token("Null")},
	{patterns.Token("wohl")},
	{patterns.Token("erst")},
	{patterns.Token("lieber")},
	{patterns.Token("besser")},
	{patterns.Token("laut")},
	{patterns.Token("research")},
	{patterns.Token("researchs")},
	{patterns.Token("security")},
	{patterns.Token("business")},
	{patterns.Token("Universal")},
	{patterns.Token("voll"), patterns.Token("Sorge")},
	{patterns.Token("Total"), patterns.TokenRegex("Tankstellen?")},
	{patterns.Token("Ganz"), patterns.Token("Gentleman")},
	{patterns.Token("Kurz"), patterns.Token("Zeit"), patterns.TokenRegex("für|um")},
	{patterns.Token("Golden"), patterns.Token("Gate")},
	{patterns.Token("Wirtschaftlich"), patterns.TokenRegex("Berechtigte[rn]?")},
	{patterns.Token("Russisch"), patterns.Token("Roulette")},
	{patterns.Token("Clever"), patterns.TokenRegex("Shuttles?")},
	{patterns.Token("Personal"), patterns.TokenRegex("(Computer|Coach|Trainer|Brand).*")},
	{
		patterns.TokenRegex("Digital|Fair|Regional|Global|Bilingual|International|National|Visual|Final|Rapid|Dual|Golden|Human"),
		patterns.TokenRegex("(Initiative|Office|Connection|Bootcamp|Leadership|Sales|Community|Service|Management|Board|Identity|City|Paper|Transfer|Transformation|Power|Shopping|Brand|Master|Gate|Drive|Learning|Publishing|Signage|Value|Entertainment|Museum|Register|Society|Union|Institute|Symposium|Style|Design|Edition).*"),
	},
	{patterns.Token("Smart")},
	{patterns.Token("International"), patterns.TokenRegex("Society|Olympic|Space")},
	{patterns.Token("GmbH")},
}

var (
	agreementRule2AntiOnce  sync.Once
	agreementRule2AntiRules []*disambigrules.DisambiguationPatternRule
)

func agreementRule2AntiPatterns() []*disambigrules.DisambiguationPatternRule {
	agreementRule2AntiOnce.Do(func() {
		aps := AgreementRule2AntiPatterns
		agreementRule2AntiRules = make([]*disambigrules.DisambiguationPatternRule, 0, len(aps))
		for _, toks := range aps {
			if len(toks) == 0 {
				continue
			}
			rule := disambigrules.NewDisambiguationPatternRule(
				"INTERNAL_ANTIPATTERN", "(no description)", "de",
				toks, "", nil, disambigrules.ActionImmunize,
			)
			agreementRule2AntiRules = append(agreementRule2AntiRules, rule)
		}
	})
	return agreementRule2AntiRules
}

func (r *AgreementRule2) getSentenceWithImmunization(sentence *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if sentence == nil {
		return nil
	}
	aps := agreementRule2AntiPatterns()
	if len(aps) == 0 {
		return sentence
	}
	src := sentence.GetTokens()
	cloned := make([]*languagetool.AnalyzedTokenReadings, len(src))
	for i, t := range src {
		if t == nil {
			continue
		}
		cloned[i] = languagetool.NewAnalyzedTokenReadingsFromOld(t, t.GetReadings(), "")
	}
	immunized := languagetool.NewAnalyzedSentence(cloned)
	for _, ap := range aps {
		if ap != nil {
			immunized = ap.Replace(immunized)
		}
	}
	return immunized
}

func (r *AgreementRule2) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if sentence == nil {
		return nil
	}
	// Java: List<RuleMatch> ruleMatches = new ArrayList<>(); ... return toRuleMatchArray(ruleMatches);
	var ruleMatches []*rules.RuleMatch
	// Java: AnalyzedTokenReadings[] tokens = getSentenceWithImmunization(sentence).getTokensWithoutWhitespace();
	tokens := r.getSentenceWithImmunization(sentence).GetTokensWithoutWhitespace()
	for i := 0; i < len(tokens); i++ {
		if tokens[i] == nil {
			continue
		}
		token := tokens[i].GetToken()
		// Java: if (!tokens[i].isSentenceStart() && !StringUtils.equalsAny(token, "\"", "„", "»", "«"))
		if tokens[i].IsSentenceStart() || isAgreementRule2Quote(token) {
			// skip SENT_START and quotes (continue loop)
			continue
		}
		// skip quotes, as these are not relevant
		if i+1 < len(tokens) &&
			tokens[i].HasPosTagStartingWith("ADJ") &&
			tokens[i+1].HasPosTagStartingWith("SUB") &&
			!tokens[i+1].HasPosTagStartingWith("EIG") {
			// Java: if (immunized || unter) continue;  — not break
			if tokens[i].IsImmunized() || tokens[i+1].IsImmunized() ||
				strings.EqualFold(tokens[i].GetToken(), "unter") {
				continue
			}
			var nextToken *languagetool.AnalyzedTokenReadings
			if i+2 < len(tokens) {
				nextToken = tokens[i+2]
			}
			if nextToken != nil && nextToken.HasPosTagStartingWith("SUB") {
				// no alarm for e.g. "Deutscher Taschenbuch Verlag"
				break
			}
			ruleMatch := r.checkAdjNounAgreement(tokens[i], tokens[i+1], sentence)
			if ruleMatch != nil {
				// Java always setSuggestedReplacements (may be empty list)
				ruleMatch.SetSuggestedReplacements(r.getSuggestions(tokens, i))
				ruleMatches = append(ruleMatches, ruleMatch)
				break
			}
			// Java: when agreement OK, do not break — fall through to next i
		} else {
			// rule only works at sentence start (minus quotes)
			break
		}
	}
	return ruleMatches
}

// checkAdjNounAgreement ports AgreementRule2.checkAdjNounAgreement.
func (r *AgreementRule2) checkAdjNounAgreement(token1, token2 *languagetool.AnalyzedTokenReadings, sentence *languagetool.AnalyzedSentence) *rules.RuleMatch {
	set := r.retainCommonCategories(token1, token2)
	if len(set) == 0 {
		// Java: new RuleMatch(this, sentence, token1.getStartPos(), token2.getEndPos(), msg, shortMsg)
		rm := rules.NewRuleMatch(r, sentence, token1.GetStartPos(), token2.GetEndPos(), agreement2Msg)
		rm.ShortMessage = agreement2Short
		return rm
	}
	return nil
}

// retainCommonCategories ports AgreementRule2.retainCommonCategories.
// Java: set1.retainAll(set2); empty common set → agreement error.
// Note: empty set1 or set2 also yields empty retainAll (Java flags mismatch; do not early-return).
func (r *AgreementRule2) retainCommonCategories(token1, token2 *languagetool.AnalyzedTokenReadings) map[string]struct{} {
	// Java: Set<AgreementRule.GrammarCategory> categoryToRelaxSet = Collections.emptySet();
	set1 := GetAgreementCategories(token1, nil, false)
	set2 := GetAgreementCategories(token2, nil, false)
	// Java: set1.retainAll(set2)
	for k := range set1 {
		if _, ok := set2[k]; !ok {
			delete(set1, k)
		}
	}
	return set1
}

// getSuggestions ports AgreementRule2.getSuggestions.
// Java: GermanSynthesizer.INSTANCE.synthesize(adjToken, "ADJ:NOM:"+number+":"+gender+":GRU:SOL", true)
func (r *AgreementRule2) getSuggestions(tokens []*languagetool.AnalyzedTokenReadings, i int) []string {
	suggestions := []string{}
	if r == nil || r.Synth == nil || i+1 >= len(tokens) || tokens[i] == nil || tokens[i+1] == nil {
		return suggestions
	}
	// Java: AnalyzedToken adjToken = tokens[i].getAnalyzedToken(0);
	adjToken := tokens[i].GetAnalyzedToken(0)
	if adjToken == nil {
		return suggestions
	}
	for _, nounToken := range tokens[i+1].GetReadings() {
		if nounToken == nil || nounToken.GetPOSTag() == nil {
			continue
		}
		gender := agreementRule2GetGender(nounToken)
		if gender == "" {
			continue
		}
		number := agreementRule2GetNumber(nounToken)
		if number == "" {
			continue
		}
		// Java third arg true → posTagRegExp
		tag := "ADJ:NOM:" + number + ":" + gender + ":GRU:SOL"
		forms, err := r.Synth.SynthesizeRE(adjToken, tag, true)
		if err != nil {
			// Java: catch (IOException e) { throw new RuntimeException(e); }
			panic(err)
		}
		for _, s := range forms {
			fullSugg := tools.UppercaseFirstChar(s) + " " + nounToken.GetToken()
			// Java: if (!suggestions.contains(fullSugg))
			found := false
			for _, existing := range suggestions {
				if existing == fullSugg {
					found = true
					break
				}
			}
			if !found {
				suggestions = append(suggestions, fullSugg)
			}
		}
	}
	return suggestions
}

// agreementRule2GetGender ports AgreementRule2.getGender.
func agreementRule2GetGender(nounToken *languagetool.AnalyzedToken) string {
	if nounToken == nil || nounToken.GetPOSTag() == nil {
		return ""
	}
	pos := *nounToken.GetPOSTag()
	switch {
	case strings.Contains(pos, ":MAS"):
		return "MAS"
	case strings.Contains(pos, ":FEM"):
		return "FEM"
	case strings.Contains(pos, ":NEU"):
		return "NEU"
	default:
		return ""
	}
}

// agreementRule2GetNumber ports AgreementRule2.getNumber.
func agreementRule2GetNumber(nounToken *languagetool.AnalyzedToken) string {
	if nounToken == nil || nounToken.GetPOSTag() == nil {
		return ""
	}
	pos := *nounToken.GetPOSTag()
	switch {
	case strings.Contains(pos, ":SIN:"):
		return "SIN"
	case strings.Contains(pos, ":PLU:"):
		return "PLU"
	default:
		return ""
	}
}

// isAgreementRule2Quote ports StringUtils.equalsAny(token, "\"", "„", "»", "«") in match().
// Only these four — no invent of curly/ASCII-single quotes.
func isAgreementRule2Quote(w string) bool {
	switch w {
	case "\"", "„", "»", "«":
		return true
	}
	return false
}
