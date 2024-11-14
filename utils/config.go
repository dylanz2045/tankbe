package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	RedisAddress string `mapstructure:"REDIS_ADDRESS"`
	Cert         string `mapstructure:"REDISCERT"`
	DbPort       string `mapstructure:"DBPORT"`
	DBHOST       string `mapstructure:"DBHOST"`
	DBName       string `mapstructure:"DBNAME"`
	DBUser       string `mapstructure:"DBUSER"`
	DBPwd        string `mapstructure:"DBPWD"`
	ServerHost   string `mapstructure:"SERVERHOST"`
	ServerPort   string `mapstructure:"SERVERPort"`
	JwtSecret    string `mapstructure:"JWTSECRET"`
	Aeskey       string `mapstructure:"USERINFOkey"`
	Aesiv        string `mapstructure:"USERINFOiv"`
}

func LoadConfig(path string) (config Config, err error) {
	// 获取当前工作目录
	wd, err := os.Getwd()
	if err != nil {
		return config, fmt.Errorf("failed to get working directory: %w", err)
	}

	// 构建配置文件的绝对路径
	configPath := filepath.Join(wd)

	viper.AddConfigPath(configPath)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()
	err = viper.ReadInConfig()
	if err != nil {
		return
	}
	err = viper.Unmarshal(&config)
	return
}

func SetConfig() Config {
	config, err := LoadConfig("..")
	if err != nil {
		fmt.Println("无法读取配置件")
		return Config{}
	}
	return config
}
