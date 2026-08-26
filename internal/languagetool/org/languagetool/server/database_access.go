package server

// DatabaseAccessAPI is the shared DB façade (open-source no-op backend).
// Full MyBatis/SQL wiring deferred.
type DatabaseAccessAPI struct {
	*DatabaseAccess
	Logger *DatabaseLogger
}

// DatabaseAccessOpenSource is the OSS implementation (no premium DB).
type DatabaseAccessOpenSource struct {
	DatabaseAccessAPI
}

func NewDatabaseAccessOpenSource() *DatabaseAccessOpenSource {
	d := &DatabaseAccessOpenSource{
		DatabaseAccessAPI: DatabaseAccessAPI{
			DatabaseAccess: DatabaseAccessInstance(),
			Logger:         DBLogger(),
		},
	}
	return d
}

func (d *DatabaseAccessOpenSource) Init() {
	if d == nil {
		return
	}
	d.DatabaseAccess.InitOpenSource()
}

// GetUserByEmail returns nil in OSS mode without a real DB.
func (d *DatabaseAccessOpenSource) GetUserByEmail(email string) *UserInfoEntry {
	return nil
}

// GetUserByAPIKey returns nil in OSS mode.
func (d *DatabaseAccessOpenSource) GetUserByAPIKey(username, apiKey string) *UserInfoEntry {
	return nil
}

// LogCheck enqueues a check log entry when logging is enabled.
func (d *DatabaseAccessOpenSource) LogCheck(userID *int64, textSize int, lang string, matchCount int) {
	if d == nil || d.Logger == nil {
		return
	}
	d.Logger.Log(NewDatabaseCheckLogEntry(userID, textSize, lang, matchCount))
}
