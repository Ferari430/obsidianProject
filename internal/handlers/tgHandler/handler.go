package tgHandler

import (
	"log"

	"github.com/Ferari430/obsidianProject/internal/services/sendService"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TgHandler struct {
	Bot         *tg.BotAPI
	SendService *sendService.SendService
}

func newTgHandler() *TgHandler {
	return &TgHandler{}
}

func (h *TgHandler) SendMessage(msg string) {
	newM := tg.NewMessage(449237834, msg)
	m, err := h.Bot.Send(newM)

	if err != nil {
		log.Fatal(err)
		return
	}

	log.Println("here m:", m.Time())
}
