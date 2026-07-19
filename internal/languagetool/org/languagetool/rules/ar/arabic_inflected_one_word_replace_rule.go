package ar

import (
	"embed"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/inflected_one_word.txt
var inflectedOneWordFS embed.FS

var (
	inflectedOnce sync.Once
	inflectedMap  map[string]rules.SuggestionWithMessage
)

func loadInflectedOneWord() map[string]rules.SuggestionWithMessage {
	inflectedOnce.Do(func() {
		inflectedMap = parseARInflectedMap()
	})
	return inflectedMap
}

func parseARInflectedMap() map[string]rules.SuggestionWithMessage {
	b, err := inflectedOneWordFS.ReadFile("data/inflected_one_word.txt")
	if err != nil {
		panic(err)
	}
	out := map[string]rules.SuggestionWithMessage{}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		msg := ""
		if len(parts) == 2 {
			msg = parts[1]
		}
		kv := strings.SplitN(parts[0], "=", 2)
		if len(kv) < 2 {
			continue
		}
		wrong := strings.TrimSpace(kv[0])
		sug := strings.TrimSpace(kv[1])
		out[wrong] = rules.SuggestionWithMessage{Suggestion: sug, Message: msg}
	}
	return out
}

// ArabicInflectedOneWordReplaceRule ports
// org.languagetool.rules.ar.ArabicInflectedOneWordReplaceRule.
// Match is lemma + POS only (Java getSuggestedWords); without POS/lemma fail closed.
// InflectLemmaLike optional synthesizer for suggestions (Java ArabicSynthesizer).
type ArabicInflectedOneWordReplaceRule struct {
	Messages map[string]string
	words    map[string]rules.SuggestionWithMessage
	// InflectLemmaLike ports synthesizer.inflectLemmaLike(targetLemma, sourceToken).
	// When nil, bare dictionary suggestions are used (no surface clitic invent).
	InflectLemmaLike func(targetLemma string, source *languagetool.AnalyzedToken) []string
}

func NewArabicInflectedOneWordReplaceRule(messages map[string]string) *ArabicInflectedOneWordReplaceRule {
	return &ArabicInflectedOneWordReplaceRule{
		Messages: messages,
		words:    loadInflectedOneWord(),
	}
}

func (r *ArabicInflectedOneWordReplaceRule) GetID() string { return "AR_INFLECTED_ONE_WORD" }

func (r *ArabicInflectedOneWordReplaceRule) GetDescription() string {
	return "قاعدة تطابق الكلمات التي يجب تجنبها وتقترح تصويبا لها"
}

// Match ports ArabicInflectedOneWordReplaceRule.match.
func (r *ArabicInflectedOneWordReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || sentence == nil || len(r.words) == 0 {
		return nil
	}
	tokens := sentence.GetTokensWithoutWhitespace()
	var matches []*rules.RuleMatch
	for i := 1; i < len(tokens); i++ {
		tok := tokens[i]
		if tok == nil {
			continue
		}
		for _, wordTok := range tok.GetReadings() {
			if wordTok == nil {
				continue
			}
			swm := r.getSuggestedWords(wordTok)
			if swm == nil {
				continue
			}
			propositions := strings.Split(swm.Suggestion, "|")
			sugMsg := swm.Message
			var forms []string
			for _, prop := range propositions {
				prop = strings.TrimSpace(prop)
				if prop == "" {
					continue
				}
				if r.InflectLemmaLike != nil {
					forms = append(forms, r.InflectLemmaLike(prop, wordTok)...)
				} else {
					// without synthesizer: bare replacement lemmas (Java always has synth)
					forms = append(forms, prop)
				}
			}
			if len(forms) == 0 {
				continue
			}
			// Message template simplified to short form used in tests
			msg := " لا تقل '" + tok.GetToken() + "' بل قل: " + strings.Join(forms, " أو ")
			if sugMsg != "" {
				msg = sugMsg + " " + strings.Join(forms, " أو ")
			}
			rm := rules.NewRuleMatch(r, sentence, tok.GetStartPos(), tok.GetEndPos(), msg)
			rm.ShortMessage = "خطأ، يفضل أن  يقال:"
			rm.SetSuggestedReplacements(forms)
			matches = append(matches, rm)
			break // one match per token surface (Java can add multiple per reading)
		}
	}
	return matches
}

// getSuggestedWords ports getSuggestedWords: requires POS + lemma in dictionary.
func (r *ArabicInflectedOneWordReplaceRule) getSuggestedWords(mytoken *languagetool.AnalyzedToken) *rules.SuggestionWithMessage {
	if mytoken == nil || mytoken.GetPOSTag() == nil {
		return nil
	}
	if mytoken.GetLemma() == nil || *mytoken.GetLemma() == "" {
		return nil
	}
	if swm, ok := r.words[*mytoken.GetLemma()]; ok {
		cp := swm
		return &cp
	}
	return nil
}
