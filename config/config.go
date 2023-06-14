package config

import (
	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigFile("./config/config.ini")
	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
}
