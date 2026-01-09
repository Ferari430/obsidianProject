package sendService

import (
	"github.com/Ferari430/obsidianProject/internal/models"
	"github.com/Ferari430/obsidianProject/internal/repo/inm"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type SendService struct {
	Db *inm.Postgres
}

func NewSendService(storage *inm.Postgres) *SendService {
	return &SendService{
		Db: storage,
	}
}

func (s *SendService) GetFile(message *tg.Message) ([]*models.File, error) {

	files, err := s.Db.GetConfirmedFiles()
	if err != nil {
		return nil, err
	}
	_ = files
	return files, nil
}
