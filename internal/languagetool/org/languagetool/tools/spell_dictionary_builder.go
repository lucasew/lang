package tools

import (
	"bufio"
	"io"
	"strings"
)

// SpellDictionaryBuilder ports org.languagetool.tools.SpellDictionaryBuilder text prep.
type SpellDictionaryBuilder struct {
	*DictionaryBuilder
}

func NewSpellDictionaryBuilder(info map[string]string) *SpellDictionaryBuilder {
	return &SpellDictionaryBuilder{DictionaryBuilder: NewDictionaryBuilder(info)}
}

// TokenizeInput copies plain-text word list, optionally attaching frequency separator.
// Each non-empty line becomes one token (Java currently treats whole line as token).
func (b *SpellDictionaryBuilder) TokenizeInput(r io.Reader, w io.Writer) (int, error) {
	sep := ""
	if b != nil {
		sep = b.Separator()
		if sep == "" && b.Props != nil {
			sep = b.Props["fsa.dict.separator"]
		}
	}
	sc := bufio.NewScanner(r)
	n := 0
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// strip existing separator payload
		token := line
		occ := ""
		if sep != "" {
			if i := strings.Index(line, sep); i >= 0 {
				token = line[:i]
				occ = line[i+len(sep):]
			}
		}
		if token == "" {
			continue
		}
		if _, err := io.WriteString(w, token); err != nil {
			return n, err
		}
		if sep != "" {
			if _, err := io.WriteString(w, sep); err != nil {
				return n, err
			}
			if occ != "" {
				if _, err := io.WriteString(w, occ); err != nil {
					return n, err
				}
			} else if len(b.FreqList) > 0 {
				if f, ok := b.FreqList[token]; ok {
					if _, err := w.Write([]byte{FreqToRange(f)}); err != nil {
						return n, err
					}
				} else {
					if _, err := w.Write([]byte{'A'}); err != nil {
						return n, err
					}
				}
			}
		}
		if _, err := io.WriteString(w, "\n"); err != nil {
			return n, err
		}
		n++
	}
	return n, sc.Err()
}

func (b *SpellDictionaryBuilder) Separator() string {
	if b == nil || b.DictionaryBuilder == nil {
		return ""
	}
	return b.DictionaryBuilder.Separator()
}
