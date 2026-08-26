package commandline

import (
	"path/filepath"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Official Morfologik POS dicts under inspiration/languagetool language modules.
// Names must match upstream resource/{code}/*.dict (not invented soft maps).
func TestDiscoverLanguagePOSDicts_Upstream(t *testing.T) {
	langs := []string{"ar", "br", "da", "el", "gl", "it", "km", "ml", "pl", "ro", "ru", "sk", "sr", "sv", "ta", "tl"}
	for _, lang := range langs {
		lang := lang
		t.Run(lang, func(t *testing.T) {
			p := DiscoverLanguagePOSDict(nil, lang)
			require.NotEmpty(t, p, "dict path for %s", lang)
			if lang == "sr" {
				// Java SerbianTagger: /sr/dictionary/ekavian/serbian.dict
				require.Contains(t, p, filepath.Join("dictionary", "ekavian", "serbian.dict"))
			}
			lt := languagetool.NewJLanguageTool(lang)
			require.True(t, languagetool.RegisterBinaryPOSTagger(lt, p), "open %s", p)
			require.NotNil(t, lt.TagWord)
		})
	}
}
