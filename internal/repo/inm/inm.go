package inm

import (
	"log"
	"os"

	"github.com/Ferari430/obsidianProject/internal/models"
)

type Postgres struct {
	table          []*models.File
	converterFiles []*models.File
}

func NewPostgres() *Postgres {
	arr := make([]*models.File, 0)
	arr1 := make([]*models.File, 0)
	return &Postgres{table: arr,
		converterFiles: arr1,
	}
}

func (s *Postgres) Add(collectedMdFiles []os.DirEntry) error {
	s.converterFiles = make([]*models.File, 0)

	for _, f := range collectedMdFiles {
		if existingFile, ok := s.checkHandledFile(f); !ok {
			modifiedAt, err := f.Info()
			if err != nil {
				return err
			}

			newFile := models.NewFile(f.Name(), modifiedAt.ModTime())
			s.table = append(s.table, newFile)
			s.converterFiles = append(s.converterFiles, newFile)
		} else {
			log.Println("file already in db")
			if !s.checkModifyFile(f, existingFile) {
				s.converterFiles = append(s.converterFiles, existingFile)
			}
		}
	}

	l := len(collectedMdFiles)
	log.Printf("добавлено %d файлов", l)
	return nil
}

func (s *Postgres) Get() []*models.File {
	log.Println("len converterFiles: ", len(s.converterFiles))
	return s.converterFiles
}

func (s *Postgres) checkModifyFile(f os.DirEntry, existingFile *models.File) bool {
	op := "postgres.CheckModifyFile"
	info, err := f.Info()
	if err != nil {
		log.Println(op, err)
		return false
	}

	return info.ModTime() != existingFile.GetTimeMod()
}

func (s *Postgres) checkHandledFile(f os.DirEntry) (*models.File, bool) {
	for _, file := range s.table {
		if file.FPath == f.Name() {
			return file, true
		}
	}

	return nil, false
}

func (s *Postgres) AddPDFFiles(pdfFiles []os.DirEntry) error {
	for _, f := range pdfFiles {
		info, err := f.Info()
		if err != nil {
			return err
		}

		newFile := &models.File{
			FPath:      f.Name(),
			IsPdf:      true,
			ModifyedAt: info.ModTime(),
		}
		s.table = append(s.table, newFile)
	}
	return nil
}

func (s *Postgres) getAllFIlesName() {
	for _, f := range s.table {
		log.Printf("in table %s", f.FPath)
	}
}
