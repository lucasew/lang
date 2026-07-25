package de

import (
	"fmt"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// GermanFillerWordsRule ports org.languagetool.rules.de.GermanFillerWordsRule
// (extends AbstractStatisticStyleRule; text-level % limit; default off; DEFAULT_MIN_PERCENT=8).
type GermanFillerWordsRule struct {
	*rules.AbstractStatisticStyleRule
	fillers map[string]struct{}
	// TestTwoFollowing / TestManyInSentence port Java UserConfig rule options (default false).
	// Java fields: testTwoFollowing / TestManyInSentence (both read cf[2] — upstream bug).
	TestTwoFollowing   bool
	TestManyInSentence bool
	sentenceMessage    string
}

const (
	germanFillerDefaultMinPercent     = 8
	germanFillerDefaultTwoFollowing   = false
	germanFillerDefaultManyInSentence = false
	germanFillerSentenceMsg1          = "Zwei potentielle Füllwörter hintereinander. Mindestens eins sollte gelöscht werden."
	germanFillerSentenceMsg2          = "Mehr als zwei potentielle Füllwörter in einem Satz. Mindestens eins sollte gelöscht werden."
)

// NewGermanFillerWordsRule ports GermanFillerWordsRule(messages, lang, null).
func NewGermanFillerWordsRule(messages map[string]string) *GermanFillerWordsRule {
	return NewGermanFillerWordsRuleWithUserConfig(messages, nil)
}

// NewGermanFillerWordsRuleWithUserConfig ports
// GermanFillerWordsRule(ResourceBundle messages, Language lang, UserConfig userConfig).
func NewGermanFillerWordsRuleWithUserConfig(messages map[string]string, userConfig *languagetool.UserConfig) *GermanFillerWordsRule {
	fillers := map[string]struct{}{}
	for _, w := range []string{
		"aber",
		"abermals",
		"allein",
		"allemal",
		"allenfalls",
		"allenthalben",
		"allerdings",
		"allesamt",
		"allzu",
		"also",
		"alt",
		"andauernd",
		"andererseits",
		"andernfalls",
		"anscheinend",
		"auch",
		"auffallend",
		"augenscheinlich",
		"ausdrücklich",
		"ausgerechnet",
		"ausnahmslos",
		"außerdem",
		"äußerst",
		"beinahe",
		"bekanntlich",
		"bereits",
		"besonders",
		"bestenfalls",
		"bestimmt",
		"bloß",
		"dabei",
		"dadurch",
		"dafür",
		"dagegen",
		"daher",
		"damals",
		"danach",
		"demgegenüber",
		"demgemäß",
		"demnach",
		"denkbar",
		"denn",
		"dennoch",
		"deshalb",
		"deswegen",
		"doch",
		"durchaus",
		"durchweg",
		"eben",
		"eigentlich",
		"einerseits",
		"einfach",
		"einige",
		"einigermaßen",
		"einmal",
		"ergo",
		"erheblich",
		"etliche",
		"etwa",
		"etwas",
		"fast",
		"folgendermaßen",
		"folglich",
		"förmlich",
		"fortwährend",
		"fraglos",
		"freilich",
		"ganz",
		"gänzlich",
		"gar",
		"gelegentlich",
		"gemeinhin",
		"genau",
		"geradezu",
		"gewiss",
		"gewissermaßen",
		"glatt",
		"gleichsam",
		"gleichwohl",
		"glücklicherweise",
		"gottseidank",
		"größtenteils",
		"häufig",
		"hingegen",
		"hinlänglich",
		"höchst",
		"höchstens",
		"immer",
		"immerhin",
		"immerzu",
		"indessen",
		"infolgedessen",
		"insbesondere",
		"inzwischen",
		"irgend",
		"irgendein",
		"irgendjemand",
		"irgendwann",
		"irgendwie",
		"irgendwo",
		"ja",
		"je",
		"jedenfalls",
		"jedoch",
		"jemals",
		"kaum",
		"keinesfalls",
		"keineswegs",
		"längst",
		"lediglich",
		"leider",
		"letztlich",
		"manchmal",
		"mehrfach",
		"meinetwegen",
		"meist",
		"meistens",
		"meistenteils",
		"mindestens",
		"mithin",
		"mitunter",
		"möglicherweise",
		"möglichst",
		"nämlich",
		"naturgemäß",
		"natürlich",
		"neuerdings",
		"neuerlich",
		"neulich",
		"nichtsdestoweniger",
		"nie",
		"niemals",
		"nun",
		"nur",
		"offenbar",
		"offenkundig",
		"offensichtlich",
		"oft",
		"ohnedies",
		"partout",
		"plötzlich",
		"praktisch",
		"quasi",
		"recht",
		"reichlich",
		"reiflich",
		"relativ",
		"restlos",
		"richtiggehend",
		"rundheraus",
		"rundum",
		"sattsam",
		"schlicht",
		"schlichtweg",
		"schließlich",
		"schlussendlich",
		"schon",
		"sehr",
		"selbst",
		"selbstredend",
		"selbstverständlich",
		"selten",
		"seltsamerweise",
		"sicher",
		"sicherlich",
		"so",
		"sogar",
		"sonst",
		"sowieso",
		"sozusagen",
		"stellenweise",
		"stets",
		"trotzdem",
		"überaus",
		"überdies",
		"überhaupt",
		"übrigens",
		"umständehalber",
		"unbedingt",
		"unerhört",
		"ungefähr",
		"ungemein",
		"ungewöhnlich",
		"ungleich",
		"unglücklicherweise",
		"unlängst",
		"unmaßgeblich",
		"unsagbar",
		"unsäglich",
		"unstreitig",
		"unzweifelhaft",
		"vergleichsweise",
		"vermutlich",
		"vielfach",
		"vielleicht",
		"voll",
		"vollends",
		"völlig",
		"vollkommen",
		"vollständig",
		"wahrscheinlich",
		"weidlich",
		"weitgehend",
		"wenigstens",
		"wieder",
		"wiederum",
		"wirklich",
		"wohl",
		"wohlgemerkt",
		"womöglich",
		"ziemlich",
		"zudem",
		"zugegeben",
		"zumeist",
		"zusehends",
		"zuweilen",
		"zweifellos",
		"zweifelsfrei",
		"zweifelsohne",
	} {
		fillers[w] = struct{}{}
	}
	// Java super: AbstractStatisticStyleRule(messages, lang, userConfig, DEFAULT_MIN_PERCENT)
	// → minPercent from cf[0], excludeDirectSpeech from cf[1] or excludeDirectSpeech().
	minPercent := germanFillerDefaultMinPercent
	excludeDS := true // Java excludeDirectSpeech()
	if userConfig != nil {
		if cf := userConfig.GetConfigValueByID("FILLER_WORDS_DE"); len(cf) > 0 && cf[0] != nil {
			if n, ok := configInt(cf[0]); ok {
				minPercent = n
			}
		}
		if cf := userConfig.GetConfigValueByID("FILLER_WORDS_DE"); len(cf) > 1 && cf[1] != nil {
			if b, ok := cf[1].(bool); ok {
				excludeDS = b
			}
		}
	}
	r := &GermanFillerWordsRule{
		AbstractStatisticStyleRule: &rules.AbstractStatisticStyleRule{
			ID:                  "FILLER_WORDS_DE",
			Description:         "Statistische Stilanalyse: Füllwörter",
			MinPercent:          minPercent,
			Denominator:         100,
			ExcludeDirectSpeech: excludeDS,
		},
		fillers:            fillers,
		TestTwoFollowing:   germanFillerGetTwoFollowing(userConfig),
		TestManyInSentence: germanFillerGetManyInSentence(userConfig),
	}
	r.ConditionFulfilled = r.conditionFulfilled
	r.SentenceConditionFulfilled = r.sentenceConditionFulfilled
	r.LimitMessage = r.getLimitMessage
	rules.InitStatisticStyleMeta(r.AbstractStatisticStyleRule, messages, false)
	if userConfig != nil {
		if ls := userConfig.GetLinguServices(); ls != nil {
			ls.SetThesaurusRelevantRule(r)
		}
	}
	return r
}

// germanFillerGetTwoFollowing ports getTwoFollowing — reads cf[2] Boolean.
func germanFillerGetTwoFollowing(userConfig *languagetool.UserConfig) bool {
	if userConfig != nil {
		cf := userConfig.GetConfigValueByID("FILLER_WORDS_DE")
		if len(cf) > 2 && cf[2] != nil {
			if b, ok := cf[2].(bool); ok {
				return b
			}
		}
	}
	return germanFillerDefaultTwoFollowing
}

// germanFillerGetManyInSentence ports getManyInSentence.
// Java bug-for-bug: also reads cf[2] (not cf[3]) — same slot as two-following.
func germanFillerGetManyInSentence(userConfig *languagetool.UserConfig) bool {
	if userConfig != nil {
		cf := userConfig.GetConfigValueByID("FILLER_WORDS_DE")
		if len(cf) > 2 && cf[2] != nil {
			if b, ok := cf[2].(bool); ok {
				return b
			}
		}
	}
	return germanFillerDefaultManyInSentence
}

func configInt(v any) (int, bool) {
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

// NewGermanFillerWordsRuleWithMinPercent builds with explicit percent (0 = show all; twin tests).
func NewGermanFillerWordsRuleWithMinPercent(messages map[string]string, minPercent int) *GermanFillerWordsRule {
	r := NewGermanFillerWordsRule(messages)
	r.MinPercent = minPercent
	return r
}

// NewGermanFillerWordsRuleWithDefaultLimit is an alias for the Java default (8%).
func NewGermanFillerWordsRuleWithDefaultLimit(messages map[string]string) *GermanFillerWordsRule {
	return NewGermanFillerWordsRule(messages)
}

func (r *GermanFillerWordsRule) GetID() string {
	if r != nil && r.AbstractStatisticStyleRule != nil {
		return r.AbstractStatisticStyleRule.GetID()
	}
	return "FILLER_WORDS_DE"
}

// GetDescription ports GermanFillerWordsRule.getDescription.
func (r *GermanFillerWordsRule) GetDescription() string {
	return "Statistische Stilanalyse: Füllwörter"
}

// GetConfigurePercentText ports getConfigurePercentText.
func (r *GermanFillerWordsRule) GetConfigurePercentText() string {
	return "Anzeigen wenn mehr als ...% eines Kapitels Füllwörter sind:"
}

// GetConfigureWithoutDirectSpeachText ports getConfigureWithoutDirectSpeachText (Java spelling).
func (r *GermanFillerWordsRule) GetConfigureWithoutDirectSpeachText() string {
	return "Keine wörtliche Rede und Zitate berücksichtigen"
}

// GetRuleOptions ports GermanFillerWordsRule.getRuleOptions (overrides abstract base's 2-option form).
func (r *GermanFillerWordsRule) GetRuleOptions() []*rules.RuleOption {
	return []*rules.RuleOption{
		rules.NewRuleOption(germanFillerDefaultMinPercent, r.GetConfigurePercentText(), 0, 100),
		rules.NewRuleOption(true, r.GetConfigureWithoutDirectSpeachText(), nil, nil), // excludeDirectSpeech()
		rules.NewRuleOption(germanFillerDefaultTwoFollowing, "Zeige zwei aufeinander folgende Füllwörter", nil, nil),
		rules.NewRuleOption(germanFillerDefaultManyInSentence, "Zeige mehr als zwei Füllwörter in einem Satz", nil, nil),
	}
}

// GetSentenceMessage ports getSentenceMessage.
func (r *GermanFillerWordsRule) GetSentenceMessage() string {
	if r == nil {
		return ""
	}
	return r.sentenceMessage
}

func (r *GermanFillerWordsRule) getLimitMessage(limit int, percent float64) string {
	// Java GermanFillerWordsRule.getLimitMessage
	if limit == 0 {
		return "Dieses Wort könnte ein Füllwort sein. Möglicherweise ist es besser es zu löschen."
	}
	return fmt.Sprintf("Mehr als %d%% Füllwörter {%d%%} gefunden. Möglicherweise ist es besser dieses potentielle Füllwort zu löschen.",
		limit, int(percent+0.5))
}

func (r *GermanFillerWordsRule) isFillerSurface(tok string) bool {
	if r == nil {
		return false
	}
	_, ok := r.fillers[tok]
	return ok
}

func (r *GermanFillerWordsRule) conditionFulfilled(tokens []*languagetool.AnalyzedTokenReadings, nToken int) int {
	if r == nil || nToken < 0 || nToken >= len(tokens) {
		return -1
	}
	tok := tokens[nToken].GetToken()
	if !r.isFillerSurface(tok) {
		return -1
	}
	if germanFillerIsException(tokens, nToken) {
		return -1
	}
	// Java conditionFulfilled:
	// (nToken < 2 || !isTwoWordException(prev, cur))
	// && (nToken > length-2 || !isTwoWordException(cur, next))
	if nToken >= 2 && germanFillerTwoWordException(tokens[nToken-1].GetToken(), tok) {
		return -1
	}
	if nToken <= len(tokens)-2 && nToken+1 < len(tokens) &&
		germanFillerTwoWordException(tok, tokens[nToken+1].GetToken()) {
		return -1
	}
	return nToken
}

func (r *GermanFillerWordsRule) sentenceConditionFulfilled(tokens []*languagetool.AnalyzedTokenReadings, nToken int) bool {
	if r == nil {
		return false
	}
	// Java sentenceConditionFulfilled for two-following uses fillerWords.contains +
	// !isException only — NOT isTwoWordException / full conditionFulfilled.
	// Java does not clear sentenceMessage first; only set when returning true.
	if r.TestTwoFollowing {
		prevFiller := nToken > 1 && tokens[nToken-1] != nil &&
			r.isFillerSurface(tokens[nToken-1].GetToken()) && !germanFillerIsException(tokens, nToken-1)
		nextFiller := nToken < len(tokens)-1 && tokens[nToken+1] != nil &&
			r.isFillerSurface(tokens[nToken+1].GetToken()) && !germanFillerIsException(tokens, nToken+1)
		if prevFiller || nextFiller {
			r.sentenceMessage = germanFillerSentenceMsg1
			r.SentenceMessage = r.sentenceMessage
			return true
		}
	}
	if r.TestManyInSentence {
		n := 0
		for i := nToken - 2; i > 0; i-- {
			if r.conditionFulfilled(tokens, i) == i {
				n++
				if n > 1 {
					r.sentenceMessage = germanFillerSentenceMsg2
					r.SentenceMessage = r.sentenceMessage
					return true
				}
			}
		}
		for i := nToken + 2; i < len(tokens); i++ {
			if r.conditionFulfilled(tokens, i) == i {
				n++
				if n > 1 {
					r.sentenceMessage = germanFillerSentenceMsg2
					r.SentenceMessage = r.sentenceMessage
					return true
				}
			}
		}
	}
	return false
}

// germanFillerIsException ports GermanFillerWordsRule.isException only
// (two-word pairs stay in conditionFulfilled, matching Java).
// Lemma/POS branches need tagged input; without tags they fail closed (not exception).
func germanFillerIsException(tokens []*languagetool.AnalyzedTokenReadings, num int) bool {
	if num < 0 || num >= len(tokens) {
		return true
	}
	tok := tokens[num].GetToken()
	// Java: if (num == 1 || ",".equals(tokens[num - 1].getToken())) return true;
	if num == 1 || (num > 0 && tokens[num-1].GetToken() == ",") {
		return true
	}
	if tok == "allein" {
		// Java: hasLemma("sein")
		for i := 1; i < len(tokens); i++ {
			if tokens[i] != nil && tokens[i].HasLemma("sein") {
				return true
			}
		}
		return false
	}
	if tok == "recht" {
		// Java: hasAnyLemma("haben", "geben"); no return false after loop (falls through).
		for i := 1; i < len(tokens); i++ {
			if tokens[i] != nil && tokens[i].HasAnyLemma("haben", "geben") {
				return true
			}
		}
	}
	if num < len(tokens)-1 && (tok == "so" || tok == "besonders") &&
		tokens[num+1].HasPosTagStartingWith("ADJ") {
		return true
	}
	if tokens[num].HasPosTagStartingWith("ADJ") && num > 0 && tokens[num-1].GetToken() == "so" {
		return true
	}
	if tok == "nur" && num > 0 && tokens[num-1].GetToken() == "nicht" {
		for i := num + 1; i < len(tokens)-2; i++ {
			if tokens[i].GetToken() == "," &&
				(tokens[i+1].GetToken() == "auch" ||
					(tokens[i+1].GetToken() == "sondern" && tokens[i+2].GetToken() == "auch")) {
				return true
			}
		}
	}
	if num > 2 && tok == "auch" && tokens[num-1].GetToken() == "sondern" && tokens[num-2].GetToken() == "," {
		for i := 1; i < num-2; i++ {
			if tokens[i].GetToken() == "nicht" && tokens[i+1].GetToken() == "nur" {
				return true
			}
		}
	}
	return false
}

func germanFillerTwoWordException(first, second string) bool {
	return (first == "aber" && (second == "nur" || second == "auch")) ||
		(first == "auch" && second == "nur") ||
		(first == "immer" && second == "wieder") ||
		(first == "genau" && second == "so") ||
		(first == "so" && (second == "etwas" || second == "viel" || second == "oft")) ||
		(first == "schon" && second == "fast")
}

// Match runs the text-level statistic path on a single sentence (Java is TextLevelRule).
func (r *GermanFillerWordsRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || sentence == nil {
		return nil
	}
	return r.MatchList([]*languagetool.AnalyzedSentence{sentence})
}

// MatchList ports TextLevelRule.match for filler statistics.
func (r *GermanFillerWordsRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || r.AbstractStatisticStyleRule == nil {
		return nil
	}
	return r.AbstractStatisticStyleRule.MatchList(sentences)
}
