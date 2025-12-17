package main

import (
	"log"

	"github.com/sunzhaoc/plant_be/pkg/db/mysql"
	"github.com/sunzhaoc/plant_be/routers"
)

func main() {
	if err := mysql.Init(mysql.Load(), []string{"ali"}); err != nil {
		log.Fatal("初始化数据库失败：%v", err)
	}
	defer mysql.Close()

	routers.InitRouter()
}
