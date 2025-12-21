package cronConverter

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Cron struct {
	t *time.Ticker
}

func NewCron(ticker *time.Ticker) *Cron {
	cron := &Cron{t: ticker}
	return cron
}

func (c *Cron) Run(mdFiles []string) {
	log.Println("start cron")
	n := 0

	out := "/home/user/programmin/obsidianProject/data/obsidianProject/"
	for n != len(mdFiles) {
		for _, mdFile := range mdFiles {
			select {
			case <-c.t.C:
				fName := filepath.Base(mdFile)
				searchPictureName(mdFile)
				fName = replaceExtension(fName, ".md", ".html")
				log.Println("new fname:", fName)
				htmlOut := fmt.Sprintf("%s/%s", out, fName)
				err := convertMDToHTML(mdFile, htmlOut)
				if err != nil {
					log.Println(err)
				}

				time.Sleep(time.Second * 1)
				fName = replaceExtension(fName, ".html", ".pdf")
				log.Println("new fname:", fName)
				pdfOut := fmt.Sprintf("%s/%s", out, fName)
				err = convertHTMLToPDF(htmlOut, pdfOut)

				n++
				log.Println("file processing finished, filename:", fName)
				log.Println("----------------------------------")
			}
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

func searchPictureName(path string) {
	// Открываем файл для чтения и записи
	f, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		log.Println("Error opening file:", err)
		return
	}

	defer func() {
		err := f.Close() // Закрываем файл в конце
		if err != nil {
			log.Println("Error closing file:", err)
		}
	}()

	// Сканер для построчного чтения файла
	scanner := bufio.NewScanner(f)

	// Мы будем использовать буфер для сохранения строк, чтобы потом их переписать с добавленным текстом
	var lines []string

	// Читаем файл построчно
	for scanner.Scan() {
		line := scanner.Text()

		// Если строка содержит картинку (например, "Screenshot")
		if strings.Contains(line, ".png") {
			//log.Println("picture name:", line)

			if strings.Count(line, ".png") >= 2 {
				log.Println("already contains")
				return
			}

			fname := extractFileName(line)
			cfn := completeFileName(fname)
			line += cfn
			log.Println("picture name:", cfn)
		}
		// Добавляем обработанную строку в буфер
		lines = append(lines, line)
	}

	// Проверяем на ошибки при сканировании
	if err := scanner.Err(); err != nil {
		log.Println("Error scanning file:", err)
		return
	}

	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		log.Println("Error seeking to start:", err)
		return
	}

	for _, line := range lines {
		_, err := f.WriteString(line + "\n") // Добавляем новую строку с концом
		if err != nil {
			log.Println("Error writing to file:", err)
			return
		}
	}

	log.Println("File updated successfully.")
}

func convertMDToHTML(inputFile, outputFile string) error {
	cmd := exec.Command("pandoc", inputFile, "-o", outputFile)

	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("Error executing pandoc: %v\nOutput: %s", err, output)
	}

	fmt.Println("Conversion completed successfully.")
	return nil
}

func convertHTMLToPDF(inputFile, outputFile string) error {
	cmd := exec.Command("wkhtmltopdf", "--enable-local-file-access", "--encoding", "UTF-8", inputFile, outputFile)

	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("Error executing wkhtmltopdf: %v\nOutput: %s", err, output)
	}

	fmt.Println("HTML to PDF conversion completed successfully.")
	return nil
}

func extractFileName(s string) string {
	result := ""
	for _, val := range s {
		if string(val) != "!" && string(val) != "[" && string(val) != "]" {
			result += string(val)
		}
	}
	return result
}

func completeFileName(s string) string {
	return fmt.Sprintf("(./%s)", s)
}

func replaceExtension(s string, oldext, newext string) string {
	// text.md --> text.html
	return strings.Replace(s, oldext, newext, -1)

}
