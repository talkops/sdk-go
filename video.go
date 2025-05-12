package talkops

type Video struct {
	Link
}

func (v *Video) ToJSON() map[string]interface{} {
	data := v.Link.ToJSON()
	data["type"] = "video"
	return data
}
