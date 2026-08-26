package wikipedia

// MediaWikiContent ports org.languagetool.dev.wikipedia.MediaWikiContent.
type MediaWikiContent struct {
	Content   string
	Timestamp string
}

func NewMediaWikiContent(content, timestamp string) MediaWikiContent {
	return MediaWikiContent{Content: content, Timestamp: timestamp}
}

func (m MediaWikiContent) GetContent() string   { return m.Content }
func (m MediaWikiContent) GetTimestamp() string { return m.Timestamp }
