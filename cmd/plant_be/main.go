package main

import (
	"log"

	"github.com/sunzhaoc/plant_be/pkg/db/mysql"
	"github.com/sunzhaoc/plant_be/pkg/db/redis"
	"github.com/sunzhaoc/plant_be/pkg/utils"
	"github.com/sunzhaoc/plant_be/routers"
)

func main() {
	if err := utils.InitSlog(); err != nil {
		panic("日志初始化失败: " + err.Error())
	}

	// 初始化 Mysql
	if err := mysql.Init(mysql.Load(), []string{"ali"}); err != nil {
		log.Fatal("初始化Mysql数据库失败：%v", err)
	}
	defer mysql.Close()

	// 初始化 Redis
	if err := redis.Init(redis.Load(), []string{"ali"}); err != nil {
		log.Fatalf("初始化Redis数据库失败：%v", err)
	}

	routers.InitRouter()
}
