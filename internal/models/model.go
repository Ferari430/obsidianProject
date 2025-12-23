package models

type File struct {
	FPath string
	IsPdf bool
}

func NewFile(path string) *File {
	return &File{
		FPath: path,
		IsPdf: false,
	}
}
