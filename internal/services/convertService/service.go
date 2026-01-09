package convertService

import "C"
import (
	"bufio"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"github.com/Ferari430/obsidianProject/internal/models"
	"github.com/Ferari430/obsidianProject/internal/repo/inm"
	"github.com/Ferari430/obsidianProject/pkg/logger"
)

type ConvertService struct {
	db     *inm.Postgres
	logger *logger.Logger
}

func NewConvertService(db *inm.Postgres, l *logger.Logger) *ConvertService {
	return &ConvertService{
		db:     db,
		logger: l,
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

func (c *ConvertService) SearchPictureName(p string) {
	path := "/home/user/programmin/obsidianProject/data/obsidianProject/" + p
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
	cmd := exec.Command("pandoc", inputFile, "-o", outputFile, "--from=markdown+hard_line_breaks")

	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("Error executing pandoc: %v\nOutput: %s", err, output)
	}

	fmt.Println("Conversion completed successfully.")
	return nil
}

func (c *ConvertService) ConvertHTMLToPDF(inputFile, outputFile string) error {
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
	log.Println("no new file for convertService")
	return result
}

func completeFileName(s string) string {
	return fmt.Sprintf("(./%s)", s)
}

func (c *ConvertService) ReplaceExtension(s string, oldext, newext string) string {
	// text.md --> text.html
	return strings.Replace(s, oldext, newext, -1)

}
