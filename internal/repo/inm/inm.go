package inm

import (
	"errors"
	"log"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

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

	newFilesCount := 0
	modifiedFilesCount := 0
	skippedFilesCount := 0

	for _, f := range collectedMdFiles {
		if existingFile, ok := s.checkHandledFileLocked(f); !ok {
			// Новый файл
			fileInfo, err := f.Info()
			if err != nil {
				return err
			}

			newFile := models.NewFile(f.Name(), f.Name(), fileInfo.ModTime()) // Используем имя файла как FullPath для обратной совместимости
			s.table = append(s.table, newFile)
			s.converterFiles = append(s.converterFiles, newFile)
			newFilesCount++
			log.Printf("[ADD] Новый файл: %s", f.Name())
		} else {
			// Файл уже известен
			timeMatches := s.checkModifyFile(f, existingFile)

			if !timeMatches {
				// Файл изменился
				if !s.isFileInConverterLocked(existingFile.FPath) {
					s.converterFiles = append(s.converterFiles, existingFile)
					modifiedFilesCount++
					log.Printf("[MODIFY] Файл изменился и добавлен в очередь: %s", existingFile.FPath)
				} else {
					log.Printf("[SKIP] Файл уже в очереди конвертации: %s", existingFile.FPath)
				}
			} else {
				skippedFilesCount++
				log.Printf("[SKIP] Файл не изменился: %s", existingFile.FPath)
			}
		}
	}

	log.Printf("[STATS] Новых: %d, Изменено: %d, Пропущено: %d, Всего в очереди: %d",
		newFilesCount, modifiedFilesCount, skippedFilesCount, len(s.converterFiles))
	return nil
}

// AddWithFullPath добавляет MD файлы с полными путями (для рекурсивного сбора)
func (s *Postgres) AddWithFullPath(collectedMdFiles []models.MDFile) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	newFilesCount := 0
	modifiedFilesCount := 0
	skippedFilesCount := 0

	for _, mdFile := range collectedMdFiles {
		// Проверяем, есть ли файл уже в таблице (по полному пути)
		existingFile, ok := s.checkHandledFileByPathLocked(mdFile.FullPath)

		if !ok {
			// Новый файл
			fileInfo, err := mdFile.DirEntry.Info()
			if err != nil {
				return err
			}

			newFile := models.NewFile(mdFile.Name, mdFile.FullPath, fileInfo.ModTime())
			s.table = append(s.table, newFile)
			s.converterFiles = append(s.converterFiles, newFile)
			newFilesCount++
			log.Printf("[ADD] Новый файл: %s (путь: %s)", mdFile.Name, mdFile.FullPath)
		} else {
			// Файл уже известен
			timeMatches := s.checkModifyFileWithPath(mdFile.DirEntry, existingFile)

			if !timeMatches {
				// Файл изменился
				if !s.isFileInConverterLocked(existingFile.FullPath) {
					s.converterFiles = append(s.converterFiles, existingFile)
					modifiedFilesCount++
					log.Printf("[MODIFY] Файл изменился и добавлен в очередь: %s", existingFile.FullPath)
				} else {
					log.Printf("[SKIP] Файл уже в очереди конвертации: %s", existingFile.FullPath)
				}
			} else {
				skippedFilesCount++
				log.Printf("[SKIP] Файл не изменился: %s", existingFile.FullPath)
			}
		}
	}

	log.Printf("[STATS] Новых: %d, Изменено: %d, Пропущено: %d, Всего в очереди: %d",
		newFilesCount, modifiedFilesCount, skippedFilesCount, len(s.converterFiles))
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

func (s *Postgres) UpdateFileModifyTime(fileName string, modifyedAt time.Time) error {
	log.Println(`UpdateFileModifyTime`)
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, f := range s.table {
		// govnokod
		name := strings.Split(fileName, `\`)

		if f.FPath == name[len(name)-1] {

			f.ModifyedAt = modifyedAt
			f.NeedToConvert = false
			log.Printf("время модификации файла %s обновлено на %v", fileName, modifyedAt)
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

	timeMatches := info.ModTime() == existingFile.GetTimeMod()

	existingFile.NeedToConvert = !timeMatches

	return timeMatches
}

// checkModifyFileWithPath проверяет изменения файла (для полных путей)
func (s *Postgres) checkModifyFileWithPath(f os.DirEntry, existingFile *models.File) bool {
	op := "postgres.CheckModifyFileWithPath"
	info, err := f.Info()
	if err != nil {
		log.Println(op, err)
		return false
	}

	timeMatches := info.ModTime() == existingFile.GetTimeMod()

	existingFile.NeedToConvert = !timeMatches

	return timeMatches
}

func (s *Postgres) checkHandledFileLocked(f os.DirEntry) (*models.File, bool) {
	for _, file := range s.table {
		if file.FPath == f.Name() {
			return file, true
		}
	}

	return nil, false
}

// checkHandledFileByPathLocked проверяет, есть ли файл в таблице по полному пути
func (s *Postgres) checkHandledFileByPathLocked(fullPath string) (*models.File, bool) {
	for _, file := range s.table {
		if file.FullPath == fullPath {
			return file, true
		}
	}
	return nil, false
}

// isFileInConverterLocked проверяет, находится ли файл уже в очереди конвертации
// Возвращает true если файл уже в очереди, false если его там нет
func (s *Postgres) isFileInConverterLocked(fileName string) bool {
	for _, file := range s.converterFiles {
		if file.FPath == fileName {
			return true
		}
	}
	return false
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
