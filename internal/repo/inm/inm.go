package inm

import (
	"errors"
	"log"
	"math/rand"
	"os"
	"sync"

	"github.com/Ferari430/obsidianProject/internal/models"
)

type Postgres struct {
	mu             sync.RWMutex
	table          []*models.File // уже обработанные файлы
	converterFiles []*models.File // файлы для конвертации
	pdfFiles       []*models.File
}

func NewPostgres() *Postgres {
	return &Postgres{
		table:          make([]*models.File, 0),
		converterFiles: make([]*models.File, 0),
	}
}

func (s *Postgres) Add(collectedMdFiles []os.DirEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	//s.converterFiles = make([]*models.File, 0)

	for _, f := range collectedMdFiles {
		if existingFile, ok := s.checkHandledFileLocked(f); !ok {
			fileInfo, err := f.Info()
			if err != nil {
				return err
			}

			newFile := models.NewFile(f.Name(), fileInfo.ModTime())
			s.table = append(s.table, newFile)
			s.converterFiles = append(s.converterFiles, newFile)
		} else {
			log.Println("file already in db")

			if !s.checkModifyFile(f, existingFile) {
				log.Println("время не совпадает, конвертация файла....", existingFile.FPath)
				s.converterFiles = append(s.converterFiles, existingFile)
			}
		}
	}

	log.Println("только пдф файлы:", s.getAllPDFFilesLocked())

	l := len(collectedMdFiles)
	log.Printf("добавлено %d файлов", l)
	return nil
}

func (s *Postgres) Get() []*models.File {
	s.mu.RLock()
	defer s.mu.RUnlock()

	log.Println("len converterFiles: ", len(s.converterFiles))

	result := make([]*models.File, len(s.converterFiles))
	copy(result, s.converterFiles)
	return result
}

func (s *Postgres) RemoveFromConverter(fileName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, f := range s.converterFiles {
		if f.FPath == fileName {
			s.converterFiles = append(s.converterFiles[:i], s.converterFiles[i+1:]...)
			log.Printf("файл %s удален из очереди конвертации", fileName)
			return nil
		}
	}
	return errors.New("файл не найден в очереди конвертации")
}

func (s *Postgres) UpdateFileModifyTime(fileName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, f := range s.table {
		if f.FPath == fileName {
			// Обновляем время модификации из файловой системы
			fileInfo, err := os.Stat("/home/user/programmin/obsidianProject/data/obsidianProject/" + fileName)
			if err != nil {
				log.Printf("ошибка получения информации о файле %s: %v", fileName, err)
				return err
			}
			f.ModifyedAt = fileInfo.ModTime()
			f.NeedToConvert = false
			log.Printf("время модификации файла %s обновлено на %v", fileName, fileInfo.ModTime())
			return nil
		}
	}
	return errors.New("файл не найден в таблице")
}

func (s *Postgres) GetConfirmedFiles() ([]*models.File, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.table == nil || len(s.table) == 0 {
		return nil, errors.New("there are no confirmed files")
	}

	log.Println("len confirmed files: ", len(s.table))

	result := make([]*models.File, len(s.table))
	copy(result, s.table)
	return result, nil
}

func (s *Postgres) checkModifyFile(f os.DirEntry, existingFile *models.File) bool {
	op := "postgres.CheckModifyFile"
	info, err := f.Info()
	if err != nil {
		log.Println(op, err)
		return false
	}

	flag := info.ModTime() == existingFile.GetTimeMod()

	existingFile.NeedToConvert = !flag

	return flag
}

func (s *Postgres) checkHandledFileLocked(f os.DirEntry) (*models.File, bool) {
	for _, file := range s.table {
		if file.FPath == f.Name() {
			return file, true
		}
	}

	return nil, false
}

func (s *Postgres) AddPDFFiles(pdfFiles []os.DirEntry, content []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var errs []error

	for _, f := range pdfFiles {
		info, err := f.Info()
		if err != nil {
			errs = append(errs, err)
			continue
		}

		newFile := &models.File{
			FPath:      f.Name(),
			IsPdf:      true,
			ModifyedAt: info.ModTime(),
			PdfContent: content,
		}

		s.table = append(s.table, newFile)
	}

	if len(errs) > 0 {
		return errs[0]
	}

	return nil
}

func (s *Postgres) GetAllFilesName() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, f := range s.table {
		log.Printf("in table %s", f.FPath)
	}
}

func (s *Postgres) GetAllPDFFiles() []*models.File {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.getAllPDFFilesLocked()
}

func (s *Postgres) getAllPDFFilesLocked() []*models.File {
	arr := make([]*models.File, 0)
	for _, f := range s.table {
		log.Println("file:", f.FPath, f.IsPdf, f.ModifyedAt)

		if f.IsPdf {
			arr = append(arr, f)
		}
	}

	return arr
}

func (s *Postgres) ClearConverterFiles() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.converterFiles = make([]*models.File, 0)
}

func (s *Postgres) AddPDFFile(f os.DirEntry, content []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var errs []error

	info, err := f.Info()
	if err != nil {
		errs = append(errs, err)
	}
	if content == nil {
		log.Println("content is nil")
		return nil
	}

	newFile := &models.File{
		FPath:      f.Name(),
		IsPdf:      true,
		ModifyedAt: info.ModTime(),
		PdfContent: content,
	}

	s.table = append(s.table, newFile)

	return nil
}

func (s *Postgres) GetRandomFile() (*models.File, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.table == nil || len(s.table) == 0 {
		return nil, errors.New("there are no confirmed files")
	}

	r := rand.Intn(len(s.table))

	file := s.table[r]

	if file == nil {
		return nil, errors.New("file in storage = nil")
	}
	//зачем так делать если передача изначального файла наверх в прогу это бесплатно
	//pdf := &models.Pdf{
	//	Name:       file.FPath,
	//	PdfContent: file.PdfContent,
	//}

	log.Println("file name:", file.FPath)
	return file, nil
}
