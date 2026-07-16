package server

import "time"

// DatabaseLogEntry ports org.languagetool.server.DatabaseLogEntry.
type DatabaseLogEntry interface {
	GetMapping() map[string]any
	GetMappingIdentifier() string
	Followup() DatabaseLogEntry
}

// DatabasePingLogEntry ports org.languagetool.server.DatabasePingLogEntry.
type DatabasePingLogEntry struct {
	Date   time.Time
	Client *int64
	User   *int64
}

func NewDatabasePingLogEntry(client, user *int64) *DatabasePingLogEntry {
	return &DatabasePingLogEntry{Date: time.Now().UTC(), Client: client, User: user}
}

func (e *DatabasePingLogEntry) GetMapping() map[string]any {
	m := map[string]any{
		"day":        e.Date.Format("2006-01-02"),
		"created_at": e.Date.Format("2006-01-02 15:04:05"),
	}
	if e.Client != nil {
		m["client"] = *e.Client
	}
	if e.User != nil {
		m["user_id"] = *e.User
	}
	return m
}

func (e *DatabasePingLogEntry) GetMappingIdentifier() string {
	return "org.languagetool.server.LogMapper.pings"
}

func (e *DatabasePingLogEntry) Followup() DatabaseLogEntry { return nil }

// DatabaseCheckLogEntry ports a simplified check log row.
type DatabaseCheckLogEntry struct {
	Date         time.Time
	UserID       *int64
	TextSize     int
	LanguageCode string
	MatchCount   int
	Client       *int64
}

func NewDatabaseCheckLogEntry(userID *int64, textSize int, lang string, matchCount int) *DatabaseCheckLogEntry {
	return &DatabaseCheckLogEntry{
		Date:         time.Now().UTC(),
		UserID:       userID,
		TextSize:     textSize,
		LanguageCode: lang,
		MatchCount:   matchCount,
	}
}

func (e *DatabaseCheckLogEntry) GetMapping() map[string]any {
	m := map[string]any{
		"created_at":    e.Date.Format("2006-01-02 15:04:05"),
		"text_size":     e.TextSize,
		"language":      e.LanguageCode,
		"match_count":   e.MatchCount,
	}
	if e.UserID != nil {
		m["user_id"] = *e.UserID
	}
	if e.Client != nil {
		m["client"] = *e.Client
	}
	return m
}

func (e *DatabaseCheckLogEntry) GetMappingIdentifier() string {
	return "org.languagetool.server.LogMapper.checks"
}

func (e *DatabaseCheckLogEntry) Followup() DatabaseLogEntry { return nil }
