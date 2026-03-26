package utils

import (
	"log/slog"
	"os"
	"path/filepath"
)

// InitSlog 初始化日志
// 控制台：Text 格式（易读）
// 文件：JSON 格式（便于解析、采集）
func InitSlog() error {
	logDir := "logs"

	// 1. 创建日志目录（不存在则自动创建）
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		return err
	}

	// 2. 打开/创建日志文件（追加写入模式）
	logFile, err := os.OpenFile(
		filepath.Join(logDir, "plant_be.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		return err
	}

	// 3. 配置日志通用选项（级别、源码位置等）
	opts := &slog.HandlerOptions{
		Level:     slog.LevelDebug, // 日志级别：Debug及以上全部输出
		AddSource: true,            // 可选：打印日志所在文件+行号，方便调试
	}

	// 4. 创建两个独立处理器
	consoleHandler := slog.NewTextHandler(os.Stdout, opts) // 控制台：TEXT 格式
	fileHandler := slog.NewJSONHandler(logFile, opts)      // 文件：JSON 格式

	// 5. 组合多处理器（核心：同时输出到两个目标）
	multiHandler := slog.NewMultiHandler(consoleHandler, fileHandler)

	// 6. 设置为全局默认 Logger（直接使用 slog.Info() 即可生效）
	slog.SetDefault(slog.New(multiHandler))

	// 初始化完成日志（测试用）
	slog.Info("日志初始化成功", "log_file", filepath.Join(logDir, "plant_be.log"))
	return nil
}
