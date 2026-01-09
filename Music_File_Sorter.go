package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	fmt.Println("=== СОРТИРОВКА МУЗЫКАЛЬНЫХ ФАЙЛОВ ПО ГРУППАМ ===")
	fmt.Println()

	// Если путь передан как аргумент
	var folderPath string
	if len(os.Args) > 1 {
		folderPath = os.Args[1]
		fmt.Printf("Путь из аргумента: %s\n", folderPath)
	} else {
		// Запрашиваем путь
		fmt.Print("Введите путь к папке: ")
		fmt.Scanln(&folderPath)
	}

	// Убираем лишние пробелы
	folderPath = strings.TrimSpace(folderPath)

	if folderPath == "" {
		fmt.Println("Ошибка: путь не указан")
		fmt.Println("Использование:")
		fmt.Println("  1. Перетащите папку на программу")
		fmt.Println("  2. Или запустите: program.exe \"C:\\Музыка\\папка\"")
		return
	}

	// Если путь в кавычках, убираем их
	if strings.HasPrefix(folderPath, "\"") && strings.HasSuffix(folderPath, "\"") {
		folderPath = folderPath[1 : len(folderPath)-1]
	}

	// Получаем абсолютный путь
	absPath, err := filepath.Abs(folderPath)
	if err != nil {
		fmt.Printf("Ошибка обработки пути: %v\n", err)
		return
	}

	// Проверяем существование папки
	info, err := os.Stat(absPath)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		fmt.Printf("Путь: %s\n", absPath)
		return
	}

	if !info.IsDir() {
		fmt.Printf("Ошибка: %s не является папкой\n", absPath)
		return
	}

	fmt.Printf("\n✅ Найдена папка: %s\n", absPath)

	// Подтверждение
	fmt.Print("\n⚠️  Начать сортировку файлов по группам? (y/N): ")

	var confirm string
	fmt.Scanln(&confirm)
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm != "y" && confirm != "yes" && confirm != "д" && confirm != "да" {
		fmt.Println("❌ Отменено")
		return
	}

	fmt.Println("\n🔍 Начинаю сканирование и сортировку...")
	fmt.Println("══════════════════════════════════════════")

	sortedCount := sortFilesByArtist(absPath)

	fmt.Println("══════════════════════════════════════════")
	fmt.Printf("✅ Готово! Обработано файлов: %d\n", sortedCount)

	// Ждем нажатия Enter перед выходом
	fmt.Print("\nНажмите Enter для выхода...")
	fmt.Scanln()
}

func sortFilesByArtist(folderPath string) int {
	// Регулярное выражение для извлечения имени группы
	// Ожидаемый формат: "Группа - Название трека.расширение"
	// или "Группа - Альбом - Название трека.расширение"
	pattern := regexp.MustCompile(`^([^\-]+?)\s*-\s*[^\.]+\.\w+$`)
	processedCount := 0

	// Поддерживаемые музыкальные форматы
	musicExtensions := map[string]bool{
		".mp3":  true,
		".flac": true,
		".wav":  true,
		".aac":  true,
		".ogg":  true,
		".m4a":  true,
		".wma":  true,
	}

	// Собираем список файлов для обработки
	var filesToProcess []string
	
	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		
		if info.IsDir() {
			return nil
		}
		
		ext := strings.ToLower(filepath.Ext(info.Name()))
		if musicExtensions[ext] {
			filesToProcess = append(filesToProcess, path)
		}
		
		return nil
	})

	if err != nil {
		fmt.Printf("⚠️  Ошибка при сканировании: %v\n", err)
		return 0
	}

	fmt.Printf("Найдено музыкальных файлов: %d\n\n", len(filesToProcess))

	// Обрабатываем каждый файл
	for _, filePath := range filesToProcess {
		// Получаем имя файла
		fileName := filepath.Base(filePath)
		
		// Пытаемся извлечь имя группы
		matches := pattern.FindStringSubmatch(fileName)
		if matches == nil || len(matches) < 2 {
			fmt.Printf("⚠️  Неверный формат названия: %s\n", fileName)
			continue
		}

		// Извлекаем и очищаем имя группы
		artistName := strings.TrimSpace(matches[1])
		
		// Очищаем имя группы от лишних символов
		artistName = cleanArtistName(artistName)

		if artistName == "" {
			fmt.Printf("⚠️  Не удалось определить группу: %s\n", fileName)
			continue
		}

		// Создаем папку для группы
		artistFolder := filepath.Join(folderPath, artistName)
		if err := os.MkdirAll(artistFolder, 0755); err != nil {
			fmt.Printf("❌ Ошибка создания папки %s: %v\n", artistName, err)
			continue
		}

		// Формируем новый путь для файла
		newPath := filepath.Join(artistFolder, fileName)
		
		// Проверяем, не существует ли уже файл с таким именем
		if _, err := os.Stat(newPath); err == nil {
			// Файл уже существует, добавляем суффикс
			ext := filepath.Ext(fileName)
			nameWithoutExt := strings.TrimSuffix(fileName, ext)
			counter := 1
			
			for {
				newFileName := fmt.Sprintf("%s (%d)%s", nameWithoutExt, counter, ext)
				newPath = filepath.Join(artistFolder, newFileName)
				
				if _, err := os.Stat(newPath); os.IsNotExist(err) {
					break
				}
				counter++
			}
		}

		// Перемещаем файл
		if err := os.Rename(filePath, newPath); err != nil {
			fmt.Printf("❌ Ошибка перемещения %s: %v\n", fileName, err)
		} else {
			fmt.Printf("✅ Перемещен: %s → %s\n", fileName, artistName)
			processedCount++
		}
	}

	return processedCount
}

func cleanArtistName(name string) string {
	// Удаляем различные префиксы и суффиксы
	patternsToRemove := []*regexp.Regexp{
		regexp.MustCompile(`^\[\d+\]\s*`),           // [123] в начале
		regexp.MustCompile(`^\d+\.\s*`),            // 1. в начале
		regexp.MustCompile(`\s+\([^)]*\)$`),        // (что-то) в конце
		regexp.MustCompile(`\s+\[[^\]]*\]$`),       // [что-то] в конце
		regexp.MustCompile(`\s+ft\.\s+.*$`),        // ft. и все после
		regexp.MustCompile(`\s+feat\.\s+.*$`),      // feat. и все после
		regexp.MustCompile(`\s+vs\.\s+.*$`),        // vs. и все после
	}

	result := strings.TrimSpace(name)
	
	for _, pattern := range patternsToRemove {
		result = pattern.ReplaceAllString(result, "")
	}

	// Заменяем недопустимые символы в именах папок
	result = strings.ReplaceAll(result, ":", " -")
	result = strings.ReplaceAll(result, "/", " & ")
	result = strings.ReplaceAll(result, "\\", " & ")
	result = strings.ReplaceAll(result, "?", "")
	result = strings.ReplaceAll(result, "*", "")
	result = strings.ReplaceAll(result, "\"", "'")
	result = strings.ReplaceAll(result, "<", "(")
	result = strings.ReplaceAll(result, ">", ")")
	result = strings.ReplaceAll(result, "|", "-")

	return strings.TrimSpace(result)
}