package server

// APINewGroup ports org.languagetool.server.APINewGroup (POST /groups body).
type APINewGroup struct {
	Name string `json:"name"`
}

func NewAPINewGroup(name string) APINewGroup {
	return APINewGroup{Name: name}
}

func (a APINewGroup) Equal(o APINewGroup) bool {
	return a.Name == o.Name
}
