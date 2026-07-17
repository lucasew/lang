package commandline

import (
	"os"
	"path/filepath"
)

// WalkUpFind walks from start (or cwd) toward root looking for relPath.
// Soft data discovery (SPEC §10 nicer data discovery).
func WalkUpFind(start, relPath string) string {
	if relPath == "" {
		return ""
	}
	dir := start
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return ""
		}
	}
	for i := 0; i < 12; i++ {
		cand := filepath.Join(dir, relPath)
		if st, err := os.Stat(cand); err == nil && (st.IsDir() || st.Mode().IsRegular()) {
			return cand
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// DiscoverGrammarDir finds a soft grammar dir via env, data-dir, or walk-up testdata/grammar.
func DiscoverGrammarDir(opts *CommandLineOptions) string {
	if d := resolveGrammarDir(opts); d != "" {
		if st, err := os.Stat(d); err == nil && st.IsDir() {
			return d
		}
		// still return configured path even if missing (caller may no-op)
		if opts != nil && opts.GetDataDir() != "" {
			return d
		}
		if os.Getenv("LANG_GRAMMAR_DIR") != "" || os.Getenv("LANG_DATA_DIR") != "" {
			return d
		}
	}
	return WalkUpFind("", filepath.Join("testdata", "grammar"))
}

// DiscoverFalseFriendsFile finds false-friends XML via env/data-dir/walk-up.
// Prefers vendored upstream (DOCTYPE-stripped) over the legacy soft subset.
func DiscoverFalseFriendsFile(opts *CommandLineOptions) string {
	if p := resolveFalseFriendsFile(opts); p != "" {
		if _, err := os.Stat(p); err == nil {
			return p
		}
		if opts != nil && (opts.GetDataDir() != "" || opts.FalseFriendsFile != "") {
			return p
		}
	}
	for _, rel := range []string{
		filepath.Join("testdata", "upstream", "false-friends-nodtd.xml"),
		filepath.Join("testdata", "upstream", "false-friends.xml"),
		filepath.Join("testdata", "false-friends-soft.xml"),
		filepath.Join("inspiration", "languagetool", "languagetool-core", "src", "main", "resources", "org", "languagetool", "rules", "false-friends.xml"),
	} {
		if p := WalkUpFind("", rel); p != "" {
			return p
		}
	}
	return ""
}

// DiscoverEnglishUSDict finds en_US.dict for the binary Morfologik speller.
// Order: LANG_EN_US_DICT, --data-dir/en/hunspell/en_US.dict, walk-up third_party and inspiration.
func DiscoverEnglishUSDict(opts *CommandLineOptions) string {
	if p := os.Getenv("LANG_EN_US_DICT"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	if opts != nil && opts.GetDataDir() != "" {
		cand := filepath.Join(opts.GetDataDir(), "en", "hunspell", "en_US.dict")
		if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
			return cand
		}
		// also flat layout
		cand = filepath.Join(opts.GetDataDir(), "en_US.dict")
		if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
			return cand
		}
	}
	relPaths := []string{
		filepath.Join("third_party", "english-pos-dict", "org", "languagetool", "resource", "en", "hunspell", "en_US.dict"),
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "en", "src", "main", "resources", "org", "languagetool", "resource", "en", "hunspell", "en_US.dict"),
	}
	for _, rel := range relPaths {
		if p := WalkUpFind("", rel); p != "" {
			return p
		}
	}
	return ""
}

// DiscoverEnglishTyposFile finds soft EN typo→suggestion TSV
// (LANG_EN_TYPOS_FILE, data-dir/spelling/en-typos.tsv, walk-up testdata).
func DiscoverEnglishTyposFile(opts *CommandLineOptions) string {
	if p := os.Getenv("LANG_EN_TYPOS_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	if opts != nil && opts.GetDataDir() != "" {
		for _, rel := range []string{
			filepath.Join(opts.GetDataDir(), "spelling", "en-typos.tsv"),
			filepath.Join(opts.GetDataDir(), "en-typos.tsv"),
		} {
			if st, err := os.Stat(rel); err == nil && st.Mode().IsRegular() {
				return rel
			}
		}
	}
	return WalkUpFind("", filepath.Join("testdata", "spelling", "en-typos.tsv"))
}

// DiscoverEnglishIgnoreSpellingList finds soft EN ignore-spelling word list
// (CLI --ignore-spelling-file, LANG_IGNORE_SPELLING_FILE, data-dir, walk-up).
func DiscoverEnglishIgnoreSpellingList(opts *CommandLineOptions) string {
	if opts != nil {
		if p := opts.GetIgnoreSpellingFile(); p != "" {
			if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
				return p
			}
			// still return configured path for diagnostics even if missing
			return p
		}
	}
	if p := os.Getenv("LANG_IGNORE_SPELLING_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	if opts != nil && opts.GetDataDir() != "" {
		for _, rel := range []string{
			filepath.Join(opts.GetDataDir(), "disambiguation", "en-ignore-spelling.txt"),
			filepath.Join(opts.GetDataDir(), "en-ignore-spelling.txt"),
		} {
			if st, err := os.Stat(rel); err == nil && st.Mode().IsRegular() {
				return rel
			}
		}
	}
	return WalkUpFind("", filepath.Join("testdata", "disambiguation", "en-ignore-spelling.txt"))
}

// DiscoverEnglishSoftDisambiguationXML finds soft EN disambiguation XML
// (CLI --disambiguation-file, LANG_DISAMBIGUATION_FILE, data-dir, walk-up).
func DiscoverEnglishSoftDisambiguationXML(opts *CommandLineOptions) string {
	if opts != nil {
		if p := opts.GetDisambiguationFile(); p != "" {
			if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
				return p
			}
			return p
		}
	}
	if p := os.Getenv("LANG_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	if opts != nil && opts.GetDataDir() != "" {
		for _, rel := range []string{
			filepath.Join(opts.GetDataDir(), "disambiguation", "en-soft.xml"),
			filepath.Join(opts.GetDataDir(), "en-soft-disambiguation.xml"),
		} {
			if st, err := os.Stat(rel); err == nil && st.Mode().IsRegular() {
				return rel
			}
		}
	}
	return WalkUpFind("", filepath.Join("testdata", "disambiguation", "en-soft.xml"))
}

// DiscoverEnglishMultiwords finds multiword dict for soft multiword disambiguation.
// Prefers LANG_EN_MULTIWORDS, data-dir, vendored upstream multiwords, then legacy soft list.
func DiscoverEnglishMultiwords(opts *CommandLineOptions) string {
	if p := os.Getenv("LANG_EN_MULTIWORDS"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	if opts != nil && opts.GetDataDir() != "" {
		for _, rel := range []string{
			filepath.Join(opts.GetDataDir(), "disambiguation", "en-multiwords-soft.txt"),
			filepath.Join(opts.GetDataDir(), "en", "multiwords.txt"),
			filepath.Join(opts.GetDataDir(), "multiwords.txt"),
		} {
			if st, err := os.Stat(rel); err == nil && st.Mode().IsRegular() {
				return rel
			}
		}
	}
	// Prefer vendored upstream multiwords, then legacy soft list, then submodule/third_party.
	relPaths := []string{
		filepath.Join("testdata", "disambiguation", "en-multiwords-upstream.txt"),
		filepath.Join("testdata", "upstream", "en", "resource", "multiwords.txt"),
		filepath.Join("testdata", "disambiguation", "en-multiwords-soft.txt"),
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "en", "src", "main", "resources", "org", "languagetool", "resource", "en", "multiwords.txt"),
		filepath.Join("third_party", "english-pos-dict", "org", "languagetool", "resource", "en", "multiwords.txt"),
	}
	for _, rel := range relPaths {
		if p := WalkUpFind("", rel); p != "" {
			return p
		}
	}
	return ""
}

// DiscoverEnglishPOSDict finds english.dict for the binary POS tagger.
// Order: LANG_ENGLISH_DICT, --data-dir/en/english.dict, walk-up third_party and inspiration.
func DiscoverEnglishPOSDict(opts *CommandLineOptions) string {
	if p := os.Getenv("LANG_ENGLISH_DICT"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	if opts != nil && opts.GetDataDir() != "" {
		for _, rel := range []string{
			filepath.Join(opts.GetDataDir(), "en", "english.dict"),
			filepath.Join(opts.GetDataDir(), "english.dict"),
		} {
			if st, err := os.Stat(rel); err == nil && st.Mode().IsRegular() {
				return rel
			}
		}
	}
	relPaths := []string{
		filepath.Join("third_party", "english-pos-dict", "org", "languagetool", "resource", "en", "english.dict"),
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "en", "src", "main", "resources", "org", "languagetool", "resource", "en", "english.dict"),
	}
	for _, rel := range relPaths {
		if p := WalkUpFind("", rel); p != "" {
			return p
		}
	}
	return ""
}
