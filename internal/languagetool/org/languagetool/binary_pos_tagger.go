package languagetool

import (
	"strings"

	atticmorfo "github.com/lucasew/lang/internal/attic/morfologik"
)

// RegisterBinaryPOSTagger installs lt.TagWord from a Morfologik POS dictionary
// (CFSA2 or FSA5), matching Java BaseTagger/MorfologikTagger lookup behavior.
// Returns false if the dictionary cannot be opened.
func RegisterBinaryPOSTagger(lt *JLanguageTool, dictPath string) bool {
	if lt == nil || dictPath == "" {
		return false
	}
	d, err := atticmorfo.OpenDictionary(dictPath)
	if err != nil || d == nil {
		return false
	}
	lt.TagWord = BinaryPOSTagWord(d)
	return true
}

// BinaryPOSTagWord returns a TagWord inject from an opened POS dictionary.
// Case retries mirror typical LT BaseTagger lowercasing when the surface miss hits.
func BinaryPOSTagWord(d *atticmorfo.Dictionary) func(token string) []TokenTag {
	if d == nil {
		return nil
	}
	lookup := func(w string) []TokenTag {
		forms, err := d.Lookup(w)
		if err != nil || len(forms) == 0 {
			return nil
		}
		out := make([]TokenTag, 0, len(forms))
		for _, f := range forms {
			out = append(out, TokenTag{POS: f.Tag, Lemma: f.Stem})
		}
		return out
	}
	return func(token string) []TokenTag {
		if token == "" {
			return nil
		}
		if tags := lookup(token); len(tags) > 0 {
			return tags
		}
		low := strings.ToLower(token)
		if low != token {
			if tags := lookup(low); len(tags) > 0 {
				return tags
			}
		}
		if low != "" {
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
