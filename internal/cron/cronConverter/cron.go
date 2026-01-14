package cronConverter

import (
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

func (c *Cron) Run() {
	log.Println("start cron converter")
	time.Sleep(time.Second)

	for range c.t.C {

		log.Println("tick converter")
		// переписать. заменить c.srv.GetFiles() на c.srv.handleFiles чтобы не  было зависимости от репозитория
		mdFiles := c.srv.GetFiles()
		log.Println("len mdFiles: ", len(mdFiles))
		for _, mdFile := range mdFiles {
			log.Println("mdFile in coverter:", mdFile.FPath, mdFile.IsPdf)
			if !mdFile.IsPdf || mdFile.NeedToConvert {
				log.Println("конвертация", mdFile.FPath)
				c.srv.ConvertMdToPDF(mdFile.FPath)

				// fName = c.srv.ReplaceExtension(fName, ".md", ".html")
				// log.Println("new fname:", fName)
				// path := fmt.Sprintf("%s%s", out, mdFile.FPath)
				// htmlOut := fmt.Sprintf("%s%s", out, fName)
				// log.Println("fileNames:", path, htmlOut)

				// time.Sleep(time.Millisecond * 350)
				// fName = c.srv.ReplaceExtension(fName, ".html", ".pdf")
				// log.Println("new fname:", fName)
				// pdfOut := fmt.Sprintf("%s/%s", out, fName)
				// err = c.srv.ConvertHTMLToPDF(htmlOut, pdfOut)
				// if err != nil {
				// 	log.Println(err)
				// 	return
				// }

				// mdFile.IsPdf = true

				// err = pkg.SetContentToModel(mdFile)
				// if err != nil {
				// 	return
				// }

				// log.Println("file processing finished, filename:", fName)
				// log.Println("----------------------------------")

				// // Обновляем время модификации и флаг конвертации
				// err = c.srv.UpdateFileModifyTime(mdFile.FPath)
				// if err != nil {
				// 	log.Println(err)
				// }

				// // Удаляем обработанный файл из очереди конвертации
				// c.srv.RemoveFromConverter(mdFile.FPath)
			}
		}

	}
}
