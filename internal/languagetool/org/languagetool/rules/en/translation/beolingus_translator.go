package translation

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	rulesde "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/de"
	base "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/translation"
	detag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/de"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// BeoLingusTranslator ports org.languagetool.rules.en.translation.BeoLingusTranslator.
// German / English translator.
type BeoLingusTranslator struct {
	tagger    *detag.GermanTagger
	de2en     map[string][]base.TranslationEntry
	en2de     map[string][]base.TranslationEntry
	inflector *Inflector
}

// Source: https://www.cafe-lingua.de/englische-grammatik/verben-mit-to-infinitiv.php
// Java BeoLingusTranslator.verbsWithTo
var verbsWithTo = map[string]struct{}{
	"afford": {}, "agree": {}, "aim": {}, "appear": {}, "arrange": {},
	"attempt": {}, "beg": {}, "care": {}, "choose": {}, "claim": {}, "condescend": {}, "consent": {},
	"dare": {}, "decide": {}, "demand": {}, "deserve": {},
	"determine": {}, "endeavour": {}, "expect": {}, "fail": {}, "forget": {}, "guarantee": {},
	"happen": {}, "have": {}, "help": {}, "hesitate": {}, "hope": {},
	"learn": {}, "long": {}, "manage": {}, "mean": {}, "need": {}, "neglect": {}, "offer": {},
	"plan": {}, "prepare": {}, "pretend": {}, "proceed": {}, "promise": {},
	"refuse": {}, "resolve": {}, "seem": {}, "stop": {}, "swear": {}, "tend": {}, "threaten": {},
	"trouble": {}, "undertake": {}, "volunteer": {}, "vow": {}, "want": {}, "wish": {},
}

// Precompiled Java Pattern / replaceAll twins.
var (
	// enUsPattern ports BeoLingusTranslator.enUsPattern
	enUsPattern = regexp.MustCompile(`.*?\w+ \[(Br|Am)\.\] ?/ ?\w+ \[(Br|Am)\.].*`)
	// replaceFirst variants for Am/Br split
	enUsVariant1 = regexp.MustCompile(`^(.*?)(\w+) (\[(?:Br|Am)\.\]) ?/ ?\w+ \[(?:Br|Am)\.\](.*)`)
	enUsVariant2 = regexp.MustCompile(`^(.*?)(\w+) (\[(?:Br|Am)\.\]) ?/ ?(\w+) (\[(?:Br|Am)\.\])(.*)`)

	reSlashAbbrev     = regexp.MustCompile(`/.*?/`)          // handleItem: replaceAll("/.*?/", "")
	reSemicolonSplit  = regexp.MustCompile(`;\s+`)
	reBraces          = regexp.MustCompile(`\{.*?\}`)
	reBrackets        = regexp.MustCompile(`\[.*?\]`)
	reParens          = regexp.MustCompile(`\(.*?\)`)
	reSlashAbbrevWord = regexp.MustCompile(`/.*?/\b`)        // cleanForLookup: replaceAll("/.*?/\\b", "")
	reAbbrevUpper     = regexp.MustCompile(`/[A-Z]+/`)        // cleanTranslationForReplace
	reAbbrevTitle     = regexp.MustCompile(` /[A-Z][a-z]+\.?/`)
	reAngle           = regexp.MustCompile(`<(.*)>`)
	reSpaces          = regexp.MustCompile(`\s+`)
)

var (
	beoInstanceMu sync.Mutex
	beoInstance   *BeoLingusTranslator
)

// GetInstanceFromFile ports BeoLingusTranslator.getInstance(File).
func GetInstanceFromFile(beolingusFile string) (*BeoLingusTranslator, error) {
	cfg := &languagetool.GlobalConfig{}
	cfg.SetBeolingusFile(beolingusFile)
	return GetInstance(cfg)
}

// GetInstance ports BeoLingusTranslator.getInstance(GlobalConfig).
func GetInstance(globalConfig *languagetool.GlobalConfig) (*BeoLingusTranslator, error) {
	beoInstanceMu.Lock()
	defer beoInstanceMu.Unlock()
	if beoInstance == nil && globalConfig != nil && globalConfig.GetBeolingusFile() != "" {
		inst, err := NewBeoLingusTranslator(globalConfig.GetBeolingusFile())
		if err != nil {
			return nil, err
		}
		beoInstance = inst
	}
	return beoInstance, nil
}

// ResetBeoLingusInstanceForTest clears the process singleton (tests only).
func ResetBeoLingusInstanceForTest() {
	beoInstanceMu.Lock()
	beoInstance = nil
	beoInstanceMu.Unlock()
}

// NewBeoLingusTranslator ports BeoLingusTranslator(File).
func NewBeoLingusTranslator(file string) (*BeoLingusTranslator, error) {
	// Java: tagger = Languages.getLanguageForShortCode("de").getTagger();
	tagger := rulesde.DiscoveredGermanTagger()
	b := &BeoLingusTranslator{
		tagger:    tagger,
		de2en:     make(map[string][]base.TranslationEntry),
		en2de:     make(map[string][]base.TranslationEntry),
		inflector: NewInflector(),
	}
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	// Beolingus lines can be long
	sc.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	for sc.Scan() {
		line := sc.Text()
		// Java: line.trim().isEmpty() || line.startsWith("#")
		if tools.JavaStringTrim(line) == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Java: line.split(" :: ")
		parts := strings.Split(line, " :: ")
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid line format (missing :: ): %s", line)
		}
		german := parts[0]
		english := parts[1]
		germanParts := strings.Split(german, "|")
		englishParts := strings.Split(english, "|")
		if len(germanParts) != len(englishParts) {
			// Java: throw new IOException("Invalid line format (DE item count != EN item count): " + line);
			return nil, fmt.Errorf("Invalid line format (DE item count != EN item count): %s", line)
		}
		for i, germanPart := range germanParts {
			b.handleItem(b.de2en, germanParts, englishParts, i, germanPart)
			// handleItem(en2de, ...) — direction not supported yet
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return b, nil
}

// handleItem ports BeoLingusTranslator.handleItem.
func (b *BeoLingusTranslator) handleItem(m map[string][]base.TranslationEntry, germanParts, englishParts []string, i int, germanPart string) {
	// Java: germanPart = germanPart.replaceAll("/.*?/", "");
	germanPart = reSlashAbbrev.ReplaceAllString(germanPart, "")
	germanSubParts := b.Split(germanPart)
	for _, germanSubPart := range germanSubParts {
		key := cleanForLookup(germanSubPart)
		oldEntries, ok := m[key]
		if ok {
			if tools.JavaStringTrim(englishParts[i]) != "" {
				oldEntries = append(oldEntries, base.NewTranslationEntry(
					b.Split(germanPart),
					b.Split(tools.JavaStringTrim(englishParts[i])),
					len(germanParts),
				))
				m[key] = oldEntries
			}
		} else {
			if tools.JavaStringTrim(englishParts[i]) != "" {
				// Java new branch: split(englishParts[i]) without trim of the whole string for empty check already done
				l := []base.TranslationEntry{
					base.NewTranslationEntry(
						b.Split(germanPart),
						b.Split(englishParts[i]),
						len(germanParts),
					),
				}
				m[key] = l
			}
		}
	}
}

// Split ports BeoLingusTranslator.split.
// "tyre [Br.]/tire [Am.] pump" -> "tyre pump [Br.]" + "tire pump [Am.]"
func (b *BeoLingusTranslator) Split(s string) []string {
	parts := b.SplitAtSemicolon(s)
	newParts := make([]string, 0, len(parts))
	for _, part := range parts {
		// Java: enUsPattern.matcher(part).matches() — full match
		if enUsPattern.MatchString(part) {
			// Java replaceFirst
			variant1 := enUsVariant1.ReplaceAllString(part, "$1$2$4 $3")
			variant2 := enUsVariant2.ReplaceAllString(part, "$1$4$6 $5")
			newParts = append(newParts, variant1, variant2)
		} else {
			newParts = append(newParts, part)
		}
	}
	return newParts
}

// GetDeEnSize ports BeoLingusTranslator.getDeEnSize.
func (b *BeoLingusTranslator) GetDeEnSize() int {
	if b == nil {
		return 0
	}
	return len(b.de2en)
}

// SplitAtSemicolon ports BeoLingusTranslator.splitAtSemicolon.
// split input like "family doctors; family physicians" at ";", unless it's in "{...}":
func (b *BeoLingusTranslator) SplitAtSemicolon(s string) []string {
	// Java: Arrays.stream(s.split(";\\s+")).map(k -> k.trim())
	raw := reSemicolonSplit.Split(s, -1)
	list := make([]string, len(raw))
	for i, k := range raw {
		list[i] = tools.JavaStringTrim(k)
	}
	mergedList := make([]string, 0, len(list))
	mergeListPos := 0
	merging := false
	for _, item := range list {
		openPos := strings.IndexByte(item, '{')
		closePos := strings.IndexByte(item, '}')
		if merging {
			mergedList[mergeListPos-1] = mergedList[mergeListPos-1] + "; " + item
			mergeListPos--
			if closePos >= 0 {
				merging = false
			}
		} else if openPos > closePos {
			// ";" inside "{...}" - merge those again
			mergedList = append(mergedList, item)
			merging = true
		} else {
			mergedList = append(mergedList, item)
		}
		mergeListPos++
	}
	return mergedList
}

// Translate ports BeoLingusTranslator.translate.
func (b *BeoLingusTranslator) Translate(term, fromLang, toLang string) ([]base.TranslationEntry, error) {
	var m map[string][]base.TranslationEntry
	if fromLang == "de" && toLang == "en" {
		m = b.de2en
	} else if fromLang == "en" && toLang == "de" {
		m = b.en2de
	} else {
		// Java: throw new RuntimeException("Not supported: " + fromLang + " -> " + toLang);
		return nil, fmt.Errorf("Not supported: %s -> %s", fromLang, toLang)
	}
	entries := m[strings.ToLower(term)]
	// HashSet<TranslationEntry> — equals on l1+l2
	entriesSet := make(map[string]base.TranslationEntry)
	entryKey := func(e base.TranslationEntry) string {
		return strings.Join(e.L1, "\x00") + "\x01" + strings.Join(e.L2, "\x00")
	}
	if entries != nil {
		for _, e := range entries {
			entriesSet[entryKey(e)] = e
		}
	}
	translationsForBaseforms := b.getTranslationsForBaseforms(term, m)
	for _, trans := range translationsForBaseforms {
		if entries != nil {
			// Java: Optional first = entries.stream().filter(k -> k.getL1().equals(trans.getL1())).findFirst();
			// if (first.isPresent() && first.get().getL1().equals(trans.getL1())) skip; else add
			foundSameL1 := false
			for _, k := range entries {
				if slicesEqualStr(k.L1, trans.L1) {
					foundSameL1 = true
					break
				}
			}
			if !foundSameL1 {
				entriesSet[entryKey(trans)] = trans
			}
		} else {
			entriesSet[entryKey(trans)] = trans
		}
	}
	sortedList := make([]base.TranslationEntry, 0, len(entriesSet))
	for _, e := range entriesSet {
		sortedList = append(sortedList, e)
	}
	// Collections.sort by itemCount descending
	sort.SliceStable(sortedList, func(i, j int) bool {
		return sortedList[i].ItemCount > sortedList[j].ItemCount
	})
	return sortedList, nil
}

func slicesEqualStr(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// getTranslationsForBaseforms ports BeoLingusTranslator.getTranslationsForBaseforms.
func (b *BeoLingusTranslator) getTranslationsForBaseforms(term string, m map[string][]base.TranslationEntry) []base.TranslationEntry {
	result := make([]base.TranslationEntry, 0)
	if b.tagger == nil {
		// Java always has de tagger; fail-closed empty when unavailable.
		return result
	}
	// Java: tagger.tag(Collections.singletonList(term));
	readings := b.tagger.Tag([]string{term})
	// readings.addAll(tagger.tag(... uppercaseFirstChar(term) ...));
	readings = append(readings, b.tagger.Tag([]string{tools.UppercaseFirstChar(term)})...)
	for _, reading := range readings {
		if reading == nil {
			continue
		}
		b.addResultsForTokens(m, reading.GetReadings(), &result)
	}
	return result
}

// addResultsForTokens ports BeoLingusTranslator.addResultsForTokens.
func (b *BeoLingusTranslator) addResultsForTokens(m map[string][]base.TranslationEntry, aTokens []*languagetool.AnalyzedToken, result *[]base.TranslationEntry) {
	for _, aToken := range aTokens {
		if aToken == nil {
			continue
		}
		lemma := aToken.GetLemma()
		if lemma != nil {
			tmp := m[strings.ToLower(*lemma)]
			if tmp != nil {
				for _, tmpEntry := range tmp {
					// Java: if (!result.contains(tmpEntry)) — equals on l1+l2 of uncleaned entry
					if !containsEntry(*result, tmpEntry) {
						entry := b.cleanTranslationEntry(tmpEntry, aToken)
						if entry != nil {
							*result = append(*result, *entry)
						}
					}
				}
			}
		}
	}
}

func containsEntry(list []base.TranslationEntry, e base.TranslationEntry) bool {
	for _, x := range list {
		if x.Equal(e) {
			return true
		}
	}
	return false
}

// cleanTranslationEntry ports BeoLingusTranslator.cleanTranslationEntry.
func (b *BeoLingusTranslator) cleanTranslationEntry(tmpEntry base.TranslationEntry, aToken *languagetool.AnalyzedToken) *base.TranslationEntry {
	l := make([]string, 0)
	// Java: aToken.getPOSTag() may be null → Inflector null branch.
	// Go Inflector: empty string is null stand-in.
	posTag := ""
	if p := aToken.GetPOSTag(); p != nil {
		posTag = *p
	}
	for _, s := range tmpEntry.GetL2() {
		// Java: inflector.inflect(cleanTranslationForReplace(s, null), aToken.getPOSTag());
		inflected := b.inflector.Inflect(b.CleanTranslationForReplace(s, ""), posTag)
		for _, inflectedForm := range inflected {
			cleanEntry := inflectedForm + " " + b.GetTranslationSuffix(s)
			l = append(l, tools.JavaStringTrim(cleanEntry))
		}
	}
	if len(l) > 0 {
		e := base.NewTranslationEntry(tmpEntry.GetL1(), l, tmpEntry.GetItemCount())
		return &e
	}
	return nil
}

// cleanForLookup ports BeoLingusTranslator.cleanForLookup.
func cleanForLookup(s string) string {
	s = reBraces.ReplaceAllString(s, "")
	s = reBrackets.ReplaceAllString(s, "")
	s = reParens.ReplaceAllString(s, "")
	// Java: replaceAll("/.*?/\\b", "")
	s = reSlashAbbrevWord.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "jdn./etw. ", "")
	s = strings.ReplaceAll(s, "jdm./etw. ", "")
	s = strings.ReplaceAll(s, "etw./jdn. ", "")
	s = strings.ReplaceAll(s, "etw./jdm. ", "")
	s = strings.ReplaceAll(s, "etw. ", "")
	s = strings.ReplaceAll(s, "jdn. ", "")
	s = strings.ReplaceAll(s, "jdm. ", "")
	return strings.ToLower(tools.JavaStringTrim(s))
}

// CleanTranslationForReplace ports BeoLingusTranslator.cleanTranslationForReplace.
// prevWord empty string maps Java null for callers that pass null.
func (b *BeoLingusTranslator) CleanTranslationForReplace(s, prevWord string) string {
	clean := reBrackets.ReplaceAllString(s, "")
	clean = reBraces.ReplaceAllString(clean, "")
	clean = reParens.ReplaceAllString(clean, "")
	clean = strings.ReplaceAll(clean, "sth./sb.", "")
	clean = strings.ReplaceAll(clean, "sb./sth.", "")
	clean = strings.ReplaceAll(clean, "sth.", "")
	clean = strings.ReplaceAll(clean, "sb.", "")
	clean = reAbbrevUpper.ReplaceAllString(clean, "")
	clean = reAbbrevTitle.ReplaceAllString(clean, "")
	clean = strings.ReplaceAll(clean, "<> ", "")
	clean = reAngle.ReplaceAllString(clean, "")
	clean = reSpaces.ReplaceAllString(clean, " ")
	clean = tools.JavaStringTrim(clean)
	if prevWord == "to" && strings.HasPrefix(clean, "to ") {
		return clean[3:]
	}
	// Java: if (!"to".equals(prevWord) && clean.startsWith("to ") && !verbsWithTo.contains(prevWord))
	// null prevWord: !"to".equals(null) is true; verbsWithTo.contains(null) is false → strip "to "
	if prevWord != "to" && strings.HasPrefix(clean, "to ") {
		if _, ok := verbsWithTo[prevWord]; !ok {
			return clean[3:]
		}
	}
	return clean
}

// GetMessage ports BeoLingusTranslator.getMessage.
func (b *BeoLingusTranslator) GetMessage() string {
	return "Translate to English?"
}

// GetTranslationSuffix ports BeoLingusTranslator.getTranslationSuffix.
func (b *BeoLingusTranslator) GetTranslationSuffix(s string) string {
	var sb strings.Builder
	lookingFor := make([]string, 0)
	contains := func(c string) bool {
		for _, x := range lookingFor {
			if x == c {
				return true
			}
		}
		return false
	}
	remove := func(c string) {
		// Java List.remove(Object) removes first occurrence
		out := lookingFor[:0]
		removed := false
		for _, x := range lookingFor {
			if !removed && x == c {
				removed = true
				continue
			}
			out = append(out, x)
		}
		lookingFor = out
	}
	// Java iterates s.charAt (UTF-16). Suffix delimiters are ASCII; range is fine for BMP test data.
	for _, r := range s {
		switch r {
		case '[':
			lookingFor = append(lookingFor, "]")
		case ']':
			if contains("]") {
				sb.WriteRune(r)
				sb.WriteByte(' ')
				remove("]")
			}
		case '<':
			lookingFor = append(lookingFor, ">")
		case '>':
			if contains(">") {
				sb.WriteRune(r)
				sb.WriteByte(' ')
				remove(">")
			}
		case '(':
			lookingFor = append(lookingFor, ")")
		case ')':
			// Java: no contains check
			sb.WriteRune(r)
			sb.WriteByte(' ')
			remove(")")
		case '{':
			lookingFor = append(lookingFor, "}")
		case '}':
			// Java: no contains check
			sb.WriteRune(r)
			sb.WriteByte(' ')
			remove("}")
		}
		if len(lookingFor) > 0 {
			sb.WriteRune(r)
		}
	}
	return tools.JavaStringTrim(sb.String())
}

// GetDataSource ports BeoLingusTranslator.getDataSource.
func (b *BeoLingusTranslator) GetDataSource() base.DataSource {
	return base.NewDataSource(
		"https://www.gnu.org/licenses/old-licenses/gpl-2.0.en.html",
		"BEOLINGUS",
		"http://dict.tu-chemnitz.de",
	)
}

// DiscoverBeolingusTestFile finds official beolingus_test.txt (Java src/test/resources).
func DiscoverBeolingusTestFile() string {
	relPaths := []string{
		filepath.Join("inspiration", "languagetool", "languagetool-standalone", "src", "test", "resources", "beolingus_test.txt"),
	}
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for {
		for _, rel := range relPaths {
			cand := filepath.Join(dir, rel)
			if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
				return cand
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

var _ base.Translator = (*BeoLingusTranslator)(nil)
