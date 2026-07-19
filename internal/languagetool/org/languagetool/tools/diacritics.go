package tools

import (
	"strings"
	"unicode"
)

// RemoveDiacritics ports StringTools.removeDiacritics (NFD + strip combining marks).
// Uses precomposed Romance map + Mn filter so we do not depend on x/text when unused.
func RemoveDiacritics(str string) string {
	if str == "" {
		return str
	}
	var b strings.Builder
	b.Grow(len(str))
	for _, r := range str {
		switch r {
		case 'á', 'à', 'â', 'ä', 'ã', 'å':
			b.WriteByte('a')
		case 'é', 'è', 'ê', 'ë':
			b.WriteByte('e')
		case 'í', 'ì', 'î', 'ï':
			b.WriteByte('i')
		case 'ó', 'ò', 'ô', 'ö', 'õ':
			b.WriteByte('o')
		case 'ú', 'ù', 'û', 'ü':
			b.WriteByte('u')
		case 'ý', 'ÿ':
			b.WriteByte('y')
		case 'ç':
			b.WriteByte('c')
		case 'ñ':
			b.WriteByte('n')
		case 'Á', 'À', 'Â', 'Ä', 'Ã', 'Å':
			b.WriteByte('A')
		case 'É', 'È', 'Ê', 'Ë':
			b.WriteByte('E')
		case 'Í', 'Ì', 'Î', 'Ï':
			b.WriteByte('I')
		case 'Ó', 'Ò', 'Ô', 'Ö', 'Õ':
			b.WriteByte('O')
		case 'Ú', 'Ù', 'Û', 'Ü':
			b.WriteByte('U')
		case 'Ý':
			b.WriteByte('Y')
		case 'Ç':
			b.WriteByte('C')
		case 'Ñ':
			b.WriteByte('N')
		default:
			if unicode.Is(unicode.Mn, r) {
				continue
			}
			b.WriteRune(r)
		}
	}
	return b.String()
}
