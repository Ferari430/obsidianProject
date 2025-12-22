package cronChecker

import (
	"log"
	"time"

	"github.com/Ferari430/obsidianProject/internal/services/checkService"
)

type CronChecker struct {
	t      *time.Ticker
	s      *checkService.Service
	signal chan struct{}
}

func NewCronChecker(ticker *time.Ticker, srv *checkService.Service, ch chan struct{}) *CronChecker {

	c := &CronChecker{
		t:      ticker,
		s:      srv,
		signal: ch,
	}
	return c
}

func (c *CronChecker) Run() {
	log.Println("start cron checker")
	for {
		select {
		case <-c.t.C:
			_, err := c.s.CollectNewMdFiles()
			if err != nil {
				log.Println("Error collecting new files:", err)
			}

		case <-c.signal:
			log.Println("stop cron checker")
			return
		}
	}
}
