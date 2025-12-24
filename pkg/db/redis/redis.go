package redis

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var redisDbInstances = make(map[string]*redis.Client)

func Init(redisConfig map[string]RedisConfig, initDb []string) error {
	for _, dbName := range initDb {
		cfg, ok := redisConfig[dbName]
		if !ok {
			return fmt.Errorf("redis实例[%s]不存在于配置中", dbName)
		}

		// 创建客户端
		rdb := redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
			Username: cfg.User,
			Password: cfg.Password,
			DB:       cfg.DbName,
			PoolSize: cfg.PoolSize,
		})

		// 连通性检查
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		if _, err := rdb.Ping(ctx).Result(); err != nil {
			cancel()
			log.Fatalf("Redis [%s] 无法连接: %v", dbName, err)
		}
		cancel()

		redisDbInstances[dbName] = rdb
		log.Printf("Redis 实例 [%s] 初始化成功", dbName)
	}
	return nil
}

func GetDb(name string) (*redis.Client, error) {
	db, exists := redisDbInstances[name]
	if !exists {
		return nil, fmt.Errorf("redis数据库实例[%s]不存在", name)
	}
	if db == nil {
		return nil, fmt.Errorf("redis数据库实例[%s]未初始化", name)
	}
	return db, nil
}
