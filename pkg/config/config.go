package config

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var rootDir = GetProjectRoot()

func GetProjectRoot() string {
	_, b, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(b)

	for {
		if matches, err := filepath.Glob(filepath.Join(basePath, "go.mod")); err == nil && len(matches) > 0 {
			log.Println("base path: ", basePath)
			return basePath
		}
		parent := filepath.Dir(basePath)
		if parent == basePath {
			log.Fatal("Project root not found")
		}
		basePath = parent
	}
}

type Config struct {
	Watcher struct {
		Path string `yaml:"path"`
	} `yaml:"watcher"`

	Discovery struct {
		BroadcastPort     int           `yaml:"broadcastPort"`
		TcpPort           int           `yaml:"tcpPort"`
		BroadcastInterval time.Duration `yaml:"broadcastInterval"`
		BufferSize        int           `yaml:"bufferSize"`
	} `yaml:"discovery"`

	Transfer struct {
		BufferSize  int `yaml:"bufferSize"`
		Consistency struct {
			OnConflict string `yaml:"onConflict"`
		} `yaml:"consistency"`
	} `yaml:"transfer"`
}

func NewConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(fmt.Sprintf("%s/config", rootDir))

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	c := Config{}

	err := viper.Unmarshal(&c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
