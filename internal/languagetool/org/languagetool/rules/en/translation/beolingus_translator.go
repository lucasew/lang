package translation

import (
	"bufio"
	"io"
	"strings"

	base "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/translation"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// BeoLingusTranslator ports org.languagetool.rules.en.translation.BeoLingusTranslator
// as a line-oriented dictionary (de→en pairs) with Inflector post-processing.
type BeoLingusTranslator struct {
	*base.MapTranslator
	Inflector *Inflector
}

func NewBeoLingusTranslator() *BeoLingusTranslator {
	src := base.NewDataSource(
		"https://www.dict.cc/?s=about%3A",
		"BEOLINGUS",
		"https://dict.tu-chemnitz.de/",
	)
	mt := base.NewMapTranslator(src)
	mt.Message = "Possible translation from German:"
	return &BeoLingusTranslator{MapTranslator: mt}
}

// LoadDict loads "german :: english", "de|en", or tab-separated lines.
func (b *BeoLingusTranslator) LoadDict(r io.Reader) error {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := tools.JavaStringTrim(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		var de, en string
		switch {
		case strings.Contains(line, " :: "):
			parts := strings.SplitN(line, " :: ", 2)
			de, en = parts[0], parts[1]
		case strings.Contains(line, "|"):
			parts := strings.SplitN(line, "|", 2)
			de, en = parts[0], parts[1]
		case strings.Contains(line, "\t"):
			parts := strings.SplitN(line, "\t", 2)
			de, en = parts[0], parts[1]
		default:
			continue
		}
		de, en = tools.JavaStringTrim(de), tools.JavaStringTrim(en)
		if de == "" || en == "" {
			continue
		}
		key := stripBraces(de)
		entry := base.NewTranslationEntry([]string{de}, []string{en}, 1)
		b.Add(key, "de", "en", entry)
	}
	return sc.Err()
}

// TranslateWithInflection looks up and optionally inflects English forms via DE POS.
func (b *BeoLingusTranslator) TranslateWithInflection(term, dePosTag string) ([]base.TranslationEntry, error) {
	entries, err := b.Translate(term, "de", "en")
	if err != nil || b.Inflector == nil || dePosTag == "" {
		return entries, err
	}
	var out []base.TranslationEntry
	for _, e := range entries {
		var l2 []string
		for _, en := range e.L2 {
			l2 = append(l2, b.Inflector.Inflect(en, dePosTag)...)
		}
		if len(l2) == 0 {
			l2 = e.L2
		}
		out = append(out, base.NewTranslationEntry(e.L1, l2, e.ItemCount))
	}
	return out, nil
}

func stripBraces(s string) string {
	for {
		i := strings.IndexAny(s, "{[")
		if i < 0 {
			break
		}
		closeCh := "}"
		if s[i] == '[' {
			closeCh = "]"
		}
		j := strings.Index(s[i:], closeCh)
		if j < 0 {
			break
		}
		s = s[:i] + s[i+j+1:]
	}
	return tools.JavaStringTrim(s)
}
