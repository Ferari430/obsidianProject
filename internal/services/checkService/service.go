package checkService

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Ferari430/obsidianProject/internal/models"
	"github.com/Ferari430/obsidianProject/internal/repo/inm"
	"github.com/Ferari430/obsidianProject/pkg"
	"github.com/Ferari430/obsidianProject/pkg/logger"
)

type Service struct {
	root   string //dir
	db     *inm.Postgres
	logger *logger.Logger
}

func NewCheckService(root string, db *inm.Postgres, l *logger.Logger) *Service {
	return &Service{
		root:   root,
		db:     db,
		logger: l,
	}

}

func (s *Service) RestorePDFFiles() error {
	op := "CronChecker.RestorePDFFiles"
	var files []os.DirEntry
	log.Println("ROOT=", s.root)
	allFiles, err := os.ReadDir(s.root)
	if err != nil {
		log.Println(op, err)
		return err
	}

	for _, f := range allFiles {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".pdf") {
			mdFileName := strings.Replace(f.Name(), ".pdf", ".md", 1)
			mdFile, err := findFileInDir(s.root, mdFileName)
			if err != nil {
				log.Println(op, err)
				return err
			}
			files = append(files, mdFile)
		}
	}

	log.Println("Restored files:", len(files))
	for _, f := range files {
		content, err := pkg.GetContentFromFile(f.Name(), s.root)
		if err != nil {
			log.Println(op, err)
			return err
		}

		err = s.db.AddPDFFile(f, content)
		if err != nil {
			log.Println(op, err)
			return err
		}
	}

	return nil
}

func (s *Service) CollectNewMdFiles() error {
	op := "CronChecker.CollectMdFiles"

	// Рекурсивно собираем все MD файлы со всех подпапок
	collectedMdFiles, err := s.collectMdFilesRecursive(s.root)
	if err != nil {
		log.Println(op, err)
		return err
	}

	log.Printf("[COLLECTOR] Найдено MD файлов: %d", len(collectedMdFiles))

	if err := s.db.AddWithFullPath(collectedMdFiles); err != nil {
		log.Println(op, err)
		return err
	}

	return nil
}

// collectMdFilesRecursive рекурсивно собирает все MD файлы из корневой папки и подпапок
func (s *Service) collectMdFilesRecursive(dir string) ([]models.MDFile, error) {
	var mdFiles []models.MDFile

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		if f.IsDir() {
			// Рекурсивно обрабатываем подпапки
			subPath := filepath.Join(dir, f.Name())
			subMdFiles, err := s.collectMdFilesRecursive(subPath)
			if err != nil {
				log.Printf("[WARNING] Ошибка при чтении папки %s: %v", subPath, err)
				continue // Продолжаем даже если была ошибка в одной из подпапок
			}
			mdFiles = append(mdFiles, subMdFiles...)
		} else if strings.HasSuffix(f.Name(), ".md") {
			fullPath := filepath.Join(dir, f.Name())
			log.Printf("[FOUND] MD файл: %s", fullPath)
			mdFiles = append(mdFiles, models.MDFile{
				Name:     f.Name(),
				FullPath: fullPath,
				DirEntry: f,
			})
		}
	}

	return mdFiles, nil
}

func findFileInDir(dir, filename string) (os.DirEntry, error) {
	files, err := os.ReadDir(dir)
	if err != nil {

		return nil, err
	}

	log.Println("current files: ", files)
	for _, f := range files {
		if f.Name() == filename {
			return f, nil
		}
	}

	return nil, fmt.Errorf("file not found")
}
