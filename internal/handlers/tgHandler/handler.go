package tgHandler

import (
	"fmt"
	"log"

	"github.com/Ferari430/obsidianProject/internal/models"
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

func (h *TgHandler) SendMessage(msg string) error {
	newM := tg.NewMessage(449237834, msg)
	m, err := h.Bot.Send(newM)

	if err != nil {
		log.Fatal(err)
		return err
	}

	log.Println("here m:", m.Time())
	return nil
}

func (h *TgHandler) SendPDF(f *models.File) {
	if f.PdfContent == nil {
		log.Println("PDF content is nil!")
	}

	name := fmt.Sprintf("%s.pdf", f.FPath)

	file := tg.FileBytes{
		Name:  name,
		Bytes: f.PdfContent,
	}

	doc := tg.NewDocument(449237834, file)

	m, err := h.Bot.Send(doc)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("here m:", m.Time())

}
