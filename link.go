package talkops

type Link struct {
	Media
	URL string
}

func (l *Link) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"type": "link",
		"url":  l.URL,
	}
}
