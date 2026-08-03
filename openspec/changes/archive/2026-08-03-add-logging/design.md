## Context

- 应用为 wails v2 + Go 1.25 + React（HashRouter）桌面端，包名 `main`，后端文件扁平（`app.go`、`auth_service.go`、`session.go` 等）
- 会话已按用户配置目录持久化（`session.go` 的 `os.UserConfigDir()/kuwave/session.json`），日志目录复用同一基址，保持一致性
- 当前 Go 后端无任何日志；前端 `console.*` 输出进 devtools，用户无法看到
- 日志量级低：登录、扫码轮询、API 调用为主，非高频热路径

## Goals / Non-Goals

**Goals:**
- 统一日志：后端（slog）与前端（console → runtime.Log*）落入同一文件
- 性能好：单 writer 常开、直写不缓冲，不影响应用主流程；写日志失败不崩溃应用
- 文件有界：按大小轮转 + 自动清理，磁盘占用可控（~10MB × 3 备份 + 7 天窗口 + 压缩）
- 覆盖 Wails 运行时自身日志（`options.Logger`）

**Non-Goals:**
- 不实现日志查看 UI（日志只是便于事后定位问题，不要求应用内可视化）
- 不做远程日志 / 日志上报
- 不做日志级别运行时热切换（默认 INFO，调试期可改常量/环境变量）
- 不引入 zap / logrus 等重型日志框架

## Decisions

### D1：日志框架用标准库 `log/slog` + `lumberjack`（方案 A，已确认）

- `slog`：Go 1.21+ 标准库，结构化、性能好，零额外依赖。`slog.NewTextHandler` 输出可读文本，适合桌面应用排查
- `lumberjack`（`gopkg.in/natefinch/lumberjack.v2`）：io.Writer 实现，久经考验（zap 默认日志轮转方案），原生支持 `MaxSize`（单文件上限 MB）、`MaxBackups`（保留备份数）、`MaxAge`（天数）、`Compress`（压缩）。**直接满足"文件不宜过大 + 自动清理"要求，无需手写轮转**
- 备选已否决：
  - **Wails `pkg/logger.FileLogger`**：无时间戳、每条日志 open+close 文件（性能差）、无轮转（文件无限膨胀）、写失败 `log.Fatal`（搞崩应用）。三项要求全不满足
  - **手写按大小轮转（方案 B）**：约 60 行，轮转并发边界、目录清理、崩溃一致性都要自己兜底，不如 lumberjack 成熟
  - **zap**：对桌面应用日志量过重

### D2：日志文件路径

`os.UserConfigDir()/kuwave/logs/kuwave.log`，与 `session.go` 的配置目录基址一致。初始化时 `os.MkdirAll` 创建目录（权限 0o700，与 session 一致）。

### D3：lumberjack 轮转参数

- `MaxSize: 10`（MB）
- `MaxBackups: 3`
- `MaxAge: 7`（天）
- `Compress: true`

磁盘上界：10MB × 3 + 当天文件，约 40MB，窗口 7 天。

### D4：直写不缓冲

日志直写 lumberjack（其内部持有文件句柄 + mutex），不加 `bufio.Writer`。原因：写缓冲有崩溃丢日志风险；本应用日志量低，直写性能足够。`slog.HandlerOptions{Level: slog.LevelInfo}` 兜底过滤 DEBUG。

### D5：Wails Logger 适配器

`logger.Logger` 接口（Print/Trace/Debug/Info/Warning/Error/Fatal）实现为 `slog` 适配器，映射：

| logger.Logger | slog |
|---|---|
| Print | Info |
| Trace/Debug | Debug |
| Info | Info |
| Warning | Warn |
| Error/Fatal | Error（Fatal 不 `os.Exit`，降级为 Error） |

作用：Wails 运行时内部日志进入同一文件；`Fatal` 不退出进程，避免单条日志错误搞崩应用。

### D6：前端 console 垫片

在 `frontend/src/main.tsx` 覆写 `console.debug/info/warn/error`，转发到 wails `runtime.LogDebug/LogInfo/LogWarning/LogError`。Wails v2 运行时已自动注入 `window.runtime`，这些方法路由到 `options.Logger` → 适配器 → slog → 同一日志文件。仅在 `window.runtime` 存在时转发，保留原 `console.*` 行为。

### D7：日志服务形态

新增 `log_service.go`（包 `main`，与现有扁平结构一致）：

```go
type Service struct { logger *slog.Logger; writer *lumberjack.Logger }
func New() *Service                    // 初始化目录 + lumberjack + slog.TextHandler
func (s *Service) Logger() *slog.Logger // 供 App / AuthService 使用
func (s *Service) Close() error        // 应用退出前关闭 lumberjack
```

`main.go` 最先初始化 `Service`，注入 `App` 与 `AuthService`，并设 `options.Logger = logservice.NewWailsAdapter(service)`。`app.go` / `auth_service.go` 增加构造函数参数接收 `*slog.Logger` 或 `*Service`，记录关键事件（启动、登录成功、扫码状态、API 错误、服务初始化失败）。

## Risks / Trade-offs

- [lumberjack 仓库已归档（2022 年停止维护）] → 功能稳定、API 简单、被 zap 等广泛使用，对只用到轮转的场景风险可忽略；如未来需替换，io.Writer 抽象使切换成本低
- [日志写入阻塞调用方（lumberjack 内部 mutex）] → 本应用日志量低，锁竞争可忽略；`slog.LevelInfo` 过滤降低写入频率
- [Fatal 降级为 Error，语义弱化] → 符合桌面应用"日志不应退出进程"原则；真实致命错误仍由错误处理链处理
- [前端 console 垫片影响面] → 仅转发，不改写原输出；`runtime` 缺失时静默跳过，不抛错

## Migration Plan

- 一次性新增依赖 + 新文件，无数据迁移；`main.go` 初始化顺序调整（日志最先）
- 回滚：删除 `log_service.go` 与注入点改动即可，不影响既有功能

## Open Questions

- 日志级别是否暴露运行时切换（当前无，默认 INFO）——留待需要时扩展
