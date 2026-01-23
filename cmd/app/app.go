package app

import (
	"path/filepath"
	"time"

	"github.com/Ferari430/obsidianProject/internal/config"
	"github.com/Ferari430/obsidianProject/internal/cron/cronChecker"
	"github.com/Ferari430/obsidianProject/internal/cron/cronConverter"
	"github.com/Ferari430/obsidianProject/internal/repo/inm"
	"github.com/Ferari430/obsidianProject/internal/services/checkService"
	"github.com/Ferari430/obsidianProject/internal/services/convertService"
	"github.com/Ferari430/obsidianProject/pkg/dirManager"
	"github.com/Ferari430/obsidianProject/pkg/logger"
)

type App struct {
	cronConverter *cronConverter.Cron
	cronChecker   *cronChecker.CronChecker
}

func NewApp() *App {
	postgres := inm.NewPostgres()
	l := logger.NewLogger()

	cfg := config.LoadConfig()

	t1 := time.NewTicker(time.Second * 60) // cronConverter
	srv1 := convertService.NewConvertService(cfg.AppCfg.Root,
		cfg.AppCfg.Sep, cfg.AppCfg.PandocPath, cfg.AppCfg.WkhtmltopdfPdf, postgres, l)
	converter := cronConverter.NewCron(t1, srv1, l)

	t2 := time.NewTicker(time.Second * 5) // cronChecker
	ch := make(chan struct{})

	srv2 := checkService.NewCheckService(cfg.AppCfg.Root, postgres, l)
	checker := cronChecker.NewCronChecker(t2, srv2, ch, l)

	application := &App{
		cronConverter: converter,
		cronChecker:   checker,
	}

	return application
}

func (a *App) Start() {
	cfg := config.LoadConfig()
	mddir := filepath.Join(cfg.AppCfg.Root, "mddir")
	htmldir := filepath.Join(cfg.AppCfg.Root, "htmldir")
	pdfdir := filepath.Join(cfg.AppCfg.Root, "pdfdir")

	allDir := []string{mddir, htmldir, pdfdir}
	dm := dirManager.NewDirManager(allDir)
	dm.Check()

	go a.cronChecker.Run()
	go a.cronConverter.Run()

	time.Sleep(time.Second * 2)
}
