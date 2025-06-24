package types

type File struct{}

func (File) ContentType() string {
	return "application/octet-stream"
}
