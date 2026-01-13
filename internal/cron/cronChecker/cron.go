package cronChecker

import (
	"log"
	"log/slog"
	"time"

	"github.com/Ferari430/obsidianProject/internal/services/checkService"
	"github.com/Ferari430/obsidianProject/pkg/logger"
)

type CronChecker struct {
	t      *time.Ticker
	s      *checkService.Service
	signal chan struct{}
	logger *logger.Logger
}

func NewCronChecker(ticker *time.Ticker, srv *checkService.Service, ch chan struct{}, l *logger.Logger) *CronChecker {

	c := &CronChecker{
		t:      ticker,
		s:      srv,
		signal: ch,
		logger: l,
	}
	return c
}

func (c *CronChecker) Run() {
	op := "cronChecker.Run"

	err := c.s.RestorePDFFiles()
	if err != nil {
		c.logger.Debug("cant restore pdf files", slog.String("op", op), slog.String("err", err.Error()))
		return
	}

	log.Println("start cron checker")
	for {
		select {
		case <-c.t.C:
			err := c.s.CollectNewMdFiles()
			if err != nil {
				log.Println("Error collecting new files:", err, op)
			}

		case <-c.signal:
			log.Println("stop cron checker")
			return
		}
	}
}
