package conf

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/kr/pretty"
	"gopkg.in/validator.v2"
	"gopkg.in/yaml.v2"
)

var (
	conf *Config
	once sync.Once
)

type Config struct {
	Env      string
	Kitex    Kitex    `yaml:"kitex"`
	MySQL    MySQL    `yaml:"mysql"`
	Redis    Redis    `yaml:"redis"`
	Registry Registry `yaml:"registry"`
	ES       ES       `yaml:"es"`
	Milvus   Milvus   `yaml:"milvus"`
	Index    Index    `yaml:"index"`
}

type MySQL struct {
	DSN string `yaml:"dsn"`
}

type Redis struct {
	Address  string `yaml:"address"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type Kitex struct {
	Service       string `yaml:"service"`
	Address       string `yaml:"address"`
	LogLevel      string `yaml:"log_level"`
	LogFileName   string `yaml:"log_file_name"`
	LogMaxSize    int    `yaml:"log_max_size"`
	LogMaxBackups int    `yaml:"log_max_backups"`
	LogMaxAge     int    `yaml:"log_max_age"`
}

type Registry struct {
	RegistryAddress []string `yaml:"registry_address"`
	Username        string   `yaml:"username"`
	Password        string   `yaml:"password"`
}

type ES struct {
	Addresses []string `yaml:"addresses"`
	Index     string   `yaml:"index"`
	Username  string   `yaml:"username"`
	Password  string   `yaml:"password"`
}

type Milvus struct {
	Address    string `yaml:"address"`
	Collection string `yaml:"collection"`
	VectorDim  int    `yaml:"vector_dim"`
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
}

type Index struct {
	TaskChunkSize      int `yaml:"task_chunk_size"`
	EmbeddingBatchSize int `yaml:"embedding_batch_size"`
	WriteBatchSize     int `yaml:"write_batch_size"`
	MinWorkers         int `yaml:"min_workers"`
	MaxWorkers         int `yaml:"max_workers"`
	QueueSize          int `yaml:"queue_size"`
	ScaleUpThreshold   int `yaml:"scale_up_threshold"`
	ScaleDownThreshold int `yaml:"scale_down_threshold"`
	IdleTimeoutSeconds int `yaml:"idle_timeout_seconds"`
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
		klog.Error("parse yaml error - %v", err)
		panic(err)
	}
	if err := validator.Validate(conf); err != nil {
		klog.Error("validate config error - %v", err)
		panic(err)
	}
	conf.Env = GetEnv()
	pretty.Printf("%+v\n", conf)
}

func GetEnv() string {
	e := os.Getenv("GO_ENV")
	if len(e) == 0 {
		return "test"
	}
	return e
}

func LogLevel() klog.Level {
	level := GetConf().Kitex.LogLevel
	switch level {
	case "trace":
		return klog.LevelTrace
	case "debug":
		return klog.LevelDebug
	case "info":
		return klog.LevelInfo
	case "notice":
		return klog.LevelNotice
	case "warn":
		return klog.LevelWarn
	case "error":
		return klog.LevelError
	case "fatal":
		return klog.LevelFatal
	default:
		return klog.LevelInfo
	}
}
