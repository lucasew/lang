package en

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// EnglishWordRepeatRule ports org.languagetool.rules.en.EnglishWordRepeatRule.
// POS branches use HasPartialPosTag (Java posIsIn); without tags those arms
// fail closed (no surface invent of VB/NN/PRP/…).
type EnglishWordRepeatRule struct {
	*rules.WordRepeatRule
}

var singleChar = regexp.MustCompile(`(?i)^[a-z]$`)

var apostropheRE = regexp.MustCompile(`['’` + "`" + `´‘]`)
var logSignRE = regexp.MustCompile(`log(ged|s)?|sign(ed|s)?`)

func NewEnglishWordRepeatRule(messages map[string]string) *EnglishWordRepeatRule {
	base := rules.NewWordRepeatRule(messages)
	base.IDOverride = "ENGLISH_WORD_REPEAT_RULE"
	r := &EnglishWordRepeatRule{WordRepeatRule: base}
	base.ExtraIgnore = r.englishIgnore
	return r
}

// englishIgnore ports EnglishWordRepeatRule.ignore.
// Base WordRepeatRule.Ignore then applies super.ignore (Phi/Li/…).
func (r *EnglishWordRepeatRule) englishIgnore(tokens []*languagetool.AnalyzedTokenReadings, position int) bool {
	if position == 0 {
		return false
	}
	// TODO (Java): What that is is a … / but you you're … / I'm so so happy / I'm very very happy
	word := tokens[position].GetToken()

	if (repetitionOf("did", tokens, position) || repetitionOf("do", tokens, position) ||
		repetitionOf("does", tokens, position)) && position+1 < len(tokens) &&
		strings.EqualFold(tokens[position+1].GetToken(), "n't") {
		return true
	}
	// "Please pass her her phone."
	if repetitionOf("her", tokens, position) &&
		posIsIn(tokens, position-2, "VB", "VBP", "VBZ", "VBG", "VBD", "VBN") &&
		posIsIn(tokens, position+1, "NN", "NNS", "NN:U", "NN:UN", "NNP") {
		return true
	}
	// "If I had had time…"
	if repetitionOf("had", tokens, position) && posIsIn(tokens, position-2, "PRP", "NN") {
		return true
	}
	// "I don't think that that is a problem."
	if repetitionOf("that", tokens, position) && posIsIn(tokens, position+1, "MD", "NN", "PRP$", "JJ", "VBZ", "VBD") {
		return true
	}
	// "The can can hold the water." — first "can" is NN
	if repetitionOf("can", tokens, position) && posIsIn(tokens, position-1, "NN") {
		return true
	}
	if repetitionOf("hip", tokens, position) && position+1 < len(tokens) &&
		strings.EqualFold(tokens[position+1].GetToken(), "hooray") {
		return true
	}
	if repetitionOf("bam", tokens, position) && position+1 < len(tokens) &&
		strings.EqualFold(tokens[position+1].GetToken(), "bigelow") {
		return true
	}
	if repetitionOf("wild", tokens, position) && position+1 < len(tokens) &&
		strings.EqualFold(tokens[position+1].GetToken(), "west") {
		return true
	}
	if repetitionOf("far", tokens, position) && position+1 < len(tokens) &&
		strings.EqualFold(tokens[position+1].GetToken(), "away") {
		return true
	}
	if repetitionOf("so", tokens, position) && position+1 < len(tokens) &&
		strings.EqualFold(tokens[position+1].GetToken(), "much") {
		return true
	}
	if repetitionOf("so", tokens, position) && position+1 < len(tokens) &&
		strings.EqualFold(tokens[position+1].GetToken(), "many") {
		return true
	}
	// It's S.T.E.A.M.
	if repetitionOf("s", tokens, position) && position > 1 &&
		apostropheRE.MatchString(tokens[position-2].GetToken()) {
		return true
	}
	if repetitionOf("in", tokens, position) && position > 2 &&
		logSignRE.MatchString(tokens[position-3].GetToken()) {
		return true
	}
	if repetitionOf("in", tokens, position) && position > 1 &&
		logSignRE.MatchString(tokens[position-2].GetToken()) {
		return true
	}
	if repetitionOf("a", tokens, position) && position > 1 && tokens[position-2].GetToken() == "." {
		return true
	}
	if repetitionOf("on", tokens, position) && position > 1 && tokens[position-2].GetToken() == "." {
		return true
	}
	// three-time word repetition
	if strings.EqualFold(tokens[position-1].GetToken(), word) {
		if (position+1 < len(tokens) && strings.EqualFold(tokens[position+1].GetToken(), word)) ||
			(position > 1 && strings.EqualFold(tokens[position-2].GetToken(), word)) {
			return true
		}
	}
	// spelling with spaces: "b a s i c a l l y"
	if singleChar.MatchString(tokens[position].GetToken()) && position > 1 &&
		singleChar.MatchString(tokens[position-2].GetToken()) &&
		position+1 < len(tokens) && singleChar.MatchString(tokens[position+1].GetToken()) {
		return true
	}

	// Fixed-phrase / onomatopoeia (Java EnglishWordRepeatRule chain order).
	for _, w := range []string{
		"aye", "blah", "mau", "uh", "paw", "cha", "yum", "wop", "woop", "fnarr", "fnar",
		"ha", "omg", "boo", "tick", "twinkle", "ta", "la", "x", "hi", "ho", "heh", "jay",
		"walla", "sri", "hey", "hah", "oh", "ouh", "chop", "ring", "beep", "bleep", "yeah",
		"gout",
	} {
		if repetitionOf(w, tokens, position) {
			return true
		}
	}
	// Java: wait wait at position==2 only ("Wait wait!" sentence start; not "Please wait wait").
	if repetitionOf("wait", tokens, position) && position == 2 {
		return true
	}
	for _, w := range []string{
		"quack", "meow", "squawk", "whoa", "si", "honk", "brum", "chi", "santorio",
		"lapu", "chow", "shh", "yummy", "boom", "bye", "ah", "aah", "bang", "woof", "wink",
		"yes", "tsk", "hush", "ding", "choo", "miu", "tuk", "yadda", "doo", "sapiens", "tse",
		"no", "Bora",
	} {
		if repetitionOf(w, tokens, position) {
			return true
		}
	}

	// Java: getToken().endsWith("ay") / endsWith("ill") — case-sensitive suffix.
	if strings.HasSuffix(word, "ay") {
		if tokens[position-1].GetToken() == "may" && tokens[position].GetToken() == "May" {
			return true
		}
		if tokens[position-1].GetToken() == "May" && tokens[position].GetToken() == "may" {
			return true
		}
		if len(tokens) > 2 && tokens[1].GetToken() == "May" && tokens[2].GetToken() == "May" {
			return true
		}
	}
	if strings.HasSuffix(word, "ill") {
		if position > 0 && tokens[position-1].GetToken() == "will" && tokens[position].GetToken() == "Will" {
			return true
		}
		if tokens[position-1].GetToken() == "Will" && tokens[position].GetToken() == "will" {
			return true
		}
		if len(tokens) > 2 && tokens[1].GetToken() == "Will" && tokens[2].GetToken() == "Will" {
			return true
		}
	}
	return false
}

// posIsIn ports EnglishWordRepeatRule.posIsIn (HasPartialPosTag substring match).
func posIsIn(tokens []*languagetool.AnalyzedTokenReadings, position int, posTags ...string) bool {
	if position < 0 || position >= len(tokens) || tokens[position] == nil {
		return false
	}
	for _, tag := range posTags {
		if tokens[position].HasPartialPosTag(tag) {
			return true
		}
	}
	return false
}

func repetitionOf(word string, tokens []*languagetool.AnalyzedTokenReadings, position int) bool {
	return position > 0 &&
		strings.EqualFold(tokens[position-1].GetToken(), word) &&
		strings.EqualFold(tokens[position].GetToken(), word)
}
