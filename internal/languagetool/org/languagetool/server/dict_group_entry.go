package server

// DictGroupEntry ports org.languagetool.server.DictGroupEntry.
type DictGroupEntry struct {
	ID          int64
	Name        string
	UserGroupID *int64
}

func NewDictGroupEntry(id int64, name string, userGroupID *int64) DictGroupEntry {
	return DictGroupEntry{ID: id, Name: name, UserGroupID: userGroupID}
}
