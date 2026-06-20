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
	LLM      LLM      `yaml:"llm"`
	ShareDir ShareDir `yaml:"ShareDir"`
	WorkPool WorkPool `yaml:"workpool"`
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
	MetricsPort   string `yaml:"metrics_port"`
}

type Registry struct {
	RegistryAddress []string `yaml:"registry_address"`
	Username        string   `yaml:"username"`
	Password        string   `yaml:"password"`
}

type LLM struct {
	BaseURL        string  `yaml:"base_url"`
	ModelName      string  `yaml:"model_name"`
	TimeoutSeconds int     `yaml:"timeout_seconds"`
	MaxTokens      int     `yaml:"max_tokens"`
	MaxRetries     int     `yaml:"max_retries"`
	Temperature    float32 `yaml:"temperature"`
	TopP           float32 `yaml:"top_p"`
}

type ShareDir struct {
	ShareDir string `yaml:"share_dir"`
}

type WorkPool struct {
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
	confFilePath, err := findConfFile()
	if err != nil {
		panic(err)
	}
	content, err := ioutil.ReadFile(confFilePath)
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
	normalizeLLMConfig(&conf.LLM)
	pretty.Printf("%+v\n", conf)
}

func normalizeLLMConfig(llm *LLM) {
	if llm.BaseURL == "" {
		llm.BaseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	}
	if llm.ModelName == "" {
		llm.ModelName = "qwen3.7-max"
	}
	if llm.TimeoutSeconds <= 0 {
		llm.TimeoutSeconds = 120
	}
	if llm.MaxTokens <= 0 {
		llm.MaxTokens = 8192
	}
	if llm.MaxRetries < 0 {
		llm.MaxRetries = 0
	}
	if llm.Temperature <= 0 {
		llm.Temperature = 0.7
	}
	if llm.TopP <= 0 {
		llm.TopP = 0.7
	}
}

func findConfFile() (string, error) {
	if p := os.Getenv("AI_CONF_PATH"); p != "" {
		return p, nil
	}

	rel := filepath.Join("app", "ai", "conf", GetEnv(), "conf.yaml")
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		candidate := filepath.Join(wd, rel)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}

		parent := filepath.Dir(wd)
		if parent == wd {
			break
		}
		wd = parent
	}

	return filepath.Join("conf", GetEnv(), "conf.yaml"), nil
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
