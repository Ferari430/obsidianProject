package cronSender

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
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

func (s *CronSender) SendAllFiles() {
	log.Println("start cron sender")
	for {
		select {
		case <-s.t.C:
			log.Println("sending files:...")
			arr := make([]string, 0)
			files, err := s.handler.SendService.Db.GetConfirmedFiles()
			if err != nil {
				log.Println(err)
			}
			wg := sync.WaitGroup{}
			wg.Add(len(files))

			for _, file := range files {

				go func() {
					defer wg.Done()
					str := fmt.Sprintf("Reqest number i = %d. Values in storage: %v", rand.Intn(9), arr)
					s.handler.SendMessage(str)
					s.handler.SendPDF(file)

				}()
			}
			log.Println("sendind files confirmed")
			wg.Wait()
		}
	}
}

func (s *CronSender) Start() {
	log.Println("start cron sender")
	for {
		select {
		case <-s.t.C:
			log.Println("sending file:...")
			file, err := s.handler.SendService.GetRandomPdf()
			if err != nil {
				log.Fatal(err)
			}

			str := fmt.Sprint("New File:")
			err = s.handler.SendMessage(str)
			if err != nil {
				log.Fatal(err)
			}

			s.handler.SendPDF(file)

			log.Println("sendind files confirmed")
		}
	}
}
