package cronConverter

import (
	"fmt"
	"log"
	"time"

	"github.com/Ferari430/obsidianProject/internal/models"
	"github.com/Ferari430/obsidianProject/internal/services/convertService"
)

type Cron struct {
	t   *time.Ticker
	srv *convertService.ConvertService
}

func NewCron(ticker *time.Ticker, s *convertService.ConvertService) *Cron {
	cron := &Cron{t: ticker,
		srv: s}
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
				if !mdFile.IsPdf {
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

//func OpenMDFile(path string) (*os.File, error) {
//	f, err := os.OpenFile(path, os.O_RDONLY, 0666)
//	if err != nil {
//		return nil, err
//	}
//	defer f.Close()
//
//	fileInfo, err := f.Stat()
//	if err != nil {
//		return nil, err
//	}
//
//	buf := make([]byte, fileInfo.Size())
//
//	_, err = f.Read(buf)
//	if err != nil {
//		return nil, err
//	}
//
//	log.Println(string(buf))
//	return f, nil
//}
//
//func convert(inputFile string) error {
//	OutputFile := "output.pdf"
//	cmd := exec.Command("pandoc", inputFile, "-o", OutputFile,
//		"--pdf-engine=xelatex",
//		"-V", "mainfont=DejaVu Sans",
//		"-V", "monofont=DejaVu Sans Mono")
//
//	output, err := cmd.CombinedOutput()
//	if err != nil {
//		log.Println(string(output))
//		return err
//	}
//
//	filename := strings.Split(inputFile, "/")
//	name := filename[len(filename)-1]
//	log.Printf("convertation from %s to %s successfully", name, OutputFile)
//
//	return nil
//}
