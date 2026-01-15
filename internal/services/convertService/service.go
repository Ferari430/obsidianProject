package convertService

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Ferari430/obsidianProject/internal/models"
	"github.com/Ferari430/obsidianProject/internal/repo/inm"
	"github.com/Ferari430/obsidianProject/pkg/dirManager"
	"github.com/Ferari430/obsidianProject/pkg/logger"
)

type ConvertService struct {
	db              *inm.Postgres
	logger          *logger.Logger
	Path            string
	Sep             string
	PandocPath      string
	WkhtmltopdfPath string
}

func NewConvertService(p, sep, pandocPath, wkhtmltopdf string, db *inm.Postgres, l *logger.Logger) *ConvertService {
	return &ConvertService{
		db:              db,
		logger:          l,
		Path:            p,
		Sep:             sep,
		PandocPath:      pandocPath,
		WkhtmltopdfPath: wkhtmltopdf,
	}
}

func (c *ConvertService) GetFiles() []*models.File {
	op := "convertService.GetFiles"
	arr := c.db.Get()
	if len(arr) == 0 {
		c.logger.Debug("no new file for convertService", slog.String("op", op))
		return nil
	}

	log.Printf("Крон Конвертер получил  %d файлов", len(arr))
	return arr
}

func (c *ConvertService) RemoveFromConverter(fileName string) error {
	return c.db.RemoveFromConverter(fileName)
}

func (c *ConvertService) UpdateFileModifyTime(fileName string) error {
	log.Println("filename:", fileName)
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		log.Printf("---------%s----------", fileName)
		log.Printf("ошибка получения информации о файле %s: %v", fileName, err)
		return err
	}

	return c.db.UpdateFileModifyTime(fileName, fileInfo.ModTime())
}

func (c *ConvertService) SearchPictureName(path string) {

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

		if strings.Contains(line, ".png") {

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

func (c *ConvertService) ConvertMDToHTML(inputFile, outputFile string) error {
	cmd := exec.Command(c.PandocPath, inputFile, "-o", outputFile, "--from=markdown+hard_line_breaks")

	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("Error executing pandoc: %v\nOutput: %s", err, output)
	}

	fmt.Println("Conversion completed successfully.")
	return nil
}

func (c *ConvertService) prepareFilePath(fileName string) {

}

func (c *ConvertService) ConvertMdToPDF(fileName string) {
	op := "convertService.convertMdToPDF"
	md, html, pdf := c.getAbsPath(fileName)

	c.SearchPictureName(md)
	log.Printf("имя для html: %s\nимя для pdf:%s", html, pdf)
	err := c.ConvertMDToHTML(md, html)
	if err != nil {
		c.logger.Error("cant convert md file to html", slog.String("op", op), slog.String("error:", err.Error()))
	}
	err = c.ConvertHTMLToPDF(html, pdf)
	if err != nil {
		c.logger.Error("cant convert html file to pdf", slog.String("op", op), slog.String("error:", err.Error()))
	}

	err = c.UpdateFileModifyTime(md)
	if err != nil {
		c.logger.Error("cant update file time modification", slog.String("op", op), slog.String("error:", err.Error()))
	}

	err = c.RemoveFromConverter(fileName)
	if err != nil {
		c.logger.Error("cant delete file from converter slice", slog.String("op", op), slog.String("error:", err.Error()))
	}
	if err == nil {
		log.Printf("конвертация файла %s прошла успешно", fileName)
	}

}

func (c *ConvertService) getAbsPath(fileName string) (string, string, string) {
	absFileName := c.Path + c.Sep + fileName
	log.Printf("из имени %s сделали %s путь", fileName, absFileName)
	html := c.ReplaceExtension(absFileName, ".md", ".html")
	pdf := c.ReplaceExtension(absFileName, ".md", ".pdf")

	return absFileName, html, pdf
}

func (c *ConvertService) ConvertHTMLToPDF(inputFile, outputFile string) error {
	cmd := exec.Command(c.WkhtmltopdfPath, "--enable-local-file-access", "--encoding", "UTF-8", inputFile, outputFile)

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
	log.Println("no new file for convertService")
	return result
}

func completeFileName(s string) string {
	return fmt.Sprintf("(./%s)", s)
}

func (c *ConvertService) ReplaceExtension(s string, oldExt, newExt string) string {
	// text.md --> text.html
	return strings.Replace(s, oldExt, newExt, -1)
}

func (c *ConvertService) SetContentToModel(file *models.File) error {
	op := "convertService.SetContentToModel"
	newFilename := file.FPath
	a := dirManager.ReplaceExtension(newFilename, ".md", ".pdf")
	path := filepath.Join(c.Path, a)
	log.Println("path: ", path)
	f, err := os.OpenFile(path, os.O_RDWR, 0666)

	defer func() {
		if err := f.Close(); err != nil {
			log.Println("Error closing file:", err)
		}
	}()

	if err != nil {
		log.Println(op, err)
		return err
	}

	content, err := io.ReadAll(f)
	if err != nil {
		log.Println(op, err)
		return err
	}

	file.SetPdfContent(content)
	log.Println("content setted")
	return nil
}
