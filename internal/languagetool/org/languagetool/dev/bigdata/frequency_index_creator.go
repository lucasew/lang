package bigdata

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"
)

const (
	MinYear             = 1910
	LTCompleteMarker    = "languagetool_index_complete"
	GoogleSentenceStart = "_START_"
	GoogleSentenceEnd   = "_END_"
)

// NgramCount is one aggregated ngram → total count.
type NgramCount struct {
	Text  string
	Count int64
}

// IsRealPOSTag ports FrequencyIndexCreator.isRealPosTag (POS-tagged ngrams to skip).
func IsRealPOSTag(text string) bool {
	idx := strings.Index(text, "_")
	if idx < 0 {
		return false
	}
	// _START_ / _END_ are not "real" POS tags to skip for IGNORE_POS? Java returns false for them
	tag := ""
	if idx+7 <= len(text) {
		tag = text[idx : idx+7]
	}
	if tag == GoogleSentenceStart {
		return false
	}
	tag2 := ""
	if idx+5 <= len(text) {
		tag2 = text[idx : idx+5]
	}
	if tag2 == GoogleSentenceEnd {
		return false
	}
	return true
}

// AggregateGoogleNgramLines ports corpus-mode year aggregation from FrequencyIndexCreator.
// Lines: ngram \t year \t count [\t ...]
// ignorePOS skips real POS-tagged ngrams when true.
func AggregateGoogleNgramLines(r io.Reader, ignorePOS bool) ([]NgramCount, error) {
	sc := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 4*1024*1024)
	var out []NgramCount
	var prevText string
	var docCount int64
	first := true
	flush := func() {
		if !first && prevText != "" {
			out = append(out, NgramCount{Text: prevText, Count: docCount})
		}
	}
	for sc.Scan() {
		line := sc.Text()
		parts := strings.Split(line, "\t")
		if len(parts) < 3 {
			continue
		}
		text := parts[0]
		if ignorePOS && IsRealPOSTag(text) {
			continue
		}
		year, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}
		if year < MinYear {
			continue
		}
		count, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			continue
		}
		if first {
			prevText = text
			docCount = count
			first = false
			continue
		}
		if prevText == text {
			docCount += count
		} else {
			out = append(out, NgramCount{Text: prevText, Count: docCount})
			prevText = text
			docCount = count
		}
	}
	flush()
	return out, sc.Err()
}

// AggregateHiveNgramLines ports hive mode: ngram \t count (no year aggregation).
func AggregateHiveNgramLines(r io.Reader, ignorePOS bool) ([]NgramCount, error) {
	sc := bufio.NewScanner(r)
	var out []NgramCount
	for sc.Scan() {
		parts := strings.Split(sc.Text(), "\t")
		if len(parts) < 2 {
			continue
		}
		text := parts[0]
		if ignorePOS && IsRealPOSTag(text) {
			continue
		}
		count, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			continue
		}
		out = append(out, NgramCount{Text: text, Count: count})
	}
	return out, sc.Err()
}

// MatchesGoogleNgramFilename ports NAME_REGEX1 match.
var (
	nameRegex1 = regexp.MustCompile(`googlebooks-[a-z]{3}-all-[1-5]gram-20120701-(.*?)\.gz`)
	nameRegex2 = regexp.MustCompile(`[a-z0-9]+-[a-z0-9]+-[a-z0-9]+-[a-z0-9]+-[a-z0-9]+[_-](.*?)\.gz`)
)

func IsCorpusModeFilename(name string) bool { return nameRegex1.MatchString(name) }
func IsHiveModeFilename(name string) bool   { return nameRegex2.MatchString(name) }
func ShouldSkipPOSFilename(name string) bool {
	return strings.Contains(name, "_") && regexp.MustCompile(`.*_[A-Z]+_.*`).MatchString(name)
}
