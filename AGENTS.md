# AGENTS.md

本文件是 AI 编程代理的操作指令，只记录代理容易出错、必须遵守的规则。
项目背景、架构、目录约定和编码规范见 `PROJECT.md`，此处不重复。
本文件适用于整个仓库。

## 改动前必读

- 修改跨端数据结构前，先读 `PROJECT.md` 的“关键决策”和“约束”两节。
- 改动 Go 暴露的 Wails 方法签名、返回结构或 JSON 字段名属于契约变更：必须同步前端类型与调用处，并运行 `wails generate module`。
- 新增业务服务时，字段回退、`is_original`、歌曲名后缀、封面优先级、分页回退、原子写入等逻辑必须调用 `internal/catalogmap/`、`internal/paging/`、`internal/atomicfile/` 等现有包，不得在 service 内重写。发现重复实现时报告，不要自行抽取。
- 不新建 `utils`、`common`、`helpers` 等杂物袋包；只有出现两个以上真实调用点且语义一致时才抽取公共代码。

## 生成代码与 SDK

- `frontend/wailsjs/` 由 Wails 自动生成，不手动修改，也不提交到 Git。Go 暴露方法变化后，在仓库根目录运行 `wails generate module`，再检查前端调用的类型。
- `github.com/rushingrain/kugou-music-api` 是从 JavaScript 版移植的 Go SDK。排查接口异常时，考虑 JS 与 Go 在 `null`、数值和动态类型等方面的语义差异。
- JS 原版仓库是排查移植差异的唯一参照标准，本地路径：`../KuGouMusicApi`。仅作只读参考，禁止修改。
- 若 JS 原版目录不存在或无法读取，视为缺少修复依据：立即停止，向开发者报告“需要提供 JS 原版仓库”，等待提供后再继续。不得凭记忆或推测替代参照物。
- 怀疑问题出在 SDK 时，先对照 JS 原版确认预期行为，再停下向开发者报告，未经明确指示不得修改 SDK 源码。
- 严禁在业务代码中重写 JS 版 SDK 已有的逻辑。若 Go 版 SDK 与 JS 版有差异，应优先在 Go 版 SDK 中修复。
- SDK 本地联调使用仓库根目录的 `go.work`（指向本地 SDK，`go.work`、`go.work.sum` 不提交）；`go.mod` 只保留可获取的固定 SDK 版本，不写本机绝对路径 `replace`。
- 提交前必须执行 `GOWORK=off go build ./...` 验证干净构建；若本地缺少 SDK 源码或固定版本无法获取，报告验证受阻，不擅自修改依赖指向。
- SDK 修复验证通过后，先提交并推送 SDK 变更，再把 `go.mod` 更新到对应 commit 或 tag；不要在业务代码中复制 SDK 逻辑。

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
