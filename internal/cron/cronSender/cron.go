package cronSender

import (
	"fmt"
	"log"
	"time"

	"github.com/Ferari430/obsidianProject/internal/handlers/tgHandler"
)

type CronSender struct {
	handler tgHandler.TgHandler
	t       *time.Ticker
}

func NewCronSender(h tgHandler.TgHandler, ticker *time.Ticker) *CronSender {
	return &CronSender{handler: h,
		t: ticker}

}

func (s *CronSender) Start() {
	log.Println("start cron sender")
	i := 0
	for {
		select {
		case <-s.t.C:
			arr := make([]string, 0)
			f, err := s.handler.SendService.Db.GetConfirmedFiles()
			if err != nil {
				log.Println(err)
			}

			for _, file := range f {
				arr = append(arr, file.FPath)
			}

			str := fmt.Sprintf("Reqest number i = %d. Values in storage: %v", i, arr)
			s.handler.SendMessage(str)
			i++

		}
	}
}
