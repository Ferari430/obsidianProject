package inm

import (
	"log"
	"slices"
)

type Postgres struct {
	table []string
}

func NewPostgres() *Postgres {
	arr := make([]string, 0)
	return &Postgres{table: arr}
}

func (s *Postgres) Add(collectedMdFiles []string) error {

	for _, f := range collectedMdFiles {
		if !slices.Contains(s.table, f) {
			s.table = append(s.table, f)
		}
	}
	l := len(collectedMdFiles)
	log.Printf("добавлено %d файлов", l)
	return nil
}

func (s *Postgres) Get() []string {
	return s.table
}
