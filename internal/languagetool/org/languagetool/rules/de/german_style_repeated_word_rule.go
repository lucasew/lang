package de

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// GermanStyleRepeatedWordRule ports org.languagetool.rules.de.GermanStyleRepeatedWordRule.
// Default off (Java). Compound-part matching optional (testCompoundWords; default false).
type GermanStyleRepeatedWordRule struct {
	*rules.AbstractStyleRepeatedWordRule
	TestCompoundWords bool
	// IsCorrectSpell optional; used when TestCompoundWords (Java Morfologik).
	IsCorrectSpell func(word string) bool
}

var styleLettersRE = regexp.MustCompile(`^[A-Za-zÄÖÜäöüß]+$`)

func NewGermanStyleRepeatedWordRule(messages map[string]string) *GermanStyleRepeatedWordRule {
	base := rules.NewAbstractStyleRepeatedWordRule()
	base.ID = "STYLE_REPEATED_WORD_RULE_DE"
	base.Description = "Wiederholte Worte in aufeinanderfolgenden Sätzen"
	base.MaxDistanceOfSentences = 1
	base.ExcludeDirectSpeech = true
	base.MessageSameSentence = func() string {
		return "Mögliches Stilproblem: Das Wort wird noch einmal im selben Satz verwendet."
	}
	base.MessageSentenceBefore = func() string {
		return "Mögliches Stilproblem: Das Wort wird bereits in einem vorhergehenden Satz verwendet."
	}
	base.MessageSentenceAfter = func() string {
		return "Mögliches Stilproblem: Das Wort wird auch in einem nachfolgenden Satz verwendet."
	}
	// Java: gehe … <marker>gehe</marker> (fixed drops second verb)
	base.AddExamplePair(
		rules.Wrong("Ich gehe zum Supermarkt, danach <marker>gehe</marker> ich nach Hause."),
		rules.Fixed("Ich gehe zum Supermarkt, danach nach Hause."),
	)
	r := &GermanStyleRepeatedWordRule{AbstractStyleRepeatedWordRule: base}
	base.IsTokenToCheck = r.isTokenToCheck
	base.IsTokenPair = r.isTokenPair
	base.IsPartOfWord = r.isPartOfWord
	base.IsExceptionPair = r.isExceptionPair
	// Java GermanStyleRepeatedWordRule.setURL → OpenThesaurus lemma/surface link.
	base.SetURL = germanStyleRepeatedWordURL
	// Java AbstractStyleRepeatedWordRule: STYLE + Style + defaultOff.
	rules.InitStyleRepeatedWordMeta(base, messages)
	// Java: MorfologikSpeller / LinguServices; without dict compound-part checks stay fail-closed.
	r.IsCorrectSpell = func(word string) bool {
		if !FilterDictAvailable() {
			return false
		}
		return !FilterDictIsMisspelled(tools.UppercaseFirstChar(word))
	}
	return r
}

func (r *GermanStyleRepeatedWordRule) GetID() string {
	if r != nil && r.AbstractStyleRepeatedWordRule != nil {
		return r.AbstractStyleRepeatedWordRule.GetID()
	}
	return "STYLE_REPEATED_WORD_RULE_DE"
}

// germanStyleRepeatedWordURL ports GermanStyleRepeatedWordRule.setURL.
// Java: SYNONYMS_URL + single lemma, else token surface.
const germanOpenThesaurusURL = "https://www.openthesaurus.de/synonyme/"

func germanStyleRepeatedWordURL(token *languagetool.AnalyzedTokenReadings) string {
	if token == nil {
		return ""
	}
	var lemmas []string
	for _, rd := range token.GetReadings() {
		if rd == nil {
			continue
		}
		if l := rd.GetLemma(); l != nil && *l != "" {
			lemmas = append(lemmas, *l)
		}
	}
	if len(lemmas) == 1 {
		return germanOpenThesaurusURL + lemmas[0]
	}
	return germanOpenThesaurusURL + token.GetToken()
}

// isUnknownWordStyle ports GermanStyleRepeatedWordRule.isUnknownWord:
// isPosTagUnknown && len>2 && letters only (not invent !isTagged).
func isUnknownWordStyle(token *languagetool.AnalyzedTokenReadings) bool {
	if token == nil || !token.IsPosTagUnknown() {
		return false
	}
	s := token.GetToken()
	// Java token.getToken().length() > 2 (UTF-16 code units).
	return utf16LenDE(s) > 2 && styleLettersRE.MatchString(s)
}

func (r *GermanStyleRepeatedWordRule) isTokenToCheck(tokens []*languagetool.AnalyzedTokenReadings, n int) bool {
	// Java: Frau/Herr + next EIG/unknown → false; else (SUB|EIG|VER|ADJ without PRO/ART/ADV/AUX/MOD) || unknown
	if n <= 0 || n >= len(tokens) || tokens[n] == nil {
		return false
	}
	if n > 0 && n < len(tokens)-1 && tokens[n+1] != nil {
		nextEIG := tokens[n+1].HasPosTagStartingWith("EIG") || isUnknownWordStyle(tokens[n+1])
		if nextEIG {
			switch tokens[n].GetToken() {
			case "Frau", "Fräulein", "Herr", "Herrn", "Lady", "Mister":
				return false
			}
		}
	}
	token := tokens[n]
	ok := (token.MatchesPosTagRegex(`(SUB|EIG|VER|ADJ):.*`) &&
		!token.MatchesPosTagRegex(`(PRO|A(RT|DV)|VER:(AUX|MOD)):.*`)) ||
		isUnknownWordStyle(token)
	if !ok {
		return false
	}
	switch token.GetToken() {
	case "sicher", "weit", "Sie", "Ich", "Euch", "Eure", "Der", "all":
		return false
	}
	return true
}

func (r *GermanStyleRepeatedWordRule) isTokenPair(tokens []*languagetool.AnalyzedTokenReadings, n int, before bool) bool {
	if before {
		if n > 2 && n < len(tokens) && tokens[n-2] != nil && tokens[n-1] != nil && tokens[n] != nil {
			if (tokens[n-2].HasPosTagStartingWith("SUB") && tokens[n-1].HasPosTagStartingWith("PRP") &&
				tokens[n].HasPosTagStartingWith("SUB")) ||
				(tokens[n-2].GetToken() == "hart" && tokens[n-1].GetToken() == "auf" && tokens[n].GetToken() == "hart") ||
				(tokens[n-2].GetToken() == "dicht" && tokens[n-1].GetToken() == "an" && tokens[n].GetToken() == "dicht") ||
				(tokens[n-2].GetToken() == "fressen" && tokens[n-1].GetToken() == "und" && tokens[n].GetToken() == "gefressen") {
				return true
			}
		}
	} else {
		if n > 0 && n < len(tokens)-2 && tokens[n] != nil && tokens[n+1] != nil && tokens[n+2] != nil {
			if (tokens[n].HasPosTagStartingWith("SUB") && tokens[n+1].HasPosTagStartingWith("PRP") &&
				tokens[n+2].HasPosTagStartingWith("SUB")) ||
				(tokens[n].GetToken() == "hart" && tokens[n+1].GetToken() == "auf" && tokens[n+2].GetToken() == "hart") ||
				(tokens[n].GetToken() == "dicht" && tokens[n+1].GetToken() == "an" && tokens[n+2].GetToken() == "dicht") ||
				(tokens[n].GetToken() == "fressen" && tokens[n+1].GetToken() == "und" && tokens[n+2].GetToken() == "gefressen") {
				return true
			}
		}
	}
	return false
}

func (r *GermanStyleRepeatedWordRule) isCorrectSpell(word string) bool {
	if r != nil && r.IsCorrectSpell != nil {
		return r.IsCorrectSpell(word)
	}
	// Fail-closed without hook or dict (Java requires speller).
	if !FilterDictAvailable() {
		return false
	}
	return !FilterDictIsMisspelled(tools.UppercaseFirstChar(word))
}

func (r *GermanStyleRepeatedWordRule) isSecondPartOfWord(testTokenText, tokenText string) bool {
	// Java: testTokenText.length() - tokenText.length() < 3 (UTF-16)
	if utf16LenDE(testTokenText)-utf16LenDE(tokenText) < 3 {
		return false
	}
	lowerTokenText := tools.LowercaseFirstChar(tokenText)
	if lowerTokenText == "frei" ||
		(lowerTokenText == "alten" && strings.HasSuffix(testTokenText, "halten")) {
		return false
	}
	if strings.HasPrefix(tools.LowercaseFirstChar(testTokenText), lowerTokenText) {
		// Java: testTokenText.substring(tokenText.length())
		word := javaUTF16SuffixFrom(testTokenText, utf16LenDE(tokenText))
		if r.isCorrectSpell(word) {
			return true
		}
		if strings.HasPrefix(word, "s") {
			// Java: word.substring(1)
			word = javaUTF16SuffixFrom(word, 1)
			if r.isCorrectSpell(word) {
				return true
			}
		}
		return false
	} else if strings.HasSuffix(testTokenText, lowerTokenText) {
		// Java: substring(0, testTokenText.length() - tokenText.length())
		// endsWith(lowerTokenText): lowerTokenText is ASCII-lower of tokenText first-char change —
		// for German Morphy tokens, suffix match uses UTF-16-safe HasSuffix on equal encoding.
		drop := utf16LenDE(tokenText)
		if utf16LenDE(testTokenText) < drop {
			return false
		}
		word := javaUTF16PrefixDE(testTokenText, utf16LenDE(testTokenText)-drop)
		if r.isCorrectSpell(word) {
			return true
		}
		if strings.HasSuffix(word, "s") {
			// Java: word.substring(word.length() - 1) — last UTF-16 unit only (bug-for-bug)
			word = javaUTF16SuffixFrom(word, utf16LenDE(word)-1)
			if r.isCorrectSpell(word) {
				return true
			}
		}
		return false
	}
	return false
}

func (r *GermanStyleRepeatedWordRule) isPartOfWord(testTokenText, tokenText string) bool {
	// Java: length() < 3 on both (UTF-16)
	if !r.TestCompoundWords || utf16LenDE(testTokenText) < 3 || utf16LenDE(tokenText) < 3 {
		return false
	}
	if utf16LenDE(testTokenText) > utf16LenDE(tokenText) {
		return r.isSecondPartOfWord(testTokenText, tokenText)
	}
	return r.isSecondPartOfWord(tokenText, testTokenText)
}

// javaUTF16PrefixDE ports String.substring(0, n) in UTF-16 units.
func javaUTF16PrefixDE(s string, n int) string {
	u := utf16EncodeDE(s)
	if n <= 0 {
		return ""
	}
	if n > len(u) {
		n = len(u)
	}
	return utf16DecodeDE(u[:n])
}

// javaUTF16SuffixFrom ports String.substring(n) in UTF-16 units.
func javaUTF16SuffixFrom(s string, n int) string {
	u := utf16EncodeDE(s)
	if n <= 0 {
		return s
	}
	if n >= len(u) {
		return ""
	}
	return utf16DecodeDE(u[n:])
}

func (r *GermanStyleRepeatedWordRule) isExceptionPair(token1, token2 *languagetool.AnalyzedTokenReadings) bool {
	if token1 == nil || token2 == nil {
		return false
	}
	if (token1.HasAnyLemma("nah") && token1.HasAnyLemma("nächst") && !token2.HasAnyLemma("nächst")) ||
		(token2.HasAnyLemma("nah") && token2.HasAnyLemma("nächst") && !token1.HasAnyLemma("nächst")) {
		return true
	}
	if token1.HasAnyLemma("gut") &&
		((strings.HasPrefix(token1.GetToken(), "gut") && !strings.HasPrefix(token2.GetToken(), "gut")) ||
			(strings.HasPrefix(token2.GetToken(), "gut") && !strings.HasPrefix(token1.GetToken(), "gut"))) {
		return true
	}
	return false
}

// MatchList delegates to abstract base.
func (r *GermanStyleRepeatedWordRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || r.AbstractStyleRepeatedWordRule == nil {
		return nil
	}
	return r.AbstractStyleRepeatedWordRule.MatchList(sentences)
}
