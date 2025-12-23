package checkService

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Ferari430/obsidianProject/internal/repo/inm"
)

type Service struct {
	root string //dir
	//collectedMdFiles    []string
	//alreadyUploadedToDb []string
	db *inm.Postgres
}

//func (s *Service) Add(files []string) error {
//	sliceFiles := make([]*models.File, 0)
//
//	for _, file := range files {
//		sliceFiles = append(sliceFiles, models.NewFile(file))
//	}
//
//	err := s.db.Add(sliceFiles)
//	if err != nil {
//		log.Println("Error adding files:", err)
//	}
//	return nil
//}

func NewCheckService(root string, db *inm.Postgres) *Service {
	//arr1 := make([]string, 0)
	//arr2 := make([]string, 0)
	return &Service{
		root: root,
		db:   db,
		//collectedMdFiles:    arr1,
		//alreadyUploadedToDb: arr2,
	}
}

func (s *Service) RestorePDFFiles() error {
	files := make([]os.DirEntry, 0)
	op := "CronCkecker.CheckFileType"
	allFiles, err := os.ReadDir(s.root)
	if err != nil {
		log.Println(op, err)
		return err
	}
	for _, f := range allFiles {
		if !f.IsDir() && strings.Contains(f.Name(), "pdf") {

			str := strings.Replace(f.Name(), ".pdf", ".md", 1)
			existingFile, err := findFileInDir(s.root, str)
			if err != nil {
				log.Println(op, err)
				return err
			}
			files = append(files, existingFile)
		}
	}

	log.Println("restore files", len(files))

	err = s.db.AddPDFFiles(files)
	if err != nil {
		return err
	}

	return nil
}

/*
  CollectNewMdFiles:
1. Идет в корневую дирректорию
2. Читает все файлы и папки
3. Отбирает файлы с .md
4. Кладет их в collectedMdFiles
5.
*/
//отправляю все файлы md
func (s *Service) CollectNewMdFiles() error {
	collectedMdFiles := make([]os.DirEntry, 0)
	op := "CronCkecker.CollectMdFiles"
	files, err := os.ReadDir(s.root)
	if err != nil {
		return err
	}

	for _, f := range files {
		if !f.IsDir() {
			t := strings.Split(f.Name(), ".")
			if len(t) > 1 && t[1] == "md" {
				log.Println("finded file:", f.Name())
				collectedMdFiles = append(collectedMdFiles, f)
			}
		}
	}

	err = s.db.Add(collectedMdFiles)
	if err != nil {
		log.Println(op, err)
	}
	return nil
}

func findFileInDir(dir, filename string) (os.DirEntry, error) {
	// Читаем содержимое директории
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	// Ищем файл среди содержимого
	for _, f := range files {
		if f.Name() == filename {
			return f, nil
		}
	}

	return nil, fmt.Errorf("file not found")
}
