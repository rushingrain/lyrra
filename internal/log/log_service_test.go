package log

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/natefinch/lumberjack.v2"
)

// TestServiceWritesFile 验证日志写入文件且包含时间戳与级别。
func TestServiceWritesFile(t *testing.T) {
	dir := t.TempDir()
	svc := newService(filepath.Join(dir, "kuwave.log"))
	defer func() {
		if err := svc.Close(); err != nil {
			t.Errorf("关闭日志服务失败: %v", err)
		}
	}()

	svc.Logger().Info("hello kuwave", "key", "val")

	data, err := os.ReadFile(filepath.Join(dir, "kuwave.log"))
	if err != nil {
		t.Fatalf("读取日志文件失败: %v", err)
	}
	line := string(data)
	if !strings.Contains(line, "hello kuwave") {
		t.Errorf("日志缺少消息: %s", line)
	}
	if !strings.Contains(line, "INFO") {
		t.Errorf("日志缺少级别: %s", line)
	}
	// TextHandler 默认输出 time=... 时间戳
	if !strings.HasPrefix(line, "time=") {
		t.Errorf("日志缺少时间戳: %s", line)
	}
}

// TestLumberjackRotation 验证 lumberjack 达到 MaxSize 后轮转出备份文件。
func TestLumberjackRotation(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "kuwave.log")
	lj := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    1, // 1MB，方便触发轮转
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   true,
	}
	defer func() {
		if err := lj.Close(); err != nil {
			t.Errorf("关闭 lumberjack 失败: %v", err)
		}
	}()

	// 写入 1.5MB，触发至少一次轮转
	chunk := make([]byte, 512*1024)
	for i := range chunk {
		chunk[i] = 'a'
	}
	for i := 0; i < 3; i++ {
		if _, err := lj.Write(chunk); err != nil {
			t.Fatalf("写入失败: %v", err)
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("读取目录失败: %v", err)
	}
	backups := 0
	for _, e := range entries {
		if strings.Contains(e.Name(), "kuwave-") {
			backups++
		}
	}
	if backups < 1 {
		t.Errorf("期望至少 1 个轮转备份，实际 %d 个（目录: %v）", backups, entries)
	}
}
