package language

// Spanish is the default Spain Spanish variant.
var Spanish = SpanishSpain

func NewSpanish() SpanishVariant      { return SpanishSpain }
func NewSpanishVoseo() SpanishVariant { return SpanishVoseo }
