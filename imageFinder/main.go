package main

import (
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"strings"

	"github.com/Ferari430/obsidianProject/imageFinder/mdFinder"
)

func main() {
	root := `B:\programmin-20260114T065921Z-1-001\programmin\obsidianProject\datanew\test`
	i := NewImageFinder(root)

	err := i.ScanFiles()
	if err != nil {
		return
	}
	log.Println("image files:")
	for _, val := range i.data {
		fmt.Println(val)
	}

	fmt.Print("\n")

	log.Println("md files:")
	mdfinder := mdFinder.NewMDFinder(root, ".md", i)
	err = mdfinder.ScanFiles()
	if err != nil {
		return
	}

	for _, val := range mdfinder.Data {
		err := mdfinder.ProcessFile(val)
		if err != nil {
			return
		}
		log.Println(val)
	}
	log.Println("succsses")
}

type ImageFinder struct {
	root string
	data map[string]string
}

func NewImageFinder(root string) *ImageFinder {
	return &ImageFinder{
		root: root,

		data: make(map[string]string),
	}
}

func (i *ImageFinder) GetInfo() map[string]string {
	return i.data
}

func (f *ImageFinder) ScanFiles() error {
	extensions := []string{".png", ".jpg", ".jpeg", ".gif", ".bmp", ".svg", ".webp"}

	err := filepath.WalkDir(f.root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Пропускаем ошибки доступа
		}

		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		for _, imageExt := range extensions {
			if ext == imageExt {
				filename := filepath.Base(path)

				if _, exists := f.data[filename]; !exists {
					f.data[filename] = path
				}
			}
		}
		return nil
	})
	return err
}
