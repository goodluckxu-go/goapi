package types

type Mp4 struct{}

func (Mp4) ContentType() string {
	return "video/mp4"
}

type Avi struct{}

func (Avi) ContentType() string {
	return "video/avi"
}
