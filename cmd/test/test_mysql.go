package main

import (
	"fmt"
	"log"

	"github.com/sunzhaoc/plant_be/pkg/db/mysql"
)

func main() {
	// 修复日志格式化：log.Fatal → log.Fatalf（否则%v占位符不生效）
	if err := mysql.Init(mysql.Load(), []string{"ali", "ali2"}); err != nil {
		log.Fatalf("初始化数据库失败：%v", err)
	}
	defer mysql.Close()

	// 关键改动：结构体标签从 db:"xxx" 改为 gorm:"column:xxx"（适配GORM映射规则）
	type Month struct {
		ID    string `gorm:"column:id"` // GORM通过column标签映射数据库列名
		Month string `gorm:"column:month"`
	}

	// 批量查询示例
	var months []Month
	sqlStr := "SELECT id, month FROM szc.dim_month"
	err := mysql.ExecuteSql("ali", sqlStr, nil, &months)
	if err != nil {
		fmt.Println("查询失败:", err)
		return
	}
	// 打印批量结果
	for _, m := range months {
		fmt.Printf("批量结果：ID=%s, Month=%s\n", m.ID, m.Month)
	}

	// 单行查询示例
	var singleMonth Month
	sqlStrSingle := "SELECT id, month FROM szc.dim_month WHERE id = ?"
	err = mysql.ExecuteSql("ali2", sqlStrSingle, []interface{}{"1"}, &singleMonth)
	if err != nil {
		fmt.Println("单行查询失败:", err)
		return
	}
	fmt.Printf("单行结果：ID=%s, Month=%s\n", singleMonth.ID, singleMonth.Month)
}
