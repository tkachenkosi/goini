package goini

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

const NAME_FL = ".ini"

type conf struct {
	secName  string
	fileName string
	result   map[string]string
}

func Load[T any]() (T, error) {
	const msg_text = "ошибка преобразования значения для поля"
	var cfg T

	c := conf{
		fileName: NAME_FL,
		secName:  "[" + reflect.TypeOf(cfg).Name() + "]", // получение имени структуры оно же имя секции
		result:   make(map[string]string, 2),
	}

	if err := c.parser(); err != nil {
		return cfg, err
	}

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
				result, err := strconv.Atoi(value)
				if err != nil {
					return cfg, fmt.Errorf("%s %s: %v", msg_text, tag, err)
				}
				field.SetInt(int64(result))
			default:
				return cfg, fmt.Errorf("неподдерживаемый тип поля: %s", field.Kind())
			}
		}
	}

	return cfg, nil
}

// упращенное чтение, возвращает map и только string
func SimpleLoad(sec string) (map[string]string, error) {
	c := conf{
		fileName: NAME_FL,
		secName:  sec,
		result:   make(map[string]string),
	}
	if err := c.parser(); err != nil {
		return nil, err
	}
	return c.result, nil
}

func (c *conf) parser() error {
	var ok bool

	// 1 - поиск файла .ini в текущем каталоге
	if _, err := os.Stat(c.fileName); os.IsNotExist(err) {
		// 2 - попытаемся найти конфиг в общей паке /usr/local/etc/<имя программы>/.ini
		c.fileName = "/usr/local/etc/" + os.Args[0][strings.LastIndex(os.Args[0], "/")+1:] + "/" + c.fileName
		if _, err := os.Stat(c.fileName); os.IsNotExist(err) {
			return fmt.Errorf("файл конфигурации не найден: %s", c.fileName)
		}
	}

	f, err := os.Open(c.fileName)
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

		if !ok && strings.Contains(line, c.secName) {
			ok = true
			continue
		}

		if ok && strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			break // новая секция
		}

		if ok && strings.Contains(line, "=") {
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

	return nil
}
