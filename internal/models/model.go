package models

import (
	"os"
	"time"
)

type File struct {
	FileInstans   *os.File
	FPath         string
	IsPdf         bool
	ModifyedAt    time.Time
	NeedToConvert bool
}

func NewFile(path string, t time.Time) *File {
	return &File{
		FPath:         path,
		IsPdf:         false,
		ModifyedAt:    t,
		NeedToConvert: false,
	}
}

// delete this
func (f *File) GetTimeMod() time.Time {
	return f.ModifyedAt
}
