package checkService

import (
	"fmt"
	"log"
	"os"
	"strings"

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
		content, err := pkg.GetContentFromFile(f.Name())
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
	var collectedMdFiles []os.DirEntry

	files, err := os.ReadDir(s.root)
	if err != nil {
		log.Println(op, err)
		return err
	}
	log.Println("files:", files)

	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".md") {
			log.Println("Found MD file:", f.Name())
			collectedMdFiles = append(collectedMdFiles, f)
		}
	}

	if err := s.db.Add(collectedMdFiles); err != nil {
		log.Println(op, err)
		return err
	}

	return nil
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
