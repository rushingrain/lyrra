<p align="center">
  <img src="build/appicon.png" width="128" height="128" alt="lyrra icon" />
</p>

<h1 align="center">lyrra</h1>

<p align="center">基于 Wails、Go、React 与 TypeScript 的第三方酷狗音乐桌面客户端。</p>

<p align="center">
  <img alt="Wails" src="https://img.shields.io/badge/Wails-v2-DD3A2B" />
  <img alt="Go" src="https://img.shields.io/badge/Go-1.25+-00ADD8" />
  <img alt="React" src="https://img.shields.io/badge/React-19-61DAFB" />
  <img alt="TypeScript" src="https://img.shields.io/badge/TypeScript-6-3178C6" />
  <img alt="License" src="https://img.shields.io/badge/License-MIT-yellow" />
</p>

> **重要说明**：lyrra 是非官方第三方项目，与酷狗音乐及其关联公司不存在隶属、授权、合作或背书关系。本项目仅供学习、研究与技术交流，禁止用于商业用途。音乐、封面、歌词、商标等内容的权利归原权利人所有，请支持正版音乐。

## 功能特性

- 扫码登录、登录态恢复与五天固定会话有效期。
- 首页精选活动、官方歌单、每日推荐、最新音乐和热门音乐。
- 综合搜索及单曲、歌单、专辑、歌手、MV、歌词分类搜索，支持滚动分页。
- 官方歌单列表、歌单详情、专辑详情与“播放全部”。
- 桌面播放器：播放/暂停、上一首/下一首、进度拖动、音量、播放模式与播放队列。
- 播放音质标识与降级识别，播放地址签名校验，播放队列和当前歌曲本地恢复。
- 沉浸式歌词页、收藏数/评论数展示、播放列表侧栏。
- 首页数据缓存、搜索请求去重、歌词与统计信息缓存，减少重复网络请求。
- 单实例运行、统一日志、本地账号配置与播放状态持久化。

## 项目状态

| 模块 | 状态 | 说明 |
| --- | --- | --- |
| 首页、官方歌单、歌单/专辑详情 | 已实现 | 接入真实后端数据 |
| 搜索 | 已实现 | 登录后可用，支持综合与分类搜索 |
| 播放器、歌词、播放队列 | 已实现 | 支持本地状态恢复 |
| 扫码登录、登出、会话过期 | 已实现 | 手机号登录仍为待接入状态 |
| 关注、FM 电台、我的喜欢、最近播放、我的云盘 | 页面占位 | 导航入口已保留，业务尚未接入 |

更完整的实现范围、限制和架构说明见 [PROJECT.md](PROJECT.md)。

## 技术栈

| 层级 | 技术 |
| --- | --- |
| 桌面容器 | Wails v2.13 |
| 后端 | Go 1.25、`log/slog`、Lumberjack |
| 前端 | React 19、TypeScript 6、Vite 8 |
| UI | Tailwind CSS 4、Lucide、SimpleBar |
| 状态与路由 | Zustand、React Router 7 |
| 音乐 SDK | [`rushingrain/kugou-music-api`](https://github.com/rushingrain/kugou-music-api) |

## 环境要求

- Go 1.25 或更高版本。
- Node.js `^20.19.0` 或 `>=22.12.0`。
- Wails CLI v2。
- Windows 需要 WebView2 运行时。

安装 Wails CLI：

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

> **SDK 依赖说明**：`go.mod` 固定了可直接获取的 SDK 版本，正常克隆后无需额外配置即可构建。若需要同时修改本地 SDK，在仓库根目录创建不提交的 `go.work`：
>
> ```bash
> go work init . ../kugou-music-api
> ```
>
> 日常 `go build`、`go test` 会使用本地 SDK；提交或发布前执行 `GOWORK=off go build ./...`，验证 `go.mod` 固定版本可以独立构建。

## 快速开始

克隆仓库并进入项目目录后：

```bash
# 安装前端依赖
cd frontend
npm ci
cd ..

# 启动桌面开发环境
wails dev
```

构建当前平台的可执行文件：

```bash
wails build
```

构建产物默认位于 `build/bin/`。

### 仅预览前端界面

```bash
cd frontend
npm ci
npm run dev
```

纯浏览器模式只适合调试界面，部分服务会使用 mock 数据。登录、真实搜索、播放地址解析等 Wails 跨端能力需要在 `wails dev` 环境中验证。

## 数据与日志

应用数据按操作系统写入用户配置目录下的 `lyrra/`：

- macOS：`~/Library/Application Support/lyrra/`
- Windows：`%AppData%\lyrra\`
- Linux：`~/.config/lyrra/`

主要文件包括 `session.json`、`player-device.json`、`player-account.json`、`vip.json` 和 `logs/lyrra.log`。登录态、播放记录等敏感信息不会写入日志。

## 开发检查

```bash
# 前端类型检查
cd frontend && npx tsc --noEmit

# Go 编译检查
go build ./...
```

Go 暴露给前端的 Wails 方法或 DTO 发生变化后，请在仓库根目录运行：

```bash
wails generate module
```

`frontend/wailsjs/` 是生成目录，不应手动修改或提交。

## 项目结构

```text
.
├── main.go                 # Wails 启动、服务注册与单实例配置
├── app.go                  # 窗口级 Wails 方法
├── internal/               # Go 业务服务、DTO、SDK 适配与持久化
├── frontend/src/           # React 页面、组件、状态、hooks 与 Wails 调用封装
├── frontend/wailsjs/       # Wails 自动生成的前端绑定（不提交）
├── build/                  # 应用图标与平台构建资源
├── AGENTS.md               # 自动化编码代理工作规范
└── PROJECT.md              # 项目上下文、架构与实现说明
```

## 参与贡献

欢迎提交 Issue 和 Pull Request。提交前请：

1. 阅读 [PROJECT.md](PROJECT.md) 了解架构和边界。
2. 保持改动聚焦，不混入无关重构。
3. 不提交账号凭证、Cookie、密钥、播放记录或生成代码。

## 免责声明

1. 本项目为个人学习和技术研究项目，不是酷狗音乐官方客户端，也未获得酷狗音乐或其关联公司的授权、认可或支持。
2. 本项目仅用于学习、研究与技术交流，不得用于任何商业用途、非法用途或违反第三方服务条款的用途。
3. 本项目不存储、不提供、不分发任何音乐文件。应用展示和播放的内容来自第三方服务及其 CDN，相关音乐、封面、歌词、商标和其他内容的著作权、商标权及其他权利归原权利人所有。
4. 使用本项目所产生的一切后果由使用者自行承担，包括但不限于第三方账号限制、服务中断、数据丢失或法律风险。项目作者与贡献者不对上述后果承担责任。
5. 请遵守所在地区法律法规以及第三方平台的服务条款，尊重版权，支持正版音乐。请勿将本项目用于大规模抓取、传播盗版内容或其他侵权活动。
6. 本项目按“现状”提供，不提供任何明示或暗示的担保，包括但不限于可用性、稳定性、安全性和特定用途适用性。

### 侵权处理

若任何权利人认为本项目内容侵犯其合法权益，请通过 GitHub Issue 联系，并提供权属证明、具体内容和链接。确认后会尽快删除、修改或下架相关内容。

## 致谢

- [Wails](https://wails.io/)
- [React](https://react.dev/)
- [kugou-music-api](https://github.com/lfhy/kugou-music-api)
- [KuGouMusicApi](https://github.com/MakcRe/KuGouMusicApi)

## 许可证

本项目基于 [MIT License](LICENSE) 开源，Copyright (c) 2026 rushingrain。
