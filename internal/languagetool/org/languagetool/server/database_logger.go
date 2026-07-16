package server

import (
	"sync"
	"time"
)

// DatabaseLogger ports org.languagetool.server.DatabaseLogger as an in-memory queue.
// SQL batch commit is deferred until a real DB backend is wired.
type DatabaseLogger struct {
	mu       sync.Mutex
	messages []DatabaseLogEntry
	maxQueue int
	enabled  bool
}

var defaultDBLogger = &DatabaseLogger{maxQueue: 50000}

func DBLogger() *DatabaseLogger { return defaultDBLogger }

func NewDatabaseLogger(maxQueue int) *DatabaseLogger {
	if maxQueue <= 0 {
		maxQueue = 50000
	}
	return &DatabaseLogger{maxQueue: maxQueue}
}

// Init enables logging (called when DatabaseAccess is ready).
func (l *DatabaseLogger) Init() {
	if l != nil {
		l.enabled = true
	}
}

func (l *DatabaseLogger) IsEnabled() bool {
	return l != nil && l.enabled
}

// Log enqueues an entry (drops when over capacity).
func (l *DatabaseLogger) Log(entry DatabaseLogEntry) {
	if l == nil || entry == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.enabled {
		return
	}
	if len(l.messages) >= l.maxQueue {
		return
	}
	l.messages = append(l.messages, entry)
}

// Poll drains up to n entries (for tests / worker loop).
func (l *DatabaseLogger) Poll(n int) []DatabaseLogEntry {
	if l == nil || n <= 0 {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.messages) == 0 {
		return nil
	}
	if n > len(l.messages) {
		n = len(l.messages)
	}
	out := append([]DatabaseLogEntry{}, l.messages[:n]...)
	l.messages = l.messages[n:]
	return out
}

func (l *DatabaseLogger) QueueSize() int {
	if l == nil {
		return 0
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.messages)
}

// DatabaseAccess is a minimal open-source stub (no real DB).
type DatabaseAccess struct {
	Ready bool
}

var dbAccess = &DatabaseAccess{}

func DatabaseAccessInstance() *DatabaseAccess { return dbAccess }

func (d *DatabaseAccess) InitOpenSource() {
	if d == nil {
		return
	}
	d.Ready = true
	DBLogger().Init()
}

func (d *DatabaseAccess) IsReady() bool {
	return d != nil && d.Ready
}

// DBGroup ports org.languagetool.server.DBGroup.
type DBGroup struct {
	ID   int64
	Name string
}

// DBGroupMember ports org.languagetool.server.DBGroupMember.
type DBGroupMember struct {
	ID      int64
	GroupID int64
	UserID  int64
	Role    string
}

func NewDBGroupMember(id, groupID, userID int64, roles []GroupRole) DBGroupMember {
	return DBGroupMember{
		ID:      id,
		GroupID: groupID,
		UserID:  userID,
		Role:    EncodeGroupRoles(roles),
	}
}

// DBInvite ports org.languagetool.server.DBInvite.
type DBInvite struct {
	ID        int64
	GroupID   int64
	Email     string
	Token     string
	CreatedAt time.Time
}

func NewDBInvite(id, groupID int64, email, token string) DBInvite {
	return DBInvite{
		ID:        id,
		GroupID:   groupID,
		Email:     email,
		Token:     token,
		CreatedAt: time.Now().UTC(),
	}
}
