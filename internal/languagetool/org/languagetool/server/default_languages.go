package server

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/corepack"

// commonVariants soft-expands /v2/languages the way Java lists short + long codes.
// Values are (longCode, display name suffix).
var commonVariants = map[string][]struct {
	Long string
	Name string
}{
	"en": {{"en-US", "English (US)"}, {"en-GB", "English (GB)"}, {"en-AU", "English (Australian)"}, {"en-CA", "English (Canadian)"}},
	"de": {{"de-DE", "German (Germany)"}, {"de-AT", "German (Austria)"}, {"de-CH", "German (Swiss)"}},
	"pt": {{"pt-PT", "Portuguese (Portugal)"}, {"pt-BR", "Portuguese (Brazil)"}},
	"es": {{"es", "Spanish"}, {"es-ES", "Spanish (Spain)"}, {"es-AR", "Spanish (Argentina)"}, {"es-MX", "Spanish (Mexico)"}},
	"fr": {{"fr", "French"}, {"fr-FR", "French (France)"}, {"fr-CA", "French (Canada)"}},
	"nl": {{"nl", "Dutch"}, {"nl-NL", "Dutch (Netherlands)"}, {"nl-BE", "Dutch (Belgium)"}},
	"it": {{"it", "Italian"}, {"it-IT", "Italian (Italy)"}},
	"pl": {{"pl-PL", "Polish (Poland)"}},
	"ru": {{"ru-RU", "Russian (Russia)"}},
	"uk": {{"uk-UA", "Ukrainian (Ukraine)"}},
	"ca": {{"ca-ES", "Catalan (Spain)"}},
	"sv": {{"sv-SE", "Swedish (Sweden)"}},
	"da": {{"da-DK", "Danish (Denmark)"}},
	"sk": {{"sk-SK", "Slovak (Slovakia)"}},
	"el": {{"el-GR", "Greek (Greece)"}},
	"gl": {{"gl-ES", "Galician"}},
	"ro": {{"ro-RO", "Romanian (Romania)"}},
	"sl": {{"sl-SI", "Slovenian (Slovenia)"}},
	"ar": {{"ar", "Arabic"}},
	"ja": {{"ja-JP", "Japanese"}},
	"zh": {{"zh-CN", "Chinese (Simplified)"}},
}

// DefaultCoreLanguages returns LanguageInfo for all corepack-supported languages,
// expanded with soft longCode variants for LibreOffice/clients that need them.
func DefaultCoreLanguages() []LanguageInfo {
	out := make([]LanguageInfo, 0, len(corepack.Supported)*3)
	seen := map[string]struct{}{}
	add := func(name, code, long string) {
		key := long
		if key == "" {
			key = code
		}
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		if long == "" {
			long = code
		}
		out = append(out, LanguageInfo{Name: name, Code: code, LongCode: long})
	}
	for _, s := range corepack.Supported {
		base := s.Code
		if variants, ok := commonVariants[base]; ok {
			for _, v := range variants {
				// code is always the short base for LT clients
				add(v.Name, base, v.Long)
			}
		} else {
			add(s.Name, base, base)
		}
	}
	return out
}
