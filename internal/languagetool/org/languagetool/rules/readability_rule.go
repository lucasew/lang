package rules

import (
	"strconv"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// Java ReadabilityRule constants.
const (
	readabilityMarkWords = 3
	readabilityMinWords  = 10
)

// ReadabilityRule ports org.languagetool.rules.ReadabilityRule (extends TextLevelRule).
// Flesch-Reading-Ease for English; tooEasyTest selects simple vs difficult check.
// Java: Category TEXT_ANALYSIS, ITS Style; setDefaultOff when !defaultOn.
type ReadabilityRule struct {
	Messages map[string]string
	// TooEasyTest ports tooEasyTest (true → too easy / SIMPLE id).
	TooEasyTest bool
	// Level is the threshold 0–6 (Java default 3).
	Level int
	// SingleLineBreaksMarksPara ports Language.getSentenceTokenizer().singleLineBreaksMarksPara()
	// for Tools.isParagraphEnd (Demo/SRX default false).
	SingleLineBreaksMarksPara bool
	// LangCode ports Language for LinguServices.getNumberOfSyllables(…, lang.getDefaultLanguageVariant()).
	// Empty when unused (linguServices == null path).
	LangCode string
	// LinguServices ports userConfig.getLinguServices() (nil → simpleSyllablesCount).
	LinguServices *languagetool.LinguServices
	Category      *Category
	IssueType     ITSIssueType
	DefaultOff    bool

	// Override hooks map Java polymorphic overrides from language subclasses
	// (e.g. GermanReadabilityRule) into MatchList without re-implementing match().
	// Nil → English base behavior (must keep core tests green).
	FleschReadingEaseOverride    func(asl, asw float64) float64
	SimpleSyllablesCountOverride func(word string) int
	MessageOverride              func(level, fre, asl, asw int) string

	// Cumulative stats from last match() (Java instance fields).
	nAllSentences int
	nAllWords     int
	nAllSyllables int
}

// NewReadabilityRule ports ReadabilityRule(messages, lang, null, tooEasyTest, level, false)
// with Demo-like paragraph breaks (singleLine=false). level < 0 → default 3.
func NewReadabilityRule(tooEasy bool, level int) *ReadabilityRule {
	return NewReadabilityRuleFull(nil, nil, tooEasy, level, false)
}

// NewReadabilityRuleFull ports
// ReadabilityRule(ResourceBundle, Language, UserConfig, boolean tooEasyTest, int level, boolean defaultOn).
// Language is carried as SingleLineBreaksMarksPara + LangCode (set by callers when needed).
func NewReadabilityRuleFull(
	messages map[string]string,
	userConfig *languagetool.UserConfig,
	tooEasyTest bool,
	level int,
	defaultOn bool,
) *ReadabilityRule {
	cat := NewCategoryFull(NewCategoryId("TEXT_ANALYSIS"), "Text Analysis", CategoryInternal, false, "")
	r := &ReadabilityRule{
		Messages:    messages,
		TooEasyTest: tooEasyTest,
		Category:    cat,
		IssueType:   ITSStyle,
		DefaultOff:  !defaultOn,
	}

	tmpLevel := -1
	if userConfig != nil {
		r.LinguServices = userConfig.GetLinguServices()
		// Java: userConfig.getConfigValueByID(getId()) — id depends on tooEasyTest.
		if cf := userConfig.GetConfigValueByID(r.GetID()); len(cf) > 0 && cf[0] != nil {
			if n, ok := configValueAsInt(cf[0]); ok {
				tmpLevel = n
			}
		}
	}

	if tmpLevel >= 0 {
		r.Level = tmpLevel
	} else if level >= 0 {
		r.Level = level
	} else {
		r.Level = 3
	}
	return r
}

// GetID ports getId() / getId(tooEasyTest).
func (r *ReadabilityRule) GetID() string {
	return r.GetIDFor(r != nil && r.TooEasyTest)
}

// GetIDFor ports getId(boolean tooEasyTest).
func (r *ReadabilityRule) GetIDFor(tooEasyTest bool) string {
	if tooEasyTest {
		return "READABILITY_RULE_SIMPLE"
	}
	return "READABILITY_RULE_DIFFICULT"
}

// GetDescription ports getDescription().
func (r *ReadabilityRule) GetDescription() string {
	if r != nil && r.TooEasyTest {
		return "Readability: Too easy text"
	}
	return "Readability: Too difficult text"
}

// GetConfigureText ports getConfigureText().
func (r *ReadabilityRule) GetConfigureText() string {
	return "Level of readability 0 (very difficult) to 6 (very easy):"
}

// GetRuleOptions ports getRuleOptions() — default option value 3, range 0–6.
func (r *ReadabilityRule) GetRuleOptions() []*RuleOption {
	return []*RuleOption{NewRuleOption(3, r.GetConfigureText(), 0, 6)}
}

func (r *ReadabilityRule) GetCategory() *Category {
	if r == nil {
		return nil
	}
	return r.Category
}

func (r *ReadabilityRule) GetLocQualityIssueType() ITSIssueType {
	if r == nil || r.IssueType == "" {
		return ITSStyle
	}
	return r.IssueType
}

func (r *ReadabilityRule) IsDefaultOff() bool { return r != nil && r.DefaultOff }

// GetAllSentences ports getAllSentences().
func (r *ReadabilityRule) GetAllSentences() int {
	if r == nil {
		return 0
	}
	return r.nAllSentences
}

// GetAllWords ports getAllWords().
func (r *ReadabilityRule) GetAllWords() int {
	if r == nil {
		return 0
	}
	return r.nAllWords
}

// GetAllSyllables ports getAllSyllables().
func (r *ReadabilityRule) GetAllSyllables() int {
	if r == nil {
		return 0
	}
	return r.nAllSyllables
}

// PrintMessageLevel ports printMessageLevel(level).
func (r *ReadabilityRule) PrintMessageLevel(level int) string {
	var sLevel string
	switch level {
	case 0:
		sLevel = "Very difficult"
	case 1:
		sLevel = "Difficult"
	case 2:
		sLevel = "Fairly difficult"
	case 3:
		sLevel = "Medium"
	case 4:
		sLevel = "Fairly easy"
	case 5:
		sLevel = "Easy"
	case 6:
		sLevel = "Very easy"
	default:
		return ""
	}
	return " {Level " + strconv.Itoa(level) + ": " + sLevel + "}"
}

// GetMessage ports getMessage(level, FRE, ASL, ASW) — FRE/ASL/ASW unused in English template.
// MessageOverride ports language-subclass overrides (e.g. GermanReadabilityRule.getMessage).
func (r *ReadabilityRule) GetMessage(level, fre, asl, asw int) string {
	if r != nil && r.MessageOverride != nil {
		return r.MessageOverride(level, fre, asl, asw)
	}
	simple, few := "difficult", "many"
	if r != nil && r.TooEasyTest {
		simple, few = "simple", "few"
	}
	return "Readability: The text of this paragraph is too " + simple + r.PrintMessageLevel(level) +
		". Too " + few + " words per sentence and too " + few + " syllables per word."
}

// getReadabilityLevel ports private getReadabilityLevel(fre).
// Higher FRE → easier text → higher level (0 very difficult … 6 very easy).
func getReadabilityLevel(fre float64) int {
	if fre < 30 {
		return 0
	} else if fre < 50 {
		return 1
	} else if fre < 60 {
		return 2
	} else if fre < 70 {
		return 3
	} else if fre < 80 {
		return 4
	} else if fre < 90 {
		return 5
	}
	return 6
}

// GetReadabilityLevel exposes getReadabilityLevel for tests / subclasses.
func (r *ReadabilityRule) GetReadabilityLevel(fre float64) int {
	return getReadabilityLevel(fre)
}

// GetFleschReadingEase ports getFleschReadingEase(asl, asw) — English FRE formula.
// Subclasses (DE Amstad, PT) override via FleschReadingEaseOverride.
func (r *ReadabilityRule) GetFleschReadingEase(asl, asw float64) float64 {
	if r != nil && r.FleschReadingEaseOverride != nil {
		return r.FleschReadingEaseOverride(asl, asw)
	}
	return 206.835 - (1.015 * asl) - (84.6 * asw)
}

// isVowel ports private static isVowel(char c).
func isVowelReadability(c rune) bool {
	return c == 'a' || c == 'e' || c == 'i' || c == 'o' || c == 'u' || c == 'y' ||
		c == 'A' || c == 'E' || c == 'I' || c == 'O' || c == 'U' || c == 'Y'
}

// SimpleSyllablesCount ports protected simpleSyllablesCount(String word).
// Uses Java String.charAt / length (UTF-16 code units), double-vowel lastDouble logic,
// and trailing e+s/d / ue rules. Language subclasses set SimpleSyllablesCountOverride.
func (r *ReadabilityRule) SimpleSyllablesCount(word string) int {
	if r != nil && r.SimpleSyllablesCountOverride != nil {
		return r.SimpleSyllablesCountOverride(word)
	}
	return simpleSyllablesCount(word)
}

func simpleSyllablesCount(word string) int {
	n := utf16LenReadability(word)
	if n == 0 {
		return 0
	}
	if n == 1 {
		return 1
	}
	nSyllables := 0
	lastDouble := false
	for i := 0; i < n-1; i++ {
		c := javaCharAtReadability(word, i)
		if isVowelReadability(c) {
			cn := javaCharAtReadability(word, i+1)
			if lastDouble {
				nSyllables++
				lastDouble = false
			} else if ((c == 'e' || c == 'E') && (cn == 'a' || cn == 'o' || cn == 'e' || cn == 'i' || cn == 'y')) ||
				((c == 'a' || c == 'A') && (cn == 'e' || cn == 'i' || cn == 'u')) ||
				((c == 'o' || c == 'O') && (cn == 'o' || cn == 'i' || cn == 'u' || cn == 'a')) ||
				((c == 'u' || c == 'U') && (cn == 'i' || cn == 'a')) ||
				((c == 'i' || c == 'I') && (cn == 'e' || cn == 'o')) {
				lastDouble = true
			} else {
				nSyllables++
				lastDouble = false
			}
		} else {
			lastDouble = false
		}
	}
	c := javaCharAtReadability(word, n-1)
	cl := javaCharAtReadability(word, n-2)
	// Java: if (cl == 'e' && (c == 's' || c == 'd') || cl == 'u' && c == 'e')
	// ⇒ (cl=='e' && (c=='s'||c=='d')) || (cl=='u' && c=='e')
	if (cl == 'e' && (c == 's' || c == 'd')) || (cl == 'u' && c == 'e') {
		nSyllables--
	} else if isVowelReadability(c) && c != 'e' {
		nSyllables++
	}
	if nSyllables <= 0 {
		return 1
	}
	return nSyllables
}

func utf16LenReadability(s string) int {
	return len(utf16.Encode([]rune(s)))
}

func javaCharAtReadability(s string, i int) rune {
	u := utf16.Encode([]rune(s))
	if i < 0 || i >= len(u) {
		return 0
	}
	return rune(u[i])
}

// MatchList ports ReadabilityRule.match(List<AnalyzedSentence>).
// Bug-for-bug control flow including the final return operator-precedence quirk:
//
//	if (nParagraph > 1 && (tooEasyTest && rLevel > level) || (!tooEasyTest && rLevel < level))
//	  return ruleMatches;
//	else
//	  return empty;
func (r *ReadabilityRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*RuleMatch {
	if r == nil {
		return nil
	}
	var ruleMatches []*RuleMatch
	nParagraph := 0
	r.nAllSentences = 0
	r.nAllWords = 0
	r.nAllSyllables = 0
	nSentences := 0
	nWords := 0
	nSyllables := 0
	pos := 0
	startPos := -1
	endPos := -1
	level := r.Level
	tooEasyTest := r.TooEasyTest
	singleLine := r.SingleLineBreaksMarksPara

	for n := 0; n < len(sentences); n++ {
		sentence := sentences[n]
		if sentence == nil {
			// Java would NPE; skip nil for Go safety without changing multi-sentence stats path.
			continue
		}
		tokens := sentence.GetTokensWithoutWhitespace()
		if startPos < 0 && len(tokens) > 1 && tokens[1] != nil {
			startPos = pos + tokens[1].GetStartPos()
		}
		if endPos < 0 && len(tokens) > readabilityMarkWords && tokens[readabilityMarkWords] != nil {
			endPos = pos + tokens[readabilityMarkWords].GetEndPos()
		}
		nSentences++
		for _, token := range tokens {
			if token == nil {
				continue
			}
			sToken := token.GetToken()
			if !token.IsWhitespace() && !token.IsNonWord() {
				nWords++
				if r.LinguServices == nil {
					nSyllables += r.SimpleSyllablesCount(sToken)
				} else {
					nSyllables += r.LinguServices.GetNumberOfSyllables(sToken, r.LangCode)
				}
			}
		}
		if languagetool.IsParagraphEnd(sentences, n, singleLine) {
			if nWords >= readabilityMinWords {
				asl := float64(nWords) / float64(nSentences)
				asw := float64(nSyllables) / float64(nWords)
				fre := r.GetFleschReadingEase(asl, asw)
				rLevel := getReadabilityLevel(fre)
				if (tooEasyTest && rLevel > level) || (!tooEasyTest && rLevel < level) {
					msg := r.GetMessage(rLevel, int(fre), int(asl), int(asw))
					ruleMatches = append(ruleMatches, NewRuleMatch(r, sentence, startPos, endPos, msg))
				}
			}
			r.nAllSentences += nSentences
			r.nAllWords += nWords
			r.nAllSyllables += nSyllables
			nSentences = 0
			nWords = 0
			nSyllables = 0
			startPos = -1
			endPos = -1
			nParagraph++
		}
		pos += sentence.GetCorrectedTextLength()
	}

	// Final FRE/rLevel from cumulative nAll* (Java may divide by zero → Inf/NaN).
	asl := float64(r.nAllWords) / float64(r.nAllSentences)
	asw := float64(r.nAllSyllables) / float64(r.nAllWords)
	fre := r.GetFleschReadingEase(asl, asw)
	rLevel := getReadabilityLevel(fre)
	// Java: if (nParagraph > 1 && (tooEasyTest && rLevel > level) || (!tooEasyTest && rLevel < level))
	// Parse as (nParagraph > 1 && (tooEasyTest && rLevel > level)) || (!tooEasyTest && rLevel < level)
	if nParagraph > 1 && (tooEasyTest && rLevel > level) || (!tooEasyTest && rLevel < level) {
		return ruleMatches
	}
	return []*RuleMatch{}
}

// MinToCheckParagraph ports minToCheckParagraph() → -1.
func (r *ReadabilityRule) MinToCheckParagraph() int { return -1 }
