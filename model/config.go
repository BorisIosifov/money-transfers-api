package model

import (
	"io/ioutil"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	Port           int            `yaml:"port"`
	Postgres       PostgresConfig `yaml:"postgres"`
	TelegramChatID string         `yaml:"TelegramChatID"`
	SMTPhost       string         `yaml:"SMTPhost"`
	SMTPport       int            `yaml:"SMTPport"`
	SMTPlogin      string         `yaml:"SMTPlogin"`
	SMTPpassword   string         `yaml:"SMTPpassword"`
}

type PostgresConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
}

func LoadConfig() (config Config, err error) {
	env := os.Getenv("MT_ENV")
	if env == "" {
		env = "desktop"
	}
	confDir := "config/" + env + "/"

	if env == "test" {
		confDir = "../" + confDir
	}

	fileContent, err := ioutil.ReadFile(filepath.Join(confDir, "config.yaml"))
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(fileContent, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}
