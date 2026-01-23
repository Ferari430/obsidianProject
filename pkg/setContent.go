package pkg

import (
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/Ferari430/obsidianProject/internal/models"
	"github.com/Ferari430/obsidianProject/pkg/dirManager"
)

func SetContentToModel(file *models.File, root string) error {
	op := "pkg.SetContentToModel"
	newFilename := file.FPath
	a := dirManager.ReplaceExtension(newFilename, ".md", ".pdf")
	path := filepath.Join(root, a)
	log.Println("path: ", path)
	f, err := os.OpenFile(path, os.O_RDWR, 0666)

	defer func() {
		if err := f.Close(); err != nil {
			log.Println("Error closing file:", err)
		}
	}()

	if err != nil {
		log.Println(op, err)
		return err
	}

	content, err := io.ReadAll(f)
	if err != nil {
		log.Println(op, err)
		return err
	}

	file.SetPdfContent(content)
	log.Println("content setted")
	return nil
}

func GetContentFromFile(fname string, root string) ([]byte, error) {
	op := "pkg.SetContentToModel"
	newFilename := fname
	a := dirManager.ReplaceExtension(newFilename, ".md", ".pdf")
	path := filepath.Join(root, a)
	log.Println("path: ", path)
	f, err := os.OpenFile(path, os.O_RDWR, 0666)

	defer func() {
		if err := f.Close(); err != nil {
			log.Println("Error closing file:", err)
		}
	}()

	if err != nil {
		log.Println(op, err)
		return nil, err
	}

	content, err := io.ReadAll(f)
	if err != nil {
		log.Println(op, err)
		return nil, err
	}

	return content, nil

}
