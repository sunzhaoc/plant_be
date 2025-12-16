package mysql

import (
	"log"

	"github.com/spf13/viper"
)

type MySQLConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"db_name"`
	Charset  string `mapstructure:"charset"`
	MaxOpen  int    `mapstructure:"max_open_conns"`     // 最大打开连接数
	MaxIdle  int    `mapstructure:"max_idle_conns"`     // 最大空闲连接数
	LifeTime int    `mapstructure:"conn_max_life_time"` // 连接最大存活时间（秒）
}

var MySQLCfg map[string]MySQLConfig

func Load() map[string]MySQLConfig {
	// 配置文件路径和名称
	viper.SetConfigName("config")   // 配置文件名（无后缀）
	viper.SetConfigType("yaml")     // 配置文件类型
	viper.AddConfigPath("./config") // 配置文件所在目录

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}

	// 解析配置到结构体
	if err := viper.UnmarshalKey("mysql", &MySQLCfg); err != nil {
		log.Fatalf("解析配置失败: %v", err)
	}
	return MySQLCfg
}
