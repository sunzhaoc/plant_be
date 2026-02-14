package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Ip9Response struct {
	Ret  int    `json:"ret"`
	Data IpData `json:"data"`
	Qt   int    `json:"qt"`
}

type IpData struct {
	Country string `json:"country"`
	Prov    string `json:"prov"`
	City    string `json:"city"`
	Area    string `json:"area"`
	Isp     string `json:"isp"`
	BigArea string `json:"big_area"`
}

func GetIpInfo(ip string) (map[string]string, error) {
	url := fmt.Sprintf("https://ip9.com.cn/get?ip=%s", ip)

	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		return nil, fmt.Errorf("创建请求失败：%w", err)
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败：%w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("接口返回非200状态码：%d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败：%w", err)
	}

	var ipResp Ip9Response
	err = json.Unmarshal(body, &ipResp)
	if err != nil {
		return nil, fmt.Errorf("解析JSON失败：%w，原始响应：%s", err, string(body))
	}

	ipInfoMap := map[string]string{
		"Country": ipResp.Data.Country,
		"Prov":    ipResp.Data.Prov,
		"City":    ipResp.Data.City,
		"Area":    ipResp.Data.Area,
		"Isp":     ipResp.Data.Isp,
	}

	return ipInfoMap, nil
}
