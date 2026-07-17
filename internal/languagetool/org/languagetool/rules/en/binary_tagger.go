package en

import (
	"strings"

	atticmorfo "github.com/lucasew/lang/internal/attic/morfologik"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// RegisterBinaryEnglishTagger installs lt.TagWord backed by CFSA2 english.dict POS lookup.
// Returns false if the dictionary cannot be opened.
func RegisterBinaryEnglishTagger(lt *languagetool.JLanguageTool, dictPath string) bool {
	if lt == nil || dictPath == "" {
		return false
	}
	d, err := atticmorfo.OpenDictionary(dictPath)
	if err != nil || d == nil {
		return false
	}
	lt.TagWord = BinaryEnglishTagWord(d)
	return true
}

// BinaryEnglishTagWord returns a TagWord inject from an opened POS dictionary.
func BinaryEnglishTagWord(d *atticmorfo.Dictionary) func(token string) []languagetool.TokenTag {
	if d == nil {
		return nil
	}
	lookup := func(w string) []languagetool.TokenTag {
		forms, err := d.Lookup(w)
		if err != nil || len(forms) == 0 {
			return nil
		}
		out := make([]languagetool.TokenTag, 0, len(forms))
		for _, f := range forms {
			out = append(out, languagetool.TokenTag{POS: f.Tag, Lemma: f.Stem})
		}
		return out
	}
	return func(token string) []languagetool.TokenTag {
		if token == "" {
			return nil
		}
		if tags := lookup(token); len(tags) > 0 {
			return tags
		}
		// case retries: lower, then Title
		low := strings.ToLower(token)
		if low != token {
			if tags := lookup(low); len(tags) > 0 {
				return tags
			}
		}
		if low != "" {
			// Title case: first upper rest lower (This → This already tried; house → House rare)
			runes := []rune(low)
			if len(runes) > 0 {
				title := strings.ToUpper(string(runes[0])) + string(runes[1:])
				if title != token && title != low {
					if tags := lookup(title); len(tags) > 0 {
						return tags
					}
				}
			}
		}
		return nil
	}
}
