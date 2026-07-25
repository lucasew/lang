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
	// TestCompoundWords ports Java testCompoundWords (UserConfig cf[2]; default false).
	TestCompoundWords bool
	// messages ports ResourceBundle for getRuleOptions (may be nil).
	messages map[string]string
	// IsCorrectSpell optional; used when TestCompoundWords (Java Morfologik Speller).
	// Default: FilterDictIsMisspelled after UppercaseFirstChar (de_DE.dict).
	IsCorrectSpell func(word string) bool
}

const (
	germanStyleMaxDistanceDefault  = 1
	germanStyleExcludeDirectSpeech = true
	germanStyleTestCompoundDefault = false
	germanStyleSynonymsURL         = "https://www.openthesaurus.de/synonyme/"
	germanStyleRuleID              = "STYLE_REPEATED_WORD_RULE_DE"
	germanStyleDesc                = "Wiederholte Worte in aufeinanderfolgenden Sätzen"
	germanStyleOptExcludeSpeech    = "Direkte Rede und Zitate ausschließen"
	germanStyleOptTestCompound     = "Auch zusammengesetzte Wörter prüfen"
	germanStyleMsgSameSentence     = "Mögliches Stilproblem: Das Wort wird noch einmal im selben Satz verwendet."
	germanStyleMsgSentenceBefore   = "Mögliches Stilproblem: Das Wort wird bereits in einem vorhergehenden Satz verwendet."
	germanStyleMsgSentenceAfter    = "Mögliches Stilproblem: Das Wort wird auch in einem nachfolgenden Satz verwendet."
)

var styleLettersRE = regexp.MustCompile(`^[A-Za-zÄÖÜäöüß]+$`)

// NewGermanStyleRepeatedWordRule ports GermanStyleRepeatedWordRule(messages, lang, null).
func NewGermanStyleRepeatedWordRule(messages map[string]string) *GermanStyleRepeatedWordRule {
	return NewGermanStyleRepeatedWordRuleWithUserConfig(messages, nil)
}

// NewGermanStyleRepeatedWordRuleWithUserConfig ports
// GermanStyleRepeatedWordRule(ResourceBundle messages, Language lang, UserConfig userConfig).
func NewGermanStyleRepeatedWordRuleWithUserConfig(messages map[string]string, userConfig *languagetool.UserConfig) *GermanStyleRepeatedWordRule {
	base := rules.NewAbstractStyleRepeatedWordRule()
	base.ID = germanStyleRuleID
	base.Description = germanStyleDesc
	base.MaxDistanceOfSentences = germanStyleMaxDistanceDefault
	base.ExcludeDirectSpeech = germanStyleExcludeDirectSpeech
	base.MessageSameSentence = func() string { return germanStyleMsgSameSentence }
	base.MessageSentenceBefore = func() string { return germanStyleMsgSentenceBefore }
	base.MessageSentenceAfter = func() string { return germanStyleMsgSentenceAfter }
	// Java: gehe … <marker>gehe</marker> (fixed drops second verb)
	base.AddExamplePair(
		rules.Wrong("Ich gehe zum Supermarkt, danach <marker>gehe</marker> ich nach Hause."),
		rules.Fixed("Ich gehe zum Supermarkt, danach nach Hause."),
	)
	// Java abstract ctor: UserConfig cf[0]/cf[1] via getId() (DE id set above).
	base.ApplyUserConfigFromID(germanStyleRuleID, userConfig)

	r := &GermanStyleRepeatedWordRule{
		AbstractStyleRepeatedWordRule: base,
		TestCompoundWords:             germanStyleTestCompoundDefault,
		messages:                      messages,
	}
	// Java subclass: if (cf != null && cf.length > 2) testCompoundWords = (boolean) cf[2];
	if userConfig != nil {
		if cf := userConfig.GetConfigValueByID(germanStyleRuleID); len(cf) > 2 && cf[2] != nil {
			if v, ok := cf[2].(bool); ok {
				r.TestCompoundWords = v
			}
		}
	}
	base.IsTokenToCheck = r.isTokenToCheck
	base.IsTokenPair = r.isTokenPair
	base.IsPartOfWord = r.isPartOfWord
	base.IsExceptionPair = r.isExceptionPair
	// Java GermanStyleRepeatedWordRule.setURL → OpenThesaurus lemma/surface link.
	base.SetURL = germanStyleRepeatedWordURL
	// Java AbstractStyleRepeatedWordRule: STYLE + Style + defaultOff (+ subclass setCategory STYLE).
	rules.InitStyleRepeatedWordMeta(base, messages)
	// Java: lazy Morfologik Speller(/de/hunspell/de_DE.dict) or LinguServices.
	// Without dict, compound-part checks fail-closed (return false) — isPartOfWord only
	// runs when TestCompoundWords; default-off path never needs the speller.
	r.IsCorrectSpell = func(word string) bool {
		if !FilterDictAvailable() {
			return false
		}
		// Java isCorrectSpell uppercases first then Speller.isMisspelled.
		return !FilterDictIsMisspelled(tools.UppercaseFirstChar(word))
	}
	return r
}

// GetRuleOptions ports GermanStyleRepeatedWordRule.getRuleOptions (3 options).
func (r *GermanStyleRepeatedWordRule) GetRuleOptions() []*rules.RuleOption {
	// Java: messages.getString("guiStyleRepeatedWordText") for option 0;
	// hardcoded DE strings for exclude speech + compound words.
	distText := "Maximum distance between two repeated words (0 - 5 sentences):"
	if r != nil && r.messages != nil {
		if s, ok := r.messages["guiStyleRepeatedWordText"]; ok && s != "" {
			distText = s
		}
	}
	return []*rules.RuleOption{
		rules.NewRuleOption(germanStyleMaxDistanceDefault, distText, 0, 5),
		rules.NewRuleOption(germanStyleExcludeDirectSpeech, germanStyleOptExcludeSpeech, 0, 5),
		rules.NewRuleOption(germanStyleTestCompoundDefault, germanStyleOptTestCompound, 0, 5),
	}
}

func (r *GermanStyleRepeatedWordRule) GetID() string {
	if r != nil && r.AbstractStyleRepeatedWordRule != nil {
		return r.AbstractStyleRepeatedWordRule.GetID()
	}
	return germanStyleRuleID
}

// germanStyleRepeatedWordURL ports GermanStyleRepeatedWordRule.setURL.
// Java: SYNONYMS_URL + single lemma, else token surface.
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
		return germanStyleSynonymsURL + lemmas[0]
	}
	return germanStyleSynonymsURL + token.GetToken()
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
	// Note: n==0 is allowed (SENT_START fails POS); only bounds-check n.
	if n < 0 || n >= len(tokens) || tokens[n] == nil {
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
	// Java: ((matches SUB|EIG|VER|ADJ && !PRO|ART|ADV|AUX|MOD) || isUnknownWord) && !equalsAny(...)
	ok := (token.MatchesPosTagRegex(`(SUB|EIG|VER|ADJ):.*`) &&
		!token.MatchesPosTagRegex(`(PRO|A(RT|DV)|VER:(AUX|MOD)):.*`)) ||
		isUnknownWordStyle(token)
	if !ok {
		return false
	}
	// Java Strings.CS.equalsAny — case-sensitive surface blocklist
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
