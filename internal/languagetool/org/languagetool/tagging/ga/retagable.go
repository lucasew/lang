package ga

// Retaggable ports org.languagetool.tagging.ga.Retaggable.
type Retaggable struct {
	Word          string
	RestrictToPos string
	AppendTag     string
	Prefix        string
}

func NewRetaggable(word, restrictToPos, appendTag string) *Retaggable {
	return &Retaggable{Word: word, RestrictToPos: restrictToPos, AppendTag: appendTag}
}

func NewRetaggableWithPrefix(word, restrictToPos, appendTag, prefix string) *Retaggable {
	return &Retaggable{Word: word, RestrictToPos: restrictToPos, AppendTag: appendTag, Prefix: prefix}
}

func (r *Retaggable) GetWord() string          { return r.Word }
func (r *Retaggable) GetRestrictToPos() string { return r.RestrictToPos }
func (r *Retaggable) GetAppendTag() string     { return r.AppendTag }
func (r *Retaggable) GetPrefix() string        { return r.Prefix }

func (r *Retaggable) SetAppendTag(appendTag string) {
	if r.AppendTag == "" {
		r.AppendTag = appendTag
	} else {
		r.AppendTag += appendTag
	}
}

func (r *Retaggable) SetRestrictToPos(restrictToPos string) {
	if r.RestrictToPos == "" {
		r.RestrictToPos = restrictToPos
	} else {
		r.RestrictToPos += "|" + restrictToPos
	}
}
