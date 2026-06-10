## goini – Minimal INI Config Reader for Go

goini is a lightweight, zero-dependency INI configuration reader designed for fast, type-safe loading of specific sections into Go structs.
It prioritizes simplicity and performance over full INI standard compliance.

### Key Features
- Type-safe loading – Load a section directly into a typed struct.
- Minimal API – Only two exported functions: Load[T any]() and SimpleLoad().
- No reflection overhead after load – struct fields are mapped once.
- Supports string and int fields – other types are deliberately not supported.
- Automatic section name – section name equals struct type name ([TypeName]).
- Fallback file locations – works well for Linux daemons without home directory.
- Environment variable override – APP_INI_CONFIG takes highest priority.

### Why goini?
Unlike full-featured INI parsers (like go-ini, viper), goini is not a general-purpose configuration library.
#### It is perfect for:
- Small services, daemons, or CLI tools that need to read one specific section.
- Cases where you control the INI file syntax and don’t need advanced features (global keys, nested sections, etc.).
- When you want explicit, compile-time safety for config fields.
It trades flexibility for speed and dead-simple usage – no GetString(), GetInt() calls, just a struct.

### Installation
```bash
go get github.com/yourusername/goini
```
### Example INI File (yourapp.ini)
```ini
[pg]
host=10.10.10.10
port=5432
db=base
conns=5
user=user
passwd=user

[web]
host=localhost
port=3001
```
### Usage
Define structs with ini tags matching the keys:
```go
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
```
#### Load a section:
```go
import "github.com/tkachenkosi/goini"

func main() {
    cfg, err := goini.Load[web]()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("host=%s port=%s\n", cfg.Host, cfg.Port)
}
```
goini.Load[pg]() would read the [pg] section.
SimpleLoad("[web]") returns map[string]string for quick access.
### File Search Order
When you call Load or SimpleLoad, goini looks for the configuration file in this order:
1. Environment variable APP_INI_CONFIG (if set) – highest priority.
2. Current directory – ./yourapp.ini (where yourapp is the executable name without extension).
3. System directory – /usr/local/etc/yourapp/yourapp.ini (ideal for Linux daemons).
4. Fallback – ./yourapp.ini (will cause an error if missing).
#### Supported Field Types
- string
- int
If a field has a different type, Load returns an error.
Empty values in the INI file are ignored (the struct field retains its zero value).upported Field Types
### License
MIT
### Notes for Contributors
- Keep the API minimal (maximum 2–3 exported functions).
- Do not add support for floats, bools, slices, or nested structs.
- File creation/modification is out of scope (admins create the config manually).
