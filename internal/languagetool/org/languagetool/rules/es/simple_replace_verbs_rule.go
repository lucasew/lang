package es

import (
	"embed"
	"regexp"
	"strings"
	"sync"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

//go:embed data/replace_verbs.txt
var verbsFS embed.FS

var (
	verbsOnce sync.Once
	verbsMap  map[string][]string

	// Java SimpleReplaceVerbsRule endings / desinencies_1conj_0|1
	verbEndings = "a|aba|abais|aban|abas|ad|ada|adas|ado|ados|amos|an|ando|ar|ara|" +
		"arais|aran|aras|are|areis|aremos|aren|ares|aron|ará|arán|arás|aré|aréis|aría|aríais|aríamos|arían|" +
		"arías|as|ase|aseis|asen|ases|aste|asteis|e|emos|en|es|o|ábamos|áis|áramos|áremos|ásemos|é|éis|ó|" +
		"arse|arme|arte|arlos|arles|arlas|arnos|aros"
	desinencies1Conj0 = regexp.MustCompile("^(.+?)(" + verbEndings + ")$")
	desinencies1Conj1 = regexp.MustCompile("^(.+)(" + verbEndings + ")$")
)

func loadVerbs() map[string][]string {
	verbsOnce.Do(func() {
		f, err := verbsFS.Open("data/replace_verbs.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		verbsMap = m
	})
	return verbsMap
}

// SimpleReplaceVerbsRule ports org.languagetool.rules.es.SimpleReplaceVerbsRule.
// Conjugation stripping is 1:1 with Java; Tag/Synthesize are optional (fail closed when nil).
// IgnoreTaggedWords matches Java setIgnoreTaggedWords().
type SimpleReplaceVerbsRule struct {
	*rules.AbstractSimpleReplaceRule
	// Tag ports SpanishTagger.tag (used on "am"+desinence). Nil → no matches after strip.
	Tag func(words []string) []*languagetool.AnalyzedTokenReadings
	// Synthesize ports SpanishSynthesizer.synthesize. Nil → no synthetic forms.
	Synthesize func(token *languagetool.AnalyzedToken, posTag string) ([]string, error)
}

// NewSimpleReplaceVerbsRule constructs ES_SIMPLE_REPLACE_VERBS without tagger/synth wired.
func NewSimpleReplaceVerbsRule(messages map[string]string) *SimpleReplaceVerbsRule {
	base := &rules.AbstractSimpleReplaceRule{
		Messages:          messages,
		WrongWords:        loadVerbs(),
		CaseSensitive:     false,
		CheckLemmas:       false,
		IgnoreTaggedWords: true, // Java setIgnoreTaggedWords()
		ID:                "ES_SIMPLE_REPLACE_VERBS",
		Description:       "Detecta verbos incorrectos y propone sugerencias.",
		ShortMsg:          "Verbo incorrecto",
		MessageFn: func(tokenStr string, replacements []string) string {
			return "Verbo incorrecto: " + tokenStr
		},
	}
	return &SimpleReplaceVerbsRule{AbstractSimpleReplaceRule: base}
}

// Match ports SimpleReplaceVerbsRule.match (not AbstractSimpleReplaceRule surface lookup alone).
func (r *SimpleReplaceVerbsRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || sentence == nil {
		return nil
	}
	wrong := loadVerbs()
	var ruleMatches []*rules.RuleMatch
	for _, tokenReadings := range sentence.GetTokensWithoutWhitespace() {
		if tokenReadings == nil || tokenReadings.IsSentenceStart() {
			continue
		}
		if r.IgnoreTaggedWords && tokenReadings.IsTagged() {
			continue
		}
		originalTokenStr := tokenReadings.GetToken()
		tokenString := strings.ToLower(originalTokenStr)
		// Spanish locale lower is mostly ToLower for verb endings here.
		var analyzed *languagetool.AnalyzedTokenReadings
		var infinitive string
		for attempt := 0; attempt < 2 && analyzed == nil; attempt++ {
			var m *regexp.Regexp
			if attempt == 0 {
				m = desinencies1Conj0
			} else {
				m = desinencies1Conj1
			}
			sub := m.FindStringSubmatch(tokenString)
			if sub == nil {
				continue
			}
			lexeme, desinence := sub[1], sub[2]
			// orthographic adjustments before -e/-é/-i/-ï endings
			if startsWithEI(desinence) {
				lexeme = adjustLexemeBeforeEI(lexeme)
			}
			infinitive = lexeme + "ar"
			if _, ok := wrong[infinitive]; !ok {
				continue
			}
			// Java: tagger.tag(["am"+desinence]); require non-null POS on first reading
			if r.Tag == nil {
				continue
			}
			tagged := r.Tag([]string{"am" + desinence})
			if len(tagged) == 0 || tagged[0] == nil {
				continue
			}
			first := tagged[0].GetAnalyzedToken(0)
			if first == nil || first.GetPOSTag() == nil {
				continue
			}
			analyzed = tagged[0]
		}
		if analyzed == nil {
			continue
		}
		// synthesize replacements
		var possible []string
		replacementInfinitives := wrong[infinitive]
		for _, replacementInfinitive := range replacementInfinitives {
			if strings.HasPrefix(replacementInfinitive, "(") {
				possible = appendUnique(possible, replacementInfinitive)
				continue
			}
			parts := strings.Split(replacementInfinitive, " ")
			lemma := parts[0]
			posTemplate := "V.*"
			infinitiveAsAnTkn := languagetool.NewAnalyzedToken(lemma, &posTemplate, &lemma)
			for _, analyzedToken := range analyzed.GetReadings() {
				if analyzedToken == nil || analyzedToken.GetPOSTag() == nil {
					continue
				}
				posTag := *analyzedToken.GetPOSTag()
				if lemma == "haver" && len(posTag) >= 2 {
					posTag = "VA" + posTag[2:]
				}
				if r.Synthesize == nil {
					continue
				}
				synthesized, err := r.Synthesize(infinitiveAsAnTkn, posTag)
				if err != nil || len(synthesized) == 0 {
					continue
				}
				for _, s := range synthesized {
					for j := 1; j < len(parts); j++ {
						s = s + " " + parts[j]
					}
					possible = appendUnique(possible, s)
				}
			}
		}
		if len(possible) == 0 {
			continue
		}
		// Case-adjust suggestions like AbstractSimpleReplaceRule
		if tools.IsAllUppercase(originalTokenStr) {
			for i, s := range possible {
				if !strings.HasPrefix(s, "(") {
					possible[i] = strings.ToUpper(s)
				}
			}
		} else if startsUpper(originalTokenStr) {
			for i, s := range possible {
				if !strings.HasPrefix(s, "(") {
					possible[i] = tools.UppercaseFirstChar(s)
				}
			}
		}
		rm := rules.NewRuleMatch(r, sentence, tokenReadings.GetStartPos(), tokenReadings.GetEndPos(),
			"Verbo incorrecto: "+originalTokenStr)
		rm.ShortMessage = "Verbo incorrecto"
		rm.SetSuggestedReplacements(possible)
		ruleMatches = append(ruleMatches, rm)
	}
	return ruleMatches
}

func startsWithEI(desinence string) bool {
	if desinence == "" {
		return false
	}
	r := []rune(desinence)[0]
	return r == 'e' || r == 'é' || r == 'i' || r == 'ï'
}

func adjustLexemeBeforeEI(lexeme string) string {
	switch {
	case strings.HasSuffix(lexeme, "c"):
		return lexeme[:len(lexeme)-1] + "z"
	case strings.HasSuffix(lexeme, "qu"):
		return lexeme[:len(lexeme)-2] + "c"
	case strings.HasSuffix(lexeme, "g"):
		return lexeme[:len(lexeme)-1] + "j"
	case strings.HasSuffix(lexeme, "gü"):
		return lexeme[:len(lexeme)-2] + "gu"
	case strings.HasSuffix(lexeme, "gu"):
		return lexeme[:len(lexeme)-2] + "g"
	default:
		return lexeme
	}
}

func appendUnique(dst []string, s string) []string {
	for _, x := range dst {
		if x == s {
			return dst
		}
	}
	return append(dst, s)
}

func startsUpper(s string) bool {
	for _, r := range s {
		return unicode.IsUpper(r)
	}
	return false
}
