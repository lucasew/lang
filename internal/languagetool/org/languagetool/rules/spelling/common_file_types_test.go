package spelling

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// javaMatches emulates Java Pattern.matcher(s).matches() (full-region match).
// GetSuffixPattern returns the unanchored Pattern twin; callers in Java use matches().
func javaMatches(re *regexp.Regexp, s string) bool {
	return regexp.MustCompile(`^(?:` + re.String() + `)$`).MatchString(s)
}

func TestCommonFileTypes_ListOrderAndDuplicateCGI(t *testing.T) {
	// Exact COMMON_FILE_TYPES from Java Arrays.asList (order + duplicate "cgi").
	want := []string{
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
	require.Equal(t, want, commonFileTypes)
	cgi := 0
	for _, e := range commonFileTypes {
		if e == "cgi" {
			cgi++
		}
	}
	require.Equal(t, 2, cgi, "Java lists cgi twice")
}

func TestCommonFileTypes_PatternSource(t *testing.T) {
	// Java Pattern.pattern() body (flags separate as CASE_INSENSITIVE=2).
	wantBody := `[\wáàâóòìíéèùúôîêûäöüß\-.()]*?.+\.(` + strings.Join(commonFileTypes, "|") + `)`
	got := GetSuffixPattern().String()
	require.Equal(t, `(?i)`+wantBody, got, "Go (?i) maps Java Pattern.CASE_INSENSITIVE; body must match Java pattern()")
	// Reluctant quantifier and character class preserved.
	require.Contains(t, got, `*?`)
	require.Contains(t, got, `áàâóòìíéèùúôîêûäöüß`)
	require.Contains(t, got, `tar.gz`)
}

func TestCommonFileTypes_SuffixMatches_JavaOracle(t *testing.T) {
	re := GetSuffixPattern()
	// Outcomes from Java CommonFileTypes.getSuffixPattern().matcher(s).matches()
	// (oracle: same list+pattern compiled with Pattern.CASE_INSENSITIVE).
	cases := []struct {
		s    string
		want bool
	}{
		// agreed matrix
		{"photo.JPG", true},
		{"report.pdf", true},
		{"hello", false},
		{".pdf", false}, // needs something before extension per .+\\.
		// path-like / punctuation / German chars in class
		{"a.pdf", true},
		{"my-file_v2.pdf", true},
		{"äöü.pdf", true},
		{"file(1).pdf", true},
		{"dir/file.pdf", true}, // .+ may consume '/'
		{"hello world.pdf", true},
		// extension case folding (CASE_INSENSITIVE)
		{"photo.JPEG", true},
		{"x.PDF", true},
		{"FILE.DOCX", true},
		{"Report.PDF", true},
		// multi-part extension in list (dot unescaped as in Java join)
		{"a.tar.gz", true},
		{"x.cgi", true},
		{"sketch.ai", true},
		// negatives
		{"file.unknown", false},
		{"not.pdf.extra", false},
		{"xx file.pdf yy", false},
	}
	for _, tc := range cases {
		t.Run(tc.s, func(t *testing.T) {
			require.Equal(t, tc.want, javaMatches(re, tc.s), "Java matches() twin for %q", tc.s)
		})
	}
}
