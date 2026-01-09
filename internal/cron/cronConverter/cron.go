package cronConverter

import (
	"fmt"
	"log"
	"time"

	"github.com/Ferari430/obsidianProject/internal/models"
	"github.com/Ferari430/obsidianProject/internal/services/convertService"
	"github.com/Ferari430/obsidianProject/pkg/logger"
)

type Cron struct {
	t      *time.Ticker
	srv    *convertService.ConvertService
	logger *logger.Logger
}

func NewCron(ticker *time.Ticker, s *convertService.ConvertService, l *logger.Logger) *Cron {
	cron := &Cron{t: ticker,
		srv:    s,
		logger: l,
	}
	return cron
}

func (c *Cron) GetFiles() []*models.File {
	return c.srv.GetFiles()
}

// TODO: доработать, логика с for не нравится
func (c *Cron) Run() {
	log.Println("start cron converter")
	time.Sleep(time.Second)
	out := "/home/user/programmin/obsidianProject/data/obsidianProject/"
	for {
		select {
		case <-c.t.C:
			log.Println("tick converter")
			mdFiles := c.srv.GetFiles()
			log.Println("len mdFiles: ", len(mdFiles))
			for _, mdFile := range mdFiles {
				log.Println("mdFile in coverter:", mdFile.FPath, mdFile.IsPdf)
				if !mdFile.IsPdf || mdFile.NeedToConvert {
					log.Println("конвертация", mdFile.FPath)
					fName := mdFile.FPath
					c.srv.SearchPictureName(mdFile.FPath)
					fName = c.srv.ReplaceExtension(fName, ".md", ".html")
					log.Println("new fname:", fName)
					path := fmt.Sprintf("%s%s", out, mdFile.FPath)
					htmlOut := fmt.Sprintf("%s%s", out, fName)
					log.Println("fileNames:", path, htmlOut)
					err := c.srv.ConvertMDToHTML(path, htmlOut)
					if err != nil {
						log.Println(err)
					}
					time.Sleep(time.Millisecond * 350)
					fName = c.srv.ReplaceExtension(fName, ".html", ".pdf")
					log.Println("new fname:", fName)
					pdfOut := fmt.Sprintf("%s/%s", out, fName)
					err = c.srv.ConvertHTMLToPDF(htmlOut, pdfOut)
					if err != nil {
						log.Println(err)
						return
					}
					mdFile.IsPdf = true
					log.Println("file processing finished, filename:", fName)
					log.Println("----------------------------------")
				}
				log.Println("mdFile already exists in pdf")
			}
		}
	}
}
