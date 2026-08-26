package spelling

import (
	"regexp"
	"strings"
)

// Twin of org.languagetool.rules.spelling.CommonFileTypes
// (inspiration/languagetool/languagetool-core/src/main/java/org/languagetool/rules/spelling/CommonFileTypes.java).
// Static utility: private constructor in Java → no constructible type here.

// commonFileTypes ports CommonFileTypes.COMMON_FILE_TYPES (exact order/content).
// Java lists "cgi" twice; keep the same multiset so String.join order matches.
var commonFileTypes = []string{
	"jpeg", "jpg", "gif", "png", "bmp", "svg", "ai", "sketch", "ico", "ps", "psd", "tiff", "tif",
	"mp3", "wav", "midi", "mid", "aif", "mpa", "ogg", "wma", "wpl", "cda",
	"7z", "arj", "deb", "pkg", "plist", "rar", "rpm", "tar.gz", "tar", "zip",
	"bin", "dmg", "iso", "toast", "vcd", "csv", "dat", "db", "log", "mdb", "sav", "sql", "xml",
	"apk", "bat", "cgi", "com", "exe", "gadget", "jar", "py", "js", "jsx", "json", "wsf", "ts", "tsx",
	"fnt", "fon", "otf", "ttf", "woff", "woff2",
	"rb", "java", "php", "html", "asp", "aspx", "cer", "cfm", "cgi", "pl", "css", "scss", "htm", "jsp", "part", "rss", "xhtml",
	"key", "odp", "pps", "ppt", "pptx", "class", "cpp", "cs", "h", "sh", "swift", "vb",
	"ods", "odt", "xlr", "xls", "xlsx", "xlt", "xltx", "bak", "cab", "cfg", "cpl", "cur", "dll", "dmp", "msi", "ini", "tmp",
	"3g2", "3gp", "avi", "flv", "h264", "m4v", "mkv", "mov", "mp4", "mpg", "mpeg", "rm", "swf", "vob", "wmv",
	"doc", "docx", "dot", "dotx", "pdf", "rtf", "srx", "text", "tex", "wks", "wps", "wpd", "txt", "yaml", "yml", "csl", "md", "adm", "webm", "webp",
}

// suffixPatternSource is the Pattern body as Java builds it (before CASE_INSENSITIVE flag).
// Java:
//
//	Pattern.compile("[\\wáàâóòìíéèùúôîêûäöüß\\-.()]*?.+\\.(" + String.join("|", COMMON_FILE_TYPES) + ")",
//	  Pattern.CASE_INSENSITIVE);
const suffixPatternBodyPrefix = `[\wáàâóòìíéèùúôîêûäöüß\-.()]*?.+\.(`

// suffixPattern ports CommonFileTypes.suffixPattern (static final, CASE_INSENSITIVE).
// Go maps Pattern.CASE_INSENSITIVE to the (?i) flag (extensions are ASCII; parity holds).
var suffixPattern = regexp.MustCompile(`(?i)` + suffixPatternBodyPrefix + strings.Join(commonFileTypes, "|") + `)`)

// GetSuffixPattern ports CommonFileTypes.getSuffixPattern.
func GetSuffixPattern() *regexp.Regexp {
	return suffixPattern
}
