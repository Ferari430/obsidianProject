package models

import "time"

type File struct {
	FPath      string
	IsPdf      bool
	ModifyedAt time.Time
}

func NewFile(path string, t time.Time) *File {
	return &File{
		FPath:      path,
		IsPdf:      false,
		ModifyedAt: t,
	}
}

// delete this
func (f *File) GetTimeMod() time.Time {
	return f.ModifyedAt
}
