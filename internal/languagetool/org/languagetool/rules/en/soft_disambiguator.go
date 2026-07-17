package en

import (
	"bufio"
	"os"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
	entag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/en"
)

// softEnglishMultiwords is a small tab-separated multiword list for the live check path.
// Full multiwords.txt uses glued tags for some lines; keep this soft and safe.
var softEnglishMultiwords = []string{
	"New York\tNNP",
	"Los Angeles\tNNP",
	"United States\tNNP",
	"United Kingdom\tNNP",
	"status quo\tNN",
	"Status Quo\tNN",
	"as well\tRB",
	"Taj Mahal\tNNP",
	"Yom Kippur\tNNP",
}

// RegisterSoftEnglishDisambiguator installs an EnglishHybridDisambiguator with a
// MultiWordChunker (soft built-in lines + optional tab-separated multiwords file).
func RegisterSoftEnglishDisambiguator(lt *languagetool.JLanguageTool, multiwordsPath string) {
	if lt == nil {
		return
	}
	lines := append([]string(nil), softEnglishMultiwords...)
	if multiwordsPath != "" {
		if f, err := os.Open(multiwordsPath); err == nil {
			// Only append tab-separated lines to avoid panics on glued-tag format.
			if loaded, err := loadTabSeparatedMultiwords(f); err == nil && len(loaded) > 0 {
				lines = append(lines, loaded...)
			}
			_ = f.Close()
		}
	}
	chunker := disambiguation.NewMultiWordChunker(lines, disambiguation.MultiWordChunkerSettings{
		AllowFirstCapitalized: true,
		AllowAllUppercase:     true,
		AllowTitlecase:        true,
	})
	chunker.SetIgnoreSpelling(true)
	hyb := entag.NewEnglishHybridDisambiguator()
	hyb.Chunker = chunker
	lt.Disambiguator = hyb
}

func loadTabSeparatedMultiwords(f *os.File) ([]string, error) {
	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		// require a tab so glued "phraseTAG" lines from upstream multiwords.txt are skipped
		if !strings.Contains(line, "\t") {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			continue
		}
		lines = append(lines, parts[0]+"\t"+parts[1])
	}
	return lines, sc.Err()
}
