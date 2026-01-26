package cronConverter

import (
	"log/slog"
	"time"

	"github.com/Ferari430/obsidianProject/internal/models"
	"github.com/Ferari430/obsidianProject/internal/services/convertService"
	"github.com/Ferari430/obsidianProject/pkg/logger"
)

type Cron struct {
	t      *time.Ticker
	srv    *convertService.ConvertService
	logger *logger.Logger
	stopCh chan struct{}
}

func NewCron(ticker *time.Ticker, s *convertService.ConvertService, l *logger.Logger) *Cron {
	cron := &Cron{
		t:      ticker,
		srv:    s,
		logger: l,
		stopCh: make(chan struct{}),
	}
	return cron
}

func (c *Cron) GetFiles() []*models.File {
	return c.srv.GetFiles()
}

// HandleFiles обрабатывает файлы для конвертации
func (c *Cron) HandleFiles() error {
	op := "cronConverter.HandleFiles"

	// Получаем файлы через сервис
	mdFiles := c.srv.GetFiles()
	if len(mdFiles) == 0 {
		c.logger.Info("No files to convert", slog.String("op", op))
		return nil
	}

	c.logger.Info("Files to convert",
		slog.String("op", op),
		slog.Int("count", len(mdFiles)))

	// Обрабатываем каждый файл
	for _, mdFile := range mdFiles {
		c.logger.Info("Processing file",
			slog.String("op", op),
			slog.String("path", mdFile.FPath),
			slog.Bool("isPDF", mdFile.IsPdf),
			slog.Bool("needConvert", mdFile.NeedToConvert))

		// Проверяем нужно ли конвертировать
		if !mdFile.IsPdf || mdFile.NeedToConvert {
			c.logger.Info("Starting conversion",
				slog.String("op", op),
				slog.String("file", mdFile.FPath))

			// Конвертируем MD в PDF
			c.srv.ConvertMdToPDF(mdFile.FPath)

			c.logger.Info("Conversion completed",
				slog.String("op", op),
				slog.String("file", mdFile.FPath))
		} else {
			c.logger.Debug("Skipping file - no conversion needed",
				slog.String("op", op),
				slog.String("file", mdFile.FPath))
		}
	}

	return nil
}

func (c *Cron) Run() {
	op := "cronConverter.Run"
	c.logger.Info("Starting cron converter service", slog.String("op", op))

	// Даем время на инициализацию других сервисов
	time.Sleep(2 * time.Second)

	for {
		select {
		case <-c.t.C:
			c.logger.Debug("Cron converter tick",
				slog.String("op", op),
				slog.Time("time", time.Now()))

			// Используем новый метод вместо прямого вызова GetFiles
			if err := c.HandleFiles(); err != nil {
				c.logger.Error("Error handling files",
					slog.String("op", op),
					slog.String("error", err.Error()))
			}

		case <-c.stopCh:
			c.logger.Info("Cron converter stopped",
				slog.String("op", op))
			c.t.Stop()
			return
		}
	}
}

// Stop останавливает крон
func (c *Cron) Stop() {
	op := "cronConverter.Stop"
	c.logger.Info("Stopping cron converter", slog.String("op", op))
	close(c.stopCh)
}

// RunOnce запускает обработку файлов один раз (для тестов)
func (c *Cron) RunOnce() error {
	op := "cronConverter.RunOnce"
	c.logger.Info("Running converter once", slog.String("op", op))
	return c.HandleFiles()
}
