package types

type Pdf struct{}

func (Pdf) ContentType() string {
	return "application/pdf"
}

type Doc struct{}

func (Doc) ContentType() string {
	return "application/msword"
}

type Docx struct{}

func (Docx) ContentType() string {
	return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
}

type Xls struct{}

func (Xls) ContentType() string {
	return "application/vnd.ms-excel"
}

type Xlsx struct{}

func (Xlsx) ContentType() string {
	return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
}

type Ppt struct{}

func (Ppt) ContentType() string {
	return "application/vnd.ms-powerpoint"
}

type Pptx struct{}

func (Pptx) ContentType() string {
	return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
}
