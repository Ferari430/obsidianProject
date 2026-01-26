package convertService

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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
	RootPath        string // Корневая папка для рекурсивного поиска изображений
	OutputPath      string // Путь к выходной папке (например: B:\output)
	Sep             string
	PandocPath      string
	WkhtmltopdfPath string
	imageCache      map[string]string // Кэш: имя файла -> полный путь
}

func NewConvertService(p, sep, pandocPath, wkhtmltopdf, rootPath string, db *inm.Postgres, l *logger.Logger) *ConvertService {
	// OutputPath должна быть отдельной папкой для PDF/HTML файлов
	outputPath := `B:\programmin-20260114T065921Z-1-001\programmin\obsidianProject\data\obsidianProject`

	// RootPath - корневая папка где все ищем (MD файлы + картинки)
	basePath := `B:\programmin-20260114T065921Z-1-001\programmin\obsidianProject\data\obsidianProject`

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		os.MkdirAll(outputPath, 0755)
	}

	return &ConvertService{
		db:              db,
		logger:          l,
		Path:            p,          // путь к MD файлам
		RootPath:        basePath,   // корень для рекурсивного поиска ВСЕГО
		OutputPath:      outputPath, // папка для выходных PDF/HTML
		Sep:             sep,
		PandocPath:      pandocPath,
		WkhtmltopdfPath: wkhtmltopdf,
		imageCache:      make(map[string]string),
	}
}

// GetFiles получает файлы из БД для конвертации
func (c *ConvertService) GetFiles() []*models.File {
	op := "convertService.GetFiles"
	arr := c.db.Get()
	if len(arr) == 0 {
		c.logger.Debug("no new file for convertService", slog.String("op", op))
		return nil
	}

	c.logger.Info("Got files for conversion",
		slog.String("op", op),
		slog.Int("count", len(arr)))
	return arr
}

// RemoveFromConverter удаляет файл из очереди конвертации
func (c *ConvertService) RemoveFromConverter(fileName string) error {
	return c.db.RemoveFromConverter(fileName)
}

// UpdateFileModifyTime обновляет время модификации файла в БД
func (c *ConvertService) UpdateFileModifyTime(fileName string) error {
	op := "convertService.UpdateFileModifyTime"

	fileInfo, err := os.Stat(fileName)
	if err != nil {
		c.logger.Error("Error getting file info",
			slog.String("op", op),
			slog.String("file", fileName),
			slog.String("error", err.Error()))
		return err
	}

	return c.db.UpdateFileModifyTime(fileName, fileInfo.ModTime())
}
func (c *ConvertService) SearchPictureName(mdFilePath string) {
	op := "convertService.SearchPictureName"

	// Читаем MD файл
	content, err := os.ReadFile(mdFilePath)
	if err != nil {
		c.logger.Error("Error reading file",
			slog.String("op", op),
			slog.String("file", mdFilePath),
			slog.String("error", err.Error()))
		return
	}

	mdContent := string(content)

	// Логируем начало обработки
	c.logger.Debug("Processing file for images",
		slog.String("op", op),
		slog.String("file", mdFilePath),
		slog.Int("content_length", len(mdContent)))
	alreadyProcessed := false
	reProcessed := regexp.MustCompile(`!\[\[?[^\]]*\]\]?\([^)]*[\\/][^)]*\)`)
	if reProcessed.MatchString(mdContent) {
		c.logger.Debug("File already has processed image links, skipping",
			slog.String("op", op),
			slog.String("file", mdFilePath))
		alreadyProcessed = true
	}

	if alreadyProcessed {
		return
	}

	// Выводим первые 200 символов для отладки
	previewLength := min(200, len(mdContent))
	c.logger.Debug("Content preview",
		slog.String("op", op),
		slog.String("preview", mdContent[:previewLength]))

	// 1. Находим все имена изображений в файле
	imageNames := c.extractImageNames(mdContent)
	if len(imageNames) == 0 {
		c.logger.Debug("No images found in file",
			slog.String("op", op),
			slog.String("file", mdFilePath))
		return
	}

	c.logger.Info("Found image references in file",
		slog.String("op", op),
		slog.String("file", mdFilePath),
		slog.Int("count", len(imageNames)),
		slog.Any("images", imageNames))

	// 2. Ищем файлы изображений рекурсивно
	imagePaths := make(map[string]string) // имя файла (без расширения) -> относительный путь
	mdDir := filepath.Dir(mdFilePath)

	for _, imageName := range imageNames {
		c.logger.Debug("Searching for image",
			slog.String("op", op),
			slog.String("image", imageName))

		// Ищем файл в файловой системе
		fullPath, err := c.findImageRecursive(imageName)
		if err != nil {
			c.logger.Warn("Image not found",
				slog.String("op", op),
				slog.String("image", imageName),
				slog.String("mdFile", mdFilePath),
				slog.String("error", err.Error()))
			continue
		}

		// Вычисляем относительный путь от MD файла к изображению
		relPath, err := filepath.Rel(mdDir, fullPath)
		if err != nil {
			c.logger.Error("Cannot compute relative path",
				slog.String("op", op),
				slog.String("mdFile", mdFilePath),
				slog.String("image", fullPath),
				slog.String("error", err.Error()))
			relPath = fullPath
		}

		// Форматируем путь для Markdown
		relPath = filepath.ToSlash(relPath)
		if !strings.HasPrefix(relPath, ".") && !strings.HasPrefix(relPath, "/") {
			relPath = "./" + relPath
		}

		imagePaths[imageName] = relPath
		c.logger.Info("Found image and computed path",
			slog.String("op", op),
			slog.String("image", imageName),
			slog.String("fullPath", fullPath),
			slog.String("relativePath", relPath))
	}

	// 3. Заменяем ссылки в MD файле
	originalContent := mdContent
	updatedContent := c.updateImageLinks(mdContent, imagePaths)

	// Проверяем изменения
	if updatedContent == originalContent {
		c.logger.Debug("No changes made to file",
			slog.String("op", op),
			slog.String("file", mdFilePath))
		return
	}

	// Логируем различия
	c.logger.Info("Changes detected in file",
		slog.String("op", op),
		slog.String("file", mdFilePath),
		slog.Int("imagesFound", len(imagePaths)))

	// Находим и логируем конкретные изменения
	originalLines := strings.Split(originalContent, "\n")
	updatedLines := strings.Split(updatedContent, "\n")

	for i := 0; i < len(originalLines) && i < len(updatedLines); i++ {
		if originalLines[i] != updatedLines[i] {
			c.logger.Info("Line changed",
				slog.String("op", op),
				slog.Int("line", i+1),
				slog.String("before", originalLines[i]),
				slog.String("after", updatedLines[i]))
		}
	}

	// 4. Сохраняем обновленный файл
	err = os.WriteFile(mdFilePath, []byte(updatedContent), 0644)
	if err != nil {
		c.logger.Error("Error writing file",
			slog.String("op", op),
			slog.String("file", mdFilePath),
			slog.String("error", err.Error()))
		return
	}

	c.logger.Info("File successfully updated",
		slog.String("op", op),
		slog.String("file", mdFilePath),
		slog.Int("imagesUpdated", len(imagePaths)))
}

// Вспомогательная функция для получения diff
func getDiff(old, new string, contextLines int) string {
	oldLines := strings.Split(old, "\n")
	newLines := strings.Split(new, "\n")

	var diff []string
	for i := 0; i < len(oldLines) && i < len(newLines); i++ {
		if oldLines[i] != newLines[i] {
			start := max(0, i-contextLines)

			diff = append(diff, fmt.Sprintf("Line %d:", i+1))
			diff = append(diff, fmt.Sprintf("  - %s", oldLines[i]))
			diff = append(diff, fmt.Sprintf("  + %s", newLines[i]))

			// Добавляем контекст до измененной строки
			if start < i {
				diff = append(diff, "Context before:")
				for j := start; j < i; j++ {
					diff = append(diff, fmt.Sprintf("    %s", oldLines[j]))
				}
			}

			// Добавляем контекст после измененной строки
			end := min(len(newLines), i+contextLines+1)
			if i+1 < end {
				diff = append(diff, "Context after:")
				for j := i + 1; j < end; j++ {
					diff = append(diff, fmt.Sprintf("    %s", newLines[j]))
				}
			}
			break
		}
	}

	if len(diff) == 0 {
		// Если строки полностью совпадают или одна короче другой
		if len(oldLines) != len(newLines) {
			return "File length changed"
		}
		return "No differences found"
	}

	return strings.Join(diff, "\n")
}
func (c *ConvertService) extractImageNames(mdContent string) []string {
	var imageNames []string
	seen := make(map[string]bool)

	// Регулярные выражения для разных форматов ссылок
	patterns := []struct {
		name   string
		regex  *regexp.Regexp
		isFull bool // Является ли ссылка полной (уже имеет путь)
	}{
		// Неполные ссылки: ![screen2] или ![image.png]
		{"incomplete", regexp.MustCompile(`!\[([^\]]+)\]`), false},

		// Полные ссылки: ![alt](path.png) или ![alt](path)
		{"full", regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`), true},

		// Obsidian ссылки: ![[image.png]] или ![[image]]
		{"obsidian", regexp.MustCompile(`!\[\[([^\]]+)\]\]`), false},
	}

	for _, pattern := range patterns {
		matches := pattern.regex.FindAllStringSubmatch(mdContent, -1)
		for _, match := range matches {
			if len(match) < 2 {
				continue
			}

			var imageRef string

			if pattern.isFull {
				// Для полных ссылок берем путь (вторая группа)
				if len(match) >= 3 {
					imageRef = match[2]

					// Если ссылка уже имеет путь с разделителями, пропускаем ее
					if strings.Contains(imageRef, "/") || strings.Contains(imageRef, "\\") {
						c.logger.Debug("Skipping already processed link",
							slog.String("type", pattern.name),
							slog.String("ref", imageRef))
						continue
					}
				}
			} else {
				// Для неполных и Obsidian ссылок берем содержимое квадратных скобок
				imageRef = match[1]

				// Проверяем, не является ли это уже преобразованной Obsidian ссылкой
				// вида ![[image2]](path)
				if pattern.name == "obsidian" && strings.Contains(match[0], "](") {
					c.logger.Debug("Skipping converted obsidian link",
						slog.String("match", match[0]))
					continue
				}
			}

			if imageRef == "" {
				continue
			}

			// Убираем путь, оставляем только имя файла
			if strings.Contains(imageRef, "/") || strings.Contains(imageRef, "\\") {
				imageRef = filepath.Base(imageRef)
			}

			// Убираем расширение для поиска
			imageNameWithoutExt := strings.TrimSuffix(imageRef, filepath.Ext(imageRef))

			if imageNameWithoutExt != "" && !seen[imageNameWithoutExt] {
				seen[imageNameWithoutExt] = true
				imageNames = append(imageNames, imageNameWithoutExt)
				c.logger.Debug("Found image reference",
					slog.String("type", pattern.name),
					slog.String("match", match[0]),
					slog.String("extracted", imageNameWithoutExt))
			}
		}
	}

	return imageNames
}

// findImageRecursive рекурсивно ищет изображение по всему RootPath
func (c *ConvertService) findImageRecursive(imageName string) (string, error) {
	// Проверяем кэш
	if path, found := c.imageCache[imageName]; found {
		return path, nil
	}

	var foundPath string

	// Список расширений для поиска
	extensions := []string{"", ".png", ".jpg", ".jpeg", ".gif", ".bmp", ".svg", ".webp"}

	// Рекурсивно ищем файл
	err := filepath.Walk(c.RootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || foundPath != "" {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		filename := info.Name()
		filenameWithoutExt := strings.TrimSuffix(filename, filepath.Ext(filename))

		// Сравниваем имена файлов без расширения
		if strings.EqualFold(filenameWithoutExt, imageName) {
			foundPath = path
			return filepath.SkipAll
		}

		// Также проверяем полное имя с каждым расширением
		for _, ext := range extensions {
			if ext == "" {
				continue
			}
			if strings.EqualFold(filename, imageName+ext) {
				foundPath = path
				return filepath.SkipAll
			}
		}

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("error searching for image: %v", err)
	}

	if foundPath == "" {
		return "", fmt.Errorf("image '%s' not found in %s", imageName, c.RootPath)
	}

	// Сохраняем в кэш
	c.imageCache[imageName] = foundPath
	c.logger.Info("Found image",
		slog.String("image", imageName),
		slog.String("path", foundPath))
	return foundPath, nil
}

// updateImageLinks обновляет ссылки на изображения в MD контенте
func (c *ConvertService) updateImageLinks(mdContent string, imagePaths map[string]string) string {
	result := mdContent

	// 1. Обработка полных ссылок (сначала их, чтобы не мешали)
	// Проверяем, не является ли ссылка уже полной с путем
	reFullWithPath := regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)
	result = reFullWithPath.ReplaceAllStringFunc(result, func(match string) string {
		innerMatch := reFullWithPath.FindStringSubmatch(match)
		if len(innerMatch) < 3 {
			return match
		}

		altText := innerMatch[1]
		currentPath := innerMatch[2]

		// Если ссылка уже содержит путь (с разделителями), оставляем как есть
		if strings.Contains(currentPath, "/") || strings.Contains(currentPath, "\\") {
			return match
		}

		// Если это просто имя файла без пути
		imageRef := currentPath
		imageName := imageRef
		if strings.Contains(imageRef, ".") {
			imageName = strings.TrimSuffix(imageRef, filepath.Ext(imageRef))
		}

		if relPath, found := imagePaths[imageName]; found {
			return fmt.Sprintf("![%s](%s)", altText, relPath)
		}

		return match
	})

	// 2. Обработка неполных ссылок: ![screen2]
	reIncomplete := regexp.MustCompile(`!\[([^\]]+)\]`)
	result = reIncomplete.ReplaceAllStringFunc(result, func(match string) string {
		innerMatch := reIncomplete.FindStringSubmatch(match)
		if len(innerMatch) < 2 {
			return match
		}

		imageRef := innerMatch[1]

		// Пропускаем, если это уже обработанная ссылка (имеет закрывающую скобку с путем)
		if strings.Contains(match, "](") {
			return match
		}

		imageName := imageRef
		if strings.Contains(imageRef, ".") {
			imageName = strings.TrimSuffix(imageRef, filepath.Ext(imageRef))
		}

		if relPath, found := imagePaths[imageName]; found {
			return fmt.Sprintf("![%s](%s)", imageRef, relPath)
		}

		return match
	})

	// 3. Обработка Obsidian-ссылок: ![[image2]]
	reObsidian := regexp.MustCompile(`!\[\[([^\]]+)\]\]`)
	result = reObsidian.ReplaceAllStringFunc(result, func(match string) string {
		innerMatch := reObsidian.FindStringSubmatch(match)
		if len(innerMatch) < 2 {
			return match
		}

		imageRef := innerMatch[1]

		// Проверяем, не является ли это уже преобразованной ссылкой вида ![[image2]](path)
		// Для этого ищем паттерн ![[...]](...)
		if strings.Contains(match, "](") {
			return match
		}

		imageName := imageRef
		if strings.Contains(imageRef, ".") {
			imageName = strings.TrimSuffix(imageRef, filepath.Ext(imageRef))
		}

		if relPath, found := imagePaths[imageName]; found {
			return fmt.Sprintf("![[%s]](%s)", imageRef, relPath)
		}

		return match
	})

	return result
}

// ConvertMDToHTML конвертирует MD файл в HTML
func (c *ConvertService) ConvertMDToHTML(inputFile, outputFile string) error {
	op := "convertService.ConvertMDToHTML"

	// Проверяем существование файла
	if _, err := os.Stat(inputFile); err != nil {
		return fmt.Errorf("input file not found: %s", inputFile)
	}

	// Создаем команду Pandoc
	cmd := exec.Command(c.PandocPath,
		inputFile,
		"-o", outputFile,
		"--from=markdown+hard_line_breaks",
		"--embed-resources",
		"--standalone",
		"--resource-path=.",
	)

	// Выполняем команду
	output, err := cmd.CombinedOutput()
	if err != nil {
		errorMsg := string(output)
		c.logger.Error("Pandoc failed",
			slog.String("op", op),
			slog.String("input", inputFile),
			slog.String("output", outputFile),
			slog.String("error", err.Error()),
			slog.String("pandoc_output", errorMsg))

		return fmt.Errorf("pandoc error: %v\nOutput: %s", err, output)
	}

	c.logger.Info("MD to HTML conversion completed",
		slog.String("op", op),
		slog.String("input", inputFile),
		slog.String("output", outputFile))
	return nil
}

// ConvertMdToPDF конвертирует MD файл в PDF
func (c *ConvertService) ConvertMdToPDF(fileName string) {
	op := "convertService.ConvertMdToPDF"

	// Получаем пути к файлам
	md, html, pdf := c.getAbsPath(fileName)

	c.logger.Info("Starting conversion",
		slog.String("op", op),
		slog.String("md", md),
		slog.String("html", html),
		slog.String("pdf", pdf))

	// 1. Исправляем пути к изображениям
	c.SearchPictureName(md)

	// 2. Конвертируем MD -> HTML
	err := c.ConvertMDToHTML(md, html)
	if err != nil {
		c.logger.Error("Can't convert MD to HTML",
			slog.String("op", op),
			slog.String("error", err.Error()))
		return
	}

	// 3. Конвертируем HTML -> PDF
	err = c.ConvertHTMLToPDF(html, pdf)
	if err != nil {
		c.logger.Error("Can't convert HTML to PDF",
			slog.String("op", op),
			slog.String("error", err.Error()))
		return
	}

	// 4. Обновляем время модификации в БД
	err = c.UpdateFileModifyTime(md)
	if err != nil {
		c.logger.Error("Can't update file modification time",
			slog.String("op", op),
			slog.String("error", err.Error()))
	}

	// 5. Удаляем файл из списка конвертации
	err = c.RemoveFromConverter(fileName)
	if err != nil {
		c.logger.Error("Can't remove file from converter",
			slog.String("op", op),
			slog.String("error", err.Error()))
	}

	c.logger.Info("Conversion completed successfully",
		slog.String("op", op),
		slog.String("file", fileName))
}

// getAbsPath возвращает абсолютные пути к MD, HTML и PDF файлам
func (c *ConvertService) getAbsPath(fileName string) (string, string, string) {
	op := "convertService.getAbsPath"

	// Определяем абсолютный путь к MD файлу
	var absFileName string
	if filepath.IsAbs(fileName) {
		absFileName = fileName
	} else {
		absFileName = filepath.Join(c.Path, fileName)
	}

	c.logger.Debug("getAbsPath calculation",
		slog.String("op", op),
		slog.String("input", fileName),
		slog.String("basePath", c.Path),
		slog.String("result", absFileName))

	// Проверяем существование файла
	if _, err := os.Stat(absFileName); err != nil {
		c.logger.Error("MD file does not exist",
			slog.String("op", op),
			slog.String("file", absFileName),
			slog.String("error", err.Error()))
	}

	// Получаем имя файла без расширения
	fileNameWithoutExt := strings.TrimSuffix(filepath.Base(fileName), ".md")

	// Создаем пути для HTML и PDF
	html := filepath.Join(c.OutputPath, fileNameWithoutExt+".html")
	pdf := filepath.Join(c.OutputPath, fileNameWithoutExt+".pdf")

	return absFileName, html, pdf
}

// ConvertHTMLToPDF конвертирует HTML файл в PDF
func (c *ConvertService) ConvertHTMLToPDF(inputFile, outputFile string) error {
	op := "convertService.ConvertHTMLToPDF"

	cmd := exec.Command(c.WkhtmltopdfPath,
		"--enable-local-file-access",
		"--encoding", "UTF-8",
		inputFile, outputFile)

	output, err := cmd.CombinedOutput()
	if err != nil {
		c.logger.Error("Wkhtmltopdf error",
			slog.String("op", op),
			slog.String("input", inputFile),
			slog.String("output", outputFile),
			slog.String("error", err.Error()),
			slog.String("output", string(output)))
		return fmt.Errorf("error executing wkhtmltopdf: %v\nOutput: %s", err, output)
	}

	c.logger.Info("HTML to PDF conversion completed",
		slog.String("op", op),
		slog.String("input", inputFile),
		slog.String("output", outputFile))
	return nil
}

// SetContentToModel загружает содержимое PDF файла в модель
func (c *ConvertService) SetContentToModel(file *models.File) error {
	op := "convertService.SetContentToModel"

	// Формируем путь к PDF файлу
	newFilename := file.FPath
	a := dirManager.ReplaceExtension(newFilename, ".md", ".pdf")
	path := filepath.Join(c.OutputPath, filepath.Base(a))

	// Открываем PDF файл
	f, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		c.logger.Error("Error opening PDF file",
			slog.String("op", op),
			slog.String("path", path),
			slog.String("error", err.Error()))
		return err
	}
	defer f.Close()

	// Читаем содержимое
	content, err := io.ReadAll(f)
	if err != nil {
		c.logger.Error("Error reading PDF content",
			slog.String("op", op),
			slog.String("path", path),
			slog.String("error", err.Error()))
		return err
	}

	// Сохраняем в модель
	file.SetPdfContent(content)
	c.logger.Info("PDF content set to model",
		slog.String("op", op),
		slog.String("file", file.FPath))

	return nil
}
