package talkops

type Image struct {
	Link
}

func (i *Image) ToJSON() map[string]interface{} {
	data := i.Link.ToJSON()
	data["type"] = "image"
	return data
}
