package mdFinder

import (
	"bufio"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type imageFinder interface {
	GetInfo() map[string]string
}

type MDFinder struct {
	root        string
	dataType    string
	Data        map[string]string
	ImageFinder imageFinder
}

type Config struct {
	inputPath  string
	outputPath string
}

func NewMDFinder(root string, t string, imageFinder imageFinder) *MDFinder {
	return &MDFinder{
		root:        root,
		dataType:    t,
		Data:        make(map[string]string),
		ImageFinder: imageFinder,
	}
}

func (f *MDFinder) ScanFiles() error {

	err := filepath.WalkDir(f.root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext == f.dataType {
			filename := filepath.Base(path)

			if _, exists := f.Data[filename]; !exists {
				f.Data[filename] = path
			}
		}
		return nil
	})
	return err

}

func (f *MDFinder) ProcessFile(inputPath string) error {
	input, err := os.Open(inputPath)
	if err != nil {
		return err
	}

	defer input.Close()

	mainDir := `B:\programmin-20260114T065921Z-1-001\programmin\obsidianProject\imageFinder`
	tmpDir := filepath.Join(mainDir, "tmp")

	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return fmt.Errorf("не удалось создать папку tmp в %s: %v", tmpDir, err)
	}

	inputFilename := filepath.Base(inputPath)
	outputPath := filepath.Join(tmpDir, fmt.Sprintf("%s_TEMP.md",
		strings.TrimSuffix(inputFilename, ".md")))

	output, err := os.Create(outputPath)
	if err != nil {
		return err
	}

	defer output.Close()

	log.Printf("Создаю файл в: %s", outputPath)

	scanner := bufio.NewScanner(input)
	writer := bufio.NewWriter(output)
	defer writer.Flush()

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		processedLine := processLine(line, f.ImageFinder.GetInfo())
		_, err = writer.WriteString(processedLine + "\n")
		if err != nil {
			log.Println(err)
		}
	}

	return scanner.Err()
}

var (
	cfg = Config{
		inputPath:  "",
		outputPath: `B:\programmin-20260114T065921Z-1-001\programmin\obsidianProject\imageFinder\tmp`,
	}
)

//
//func processLine(line string, imagePath map[string]string) string {
//	if strings.Contains(line, "png") || strings.Contains(line, "![[") {
//		return handleLineWithImage(line, imagePath, cfg)
//	}
//	return line
//}
//
//var (
//	imageRegex = regexp.MustCompile(`!\[(\[)?([^\[\]]+\.(?:png|jpg|jpeg|gif|bmp|svg|webp|tiff|ico))(\])?\]`)
//)
//
//func handleLineWithImage(l string, imagePath map[string]string, cfg Config) string {
//	return imageRegex.ReplaceAllStringFunc(l, func(match string) string {
//		isDoubleBracket := strings.HasPrefix(match, "![[")
//		var filename string
//		if isDoubleBracket {
//			filename = match[3 : len(match)-2]
//		} else {
//			filename = match[2 : len(match)-1]
//		}
//		path, ok := imagePath[filename]
//		if !ok {
//			log.Println("cant find path for file:", filename)
//		}
//
//		newp, err := filepath.Rel(cfg.outputPath, path)
//		if err != nil {
//			log.Println("err:", err)
//			return ""
//		}
//
//		if isDoubleBracket {
//			return fmt.Sprintf("![[%s]](%s)", filename, newp)
//		}
//		return fmt.Sprintf("![%s](%s)", filename, newp)
//	})
//}

var (
	// Основное регулярное выражение - ловим ВСЁ внутри скобок
	imageRegex = regexp.MustCompile(`!\[(\[)?([^\]]+?)(\])?\]`)

	// Дополнительное: для извлечения чистого имени файла (без пути в скобках)
	fullLinkRegex = regexp.MustCompile(`!\[(\[)?([^\]]+?)(\])?\]\(([^)]+)\)`)
)

func processLine(line string, imagePath map[string]string) string {
	if strings.Contains(line, "![") {
		return handleLineWithImage(line, imagePath, cfg)
	}
	return line
}

func handleLineWithImage(l string, imagePath map[string]string, cfg Config) string {
	// Сначала обрабатываем ссылки с путями в скобках
	result := fullLinkRegex.ReplaceAllStringFunc(l, func(match string) string {
		parts := fullLinkRegex.FindStringSubmatch(match)
		if len(parts) < 5 {
			return match
		}

		filename := cleanFilename(parts[2]) // Очищаем имя файла
		currentPath := strings.TrimSpace(parts[4])
		isDoubleBracket := parts[1] == "[" && parts[3] == "]"

		// Если путь уже есть - возможно, его не нужно менять
		if isGoodPath(currentPath) {
			return match
		}

		// Ищем файл в мапе
		return processImageLink(filename, currentPath, isDoubleBracket, imagePath, cfg)
	})

	// Затем обрабатываем простые ![[filename]]
	result = imageRegex.ReplaceAllStringFunc(result, func(match string) string {
		// Пропускаем, если это уже обработано как полная ссылка
		if strings.Contains(match, "](") {
			return match
		}

		parts := imageRegex.FindStringSubmatch(match)
		if len(parts) < 4 {
			return match
		}

		filename := cleanFilename(parts[2])
		isDoubleBracket := parts[1] == "[" && parts[3] == "]"

		// Ищем файл в мапе
		return processImageLink(filename, "", isDoubleBracket, imagePath, cfg)
	})

	return result
}

// Очистка имени файла от лишних символов
func cleanFilename(raw string) string {
	// Убираем пробелы по краям
	filename := strings.TrimSpace(raw)

	// Убираем markdown-экранирование (если есть)
	filename = strings.ReplaceAll(filename, `\ `, " ") // \_ -> _
	filename = strings.ReplaceAll(filename, `\[`, "[") // \[ -> [
	filename = strings.ReplaceAll(filename, `\]`, "]") // \] -> ]
	filename = strings.ReplaceAll(filename, `\(`, "(") // \( -> (
	filename = strings.ReplaceAll(filename, `\)`, ")") // \) -> )

	return filename
}

// Поиск файла с учетом пробелов и разных вариантов
func findImageInMap(filename string, imagePath map[string]string) (string, bool) {
	// Вариант 1: Точное совпадение
	if path, ok := imagePath[filename]; ok {
		return path, true
	}

	// Вариант 2: Без учета регистра
	lowerFilename := strings.ToLower(filename)
	for key, path := range imagePath {
		if strings.ToLower(key) == lowerFilename {
			return path, true
		}
	}

	// Вариант 3: Частичное совпадение (для файлов с путями)
	baseName := filepath.Base(filename)
	for key, path := range imagePath {
		if strings.ToLower(filepath.Base(key)) == strings.ToLower(baseName) {
			return path, true
		}
	}

	// Вариант 4: Ищем с добавлением расширений
	if !strings.Contains(filename, ".") {
		extensions := []string{".png", ".jpg", ".jpeg", ".gif", ".bmp", ".svg"}
		for _, ext := range extensions {
			if path, ok := imagePath[filename+ext]; ok {
				return path, true
			}
			if path, ok := imagePath[filename+strings.ToUpper(ext)]; ok {
				return path, true
			}
		}
	}

	return "", false
}

func processImageLink(filename, currentPath string, isDoubleBracket bool,
	imagePath map[string]string, cfg Config) string {

	// Ищем файл
	path, found := findImageInMap(filename, imagePath)
	if !found && currentPath != "" {
		// Если не нашли в мапе, но есть текущий путь - используем его
		path = currentPath
		found = true
	}

	if !found {
		log.Printf("Не найден файл: %s", filename)
		// Возвращаем оригинал в правильном формате
		if isDoubleBracket {
			return fmt.Sprintf("![[%s]]", filename)
		}
		return fmt.Sprintf("![%s]", filename)
	}

	// Преобразуем путь
	relativePath, err := filepath.Rel(cfg.outputPath, path)
	if err != nil {
		log.Printf("Ошибка преобразования пути для %s: %v", filename, err)
		// Возвращаем оригинал
		if isDoubleBracket {
			return fmt.Sprintf("![[%s]]", filename)
		}
		return fmt.Sprintf("![%s]", filename)
	}

	// Форматируем результат
	relativePath = filepath.ToSlash(relativePath)

	// Экранируем пробелы в путях для markdown
	if strings.Contains(relativePath, " ") {
		relativePath = strings.ReplaceAll(relativePath, " ", "%20")
	}

	if isDoubleBracket {
		return fmt.Sprintf("![[%s]](%s)", filename, relativePath)
	}
	return fmt.Sprintf("![%s](%s)", filename, relativePath)
}

func isGoodPath(path string) bool {
	// Проверяем, является ли путь "хорошим" (относительным, рабочим)
	return strings.HasPrefix(path, "./") ||
		strings.HasPrefix(path, "../") ||
		strings.HasPrefix(path, "/") ||
		strings.Contains(path, "://") // http://, https://
}
