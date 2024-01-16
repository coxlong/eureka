package config

import (
	"github.com/spf13/viper"
)

const (
	Production  = "production"
	Development = "development"
)

type Config struct {
	Env           Env
	Server        Server
	Logger        Logger
	Database      Database
	Authorization Authorization
	OpenAI        OpenAI
}
type Env struct {
	Mode         string
	FrontendAddr string
}

type Server struct {
	Addr    string
	CrtFile string
	KeyFile string
}

type Logger struct {
	Filename   string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
	Level      string
}

type Database struct {
	Type       string
	DSN        string
	SQLitePath string
}

type Authorization struct {
	SessionKey         string
	GithubClient       string
	GithubClientSecret string
}

type OpenAI struct {
	BaseURL string
}

func LoadConf(fileName string) (*Config, error) {
	viper.SetConfigFile(fileName)
	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	var config Config
	// 解析配置文件
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}
	return &config, nil
}
