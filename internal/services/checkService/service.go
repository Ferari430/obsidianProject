package checkService

import (
	"log"
	"os"
	"slices"
	"strings"

	"github.com/Ferari430/obsidianProject/internal/repo/inm"
)

type Service struct {
	root                string //dir
	collectedMdFiles    []string
	alreadyUploadedToDb []string
	db                  *inm.Postgres
}

func (s *Service) Add(files []string) error {
	err := s.db.Add(files)
	if err != nil {
		log.Println("Error adding files:", err)
	}
	return nil
}

func NewCheckService(root string, db *inm.Postgres) *Service {
	arr1 := make([]string, 0)
	arr2 := make([]string, 0)
	return &Service{
		root:                root,
		db:                  db,
		collectedMdFiles:    arr1,
		alreadyUploadedToDb: arr2,
	}
}

// todo: not used
func (s *Service) CheckFileType() {
	op := "CronCkecker.CheckFileType"
	files, err := os.ReadDir(s.root)
	if err != nil {
		log.Println(op, err)
	}
	for _, f := range files {
		log.Println(f)
	}
}

/*
  CollectNewMdFiles:
1. Идет в корневую дирректорию
2. Читает все файлы и папки
3. Отбирает файлы с .md
4. Кладет их в collectedMdFiles
5.
*/

func (s *Service) CollectNewMdFiles() ([]string, error) {
	collectedMdFiles := make([]string, 0)
	op := "CronCkecker.CollectMdFiles"
	files, err := os.ReadDir(s.root)
	if err != nil {
		log.Println(op, err)
		return nil, err
	}

	for _, f := range files {
		if !f.IsDir() {
			if !slices.Contains(s.alreadyUploadedToDb, f.Name()) {
				t := strings.Split(f.Name(), ".")
				if len(t) > 1 && t[1] == "md" {
					log.Println("finded file:", f.Name())

					collectedMdFiles = append(collectedMdFiles, f.Name())

					s.alreadyUploadedToDb = append(s.alreadyUploadedToDb, f.Name())

					err := s.Add(collectedMdFiles)
					if err != nil {
						log.Println("Error adding file:", err)
					}
				}
			}
		}
	}

	// Обновляем собранные файлы
	s.collectedMdFiles = collectedMdFiles

	// Логируем обработанные файлы
	log.Println("unHandled md files:")
	for _, f := range collectedMdFiles {
		log.Println(f)
	}

	// Возвращаем собранные файлы
	return collectedMdFiles, nil
}

// todo: отправка файлов по запросу через кнопку тг
func (s *Service) GetMdFiles() []string {
	if len(s.collectedMdFiles) == 0 {
		log.Println("mdfiles not collected yet")
		return nil
	}

	files := s.collectedMdFiles
	return files
}
