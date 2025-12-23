package app

import (
	"time"

	"github.com/Ferari430/obsidianProject/internal/cron/cronChecker"
	"github.com/Ferari430/obsidianProject/internal/cron/cronConverter"
	"github.com/Ferari430/obsidianProject/internal/repo/inm"
	"github.com/Ferari430/obsidianProject/internal/services/checkService"
	"github.com/Ferari430/obsidianProject/internal/services/convertService"
	"github.com/Ferari430/obsidianProject/pkg/dirManager"
)

const (
	mddir   = "/home/user/programmin/obsidianProject/data/obsidianProject/mddir"
	htmldir = "/home/user/programmin/obsidianProject/data/obsidianProject/htmldir"
	pdfdir  = "/home/user/programmin/obsidianProject/data/obsidianProject/pdfdir"
)

type app struct {
	cronConverter *cronConverter.Cron
	cronChecker   *cronChecker.CronChecker
}

func NewApp() *app {
	postgres := inm.NewPostgres()

	t1 := time.NewTicker(time.Second * 20) // cronConverter
	srv1 := convertService.NewConvertService(postgres)
	converter := cronConverter.NewCron(t1, srv1)

	t2 := time.NewTicker(time.Second * 10) // cronChecker
	root := "/home/user/programmin/obsidianProject/data/obsidianProject/"
	ch := make(chan struct{})
	srv2 := checkService.NewCheckService(root, postgres)
	checker := cronChecker.NewCronChecker(t2, srv2, ch)
	application := &app{
		cronConverter: converter,
		cronChecker:   checker,
	}
	return application
}

func (a *app) Start() {

	allDir := []string{mddir, htmldir, pdfdir}
	dm := dirManager.NewDirManager(allDir)
	dm.Check()

	go a.cronChecker.Run()
	time.Sleep(time.Second * 2)
	go a.cronConverter.Run()
}
