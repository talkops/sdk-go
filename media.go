package talkops

type Media struct {
	Type string
}

func (m *Media) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"type": m.Type,
	}
}
