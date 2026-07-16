package spelling

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"

// IsUrl ports SpellingCheckRule.isUrl → WordTokenizer.isUrl.
func IsUrl(token string) bool { return tokenizers.IsURL(token) }

// IsEMail ports SpellingCheckRule.isEMail → WordTokenizer.isEMail.
func IsEMail(token string) bool { return tokenizers.IsEMail(token) }
