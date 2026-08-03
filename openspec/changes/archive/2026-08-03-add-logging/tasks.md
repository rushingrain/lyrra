## 1. 依赖与日志服务

- [x] 1.1 添加依赖 `gopkg.in/natefinch/lumberjack.v2` 到 `go.mod`（`go get`）
- [x] 1.2 新建 `log_service.go`：实现 `Service` 结构体（slog + lumberjack），含 `New()`、`Logger()`、`Close()`；目录 `os.UserConfigDir()/kuwave/logs`，`MkdirAll` 0o700
- [x] 1.3 配置 lumberjack：`MaxSize=10`、`MaxBackups=3`、`MaxAge=7`、`Compress=true`；slog 用 `TextHandler`，级别 `slog.LevelInfo`

## 2. Wails 适配器

- [x] 2.1 在 `log_service.go` 实现 `logger.Logger` 接口适配器（Print/Trace/Debug/Info/Warning/Error/Fatal → slog 映射，Fatal 降级 Error，不退出）
- [x] 2.2 提供构造适配器的方法（如 `NewWailsAdapter(service)`）

## 3. 接入 main 与服务

- [x] 3.1 `main.go` 最先初始化日志服务；设 `options.Logger` 为适配器；应用退出前调用 `Close`
- [x] 3.2 `app.go` 构造函数接收日志器，`startup` 记录应用启动
- [x] 3.3 `auth_service.go` 构造函数接收日志器，记录登录/扫码/API 失败等关键事件（如 `NewAuthService` 初始化失败、`GetLoginQrCode` 错误、`CheckQrLoginStatus` 登录成功）

## 4. 前端 console 转发

- [x] 4.1 `frontend/src/main.tsx` 覆写 `console.debug/info/warn/error`，转发到 `window.runtime.LogDebug/LogInfo/LogWarning/LogError`；`runtime` 缺失时静默跳过，保留原输出

## 5. 构建与验证

- [x] 5.1 `go build` 通过（含 lumberjack 依赖解析）
- [x] 5.2 前端 `tsc`/构建通过
- [x] 5.3 运行应用，确认 `.../kuwave/logs/kuwave.log` 生成、含时间戳与级别、前端 console 与后端日志落入同一文件
- [x] 5.4 验证轮转/清理逻辑（如临时调小 MaxSize 触发轮转，或单测覆盖 lumberjack 配置）
