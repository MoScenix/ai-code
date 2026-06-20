package conf

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/kr/pretty"
	"gopkg.in/validator.v2"
	"gopkg.in/yaml.v2"
)

var (
	conf *Config
	once sync.Once
)

type Config struct {
	Env    string
	Hertz  Hertz  `yaml:"hertz"`
	MySQL  MySQL  `yaml:"mysql"`
	Redis  Redis  `yaml:"redis"`
	Consul Consul `yaml:"consul"`
	Static Static `yaml:"static"`
	Deploy Deploy `yaml:"deploy"`
	File   File   `yaml:"file"`
}

type Consul struct {
	Address string `yaml:"address"`
}

type MySQL struct {
	DSN string `yaml:"dsn"`
}

type Redis struct {
	Address  string `yaml:"address"`
	Password string `yaml:"password"`
	Username string `yaml:"username"`
	DB       int    `yaml:"db"`
}

type Hertz struct {
	Service            string `yaml:"service"`
	Address            string `yaml:"address"`
	MaxRequestBodySize int    `yaml:"max_request_body_size"`
	EnablePprof        bool   `yaml:"enable_pprof"`
	EnableGzip         bool   `yaml:"enable_gzip"`
	LogLevel           string `yaml:"log_level"`
	LogFileName        string `yaml:"log_file_name"`
	LogMaxSize         int    `yaml:"log_max_size"`
	LogMaxBackups      int    `yaml:"log_max_backups"`
	LogMaxAge          int    `yaml:"log_max_age"`
}

type Static struct {
	Route     string `yaml:"route"`
	Root      string `yaml:"root"`
	URLPrefix string `yaml:"url_prefix"`
	Project   string `yaml:"project"`
	Avatar    string `yaml:"avatar"`
	Cover     string `yaml:"cover"`
}

type Deploy struct {
	Root      string `yaml:"root"`
	URLPrefix string `yaml:"url_prefix"`
}

type File struct {
	BigThresholdBytes int64 `yaml:"big_threshold_bytes"`
	ChunkMinSize      int64 `yaml:"chunk_min_size"`
	ChunkMaxSize      int64 `yaml:"chunk_max_size"`
}

// GetConf gets configuration instance
func GetConf() *Config {
	once.Do(initConf)
	return conf
}

func initConf() {
	prefix := "conf"
	confFileRelPath := filepath.Join(prefix, filepath.Join(GetEnv(), "conf.yaml"))
	content, err := ioutil.ReadFile(confFileRelPath)
	if err != nil {
		panic(err)
	}

	conf = new(Config)
	err = yaml.Unmarshal(content, conf)
	if err != nil {
		hlog.Error("parse yaml error - %v", err)
		panic(err)
	}
	if err := validator.Validate(conf); err != nil {
		hlog.Error("validate config error - %v", err)
		panic(err)
	}

	conf.Env = GetEnv()
	normalizeConfig(conf)

	pretty.Printf("%+v\n", conf)
}

func normalizeConfig(conf *Config) {
	if conf.Hertz.MaxRequestBodySize <= 0 {
		conf.Hertz.MaxRequestBodySize = 200 << 20
	}
}

func GetEnv() string {
	e := os.Getenv("GO_ENV")
	if len(e) == 0 {
		return "test"
	}
	return e
}

func LogLevel() hlog.Level {
	level := GetConf().Hertz.LogLevel
	switch level {
	case "trace":
		return hlog.LevelTrace
	case "debug":
		return hlog.LevelDebug
	case "info":
		return hlog.LevelInfo
	case "notice":
		return hlog.LevelNotice
	case "warn":
		return hlog.LevelWarn
	case "error":
		return hlog.LevelError
	case "fatal":
		return hlog.LevelFatal
	default:
		return hlog.LevelInfo
	}
}
