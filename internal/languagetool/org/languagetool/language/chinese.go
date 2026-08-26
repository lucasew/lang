package language

// Chinese re-exports SmallLang Chinese with a dedicated file twin.
var ChineseLang = Chinese

func NewChinese() SmallLang { return Chinese }
