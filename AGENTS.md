# 项目说明

lyrra 是基于 Wails v2、Go、React 和 TypeScript 的第三方酷狗音乐桌面客户端。本文件适用于整个仓库。

## 职责与调用边界

- 涉及酷狗服务器的业务请求、登录态、响应解析和持久化由 Go 后端负责。前端负责界面、交互、播放控制，以及调用 Go 暴露的 Wails 方法。
- 保留现有例外：WebView 可直接加载后端返回的图片和音频 CDN 地址；纯浏览器开发模式可使用现有 mock 展示界面。真实跨端功能须在 Wails 环境中验证。
- Wails 服务在 `main.go` 的 `Bind` 中注册。前端在 `frontend/src/services/` 封装生成的绑定，页面和组件优先通过该层调用后端。
- Go 侧将 SDK 的动态响应整理为稳定 DTO；修改跨端数据结构时同步检查 Go 的 JSON 字段、前端类型和调用处。

## 目录与命名

- `internal/` 按业务域分包，如 `auth`、`search`、`play`、`vip`。服务入口通常放在 `service.go`，模型、持久化等逻辑分别放在 `model.go`、`store.go` 等文件。跨服务复用的音乐实体放在 `internal/dto/`；SDK 动态值读取放在 `internal/kgapi/`，音乐实体映射放在 `internal/catalogmap/`，通用分页回退放在 `internal/paging/`，原子文件写入放在 `internal/atomicfile/`。业务服务负责流程编排，不复制这些共享能力。
- `frontend/src/pages/` 放页面，`components/` 放组件，`services/` 放 Wails 调用封装，`stores/` 放全局状态，`types/` 放共用类型，`hooks/` 放 React hooks。
- 沿用现有命名：Go 包和文件名使用小写；React 页面与组件文件使用 PascalCase，服务与 hook 文件使用 camelCase，测试文件与对应 Go 代码相邻并以 `_test.go` 结尾。

## Go 后端复用约定

- `internal/dto/` 只放稳定的跨端实体；`internal/kgapi/` 统一处理 SDK client 和动态字段；`internal/catalogmap/` 统一处理响应到实体的字段兼容与领域映射。业务服务不重复写字段回退、`is_original`、歌曲名后缀、封面优先级等逻辑。
- 同一实体在不同接口的字段差异只在 `catalogmap` 适配；新增同类接口时新增语义明确的映射函数，不在业务服务里再写一套 mapper。
- `play` 的 hash 优先级、`paging` 的分页回退、`atomicfile` 的原子写入、`auth` 的登录 cookie 选择，分别以现有包为准，不在调用方复制。
- 只有出现两个以上真实调用点且语义一致时才抽取公共代码；流程相似但副作用、重试或错误处理不同的代码，保留各自实现。
- 不建 `utils`、`common`、`helpers` 等杂物袋包，不为一次性逻辑强行抽象。
- 重构不得改变 Wails 方法签名、返回结构和 JSON 字段名；新增契约时再同步前端并运行 `wails generate module`。

## 注释语言
- 代码注释使用中文书写，力求简洁、清晰、准确表达意图。不要为了注释而注释；只在逻辑不直观、存在隐含约束或需要解释“为什么”时添加注释。
- 避免直译式或啰嗦的注释；注释应说明目的或原因，而非重复代码字面含义。

## 生成代码与 SDK

- `frontend/wailsjs/` 由 Wails 自动生成，不手动修改，也不提交到 Git。Go 暴露方法发生变化后，在仓库根目录运行 `wails generate module`，再检查前端调用的类型。
- `github.com/rushingrain/kugou-music-api` 是从 JavaScript 版移植的 Go SDK。排查接口异常时，考虑 JS 与 Go 在 `null`、数值和动态类型等方面的语义差异。
- JS 原版仓库是排查移植差异的唯一参照标准，本地路径：`../KuGouMusicApi`。仅作只读参考，禁止修改。
- 若 JS 原版目录不存在或无法读取，视为缺少修复依据：立即停止，向开发者报告"需要提供 JS 原版仓库"，等待提供后再继续。不得凭记忆或推测替代参照物。
- 怀疑问题出在 SDK 时，先对照 JS 原版确认预期行为，再停下向开发者报告，未经明确指示不得修改 SDK 源码。
- 严禁在业务代码中重写 JS 版 SDK 已有的逻辑。若 Go 版 SDK 与 JS 版有差异，应优先在 Go 版 SDK 中修复。
- `go.mod` 当前用本机绝对路径 `replace` 指向 SDK 源码，便于开发期间修复 SDK。暂时保留该配置；本地缺少该目录时，说明验证受阻的原因，不擅自更改依赖指向。

## 敏感区域

- 不要硬编码任何密钥、token 或 API Key，也不要将其提交到 Git。
- 用户登录态、播放记录等数据不得写入日志，也不得上传到任何未明确授权的地方。调试日志只记录必要的非敏感信息；媒体播放地址中的账号标识也要遵循现有脱敏处理。
- 不要擅自修改加密或解密逻辑，包括算法、密钥来源和密钥处理方式。
- 不要擅自修改 CI/CD 配置。

## 验证

- 前端改动后在 `frontend/` 执行 `npx tsc --noEmit`，确保没有 TypeScript 错误。
- Go 改动后在仓库根目录执行 `go build ./...`，确保 Go 代码可编译。干净检出中先运行 `wails generate module` 生成前端绑定；若缺少被根包嵌入的 `frontend/dist`，再在 `frontend/` 执行 `npm run build`，然后检查 Go 构建。
- Agent 负责可自动化的静态检查，并报告实际执行的命令与结果。涉及跨端调用、登录、播放等运行行为，由开发者在桌面应用中验证，agent 不要自行启动应用。
- 仅在需要时运行相关 Go 测试。带 `KUGOU_LIVE_TEST=1` 的测试会访问真实服务并可能读取本地登录态，不作为常规静态检查运行。
