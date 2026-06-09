// Package goini
package goini

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

type conf struct {
	confFile     string
	sectionName  string
	result       map[string]string
	sectionFound bool
}

var UsedConfigPath string // заполняется в Load/SimpleLoad

// getFileIniName возвращает путь файла конфигурации
func getFileIniName() string {
	exe, err := os.Executable()
	if err != nil {
		return "app.ini"
	}
	base := filepath.Base(exe)
	ext := filepath.Ext(base)
	app := strings.TrimSuffix(base, ext) // возвращает имя исполняемого файла без расширения
	ini := app + ".ini"

	// 3. Переменная окружения (высший приоритет)
	if envPath := os.Getenv("APP_INI_CONFIG"); envPath != "" {
		return envPath
	}

	// 2. Системный каталог
	sysPath := filepath.Join("/usr/local/etc", app, ini)
	if _, err := os.Stat(sysPath); err == nil {
		return sysPath
	}

	// 1. Текущий каталог
	if _, err := os.Stat(ini); err == nil {
		return ini
	}

	// 0. Если не удалось создать – возвращаем
	return "app.ini"
}

// Load возвращает структуру данных T переданных в параметрах
func Load[T any]() (T, error) {
	var cfg T

	t := reflect.TypeOf(cfg)
	if t.Kind() != reflect.Struct {
		return cfg, fmt.Errorf("Load ожидает структуру, получен %v", t.Kind())
	}
	typeName := t.Name()
	if typeName == "" {
		// значение по умолчанию
		// return cfg, fmt.Errorf("Load: анонимные структуры не поддерживаются (используйте именованную структуру)")
		typeName = "main"
	}

	// 2. Парсим файл
	UsedConfigPath = getFileIniName()
	c := conf{
		confFile:    UsedConfigPath,
		sectionName: "[" + typeName + "]", // получение имени структуры оно же имя секции
		result:      make(map[string]string, 2),
	}

	if err := c.parser(); err != nil {
		return cfg, err
	}

	// 3. Заполняем структуру
	v := reflect.ValueOf(&cfg).Elem()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i) // Значение поля
		tag := v.Type().Field(i).Tag.Get("ini")

		// Преобразуем значение из map в тип поля структуры
		if value, ok := c.result[tag]; ok {
			switch field.Kind() {
			case reflect.String:
				field.SetString(value)
			case reflect.Int:
				intVal, err := strconv.Atoi(value)
				if err != nil {
					return cfg, fmt.Errorf("%s %s: %v", "ошибка преобразования значения для поля", tag, err)
				}
				field.SetInt(int64(intVal))
			default:
				return cfg, fmt.Errorf("неподдерживаемый тип поля: %s", field.Kind())
			}
		}
	}

	return cfg, nil
}

// SimpleLoad упращенное чтение, возвращает map и только string
func SimpleLoad(section string) (map[string]string, error) {
	UsedConfigPath = getFileIniName()
	c := conf{
		confFile:    UsedConfigPath,
		sectionName: section,
		result:      make(map[string]string),
	}
	if err := c.parser(); err != nil {
		return nil, err
	}
	return c.result, nil
}

func (c *conf) parser() error {
	f, err := os.Open(c.confFile)
	if err != nil {
		return fmt.Errorf("ошибка открытия файла: %w", err)
	}
	defer f.Close()

	buf := bufio.NewScanner(f)
	for buf.Scan() {
		line := strings.TrimSpace(buf.Text())

		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		if !c.sectionFound && strings.Contains(line, c.sectionName) {
			c.sectionFound = true
			continue
		}

		if c.sectionFound && strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			break // новая секция
		}

		if c.sectionFound && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			if len(key) > 0 && len(value) > 0 {
				c.result[key] = value
			}
		}
	}

	if err := buf.Err(); err != nil {
		return fmt.Errorf("ошибка чтения файла: %w", err)
	}

	if !c.sectionFound {
		return fmt.Errorf("секция %s не найдена в файле %s", c.sectionName, c.confFile)
	}
	return nil
}
