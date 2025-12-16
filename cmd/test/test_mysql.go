package main

import (
	"fmt"
	"log"

	"github.com/sunzhaoc/plant_be/pkg/db/mysql"
)

func main() {
	if err := mysql.Init(mysql.Load(), []string{"ali", "ali2"}); err != nil {
		log.Fatal("初始化数据库失败：%v", err)
	}
	defer mysql.Close()

	type Month struct {
		ID    string `db:"id"` // 字段名与列名一致（或用db标签映射）
		Month string `db:"month"`
	}
	var months []Month
	sqlStr := "SELECT id, month FROM szc.dim_month"
	err := mysql.ExecuteSql("ali", sqlStr, nil, &months)
	if err != nil {
		fmt.Println("查询失败:", err)
		return
	}
	// 打印结果
	for _, m := range months {
		fmt.Println(m.ID, m.Month)
	}

	var singleMonth Month
	sqlStrSingle := "SELECT id, month FROM szc.dim_month WHERE id = ?"
	err = mysql.ExecuteSql("ali2", sqlStrSingle, []interface{}{"1"}, &singleMonth)
	if err != nil {
		fmt.Println("单行查询失败:", err)
		return
	}
	fmt.Println("单行结果:", singleMonth.ID, singleMonth.Month)

}
