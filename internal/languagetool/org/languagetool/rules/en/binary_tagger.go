package en

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	atticmorfo "github.com/lucasew/lang/internal/attic/morfologik"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	entok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/en"
)

// enPunctPCTRE ports EN disambiguation UNKNOWN_PCT: [\.,;:…!\?] → add POS PCT.
var enPunctPCTRE = regexp.MustCompile(`^[\.,;:…!\?]+$`)

// RegisterBinaryEnglishTagger installs lt.TagWord backed by CFSA2 english.dict POS
// lookup plus Java BaseTagger manual added/removed files (EnglishTagger extends BaseTagger).
// Returns false if the dictionary cannot be opened.
func RegisterBinaryEnglishTagger(lt *languagetool.JLanguageTool, dictPath string) bool {
	if lt == nil || dictPath == "" {
		return false
	}
	d, err := atticmorfo.OpenDictionary(dictPath)
	if err != nil || d == nil {
		return false
	}
	wt := englishWordTaggerFromDict(d, dictPath)
	tw := BinaryEnglishTagWordFrom(wt)
	lt.TagWord = tw
	// MatchState suppress_misspelled (Java lang.getTagger().tag on synthesized forms).
	patterns.RegisterLanguageTagger("en", tw)
	if lt.LanguageCode != "" {
		patterns.RegisterLanguageTagger(lt.LanguageCode, tw)
	}
	// Java EnglishWordTokenizer uses EnglishTagger.INSTANCE.tag(...).isTagged().
	wireEnglishTokenizerIsTagged(tw)
	return true
}

// englishWordTaggerFromDict builds Java EnglishTagger.getWordTagger():
// CombiningTagger(Morfologik, Manual(added*), Manual(removed*), overwrite=false)
// when added.txt is present (BaseTagger.initWordTagger).
func englishWordTaggerFromDict(d *atticmorfo.Dictionary, dictPath string) tagging.WordTagger {
	var wt tagging.WordTagger = morfologikEnglishWordTagger{d: d}
	manual := loadEnglishManualTagger(dictPath, []string{"added.txt", "added_custom.txt"})
	if manual == nil {
		return wt
	}
	removal := loadEnglishManualTagger(dictPath, []string{"removed.txt", "removed_custom.txt"})
	return tagging.NewCombiningTaggerWithRemoval(wt, manual, removal, false)
}

// loadEnglishManualTagger finds official EN manual files: beside english.dict first,
// then inspiration resource/en (dict often lives in third_party without manuals).
func loadEnglishManualTagger(dictPath string, names []string) tagging.WordTagger {
	if wt := languagetool.LoadManualTaggerBesideDict(dictPath, names); wt != nil {
		return wt
	}
	var dirs []string
	if dictPath != "" {
		dirs = append(dirs, filepath.Dir(dictPath))
	}
	// Walk-up official inspiration EN resource (Java classpath resource/en/).
	if p := walkUpEnglishResourceDir(); p != "" {
		dirs = append(dirs, p)
	}
	return languagetool.LoadManualTaggerFromDirs(dirs, names)
}

func walkUpEnglishResourceDir() string {
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "en",
		"src", "main", "resources", "org", "languagetool", "resource", "en")
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for i := 0; i < 12; i++ {
		p := filepath.Join(dir, rel)
		if st, err := os.Stat(p); err == nil && st.IsDir() {
			return p
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

type morfologikEnglishWordTagger struct {
	d *atticmorfo.Dictionary
}

func (w morfologikEnglishWordTagger) Tag(word string) []tagging.TaggedWord {
	if w.d == nil || word == "" {
		return nil
	}
	forms, err := w.d.Lookup(word)
	if err != nil || len(forms) == 0 {
		return nil
	}
	out := make([]tagging.TaggedWord, 0, len(forms))
	for _, f := range forms {
		out = append(out, tagging.NewTaggedWord(f.Stem, f.Tag))
	}
	return out
}

// wireEnglishTokenizerIsTagged installs EnglishWordTokenizer.IsTaggedEN from a
// BinaryEnglishTagWord (Java EnglishTagger.INSTANCE for wordsToAdd).
func wireEnglishTokenizerIsTagged(tw func(token string) []languagetool.TokenTag) {
	if tw == nil {
		entok.IsTaggedEN = nil
		return
	}
	entok.IsTaggedEN = func(s string) bool {
		for _, t := range tw(s) {
			if t.POS != "" {
				return true
			}
		}
		return false
	}
}

// BinaryEnglishTagWord returns a TagWord inject from an opened POS dictionary only
// (no manuals). Prefer RegisterBinaryEnglishTagger / BinaryEnglishTagWordFrom for engine.
func BinaryEnglishTagWord(d *atticmorfo.Dictionary) func(token string) []languagetool.TokenTag {
	if d == nil {
		return nil
	}
	return BinaryEnglishTagWordFrom(morfologikEnglishWordTagger{d: d})
}

// BinaryEnglishTagWordFrom ports EnglishTagger.tag case/apostrophe logic over a WordTagger
// (Morfologik ± CombiningTagger manuals, same as Java getWordTagger()).
func BinaryEnglishTagWordFrom(wt tagging.WordTagger) func(token string) []languagetool.TokenTag {
	if wt == nil {
		return nil
	}
	lookup := func(w string) []languagetool.TokenTag {
		tws := wt.Tag(w)
		if len(tws) == 0 {
			return nil
		}
		out := make([]languagetool.TokenTag, 0, len(tws))
		for _, tw := range tws {
			out = append(out, languagetool.TokenTag{POS: tw.PosTag, Lemma: tw.Lemma})
		}
		return out
	}
	return func(token string) []languagetool.TokenTag {
		if token == "" {
			return nil
		}
		// Java: typewriter apostrophe so dict entries match.
		word := strings.ReplaceAll(token, "’", "'")
		low := strings.ToLower(word)
		isLower := word == low
		isMixed := englishIsMixedCase(word)
		isAllUpper := word != "" && word == strings.ToUpper(word) && hasLetterEN(word)

		var out []languagetool.TokenTag
		seen := map[string]struct{}{}
		add := func(tags []languagetool.TokenTag) {
			for _, t := range tags {
				key := t.POS + "\x00" + t.Lemma
				if _, ok := seen[key]; ok {
					continue
				}
				seen[key] = struct{}{}
				out = append(out, t)
			}
		}
		// normal case
		add(lookup(word))
		// tag non-lowercase (alluppercase or startuppercase), but not mixed-case,
		// with lowercase word tags (Java EnglishTagger)
		if !isLower && !isMixed {
			add(lookup(low))
		}
		// tag all-uppercase proper nouns (ex. FRANCE) via Title case of lower
		if len(out) == 0 && isAllUpper && low != "" {
			runes := []rune(low)
			title := strings.ToUpper(string(runes[0])) + string(runes[1:])
			if title != word {
				add(lookup(title))
			}
		}
		// walkin' → walking style (Java endsWith "in'")
		if len(out) == 0 && strings.HasSuffix(low, "in'") {
			corrected := word
			if isAllUpper {
				corrected = word[:len(word)-1] + "G"
			} else {
				corrected = word[:len(word)-1] + "g"
			}
			add(lookup(corrected))
			if !isLower && !isMixed {
				add(lookup(strings.ToLower(corrected)))
			}
		}
		// Java disambiguation UNKNOWN_PCT: add PCT on .,;:…!? so grammar
		// patterns postag="…|PCT" match commas (ALL_OF_SUDDEN, etc.).
		if enPunctPCTRE.MatchString(word) {
			add([]languagetool.TokenTag{{POS: "PCT", Lemma: word}})
		}
		return out
	}
}

func hasLetterEN(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

// englishIsMixedCase ports StringTools.isMixedCase: both upper and lower letters,
// not merely initial capital (Title case is not mixed).
func englishIsMixedCase(s string) bool {
	hasUpper, hasLower := false, false
	first := true
	for _, r := range s {
		if !unicode.IsLetter(r) {
			continue
		}
		if unicode.IsUpper(r) {
			if !first {
				// upper after first letter → mixed (iPhone) or ALL_CAPS handled elsewhere
				hasUpper = true
			} else {
				hasUpper = true
			}
		}
		if unicode.IsLower(r) {
			hasLower = true
		}
		first = false
	}
	if !hasUpper || !hasLower {
		return false
	}
	// Title case (first upper, rest lower) is not mixed in LT.
	rs := []rune(s)
	// skip non-letters at start
	i := 0
	for i < len(rs) && !unicode.IsLetter(rs[i]) {
		i++
	}
	if i >= len(rs) {
		return false
	}
	if !unicode.IsUpper(rs[i]) {
		return true // lower then upper somewhere → mixed
	}
	for _, r := range rs[i+1:] {
		if unicode.IsLetter(r) && unicode.IsUpper(r) {
			return true // e.g. iPhone or McDonald - has upper after first
		}
	}
	return false // Title case only
}
