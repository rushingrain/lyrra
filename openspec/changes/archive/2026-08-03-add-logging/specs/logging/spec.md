## ADDED Requirements

### Requirement: 后端日志写入文件

系统 SHALL 在启动时初始化统一日志服务，后端日志使用标准库 `log/slog` 结构化输出，写入用户配置目录 `os.UserConfigDir()/kuwave/logs/kuwave.log`（目录不存在时自动创建，权限 0o700）。日志 SHALL 默认以 INFO 级别过滤。

#### Scenario: 日志目录不存在

- **WHEN** 应用启动且 `.../kuwave/logs/` 目录不存在
- **THEN** 系统自动创建该目录，且日志能正常写入 `kuwave.log`

#### Scenario: 后端服务记录日志

- **WHEN** 后端执行关键业务事件（登录、扫码状态查询、API 失败、服务初始化）
- **THEN** 结构化日志写入 `kuwave.log`，包含时间戳与日志级别

### Requirement: 日志文件自动轮转与清理

系统 SHALL 使用 `lumberjack` 对日志文件按大小轮转，并自动清理过期备份，保证磁盘占用有界：单文件上限 10MB、保留最多 3 个备份、最长保留 7 天、过期备份自动压缩。

#### Scenario: 单文件达到大小上限

- **WHEN** `kuwave.log` 大小达到 10MB
- **THEN** 系统将其轮转为备份文件，并开始写入新的 `kuwave.log`，备份数不超过 3 个

#### Scenario: 备份超过保留窗口

- **WHEN** 备份文件存在时间超过 7 天或数量超过 3 个
- **THEN** 系统自动删除过期备份，磁盘占用维持有界

### Requirement: Wails 运行时日志接入

系统 SHALL 提供 `logger.Logger` 接口适配器，将 Wails 运行时内部日志路由到 slog，写入同一日志文件。`Fatal` 级别 SHALL 降级为 `Error`，不得导致进程退出。

#### Scenario: Wails 运行时产生日志

- **WHEN** Wails 运行时内部产生 debug/info/warning/error 日志
- **THEN** 日志经适配器写入 `kuwave.log`，级别正确映射

#### Scenario: 适配器收到 Fatal 日志

- **WHEN** Wails 运行时产生 Fatal 级日志
- **THEN** 以 Error 级别写入文件，应用进程不退出

### Requirement: 前端 console 日志转发

系统 SHALL 在前端启动时覆写 `console.debug/info/warn/error`，通过 wails `runtime.LogDebug/LogInfo/LogWarning/LogError` 转发到后端日志，使前端日志落入同一文件。运行时对象不存在时 SHALL 静默跳过，不影响原 console 输出。

#### Scenario: 前端产生日志

- **WHEN** 前端代码执行 `console.log` / `console.error` 且 `window.runtime` 存在
- **THEN** 对应级别日志写入 `kuwave.log`，与原 console 输出同时保留

#### Scenario: 运行时对象缺失

- **WHEN** 前端执行 console 方法且 `window.runtime` 不存在（如浏览器调试环境）
- **THEN** 控制台正常输出，不抛错、不转发

### Requirement: 日志服务注入后端服务

系统 SHALL 在 `main.go` 最先初始化日志服务，并将日志器注入 `App` 与 `AuthService`；应用退出前 SHALL 关闭日志写入器（`Close`）。

#### Scenario: 日志服务初始化与注入

- **WHEN** 应用启动，`main.go` 初始化日志服务并创建 `App`、`AuthService`
- **THEN** 两服务持有的日志器可正常记录日志，`options.Logger` 设置为适配器

#### Scenario: 应用退出

- **WHEN** 应用正常退出
- **THEN** 日志写入器被关闭，未写入数据落盘
