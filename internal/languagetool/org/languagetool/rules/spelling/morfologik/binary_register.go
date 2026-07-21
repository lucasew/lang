package morfologik

import (
	"os"
	"path/filepath"
	"strings"

	atticmorfo "github.com/lucasew/lang/internal/attic/morfologik"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
)

// DiscoverLanguageDict finds a Java resource-path Morfologik speller .dict on disk.
// classpath is Java getFileName(), e.g. "/pl/hunspell/pl_PL.dict".
// Walks inspiration language-modules and optional third_party mirrors. Empty if missing
// (fail closed — do not invent a different locale dict).
func DiscoverLanguageDict(classpath string) string {
	rel := strings.TrimPrefix(classpath, "/")
	if rel == "" || !strings.HasSuffix(rel, ".dict") {
		return ""
	}
	// first path segment is language short code (pl, de, …)
	lang, rest, ok := strings.Cut(rel, "/")
	if !ok || lang == "" || rest == "" {
		return ""
	}
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for {
		candidates := []string{
			filepath.Join(dir, "inspiration", "languagetool", "languagetool-language-modules", lang,
				"src", "main", "resources", "org", "languagetool", "resource", lang, rest),
			// some modules put hunspell under resource/ without repeating lang in rest already including lang/
			filepath.Join(dir, "inspiration", "languagetool", "languagetool-language-modules", lang,
				"src", "main", "resources", "org", "languagetool", "resource", rel),
			filepath.Join(dir, "third_party", "languagetool-dicts", "org", "languagetool", "resource", rel),
			filepath.Join(dir, "third_party", "english-pos-dict", "org", "languagetool", "resource", rel),
		}
		for _, p := range candidates {
			if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
				return p
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

// TryRegisterBinarySpeller opens a CFSA2/FSA speller dict and registers ruleID
// via SimplePredicateSpellerChecker (Java MorfologikSpellerRule Match parity for
// isMisspelled + SuggestEdits). Returns false if path empty or dict cannot open.
func TryRegisterBinarySpeller(lt *languagetool.JLanguageTool, ruleID, classpathOrPath string) bool {
	if lt == nil || ruleID == "" || classpathOrPath == "" {
		return false
	}
	dictPath := classpathOrPath
	if strings.HasPrefix(classpathOrPath, "/") || !filepath.IsAbs(classpathOrPath) {
		// Java classpath form or relative — discover when not an existing file
		if st, err := os.Stat(classpathOrPath); err != nil || !st.Mode().IsRegular() {
			dictPath = DiscoverLanguageDict(classpathOrPath)
		}
	}
	if dictPath == "" {
		return false
	}
	d, err := atticmorfo.OpenDictionary(dictPath)
	if err != nil || d == nil {
		return false
	}
	// Java SpellingCheckRule.init word lists for language from classpath (/pl/hunspell/…).
	langCode := ""
	rel := strings.TrimPrefix(classpathOrPath, "/")
	if i := strings.IndexByte(rel, '/'); i > 0 {
		langCode = rel[:i]
	}
	meta := spelling.NewSpellingCheckRule(ruleID, "Possible spelling mistake", langCode)
	spelling.ApplyDefaultSpellingWordLists(meta)
	// MorfologikSpeller with binary FSA + .info flags (Java Speller.isMisspelled gates).
	msp := NewMorfologikSpeller(classpathOrPath, 1)
	if !msp.AttachBinaryDictionary(dictPath) {
		// Open already succeeded above; Attach should use same path.
		msp.InDictionaryFn = d.Contains
		msp.BinaryDictPath = dictPath
		msp.LoadInfoBesideDict(dictPath)
	}
	isKnown := func(w string) bool {
		if meta.IsProhibited(w) {
			return false
		}
		if _, ok := meta.IgnoreWords[w]; ok {
			return true
		}
		// Java: !speller.isMisspelled(word)
		return !msp.IsMisspelled(w)
	}
	suggestFn := func(w string) []string {
		raw := d.SuggestEdits(w, 8)
		if len(meta.ProhibitedWords) == 0 {
			return raw
		}
		out := make([]string, 0, len(raw))
		for _, s := range raw {
			if !meta.IsProhibited(s) {
				out = append(out, s)
			}
		}
		return out
	}
	// Wrap so multi-token IGNORE_SPELLING phrases mark tokens before spellcheck.
	inner := languagetool.SimplePredicateSpellerChecker(
		ruleID, isKnown, map[string][]string{}, nil, suggestFn,
	)
	lt.AddRuleChecker(ruleID, func(s *languagetool.AnalyzedSentence) []languagetool.LocalMatch {
		meta.MarkMultiWordIgnoreSpelling(s)
		return inner(s)
	})
	return true
}

// RegisterSpellerOrEmpty registers a binary CFSA2 speller when the Java resource
// dict is on disk; otherwise calls registerEmpty (map Morfologik shell / fail-closed).
func RegisterSpellerOrEmpty(lt *languagetool.JLanguageTool, ruleID, javaClasspath string, registerEmpty func()) {
	if TryRegisterBinarySpeller(lt, ruleID, javaClasspath) {
		return
	}
	if registerEmpty != nil {
		registerEmpty()
	}
}
