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

// DiscoverFalseFriendsFile finds soft false-friends XML via env/data-dir/walk-up.
func DiscoverFalseFriendsFile(opts *CommandLineOptions) string {
	if p := resolveFalseFriendsFile(opts); p != "" {
		if _, err := os.Stat(p); err == nil {
			return p
		}
		if opts != nil && (opts.GetDataDir() != "" || opts.FalseFriendsFile != "") {
			return p
		}
	}
	return WalkUpFind("", filepath.Join("testdata", "false-friends-soft.xml"))
}
