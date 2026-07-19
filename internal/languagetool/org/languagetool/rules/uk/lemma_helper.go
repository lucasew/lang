package uk

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// LemmaHelper ports lemma sets and membership helpers from
// org.languagetool.rules.uk.LemmaHelper (token-array search deferred).

const ignoreChars = "\u00AD\u0301"

// CityAvenu is Latin-script street/city tokens treated as foreign toponyms.
var CityAvenu = setOf(
	"сіті", "ситі", "стріт", "стрит", "рівер", "ривер", "авеню",
	"штрасе", "штрассе", "сьоркл", "сквер", "плац",
)

// MonthLemmas are Ukrainian month lemmas.
var MonthLemmas = []string{
	"січень", "лютий", "березень", "квітень", "травень", "червень", "липень",
	"серпень", "вересень", "жовтень", "листопад", "грудень",
}

// DaysOfWeek are Ukrainian weekday lemmas.
var DaysOfWeek = []string{
	"понеділок", "вівторок", "середа", "четвер", "п'ятниця", "субота", "неділя",
}

// TimeLemmas are time-unit lemmas used by agreement rules.
var TimeLemmas = []string{
	"секунда", "хвилина", "хвилинка", "хвилина-дві", "хвилинка-друга",
	"година", "годинка", "півгодини", "година-друга", "година-дві",
	"час", "день", "день-другий", "півдня", "ніч", "ніченька", "вечір", "ранок",
	"тиждень", "тиждень-два", "тиждень-другий",
	"місяць", "місяць-два", "місяць-другий", "місяць-півтора", "доба", "мить", "хвилька",
	"рік", "рік-два", "рік-півтора", "півроку", "півроку-рік", "десятиліття", "десятиріччя",
	"століття", "півстоліття", "сторіччя", "півсторіччя", "тисячоліття", "півтисячоліття",
	"квартал", "годочок",
	"літо", "зима", "весна", "осінь",
	"тайм", "мить", "період", "термін", "сезон", "декада", "каденція", "раунд", "сезон",
}

// DistanceLemmas are measurement unit lemmas.
var DistanceLemmas = []string{
	"міліметр", "сантиметр", "метр", "кілометр", "кілограм", "кілограм–півтора",
	"гектар", "миля", "аршин", "дециметр", "верства", "верста",
	"грам", "літр", "фунт", "тонна", "центнер",
}

// PseudoNumLemmas are group/quantity nouns.
var PseudoNumLemmas = []string{
	"десяток", "десяток-другий", "сотня", "сотка", "тисяча", "п'ятірка", "пара",
	"третина", "чверть", "половина", "дюжина", "жменя", "жменька", "купа", "купка",
	"парочка", "оберемок", "безліч",
}

// MoneyLemmas are currency lemmas.
var MoneyLemmas = []string{"гривня", "копійка"}

// TimeLemmasShort is a short time-unit list.
var TimeLemmasShort = []string{"секунда", "хвилина", "година", "рік"}

// PlusMinus are quantitative plus/minus words.
var PlusMinus = setOf("плюс", "мінус", "максимум", "мінімум")

// AdvQuantPattern matches adverbial quantifiers.
var AdvQuantPattern = regexp.MustCompile(
	`^(більше|менше|чимало|багато|мало|забагато|замало|немало|багатенько|чималенько|стільки|обмаль|вдосталь|удосталь|трохи|трошки|досить|достатньо|недостатньо|предостатньо|багацько|чимбільше|побільше|порівну|більшість|трішки|предосить|повно|повнісінько|мільйон|тисяча|сотня|мільярд|трильйон|десяток|нуль|безліч|кілька|декілька|пара|парочка|купа|купка|безліч|мінімум|максимум)$`,
)

// PartInsertPattern matches parenthetical insert particles.
var PartInsertPattern = regexp.MustCompile(
	`^(бодай|буцім(то)?|геть|дедалі|десь|іще|ледве|мов(би(то)?)?|навіть|наче(б(то)?)?|неначе(бто)?|немов(би(то)?)?|ніби(то)?|попросту|просто(-напросто)?|справді|усього-на-всього|хай|хоча?|якраз|ж|би?|власне)$`,
)

// DashesPattern / QuotesPattern match dash and quotation punctuation.
var (
	DashesPattern = regexp.MustCompile(`^[\x{2010}-\x{2015}-]$`)
	QuotesPattern = regexp.MustCompile(`^[\p{Pi}\p{Pf}]$`)
)

// TimePlusLemmas is the union of time/distance/week/month/pseudo-num/money lemmas.
var TimePlusLemmas map[string]struct{}

// TimePlusLemmasPattern matches any TimePlus lemma.
var TimePlusLemmasPattern *regexp.Regexp

func init() {
	TimePlusLemmas = map[string]struct{}{}
	addAll(TimePlusLemmas, TimeLemmas)
	addAll(TimePlusLemmas, DistanceLemmas)
	addAll(TimePlusLemmas, DaysOfWeek)
	addAll(TimePlusLemmas, MonthLemmas)
	addAll(TimePlusLemmas, PseudoNumLemmas)
	addAll(TimePlusLemmas, MoneyLemmas)
	for _, s := range []string{"вихідний", "уїк-енд", "уїкенд", "вікенд", "відсоток", "раз", "крок"} {
		TimePlusLemmas[s] = struct{}{}
	}
	parts := make([]string, 0, len(TimePlusLemmas))
	for s := range TimePlusLemmas {
		parts = append(parts, regexp.QuoteMeta(s))
	}
	TimePlusLemmasPattern = regexp.MustCompile("^(?:" + strings.Join(parts, "|") + ")$")
}

// HasLemma reports whether any lemma is in the collection.
func HasLemma(lemmas []string, want map[string]struct{}) bool {
	for _, l := range lemmas {
		if _, ok := want[l]; ok {
			return true
		}
	}
	return false
}

// HasLemmaInList is HasLemma for a slice of wanted lemmas.
func HasLemmaInList(lemmas, want []string) bool {
	set := setOf(want...)
	return HasLemma(lemmas, set)
}

// HasLemmaString reports whether any reading lemma equals want.
func HasLemmaString(lemmas []string, want string) bool {
	for _, l := range lemmas {
		if l == want {
			return true
		}
	}
	return false
}

// HasLemmaWithPartPos ports LemmaHelper.hasLemma(readings, lemmas, partPos):
// lemma equals one of lemmas AND POS tag contains partPos (Java String.contains).
func HasLemmaWithPartPos(tok *languagetool.AnalyzedTokenReadings, lemmas []string, partPos string) bool {
	if tok == nil || partPos == "" || len(lemmas) == 0 {
		return false
	}
	want := setOf(lemmas...)
	for _, r := range tok.GetReadings() {
		if r == nil || r.GetLemma() == nil || r.GetPOSTag() == nil {
			continue
		}
		if _, ok := want[*r.GetLemma()]; !ok {
			continue
		}
		if strings.Contains(*r.GetPOSTag(), partPos) {
			return true
		}
	}
	return false
}

// CleanIgnoreChars strips soft hyphen / combining acute from a token.
func CleanIgnoreChars(token string) string {
	return strings.Map(func(r rune) rune {
		if strings.ContainsRune(ignoreChars, r) {
			return -1
		}
		return r
	}, token)
}

// IsTimePlusLemma reports membership in TimePlusLemmas.
func IsTimePlusLemma(lemma string) bool {
	_, ok := TimePlusLemmas[lemma]
	return ok
}

func setOf(ss ...string) map[string]struct{} {
	m := make(map[string]struct{}, len(ss))
	for _, s := range ss {
		m[s] = struct{}{}
	}
	return m
}

func addAll(dst map[string]struct{}, src []string) {
	for _, s := range src {
		dst[s] = struct{}{}
	}
}

// IsCapitalized ports LemmaHelper.isCapitalized (Ukrainian title-case heuristics).
func IsCapitalized(word string) bool {
	if word == "" {
		return false
	}
	runes := []rune(word)
	if len(runes) < 2 {
		return false
	}
	char0 := runes[0]
	if !unicode.IsUpper(char0) {
		return false
	}
	// lax on Latin: EuroGas
	if char0 >= 'A' && char0 <= 'Z' && unicode.IsLower(runes[1]) {
		return true
	}
	prevDash := false
	sz := len(runes)
	for i := 1; i < sz; i++ {
		ch := runes[i]
		if strings.ContainsRune(ignoreChars, ch) {
			continue
		}
		dash := ch == '-' || ch == '\u2013'
		if dash {
			if i == sz-2 && unicode.IsDigit(runes[i+1]) {
				return true
			}
			prevDash = true
			continue
		}
		if ch != '\'' && ch != '\u0301' && ch != '\u00AD' {
			// prevDash != Character.isUpperCase(ch)
			if prevDash != unicode.IsUpper(ch) {
				return false
			}
		}
		prevDash = false
	}
	return true
}
