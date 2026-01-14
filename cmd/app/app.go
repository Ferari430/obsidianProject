package app

import (
	"log"
	"time"

	"github.com/Ferari430/obsidianProject/internal/config"
	"github.com/Ferari430/obsidianProject/internal/cron/cronChecker"
	"github.com/Ferari430/obsidianProject/internal/cron/cronConverter"
	"github.com/Ferari430/obsidianProject/internal/cron/cronSender"
	"github.com/Ferari430/obsidianProject/internal/repo/inm"
	"github.com/Ferari430/obsidianProject/internal/services/checkService"
	"github.com/Ferari430/obsidianProject/internal/services/convertService"
	"github.com/Ferari430/obsidianProject/pkg/dirManager"
	"github.com/Ferari430/obsidianProject/pkg/logger"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	mddir   = "/home/user/programmin/obsidianProject/data/obsidianProject/mddir"
	htmldir = "/home/user/programmin/obsidianProject/data/obsidianProject/htmldir"
	pdfdir  = "/home/user/programmin/obsidianProject/data/obsidianProject/pdfdir"
)

type App struct {
	cronConverter *cronConverter.Cron
	cronChecker   *cronChecker.CronChecker
	cronSender    *cronSender.CronSender
}

func NewApp() *App {
	postgres := inm.NewPostgres()
	l := logger.NewLogger()

	cfg := config.LoadConfig()

	// bot, err := initTg(cfg.Tg)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	t1 := time.NewTicker(time.Second * 9) // cronConverter
	srv1 := convertService.NewConvertService(cfg.AppCfg.Root,
		cfg.AppCfg.Sep, cfg.AppCfg.PandocPath, cfg.AppCfg.WkhtmltopdfPdf, postgres, l)
	converter := cronConverter.NewCron(t1, srv1, l)

	t2 := time.NewTicker(time.Second * 5) // cronChecker
	ch := make(chan struct{})

	srv2 := checkService.NewCheckService(cfg.AppCfg.Root, postgres, l)
	checker := cronChecker.NewCronChecker(t2, srv2, ch, l)

	// t3 := time.NewTicker(time.Second * 60)

	// sendS := sendService.NewSendService(postgres)
	// tgh := tgHandler.TgHandler{Bot: bot,
	// 	SendService: sendS,
	// }
	// cS := cronSender.NewCronSender(tgh, t3)

	application := &App{
		cronConverter: converter,
		cronChecker:   checker,
		// cronSender:    cS,
	}
	return application
}

func (a *App) Start() {
	//go a.cronSender.SendAllFiles()
	allDir := []string{mddir, htmldir, pdfdir}
	dm := dirManager.NewDirManager(allDir)
	dm.Check()

	// go a.cronSender.Start()

	go a.cronChecker.Run()
	time.Sleep(time.Second * 2)
	go a.cronConverter.Run()
}

func initTg(cfg config.TgBotCfg) (*tg.BotAPI, error) {

	bot, err := tg.NewBotAPI(cfg.Token)
	if err != nil {
		log.Println("err")
		log.Println(err)
		return nil, err
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	//
	//u := tg.NewUpdate(0)
	//u.Timeout = 60
	//
	//updates := bot.GetUpdatesChan(u)
	//
	//for update := range updates {
	//	log.Println(update.Message.Text)
	//}

	return bot, nil
}
