package models

import (
	"time"
)

type File struct {
	FPath         string
	IsPdf         bool
	ModifyedAt    time.Time
	NeedToConvert bool
	PdfContent    []byte
}

//type Pdf struct {
//	Name       string
//	PdfContent []byte
//}

func NewFile(path string, t time.Time) *File {
	return &File{
		FPath:         path,
		IsPdf:         false,
		ModifyedAt:    t,
		NeedToConvert: false,
	}
}

func (f *File) GetTimeMod() time.Time {
	return f.ModifyedAt
}

func (f *File) SetPdfContent(content []byte) {
	f.PdfContent = content
}

func (f *File) convertFileToPdf(md *File) {

}
