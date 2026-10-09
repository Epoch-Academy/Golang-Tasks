package main

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config 对应配置文件 config.yaml 的结构。
type Config struct {
	Server struct {
		Port string
	}
	Database struct {
		Path string
	}
	Debug bool
}

// loadConfig 用 viper 读取 config.yaml;找不到文件时退回到默认值。
func loadConfig() (*Config, error) {
	viper.SetConfigName("config") // 配置文件名(不带扩展名)
	viper.SetConfigType("yaml")   // 格式为 yaml
	viper.AddConfigPath(".")      // 在当前目录查找
	viper.AutomaticEnv()          // 允许用环境变量覆盖配置项

	// 默认值:即使没有 config.yaml,程序也能用这些值正常启动
	viper.SetDefault("server.port", ":8080")
	viper.SetDefault("database.path", "todos.db")
	viper.SetDefault("debug", true)

	if err := viper.ReadInConfig(); err != nil {
		// "文件不存在"可以容忍(用默认值);其它读取错误要上报
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("读取配置文件出错:%w", err)
		}
		fmt.Println("提示:未找到 config.yaml,使用默认配置。")
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置出错:%w", err)
	}
	return &cfg, nil
}
