package redis

import (
	"log"

	"github.com/spf13/viper"
)

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DbName   int    `mapstructure:"db_name"`
	PoolSize int    `mapstructure:"pool_size"`
}

var RedisCfg map[string]RedisConfig

func Load() map[string]RedisConfig {
	// 配置文件路径和名称
	viper.SetConfigName("config")   // 配置文件名（无后缀）
	viper.SetConfigType("yaml")     // 配置文件类型
	viper.AddConfigPath("./config") // 配置文件所在目录

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}

	// 解析配置到结构体
	if err := viper.UnmarshalKey("redis", &RedisCfg); err != nil {
		log.Fatalf("解析配置失败: %v", err)
	}
	return RedisCfg
}
