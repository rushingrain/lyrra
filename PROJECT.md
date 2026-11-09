# PROJECT.md

lyrra 是基于 Wails v2、Go、React 和 TypeScript 的第三方酷狗音乐桌面客户端。

本文件记录项目背景、架构决策、目录约定和编码规范，供开发者和 AI 代理理解项目上下文。
AI 代理的操作指令见 `AGENTS.md`；项目启动方式见 `README.md`。

## 项目定位

- 第三方酷狗音乐桌面客户端，基于 Wails v2、Go、React、TypeScript。
- 目标：在桌面端提供酷狗音乐的核心使用体验（登录、搜索、播放、VIP 相关能力）。

## 关键决策：职责与调用边界

- 涉及酷狗服务器的业务请求、登录态、响应解析和持久化由 Go 后端负责。前端负责界面、交互、播放控制，以及调用 Go 暴露的 Wails 方法。
- 保留现有例外：WebView 可直接加载后端返回的图片和音频 CDN 地址；纯浏览器开发模式可使用现有 mock 展示界面。真实跨端功能须在 Wails 环境中验证。
- Wails 服务在 `main.go` 的 `Bind` 中注册。前端在 `frontend/src/services/` 封装生成的绑定，页面和组件优先通过该层调用后端。
- Go 侧将 SDK 的动态响应整理为稳定 DTO；修改跨端数据结构时需同步检查 Go 的 JSON 字段、前端类型和调用处。

## 目录与命名

- `internal/` 按业务域分包，如 `auth`、`search`、`play`、`vip`。服务入口通常放在 `service.go`，模型、持久化等逻辑分别放在 `model.go`、`store.go` 等文件。
- 跨服务复用的音乐实体放在 `internal/dto/`；SDK 动态值读取放在 `internal/kgapi/`；音乐实体映射放在 `internal/catalogmap/`；通用分页回退放在 `internal/paging/`；原子文件写入放在 `internal/atomicfile/`。业务服务负责流程编排，不复制这些共享能力。
- `frontend/src/pages/` 放页面，`components/` 放组件，`services/` 放 Wails 调用封装，`stores/` 放全局状态，`types/` 放共用类型，`hooks/` 放 React hooks。
- 沿用现有命名：Go 包和文件名使用小写；React 页面与组件文件使用 PascalCase，服务与 hook 文件使用 camelCase，测试文件与对应 Go 代码相邻并以 `_test.go` 结尾。

## 约束：Go 后端复用

- `internal/dto/` 只放稳定的跨端实体；`internal/kgapi/` 统一处理 SDK client 和动态字段；`internal/catalogmap/` 统一处理响应到实体的字段兼容与领域映射。业务服务不重复写字段回退、`is_original`、歌曲名后缀、封面优先级等逻辑。
- 同一实体在不同接口的字段差异只在 `catalogmap` 适配；新增同类接口时新增语义明确的映射函数，不在业务服务里再写一套 mapper。
- `play` 的 hash 优先级、`paging` 的分页回退、`atomicfile` 的原子写入、`auth` 的登录 cookie 选择，分别以现有包为准，不在调用方复制。
- 只有出现两个以上真实调用点且语义一致时才抽取公共代码；流程相似但副作用、重试或错误处理不同的代码，保留各自实现。
- 不建 `utils`、`common`、`helpers` 等杂物袋包，不为一次性逻辑强行抽象。
- 重构不得改变 Wails 方法签名、返回结构和 JSON 字段名；新增契约时再同步前端并运行 `wails generate module`。

## 编码规范：注释语言

- 代码注释使用中文书写，力求简洁、清晰、准确表达意图。不要为了注释而注释；只在逻辑不直观、存在隐含约束或需要解释“为什么”时添加注释。
- 避免直译式或啰嗦的注释；注释应说明目的或原因，而非重复代码字面含义。

## SDK 依赖与本地开发

- `go.mod` 固定可获取的 SDK 版本，不使用本机绝对路径 `replace`；克隆、CI 和发布构建都以此为准。
- 需要同步修改 SDK 时，在仓库根目录创建不提交的 `go.work`：

  ```bash
  go work init . ../kugou-music-api
  ```

  日常构建会使用本地 SDK；`go.work` 和 `go.work.sum` 已加入 `.gitignore`。
- 提交或发布前执行 `GOWORK=off go build ./...`，确保固定版本可以独立构建。SDK 修复完成后先提交并推送 SDK，再更新 `go.mod` 到对应 commit 或 tag。
