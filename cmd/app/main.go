package app

import (
	"time"

	"github.com/Ferari430/obsidianProject/internal/services/cron/cronConverter"
	"github.com/Ferari430/obsidianProject/pkg/dirManager"
)

const (
	mddir   = "/home/user/programmin/obsidianProject/data/obsidianProject/mddir"
	htmldir = "/home/user/programmin/obsidianProject/data/obsidianProject/htmldir"
	pdfdir  = "/home/user/programmin/obsidianProject/data/obsidianProject/pdfdir"
)

type app struct {
	cron *cronConverter.Cron
}

func NewApp() *app {
	t := time.NewTicker(time.Second * 2)
	cron := cronConverter.NewCron(t)
	app := &app{
		cron: cron,
	}
	return app
}

func (a *app) Start() {
	mdFiles := []string{
		"/home/user/programmin/obsidianProject/data/obsidianProject/testmd.md",
		"/home/user/programmin/obsidianProject/data/obsidianProject/zagolovok.md",
	}
	allDir := []string{mddir, htmldir, pdfdir}
	dm := dirManager.NewDirManager(allDir)
	dm.Check()
	a.cron.Run(mdFiles)
}
