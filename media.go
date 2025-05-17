package talkops

const (
	MediaTypeAttachment = "attachment"
	MediaTypeImage = "image"
	MediaTypeLink = "link"
	MediaTypeVideo = "video"
)

type Media struct {
	Type string `json:"type"`
	URL string `json:"url"`
	Filename string `json:"filename,omitempty"`
}

func NewAttachment(url string, filename string) *Media {
	return &Media{
		Type: MediaTypeAttachment,
		URL: url,
		Filename: filename,
	}
}

func NewImage(url string) *Media {
	return &Media{
		Type: MediaTypeImage,
		URL: url,
	}
}

func NewLink(url string) *Media {
	return &Media{
		Type: MediaTypeLink,
		URL: url,
	}
}

func NewVideo(url string) *Media {
	return &Media{
		Type: MediaTypeVideo,
		URL: url,
	}
}
