package inm

import (
	"log"
	"slices"

	"github.com/Ferari430/obsidianProject/internal/models"
)

type Postgres struct {
	table []*models.File
	arr   []*models.File
}

func NewPostgres() *Postgres {
	arr := make([]*models.File, 0)
	arr1 := make([]*models.File, 0)
	return &Postgres{table: arr,
		arr: arr1,
	}
}

func (s *Postgres) Add(collectedMdFiles []*models.File) error {
	for _, f := range collectedMdFiles {
		if !slices.Contains(s.table, f) {
			s.table = append(s.table, f)
		}
	}

	l := len(collectedMdFiles)
	log.Printf("добавлено %d файлов", l)
	return nil
}

func (s *Postgres) Get() []*models.File {
	return s.table
}
