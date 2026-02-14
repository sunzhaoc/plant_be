package main

import (
	"fmt"

	"github.com/sunzhaoc/plant_be/pkg/utils"
)

func main() {
	ipInfo, err := utils.GetIpInfo("1.207.133.71")
	if err != nil {
		fmt.Println("获取IP信息失败：", err)
		return
	}

	// 遍历打印字典中的信息
	fmt.Println("国家：", ipInfo["Country"])
	fmt.Println("省份：", ipInfo["Prov"])
	fmt.Println("城市：", ipInfo["City"])
	fmt.Println("区域：", ipInfo["Area"])
	fmt.Println("运营商：", ipInfo["Isp"])
}
