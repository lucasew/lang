package uk

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// UkrainianWordTokenizer ports org.languagetool.tokenizers.uk.UkrainianWordTokenizer.
type UkrainianWordTokenizer struct{}

func NewUkrainianWordTokenizer() *UkrainianWordTokenizer { return &UkrainianWordTokenizer{} }

const (
	decimalCommaSubst     = '\uE001'
	nonBreakingSpaceSubst = '\uE002'
	nonBreakingDotSubst   = '\uE003'
	nonBreakingColonSubst = '\uE004'
	leftBraceSubst        = '\uE005'
	rightBraceSubst       = '\uE006'
	nonBreakingSlashSubst = '\uE007'
	leftAngleSubst        = '\uE008'
	rightAngleSubst       = '\uE009'
	slashSubst            = '\uE010'
	nonBreakingPlaceholder  = "\uE109"
	breakingPlaceholder     = "\uE110"
	nonBreakingPlaceholder2 = "\uE120"
	softHyphenWrap          = "\u00AD\n"
	softHyphenWrapSubst     = "\uE103"
	urlStartReplaceChar     = 0xE300
)

var (
	weirdApostroph = regexp.MustCompile(`(?i)([бвджзклмнпрстфхш])(["\x{201D}\x{201F}` + "`" + `´])([єїюя])`)
	wordsWithBrackets = regexp.MustCompile(`(?i)([а-яіїєґ])\[([а-яіїєґ]+)\]`)
	decimalComma      = regexp.MustCompile(`([\d]),([\d])`)
	// spaced thousands without lookbehind: handle via ReplaceAllStringFunc
	decimalSpace = regexp.MustCompile(`(?:^|[\s\x{00A0}\x{202F}(])(\d{1,3}(?:[\s\x{00A0}\x{202F}]\d{3})+)(?:[\s\x{00A0}\x{202F}(]|$)`)
	// better: find numbers with spaces inside
	decimalSpaceInner = regexp.MustCompile(`\d{1,3}(?:[\s\x{00A0}\x{202F}]\d{3})+`)

	dashNumbers   = regexp.MustCompile(`([IVXІХ]+)([\x{2013}-])([IVXІХ]+)`)
	nDashSpace    = regexp.MustCompile(`(?i)([а-яіїєґa-z0-9])(\x{2013}[\s\x{00A0}\x{202F}])`)
	nDashSpace2   = regexp.MustCompile(`(?i)([\s.,;!?]\x{2013})([а-яіїєґa-z])`)
	dottedNumbers = regexp.MustCompile(`([\d])\.([\d])`)
	dottedNumbers3 = regexp.MustCompile(`([\d])\.([\d]+)\.([\d])`)
	colonNumbers  = regexp.MustCompile(`([\d]):([\d])`)
	braceInWord   = regexp.MustCompile(`(?i)([а-яіїєґ])\(([а-яіїєґ']+)\)`)
	xmlTag        = regexp.MustCompile(`(?i)<(/?[a-z_]+/?)>`)

	initialsSP2 = regexp.MustCompile(`([А-ЯІЇЄҐ])\.([\s\x{00A0}\x{202F}]{0,5}[А-ЯІЇЄҐ])\.([\s\x{00A0}\x{202F}]{0,5}[А-ЯІЇЄҐ][а-яіїєґ']+)`)
	initialsSP1 = regexp.MustCompile(`([А-ЯІЇЄҐ])\.([\s\x{00A0}\x{202F}]{0,5}[А-ЯІЇЄҐ][а-яіїєґ']+)`)
	initialsRSP2 = regexp.MustCompile(`([А-ЯІЇЄҐ][а-яіїєґ']+)([\s\x{00A0}\x{202F}]?[А-ЯІЇЄҐ])\.([\s\x{00A0}\x{202F}]?[А-ЯІЇЄҐ])\.`)
	initialsRSP1 = regexp.MustCompile(`([А-ЯІЇЄҐ][а-яіїєґ']+)([\s\x{00A0}\x{202F}]?[А-ЯІЇЄҐ])\.`)

	abbrDotVO1 = regexp.MustCompile(`([вВу])\.([\s\x{00A0}\x{202F}]*о)\.`)
	abbrDotVO2 = regexp.MustCompile(`(к)\.([\s\x{00A0}\x{202F}]*с)\.`)
	abbrDotVO3 = regexp.MustCompile(`(ч|ст)\.([\s\x{00A0}\x{202F}]*л)\.`)
	abbrDotTys1 = regexp.MustCompile(`([0-9IІ][\s\x{00A0}\x{202F}]+)(тис|арт)\.`)
	abbrDotTys2 = regexp.MustCompile(`(тис|арт)\.([\s\x{00A0}\x{202F}]+[а-яіїєґ0-9])`)
	abbrDotArt = regexp.MustCompile(`([Аа]рт|[Мм]ал|[Рр]ис|[Сс]пр)\.([\s\x{00A0}\x{202F}]*(?:№[\s\x{00A0}\x{202F}]*)?[0-9])`)
	abbrDotMan = regexp.MustCompile(`(Ман)\.([\s\x{00A0}\x{202F}]*(?:Сіті|[Юю]н))`)
	// lat. without lookbehind: use non-letter prefix
	abbrDotLat = regexp.MustCompile(`(?i)(^|[^а-яіїєґ'\x{0301}-])(лат)\.([\s\x{00A0}\x{202F}]+[a-zA-Z])`)
	abbrDotProf = regexp.MustCompile(`(?i)(^|[^а-яіїєґ'\x{0301}-])([Аа]кад|[Пп]роф|[Дд]оц|[Аа]сист|[Аа]рх|ап|тов|вул|бул|бульв|о|р|ім|упорядн?|др|[Пп]реп|Ів|Дж|Ол|[сС]вт|Авг)\.([\s\x{00A0}\x{202F}]+[А-ЯІЇЄҐа-яіїєґ])`)
	abbrDotGub = regexp.MustCompile(`(.[А-ЯІЇЄҐ][а-яіїєґ'-]+[\s\x{00A0}\x{202F}]+губ)\.`)
	// Go \b is ASCII-only; use explicit non-letter left edge for Cyrillic initials like К.-Святошинський
	abbrDotDash = regexp.MustCompile(`(^|[^а-яіїєґА-ЯІЇЄҐ'])([А-ЯІЇЄҐ]ж?)\.([-\x{2013}](?:[А-ЯІЇЄҐ][а-яіїєґ']{2}|[А-ЯІЇЄҐ]\.))`)
	abbrDotKub = regexp.MustCompile(`(кв|куб)\.([\s\x{00A0}\x{202F}]*(?:[смкд]|мк)?м)`)
	abbrDotSG = regexp.MustCompile(`(с)\.(-г)\.`)
	abbrDotChl = regexp.MustCompile(`(чл)\.(-кор)\.`)
	abbrDotPn = regexp.MustCompile(`(пн|пд)\.(-(зах|сх))\.`)
	invalidMln = regexp.MustCompile(`(млн|млрд)\.( [а-яіїєґ])`)
	// Java ABBR_DOT_2_SMALL_LETTERS_PATTERN: second group has (?![смкд]?м\.) — RE2 has no
	// lookahead; applied via replaceAbbrDot2Small.
	abbrDot2Small = regexp.MustCompile(`(^|[^а-яіїєґА-ЯІЇЄҐ'\x{0301}-])([векнпрстцч]{1,2})\.([\s\x{00A0}\x{202F}]*[екмнпрстч]{1,2})\.`)
	// meter-like second abbr that Java excludes via (?![смкд]?м\.)
	abbrDot2SmallMeterSecond = regexp.MustCompile(`^[\s\x{00A0}\x{202F}]*(?:[смкд]?м|мк)$`)

	// non-ending abbreviations (long list); bare "в." handled carefully (not "в...")
	// Java: (?!\uE120|\.+[\h\v]*$) after the dot — applied via replaceAbbrNonEnding.
	abbrNonEndingList = `абз|австрал|ам|амер|англ|акад(?:ем)?|арк|ауд|біол|бл(?:изьк)?|болг|буд|вип|вірм|грец(?:ьк)?|держ|див|дир|діал|дод|дол|досл|доц|доп|екон|ел|жін|зав|заст|зах|зб|зв|зневажл?|зовн|іл|ім|івр|інж|ісп|іст|італ|к|каб|каф|канд|кв|[1-9]-кімн|кімн|кін|кл|кн|коеф|крим|латин|мал|моб|н|[Нн]апр|нач|нім|нац|нпр|образн|оз|оп|оф|п|пен|перекл|перен|пл|пол|пом|пор|порівн|[Пп]оч|пп|прибл|прикм|прим|присл|пров|пром|просп|[Рр]ед|[Рр]еж|розд|розм|рос|рт|рум|с|санскр|[Сс]вв?|скор|соц|співавт|[сС]т|стор|суч|сх|табл|тт|[тТ]ел|техн|укр|філол|фр|франц|худ|[цЦ]ит|ч|чайн|част|ц|яп|япон`
	abbrNonEnding = regexp.MustCompile(`(?i)(^|[^а-яіїєґ'\x{0301}-])(` + abbrNonEndingList + `)\.`)
	// single-letter в. abbreviation when not ellipsis / not already protected
	abbrDotV = regexp.MustCompile(`(^|[^а-яіїєґА-ЯІЇЄҐ'\x{0301}-])([вВ])\.([^\.\x{E120}])`)
	abbrNonEnding2 = regexp.MustCompile(`([^а-яіїєґА-ЯІЇЄҐ'-]м\.)([\s\x{00A0}\x{202F}]*[А-ЯІЇЄҐ])`)

	abbrNar1 = regexp.MustCompile(`(([0-9]|рік|[рp]\.|[-–—])[\s\x{00A0}\x{202F}]+нар)\.`)
	abbrNar2 = regexp.MustCompile(`(^|[^а-яіїєґА-ЯІЇЄҐ'])(нар)\.([\s\x{00A0}\x{202F}]+[0-9а-яіїєґ])`)

	// ending abbreviations: Java includes р|РР|ст (single) in addition to рр|стст|...
	// Java: ([^letter-](abbr))\. (?!\uE120) — left boundary required so "пародист." is not "ст."
	// Applied via replaceAbbrEnding.
	abbrEnding = regexp.MustCompile(`(?i)(^|[^а-яіїєґА-ЯІЇЄҐ'\x{0301}-])((?:та|й|і) (?:інш?|под)|атм|відс|гр|коп|дес|дол|обл|пов|рр|РР|р|руб|стст|ст|стол|стор|чол|шт)\.`)
	abbrITP = regexp.MustCompile(`([ій][\s\x{00A0}\x{202F}]+т\.)([\s\x{00A0}\x{202F}]*(д|п|ін)\.)`)
	abbrITCH = regexp.MustCompile(`([ву][\s\x{00A0}\x{202F}]+т\.)([\s\x{00A0}\x{202F}]*ч\.)`)
	abbrTZV = regexp.MustCompile(`([\s\x{00A0}\x{202F}(]+т\.)([\s\x{00A0}\x{202F}]*зв\.)`)
	abbrAtEnd = regexp.MustCompile(`(^|[^а-яіїєґА-ЯІЇЄҐ'])(тис|губ|[А-ЯІЇЄҐ])\.[\s\x{00A0}\x{202F}]*$`)
	abbrRedAvt = regexp.MustCompile(`([\s\x{00A0}\x{202F}]+([Рр]ед|[Аа]вт))\.([\s\x{00A0}\x{202F}]*[)\]а-яіїєґ])`)

	// Year with р.
	yearWithR = regexp.MustCompile(`((?:[12][0-9]{3}[—–-])?[12][0-9]{3})(рр?\.)`)

	compoundQuotes1 = regexp.MustCompile(`(?i)([а-яіїє]-)([«"„])([а-яіїєґ'-]+)([»"“])`)
	compoundQuotes2 = regexp.MustCompile(`(?i)([«"„])([а-яіїєґ0-9'-]+)([»"“])(-[а-яіїє])`)

	urlPattern = regexp.MustCompile(`(?i)((https?|ftp)://|www\.)[^\s\x{00A0}\x{202F}/$.?#),]+\.[^\s\x{00A0}\x{202F}),">]*|(mailto:)?[\p{L}\d._-]+@[\p{L}\d_-]+(\.[\p{L}\d_-]+)+`)

	leadingDash = regexp.MustCompile(`^([\x{2014}\x{2013}])([а-яіїєґА-ЯІЇЄҐA-Z])`)
	leadingDash2 = regexp.MustCompile(`^(-)([А-ЯІЇЄҐA-Z])`)
	numberMissingSpace = regexp.MustCompile(`((?:[\s\x{00A0}\x{202F}\x{E110}]|^)[а-яїієґА-ЯІЇЄҐ'-]*[а-яїієґ']?[а-яїієґ])([0-9]+(?:$|[^а-яіїєґА-ЯІЇЄҐa-zA-Z»"“]))`)
	// Java WEB_ENTITIES: CASE_INSENSITIVE|UNICODE_CHARACTER_CLASS + \b after TLD.
	// Go \b is ASCII-only — boundary checked in replaceWebEntities (Unicode letter/digit/_).
	// Enumerate Cyrillic case variants; Latin listed in common cases (+ ToLower check).
	webEntities = regexp.MustCompile(`([а-яіїєґА-ЯІЇЄҐ])\.([Nn][Ee][Tt]|[Ii][Nn][Ff][Oo]|[Cc][Ii][Tt][Yy]|[Ll][Ii][Ff][Ee]|[Uu][Aa]|[Mm][Ee][Dd][Ii][Aa]|[Cc][Oo][Mm]|[Rr][Uu]|НЕТ|Нет|нет|Інфо|інфо|юа|лі|фм|ру|Ру|РУ|орг)`)
	webEntities2 = regexp.MustCompile(`(?i)\.([a-z_-]+)\.(ua)`)

	// colloquial forms ending with ' that must stay attached
	colloquialApos = map[string]bool{
		"мо": true, "тре": true, "тра": true, "чо": true, "нічо": true,
		"бо": true, "зара": true, "пра": true,
	}
)

func (w *UkrainianWordTokenizer) Tokenize(text string) []string {
	urls := map[string]string{}
	// Java: if (!text.trim().isEmpty()) adjustTextForTokenizing — String.trim.
	if !tools.JavaStringTrimIsEmpty(text) {
		text = adjustTextForTokenizing(text, urls)
	}

	var tokenList []string
	for _, token := range splitWithDelimiters(text) {
		if token == breakingPlaceholder {
			continue
		}
		token = strings.ReplaceAll(token, string(decimalCommaSubst), ",")
		token = strings.ReplaceAll(token, string(nonBreakingSlashSubst), "/")
		token = strings.ReplaceAll(token, string(nonBreakingColonSubst), ":")
		token = strings.ReplaceAll(token, string(nonBreakingSpaceSubst), " ")
		token = strings.ReplaceAll(token, string(leftBraceSubst), "(")
		token = strings.ReplaceAll(token, string(rightBraceSubst), ")")
		token = strings.ReplaceAll(token, string(leftAngleSubst), "<")
		token = strings.ReplaceAll(token, string(rightAngleSubst), ">")
		token = strings.ReplaceAll(token, string(slashSubst), "/")
		token = strings.ReplaceAll(token, string(nonBreakingDotSubst), ".")
		token = strings.ReplaceAll(token, softHyphenWrapSubst, softHyphenWrap)
		token = strings.ReplaceAll(token, nonBreakingPlaceholder, "")
		token = strings.ReplaceAll(token, nonBreakingPlaceholder2, "")
		for k, v := range urls {
			token = strings.ReplaceAll(token, k, v)
		}
		tokenList = append(tokenList, token)
	}
	return tokenList
}

func cleanup(text string) string {
	text = strings.ReplaceAll(text, "\u2019", "'")
	text = strings.ReplaceAll(text, "\u02BC", "'")
	text = strings.ReplaceAll(text, "\u2018", "'")
	text = strings.ReplaceAll(text, "\u201A", ",")
	text = strings.ReplaceAll(text, "\u2011", "-")
	text = weirdApostroph.ReplaceAllString(text, "$1"+nonBreakingPlaceholder2+"$2"+nonBreakingPlaceholder2+"$3")
	return text
}

func adjustTextForTokenizing(text string, urls map[string]string) string {
	text = cleanup(text)

	// Java: "\u2014\u2013-".indexOf(text.charAt(0)) >= 0 — first *char* (rune), not first byte.
	if text != "" {
		r, _ := utf8.DecodeRuneInString(text)
		if r == '\u2014' || r == '\u2013' || r == '-' {
			if m := leadingDash.FindStringSubmatch(text); m != nil {
				text = leadingDash.ReplaceAllString(text, "$1"+breakingPlaceholder+"$2")
			} else if m := leadingDash2.FindStringSubmatch(text); m != nil {
				text = leadingDash2.ReplaceAllString(text, "$1"+breakingPlaceholder+"$2")
			}
		}
	}

	if strings.Contains(text, ",") {
		text = decimalComma.ReplaceAllString(text, "$1"+string(decimalCommaSubst)+"$2")
	}

	if strings.Contains(text, "http") || strings.Contains(text, "www") || strings.Contains(text, "@") || strings.Contains(text, "ftp") {
		urlReplaceChar := urlStartReplaceChar
		for {
			loc := urlPattern.FindStringIndex(text)
			if loc == nil {
				break
			}
			urlGroup := text[loc[0]:loc[1]]
			replaceChar := string(rune(urlReplaceChar))
			urls[replaceChar] = urlGroup
			text = text[:loc[0]] + replaceChar + text[loc[1]:]
			urlReplaceChar++
		}
	}

	if strings.Contains(text, "\u2014") {
		text = regexp.MustCompile(`\x{2014}([\s\x{00A0}\x{202F}])`).ReplaceAllString(text, breakingPlaceholder+"\u2014$1")
		// also mid-word emdash without space
		text = regexp.MustCompile(`([а-яіїєґА-ЯІЇЄҐ])\x{2014}`).ReplaceAllString(text, "$1"+breakingPlaceholder+"\u2014")
		text = regexp.MustCompile(`\x{2014}([а-яіїєґА-ЯІЇЄҐ])`).ReplaceAllString(text, "\u2014"+breakingPlaceholder+"$1")
	}

	nDashPresent := strings.Contains(text, "\u2013")
	if strings.Contains(text, "-") || nDashPresent {
		text = dashNumbers.ReplaceAllString(text, "$1"+breakingPlaceholder+"$2"+breakingPlaceholder+"$3")
		if nDashPresent {
			// N_DASH_SPACE: break unless followed by та|чи|і|й (Java negative lookahead)
			text = breakNDashSpace(text)
			text = nDashSpace2.ReplaceAllString(text, "$1"+breakingPlaceholder+"$2")
		}
	}

	if strings.Contains(text, "с/г") {
		text = strings.ReplaceAll(text, "с/г", "с"+string(nonBreakingSlashSubst)+"г")
	}
	if strings.Contains(text, "Л/ДНР") {
		text = strings.ReplaceAll(text, "Л/ДНР", "Л"+string(nonBreakingSlashSubst)+"ДНР")
	}

	if strings.Contains(text, "р.") {
		text = yearWithR.ReplaceAllString(text, "$1"+breakingPlaceholder+"$2")
	}

	text = strings.ReplaceAll(text, "#", breakingPlaceholder+"#")
	if strings.Contains(text, "%") {
		text = regexp.MustCompile(`%([^-])`).ReplaceAllString(text, "%"+breakingPlaceholder+"$1")
	}

	text = compoundQuotes1.ReplaceAllString(text, "$1$2"+nonBreakingPlaceholder2+"$3"+nonBreakingPlaceholder2+"$4"+nonBreakingPlaceholder2)
	text = compoundQuotes2.ReplaceAllString(text, "$1"+nonBreakingPlaceholder2+"$2"+nonBreakingPlaceholder2+"$3"+nonBreakingPlaceholder2+"$4")
	if strings.Contains(text, "[") {
		text = wordsWithBrackets.ReplaceAllString(text, "$1["+nonBreakingPlaceholder2+"$2]"+nonBreakingPlaceholder2)
	}

	dotIndex := strings.IndexByte(text, '.')
	textRtrimmed := strings.TrimRight(text, " \t\n\r\u00A0\u202F")
	dotInsideSentence := dotIndex >= 0 && dotIndex < len(textRtrimmed)-1
	abbrAtEnd := abbrAtEnd.MatchString(text)

	if dotInsideSentence || (dotIndex == len(textRtrimmed)-1 && abbrAtEnd) {
		text = dottedNumbers3.ReplaceAllString(text, "$1."+nonBreakingPlaceholder2+"$2."+nonBreakingPlaceholder2+"$3")
		text = dottedNumbers.ReplaceAllString(text, "$1."+nonBreakingPlaceholder2+"$2")

		text = abbrNar1.ReplaceAllString(text, "$1."+nonBreakingPlaceholder2+breakingPlaceholder)
		text = abbrNar2.ReplaceAllString(text, "$1$2."+nonBreakingPlaceholder2+breakingPlaceholder+"$3")

		// Java: $1.\uE120\uE110$2.\uE120\uE110 with meter exclusion on $2
		text = replaceAbbrDot2Small(text)
		nb2 := string(nonBreakingDotSubst) + breakingPlaceholder
		text = abbrDotVO1.ReplaceAllString(text, "$1"+nb2+"$2"+nb2)
		text = abbrDotVO2.ReplaceAllString(text, "$1"+nb2+"$2"+nb2)
		text = abbrDotVO3.ReplaceAllString(text, "$1"+nb2+"$2"+nb2)
		text = abbrDotArt.ReplaceAllString(text, "$1"+string(nonBreakingDotSubst)+breakingPlaceholder+"$2")
		text = abbrDotMan.ReplaceAllString(text, "$1"+string(nonBreakingDotSubst)+breakingPlaceholder+"$2")
		text = abbrDotTys1.ReplaceAllString(text, "$1$2"+string(nonBreakingDotSubst)+breakingPlaceholder)
		text = abbrDotTys2.ReplaceAllString(text, "$1"+string(nonBreakingDotSubst)+breakingPlaceholder+"$2")
		text = abbrDotLat.ReplaceAllString(text, "$1$2"+string(nonBreakingDotSubst)+breakingPlaceholder+"$3")
		text = abbrDotProf.ReplaceAllString(text, "$1$2"+string(nonBreakingDotSubst)+breakingPlaceholder+"$3")
		text = abbrDotGub.ReplaceAllString(text, "$1"+string(nonBreakingDotSubst)+breakingPlaceholder)
		text = abbrDotDash.ReplaceAllString(text, "$1$2"+string(nonBreakingDotSubst)+"$3")

		text = initialsSP2.ReplaceAllString(text, "$1"+string(nonBreakingDotSubst)+breakingPlaceholder+"$2"+string(nonBreakingDotSubst)+breakingPlaceholder+"$3")
		text = initialsSP1.ReplaceAllString(text, "$1"+string(nonBreakingDotSubst)+breakingPlaceholder+"$2")
		text = initialsRSP2.ReplaceAllString(text, "$1"+breakingPlaceholder+"$2"+string(nonBreakingDotSubst)+breakingPlaceholder+"$3"+string(nonBreakingDotSubst)+breakingPlaceholder)
		text = initialsRSP1.ReplaceAllString(text, "$1"+breakingPlaceholder+"$2"+string(nonBreakingDotSubst)+breakingPlaceholder)

		text = abbrDotKub.ReplaceAllString(text, "$1."+nonBreakingPlaceholder2+breakingPlaceholder+"$2")
		text = abbrDotSG.ReplaceAllString(text, "$1"+string(nonBreakingDotSubst)+"$2"+string(nonBreakingDotSubst)+breakingPlaceholder)
		text = abbrDotChl.ReplaceAllString(text, "$1."+nonBreakingPlaceholder2+"$2."+nonBreakingPlaceholder2+breakingPlaceholder)
		text = abbrDotPn.ReplaceAllString(text, "$1."+nonBreakingPlaceholder2+breakingPlaceholder+"$2."+nonBreakingPlaceholder2+breakingPlaceholder)
		text = abbrITP.ReplaceAllString(text, "$1"+nonBreakingPlaceholder2+breakingPlaceholder+"$2"+nonBreakingPlaceholder2+breakingPlaceholder)
		text = abbrITCH.ReplaceAllString(text, "$1"+nonBreakingPlaceholder2+breakingPlaceholder+"$2"+nonBreakingPlaceholder2+breakingPlaceholder)
		text = abbrTZV.ReplaceAllString(text, "$1"+nonBreakingPlaceholder2+breakingPlaceholder+"$2"+nonBreakingPlaceholder2+breakingPlaceholder)
		text = abbrRedAvt.ReplaceAllString(text, "$1."+nonBreakingPlaceholder2+breakingPlaceholder+"$3")
		// Java ABBR_DOT_NON_ENDING: \.(?!\uE120|\.+[\h\v]*$)
		text = replaceAbbrNonEnding(text)
		text = replaceAbbrDotV(text)
		text = abbrNonEnding2.ReplaceAllString(text, "$1"+nonBreakingPlaceholder2+breakingPlaceholder+"$2")
		text = invalidMln.ReplaceAllString(text, "$1."+nonBreakingPlaceholder2+breakingPlaceholder+"$2")
	}

	if dotInsideSentence {
		text = replaceWebEntities(text)
		text = webEntities2.ReplaceAllString(text, "."+nonBreakingPlaceholder2+"$1."+nonBreakingPlaceholder2+"$2")
	}

	// Java ABBR_DOT_ENDING: \.(?!\uE120)
	text = replaceAbbrEnding(text)

	// spaced decimals: protect groups like "2 000" and "12 000 000"
	text = protectSpacedNumbers(text)

	if strings.Contains(text, ":") {
		text = colonNumbers.ReplaceAllString(text, "$1"+string(nonBreakingColonSubst)+"$2")
	}

	if strings.Contains(text, "(") {
		text = braceInWord.ReplaceAllString(text, "$1"+string(leftBraceSubst)+"$2"+string(rightBraceSubst))
	}

	if strings.Contains(text, "<") {
		text = xmlTag.ReplaceAllString(text, breakingPlaceholder+string(leftAngleSubst)+"$1"+string(rightAngleSubst)+breakingPlaceholder)
		text = strings.ReplaceAll(text, string(leftAngleSubst)+"/", string(leftAngleSubst)+string(slashSubst))
		text = strings.ReplaceAll(text, "/"+string(rightAngleSubst), string(slashSubst)+string(rightAngleSubst))
	}

	if strings.Contains(text, "-") {
		text = regexp.MustCompile(`([а-яіїєґА-ЯІЇЄҐ])([»"\-]+-)`).ReplaceAllString(text, "$1"+breakingPlaceholder+"$2")
		text = regexp.MustCompile(`([»"\-]+-)([а-яіїєґА-ЯІЇЄҐ])`).ReplaceAllString(text, "$1"+breakingPlaceholder+"$2")
	}

	if strings.Contains(text, softHyphenWrap) {
		text = regexp.MustCompile(`([^\s])`+softHyphenWrap).ReplaceAllString(text, "$1"+softHyphenWrapSubst)
	}

	if strings.Contains(text, "'") {
		text = splitBeginApostrophe(text)
		text = protectOrSplitEndApostrophe(text)
	}

	if strings.Contains(text, "+") {
		text = breakPlus(text)
	}

	if len(text) > 1 && (strings.Contains(text, "-") || strings.Contains(text, "\u2013")) {
		text = breakLeadingSignedNumber(text)
	}

	text = numberMissingSpace.ReplaceAllString(text, "$1"+breakingPlaceholder+"$2")
	return text
}

// replaceWebEntities ports Java WEB_ENTITIES with Unicode-aware \b after the TLD.
func replaceWebEntities(text string) string {
	var b strings.Builder
	last := 0
	for _, loc := range webEntities.FindAllStringSubmatchIndex(text, -1) {
		full0, full1 := loc[0], loc[1]
		// Java \b: right edge must be end or non-word (Unicode letter/digit/_)
		if full1 < len(text) {
			r, _ := utf8.DecodeRuneInString(text[full1:])
			if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
				continue
			}
		}
		b.WriteString(text[last:full0])
		g1 := text[loc[2]:loc[3]]
		g2 := text[loc[4]:loc[5]]
		b.WriteString(g1 + "." + nonBreakingPlaceholder2 + g2)
		last = full1
	}
	b.WriteString(text[last:])
	return b.String()
}

// replaceAbbrDot2Small ports Java ABBR_DOT_2_SMALL_LETTERS_PATTERN +
// replaceAll("$1.\uE120\uE110$2.\uE120\uE110") with (?![смкд]?м\.) on the second group.
func replaceAbbrDot2Small(text string) string {
	var b strings.Builder
	last := 0
	for _, loc := range abbrDot2Small.FindAllStringSubmatchIndex(text, -1) {
		// groups: full, $1 prefix, $2 first abbr, $3 second abbr (with optional spaces)
		full0, full1 := loc[0], loc[1]
		g3 := text[loc[6]:loc[7]]
		// Java (?![смкд]?м\.) — skip when second token is a meter unit
		if abbrDot2SmallMeterSecond.MatchString(g3) {
			continue
		}
		b.WriteString(text[last:full0])
		g1 := text[loc[2]:loc[3]]
		g2 := text[loc[4]:loc[5]]
		// Java: $1 includes prefix+letters then .\uE120\uE110; Go pattern splits prefix/$2
		b.WriteString(g1 + g2 + "." + nonBreakingPlaceholder2 + breakingPlaceholder + g3 + "." + nonBreakingPlaceholder2 + breakingPlaceholder)
		last = full1
	}
	b.WriteString(text[last:])
	return b.String()
}

// replaceAbbrNonEnding ports Java ABBR_DOT_NON_ENDING_PATTERN with
// \.(?!\uE120|\.+[\h\v]*$) after the abbreviation.
func replaceAbbrNonEnding(text string) string {
	var b strings.Builder
	last := 0
	for _, loc := range abbrNonEnding.FindAllStringSubmatchIndex(text, -1) {
		full0, full1 := loc[0], loc[1]
		rest := text[full1:]
		if strings.HasPrefix(rest, nonBreakingPlaceholder2) {
			continue // already protected (\uE120)
		}
		// Java \.+[\h\v]*$ — ellipsis (or more dots) to end of string: do not protect
		if isDotsThenEnd(rest) {
			continue
		}
		b.WriteString(text[last:full0])
		g1 := text[loc[2]:loc[3]]
		g2 := text[loc[4]:loc[5]]
		b.WriteString(g1 + g2 + "." + nonBreakingPlaceholder2 + breakingPlaceholder)
		last = full1
	}
	b.WriteString(text[last:])
	return b.String()
}

// replaceAbbrDotV ports bare "в." / "В." when not ellipsis and not already protected.
func replaceAbbrDotV(text string) string {
	var b strings.Builder
	last := 0
	for _, loc := range abbrDotV.FindAllStringSubmatchIndex(text, -1) {
		full0, full1 := loc[0], loc[1]
		// full match includes the char after the dot ($3); only rewrite if not E120
		g3 := text[loc[6]:loc[7]]
		if g3 == nonBreakingPlaceholder2 {
			continue
		}
		b.WriteString(text[last:full0])
		g1 := text[loc[2]:loc[3]]
		g2 := text[loc[4]:loc[5]]
		b.WriteString(g1 + g2 + "." + nonBreakingPlaceholder2 + breakingPlaceholder + g3)
		last = full1
	}
	b.WriteString(text[last:])
	return b.String()
}

// replaceAbbrEnding ports Java ABBR_DOT_ENDING_PATTERN with \.(?!\uE120).
// Replacement is $1.\uE120\uE110 where $1 is (prefix+abbr) including the left boundary char.
func replaceAbbrEnding(text string) string {
	var b strings.Builder
	last := 0
	for _, loc := range abbrEnding.FindAllStringSubmatchIndex(text, -1) {
		full0, full1 := loc[0], loc[1]
		rest := text[full1:]
		if strings.HasPrefix(rest, nonBreakingPlaceholder2) {
			continue
		}
		b.WriteString(text[last:full0])
		// full match ends with '.'; protect: (prefix+abbr).\uE120\uE110
		m := text[full0:full1]
		b.WriteString(m[:len(m)-1] + "." + nonBreakingPlaceholder2 + breakingPlaceholder)
		last = full1
	}
	b.WriteString(text[last:])
	return b.String()
}

// isDotsThenEnd ports Java \.+[\h\v]*$ for the non-ending negative lookahead.
func isDotsThenEnd(rest string) bool {
	if rest == "" {
		return false
	}
	i := 0
	rr := []rune(rest)
	if rr[0] != '.' {
		return false
	}
	for i < len(rr) && rr[i] == '.' {
		i++
	}
	if i == 0 {
		return false
	}
	for i < len(rr) {
		if !isJavaHOrVSpace(rr[i]) {
			return false
		}
		i++
	}
	return true
}

func protectSpacedNumbers(text string) string {
	// Java: (?<=^|[\h\v(])\d{1,3}([\h][\d]{3})+(?=[\h\v(]|$)
	// Scan only at valid left boundaries so "01.01.42 400 000" still protects "400 000".
	re := regexp.MustCompile(`^\d{1,3}(?:[\s\x{00A0}\x{202F}]\d{3})+`)
	var b strings.Builder
	runes := []rune(text)
	i := 0
	for i < len(runes) {
		atBoundary := i == 0 || isSpaceLike(runes[i-1]) || runes[i-1] == '('
		if atBoundary && runes[i] >= '0' && runes[i] <= '9' {
			// try match from byte offset
			byteOff := len(string(runes[:i]))
			if loc := re.FindStringIndex(text[byteOff:]); loc != nil && loc[0] == 0 {
				endByte := byteOff + loc[1]
				// right boundary
				okRight := endByte == len(text)
				if !okRight {
					next, _ := utf8.DecodeRuneInString(text[endByte:])
					okRight = isSpaceLike(next) || next == '('
				}
				if okRight {
					m := text[byteOff:endByte]
					out := strings.ReplaceAll(m, " ", string(nonBreakingSpaceSubst))
					out = strings.ReplaceAll(out, "\u00A0", string(nonBreakingSpaceSubst))
					out = strings.ReplaceAll(out, "\u202F", string(nonBreakingSpaceSubst))
					b.WriteString(out)
					// advance i by rune count of m
					i += len([]rune(m))
					continue
				}
			}
		}
		b.WriteRune(runes[i])
		i++
	}
	return b.String()
}

func breakPlus(text string) string {
	// Java: text.replaceAll("\\+(?=[а-яіїєґА-ЯІЇЄҐ0-9])", BREAKING_PLACEHOLDER + "+" + BREAKING_PLACEHOLDER)
	// Only Ukrainian letters + digits in the lookahead — not Latin (foo+bar stays one token).
	var b strings.Builder
	runes := []rune(text)
	for i := 0; i < len(runes); i++ {
		if runes[i] == '+' {
			// +20, +займенник, mid word+word (Cyrillic)
			if i+1 < len(runes) {
				n := runes[i+1]
				if isUKCyrLetter(n) || (n >= '0' && n <= '9') {
					b.WriteString(breakingPlaceholder)
					b.WriteRune('+')
					b.WriteString(breakingPlaceholder)
					continue
				}
			}
		}
		b.WriteRune(runes[i])
	}
	return b.String()
}

func breakLeadingSignedNumber(text string) string {
	// (^|space)(-|–)(?=digit)
	var b strings.Builder
	runes := []rune(text)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if (r == '-' || r == '\u2013') && i+1 < len(runes) && runes[i+1] >= '0' && runes[i+1] <= '9' {
			if i == 0 || isSpaceLike(runes[i-1]) {
				b.WriteRune(r)
				b.WriteString(breakingPlaceholder)
				continue
			}
		}
		b.WriteRune(r)
	}
	return b.String()
}

func breakNDashSpace(text string) string {
	// Java N_DASH_SPACE_PATTERN:
	// ([а-яіїєґa-z0-9])(\u2013\h)(?!(та|чи|і|й)[\h\v])  CASE_INSENSITIVE|UNICODE_CASE
	// RE2 has no lookahead — skip break only when rest matches (та|чи|і|й)+whitespace.
	re := nDashSpace
	var b strings.Builder
	last := 0
	for _, loc := range re.FindAllStringSubmatchIndex(text, -1) {
		// loc: full, g1, g2 — byte indices
		full0, full1 := loc[0], loc[1]
		rest := text[full1:]
		// Java negative lookahead: do not break when (та|чи|і|й)[\h\v] follows
		if followedByConjunctionHV(rest) {
			b.WriteString(text[last:full1])
			last = full1
			continue
		}
		b.WriteString(text[last:full0])
		// groups: letter, ndash+space
		g1 := text[loc[2]:loc[3]]
		g2 := text[loc[4]:loc[5]]
		b.WriteString(g1 + breakingPlaceholder + g2)
		last = full1
	}
	b.WriteString(text[last:])
	return b.String()
}

// followedByConjunctionHV ports Java (?=(та|чи|і|й)[\h\v]) for the negative lookahead.
// Conjunction alone at end of string does NOT block (Java requires whitespace after).
func followedByConjunctionHV(rest string) bool {
	rr := []rune(strings.ToLower(rest))
	for _, w := range []string{"та", "чи", "і", "й"} {
		wr := []rune(w)
		if len(rr) < len(wr)+1 {
			continue
		}
		match := true
		for i := range wr {
			if rr[i] != wr[i] {
				match = false
				break
			}
		}
		if !match {
			continue
		}
		// Java [\h\v] after conjunction
		if isJavaHOrVSpace(rr[len(wr)]) {
			return true
		}
	}
	return false
}

// isJavaHOrVSpace approximates Java Pattern \h (horizontal) | \v (vertical) whitespace.
func isJavaHOrVSpace(r rune) bool {
	switch r {
	case '\t', '\n', '\v', '\f', '\r', ' ', '\u00A0', '\u1680', '\u180E',
		'\u2000', '\u2001', '\u2002', '\u2003', '\u2004', '\u2005', '\u2006',
		'\u2007', '\u2008', '\u2009', '\u200A', '\u200B', // ZWSP appears in some \h tables
		'\u2028', '\u2029', // line/paragraph separator (\v)
		'\u202F', '\u205F', '\u3000':
		return true
	default:
		return unicode.Is(unicode.Zs, r) || unicode.Is(unicode.Zl, r) || unicode.Is(unicode.Zp, r)
	}
}

func isSpaceLike(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == '\u00A0' || r == '\u202F'
}

func splitBeginApostrophe(text string) string {
	// (^|[\s(„«"'])'(?!дно)(\p{L})
	var b strings.Builder
	runes := []rune(text)
	for i := 0; i < len(runes); i++ {
		if runes[i] == '\'' {
			atBoundary := i == 0 || isSpaceLike(runes[i-1]) || strings.ContainsRune(`(„«"'`, runes[i-1])
			if atBoundary && i+1 < len(runes) && unicode.IsLetter(runes[i+1]) {
				// don't split 'дно...
				rest := string(runes[i+1:])
				if strings.HasPrefix(strings.ToLower(rest), "дно") {
					b.WriteRune('\'')
					continue
				}
				b.WriteRune('\'')
				b.WriteString(breakingPlaceholder)
				continue
			}
		}
		b.WriteRune(runes[i])
	}
	return b.String()
}

func protectOrSplitEndApostrophe(text string) string {
	// (\p{L})'(?![p{L}-]) but keep мо' тре' etc
	var b strings.Builder
	runes := []rune(text)
	for i := 0; i < len(runes); i++ {
		if runes[i] == '\'' && i > 0 && unicode.IsLetter(runes[i-1]) {
			// find word start
			j := i - 1
			for j >= 0 && unicode.IsLetter(runes[j]) {
				j--
			}
			word := string(runes[j+1 : i])
			// next char
			nextOK := i+1 >= len(runes) || !unicode.IsLetter(runes[i+1]) && runes[i+1] != '-'
			if nextOK && !colloquialApos[strings.ToLower(word)] {
				b.WriteString(breakingPlaceholder)
				b.WriteRune('\'')
				continue
			}
		}
		b.WriteRune(runes[i])
	}
	return b.String()
}

// splitWithDelimiters implements SPLIT_CHARS logic with context (no RE2 lookbehind).
func splitWithDelimiters(str string) []string {
	return splitUK(str)
}

func matchMultiPunct(runes []rune, i int) int {
	// !{2,3} ?{2,3} .{3} [!?][!?.]{1,2}
	if i >= len(runes) {
		return 0
	}
	r := runes[i]
	if r == '!' {
		n := 1
		for i+n < len(runes) && runes[i+n] == '!' && n < 3 {
			n++
		}
		if n >= 2 {
			return n
		}
		// !.? style
		if i+1 < len(runes) && (runes[i+1] == '!' || runes[i+1] == '?' || runes[i+1] == '.') {
			n = 1
			for i+n < len(runes) && n < 3 {
				c := runes[i+n]
				if c == '!' || c == '?' || c == '.' {
					n++
				} else {
					break
				}
			}
			if n >= 2 {
				return n
			}
		}
	}
	if r == '?' {
		n := 1
		for i+n < len(runes) && runes[i+n] == '?' && n < 3 {
			n++
		}
		if n >= 2 {
			return n
		}
		if i+1 < len(runes) && (runes[i+1] == '!' || runes[i+1] == '?' || runes[i+1] == '.') {
			n = 1
			for i+n < len(runes) && n < 3 {
				c := runes[i+n]
				if c == '!' || c == '?' || c == '.' {
					n++
				} else {
					break
				}
			}
			if n >= 2 {
				return n
			}
		}
	}
	if r == '.' && i+2 < len(runes) && runes[i+1] == '.' && runes[i+2] == '.' {
		return 3
	}
	return 0
}

func isUKCyrLetter(r rune) bool {
	switch {
	case r >= 'а' && r <= 'я', r >= 'А' && r <= 'Я':
		return true
	case r == 'і' || r == 'І' || r == 'ї' || r == 'Ї' || r == 'є' || r == 'Є' || r == 'ґ' || r == 'Ґ':
		return true
	}
	return false
}

func isLatinLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func isSimpleDelim(r rune) bool {
	switch r {
	case ' ', '\u00A0', '\n', '\r', '\t',
		',', '.', ';', '!', '?', '\u2014', '\u2015', ':',
		'(', ')', '[', ']', '{', '}', '<', '>', '/', '|', '\\',
		'…', '°', '$', '€', '₴', '=', '№', '§', '¿', '¡', '~', '×':
		return true
	}
	// horizontal spaces U+2000-200F partial handled below
	return false
}

func isEmojiOrSpecial(r rune) bool {
	// Breaking placeholder must split (Java SPLIT_CHARS includes \uE110).
	// Note: range is F000–FFFF (not E000); our E001–E120 substs are outside F000.
	if r == '\uE110' {
		return true
	}
	if r >= 0x2000 && r <= 0x200F {
		return true
	}
	if r == 0x201A {
		return true
	}
	if r >= 0x2020 && r <= 0x202F {
		return true
	}
	if r >= 0x2030 && r <= 0x206F {
		return true
	}
	if r >= 0x2400 && r <= 0x27FF {
		return true
	}
	if r >= 0x1F000 && r <= 0x1FFFF {
		return true
	}
	if r >= 0xF000 && r <= 0xFFFF {
		return true
	}
	return false
}

func splitUK(str string) []string {
	runes := []rune(str)
	var parts []string
	var cur []rune

	flush := func() {
		if len(cur) > 0 {
			parts = append(parts, string(cur))
			cur = nil
		}
	}
	followedByE120 := func(end int) bool {
		return end < len(runes) && runes[end] == '\uE120'
	}
	prevIsUKLetter := func(idx int) bool {
		return idx > 0 && isUKCyrLetter(runes[idx-1])
	}
	prevIsLetter := func(idx int) bool {
		return idx > 0 && (isUKCyrLetter(runes[idx-1]) || isLatinLetter(runes[idx-1]))
	}
	nextIsLetterOrDigit := func(idx int) bool {
		if idx >= len(runes) {
			return false
		}
		r := runes[idx]
		return isUKCyrLetter(r) || isLatinLetter(r) || (r >= '0' && r <= '9')
	}

	i := 0
	for i < len(runes) {
		// multi-char punct
		if n := matchMultiPunct(runes, i); n > 0 && !followedByE120(i+n) {
			flush()
			parts = append(parts, string(runes[i:i+n]))
			i += n
			continue
		}

		r := runes[i]

		// % with lookahead (?![-–][cyrillic])
		if r == '%' {
			ok := true
			if i+1 < len(runes) {
				n1 := runes[i+1]
				if (n1 == '-' || n1 == '\u2013') && i+2 < len(runes) && isUKCyrLetter(runes[i+2]) {
					ok = false // 5%-й keep together - don't split %
				}
			}
			if ok && !followedByE120(i+1) {
				flush()
				parts = append(parts, "%")
				i++
				continue
			}
			// keep with word
			cur = append(cur, r)
			i++
			continue
		}

		// quotes unless after E109
		if strings.ContainsRune(`"«»„“”`, r) {
			if i > 0 && runes[i-1] == '\uE109' {
				cur = append(cur, r)
				i++
				continue
			}
			if !followedByE120(i + 1) {
				flush()
				parts = append(parts, string(r))
				i++
				continue
			}
			cur = append(cur, r)
			i++
			continue
		}

		// superscript after cyrillic letter
		if (r == '\u00B9' || r == '\u00B2' || (r >= '\u2070' && r <= '\u2079')) && prevIsUKLetter(i) {
			if !followedByE120(i + 1) {
				flush()
				parts = append(parts, string(r))
				i++
				continue
			}
		}

		// _* sequences
		if r == '_' || r == '*' {
			// start of word: (?<![letter])[_*]+
			if !prevIsLetter(i) {
				j := i
				for j < len(runes) && (runes[j] == '_' || runes[j] == '*') {
					j++
				}
				if !followedByE120(j) {
					flush()
					parts = append(parts, string(runes[i:j]))
					i = j
					continue
				}
			}
			// end of word: [_*]+(?![letter digit])
			j := i
			for j < len(runes) && (runes[j] == '_' || runes[j] == '*') {
				j++
			}
			if !nextIsLetterOrDigit(j) && !followedByE120(j) {
				flush()
				parts = append(parts, string(runes[i:j]))
				i = j
				continue
			}
			cur = append(cur, r)
			i++
			continue
		}

		// simple single delims
		if isSimpleDelim(r) {
			if !followedByE120(i + 1) {
				flush()
				parts = append(parts, string(r))
				i++
				continue
			}
			// followed by E120 - absorb delimiter into token (protected)
			cur = append(cur, r)
			i++
			continue
		}

		if isEmojiOrSpecial(r) {
			if !followedByE120(i + 1) {
				flush()
				parts = append(parts, string(r))
				i++
				continue
			}
			cur = append(cur, r)
			i++
			continue
		}

		// + is NOT in default split chars - handled by placeholders
		// - hyphen not in split chars
		// ' apostrophe not in split chars

		cur = append(cur, r)
		i++
	}
	flush()
	return parts
}
