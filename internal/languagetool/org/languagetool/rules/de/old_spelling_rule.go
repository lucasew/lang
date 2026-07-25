package de

import (
	"embed"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"unicode"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

//go:embed data/alt_neu.csv
var altNeuFS embed.FS

// oldSpellingExceptions ports OldSpellingRule.EXCEPTIONS.
var oldSpellingExceptions = []string{
	"Schloß Holte", "Schloß Neuhaus", "Schloß Ricklingen", "Schloß-Nauses",
	"Schloß Rötteln", "Klinikum Schloß Winnenden", "Grazer Schloßberg",
	"Höchster Schloß", "Bell Telephone", "Telephone Company", "American Telephone",
	"England Telephone", "Mobile Telephone", "Cordless Telephone", "Telephone Line",
	"World Telephone", "Tip Top", "Hans Joachim Blaß", "kurz fassen",
}

// Java: private static final Pattern CHARS = Pattern.compile("[a-zA-Zöäüß]");
var oldSpellingCharsRE = regexp.MustCompile(`^[a-zA-Zöäüß]$`)

type oldSpellingHit struct {
	begin, end int // UTF-16 offsets into sentence text
	value      string
}

var (
	oldSpellingOnce sync.Once
	oldSpellingData *SpellingData // Java: Suppliers.memoize(() -> new SpellingData(FILE_PATH))
)

// oldSpellingFILEPath ports OldSpellingRule.FILE_PATH.
const oldSpellingFILEPath = "/de/alt_neu.csv"

// loadOldSpellingData ports OldSpellingRule.DATA (memoized SpellingData for FILE_PATH).
func loadOldSpellingData() *SpellingData {
	oldSpellingOnce.Do(func() {
		// Java: new SpellingData(FILE_PATH) — getCoherencyMap + GermanSynthesizer expand + tries.
		oldSpellingData = NewSpellingData(oldSpellingFILEPath)
	})
	return oldSpellingData
}

// loadOldSpelling returns body map + longest-first keys (compat for tests that assert map entries).
func loadOldSpelling() (map[string]string, []string) {
	sd := loadOldSpellingData()
	m := sd.Map()
	return m, keysLongestFirst(m)
}

// readAltNeuCSV ports JLanguageTool.getDataBroker().getFromResourceDirAsLines("/de/alt_neu.csv").
func readAltNeuCSV() ([]byte, error) {
	if root := DiscoverGermanResourceDir(); root != "" {
		p := filepath.Join(root, "alt_neu.csv")
		if b, err := os.ReadFile(p); err == nil {
			return b, nil
		}
	}
	return altNeuFS.ReadFile("data/alt_neu.csv")
}

func keysLongestFirst(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		li, lj := utf16Len(keys[i]), utf16Len(keys[j])
		if li != lj {
			return li > lj
		}
		return keys[i] < keys[j]
	})
	return keys
}

// OldSpellingRule ports org.languagetool.rules.de.OldSpellingRule.
// Match uses full-text scan of SpellingData tries (body + sentence-start), not token invent.
type OldSpellingRule struct {
	messages  map[string]string
	austrian  bool
	Category  *rules.Category
	IssueType rules.ITSIssueType
}

func NewOldSpellingRule(messages map[string]string) *OldSpellingRule {
	_ = loadOldSpellingData()
	return &OldSpellingRule{
		messages:  messages,
		Category:  rules.CatTypos.GetCategory(messages),
		IssueType: rules.ITSMisspelling,
	}
}

// NewOldSpellingRuleAT is the Austrian variant (Geschoß remains acceptable).
func NewOldSpellingRuleAT(messages map[string]string) *OldSpellingRule {
	r := NewOldSpellingRule(messages)
	r.austrian = true
	return r
}

func (r *OldSpellingRule) GetID() string { return "OLD_SPELLING_RULE" }

func (r *OldSpellingRule) GetDescription() string {
	return "Findet Schreibweisen, die nur in der alten Rechtschreibung gültig waren"
}

func (r *OldSpellingRule) GetCategory() *rules.Category {
	if r == nil {
		return nil
	}
	return r.Category
}

func (r *OldSpellingRule) GetLocQualityIssueType() rules.ITSIssueType {
	if r == nil || r.IssueType == "" {
		return rules.ITSMisspelling
	}
	return r.IssueType
}

// Match ports OldSpellingRule.match: SpellingData AhoCorasick parseText on full text,
// Collections.reverse (longest first), ignoreMatch, then sentence-start trie at begin==0.
func (r *OldSpellingRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || sentence == nil {
		return nil
	}
	sd := loadOldSpellingData()
	text := sentence.GetText()
	// Java: DATA.get().getTrie().parseText(text)
	rawHits := sd.GetTrie().ParseText(text)
	hits := make([]oldSpellingHit, len(rawHits))
	for i, h := range rawHits {
		hits[i] = oldSpellingHit{begin: h.Begin, end: h.End, value: h.Value}
	}
	// Java: Collections.reverse(hits) — work on longest matches first.
	// hankcs order is left-to-right by end; reverse + length-desc sort is faithful for overlap skip.
	sort.SliceStable(hits, func(i, j int) bool {
		li, lj := hits[i].end-hits[i].begin, hits[j].end-hits[j].begin
		if li != lj {
			return li > lj
		}
		return hits[i].begin < hits[j].begin
	})

	startPositions := map[int]struct{}{}
	var matches []*rules.RuleMatch
	for _, hit := range hits {
		if _, ok := startPositions[hit.begin]; ok {
			continue
		}
		if ignoreOldSpellingMatch(hit, text) {
			continue
		}
		// Java: always add hit.begin to startPositions after addMatch, even when
		// addMatch returns early (de-AT Geschoß skip) without appending a match.
		if m := r.addOldSpellingMatch(sentence, hit); m != nil {
			matches = append(matches, m)
		}
		startPositions[hit.begin] = struct{}{}
	}
	// Sentence-start trie: only hit.begin == 0
	// Java: DATA.get().getSentenceStartTrie().parseText(text)
	sentRaw := sd.GetSentenceStartTrie().ParseText(text)
	sentHits := make([]oldSpellingHit, len(sentRaw))
	for i, h := range sentRaw {
		sentHits[i] = oldSpellingHit{begin: h.Begin, end: h.End, value: h.Value}
	}
	sort.SliceStable(sentHits, func(i, j int) bool {
		return (sentHits[i].end - sentHits[i].begin) > (sentHits[j].end - sentHits[j].begin)
	})
	for _, hit := range sentHits {
		if hit.begin != 0 {
			continue
		}
		if _, ok := startPositions[hit.begin]; ok {
			continue
		}
		if ignoreOldSpellingMatch(hit, text) {
			continue
		}
		if m := r.addOldSpellingMatch(sentence, hit); m != nil {
			matches = append(matches, m)
		}
		break // Java: only one match at sentence start
	}
	return matches
}

// addOldSpellingMatch ports addMatch (may skip AT Geschoß).
func (r *OldSpellingRule) addOldSpellingMatch(sentence *languagetool.AnalyzedSentence, hit oldSpellingHit) *rules.RuleMatch {
	message := "Diese Schreibweise war nur in der alten Rechtschreibung korrekt."
	suggs := strings.Split(hit.value, "|")
	covered := substringByUTF16(sentence.GetText(), hit.begin, hit.end)
	if len(suggs) > 0 {
		// Java: StringUtils.replaceOnce(suggestions[0], "ss", "ß").equals(covered)
		ssForm := strings.Replace(suggs[0], "ss", "ß", 1)
		if ssForm == covered {
			if r != nil && r.austrian && strings.Contains(strings.ToLower(covered), "geschoß") {
				return nil
			}
			message += " Das Wort wird mit 'ss' geschrieben, wenn davor eine kurz gesprochene Silbe steht."
		}
	}
	rm := rules.NewRuleMatch(r, sentence, hit.begin, hit.end, message)
	rm.ShortMessage = "Alte Rechtschreibung"
	rm.SetSuggestedReplacements(suggs)
	return rm
}

// ignoreOldSpellingMatch ports OldSpellingRule.ignoreMatch.
func ignoreOldSpellingMatch(hit oldSpellingHit, text string) bool {
	// EXCEPTIONS via regionMatches(true, …) at hit.begin or hit.end-exception.length
	for _, exception := range oldSpellingExceptions {
		exLen := utf16Len(exception)
		if regionMatchesIgnoreCaseUTF16(text, hit.begin, exception, 0, exLen) {
			return true
		}
		if hit.end >= exLen && regionMatchesIgnoreCaseUTF16(text, hit.end-exLen, exception, 0, exLen) {
			return true
		}
	}
	tl := utf16Len(text)
	// boundary: previous char
	if hit.begin > 0 {
		prev := substringByUTF16(text, hit.begin-1, hit.begin)
		if !isOldSpellingBoundary(prev) {
			return true
		}
	}
	// boundary: next char
	if hit.end < tl {
		next := substringByUTF16(text, hit.end, hit.end+1)
		if !isOldSpellingBoundary(next) {
			return true
		}
	}
	// Prof. before (6 UTF-16 units back) — Java text.startsWith("Prof.", hit.begin-6)
	if hit.begin-6 >= 0 && utf16StartsWithAt(text, hit.begin-6, "Prof.") {
		return true
	}
	if hit.begin-5 >= 0 {
		before5 := substringByUTF16(text, hit.begin-5, hit.begin-1)
		if before5 == "Herr" || before5 == "Frau" {
			return true
		}
	}
	if hit.begin-4 >= 0 {
		before4 := substringByUTF16(text, hit.begin-4, hit.begin-1)
		if before4 == "Hr." || before4 == "Fr." || before4 == "Dr." {
			return true
		}
	}
	return false
}

// isOldSpellingBoundary ports isBoundary: !CHARS.matcher(s).matches()
func isOldSpellingBoundary(s string) bool {
	return !oldSpellingCharsRE.MatchString(s)
}

// regionMatchesIgnoreCaseUTF16 ports String.regionMatches(ignoreCase, toffset, other, ooffset, len).
func regionMatchesIgnoreCaseUTF16(text string, toffset int, other string, ooffset, length int) bool {
	if length < 0 || toffset < 0 || ooffset < 0 {
		return false
	}
	tu := utf16.Encode([]rune(text))
	ou := utf16.Encode([]rune(other))
	if toffset+length > len(tu) || ooffset+length > len(ou) {
		return false
	}
	for i := 0; i < length; i++ {
		a, b := rune(tu[toffset+i]), rune(ou[ooffset+i])
		if a == b {
			continue
		}
		// Java Character.toLowerCase for regionMatches ignoreCase
		if unicode.ToLower(a) != unicode.ToLower(b) {
			return false
		}
	}
	return true
}

// utf16StartsWithAt ports String.startsWith(prefix, toffset) — case-sensitive UTF-16.
func utf16StartsWithAt(text string, at int, prefix string) bool {
	pl := utf16Len(prefix)
	if at < 0 || at+pl > utf16Len(text) {
		return false
	}
	return substringByUTF16(text, at, at+pl) == prefix
}

// lookupOldSpelling remains for tests of SpellingData maps.
func lookupOldSpelling(phrase string, m map[string]string) (string, bool) {
	if neu, ok := m[phrase]; ok {
		return neu, true
	}
	// Sentence-start style: title of lowercase key (tests / expand path)
	if tools.StartsWithUppercase(phrase) && utf16Len(phrase) > 0 {
		// lowercase first UTF-16 unit only
		u := utf16.Encode([]rune(phrase))
		first := string(utf16.Decode(u[:1]))
		rest := string(utf16.Decode(u[1:]))
		low := strings.ToLower(first) + rest
		if neu, ok := m[low]; ok {
			return capitalizeSuggestions(neu), true
		}
	}
	return "", false
}

func capitalizeSuggestions(neu string) string {
	parts := strings.Split(neu, "|")
	for i, p := range parts {
		parts[i] = tools.UppercaseFirstChar(p)
	}
	return strings.Join(parts, "|")
}

func substringByUTF16(s string, from, to int) string {
	if from < 0 {
		from = 0
	}
	u := utf16.Encode([]rune(s))
	if from > len(u) {
		return ""
	}
	if to > len(u) {
		to = len(u)
	}
	if from >= to {
		return ""
	}
	return string(utf16.Decode(u[from:to]))
}

func utf16Len(s string) int {
	return len(utf16.Encode([]rune(s)))
}
