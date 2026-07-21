package detector

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

const (
	fastTextKHighest = 5
	fastTextBufSize  = 4096
)

// FastTextException ports FastTextDetector.FastTextException.
type FastTextException struct {
	Msg      string
	Disabled bool
}

func (e *FastTextException) Error() string {
	if e == nil {
		return "FastTextException"
	}
	return e.Msg
}

// NewFastTextException ports FastTextException(String message, boolean disabled).
func NewFastTextException(message string, disabled bool) *FastTextException {
	return &FastTextException{Msg: message, Disabled: disabled}
}

// IsDisabled ports isDisabled (Lombok/boolean field).
func (e *FastTextException) IsDisabled() bool {
	return e != nil && e.Disabled
}

// FastTextDetector ports org.languagetool.language.identifier.detector.FastTextDetector.
// Can run a real fastText binary or use a pluggable Runner for tests.
type FastTextDetector struct {
	mu         sync.Mutex
	modelPath  string
	binaryPath string
	cmd        *exec.Cmd
	stdin      io.WriteCloser
	stdout     io.Reader
	// Runner optional: if set, used instead of external process.
	// Input is one line of lowercased text; output is fastText predict-prob format.
	Runner func(line string) (string, error)
	// CanDetect filters language codes (nil accepts all).
	CanDetect func(langCode string, additional []string) bool
}

// NewFastTextDetector starts fastText predict-prob subprocess.
func NewFastTextDetector(modelPath, binaryPath string) (*FastTextDetector, error) {
	d := &FastTextDetector{modelPath: modelPath, binaryPath: binaryPath}
	if err := d.initProcess(); err != nil {
		return nil, err
	}
	return d, nil
}

// NewFastTextDetectorForTest builds a detector without a process.
func NewFastTextDetectorForTest() *FastTextDetector {
	return &FastTextDetector{}
}

func (d *FastTextDetector) initProcess() error {
	if d.binaryPath == "" || d.modelPath == "" {
		return fmt.Errorf("fasttext binary and model paths required")
	}
	cmd := exec.Command(d.binaryPath, "predict-prob", d.modelPath, "-", strconv.Itoa(fastTextKHighest))
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	d.cmd = cmd
	d.stdin = stdin
	d.stdout = stdout
	return nil
}

// RunFasttext returns language→probability for text.
// Java: text.replace('\n', ' ').toLowerCase(Locale.ROOT) + System.lineSeparator().
// Go strings.ToLower is Unicode default case mapping (= Locale.ROOT, not Turkish locale).
func (d *FastTextDetector) RunFasttext(text string, additionalLanguageCodes []string) (map[string]float64, error) {
	// twin: String.join prep — newline→space then Locale.ROOT lower
	joined := javaLocaleRootToLower(strings.ReplaceAll(text, "\n", " "))
	d.mu.Lock()
	defer d.mu.Unlock()
	var buf string
	var err error
	if d.Runner != nil {
		buf, err = d.Runner(joined)
		if err != nil {
			return nil, err
		}
	} else if d.stdin != nil && d.stdout != nil {
		// Java Writer uses System.lineSeparator(); Linux/mac → "\n"
		if _, err := io.WriteString(d.stdin, joined+"\n"); err != nil {
			return nil, err
		}
		// read one line of output
		r := bufio.NewReaderSize(d.stdout, fastTextBufSize)
		line, err := r.ReadString('\n')
		if err != nil && err != io.EOF {
			return nil, err
		}
		if line == "" && err == io.EOF {
			return nil, &FastTextException{Msg: "fasttext returned no data", Disabled: true}
		}
		buf = line
	} else {
		return nil, fmt.Errorf("fasttext not configured")
	}
	return d.ParseBuffer(buf, additionalLanguageCodes)
}

// ParseBuffer ports FastTextDetector.parseBuffer.
// Java always filters with LanguageIdentifierService.canLanguageBeDetected
// (supported language OR listed in additionalLanguageCodes) — not "only additional".
func (d *FastTextDetector) ParseBuffer(buffer string, additionalLanguageCodes []string) (map[string]float64, error) {
	// Java: String[] values = WHITESPACE.split(buffer.trim());
	// WHITESPACE = Pattern.compile("\\s+") without UNICODE_CHARACTER_CLASS;
	// buffer.trim() is String.trim (≤U+0020). startsWith checks the *raw* buffer.
	values := javaFastTextWhitespaceSplit(tools.JavaStringTrim(buffer))
	if !strings.HasPrefix(buffer, "__label__") {
		return nil, &FastTextException{
			Msg:      "FastText output is expected to start with '__label__': '" + buffer + "'",
			Disabled: true,
		}
	}
	if len(values)%2 != 0 {
		return nil, &FastTextException{
			Msg:      "Error while parsing fasttext output, expected pairs: '" + buffer + "'",
			Disabled: true,
		}
	}
	probs := map[string]float64{}
	for i := 0; i < len(values); i += 2 {
		lang := values[i]
		idx := strings.LastIndex(lang, "__")
		langCode := lang
		if idx >= 0 && idx+2 < len(lang) {
			langCode = lang[idx+2:]
		}
		prob, err := strconv.ParseFloat(values[i+1], 64)
		if err != nil {
			return nil, err
		}
		// CanDetect override for tests; else Java canLanguageBeDetected
		if d.CanDetect != nil {
			if !d.CanDetect(langCode, additionalLanguageCodes) {
				continue
			}
		} else if !CanLanguageBeDetected(langCode, additionalLanguageCodes) {
			continue
		}
		probs[langCode] = prob
	}
	return probs, nil
}

// Destroy stops the subprocess if any.
func (d *FastTextDetector) Destroy() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.stdin != nil {
		_ = d.stdin.Close()
	}
	if d.cmd != nil && d.cmd.Process != nil {
		_ = d.cmd.Process.Kill()
		_, _ = d.cmd.Process.Wait()
	}
	d.cmd = nil
}

// javaLocaleRootToLower ports String.toLowerCase(Locale.ROOT).
// Go unicode.ToLower / strings.ToLower use the Unicode default mapping (same as Locale.ROOT),
// not a Turkish/Azeri locale ("I" → "i", never "ı").
func javaLocaleRootToLower(s string) string {
	return strings.ToLower(s)
}

// javaFastTextWhitespaceSplit ports Pattern.compile("\\s+").split without UNICODE_CHARACTER_CLASS.
// limit 0: trailing empties dropped; empty input → [""] like Java Pattern.split("").
func javaFastTextWhitespaceSplit(s string) []string {
	if s == "" {
		return []string{""}
	}
	var parts []string
	start := 0
	i := 0
	for i < len(s) {
		c := s[i]
		if c == ' ' || c == '\t' || c == '\n' || c == '\v' || c == '\f' || c == '\r' {
			if start < i {
				parts = append(parts, s[start:i])
			}
			j := i + 1
			for j < len(s) {
				c2 := s[j]
				if c2 != ' ' && c2 != '\t' && c2 != '\n' && c2 != '\v' && c2 != '\f' && c2 != '\r' {
					break
				}
				j++
			}
			start = j
			i = j
			continue
		}
		i++
	}
	if start < len(s) {
		parts = append(parts, s[start:])
	}
	if len(parts) == 0 {
		return []string{""}
	}
	return parts
}
