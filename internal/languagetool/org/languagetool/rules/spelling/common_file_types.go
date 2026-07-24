package spelling

import (
	"regexp"
	"strings"
	"sync"
)

// commonFileTypes ports CommonFileTypes.COMMON_FILE_TYPES (order/content as Java).
// Note: Java lists "cgi" twice; keep the same multiset for 1:1 join.
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

var (
	suffixPatternOnce sync.Once
	suffixPattern     *regexp.Regexp
)

// GetSuffixPattern ports CommonFileTypes.getSuffixPattern (static CASE_INSENSITIVE Pattern).
func GetSuffixPattern() *regexp.Regexp {
	suffixPatternOnce.Do(func() {
		// Java: [\\wáàâóòìíéèùúôîêûäöüß\\-.()]*?.+\\.(" + join + ")
		pat := `[\wáàâóòìíéèùúôîêûäöüß\-.()]*?.+\.(` + strings.Join(commonFileTypes, "|") + `)`
		suffixPattern = regexp.MustCompile(`(?i)` + pat)
	})
	return suffixPattern
}
