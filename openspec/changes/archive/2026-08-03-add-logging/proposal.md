## Why

桌面应用（wails + Go + React）已实现部分功能，但没有任何日志记录。后续出现问题难以快速定位。需要一套统一、高效、有自动清理能力的日志系统，覆盖 Go 后端与前端。

## What Changes

- 新增统一日志服务：Go 后端使用标准库 `log/slog` 输出结构化日志，写入应用配置目录下的 `logs/kuwave.log`
- 日志文件按大小轮转并自动清理（`lumberjack`）：单文件上限 10MB、保留 3 个备份、最长保留 7 天、过期备份自动压缩删除，防止日志文件无限膨胀
- 新增 Wails `Logger` 适配器：Wails 运行时内部日志路由到 slog（`Fatal` 不退出进程，降级为 `Error`）
- 前端 `console.*` 通过 wails `runtime.LogDebug/LogInfo/LogWarning/LogError` 覆写，前端 React 报错与后端日志落入同一文件
- `main.go` 启动时初始化日志服务，并注入 `App`、`AuthService` 等后端服务；关键业务事件（登录、扫码轮询、API 失败）记录日志

## Capabilities

### New Capabilities
- `logging`: 统一日志能力——后端 slog 结构化日志、Wails 日志适配、前端 console 转发、按大小轮转与自动清理

### Modified Capabilities

（无既有 capability 变更）

## Impact

- **新增依赖**：`gopkg.in/natefinch/lumberjack.v2`（Go）
- **新增文件**：`log_service.go`（日志服务 + Wails 适配器）
- **修改文件**：`main.go`（初始化日志并注入）、`app.go` / `auth_service.go`（接收 logger 并记录关键事件）、`frontend/src/main.tsx`（console 垫片）
- **运行行为**：应用启动后在用户配置目录（macOS `~/Library/Application Support/kuwave/logs/`）生成日志；日志磁盘占用有界（~40MB 上限 + 7 天窗口）
