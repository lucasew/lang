package uk

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// TokenAgreementVerbNounExceptionHelper anchors the Java twin; logic is package funcs.
type TokenAgreementVerbNounExceptionHelper struct{}

func NewTokenAgreementVerbNounExceptionHelper() *TokenAgreementVerbNounExceptionHelper {
	return &TokenAgreementVerbNounExceptionHelper{}
}

// Exception reports IsVerbNounException for ATR tokens.
func (h *TokenAgreementVerbNounExceptionHelper) Exception(tokens []*languagetool.AnalyzedTokenReadings, verbPos, nounPos int) bool {
	return IsVerbNounException(tokens, verbPos, nounPos)
}

// partsCantSkip ports PARTS_CANT_SKIP for isExceptionSkip.
var partsCantSkipRE = regexp.MustCompile(
	`^(?:і|й|та|чи|або|але|як|де|куди|наче|ніби|хоч|навіщо|немов|вдвічі|дедалі|щойно|наскільки)$`)

// IsVerbNounHardAdjNoun returns skip count (>=0) or -1 (Java isExceptionHardAdjNoun).
// Used when scanning after a verb; for pair-checker, treat skip>=0 as exception.
func IsVerbNounHardAdjNoun(tokens []*languagetool.AnalyzedTokenReadings, i int, verbPos int) int {
	if tokens == nil || i < 0 || i >= len(tokens) || tokens[i] == nil {
		return -1
	}
	clean := CleanTokenLower(tokens[i])
	if regexp.MustCompile(`^(?:[0-9]{4}-.+|нікому|нічому|нічого|нікого|нічим|решту|ніщо)$`).MatchString(clean) {
		return 1
	}
	if HasLemmaTokenAny(tokens[i], []string{"сам", "самий", "себе", "один"}) {
		return 1
	}
	if i < len(tokens)-1 {
		next := CleanTokenLower(tokens[i+1])
		if HasPosTagRE(tokens[i], regexp.MustCompile(`adj:m:v_rod.*`)) &&
			regexp.MustCompile(`^(?:роду|разу|типу|штибу|розміру)$`).MatchString(next) {
			return 1
		}
		if HasPosTagRE(tokens[i], regexp.MustCompile(`(?:adj|numr):[mp]:v_oru.*`)) &&
			regexp.MustCompile(`^(?:чином|способом|робом|ходом|шляхом|коштом)$`).MatchString(next) {
			return 1
		}
		if verbPos >= 0 && verbPos < len(tokens) && HasPosTagStart(tokens[verbPos], "advp") &&
			strings.EqualFold(tokens[i].GetCleanToken(), "тим") &&
			strings.EqualFold(tokens[i+1].GetCleanToken(), "самим") {
			return 1
		}
		if HasPosTagRE(tokens[i], regexp.MustCompile(`adj:f:v_oru.*`)) && next == "мірою" {
			return 1
		}
		if HasPosTagRE(tokens[i], regexp.MustCompile(`adj:f:v_rod.*`)) &&
			regexp.MustCompile(`^(?:якості|свіжості)$`).MatchString(next) {
			return 1
		}
		if next == "темпами" {
			return 1
		}
	}

	// fixed multi-token phrases: Java mNow == i+len-1 style returns skip length
	phrases := []struct {
		line string
		skip int
	}{
		{"не те щоб", 3},
		{"не те що", 3},
		{"не останньою чергою", 3},
		{"не те , що", 4},
		{"світ за очі", 3},
		{"ні світ ні", 3},
		{"куди очі", 3},
		{"станом на", 3},
		{"страх як", 3},
		{"жах як", 3},
	}
	for _, p := range phrases {
		if NewSearchMatch(p.line).MNowATR(tokens, i) >= 0 {
			return p.skip
		}
	}

	if i > 0 && tokens[i-1] != nil && tokens[i-1].GetCleanToken() == "не" &&
		regexp.MustCompile(`^(?:указ|варіант|рідкість)$`).MatchString(clean) {
		return 0
	}
	return -1
}

// IsVerbNounExceptionSkip returns skip count or -1 (Java isExceptionSkip).
func IsVerbNounExceptionSkip(tokens []*languagetool.AnalyzedTokenReadings, i int) int {
	if tokens == nil || i < 0 || i >= len(tokens) || tokens[i] == nil {
		return -1
	}
	clean := CleanTokenLower(tokens[i])
	if hasPosTagAllPartAdv(tokens[i]) &&
		!AdvQuantPattern.MatchString(clean) &&
		!partsCantSkipRE.MatchString(clean) {
		return 0
	}
	if HasPosTagRE(tokens[i], regexp.MustCompile(`^part`)) &&
		hasPosTagAllPartConjAdv(tokens[i]) &&
		!partsCantSkipRE.MatchString(clean) {
		return 0
	}
	return -1
}

func hasPosTagAllPartAdv(tok *languagetool.AnalyzedTokenReadings) bool {
	tags := CollectPOSTags(tok)
	if len(tags) == 0 {
		return false
	}
	for _, p := range tags {
		if !strings.HasPrefix(p, "part") && !strings.HasPrefix(p, "adv") {
			return false
		}
	}
	return true
}

func hasPosTagAllPartConjAdv(tok *languagetool.AnalyzedTokenReadings) bool {
	tags := CollectPOSTags(tok)
	if len(tags) == 0 {
		return false
	}
	for _, p := range tags {
		if !strings.HasPrefix(p, "part") && !strings.HasPrefix(p, "conj") && !strings.HasPrefix(p, "adv") {
			return false
		}
	}
	return true
}

// IsExceptionVerb reports verb-side soft exception (Java isExceptionVerb Type.exception).
func IsExceptionVerb(tokens []*languagetool.AnalyzedTokenReadings, i int) bool {
	if tokens == nil || i < 0 || i >= len(tokens) || tokens[i] == nil {
		return false
	}
	if HasLemmaToken(tokens[i], "мусити") {
		return true
	}
	clean := CleanTokenLower(tokens[i])
	if clean == "може" {
		return true
	}
	// як є / як могти
	if i > 1 && (clean == "є" || HasLemmaToken(tokens[i], "могти")) &&
		strings.EqualFold(tokens[i-1].GetCleanToken(), "як") {
		return true
	}
	// будь то
	if i < len(tokens)-2 && clean == "будь" &&
		strings.EqualFold(tokens[i+1].GetCleanToken(), "то") {
		return true
	}
	return false
}

// IsExceptionVerbSkip reports verb-side skip patterns (спати after класти, pluperfect був).
func IsExceptionVerbSkip(tokens []*languagetool.AnalyzedTokenReadings, i int) bool {
	if tokens == nil || i < 1 || i >= len(tokens) || tokens[i] == nil {
		return false
	}
	clean := CleanTokenLower(tokens[i])
	// вкласти спати
	if i < len(tokens)-1 && clean == "спати" &&
		HasLemmaTokenRE(tokens[i-1], regexp.MustCompile(`^(?:по|в)?кла(?:сти|вши)$`)) {
		return true
	}
	// розпочав був / pluperfect
	if regexp.MustCompile(`^(?:був|було)$`).MatchString(clean) &&
		HasPosTagRE(tokens[i-1], regexp.MustCompile(`verb.*:past:m.*`)) {
		return true
	}
	if regexp.MustCompile(`^(?:були|було)$`).MatchString(clean) &&
		HasPosTagRE(tokens[i-1], regexp.MustCompile(`verb.*:past:p.*`)) {
		return true
	}
	if clean == "було" && HasPosTagRE(tokens[i-1], regexp.MustCompile(`verb.*:past:n.*`)) {
		return true
	}
	if regexp.MustCompile(`^(?:була|було)$`).MatchString(clean) &&
		HasPosTagRE(tokens[i-1], regexp.MustCompile(`verb.*:past:f.*`)) {
		return true
	}
	// чути/проголошено було
	if regexp.MustCompile(`^(?:було|буде)$`).MatchString(clean) &&
		HasPosTagRE(tokens[i-1], regexp.MustCompile(`verb.*(?:impers|predic).*`)) {
		return true
	}
	return false
}
