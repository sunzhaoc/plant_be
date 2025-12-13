package main

import (
	"log"
	"os"

	"github.com/sunzhaoc/plant_be/config"
	"github.com/sunzhaoc/plant_be/internal/oss"
)

func main() {
	// 加载配置
	cfg := config.LoadAliConfig()

	// 下载图片
	objectKey := "plant/squamellaria/squamellaria_grayi/img.png"
	imageBytes, err := oss.DownloadImage(cfg, objectKey)
	if err != nil {
		log.Fatalf("获取图片失败: %v", err)
	}

	// 保存到本地
	if err = os.WriteFile("downloaded.jpg", imageBytes, 0644); err != nil {
		log.Fatalf("写入文件失败: %v", err)
	}

	log.Printf("成功！图片大小：%d 字节，已保存为downloaded.jpg", len(imageBytes))
}
