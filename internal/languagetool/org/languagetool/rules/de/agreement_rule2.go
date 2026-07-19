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
}

func NewAgreementRule2(messages map[string]string) *AgreementRule2 {
	return &AgreementRule2{
		Messages: messages,
		Category: rules.CatGrammar.GetCategory(messages),
	}
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
	imm := r.getSentenceWithImmunization(sentence)
	tokens := imm.GetTokensWithoutWhitespace()
	// Java: scan from start, skip SENT_START and quotes; rule only at sentence start (minus quotes).
	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]
		if tok == nil {
			continue
		}
		if tok.IsSentenceStart() {
			continue
		}
		w := tok.GetToken()
		if isQuoteTok(w) {
			continue
		}
		// first content token
		if i+1 >= len(tokens) {
			break
		}
		// Java: ADJ + SUB tags only (no surface invent)
		if tok.HasPosTagStartingWith("ADJ") && tokens[i+1].HasPosTagStartingWith("SUB") &&
			!tokens[i+1].HasPosTagStartingWith("EIG") {
			if tok.IsImmunized() || tokens[i+1].IsImmunized() || strings.EqualFold(w, "unter") {
				break
			}
			if i+2 < len(tokens) && tokens[i+2] != nil && tokens[i+2].HasPosTagStartingWith("SUB") {
				// "Deutscher Taschenbuch Verlag"
				break
			}
			if rm := r.checkAdjNounAgreement(tok, tokens[i+1], sentence); rm != nil {
				sugs := r.getSuggestions(tokens, i)
				if len(sugs) > 0 {
					rm.SetSuggestedReplacements(sugs)
				}
				return []*rules.RuleMatch{rm}
			}
			break
		}
		// rule only works at sentence start (minus quotes)
		break
	}
	return nil
}

func (r *AgreementRule2) checkAdjNounAgreement(adj, noun *languagetool.AnalyzedTokenReadings, sentence *languagetool.AnalyzedSentence) *rules.RuleMatch {
	set1 := GetAgreementCategories(adj, nil, false)
	set2 := GetAgreementCategories(noun, nil, false)
	if len(set1) == 0 || len(set2) == 0 {
		return nil
	}
	if CategoriesIntersect(set1, set2) {
		return nil
	}
	rm := rules.NewRuleMatch(r, sentence, adj.GetStartPos(), noun.GetEndPos(), agreement2Msg)
	rm.ShortMessage = agreement2Short
	return rm
}

func (r *AgreementRule2) getSuggestions(tokens []*languagetool.AnalyzedTokenReadings, i int) []string {
	if r == nil || r.Synth == nil || i+1 >= len(tokens) {
		return nil
	}
	adjReadings := tokens[i].GetReadings()
	if len(adjReadings) == 0 || adjReadings[0] == nil {
		return nil
	}
	adjToken := adjReadings[0]
	var suggestions []string
	seen := map[string]struct{}{}
	for _, nounToken := range tokens[i+1].GetReadings() {
		if nounToken == nil || nounToken.GetPOSTag() == nil {
			continue
		}
		gender := genderFromPOS(*nounToken.GetPOSTag())
		number := numberFromPOS(*nounToken.GetPOSTag())
		if gender == "" || number == "" {
			continue
		}
		// Java: synthesize(adjToken, "ADJ:NOM:"+number+":"+gender+":GRU:SOL", true)
		tag := "ADJ:NOM:" + number + ":" + gender + ":GRU:SOL"
		forms, err := r.Synth.Synthesize(adjToken, tag)
		if err != nil {
			continue
		}
		for _, s := range forms {
			full := tools.UppercaseFirstChar(s) + " " + nounToken.GetToken()
			if _, ok := seen[full]; ok {
				continue
			}
			seen[full] = struct{}{}
			suggestions = append(suggestions, full)
		}
	}
	return suggestions
}

func genderFromPOS(pos string) string {
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

func numberFromPOS(pos string) string {
	switch {
	case strings.Contains(pos, ":SIN:"):
		return "SIN"
	case strings.Contains(pos, ":PLU:"):
		return "PLU"
	default:
		return ""
	}
}

func isQuoteTok(w string) bool {
	switch w {
	case "\"", "„", "»", "«", "'", "“", "”":
		return true
	}
	return false
}
