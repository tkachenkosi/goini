package goini

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// Структуры для тестов (совпадают с примерами из README)
type pg struct {
	Host   string `ini:"host"`
	Port   string `ini:"port"`
	Db     string `ini:"db"`
	Conns  string `ini:"conns"`
	User   string `ini:"user"`
	Passwd string `ini:"passwd"`
}

type web struct {
	Host string `ini:"host"`
	Port string `ini:"port"`
}

type dic struct {
	Inp  string `ini:"inp"`
	Dic  string `ini:"dict"`
	Mode int    `ini:"mode"`
}

// Вспомогательная функция: создаёт временный файл с содержимым
func createTempIni(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	iniPath := filepath.Join(tmpDir, "app.ini")
	err := os.WriteFile(iniPath, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("не удалось создать временный .ini: %v", err)
	}
	return iniPath
}

// Сохраняем и восстанавливаем переменную окружения и текущий каталог, чтобы тесты не мешали друг другу
func setEnvAndChdir(t *testing.T, envValue string, chdirDir string) (restore func()) {
	t.Helper()
	oldEnv := os.Getenv("APP_INI_CONFIG")
	oldDir, _ := os.Getwd()
	if envValue != "" {
		os.Setenv("APP_INI_CONFIG", envValue)
	}
	if chdirDir != "" {
		os.Chdir(chdirDir)
	}
	return func() {
		if envValue != "" {
			os.Setenv("APP_INI_CONFIG", oldEnv)
		}
		if chdirDir != "" {
			os.Chdir(oldDir)
		}
	}
}

// ----------------------------------------------------------------------
// Тесты для Load[T]
// ----------------------------------------------------------------------

func TestLoad_ValidSection_Web(t *testing.T) {
	content := `[web]
host=localhost
port=3001`
	iniPath := createTempIni(t, content)

	restore := setEnvAndChdir(t, iniPath, "")
	defer restore()

	// Перед вызовом обнуляем UsedConfigPath (если она глобальная)
	UsedConfigPath = ""

	cfg, err := Load[web]()
	if err != nil {
		t.Fatalf("ожидался успех, получили ошибку: %v", err)
	}
	expected := web{Host: "localhost", Port: "3001"}
	if !reflect.DeepEqual(cfg, expected) {
		t.Errorf("результат не совпадает: %+v, ожидалось %+v", cfg, expected)
	}
	// Проверяем, что UsedConfigPath установлен
	if UsedConfigPath != iniPath {
		t.Errorf("UsedConfigPath = %q, ожидался %q", UsedConfigPath, iniPath)
	}
}

func TestLoad_ValidSection_Dic(t *testing.T) {
	content := `[dic]
dict=/usr/local/share/dict/dict24.dic
inp=/usr/local/share/dict/dict24.txt
mode=1`
	iniPath := createTempIni(t, content)
	restore := setEnvAndChdir(t, iniPath, "")
	defer restore()

	cfg, err := Load[dic]()
	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}
	expected := dic{
		Dic:  "/usr/local/share/dict/dict24.dic",
		Inp:  "/usr/local/share/dict/dict24.txt",
		Mode: 1,
	}
	if !reflect.DeepEqual(cfg, expected) {
		t.Errorf("dic: %+v, ожидалось %+v", cfg, expected)
	}
}

func TestLoad_ValidSection_Pg(t *testing.T) {
	content := `[pg]
host=10.158.10.10
port=5432
db=base
conns=5
user=user
passwd=user`
	iniPath := createTempIni(t, content)
	restore := setEnvAndChdir(t, iniPath, "")
	defer restore()

	cfg, err := Load[pg]()
	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}
	expected := pg{
		Host:   "10.158.10.10",
		Port:   "5432",
		Db:     "base",
		Conns:  "5",
		User:   "user",
		Passwd: "user",
	}
	if !reflect.DeepEqual(cfg, expected) {
		t.Errorf("pg: %+v, ожидалось %+v", cfg, expected)
	}
}

func TestLoad_SectionNotFound(t *testing.T) {
	content := `[other]
key=value`
	iniPath := createTempIni(t, content)
	restore := setEnvAndChdir(t, iniPath, "")
	defer restore()

	_, err := Load[web]()
	if err == nil {
		t.Fatal("ожидалась ошибка о ненайденной секции, но её нет")
	}
	if !contains(err.Error(), "секция [web] не найдена") {
		t.Errorf("неправильное сообщение об ошибке: %v", err)
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	// Убедимся, что переменная окружения не указывает на существующий файл
	restore := setEnvAndChdir(t, "/nonexistent/path/to.ini", "")
	defer restore()
	_, err := Load[web]()
	if err == nil {
		t.Fatal("ожидалась ошибка открытия файла")
	}
}

func TestLoad_InvalidIntField(t *testing.T) {
	content := `[dic]
dict=/some/path
inp=/other
mode=notnumber`
	iniPath := createTempIni(t, content)
	restore := setEnvAndChdir(t, iniPath, "")
	defer restore()

	_, err := Load[dic]()
	if err == nil {
		t.Fatal("ожидалась ошибка преобразования int, но её нет")
	}
}

func TestLoad_UnsupportedFieldType(t *testing.T) {
	// Определим структуру с неподдерживаемым типом (bool)
	type unsupported struct {
		Flag bool `ini:"flag"`
	}
	content := `[unsupported]
flag=true`
	iniPath := createTempIni(t, content)
	restore := setEnvAndChdir(t, iniPath, "")
	defer restore()

	_, err := Load[unsupported]()
	if err == nil {
		t.Fatal("ожидалась ошибка неподдерживаемого типа")
	}
}

func TestLoad_NotStruct(t *testing.T) {
	// Попытка загрузить в int
	type myInt int
	_, err := Load[myInt]()
	if err == nil {
		t.Fatal("ожидалась ошибка 'ожидает структуру'")
	}
}

// ----------------------------------------------------------------------
// Тесты для SimpleLoad
// ----------------------------------------------------------------------

func TestSimpleLoad_ValidSection(t *testing.T) {
	content := `[web]
host=example.com
port=8080
`
	iniPath := createTempIni(t, content)
	restore := setEnvAndChdir(t, iniPath, "")
	defer restore()

	m, err := SimpleLoad("[web]")
	if err != nil {
		t.Fatalf("SimpleLoad вернул ошибку: %v", err)
	}
	expected := map[string]string{"host": "example.com", "port": "8080"}
	if !reflect.DeepEqual(m, expected) {
		t.Errorf("SimpleLoad: %+v, ожидалось %+v", m, expected)
	}
}

func TestSimpleLoad_SectionNotFound(t *testing.T) {
	content := `[other]
key=val`
	iniPath := createTempIni(t, content)
	restore := setEnvAndChdir(t, iniPath, "")
	defer restore()

	_, err := SimpleLoad("[web]")
	if err == nil {
		t.Fatal("ожидалась ошибка ненайденной секции")
	}
}

// ----------------------------------------------------------------------
// Тесты приоритета поиска: env, текущий каталог, системный
// ----------------------------------------------------------------------

// Проверяем, что переменная окружения APP_INI_CONFIG имеет наивысший приоритет.
func TestConfigPriority_EnvFirst(t *testing.T) {
	// Создаём два временных файла
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	content1 := `[web]
host=from_env`
	content2 := `[web]
host=from_local`

	envFile := filepath.Join(dir1, "env.ini")
	localFile := filepath.Join(dir2, "app.ini") // имя по умолчанию (app.ini)

	if err := os.WriteFile(envFile, []byte(content1), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(localFile, []byte(content2), 0o644); err != nil {
		t.Fatal(err)
	}

	// Переходим в каталог, где лежит localFile (чтобы текущий каталог подхватился)
	restore := setEnvAndChdir(t, envFile, dir2)
	defer restore()

	cfg, err := Load[web]()
	if err != nil {
		t.Fatalf("Load вернул ошибку: %v", err)
	}
	if cfg.Host != "from_env" {
		t.Errorf("должен использоваться файл из окружения, получено host=%s", cfg.Host)
	}
}

// Проверяем, что при отсутствии env используется текущий каталог.
func TestConfigPriority_CurrentDir(t *testing.T) {
	dir := t.TempDir()
	content := `[web]
host=from_current`
	iniPath := filepath.Join(dir, "app.ini")
	if err := os.WriteFile(iniPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	// Переменная окружения не задана, рабочий каталог — dir
	restore := setEnvAndChdir(t, "", dir)
	defer restore()

	cfg, err := Load[web]()
	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}
	if cfg.Host != "from_current" {
		t.Errorf("host=%s, ожидалось from_current", cfg.Host)
	}
}

// Проверяем, что если нет ни env, ни файла в текущем каталоге, то ищется /usr/local/etc/...
// В тесте мы не можем создать /usr/local/etc, поэтому проверяем только что при отсутствии возвращается ошибка (нет файла).
// Чтобы не зависеть от прав, мы можем подменить getFileIniName через monkey patch, но лучше просто проверить, что ошибка возникает.
func TestConfigPriority_SysPathFallback(t *testing.T) {
	// Создаём временную директорию и делаем вид, что это системная.
	// Однако в реальности sysPath жёстко зашит в getFileIniName. Для теста можно подменить через env,
	// но проще проверить, что без файла вообще возвращается ошибка.
	restore := setEnvAndChdir(t, "", "")
	defer restore()

	// Удаляем любые app.ini в текущем каталоге (если случайно есть)
	os.Remove("app.ini") // не критично

	_, err := Load[web]()
	if err == nil {
		t.Fatal("ожидалась ошибка отсутствия файла")
	}
	// Дополнительно: если бы файл лежал в /usr/local/etc/... он был бы найден. Но тест не может это проверить без прав.
}

// ----------------------------------------------------------------------
// Тест на пустые значения (если вы решите их разрешить)
// Сейчас пакет игнорирует пустые значения, но тест поможет отследить изменение поведения.
func TestLoad_EmptyValue(t *testing.T) {
	content := `[web]
host=
port=8080`
	iniPath := createTempIni(t, content)
	restore := setEnvAndChdir(t, iniPath, "")
	defer restore()

	cfg, err := Load[web]()
	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}
	// В текущей реализации поле host останется пустым, т.к. пустое значение игнорируется
	if cfg.Host != "" {
		t.Errorf("ожидался пустой host, получен %q", cfg.Host)
	}
	if cfg.Port != "8080" {
		t.Errorf("port=%s, ожидался 8080", cfg.Port)
	}
}

// ----------------------------------------------------------------------
// Вспомогательная функция
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > 0 && len(substr) > 0 && findSubstr(s, substr)))
}

func findSubstr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
