The package is designed to work with configuration files in the INI format, which contains sections and keys. 
It reads information from a specific section, dividing it into keys and values, and stores these values in the corresponding fields of the structure, which correspond to the names of the keys in the configuration file.
~~~
.ini file:
[pg]
host=10.10.10.10
port=5432
db=base
conns=5
user=user
passwd=user
[dic]
dict=/usr/local/share/dict/dict24.dic
inp=/usr/local/share/dict/dict24.txt
mode=1
[web]
host=localhost
port=3001
***

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

// Loads the "web" section
c, err := ini.Load[web]()
if err != nil {
    slog.Error(err.Error())
}
fmt.Printf("host=%s port=%s\n", c.Host, c.Port)

// Loads the "dic" section
d, err := ini.Load[dic]()
if err != nil {
    slog.Error(err.Error())
}
fmt.Printf("dic=%s inp=%s mode=%d\n", d.Dic, d.Inp, d.Mode)

// Loads the "pg" section
p, err := ini.Load[pg]()
if err != nil {
    slog.Error(err.Error())
}
fmt.Printf("user=%s password=%s host=%s port=%s dbname=%s pool_max_conns=%s\n", p.User, p.Passwd, p.Host, p.Port, p.Db, p.Conns)

// Working with the map type
m, err := ini.SimpleLoad("[web]")
if err != nil {
    slog.Error(err.Error())
}
fmt.Println("Si
mple:", m["host"], m["port"])
~~~
