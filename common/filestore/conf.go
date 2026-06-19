package filestore

import (
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v2"
)

var (
	conf     *Config
	confOnce sync.Once
)

type Config struct {
	Env      string
	ShareDir ShareDir `yaml:"ShareDir"`
	Cache    Cache    `yaml:"cache"`
}

type ShareDir struct {
	ShareDir string `yaml:"share_dir"`
}

type Cache struct {
	CacheDir   string `yaml:"cache_dir"`
	TTLSeconds int64  `yaml:"ttl_seconds"`
	NeedFlush  bool   `yaml:"need_flush"`
}

func GetConf() *Config {
	confOnce.Do(initConf)
	return conf
}

func initConf() {
	confFilePath, err := findConfFile()
	if err != nil {
		panic(err)
	}
	content, err := os.ReadFile(confFilePath)
	if err != nil {
		panic(err)
	}

	conf = new(Config)
	if err := yaml.Unmarshal(content, conf); err != nil {
		panic(err)
	}
	conf.Env = GetEnv()
	if conf.ShareDir.ShareDir == "" {
		panic("filestore: ShareDir.share_dir is empty")
	}
	if conf.Cache.CacheDir == "" {
		conf.Cache.CacheDir = filepath.Join(filepath.Dir(filepath.Clean(conf.ShareDir.ShareDir)), "cache")
	}
	if conf.Cache.TTLSeconds <= 0 {
		conf.Cache.TTLSeconds = 7200
	}
}

func GetEnv() string {
	e := os.Getenv("GO_ENV")
	if e == "" {
		return "test"
	}
	return e
}

func findConfFile() (string, error) {
	if p := os.Getenv("FILESTORE_CONF_PATH"); p != "" {
		return p, nil
	}

	rel := filepath.Join("common", "filestore", "conf", GetEnv(), "conf.yaml")
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

	return filepath.Join(rel), nil
}
