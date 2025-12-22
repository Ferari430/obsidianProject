package cronConverter

import (
	"fmt"
	"log"
	"time"

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

func (cron *Cron) GetFiles() []string {
	return cron.srv.GetFiles()
}

// TODO: доработать, логика с for не нравится
func (c *Cron) Run() {
	log.Println("start cron converter")
	n := 0
	time.Sleep(time.Second)
	mdFiles := c.GetFiles()
	log.Println("mdFiles:", mdFiles)
	out := "/home/user/programmin/obsidianProject/data/obsidianProject/"
	for _, mdFile := range mdFiles {
		select {
		case <-c.t.C:
			log.Println("tick converter")
			//fName := filepath.Base(mdFile)
			fName := mdFile
			c.srv.SearchPictureName(mdFile)
			fName = c.srv.ReplaceExtension(fName, ".md", ".html")
			log.Println("new fname:", fName)
			mdFile := fmt.Sprintf("%s%s", out, mdFile)
			htmlOut := fmt.Sprintf("%s%s", out, fName)
			log.Println("fileNames:", mdFile, htmlOut)
			err := c.srv.ConvertMDToHTML(mdFile, htmlOut)
			if err != nil {
				log.Println(err)
			}
			time.Sleep(time.Millisecond * 350)
			fName = c.srv.ReplaceExtension(fName, ".html", ".pdf")
			log.Println("new fname:", fName)
			pdfOut := fmt.Sprintf("%s/%s", out, fName)
			err = c.srv.ConvertHTMLToPDF(htmlOut, pdfOut)

			n++
			log.Println("file processing finished, filename:", fName)
			log.Println("----------------------------------")
		}
	}
	log.Println("cron finished work")
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
