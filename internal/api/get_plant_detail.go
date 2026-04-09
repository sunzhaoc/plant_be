package api

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sunzhaoc/plant_be/pkg/db/mysql"
	"github.com/sunzhaoc/plant_be/pkg/db/mysql/models"
	"github.com/sunzhaoc/plant_be/pkg/db/redis"
	"github.com/sunzhaoc/plant_be/pkg/utils"
)

func recordUserAccessPlantDetailLog(ctx *gin.Context, plantId string) {
	userIdVal, exists := ctx.Get("userId")
	if !exists || userIdVal == nil {
		slog.Warn("asyncPlantTask: userId不存在或为空")
		return
	}
	userId, ok := userIdVal.(uint)
	if !ok {
		slog.Error("asyncPlantTask: userId类型错误，期望uint", "actual_type", slog.Any("type", userIdVal))
		return
	}

	deviceInfoJSON := json.RawMessage("{}")
	if encodedDeviceInfo := ctx.GetHeader("X-Device-Info"); encodedDeviceInfo != "" {
		if decodedStr, err := url.QueryUnescape(encodedDeviceInfo); err == nil {
			var temp json.RawMessage
			if err := json.Unmarshal([]byte(decodedStr), &temp); err == nil {
				deviceInfoJSON = temp
			}
		}
	}

	// 获取客户端IP
	clientIP := ctx.ClientIP()
	ipInfo := map[string]string{
		"Country": "",
		"Prov":    "",
		"City":    "",
		"Area":    "",
		"Isp":     "",
	}

	if clientIP == "" {
		// 客户端IP为空，获取失败
		slog.Warn("asyncPlantTask: 无法获取客户端IP")
		clientIP = "unknown"
	} else {
		// 客户端IP不为空，获取成功
		rdb, err := redis.GetDb("ali")
		if err != nil {
			// Redis 客户端异常
			slog.Warn("asyncPlantTask: 获取Redis客户端失败，降级到直接查询IP信息", "error", err)
		} else {
			// Redis 客户端正常
			redisKey := "ip_info:" + clientIP // 定义Redis key（格式：ip_info:客户端IP）
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			ipInfoStr, err := rdb.Get(ctx, redisKey).Result() // 从Redis获取用户ip信息

			// Redis获取成功则反序列化使用
			if err == nil && ipInfoStr != "" {
				if err := json.Unmarshal([]byte(ipInfoStr), &ipInfo); err == nil {
					slog.Debug("asyncPlantTask: 从Redis缓存获取IP信息成功", "clientIP", clientIP)
				} else {
					slog.Warn("asyncPlantTask: Redis缓存IP信息反序列化失败，降级到查询IP信息", "error", err)
				}
			} else if !errors.Is(err, redis.Nil) {
				slog.Warn("asyncPlantTask: 从Redis获取IP信息失败，降级到查询IP信息", "error", err)
			}
		}
		if ipInfo["Country"] == "" && ipInfo["Prov"] == "" && ipInfo["City"] == "" {
			var err error
			ipInfo, err = utils.GetIpInfo(clientIP)
			if err != nil {
				slog.Warn("asyncPlantTask: 获取IP信息失败", "error", err)
			} else {
				if rdb != nil {
					ctxRedis, cancel := context.WithTimeout(context.Background(), 2*time.Second)
					defer cancel()
					ipInfoBytes, _ := json.Marshal(ipInfo)
					err := rdb.SetEx(
						ctxRedis,
						"ip_info:"+clientIP,
						string(ipInfoBytes),
						15*24*time.Hour, // 15天有效期
					).Err()
					if err != nil {
						slog.Warn("asyncPlantTask: IP信息写入Redis失败", "error", err, "clientIP", clientIP)
					} else {
						slog.Debug("asyncPlantTask: IP信息写入Redis成功", "clientIP", clientIP)
					}
				}
			}
		}
	}

	// 获取数据库连接
	db, err := mysql.GetDB("ali")
	if err != nil {
		slog.Error("asyncPlantTask: 数据库连接失败", "error", err)
		return
	}

	plantIdUint, err := strconv.ParseUint(plantId, 10, 64)
	if err != nil {
		slog.Error("asyncPlantTask: plantId解析为uint失败", "plantId", plantId, "error", err)
		return
	}
	plantIdFinal := uint(plantIdUint)

	newAccessLog := models.UserAccessPlantDetailLog{
		UserId:     userId,
		PlantId:    plantIdFinal,
		Ip:         clientIP,
		Country:    ipInfo["Country"],
		Province:   ipInfo["Prov"],
		City:       ipInfo["City"],
		Area:       ipInfo["Area"],
		Isp:        ipInfo["Isp"],
		DeviceInfo: deviceInfoJSON,
	}
	if err := db.Create(&newAccessLog).Error; err != nil {
		slog.Error("asyncPlantTask: 插入用户访问日志失败", "error", err, "user_id", userId, "plant_id", plantIdFinal)
		return
	}
	return
}

func GetPlantDetail(c *gin.Context) {
	plantId := c.Param("plantId")
	go recordUserAccessPlantDetailLog(c.Copy(), plantId) // 记录一下日志

	db, err := mysql.GetDB("ali")
	if err != nil {
		slog.Error("数据库连接失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器内部错误"})
		return
	}

	// 获取植物规格数据
	type PlantSku = struct {
		SkuId uint64  `json:"sku_id"`
		Size  string  `json:"size"`
		Price float64 `json:"price"`
		Stock uint    `json:"stock"`
	}
	var plantSkuList []PlantSku
	query := "SELECT `id` sku_id, `size`, price, stock FROM plant.plant_sku WHERE plant_id = ? ORDER BY sort;"
	skuResult := db.Raw(query, plantId).Scan(&plantSkuList)
	if skuResult.Error != nil {
		slog.Error("查询植物SKU列表失败", slog.Any("error", skuResult.Error))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "查询植物SKU列表失败",
		})
		return
	}

	// 获取植物规格图片
	type PlantImage = struct {
		ImgUrl string `json:"img_url"`
	}
	var plantImageList []PlantImage
	query = "SELECT img_url FROM plant.plant_image WHERE plant_id = ? ORDER BY sort;"
	imageResult := db.Raw(query, plantId).Scan(&plantImageList)
	if imageResult.Error != nil {
		slog.Error("查询植物图片列表失败", slog.Any("error", imageResult.Error))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "查询植物图片列表失败",
		})
		return
	}

	// 获取植物的介绍
	type PlantMain = struct {
		Name      string `json:"name"`
		LatinName string `json:"latin_name"`
		Detail    string `json:"detail"`
	}
	var plantDetail PlantMain
	query = "SELECT `name`, `latin_name`, `detail` FROM plant.plants WHERE id = ?;"
	detailResult := db.Raw(query, plantId).Scan(&plantDetail)
	if detailResult.Error != nil {
		slog.Error("查询植物介绍失败", slog.Any("error", detailResult.Error))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "查询植物图片列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取植物详情成功",
		"data": gin.H{
			"plant_name":       plantDetail.Name,
			"plant_latin_name": plantDetail.LatinName,
			"plant_detail":     plantDetail.Detail,
			"skus":             plantSkuList,
			"images":           plantImageList,
		},
	})
}
