package server

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// ConfidenceMapLoader ports org.languagetool.server.ConfidenceMapLoader.
type ConfidenceMapLoader struct{}

func NewConfidenceMapLoader() *ConfidenceMapLoader { return &ConfidenceMapLoader{} }

// Load reads rule confidence files for the given language short codes.
// filePattern must contain "{lang}" placeholder (e.g. "/data/conf-{lang}.csv").
// Each line: RULE_ID,float_value[,...]
func (l *ConfidenceMapLoader) Load(filePattern string, langCodes []string) (map[tools.ConfidenceKey]float32, error) {
	if !strings.Contains(filePattern, "{lang}") {
		return nil, fmt.Errorf("the 'ruleIdToConfidenceFile' parameter must contain '{lang}' as a placeholder for the language code")
	}
	confMap := map[tools.ConfidenceKey]float32{}
	for _, lang := range langCodes {
		fileName := strings.ReplaceAll(filePattern, "{lang}", lang)
		f, err := os.Open(fileName)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.Split(line, ",")
			if len(parts) < 2 {
				_ = f.Close()
				return nil, fmt.Errorf("invalid line in %s, expected 'RULE_ID,float_value[,...]': %s", fileName, line)
			}
			conf, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 32)
			if err != nil {
				_ = f.Close()
				return nil, fmt.Errorf("invalid confidence float value in %s: %s", fileName, line)
			}
			key := tools.NewConfidenceKey(lang, strings.TrimSpace(parts[0]))
			confMap[key] = float32(conf)
		}
		_ = f.Close()
		if err := sc.Err(); err != nil {
			return nil, err
		}
	}
	if len(confMap) == 0 {
		return nil, fmt.Errorf("no confidence values could be loaded for %s", filePattern)
	}
	return confMap, nil
}
