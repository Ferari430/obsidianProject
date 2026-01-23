package models

import (
	"os"
	"time"
)

// MDFile содержит информацию о MD файле с полным путём
type MDFile struct {
	Name     string // Имя файла
	FullPath string // Полный путь до файла
	DirEntry os.DirEntry
}

type File struct {
	FPath         string // Имя файла (например: "1.md")
	FullPath      string // Полный путь до файла (например: "B:\data\sobesVault\interview\1.md")
	IsPdf         bool
	ModifyedAt    time.Time
	NeedToConvert bool
	PdfContent    []byte
}

func NewFile(path string, fullPath string, t time.Time) *File {
	return &File{
		FPath:         path,
		FullPath:      fullPath,
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
