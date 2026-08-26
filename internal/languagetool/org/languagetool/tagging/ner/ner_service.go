package ner

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

const nerTimeout = 500 * time.Millisecond

// Span ports NERService.Span (PERSON entity span).
type Span struct {
	FromPos int
	ToPos   int
}

func NewSpan(from, to int) Span {
	if from >= to {
		panic(fmt.Sprintf("fromPos must be < toPos: fromPos: %d, toPos: %d", from, to))
	}
	return Span{FromPos: from, ToPos: to}
}

func (s Span) GetStart() int { return s.FromPos }
func (s Span) GetEnd() int   { return s.ToPos }
func (s Span) String() string {
	return fmt.Sprintf("%d-%d", s.FromPos, s.ToPos)
}

// NERService ports org.languagetool.tagging.ner.NERService.
// HTTP client is pluggable; ParseBuffer is the main offline surface.
type NERService struct {
	URLStr  string
	Client  *http.Client
	// Post optional override for tests (body is form-encoded input=...).
	Post func(endpoint, formBody string) (string, error)
	// Breaker optional circuit breaker.
	Breaker *tools.CircuitBreaker
}

func NewNERService(urlStr string) *NERService {
	return &NERService{
		URLStr:  urlStr,
		Client:  &http.Client{Timeout: nerTimeout},
		Breaker: tools.CircuitBreakerRegistry().GetOrCreate("ner-service"),
	}
}

// RunNER posts text and returns PERSON spans (empty on failure).
func (s *NERService) RunNER(text string) []Span {
	if s == nil {
		return nil
	}
	joined := strings.ReplaceAll(text, "\n", " ")
	if s.Breaker != nil && !s.Breaker.Allow() {
		return nil
	}
	form := "input=" + url.QueryEscape(joined)
	var (
		result string
		err    error
	)
	if s.Post != nil {
		result, err = s.Post(s.URLStr, form)
	} else {
		result, err = s.postTo(s.URLStr, form)
	}
	if err != nil {
		if s.Breaker != nil {
			s.Breaker.OnFailure()
		}
		return nil
	}
	if s.Breaker != nil {
		s.Breaker.OnSuccess()
	}
	return ParseBuffer(result)
}

func (s *NERService) postTo(endpoint, formBody string) (string, error) {
	client := s.Client
	if client == nil {
		client = &http.Client{Timeout: nerTimeout}
	}
	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(formBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("NER HTTP %d", resp.StatusCode)
	}
	return string(b), nil
}

// ParseBuffer ports NERService.parseBuffer.
// Tokens look like: word/TAG/from/to  (slashes from the right).
// Java: buffer.trim().split(" ") — String.trim + single-space split (not Fields).
func ParseBuffer(buffer string) []Span {
	values := strings.Split(tools.JavaStringTrim(buffer), " ")
	var res []Span
	for _, value := range values {
		if value == "" {
			continue
		}
		slash3 := lastSlashFrom(value, len(value)-1)
		slash2 := lastSlashFrom(value, slash3-1)
		slash1 := lastSlashFrom(value, slash2-1)
		if slash1 < 0 || slash2 < 0 || slash3 < 0 {
			continue
		}
		tag := value[slash1+1 : slash2]
		fromPos, err1 := strconv.Atoi(value[slash2+1 : slash3])
		toPos, err2 := strconv.Atoi(value[slash3+1:])
		if err1 != nil || err2 != nil {
			continue
		}
		if tag == "PERSON" && fromPos < toPos {
			res = append(res, Span{FromPos: fromPos, ToPos: toPos})
		}
	}
	return res
}

func lastSlashFrom(s string, startPos int) int {
	if startPos >= len(s) {
		startPos = len(s) - 1
	}
	for i := startPos; i >= 0; i-- {
		if s[i] == '/' {
			return i
		}
	}
	return -1
}
