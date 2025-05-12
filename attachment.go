package talkops

type Attachment struct {
	Link
	Filename string
}

func (a *Attachment) ToJSON() map[string]interface{} {
	data := a.Link.ToJSON()
	data["filename"] = a.Filename
	data["type"] = "attachment"
	return data
}
