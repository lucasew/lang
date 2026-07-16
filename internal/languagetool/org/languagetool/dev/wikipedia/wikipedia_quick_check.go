package wikipedia

import (
	"encoding/xml"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

var (
	wikipediaURLRegex       = regexp.MustCompile(`(?i)^https?://(..)\.wikipedia\.org/wiki/(.*)$`)
	secureWikipediaURLRegex = regexp.MustCompile(`(?i)^https://secure\.wikimedia\.org/wikipedia/(..)/wiki/(.*)$`)
	reInterlangLink         = regexp.MustCompile(`\[\[[a-z]{2,6}:.*?\]\]`)
	reCategoryLink          = regexp.MustCompile(`(?i)\[\[:?(Category|Categoria|Categoría|Catégorie|Kategorie):.*?\]\]`)
	reFileLinkPrefix        = regexp.MustCompile(`(?i)(File|Fitxer|Fichero|Ficheiro|Fichier|Datei):.*?\.(png|jpg|svg|jpeg|tiff|gif)\|(?:(?:thumb|miniatur)\|)?(?:(?:right|left)\|)?`)
	reRevPreserve           = regexp.MustCompile(`(?s)<rev[^>]*xml:space="preserve"[^>]*>(.*?)</rev>`)
)

// WikipediaQuickCheck ports org.languagetool.dev.wikipedia.WikipediaQuickCheck
// without live HTTP / full JLanguageTool (plain-text + mapping path green).
type WikipediaQuickCheck struct {
	MaxSizeBytes    int
	DisabledRuleIDs []string
	filter          *SimpleWikipediaTextFilter
}

func NewWikipediaQuickCheck() *WikipediaQuickCheck {
	return &WikipediaQuickCheck{
		MaxSizeBytes: int(^uint(0) >> 1),
		filter:       NewSimpleWikipediaTextFilter(),
	}
}

// MatchWikipediaURL returns lang code and page title, or error.
func MatchWikipediaURL(raw string) (lang, title string, err error) {
	if m := wikipediaURLRegex.FindStringSubmatch(raw); m != nil {
		return m[1], m[2], nil
	}
	if m := secureWikipediaURLRegex.FindStringSubmatch(raw); m != nil {
		return m[1], m[2], nil
	}
	return "", "", fmt.Errorf("URL does not seem to be a valid Wikipedia URL: %s", raw)
}

func (c *WikipediaQuickCheck) GetLanguageCode(rawURL string) (string, error) {
	lang, _, err := MatchWikipediaURL(rawURL)
	return lang, err
}

func (c *WikipediaQuickCheck) GetPageTitle(rawURL string) (string, error) {
	_, title, err := MatchWikipediaURL(rawURL)
	if err != nil {
		return "", err
	}
	if dec, e := url.PathUnescape(title); e == nil {
		return dec, nil
	}
	return title, nil
}

func (c *WikipediaQuickCheck) ValidateWikipediaURL(rawURL string) error {
	_, _, err := MatchWikipediaURL(rawURL)
	return err
}

func (c *WikipediaQuickCheck) SetDisabledRuleIDs(ids []string) {
	c.DisabledRuleIDs = append([]string(nil), ids...)
}

func (c *WikipediaQuickCheck) GetDisabledRuleIDs() []string {
	return append([]string(nil), c.DisabledRuleIDs...)
}

// RemoveWikipediaLinks ports WikipediaQuickCheck.removeWikipediaLinks.
func RemoveWikipediaLinks(wikiContent string) string {
	s := reInterlangLink.ReplaceAllString(wikiContent, "")
	s = reCategoryLink.ReplaceAllString(s, "")
	s = reFileLinkPrefix.ReplaceAllString(s, "")
	return s
}

// GetPlainText extracts revision wikitext from MediaWiki API XML and filters to plain text.
func (c *WikipediaQuickCheck) GetPlainText(completeWikiContent string) (string, error) {
	wiki, err := ParseMediaWikiAPIContent(completeWikiContent)
	if err != nil {
		return "", err
	}
	cleaned := RemoveWikipediaLinks(wiki.Content)
	return c.filter.Filter(cleaned), nil
}

// GetPlainTextMapping returns plain text mapping (Sweble-accurate original coords deferred).
func (c *WikipediaQuickCheck) GetPlainTextMapping(completeWikiContent string) (*PlainTextMapping, error) {
	wiki, err := ParseMediaWikiAPIContent(completeWikiContent)
	if err != nil {
		return nil, err
	}
	plain := c.filter.Filter(wiki.Content)
	return NewPlainTextMappingWithOriginal(plain, wiki.Content), nil
}

// CheckPlainText builds a result with an optional matcher (full LT soft-skipped).
func (c *WikipediaQuickCheck) CheckPlainText(plainText, langCode string, matchFn func(string) []*rules.RuleMatch) *WikipediaQuickCheckResult {
	var matches []*rules.RuleMatch
	if matchFn != nil {
		matches = matchFn(plainText)
	}
	return NewWikipediaQuickCheckResult(plainText, matches, langCode)
}

// ParseMediaWikiAPIContent pulls <rev> text and timestamp from API XML.
func ParseMediaWikiAPIContent(completeWikiContent string) (MediaWikiContent, error) {
	dec := xml.NewDecoder(strings.NewReader(completeWikiContent))
	var content, ts string
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		switch se.Name.Local {
		case "rev":
			for _, a := range se.Attr {
				if a.Name.Local == "timestamp" {
					ts = a.Value
				}
			}
			var text string
			for {
				t2, err := dec.Token()
				if err != nil {
					break
				}
				switch v := t2.(type) {
				case xml.CharData:
					text += string(v)
				case xml.EndElement:
					if v.Name.Local == "rev" {
						goto revDone
					}
				}
			}
		revDone:
			if text != "" {
				content = text
			}
		}
	}
	if content == "" {
		if m := reRevPreserve.FindStringSubmatch(completeWikiContent); m != nil {
			content = m[1]
			// decode entities if raw regex path
			content = strings.ReplaceAll(content, "&amp;", "&")
			content = strings.ReplaceAll(content, "&nbsp;", "\u00A0")
			content = strings.ReplaceAll(content, "&lt;", "<")
			content = strings.ReplaceAll(content, "&gt;", ">")
		}
	}
	if content == "" {
		return MediaWikiContent{}, fmt.Errorf("could not parse MediaWiki API XML content")
	}
	return NewMediaWikiContent(content, ts), nil
}
