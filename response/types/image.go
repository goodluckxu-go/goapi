package types

type Png struct{}

func (Png) ContentType() string {
	return "image/png"
}

type Jpeg struct{}

func (Jpeg) ContentType() string {
	return "image/jpeg"
}

type Gif struct{}

func (Gif) ContentType() string {
	return "image/gif"
}
