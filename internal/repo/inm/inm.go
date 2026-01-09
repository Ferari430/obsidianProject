package inm

import (
	"errors"
	"log"
	"os"

	"github.com/Ferari430/obsidianProject/internal/models"
)

type Postgres struct {
	table          []*models.File // уже обработанные файлы
	converterFiles []*models.File // файлы для конвертации
}

func NewPostgres() *Postgres {
	arr := make([]*models.File, 0)

	return &Postgres{table: arr}
}

func (s *Postgres) Add(collectedMdFiles []os.DirEntry) error {
	s.converterFiles = make([]*models.File, 0)

	for _, f := range collectedMdFiles {
		if existingFile, ok := s.checkHandledFile(f); !ok {
			fileInfo, err := f.Info()
			if err != nil {
				return err
			}

			newFile := models.NewFile(f.Name(), fileInfo.ModTime())
			// если нет такого файла то добавляю в оба слайса
			s.table = append(s.table, newFile)
			s.converterFiles = append(s.converterFiles, newFile)
		} else {
			log.Println("file already in db")
			//если есть то проверяю не изменен ли файл

			//если время не совпадает то конвертируем
			if !s.checkModifyFile(f, existingFile) {
				log.Println("время не совпадает, конвертация файла....", existingFile.FPath)
				s.converterFiles = append(s.converterFiles, existingFile)
			}
		}
	}

	log.Println("только пдф файлы:", s.GetAllPDFFiles())

	l := len(collectedMdFiles)
	log.Printf("добавлено %d файлов", l)
	return nil
}

func (s *Postgres) Get() []*models.File {
	log.Println("len converterFiles: ", len(s.converterFiles))
	return s.converterFiles
}

func (s *Postgres) GetConfirmedFiles() ([]*models.File, error) {
	if s.table == nil || len(s.table) == 0 {
		return nil, errors.New("there are no confirmed files")
	}

	log.Println("len confirmed files: ", len(s.table))
	return s.table, nil
}

func (s *Postgres) checkModifyFile(f os.DirEntry, existingFile *models.File) bool {
	op := "postgres.CheckModifyFile"
	info, err := f.Info()
	if err != nil {
		log.Println(op, err)
		return false
	}

	flag := info.ModTime() == existingFile.GetTimeMod()

	//if flag == false {
	//	log.Println("время не совпадает", "fileInfo", info.ModTime(), "fileInfoDb", existingFile.GetTimeMod())
	//}
	//
	//if flag == true {
	//	log.Println("время совпадает", "fileInfo", info.ModTime(), "fileInfoDb", existingFile.GetTimeMod())
	//}

	existingFile.NeedToConvert = true

	return flag
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

func (s *Postgres) GetAllPDFFiles() []*models.File {
	arr := make([]*models.File, 0)
	for _, f := range s.table {

		log.Println("file:", f.FPath, f.IsPdf, f.ModifyedAt)

		if f.IsPdf {
			arr = append(arr, f)
		}
	}

	return arr
}
